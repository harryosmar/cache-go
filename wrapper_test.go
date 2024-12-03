package cache_go

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/harryosmar/cache-go/mocks"
	"github.com/stretchr/testify/assert"
)

type TestData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestGetFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockCacheRepo(ctrl)
	testData := &TestData{ID: 1, Name: "test"}
	testBytes, _ := json.Marshal(testData)

	tests := []struct {
		name        string
		setupMock   func()
		id          int
		prefixKey   string
		exp         time.Duration
		fnCacheable func(ctx context.Context, id int) (*TestData, error)
		want        *TestData
		wantErr     bool
	}{
		{
			name: "Success - Data found in cache",
			setupMock: func() {
				mockRepo.EXPECT().Get(ctx, "test:1").Return(testBytes, true, nil)
			},
			id:        1,
			prefixKey: "test",
			exp:       time.Minute,
			fnCacheable: func(ctx context.Context, id int) (*TestData, error) {
				return nil, nil // Should not be called
			},
			want:    testData,
			wantErr: false,
		},
		{
			name: "Success - Data not in cache, fetched from source",
			setupMock: func() {
				mockRepo.EXPECT().Get(ctx, "test:1").Return([]byte{}, false, nil)
				mockRepo.EXPECT().Store(ctx, "test:1", testBytes, time.Minute).Return(nil)
			},
			id:        1,
			prefixKey: "test",
			exp:       time.Minute,
			fnCacheable: func(ctx context.Context, id int) (*TestData, error) {
				return testData, nil
			},
			want:    testData,
			wantErr: false,
		},
		{
			name: "Error - Cache get error",
			setupMock: func() {
				mockRepo.EXPECT().Get(ctx, "test:1").Return([]byte{}, false, errors.New("cache error"))
				mockRepo.EXPECT().Store(ctx, "test:1", testBytes, time.Minute).Return(nil)
			},
			id:        1,
			prefixKey: "test",
			exp:       time.Minute,
			fnCacheable: func(ctx context.Context, id int) (*TestData, error) {
				return testData, nil
			},
			want:    testData,
			wantErr: false,
		},
		{
			name: "Error - Source function error",
			setupMock: func() {
				mockRepo.EXPECT().Get(ctx, "test:1").Return([]byte{}, false, nil)
			},
			id:        1,
			prefixKey: "test",
			exp:       time.Minute,
			fnCacheable: func(ctx context.Context, id int) (*TestData, error) {
				return nil, errors.New("source error")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			got, err := GetFromCache(ctx, mockRepo, tt.id, tt.prefixKey, tt.exp, tt.fnCacheable)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetFromCacheWithDynamicTTL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockCacheRepo(ctrl)
	testData := &TestData{ID: 1, Name: "test"}
	testBytes, _ := json.Marshal(testData)

	tests := []struct {
		name        string
		setupMock   func()
		id          int
		prefixKey   string
		fnGetTtl    func(ctx context.Context, data *TestData) time.Duration
		fnCacheable func(ctx context.Context, id int) (*TestData, error)
		want        *TestData
		wantErr     bool
	}{
		{
			name: "Success - Data found in cache",
			setupMock: func() {
				mockRepo.EXPECT().Get(ctx, "test:1").Return(testBytes, true, nil)
			},
			id:        1,
			prefixKey: "test",
			fnGetTtl: func(ctx context.Context, data *TestData) time.Duration {
				return time.Minute
			},
			fnCacheable: func(ctx context.Context, id int) (*TestData, error) {
				return nil, nil // Should not be called
			},
			want:    testData,
			wantErr: false,
		},
		{
			name: "Success - Data not in cache, fetched from source with dynamic TTL",
			setupMock: func() {
				mockRepo.EXPECT().Get(ctx, "test:1").Return([]byte{}, false, nil)
				mockRepo.EXPECT().Store(ctx, "test:1", testBytes, time.Minute).Return(nil)
			},
			id:        1,
			prefixKey: "test",
			fnGetTtl: func(ctx context.Context, data *TestData) time.Duration {
				return time.Minute
			},
			fnCacheable: func(ctx context.Context, id int) (*TestData, error) {
				return testData, nil
			},
			want:    testData,
			wantErr: false,
		},
		{
			name: "Success - Store without TTL when TTL is 0",
			setupMock: func() {
				mockRepo.EXPECT().Get(ctx, "test:1").Return([]byte{}, false, nil)
				mockRepo.EXPECT().StoreWithoutTTL(ctx, "test:1", testBytes).Return(nil)
			},
			id:        1,
			prefixKey: "test",
			fnGetTtl: func(ctx context.Context, data *TestData) time.Duration {
				return 0
			},
			fnCacheable: func(ctx context.Context, id int) (*TestData, error) {
				return testData, nil
			},
			want:    testData,
			wantErr: false,
		},
		{
			name: "Error - Source function error",
			setupMock: func() {
				mockRepo.EXPECT().Get(ctx, "test:1").Return([]byte{}, false, nil)
			},
			id:        1,
			prefixKey: "test",
			fnGetTtl: func(ctx context.Context, data *TestData) time.Duration {
				return time.Minute
			},
			fnCacheable: func(ctx context.Context, id int) (*TestData, error) {
				return nil, errors.New("source error")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			got, err := GetFromCacheWithDynamicTTL(ctx, mockRepo, tt.id, tt.prefixKey, tt.fnGetTtl, tt.fnCacheable)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
