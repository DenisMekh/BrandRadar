package pgrepo

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"prod-pobeda-2026/internal/entity"
)

// DashboardRepo — PostgreSQL-реализация DashboardRepository.
type DashboardRepo struct {
	pool *pgxpool.Pool
	sb   sq.StatementBuilderType
}

// NewDashboardRepo создаёт DashboardRepo.
func NewDashboardRepo(pool *pgxpool.Pool) *DashboardRepo {
	return &DashboardRepo{
		pool: pool,
		sb:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *DashboardRepo) applyDateFilter(b sq.SelectBuilder, dateFrom, dateTo string) sq.SelectBuilder {
	if dateFrom != "" {
		b = b.Where(sq.GtOrEq{"ci.published_at": dateFrom})
	}
	if dateTo != "" {
		b = b.Where(sq.LtOrEq{"ci.published_at": dateTo + " 23:59:59"})
	}
	return b
}

func (r *DashboardRepo) GetBrandSentiment(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) (entity.SentimentCounts, error) {
	b := r.sb.Select(
		"COALESCE(SUM(CASE WHEN sr.sentiment = 'positive' THEN 1 ELSE 0 END), 0) AS positive",
		"COALESCE(SUM(CASE WHEN sr.sentiment = 'negative' THEN 1 ELSE 0 END), 0) AS negative",
		"COALESCE(SUM(CASE WHEN sr.sentiment NOT IN ('positive','negative') THEN 1 ELSE 0 END), 0) AS neutral",
	).From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id").
		Where(sq.Eq{"sr.brand_id": brandID})
	b = r.applyDateFilter(b, dateFrom, dateTo)

	sql, args, err := b.ToSql()
	if err != nil {
		return entity.SentimentCounts{}, fmt.Errorf("DashboardRepo.GetBrandSentiment: build: %w", err)
	}

	var s entity.SentimentCounts
	if err = r.pool.QueryRow(ctx, sql, args...).Scan(&s.Positive, &s.Negative, &s.Neutral); err != nil {
		return entity.SentimentCounts{}, fmt.Errorf("DashboardRepo.GetBrandSentiment: scan: %w", err)
	}
	return s, nil
}

func (r *DashboardRepo) GetBrandSourceStats(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) ([]entity.SourceCount, error) {
	b := r.sb.Select(
		"COALESCE(s.name, 'unknown') AS source_name",
		"COUNT(*) AS cnt",
	).From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id").
		LeftJoin("sources s ON s.id = ci.source_id").
		Where(sq.Eq{"sr.brand_id": brandID}).
		GroupBy("s.name").
		OrderBy("cnt DESC")
	b = r.applyDateFilter(b, dateFrom, dateTo)

	sql, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("DashboardRepo.GetBrandSourceStats: build: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("DashboardRepo.GetBrandSourceStats: query: %w", err)
	}
	defer rows.Close()

	var result []entity.SourceCount
	for rows.Next() {
		var sc entity.SourceCount
		if err = rows.Scan(&sc.Source, &sc.Count); err != nil {
			return nil, fmt.Errorf("DashboardRepo.GetBrandSourceStats: scan: %w", err)
		}
		result = append(result, sc)
	}
	return result, rows.Err()
}

func (r *DashboardRepo) GetBrandDailyStats(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) ([]entity.DailyCount, error) {
	b := r.sb.Select(
		"TO_CHAR(ci.published_at, 'YYYY-MM-DD') AS day",
		"COALESCE(SUM(CASE WHEN sr.sentiment = 'positive' THEN 1 ELSE 0 END), 0)",
		"COALESCE(SUM(CASE WHEN sr.sentiment = 'negative' THEN 1 ELSE 0 END), 0)",
		"COALESCE(SUM(CASE WHEN sr.sentiment NOT IN ('positive','negative') THEN 1 ELSE 0 END), 0)",
	).From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id").
		Where(sq.Eq{"sr.brand_id": brandID}).
		GroupBy("day").
		OrderBy("day ASC")
	b = r.applyDateFilter(b, dateFrom, dateTo)

	sql, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("DashboardRepo.GetBrandDailyStats: build: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("DashboardRepo.GetBrandDailyStats: query: %w", err)
	}
	defer rows.Close()

	var result []entity.DailyCount
	for rows.Next() {
		var d entity.DailyCount
		if err = rows.Scan(&d.Date, &d.Positive, &d.Negative, &d.Neutral); err != nil {
			return nil, fmt.Errorf("DashboardRepo.GetBrandDailyStats: scan: %w", err)
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

func (r *DashboardRepo) GetBrandAlertCount(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) (int64, error) {
	b := r.sb.Select("COUNT(*)").From("alerts").Where(sq.Eq{"brand_id": brandID})
	if dateFrom != "" {
		b = b.Where(sq.GtOrEq{"fired_at": dateFrom})
	}
	if dateTo != "" {
		b = b.Where(sq.LtOrEq{"fired_at": dateTo + " 23:59:59"})
	}

	sql, args, err := b.ToSql()
	if err != nil {
		return 0, fmt.Errorf("DashboardRepo.GetBrandAlertCount: build: %w", err)
	}

	var count int64
	if err = r.pool.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("DashboardRepo.GetBrandAlertCount: scan: %w", err)
	}
	return count, nil
}

func (r *DashboardRepo) GetOverallSentiment(ctx context.Context, dateFrom, dateTo string) (entity.SentimentCounts, error) {
	b := r.sb.Select(
		"COALESCE(SUM(CASE WHEN sr.sentiment = 'positive' THEN 1 ELSE 0 END), 0)",
		"COALESCE(SUM(CASE WHEN sr.sentiment = 'negative' THEN 1 ELSE 0 END), 0)",
		"COALESCE(SUM(CASE WHEN sr.sentiment NOT IN ('positive','negative') THEN 1 ELSE 0 END), 0)",
	).From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id")
	b = r.applyDateFilter(b, dateFrom, dateTo)

	sql, args, err := b.ToSql()
	if err != nil {
		return entity.SentimentCounts{}, fmt.Errorf("DashboardRepo.GetOverallSentiment: build: %w", err)
	}

	var s entity.SentimentCounts
	if err = r.pool.QueryRow(ctx, sql, args...).Scan(&s.Positive, &s.Negative, &s.Neutral); err != nil {
		return entity.SentimentCounts{}, fmt.Errorf("DashboardRepo.GetOverallSentiment: scan: %w", err)
	}
	return s, nil
}

func (r *DashboardRepo) GetOverallDailyStats(ctx context.Context, dateFrom, dateTo string) ([]entity.DailyCount, error) {
	b := r.sb.Select(
		"TO_CHAR(ci.published_at, 'YYYY-MM-DD') AS day",
		"COALESCE(SUM(CASE WHEN sr.sentiment = 'positive' THEN 1 ELSE 0 END), 0)",
		"COALESCE(SUM(CASE WHEN sr.sentiment = 'negative' THEN 1 ELSE 0 END), 0)",
		"COALESCE(SUM(CASE WHEN sr.sentiment NOT IN ('positive','negative') THEN 1 ELSE 0 END), 0)",
	).From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id").
		GroupBy("day").
		OrderBy("day ASC")
	b = r.applyDateFilter(b, dateFrom, dateTo)

	sql, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("DashboardRepo.GetOverallDailyStats: build: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("DashboardRepo.GetOverallDailyStats: query: %w", err)
	}
	defer rows.Close()

	var result []entity.DailyCount
	for rows.Next() {
		var d entity.DailyCount
		if err = rows.Scan(&d.Date, &d.Positive, &d.Negative, &d.Neutral); err != nil {
			return nil, fmt.Errorf("DashboardRepo.GetOverallDailyStats: scan: %w", err)
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

func (r *DashboardRepo) GetAllBrandsSummary(ctx context.Context, dateFrom, dateTo string) ([]entity.BrandSummary, error) {
	b := r.sb.Select(
		"b.id",
		"b.name",
		"COALESCE(SUM(CASE WHEN sr.sentiment = 'positive' THEN 1 ELSE 0 END), 0)",
		"COALESCE(SUM(CASE WHEN sr.sentiment = 'negative' THEN 1 ELSE 0 END), 0)",
		"COALESCE(SUM(CASE WHEN sr.sentiment NOT IN ('positive','negative') THEN 1 ELSE 0 END), 0)",
	).From("brands b").
		LeftJoin("sentiment_results sr ON sr.brand_id = b.id").
		LeftJoin("crawler_items ci ON ci.id = sr.item_id").
		GroupBy("b.id", "b.name").
		OrderBy("b.name ASC")

	if dateFrom != "" {
		b = b.Where(sq.Or{sq.GtOrEq{"ci.published_at": dateFrom}, sq.Eq{"ci.published_at": nil}})
	}
	if dateTo != "" {
		b = b.Where(sq.Or{sq.LtOrEq{"ci.published_at": dateTo + " 23:59:59"}, sq.Eq{"ci.published_at": nil}})
	}

	sql, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("DashboardRepo.GetAllBrandsSummary: build: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("DashboardRepo.GetAllBrandsSummary: query: %w", err)
	}
	defer rows.Close()

	var result []entity.BrandSummary
	for rows.Next() {
		var bs entity.BrandSummary
		if err = rows.Scan(&bs.BrandID, &bs.BrandName, &bs.Sentiment.Positive, &bs.Sentiment.Negative, &bs.Sentiment.Neutral); err != nil {
			return nil, fmt.Errorf("DashboardRepo.GetAllBrandsSummary: scan: %w", err)
		}
		result = append(result, bs)
	}
	return result, rows.Err()
}
