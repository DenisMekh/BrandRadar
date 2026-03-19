package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"prod-pobeda-2026/internal/entity"
)

// AlertRepo — PostgreSQL реализация хранилища сработавших алертов.
type AlertRepo struct {
	pool DB
	sb   squirrel.StatementBuilderType
}

// NewAlertRepo создаёт новый экземпляр репозитория алертов.
func NewAlertRepo(pool DB) *AlertRepo {
	return &AlertRepo{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *AlertRepo) Create(ctx context.Context, alert *entity.Alert) error {
	sql, args, err := r.sb.
		Insert("alerts").
		Columns("id", "config_id", "brand_id", "mentions_count", "window_start", "window_end", "fired_at").
		Values(alert.ID, alert.ConfigID, alert.BrandID, alert.MentionsCount, alert.WindowStart, alert.WindowEnd, alert.FiredAt).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("AlertRepo.Create: build query: %w", err)
	}

	if scanErr := r.pool.QueryRow(ctx, sql, args...).Scan(&alert.CreatedAt); scanErr != nil {
		return fmt.Errorf("AlertRepo.Create: %w", scanErr)
	}
	return nil
}

func (r *AlertRepo) ListByBrandID(ctx context.Context, brandID uuid.UUID, limit, offset int) ([]entity.Alert, int64, error) {
	countSQL, countArgs, err := r.sb.
		Select("COUNT(*)").
		From("alerts").
		Where(squirrel.Eq{"brand_id": brandID}).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("AlertRepo.ListByBrandID: build count query: %w", err)
	}

	var total int64
	if scanErr := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); scanErr != nil {
		return nil, 0, fmt.Errorf("AlertRepo.ListByBrandID: count: %w", scanErr)
	}

	dataBuilder := r.sb.
		Select("id", "config_id", "brand_id", "mentions_count", "window_start", "window_end", "fired_at", "created_at").
		From("alerts").
		Where(squirrel.Eq{"brand_id": brandID}).
		OrderBy("fired_at DESC")
	if limit > 0 {
		dataBuilder = dataBuilder.Limit(uint64(limit))
	}
	if offset > 0 {
		dataBuilder = dataBuilder.Offset(uint64(offset))
	}

	dataSQL, dataArgs, err := dataBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("AlertRepo.ListByBrandID: build data query: %w", err)
	}

	rows, err := r.pool.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("AlertRepo.ListByBrandID: query: %w", err)
	}
	defer rows.Close()

	alerts := make([]entity.Alert, 0)
	for rows.Next() {
		var alert entity.Alert
		if scanErr := rows.Scan(
			&alert.ID,
			&alert.ConfigID,
			&alert.BrandID,
			&alert.MentionsCount,
			&alert.WindowStart,
			&alert.WindowEnd,
			&alert.FiredAt,
			&alert.CreatedAt,
		); scanErr != nil {
			return nil, 0, fmt.Errorf("AlertRepo.ListByBrandID: scan: %w", scanErr)
		}
		alerts = append(alerts, alert)
	}
	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("AlertRepo.ListByBrandID: rows: %w", rows.Err())
	}

	return alerts, total, nil
}

// ListAll возвращает все алерты с пагинацией (без фильтра по бренду).
func (r *AlertRepo) ListAll(ctx context.Context, limit, offset int) ([]entity.Alert, int64, error) {
	countSQL, countArgs, err := r.sb.
		Select("COUNT(*)").
		From("alerts").
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("AlertRepo.ListAll: build count query: %w", err)
	}

	var total int64
	if scanErr := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); scanErr != nil {
		return nil, 0, fmt.Errorf("AlertRepo.ListAll: count: %w", scanErr)
	}

	dataBuilder := r.sb.
		Select("id", "config_id", "brand_id", "mentions_count", "window_start", "window_end", "fired_at", "created_at").
		From("alerts").
		OrderBy("fired_at DESC")
	if limit > 0 {
		dataBuilder = dataBuilder.Limit(uint64(limit))
	}
	if offset > 0 {
		dataBuilder = dataBuilder.Offset(uint64(offset))
	}

	dataSQL, dataArgs, err := dataBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("AlertRepo.ListAll: build data query: %w", err)
	}

	rows, err := r.pool.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("AlertRepo.ListAll: query: %w", err)
	}
	defer rows.Close()

	alerts := make([]entity.Alert, 0)
	for rows.Next() {
		var alert entity.Alert
		if scanErr := rows.Scan(
			&alert.ID,
			&alert.ConfigID,
			&alert.BrandID,
			&alert.MentionsCount,
			&alert.WindowStart,
			&alert.WindowEnd,
			&alert.FiredAt,
			&alert.CreatedAt,
		); scanErr != nil {
			return nil, 0, fmt.Errorf("AlertRepo.ListAll: scan: %w", scanErr)
		}
		alerts = append(alerts, alert)
	}
	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("AlertRepo.ListAll: rows: %w", rows.Err())
	}

	return alerts, total, nil
}

func (r *AlertRepo) GetLastFiredAt(ctx context.Context, configID uuid.UUID) (*time.Time, error) {
	sql, args, err := r.sb.
		Select("fired_at").
		From("alerts").
		Where(squirrel.Eq{"config_id": configID}).
		OrderBy("fired_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("AlertRepo.GetLastFiredAt: build query: %w", err)
	}

	var firedAt time.Time
	if scanErr := r.pool.QueryRow(ctx, sql, args...).Scan(&firedAt); scanErr != nil {
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return nil, fmt.Errorf("AlertRepo.GetLastFiredAt: %w", entity.ErrNotFound)
		}
		return nil, fmt.Errorf("AlertRepo.GetLastFiredAt: %w", scanErr)
	}
	return &firedAt, nil
}
