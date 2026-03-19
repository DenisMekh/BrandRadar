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

// BrandUseCase — бизнес-логика управления брендами.
type BrandUseCase struct {
	repo    BrandRepository
	metrics metrics.Business
}

// NewBrandUseCase — конструктор usecase брендов.
func NewBrandUseCase(repo BrandRepository, businessMetrics ...metrics.Business) *BrandUseCase {
	m := metrics.NopBusiness()
	if len(businessMetrics) > 0 && businessMetrics[0] != nil {
		m = businessMetrics[0]
	}
	return &BrandUseCase{
		repo:    repo,
		metrics: m,
	}
}

// Create создаёт бренд с валидацией обязательных полей.
func (uc *BrandUseCase) Create(ctx context.Context, b *entity.Brand) error {
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" {
		return fmt.Errorf("BrandUseCase.Create: name is empty: %w", entity.ErrValidation)
	}
	if len(b.Keywords) == 0 {
		return fmt.Errorf("BrandUseCase.Create: keywords are empty: %w", entity.ErrValidation)
	}
	b.ID = uuid.New()
	now := time.Now().UTC()
	b.CreatedAt = now
	b.UpdatedAt = now

	if err := uc.repo.Create(ctx, b); err != nil {
		return fmt.Errorf("BrandUseCase.Create: %w", err)
	}
	if err := uc.syncBrandsTotal(ctx); err != nil {
		return fmt.Errorf("BrandUseCase.Create: sync brands total: %w", err)
	}

	logrus.Infof("brand created: id=%s, name=%s, keywords_count=%d", b.ID.String(), b.Name, len(b.Keywords))
	return nil
}

// GetByID возвращает бренд по идентификатору.
func (uc *BrandUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Brand, error) {
	b, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("BrandUseCase.GetByID: %w", err)
	}
	return b, nil
}

// List возвращает бренды с пагинацией.
func (uc *BrandUseCase) List(ctx context.Context, limit, offset int) ([]entity.Brand, int, error) {
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
		return nil, 0, fmt.Errorf("BrandUseCase.List: %w", err)
	}
	return items, int(total), nil
}

// Update обновляет бренд.
func (uc *BrandUseCase) Update(ctx context.Context, b *entity.Brand) error {
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" {
		return fmt.Errorf("BrandUseCase.Update: name is empty: %w", entity.ErrValidation)
	}
	if len(b.Keywords) == 0 {
		return fmt.Errorf("BrandUseCase.Update: keywords are empty: %w", entity.ErrValidation)
	}
	b.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(ctx, b); err != nil {
		return fmt.Errorf("BrandUseCase.Update: %w", err)
	}
	return nil
}

// Delete удаляет бренд.
func (uc *BrandUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("BrandUseCase.Delete: %w", err)
	}
	if err := uc.syncBrandsTotal(ctx); err != nil {
		return fmt.Errorf("BrandUseCase.Delete: sync brands total: %w", err)
	}
	logrus.Infof("brand deleted: id=%s", id.String())
	return nil
}

func (uc *BrandUseCase) syncBrandsTotal(ctx context.Context) error {
	total, err := uc.repo.Count(ctx)
	if err != nil {
		return err
	}
	uc.metrics.SetBrandsTotal(float64(total))
	return nil
}
