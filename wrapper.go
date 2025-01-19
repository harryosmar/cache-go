package cache_go

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func GetFromCache[TData any, TId any](
	ctx context.Context,
	repo CacheRepo,
	id TId,
	prefixKey string,
	exp time.Duration,
	fnCacheable func(ctx context.Context, id TId) (*TData, error),
) (*TData, error) {
	var (
		err   error
		key   = fmt.Sprintf("%s:%v", prefixKey, id)
		data  TData
		entry = logrus.WithField("key", key)
	)

	defer func() {
		if err != nil {
			entry = entry.WithField("err", err.Error())
			entry.Errorf("GetFromCache got err")
		}
	}()

	// get from cache
	bytesFromCache, found, err := repo.Get(ctx, key)
	if found {
		err = json.Unmarshal(bytesFromCache, &data)
		if err == nil {
			return &data, nil
		}
	}

	// not found or err
	// 1. get from source
	dataFromSource, err := fnCacheable(ctx, id)
	if err != nil {
		entry = entry.WithField("err_type", "call fnCacheable")
		return nil, err
	}
	if dataFromSource == nil {
		return nil, nil
	}

	if exp.Seconds() < 0 {
		// skip cache
		return dataFromSource, nil
	}

	// 2. cache dataFromSource
	bytesFromSource, err := json.Marshal(dataFromSource)
	if err != nil {
		entry = entry.WithField("err_type", "json.Marshal dataFromSource")
		return nil, err
	}
	if exp.Seconds() == 0 {
		err = repo.StoreWithoutTTL(ctx, key, bytesFromSource)
		entry = entry.WithField("err_type", "call cache.StoreWithoutTTL")
	} else {
		err = repo.Store(ctx, key, bytesFromSource, exp)
		entry = entry.WithField("err_type", "call cache.Store")
	}

	return dataFromSource, nil
}

func GetFromCacheWithDynamicTTL[TData any, TId any](
	ctx context.Context,
	repo CacheRepo,
	id TId,
	prefixKey string,
	fnGetTtl func(ctx context.Context, data *TData) time.Duration,
	fnCacheable func(ctx context.Context, id TId) (*TData, error),
) (*TData, error) {
	var (
		err   error
		key   = fmt.Sprintf("%s:%v", prefixKey, id)
		data  TData
		entry = logrus.WithField("key", key)
	)

	defer func() {
		if err != nil {
			entry = entry.WithField("err", err.Error())
			entry.Errorf("GetFromCache got err")
		}
	}()

	// get from cache
	bytesFromCache, found, err := repo.Get(ctx, key)
	if found {
		err = json.Unmarshal(bytesFromCache, &data)
		if err == nil {
			return &data, nil
		}
	}

	// not found or err
	// 1. get from source
	dataFromSource, err := fnCacheable(ctx, id)
	if err != nil {
		entry = entry.WithField("err_type", "call fnCacheable")
		return nil, err
	}
	if dataFromSource == nil {
		return nil, nil
	}

	// 2. cache dataFromSource
	bytesFromSource, err := json.Marshal(dataFromSource)
	if err != nil {
		entry = entry.WithField("err_type", "json.Marshal dataFromSource")
		return nil, err
	}

	exp := fnGetTtl(ctx, dataFromSource)
	if exp.Seconds() < 0 {
		// skip cache
		return dataFromSource, nil
	}

	if exp.Seconds() == 0 {
		err = repo.StoreWithoutTTL(ctx, key, bytesFromSource)
		entry = entry.WithField("err_type", "call cache.StoreWithoutTTL")
	} else {
		err = repo.Store(ctx, key, bytesFromSource, exp)
		entry = entry.WithField("err_type", "call cache.Store")
	}

	return dataFromSource, nil
}
