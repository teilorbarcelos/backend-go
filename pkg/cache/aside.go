package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type Aside struct {
	client *redis.Client
	group  singleflight.Group
	ttl    time.Duration
}

func NewAside(client *redis.Client, ttl time.Duration) *Aside {
	return &Aside{client: client, ttl: ttl}
}

func (c *Aside) GetBytes(ctx context.Context, key string, fetch func() ([]byte, error)) ([]byte, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == nil {
		return data, nil
	}

	result, err, _ := c.group.Do(key, func() (interface{}, error) {
		val, err := fetch()
		if err != nil {
			return nil, err
		}
		c.client.Set(ctx, key, val, c.ttl)
		return val, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]byte), nil
}

func (c *Aside) Invalidate(ctx context.Context, key string) {
	c.group.Forget(key)
	c.client.Del(ctx, key)
}
