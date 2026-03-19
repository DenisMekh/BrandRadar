package pgrepo

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"prod-pobeda-2026/internal/entity"
)

func TestScanAlertConfig(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	id := uuid.New()
	brandID := uuid.New()

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
					*(dest[2].(*int)) = 60
					*(dest[3].(*int)) = 30
					*(dest[4].(*string)) = "negative"
					*(dest[5].(*bool)) = true
					*(dest[6].(*float64)) = 95.0
					*(dest[7].(*int)) = 10
					*(dest[8].(*time.Time)) = now
					*(dest[9].(*time.Time)) = now
					return nil
				},
			}

			got, err := scanAlertConfig(row)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, id, got.ID)
			require.Equal(t, brandID, got.BrandID)
			require.Equal(t, entity.AlertConfig{
				ID:                id,
				BrandID:           brandID,
				WindowMinutes:     60,
				CooldownMinutes:   30,
				SentimentFilter:   "negative",
				Enabled:           true,
				Percentile:        95.0,
				AnomalyWindowSize: 10,
				CreatedAt:         now,
				UpdatedAt:         now,
			}, *got)
		})
	}
}
