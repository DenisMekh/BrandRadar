package crawler

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"prod-pobeda-2026/internal/repo/redisrepo"
)

func TestNewDeduplicator(t *testing.T) {
	t.Parallel()

	repo := redisrepo.NewDeduplicate(nil)
	d := NewDeduplicator(repo)
	require.NotNil(t, d)
}

func TestGenerateHash(t *testing.T) {
	t.Parallel()

	sameA := Item{Text: "text", Link: "https://example.com/1"}
	diff := Item{Text: "different", Link: "https://example.com/1"}

	hashA := generateHash(sameA)
	hashDiff := generateHash(diff)

	require.NotEqual(t, hashA, hashDiff)
}

func TestFilterDuplicates(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := redisrepo.NewDeduplicate(client)
	d := NewDeduplicator(repo)

	items := []Item{
		{Text: "same", Link: "https://example.com/a"},
		{Text: "new", Link: "https://example.com/b"},
	}

	got, err := d.FilterDuplicates(context.Background(), items)
	require.NoError(t, err)
	require.Len(t, got, 2)

	gotSecond, err := d.FilterDuplicates(context.Background(), items)
	require.NoError(t, err)
	require.Len(t, gotSecond, 0)
}
