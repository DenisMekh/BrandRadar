package redisrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// CooldownCache — Redis реализация кэша cooldown для алертов.
type CooldownCache struct {
	client *redis.Client
}

// NewCooldownCache создаёт новый экземпляр кэша cooldown.
func NewCooldownCache(client *redis.Client) *CooldownCache {
	return &CooldownCache{client: client}
}

// TryLock пытается установить cooldown для конфига алерта.
// Возвращает true, если cooldown свободен и был успешно установлен (атомарная операция SET NX).
func (c *CooldownCache) TryLock(ctx context.Context, configID uuid.UUID, cooldownMinutes int) (bool, error) {
	key := fmt.Sprintf("brandradar:cooldown:%s", configID.String())
	ttl := time.Duration(cooldownMinutes) * time.Minute

	res, err := c.client.SetArgs(ctx, key, "locked", redis.SetArgs{
		TTL:  ttl,
		Mode: "NX",
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("CooldownCache.TryLock: %w", err)
	}
	return res == "OK", nil
}
