package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"prod-pobeda-2026/internal/entity"
)

// DashboardUseCase — бизнес-логика дашбордов.
type DashboardUseCase struct {
	dashRepo  DashboardRepository
	brandRepo BrandRepository
}

// NewDashboardUseCase создаёт DashboardUseCase.
func NewDashboardUseCase(dashRepo DashboardRepository, brandRepo BrandRepository) *DashboardUseCase {
	return &DashboardUseCase{dashRepo: dashRepo, brandRepo: brandRepo}
}

// GetBrandDashboard возвращает полную статистику по бренду за период.
func (uc *DashboardUseCase) GetBrandDashboard(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) (*entity.BrandDashboard, error) {
	brand, err := uc.brandRepo.GetByID(ctx, brandID)
	if err != nil {
		return nil, fmt.Errorf("DashboardUseCase.GetBrandDashboard: get brand: %w", err)
	}

	sentiment, err := uc.dashRepo.GetBrandSentiment(ctx, brandID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("DashboardUseCase.GetBrandDashboard: sentiment: %w", err)
	}

	sources, err := uc.dashRepo.GetBrandSourceStats(ctx, brandID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("DashboardUseCase.GetBrandDashboard: sources: %w", err)
	}

	daily, err := uc.dashRepo.GetBrandDailyStats(ctx, brandID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("DashboardUseCase.GetBrandDashboard: daily: %w", err)
	}

	alertCount, err := uc.dashRepo.GetBrandAlertCount(ctx, brandID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("DashboardUseCase.GetBrandDashboard: alerts: %w", err)
	}

	return &entity.BrandDashboard{
		BrandID:      brand.ID,
		BrandName:    brand.Name,
		Sentiment:    sentiment,
		BySources:    sources,
		ByDate:       daily,
		RecentAlerts: alertCount,
	}, nil
}

// GetOverallDashboard возвращает общую статистику по всем брендам.
func (uc *DashboardUseCase) GetOverallDashboard(ctx context.Context, dateFrom, dateTo string) (*entity.SentimentCounts, []entity.BrandSummary, []entity.DailyCount, error) {
	sentiment, err := uc.dashRepo.GetOverallSentiment(ctx, dateFrom, dateTo)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("DashboardUseCase.GetOverallDashboard: sentiment: %w", err)
	}

	brands, err := uc.dashRepo.GetAllBrandsSummary(ctx, dateFrom, dateTo)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("DashboardUseCase.GetOverallDashboard: brands: %w", err)
	}

	daily, err := uc.dashRepo.GetOverallDailyStats(ctx, dateFrom, dateTo)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("DashboardUseCase.GetOverallDashboard: daily: %w", err)
	}

	return &sentiment, brands, daily, nil
}
