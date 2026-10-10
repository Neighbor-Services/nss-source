package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedisCache_NilClientGracefulHandling(t *testing.T) {
	c := &redisCache{client: nil}

	ctx := context.Background()

	var dest string
	found, err := c.Get(ctx, "key", &dest)
	assert.False(t, found)
	assert.NoError(t, err)

	err = c.Set(ctx, "key", "value", time.Minute)
	assert.NoError(t, err)

	err = c.Delete(ctx, "key")
	assert.NoError(t, err)

	err = c.DeleteByPattern(ctx, "key*")
	assert.NoError(t, err)

	locked, err := c.AcquireLock(ctx, "lock", time.Minute)
	assert.True(t, locked)
	assert.NoError(t, err)

	err = c.ReleaseLock(ctx, "lock")
	assert.NoError(t, err)

	err = c.Publish(ctx, "chan", "msg")
	assert.NoError(t, err)

	pubsub := c.Subscribe(ctx, "chan")
	assert.Nil(t, pubsub)

	assert.Nil(t, c.GetClient())

	err = c.Ping(ctx)
	assert.Error(t, err)

	err = c.Close()
	assert.NoError(t, err)
}

func TestNewRedisCache_OfflineInitialization(t *testing.T) {
	// Should initialize without crashing even if redis server is unreachable
	c := NewRedisCache("127.0.0.1:9999", "", 0, "")
	assert.NotNil(t, c)
	assert.NotNil(t, c.GetClient())
	_ = c.Close()
}
