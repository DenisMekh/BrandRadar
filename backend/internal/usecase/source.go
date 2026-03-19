package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/entity"
	"prod-pobeda-2026/internal/metrics"
)

// SourceUseCase — бизнес-логика управления источниками данных.
type SourceUseCase struct {
	repo            SourceRepository
	eventRepo       EventRepository
	metrics         metrics.Business
	onSourceCreated func()
}

// NewSourceUseCase — конструктор usecase источников.
func NewSourceUseCase(repo SourceRepository, eventRepo EventRepository, businessMetrics ...metrics.Business) *SourceUseCase {
	m := metrics.NopBusiness()
	if len(businessMetrics) > 0 && businessMetrics[0] != nil {
		m = businessMetrics[0]
	}
	return &SourceUseCase{
		repo:      repo,
		eventRepo: eventRepo,
		metrics:   m,
	}
}

// SetOnSourceCreated устанавливает callback, вызываемый после создания источника.
func (uc *SourceUseCase) SetOnSourceCreated(fn func()) {
	uc.onSourceCreated = fn
}

// Create создаёт источник после валидации входных данных.
func (uc *SourceUseCase) Create(ctx context.Context, s *entity.Source) error {
	s.Name = strings.TrimSpace(s.Name)
	if s.Name == "" {
		return fmt.Errorf("SourceUseCase.Create: name is empty: %w", entity.ErrValidation)
	}
	if s.Type == "" {
		return fmt.Errorf("SourceUseCase.Create: type is empty: %w", entity.ErrValidation)
	}

	// Валидация URL по типу источника
	if err := entity.ValidateSourceURL(s.Type, s.URL); err != nil {
		return fmt.Errorf("SourceUseCase.Create: %w: %w", entity.ErrValidation, err)
	}

	s.ID = uuid.New()
	s.Status = entity.SourceStatusActive
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now

	if err := uc.repo.Create(ctx, s); err != nil {
		return fmt.Errorf("SourceUseCase.Create: %w", err)
	}
	if err := uc.syncSourcesActive(ctx); err != nil {
		return fmt.Errorf("SourceUseCase.Create: sync active sources: %w", err)
	}

	logrus.Infof("source created: id=%s, type=%s", s.ID.String(), s.Type)

	if uc.onSourceCreated != nil {
		uc.onSourceCreated()
	}

	return nil
}

// GetByID возвращает источник по идентификатору.
func (uc *SourceUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Source, error) {
	s, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("SourceUseCase.GetByID: %w", err)
	}
	return s, nil
}

// List возвращает источники с пагинацией.
func (uc *SourceUseCase) List(ctx context.Context, limit, offset int) ([]entity.Source, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("SourceUseCase.List: %w", err)
	}
	return items, int(total), nil
}

// Toggle переключает статус источника active <-> inactive.
func (uc *SourceUseCase) Toggle(ctx context.Context, id uuid.UUID) (*entity.Source, error) {
	s, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("SourceUseCase.Toggle: %w", err)
	}

	if s.Status == entity.SourceStatusActive {
		s.Status = entity.SourceStatusInactive
	} else {
		s.Status = entity.SourceStatusActive
	}
	s.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(ctx, s); err != nil {
		return nil, fmt.Errorf("SourceUseCase.Toggle: %w", err)
	}
	if err := uc.syncSourcesActive(ctx); err != nil {
		return nil, fmt.Errorf("SourceUseCase.Toggle: sync active sources: %w", err)
	}

	if err := uc.eventRepo.Create(ctx, &entity.Event{
		ID:         uuid.New(),
		Type:       entity.EventSourceToggled,
		Payload:    []byte(fmt.Sprintf(`{"source_id":"%s","status":"%s"}`, s.ID, s.Status)),
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		logrus.Warnf("failed to write source toggle event: %v", err)
	}

	logrus.Infof("source toggled: id=%s, status=%s", s.ID.String(), string(s.Status))
	return s, nil
}

// Delete удаляет источник.
func (uc *SourceUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("SourceUseCase.Delete: %w", err)
	}
	if err := uc.syncSourcesActive(ctx); err != nil {
		return fmt.Errorf("SourceUseCase.Delete: sync active sources: %w", err)
	}
	return nil
}

func (uc *SourceUseCase) syncSourcesActive(ctx context.Context) error {
	counts, err := uc.repo.CountActiveByType(ctx)
	if err != nil {
		return err
	}
	if len(counts) == 0 {
		uc.metrics.SetSourcesActive("all", 0)
		return nil
	}
	for sourceType, count := range counts {
		uc.metrics.SetSourcesActive(sourceType, float64(count))
	}
	return nil
}
