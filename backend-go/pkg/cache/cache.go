package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) (bool, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	Ping(ctx context.Context) error
	Close() error
}

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int, url string) Cache {
	var client *redis.Client
	if url != "" {
		opt, err := redis.ParseURL(url)
		if err == nil {
			client = redis.NewClient(opt)
		}
	}
	if client == nil {
		client = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️ [Redis] Unable to connect to Redis at %s: %v (Cache operations will be no-ops)", addr, err)
	} else {
		log.Printf("⚡ [Redis] Connected successfully to Redis at %s (DB %d)", addr, db)
	}

	return &redisCache{client: client}
}

func (r *redisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	if r.client == nil {
		return false, nil
	}

	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	if err := json.Unmarshal(val, dest); err != nil {
		return false, err
	}

	return true, nil
}

func (r *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if r.client == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

func (r *redisCache) Delete(ctx context.Context, keys ...string) error {
	if r.client == nil || len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *redisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	if r.client == nil || pattern == "" {
		return nil
	}

	var cursor uint64
	var allKeys []string
	for {
		var keys []string
		var err error
		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		allKeys = append(allKeys, keys...)
		if cursor == 0 {
			break
		}
	}

	if len(allKeys) > 0 {
		return r.client.Del(ctx, allKeys...).Err()
	}
	return nil
}

func (r *redisCache) Ping(ctx context.Context) error {
	if r.client == nil {
		return errors.New("redis client not initialized")
	}
	return r.client.Ping(ctx).Err()
}

func (r *redisCache) Close() error {
	if r.client == nil {
		return nil
	}
	return r.client.Close()
}
