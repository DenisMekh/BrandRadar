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

func TestAlertRepo_CreateAndListByBrandID(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewAlertRepo(mock)
	ctx := context.Background()
	now := time.Now().UTC()
	brandID := uuid.New()
	configID := uuid.New()
	id := uuid.New()

	alert := &entity.Alert{
		ID:            id,
		ConfigID:      configID,
		BrandID:       brandID,
		MentionsCount: 11,
		WindowStart:   now.Add(-time.Hour),
		WindowEnd:     now,
		FiredAt:       now,
	}

	mock.ExpectQuery("INSERT INTO alerts").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), alert.MentionsCount, alert.WindowStart, alert.WindowEnd, alert.FiredAt).
		WillReturnRows(pgxmock.NewRows([]string{"created_at"}).AddRow(now))
	require.NoError(t, repo.Create(ctx, alert))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM alerts").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT id, config_id, brand_id, mentions_count, window_start, window_end, fired_at, created_at FROM alerts").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "config_id", "brand_id", "mentions_count", "window_start", "window_end", "fired_at", "created_at",
		}).AddRow(id, configID, brandID, 11, alert.WindowStart, alert.WindowEnd, alert.FiredAt, now))

	items, total, err := repo.ListByBrandID(ctx, brandID, 10, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.EqualValues(t, 1, total)

	require.NoError(t, mock.ExpectationsWereMet())
}
