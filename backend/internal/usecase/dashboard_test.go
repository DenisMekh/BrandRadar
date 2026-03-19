package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"prod-pobeda-2026/internal/entity"
	"prod-pobeda-2026/internal/usecase/mocks"
)

func TestDashboardUseCase_GetBrandDashboard_Success(t *testing.T) {
	dashRepo := new(mocks.MockDashboardRepository)
	brandRepo := new(mocks.MockBrandRepository)
	uc := NewDashboardUseCase(dashRepo, brandRepo)

	brandID := uuid.New()
	brand := &entity.Brand{ID: brandID, Name: "TestBrand"}
	sentiment := entity.SentimentCounts{Positive: 10, Negative: 5, Neutral: 3}
	sources := []entity.SourceCount{{Source: "telegram", Count: 12}}
	daily := []entity.DailyCount{{Date: "2026-03-14", Positive: 5, Negative: 2, Neutral: 1}}

	brandRepo.On("GetByID", mock.Anything, brandID).Return(brand, nil)
	dashRepo.On("GetBrandSentiment", mock.Anything, brandID, "2026-03-01", "2026-03-15").Return(sentiment, nil)
	dashRepo.On("GetBrandSourceStats", mock.Anything, brandID, "2026-03-01", "2026-03-15").Return(sources, nil)
	dashRepo.On("GetBrandDailyStats", mock.Anything, brandID, "2026-03-01", "2026-03-15").Return(daily, nil)
	dashRepo.On("GetBrandAlertCount", mock.Anything, brandID, "2026-03-01", "2026-03-15").Return(int64(2), nil)

	result, err := uc.GetBrandDashboard(context.Background(), brandID, "2026-03-01", "2026-03-15")

	assert.NoError(t, err)
	assert.Equal(t, brandID, result.BrandID)
	assert.Equal(t, "TestBrand", result.BrandName)
	assert.Equal(t, int64(10), result.Sentiment.Positive)
	assert.Equal(t, int64(5), result.Sentiment.Negative)
	assert.Equal(t, int64(3), result.Sentiment.Neutral)
	assert.Len(t, result.BySources, 1)
	assert.Len(t, result.ByDate, 1)
	assert.Equal(t, int64(2), result.RecentAlerts)

	dashRepo.AssertExpectations(t)
	brandRepo.AssertExpectations(t)
}

func TestDashboardUseCase_GetBrandDashboard_BrandNotFound(t *testing.T) {
	dashRepo := new(mocks.MockDashboardRepository)
	brandRepo := new(mocks.MockBrandRepository)
	uc := NewDashboardUseCase(dashRepo, brandRepo)

	brandID := uuid.New()
	brandRepo.On("GetByID", mock.Anything, brandID).Return((*entity.Brand)(nil), entity.ErrNotFound)

	result, err := uc.GetBrandDashboard(context.Background(), brandID, "", "")

	assert.Error(t, err)
	assert.Nil(t, result)
	brandRepo.AssertExpectations(t)
}

func TestDashboardUseCase_GetOverallDashboard_Success(t *testing.T) {
	dashRepo := new(mocks.MockDashboardRepository)
	brandRepo := new(mocks.MockBrandRepository)
	uc := NewDashboardUseCase(dashRepo, brandRepo)

	sentiment := entity.SentimentCounts{Positive: 100, Negative: 50, Neutral: 30}
	brands := []entity.BrandSummary{
		{BrandID: uuid.New(), BrandName: "Brand1", Sentiment: entity.SentimentCounts{Positive: 60, Negative: 30, Neutral: 20}},
		{BrandID: uuid.New(), BrandName: "Brand2", Sentiment: entity.SentimentCounts{Positive: 40, Negative: 20, Neutral: 10}},
	}
	daily := []entity.DailyCount{
		{Date: "2026-03-14", Positive: 50, Negative: 25, Neutral: 15},
	}

	dashRepo.On("GetOverallSentiment", mock.Anything, "", "").Return(sentiment, nil)
	dashRepo.On("GetAllBrandsSummary", mock.Anything, "", "").Return(brands, nil)
	dashRepo.On("GetOverallDailyStats", mock.Anything, "", "").Return(daily, nil)

	s, b, d, err := uc.GetOverallDashboard(context.Background(), "", "")

	assert.NoError(t, err)
	assert.Equal(t, int64(100), s.Positive)
	assert.Len(t, b, 2)
	assert.Len(t, d, 1)

	dashRepo.AssertExpectations(t)
}
