package pgrepo

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/entity"
)

type ClusteringRepo struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewClusteringRepo(pool *pgxpool.Pool) *ClusteringRepo {
	return &ClusteringRepo{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

const clusterItemsLimit = 1000

// GetItemsByBrandID возвращает последние clusterItemsLimit crawler_items бренда по дате публикации.
func (r *ClusteringRepo) GetItemsByBrandID(ctx context.Context, brandID uuid.UUID) ([]entity.CrawlerItemText, error) {
	sql, args, err := r.sb.
		Select("ci.id", "ci.text").
		From("crawler_items ci").
		Join("sentiment_results sr ON sr.item_id = ci.id").
		Where(squirrel.Eq{"sr.brand_id": brandID}).
		GroupBy("ci.id", "ci.text", "ci.published_at").
		OrderBy("ci.published_at DESC").
		Limit(clusterItemsLimit).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query items by brand: %w", err)
	}
	defer rows.Close()

	var items []entity.CrawlerItemText
	for rows.Next() {
		var item entity.CrawlerItemText
		if err := rows.Scan(&item.ID, &item.Text); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// UpdateClusterLabels проставляет cluster_label в sentiment_results для каждого item бренда.
// nil означает noise — cluster_label обнуляется (NULL).
func (r *ClusteringRepo) UpdateClusterLabels(ctx context.Context, brandID uuid.UUID, assignments map[uuid.UUID]*int) error {
	if len(assignments) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for itemID, label := range assignments {
		batch.Queue(
			`UPDATE sentiment_results SET cluster_label = $1 WHERE item_id = $2 AND brand_id = $3`,
			label, itemID, brandID,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer func() {
		if err := br.Close(); err != nil {
			logrus.Warnf("batch close: %v", err)
		}
	}()

	for range assignments {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch update cluster label: %w", err)
		}
	}

	return nil
}
