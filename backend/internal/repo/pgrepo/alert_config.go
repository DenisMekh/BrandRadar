package pgrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"prod-pobeda-2026/internal/entity"
)

// AlertConfigRepo — PostgreSQL реализация хранилища настроек алертов.
type AlertConfigRepo struct {
	pool DB
	sb   squirrel.StatementBuilderType
}

// NewAlertConfigRepo создаёт новый экземпляр репозитория настроек алертов.
func NewAlertConfigRepo(pool DB) *AlertConfigRepo {
	return &AlertConfigRepo{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *AlertConfigRepo) Create(ctx context.Context, config *entity.AlertConfig) error {
	sql, args, err := r.sb.
		Insert("alert_configs").
		Columns(
			"id", "brand_id", "window_minutes",
			"cooldown_minutes", "sentiment_filter", "enabled",
			"percentile", "anomaly_window_size",
		).
		Values(
			config.ID, config.BrandID, config.WindowMinutes,
			config.CooldownMinutes, config.SentimentFilter, config.Enabled,
			config.Percentile, config.AnomalyWindowSize,
		).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("AlertConfigRepo.Create: build query: %w", err)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("AlertConfigRepo.Create: %w", entity.ErrDuplicate)
		}
		return fmt.Errorf("AlertConfigRepo.Create: %w", err)
	}
	return nil
}

func (r *AlertConfigRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AlertConfig, error) {
	sql, args, err := r.sb.
		Select(
			"id", "brand_id", "window_minutes",
			"cooldown_minutes", "sentiment_filter", "enabled",
			"percentile", "anomaly_window_size", "created_at", "updated_at",
		).
		From("alert_configs").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("AlertConfigRepo.GetByID: build query: %w", err)
	}

	config, err := scanAlertConfig(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("AlertConfigRepo.GetByID: %w", entity.ErrNotFound)
		}
		return nil, fmt.Errorf("AlertConfigRepo.GetByID: %w", err)
	}
	return config, nil
}

func (r *AlertConfigRepo) GetByBrandID(ctx context.Context, brandID uuid.UUID) (*entity.AlertConfig, error) {
	sql, args, err := r.sb.
		Select(
			"id", "brand_id", "window_minutes",
			"cooldown_minutes", "sentiment_filter", "enabled",
			"percentile", "anomaly_window_size", "created_at", "updated_at",
		).
		From("alert_configs").
		Where(squirrel.Eq{"brand_id": brandID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("AlertConfigRepo.GetByBrandID: build query: %w", err)
	}

	config, err := scanAlertConfig(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("AlertConfigRepo.GetByBrandID: %w", entity.ErrNotFound)
		}
		return nil, fmt.Errorf("AlertConfigRepo.GetByBrandID: %w", err)
	}
	return config, nil
}

func (r *AlertConfigRepo) Update(ctx context.Context, config *entity.AlertConfig) error {
	sql, args, err := r.sb.
		Update("alert_configs").
		Set("window_minutes", config.WindowMinutes).
		Set("cooldown_minutes", config.CooldownMinutes).
		Set("sentiment_filter", config.SentimentFilter).
		Set("enabled", config.Enabled).
		Set("percentile", config.Percentile).
		Set("anomaly_window_size", config.AnomalyWindowSize).
		Set("updated_at", squirrel.Expr("now()")).
		Where(squirrel.Eq{"id": config.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("AlertConfigRepo.Update: build query: %w", err)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&config.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("AlertConfigRepo.Update: %w", entity.ErrNotFound)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("AlertConfigRepo.Update: %w", entity.ErrDuplicate)
		}
		return fmt.Errorf("AlertConfigRepo.Update: %w", err)
	}
	return nil
}

// GetAllActive возвращает все активные конфигурации алертов.
func (r *AlertConfigRepo) GetAllActive(ctx context.Context) ([]*entity.AlertConfig, error) {
	sql, args, err := r.sb.
		Select(
			"id", "brand_id", "window_minutes",
			"cooldown_minutes", "sentiment_filter", "enabled",
			"percentile", "anomaly_window_size", "created_at", "updated_at",
		).
		From("alert_configs").
		Where(squirrel.Eq{"enabled": true}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("AlertConfigRepo.GetAllActive: build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("AlertConfigRepo.GetAllActive: query: %w", err)
	}
	defer rows.Close()

	var configs []*entity.AlertConfig
	for rows.Next() {
		config, err := scanAlertConfig(rows)
		if err != nil {
			return nil, fmt.Errorf("AlertConfigRepo.GetAllActive: scan: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// ListAll возвращает все конфигурации алертов (включая отключённые).
func (r *AlertConfigRepo) ListAll(ctx context.Context) ([]*entity.AlertConfig, error) {
	sql, args, err := r.sb.
		Select(
			"id", "brand_id", "window_minutes",
			"cooldown_minutes", "sentiment_filter", "enabled",
			"percentile", "anomaly_window_size", "created_at", "updated_at",
		).
		From("alert_configs").
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("AlertConfigRepo.ListAll: build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("AlertConfigRepo.ListAll: query: %w", err)
	}
	defer rows.Close()

	var configs []*entity.AlertConfig
	for rows.Next() {
		config, err := scanAlertConfig(rows)
		if err != nil {
			return nil, fmt.Errorf("AlertConfigRepo.ListAll: scan: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

func (r *AlertConfigRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.sb.
		Delete("alert_configs").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("AlertConfigRepo.Delete: build query: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("AlertConfigRepo.Delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("AlertConfigRepo.Delete: %w", entity.ErrNotFound)
	}
	return nil
}

type alertConfigScanner interface {
	Scan(dest ...any) error
}

func scanAlertConfig(row alertConfigScanner) (*entity.AlertConfig, error) {
	var config entity.AlertConfig
	err := row.Scan(
		&config.ID,
		&config.BrandID,
		&config.WindowMinutes,
		&config.CooldownMinutes,
		&config.SentimentFilter,
		&config.Enabled,
		&config.Percentile,
		&config.AnomalyWindowSize,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
