package pgrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/entity"
)

type AnalyticsRepo struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewAnalyticsRepo(pool *pgxpool.Pool) *AnalyticsRepo {
	return &AnalyticsRepo{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// BatchInsertItems сохраняет items и sentiment results в БД одной транзакцией.
func (r *AnalyticsRepo) BatchInsertItems(
	ctx context.Context,
	items []entity.CrawlerItem,
	results []entity.SentimentMLResult,
) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			return
		}
	}(tx, ctx)

	// Batch insert items
	itemsBuilder := r.sb.Insert("crawler_items").
		Columns("id", "text", "link", "published_at", "created_at", "source_id").
		Suffix("ON CONFLICT (id) DO NOTHING")

	now := time.Now().UTC()
	for _, item := range items {
		logrus.Infof("inserting item id=%s link=%s", item.ID, item.Link)
		itemsBuilder = itemsBuilder.Values(item.ID, item.Text, item.Link, item.PublishedAt, now, item.SourceID)
	}

	itemsSQL, itemsArgs, err := itemsBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build items query: %w", err)
	}

	if _, err := tx.Exec(ctx, itemsSQL, itemsArgs...); err != nil {
		return fmt.Errorf("insert items: %w", err)
	}

	// Batch insert sentiment results
	if len(results) > 0 {
		resultsBuilder := r.sb.Insert("sentiment_results").
			Columns("id", "item_id", "brand_id", "sentiment", "confidence", "created_at")

		for _, result := range results {
			resultsBuilder = resultsBuilder.Values(uuid.New(), result.ItemID, result.BrandID, result.Sentiment, result.Confidence, now)
		}

		resultsSQL, resultsArgs, err := resultsBuilder.ToSql()
		if err != nil {
			return fmt.Errorf("build results query: %w", err)
		}

		if _, err := tx.Exec(ctx, resultsSQL, resultsArgs...); err != nil {
			return fmt.Errorf("insert sentiment results: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
