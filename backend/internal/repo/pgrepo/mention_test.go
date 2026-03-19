package pgrepo

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"prod-pobeda-2026/internal/entity"
)

func TestApplyMentionFilters(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	testCases := []struct {
		name           string
		filter         entity.MentionFilter
		wantErr        bool
		wantErrSubstr  string
		wantSQLSubstrs []string
	}{
		{
			name:   "empty filter",
			filter: entity.MentionFilter{},
		},
		{
			name: "all filters",
			filter: entity.MentionFilter{
				BrandID:   uuid.New(),
				SourceID:  uuid.New(),
				Sentiment: "negative",
				Search:    "brand",
				DateFrom:  now.Add(-time.Hour).Format(time.RFC3339),
				DateTo:    now.Format(time.RFC3339),
			},
			wantSQLSubstrs: []string{
				"brand_id",
				"source_id",
				"sr.sentiment",
				"ci.text",
				"ci.published_at",
			},
		},
		{
			name: "invalid date from",
			filter: entity.MentionFilter{
				DateFrom: "not-a-date",
			},
			wantErr:       true,
			wantErrSubstr: "invalid date_from",
		},
		{
			name: "invalid date to",
			filter: entity.MentionFilter{
				DateTo: "not-a-date",
			},
			wantErr:       true,
			wantErrSubstr: "invalid date_to",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			builder := squirrel.StatementBuilder.
				PlaceholderFormat(squirrel.Dollar).
				Select("*").
				From("sentiment_results sr")

			gotBuilder, err := applyMentionFilters(builder, tc.filter)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrSubstr)
				return
			}

			require.NoError(t, err)
			sql, _, err := gotBuilder.ToSql()
			require.NoError(t, err)

			for _, substr := range tc.wantSQLSubstrs {
				require.True(t, strings.Contains(sql, substr), "sql must contain %q, got: %s", substr, sql)
			}
		})
	}
}

func TestScanMention(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	id := uuid.New()
	brandID := uuid.New()
	sourceID := uuid.New()

	testCases := []struct {
		name    string
		scanErr error
		wantErr bool
	}{
		{
			name: "success",
		},
		{
			name:    "scan error",
			scanErr: errors.New("scan failed"),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			row := scannerStub{
				scanFn: func(dest ...any) error {
					if tc.scanErr != nil {
						return tc.scanErr
					}
					*(dest[0].(*uuid.UUID)) = id
					*(dest[1].(*uuid.UUID)) = brandID
					dest[2].(*pgtype.UUID).Bytes = sourceID
					dest[2].(*pgtype.UUID).Valid = true
					*(dest[3].(*string)) = "Медуза"
					*(dest[4].(*string)) = "telegram"
					*(dest[5].(*string)) = "text"
					*(dest[6].(*string)) = "https://example.com"
					*(dest[7].(*string)) = "negative"
					*(dest[8].(*time.Time)) = now
					*(dest[9].(*time.Time)) = now
					return nil
				},
			}

			got, err := scanMention(row)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, id, got.ID)
			require.Equal(t, brandID, got.BrandID)
			require.Equal(t, sourceID, got.SourceID)
			require.Equal(t, entity.Sentiment("negative"), got.Sentiment)
		})
	}
}
