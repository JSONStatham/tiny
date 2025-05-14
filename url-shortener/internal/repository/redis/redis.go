package redis

import (
	"context"
	"encoding/json"
	"time"
	"urlshortener/internal/config"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func New(cfg *config.Config) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Cache.Addr,
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.Db,
	})

	return &Cache{client: client}
}

func (c *Cache) Close() {
	c.client.Close()
}

func (c *Cache) Get(ctx context.Context, key string, target any) error {
	bytes, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil
	}

	return json.Unmarshal(bytes, &target)
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, bytes, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
