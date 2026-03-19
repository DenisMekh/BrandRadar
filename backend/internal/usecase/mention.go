package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/entity"
)

// MentionUseCase — бизнес-логика обработки упоминаний.
type MentionUseCase struct {
	repo      MentionRepository
	eventRepo EventRepository
	alertUC   *AlertUseCase
	cache     MentionCache
}

// NewMentionUseCase — конструктор usecase упоминаний.
func NewMentionUseCase(
	repo MentionRepository,
	eventRepo EventRepository,
	alertUC *AlertUseCase,
	cache MentionCache,
) *MentionUseCase {
	return &MentionUseCase{
		repo:      repo,
		eventRepo: eventRepo,
		alertUC:   alertUC,
		cache:     cache,
	}
}

// Create создаёт упоминание и проверяет spike.
func (uc *MentionUseCase) Create(ctx context.Context, m *entity.Mention) (*entity.Mention, error) {
	if m.BrandID == uuid.Nil {
		return nil, fmt.Errorf("MentionUseCase.Create: brand_id is nil: %w", entity.ErrValidation)
	}
	if m.SourceID == uuid.Nil {
		return nil, fmt.Errorf("MentionUseCase.Create: source_id is nil: %w", entity.ErrValidation)
	}
	if m.Text == "" {
		return nil, fmt.Errorf("MentionUseCase.Create: text is empty: %w", entity.ErrValidation)
	}

	m.ID = uuid.New()
	m.CreatedAt = time.Now().UTC()

	if err := uc.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("MentionUseCase.Create: %w", err)
	}

	now := m.CreatedAt
	if err := uc.eventRepo.Create(ctx, &entity.Event{
		ID:         uuid.New(),
		Type:       entity.EventMentionCreated,
		Payload:    []byte(fmt.Sprintf(`{"mention_id":"%s","brand_id":"%s","source_id":"%s"}`, m.ID, m.BrandID, m.SourceID)),
		OccurredAt: now,
	}); err != nil {
		logrus.Warnf("failed to write mention creation event: %v", err)
	}

	if uc.cache != nil {
		if err := uc.cache.InvalidateByBrand(ctx, m.BrandID); err != nil {
			logrus.Warnf("failed to invalidate mentions cache after create: %v", err)
		}
	}

	logrus.Infof("mention created: id=%s, brand_id=%s, sentiment=%s", m.ID.String(), m.BrandID.String(), m.Sentiment)

	if uc.alertUC != nil {
		if err := uc.alertUC.CheckSpike(ctx, m.BrandID); err != nil {
			logrus.Warnf("spike check failed: %v", err)
		}
	}

	return m, nil
}

// GetByID возвращает упоминание по идентификатору.
func (uc *MentionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Mention, error) {
	m, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("MentionUseCase.GetByID: %w", err)
	}
	uc.enrichSimilar(ctx, m)
	return m, nil
}

const similarMentionsLimit = 5

// enrichSimilar заполняет m.Similar упоминаниями того же кластера.
func (uc *MentionUseCase) enrichSimilar(ctx context.Context, m *entity.Mention) {
	if m.ClusterLabel == nil {
		return
	}
	similar, err := uc.repo.GetSimilarMentions(ctx, m.BrandID, *m.ClusterLabel, m.ID, similarMentionsLimit)
	if err != nil {
		logrus.Warnf("enrichSimilar: %v", err)
		return
	}
	m.Similar = similar
}

// List возвращает упоминания по фильтрам с пагинацией.
func (uc *MentionUseCase) List(ctx context.Context, filter MentionFilter) ([]entity.Mention, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	if uc.cache != nil {
		cachedItems, cachedTotal, cacheErr := uc.cache.GetList(ctx, filter)
		if cacheErr != nil {
			logrus.Warnf("failed to get mentions list from cache: %v", cacheErr)
		} else if cachedItems != nil {
			return cachedItems, int(cachedTotal), nil
		}
	}

	items, total, err := uc.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("MentionUseCase.List: %w", err)
	}

	for i := range items {
		uc.enrichSimilar(ctx, &items[i])
	}

	if uc.cache != nil {
		if err := uc.cache.SetList(ctx, filter, items, total); err != nil {
			logrus.Warnf("failed to set mentions list to cache: %v", err)
		}
	}
	return items, int(total), nil
}
