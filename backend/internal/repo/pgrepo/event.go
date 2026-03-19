package pgrepo

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"

	"prod-pobeda-2026/internal/entity"
)

// EventRepo — PostgreSQL реализация журнала событий.
type EventRepo struct {
	pool DB
	sb   squirrel.StatementBuilderType
}

// NewEventRepo создаёт новый экземпляр репозитория событий.
func NewEventRepo(pool DB) *EventRepo {
	return &EventRepo{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *EventRepo) Create(ctx context.Context, event *entity.Event) error {
	sql, args, err := r.sb.
		Insert("events").
		Columns("id", "type", "payload", "occurred_at").
		Values(event.ID, event.Type, []byte(event.Payload), event.OccurredAt).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("EventRepo.Create: build query: %w", err)
	}

	if scanErr := r.pool.QueryRow(ctx, sql, args...).Scan(&event.CreatedAt); scanErr != nil {
		return fmt.Errorf("EventRepo.Create: %w", scanErr)
	}
	return nil
}

func (r *EventRepo) List(ctx context.Context, eventType *string, limit, offset int) ([]entity.Event, int64, error) {
	countBuilder := r.sb.Select("COUNT(*)").From("events")
	if eventType != nil && *eventType != "" {
		countBuilder = countBuilder.Where(squirrel.Eq{"type": *eventType})
	}

	countSQL, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("EventRepo.List: build count query: %w", err)
	}

	var total int64
	if scanErr := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); scanErr != nil {
		return nil, 0, fmt.Errorf("EventRepo.List: count: %w", scanErr)
	}

	dataBuilder := r.sb.
		Select("id", "type", "payload", "occurred_at", "created_at").
		From("events").
		OrderBy("occurred_at DESC")
	if eventType != nil && *eventType != "" {
		dataBuilder = dataBuilder.Where(squirrel.Eq{"type": *eventType})
	}
	if limit > 0 {
		dataBuilder = dataBuilder.Limit(uint64(limit))
	}
	if offset > 0 {
		dataBuilder = dataBuilder.Offset(uint64(offset))
	}

	dataSQL, dataArgs, err := dataBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("EventRepo.List: build data query: %w", err)
	}

	rows, err := r.pool.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("EventRepo.List: query: %w", err)
	}
	defer rows.Close()

	events := make([]entity.Event, 0)
	for rows.Next() {
		var event entity.Event
		if scanErr := rows.Scan(
			&event.ID,
			&event.Type,
			&event.Payload,
			&event.OccurredAt,
			&event.CreatedAt,
		); scanErr != nil {
			return nil, 0, fmt.Errorf("EventRepo.List: scan: %w", scanErr)
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("EventRepo.List: rows: %w", rows.Err())
	}

	return events, total, nil
}
