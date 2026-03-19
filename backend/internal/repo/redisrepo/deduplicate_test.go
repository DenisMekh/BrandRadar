package redisrepo

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestDeduplicate_SetAndExists(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewDeduplicate(client)
	ctx := context.Background()
	key := DedupKeyPrefix + ":test"

	exists, err := repo.Exists(ctx, key)
	require.NoError(t, err)
	require.False(t, exists)

	require.NoError(t, repo.Set(ctx, key, "1"))

	exists, err = repo.Exists(ctx, key)
	require.NoError(t, err)
	require.True(t, exists)
}
