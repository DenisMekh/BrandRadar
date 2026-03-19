package pgrepo

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"prod-pobeda-2026/internal/entity"
)

func TestScanSource(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	id := uuid.New()

	testCases := []struct {
		name    string
		url     string
		scanErr error
		wantErr bool
		wantURL string
	}{
		{
			name:    "with url",
			url:     "https://example.com",
			wantURL: "https://example.com",
		},
		{
			name:    "empty url",
			url:     "",
			wantURL: "",
		},
		{
			name:    "scan error",
			scanErr: errors.New("boom"),
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
					*(dest[1].(*string)) = "web"
					*(dest[2].(*string)) = "source-1"
					*(dest[3].(*string)) = tc.url
					*(dest[4].(*entity.SourceStatus)) = entity.SourceStatusActive
					*(dest[5].(*time.Time)) = now
					*(dest[6].(*time.Time)) = now
					return nil
				},
			}

			got, err := scanSource(row)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, id, got.ID)
			require.Equal(t, tc.wantURL, got.URL)
		})
	}
}
