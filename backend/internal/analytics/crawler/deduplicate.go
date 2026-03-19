package crawler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/repo/redisrepo"
)

type Deduplicator interface {
	FilterDuplicates(ctx context.Context, items []Item) ([]Item, error)
}

type DeduplicateImpl struct {
	repo *redisrepo.Deduplicate
}

func NewDeduplicator(repo *redisrepo.Deduplicate) *DeduplicateImpl {
	return &DeduplicateImpl{repo: repo}
}

func generateHash(item Item) string {
	hash := sha256.Sum256([]byte(item.Text))
	return hex.EncodeToString(hash[:])
}

func (d *DeduplicateImpl) FilterDuplicates(ctx context.Context, items []Item) ([]Item, error) {
	unique := make([]Item, 0, len(items))
	cnt := 0
	for _, item := range items {
		hash := generateHash(item)
		key := fmt.Sprintf("%s:%s", redisrepo.DedupKeyPrefix, hash)

		exists, err := d.repo.Exists(ctx, key)
		if err != nil {
			logrus.WithError(err).Warn("deduplicator unavailable, returning all items without deduplication")
			return items, nil
		}

		if !exists {
			if err := d.repo.Set(ctx, key, "1"); err != nil {
				logrus.WithError(err).Warn("deduplicator set failed, item will not be cached")
			}

			cnt++
			unique = append(unique, item)
		}
	}
	logrus.Infof("deduplicated %d items", len(items)-cnt)

	return unique, nil
}
