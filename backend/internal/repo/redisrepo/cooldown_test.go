package redisrepo

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestCooldownCacheTryLock(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	cache := NewCooldownCache(client)
	ctx := context.Background()
	configID := uuid.New()

	locked, err := cache.TryLock(ctx, configID, 1)
	require.NoError(t, err)
	require.True(t, locked)

	locked, err = cache.TryLock(ctx, configID, 1)
	require.NoError(t, err)
	require.False(t, locked)
}
