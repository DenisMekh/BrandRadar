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

func TestSourceRepo_CreateGetListActiveUpdateDelete(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewSourceRepo(mock)
	ctx := context.Background()
	now := time.Now().UTC()
	sourceID := uuid.New()

	source := &entity.Source{
		ID:     sourceID,
		Type:   "web",
		Name:   "source-1",
		URL:    "https://example.com",
		Status: entity.SourceStatusActive,
	}

	mock.ExpectQuery("INSERT INTO sources").
		WithArgs(pgxmock.AnyArg(), source.Type, source.Name, source.URL, source.Status).
		WillReturnRows(pgxmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))
	require.NoError(t, repo.Create(ctx, source))

	mock.ExpectQuery("SELECT id, type, name, url, status, created_at, updated_at FROM sources").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "type", "name", "url", "status", "created_at", "updated_at",
		}).AddRow(sourceID, "web", "source-1", "https://example.com", entity.SourceStatusActive, now, now))
	got, err := repo.GetByID(ctx, sourceID)
	require.NoError(t, err)
	require.Equal(t, sourceID, got.ID)

	mock.ExpectQuery("SELECT id, type, name, url, status, created_at, updated_at FROM sources").
		WithArgs(entity.SourceStatusActive).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "type", "name", "url", "status", "created_at", "updated_at",
		}).AddRow(sourceID, "web", "source-1", "https://example.com", entity.SourceStatusActive, now, now))
	active, err := repo.ListActive(ctx)
	require.NoError(t, err)
	require.Len(t, active, 1)

	mock.ExpectQuery("UPDATE sources").
		WithArgs("web", "source-updated", source.URL, source.Status, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"updated_at"}).AddRow(now.Add(time.Minute)))
	source.Name = "source-updated"
	require.NoError(t, repo.Update(ctx, source))

	mock.ExpectExec("DELETE FROM sources").
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	require.NoError(t, repo.Delete(ctx, sourceID))

	require.NoError(t, mock.ExpectationsWereMet())
}
