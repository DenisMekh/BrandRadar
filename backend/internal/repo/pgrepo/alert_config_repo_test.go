package pgrepo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"

	"prod-pobeda-2026/internal/entity"
)

func TestAlertConfigRepo_CRUD(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewAlertConfigRepo(mock)
	ctx := context.Background()
	now := time.Now().UTC()
	id := uuid.New()
	brandID := uuid.New()

	cfg := &entity.AlertConfig{
		ID:                id,
		BrandID:           brandID,
		WindowMinutes:     30,
		CooldownMinutes:   15,
		SentimentFilter:   "negative",
		Enabled:           true,
		Percentile:        95.0,
		AnomalyWindowSize: 10,
	}

	mock.ExpectQuery("INSERT INTO alert_configs").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), cfg.WindowMinutes, cfg.CooldownMinutes, cfg.SentimentFilter, cfg.Enabled, cfg.Percentile, cfg.AnomalyWindowSize).
		WillReturnRows(pgxmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))
	require.NoError(t, repo.Create(ctx, cfg))

	row := pgxmock.NewRows([]string{
		"id", "brand_id", "window_minutes", "cooldown_minutes", "sentiment_filter", "enabled", "percentile", "anomaly_window_size", "created_at", "updated_at",
	}).AddRow(id, brandID, 30, 15, "negative", true, 95.0, 10, now, now)

	mock.ExpectQuery("SELECT id, brand_id, window_minutes, cooldown_minutes, sentiment_filter, enabled, percentile, anomaly_window_size, created_at, updated_at FROM alert_configs").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(row)
	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, id, got.ID)

	mock.ExpectQuery("SELECT id, brand_id, window_minutes, cooldown_minutes, sentiment_filter, enabled, percentile, anomaly_window_size, created_at, updated_at FROM alert_configs").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "brand_id", "window_minutes", "cooldown_minutes", "sentiment_filter", "enabled", "percentile", "anomaly_window_size", "created_at", "updated_at",
		}).AddRow(id, brandID, 30, 15, "negative", true, 95.0, 10, now, now))
	_, err = repo.GetByBrandID(ctx, brandID)
	require.NoError(t, err)

	mock.ExpectQuery("UPDATE alert_configs").
		WithArgs(cfg.WindowMinutes, cfg.CooldownMinutes, cfg.SentimentFilter, cfg.Enabled, cfg.Percentile, cfg.AnomalyWindowSize, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"updated_at"}).AddRow(now.Add(time.Minute)))
	require.NoError(t, repo.Update(ctx, cfg))

	mock.ExpectExec("DELETE FROM alert_configs").
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	require.NoError(t, repo.Delete(ctx, id))

	require.NoError(t, mock.ExpectationsWereMet())
}
