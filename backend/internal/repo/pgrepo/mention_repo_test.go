package pgrepo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"

	"prod-pobeda-2026/internal/entity"
)

func TestMentionRepo_CreateGetListCount(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewMentionRepo(mock)
	ctx := context.Background()
	now := time.Now().UTC()
	mentionID := uuid.New()
	brandID := uuid.New()
	sourceID := uuid.New()

	mention := &entity.Mention{
		ID:          mentionID,
		BrandID:     brandID,
		SourceID:    sourceID,
		Text:        "Text",
		URL:         "https://example.com",
		Sentiment:   entity.SENTIMENT_NEGATIVE,
		PublishedAt: now,
		CreatedAt:   now,
	}

	// Create: INSERT crawler_items + INSERT sentiment_results
	mock.ExpectExec("INSERT INTO crawler_items").
		WithArgs(pgxmock.AnyArg(), mention.Text, mention.URL, mention.SourceID, mention.PublishedAt, mention.CreatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectQuery("INSERT INTO sentiment_results").
		WithArgs(mention.ID, pgxmock.AnyArg(), mention.BrandID, string(mention.Sentiment), 0.0, mention.CreatedAt).
		WillReturnRows(pgxmock.NewRows([]string{"created_at"}).AddRow(now))
	require.NoError(t, repo.Create(ctx, mention))

	pgSourceID := pgtype.UUID{Bytes: mention.SourceID, Valid: true}

	// GetByID
	mock.ExpectQuery("SELECT sr.id").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "brand_id", "source_id", "source_name", "source_type", "text", "link", "sentiment", "published_at", "created_at", "cluster_label",
		}).AddRow(
			mention.ID, mention.BrandID, pgSourceID,
			"TestSource", "telegram",
			mention.Text, mention.URL, string(mention.Sentiment),
			mention.PublishedAt, now, nil,
		))
	got, err := repo.GetByID(ctx, mentionID)
	require.NoError(t, err)
	require.Equal(t, mentionID, got.ID)

	// List
	filter := entity.MentionFilter{
		BrandID:  brandID,
		Search:   "Tex",
		Limit:    10,
		Offset:   0,
		DateFrom: now.Add(-time.Hour).Format(time.RFC3339),
		DateTo:   now.Add(time.Hour).Format(time.RFC3339),
	}
	mock.ExpectQuery("SELECT COUNT").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT sr.id").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "brand_id", "source_id", "source_name", "source_type", "text", "link", "sentiment", "published_at", "created_at", "cluster_label",
		}).AddRow(
			mention.ID, mention.BrandID, pgSourceID,
			"TestSource", "telegram",
			mention.Text, mention.URL, string(mention.Sentiment),
			mention.PublishedAt, now, nil,
		))
	items, total, err := repo.List(ctx, filter)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.EqualValues(t, 1, total)

	// CountByBrandSince
	mock.ExpectQuery("SELECT COUNT").
		WithArgs(pgxmock.AnyArg(), now, "negative").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(3)))
	count, err := repo.CountByBrandSince(ctx, brandID, now, "negative")
	require.NoError(t, err)
	require.EqualValues(t, 3, count)

	require.NoError(t, mock.ExpectationsWereMet())
}
