package redisrepo

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"prod-pobeda-2026/internal/entity"
)

func TestNewMentionCache_AndKeyHelpers(t *testing.T) {
	t.Parallel()

	cache := NewMentionCache(nil)
	require.NotNil(t, cache)

	brandID := uuid.New()
	filterA := entity.MentionFilter{BrandID: brandID, Limit: 10, Offset: 5}
	filterB := entity.MentionFilter{BrandID: brandID, Limit: 20, Offset: 0}

	keyA1, err := mentionListKey(filterA)
	require.NoError(t, err)
	keyA2, err := mentionListKey(filterA)
	require.NoError(t, err)
	keyB, err := mentionListKey(filterB)
	require.NoError(t, err)

	require.Equal(t, keyA1, keyA2)
	require.NotEqual(t, keyA1, keyB)
	require.Contains(t, keyA1, mentionListKeyPrefix)
	require.Contains(t, mentionBrandSetKey(brandID), brandID.String())
}

func TestMentionCache_SetGetInvalidate(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	cache := NewMentionCache(client)
	ctx := context.Background()
	brandID := uuid.New()

	filter := entity.MentionFilter{
		BrandID: brandID,
		Limit:   10,
	}
	items := []entity.Mention{
		{ID: uuid.New(), BrandID: brandID, Text: "one"},
		{ID: uuid.New(), BrandID: brandID, Text: "two"},
	}

	gotItems, gotTotal, err := cache.GetList(ctx, filter)
	require.NoError(t, err)
	require.Nil(t, gotItems)
	require.EqualValues(t, 0, gotTotal)

	require.NoError(t, cache.SetList(ctx, filter, items, 2))

	gotItems, gotTotal, err = cache.GetList(ctx, filter)
	require.NoError(t, err)
	require.Len(t, gotItems, 2)
	require.EqualValues(t, 2, gotTotal)

	require.NoError(t, cache.InvalidateByBrand(ctx, brandID))
	gotItems, gotTotal, err = cache.GetList(ctx, filter)
	require.NoError(t, err)
	require.Nil(t, gotItems)
	require.EqualValues(t, 0, gotTotal)

	require.NoError(t, cache.SetList(ctx, filter, items, 2))
	require.NoError(t, cache.InvalidateAll(ctx))
	gotItems, gotTotal, err = cache.GetList(ctx, filter)
	require.NoError(t, err)
	require.Nil(t, gotItems)
	require.EqualValues(t, 0, gotTotal)
}
