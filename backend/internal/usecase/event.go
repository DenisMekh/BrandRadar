package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"prod-pobeda-2026/internal/entity"
)

// EventUseCase — бизнес-логика журнала событий.
type EventUseCase struct {
	repo EventRepository
}

// NewEventUseCase — конструктор usecase событий.
func NewEventUseCase(repo EventRepository) *EventUseCase {
	return &EventUseCase{repo: repo}
}

// Create записывает событие в журнал.
func (uc *EventUseCase) Create(ctx context.Context, e *entity.Event) error {
	if e.Type == "" {
		return fmt.Errorf("EventUseCase.Create: type is empty: %w", entity.ErrValidation)
	}

	e.ID = uuid.New()
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	e.CreatedAt = time.Now().UTC()

	if err := uc.repo.Create(ctx, e); err != nil {
		return fmt.Errorf("EventUseCase.Create: %w", err)
	}
	return nil
}

// List возвращает события с фильтрацией по типу.
func (uc *EventUseCase) List(ctx context.Context, eventType string, limit, offset int) ([]entity.Event, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var eventTypePtr *string
	if eventType != "" {
		eventTypePtr = &eventType
	}

	items, total, err := uc.repo.List(ctx, eventTypePtr, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("EventUseCase.List: %w", err)
	}
	return items, int(total), nil
}
