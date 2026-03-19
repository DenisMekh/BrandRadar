package redisrepo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"prod-pobeda-2026/internal/entity"
)

const (
	mentionListKeyPrefix = "brandradar:mentions:list:"
	mentionListKeyMask   = "brandradar:mentions:list:*"
	mentionListTTL       = 30 * time.Second
	mentionBrandSetTTL   = 60 * time.Second
)

type mentionListCachePayload struct {
	Items []entity.Mention `json:"items"`
	Total int64            `json:"total"`
}

// MentionCache — Redis реализация кэша списка упоминаний.
type MentionCache struct {
	client *redis.Client
}

// NewMentionCache создаёт новый экземпляр кэша списков упоминаний.
func NewMentionCache(client *redis.Client) *MentionCache {
	return &MentionCache{client: client}
}

// GetList читает список упоминаний из Redis по фильтру.
// При cache miss возвращает (nil, 0, nil).
func (c *MentionCache) GetList(ctx context.Context, filter entity.MentionFilter) ([]entity.Mention, int64, error) {
	key, err := mentionListKey(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("MentionCache.GetList: %w", err)
	}

	raw, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("MentionCache.GetList: %w", err)
	}

	var payload mentionListCachePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, 0, fmt.Errorf("MentionCache.GetList: %w", err)
	}

	return payload.Items, payload.Total, nil
}

// SetList сохраняет список упоминаний в Redis и привязывает ключ к brand-set для инвалидации.
func (c *MentionCache) SetList(ctx context.Context, filter entity.MentionFilter, items []entity.Mention, total int64) error {
	key, err := mentionListKey(filter)
	if err != nil {
		return fmt.Errorf("MentionCache.SetList: %w", err)
	}

	payloadBytes, err := json.Marshal(mentionListCachePayload{
		Items: items,
		Total: total,
	})
	if err != nil {
		return fmt.Errorf("MentionCache.SetList: %w", err)
	}

	brandSetKey := mentionBrandSetKey(filter.BrandID)
	pipe := c.client.Pipeline()
	pipe.Set(ctx, key, payloadBytes, mentionListTTL)
	pipe.SAdd(ctx, brandSetKey, key)
	pipe.Expire(ctx, brandSetKey, mentionBrandSetTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("MentionCache.SetList: %w", err)
	}

	return nil
}

// InvalidateByBrand очищает все кэш-ключи списка для конкретного бренда.
func (c *MentionCache) InvalidateByBrand(ctx context.Context, brandID uuid.UUID) error {
	brandSetKey := mentionBrandSetKey(brandID)
	keys, err := c.client.SMembers(ctx, brandSetKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("MentionCache.InvalidateByBrand: %w", err)
	}

	pipe := c.client.Pipeline()
	if len(keys) > 0 {
		pipe.Del(ctx, keys...)
	}
	pipe.Del(ctx, brandSetKey)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("MentionCache.InvalidateByBrand: %w", err)
	}

	return nil
}

// InvalidateAll очищает все кэш-ключи списков упоминаний.
func (c *MentionCache) InvalidateAll(ctx context.Context) error {
	var cursor uint64

	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, mentionListKeyMask, 100).Result()
		if err != nil {
			return fmt.Errorf("MentionCache.InvalidateAll: %w", err)
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("MentionCache.InvalidateAll: %w", err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func mentionListKey(filter entity.MentionFilter) (string, error) {
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return "", fmt.Errorf("marshal filter: %w", err)
	}

	hash := sha256.Sum256(filterJSON)
	hashHex := hex.EncodeToString(hash[:])

	return mentionListKeyPrefix + hashHex[:16], nil
}

func mentionBrandSetKey(brandID uuid.UUID) string {
	return fmt.Sprintf("brandradar:mentions:brand:%s", brandID.String())
}
