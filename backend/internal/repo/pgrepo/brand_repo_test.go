package pgrepo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"

	"prod-pobeda-2026/internal/entity"
)

func TestBrandRepo_CRUDAndList(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewBrandRepo(mock)
	ctx := context.Background()
	now := time.Now().UTC()
	brandID := uuid.New()

	brand := &entity.Brand{
		ID:         brandID,
		Name:       "brand",
		Keywords:   []string{"k1"},
		Exclusions: []string{"e1"},
		RiskWords:  []string{"r1"},
	}

	mock.ExpectQuery("INSERT INTO brands").
		WithArgs(pgxmock.AnyArg(), brand.Name, brand.Keywords, brand.Exclusions, brand.RiskWords).
		WillReturnRows(pgxmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))

	err = repo.Create(ctx, brand)
	require.NoError(t, err)
	require.Equal(t, now, brand.CreatedAt)

	mock.ExpectQuery("SELECT id, name, keywords, exclusions, risk_words, created_at, updated_at FROM brands").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "keywords", "exclusions", "risk_words", "created_at", "updated_at",
		}).AddRow(brandID, "brand", []string{"k1"}, []string{"e1"}, []string{"r1"}, now, now))

	got, err := repo.GetByID(ctx, brandID)
	require.NoError(t, err)
	require.Equal(t, brandID, got.ID)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM brands").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT id, name, keywords, exclusions, risk_words, created_at, updated_at FROM brands").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "keywords", "exclusions", "risk_words", "created_at", "updated_at",
		}).AddRow(brandID, "brand", []string{"k1"}, []string{"e1"}, []string{"r1"}, now, now))

	items, total, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.EqualValues(t, 1, total)

	mock.ExpectQuery("UPDATE brands").
		WithArgs("brand-updated", []string{"k1"}, []string{"e1"}, []string{"r1"}, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"updated_at"}).AddRow(now.Add(time.Minute)))

	brand.Name = "brand-updated"
	err = repo.Update(ctx, brand)
	require.NoError(t, err)

	mock.ExpectExec("DELETE FROM brands").
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(ctx, brandID)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBrandRepo_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewBrandRepo(mock)
	id := uuid.New()

	mock.ExpectQuery("SELECT id, name, keywords, exclusions, risk_words, created_at, updated_at FROM brands").
		WithArgs(pgxmock.AnyArg()).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(context.Background(), id)
	require.ErrorIs(t, err, entity.ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
