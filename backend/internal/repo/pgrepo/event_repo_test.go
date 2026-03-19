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

func TestEventRepo_CreateAndList(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewEventRepo(mock)
	ctx := context.Background()
	now := time.Now().UTC()
	id := uuid.New()

	event := &entity.Event{
		ID:         id,
		Type:       entity.EventAlertFired,
		Payload:    []byte(`{"id":"1"}`),
		OccurredAt: now,
	}

	mock.ExpectQuery("INSERT INTO events").
		WithArgs(pgxmock.AnyArg(), event.Type, []byte(event.Payload), event.OccurredAt).
		WillReturnRows(pgxmock.NewRows([]string{"created_at"}).AddRow(now))
	require.NoError(t, repo.Create(ctx, event))

	eventType := string(entity.EventAlertFired)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM events").
		WithArgs(eventType).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT id, type, payload, occurred_at, created_at FROM events").
		WithArgs(eventType).
		WillReturnRows(pgxmock.NewRows([]string{"id", "type", "payload", "occurred_at", "created_at"}).
			AddRow(id, event.Type, event.Payload, now, now))

	items, total, err := repo.List(ctx, &eventType, 10, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.EqualValues(t, 1, total)

	require.NoError(t, mock.ExpectationsWereMet())
}
