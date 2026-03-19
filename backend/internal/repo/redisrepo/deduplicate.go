package redisrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	DedupKeyPrefix = "brandradar:crawler:duplicate"
	dedupTTL       = 24 * time.Hour
)

type Deduplicate struct {
	client *redis.Client
}

func NewDeduplicate(client *redis.Client) *Deduplicate {
	return &Deduplicate{client: client}
}

func (d *Deduplicate) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := d.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("Deduplicate.Exists: %w", err)
	}
	return exists > 0, nil
}

func (d *Deduplicate) Set(ctx context.Context, key, value string) error {
	err := d.client.Set(ctx, key, value, dedupTTL).Err()
	if err != nil {
		return fmt.Errorf("Deduplicate.Set: %w", err)
	}
	return nil
}
