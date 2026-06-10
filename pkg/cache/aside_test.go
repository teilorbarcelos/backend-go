package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestCacheAside(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	ctx := context.Background()
	cache := NewAside(client, 100*time.Millisecond)

	t.Run("miss then cache hit", func(t *testing.T) {
		callCount := 0
		data, err := cache.GetBytes(ctx, "test:1", func() ([]byte, error) {
			callCount++
			return []byte(`{"id":"1","name":"test"}`), nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)

		data2, err := cache.GetBytes(ctx, "test:1", func() ([]byte, error) {
			callCount++
			return []byte(`{"id":"1","name":"should not be called"}`), nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, data, data2)
	})

	t.Run("fetch error propagates", func(t *testing.T) {
		_, err := cache.GetBytes(ctx, "test:error", func() ([]byte, error) {
			return nil, errors.New("fetch error")
		})
		assert.Error(t, err)
	})

	t.Run("invalidate clears cache", func(t *testing.T) {
		callCount := 0
		cache.GetBytes(ctx, "test:inv", func() ([]byte, error) {
			callCount++
			return []byte(`"value1"`), nil
		})
		assert.Equal(t, 1, callCount)

		cache.Invalidate(ctx, "test:inv")

		cache.GetBytes(ctx, "test:inv", func() ([]byte, error) {
			callCount++
			return []byte(`"value2"`), nil
		})
		assert.Equal(t, 2, callCount)
	})
}
