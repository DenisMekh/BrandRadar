package mocks

import (
	"context"
	"time"

	"prod-pobeda-2026/internal/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockBrandRepository struct {
	mock.Mock
}

func (m *MockBrandRepository) Create(ctx context.Context, brand *entity.Brand) error {
	args := m.Called(ctx, brand)
	return args.Error(0)
}

func (m *MockBrandRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Brand, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Brand), args.Error(1)
}

func (m *MockBrandRepository) List(ctx context.Context, limit, offset int) ([]entity.Brand, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]entity.Brand), args.Get(1).(int64), args.Error(2)
}

func (m *MockBrandRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBrandRepository) Update(ctx context.Context, brand *entity.Brand) error {
	args := m.Called(ctx, brand)
	return args.Error(0)
}

func (m *MockBrandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockSourceRepository struct {
	mock.Mock
}

func (m *MockSourceRepository) Create(ctx context.Context, source *entity.Source) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *MockSourceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Source, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Source), args.Error(1)
}

func (m *MockSourceRepository) List(ctx context.Context, limit, offset int) ([]entity.Source, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]entity.Source), args.Get(1).(int64), args.Error(2)
}

func (m *MockSourceRepository) CountActiveByType(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockSourceRepository) ListActive(ctx context.Context) ([]entity.Source, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entity.Source), args.Error(1)
}

func (m *MockSourceRepository) Update(ctx context.Context, source *entity.Source) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *MockSourceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockMentionRepository struct {
	mock.Mock
}

func (m *MockMentionRepository) Create(ctx context.Context, mention *entity.Mention) error {
	args := m.Called(ctx, mention)
	return args.Error(0)
}

func (m *MockMentionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Mention, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Mention), args.Error(1)
}

func (m *MockMentionRepository) List(ctx context.Context, filter entity.MentionFilter) ([]entity.Mention, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Mention), args.Get(1).(int64), args.Error(2)
}

func (m *MockMentionRepository) CountByBrandSince(ctx context.Context, brandID uuid.UUID, since time.Time, sentimentFilter string) (int64, error) {
	args := m.Called(ctx, brandID, since, sentimentFilter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMentionRepository) CountByBrandInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time, sentimentFilter string) (int64, error) {
	args := m.Called(ctx, brandID, from, to, sentimentFilter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMentionRepository) GetAverageSentimentInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time) (float64, error) {
	args := m.Called(ctx, brandID, from, to)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMentionRepository) GetWeightedSentimentPercentileInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time, percentile float64) (float64, error) {
	args := m.Called(ctx, brandID, from, to, percentile)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMentionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.MentionStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockMentionRepository) UpdateML(ctx context.Context, id uuid.UUID, ml entity.MentionML) error {
	args := m.Called(ctx, id, ml)
	return args.Error(0)
}

func (m *MockMentionRepository) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Mention, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]entity.Mention), args.Error(1)
}

func (m *MockMentionRepository) ListByStatusWithLimit(ctx context.Context, status entity.MentionStatus, limit int) ([]entity.Mention, error) {
	args := m.Called(ctx, status, limit)
	return args.Get(0).([]entity.Mention), args.Error(1)
}

func (m *MockMentionRepository) GetSimilarMentions(ctx context.Context, brandID uuid.UUID, clusterLabel int, excludeID uuid.UUID, limit int) ([]entity.Mention, error) {
	args := m.Called(ctx, brandID, clusterLabel, excludeID, limit)
	return args.Get(0).([]entity.Mention), args.Error(1)
}

type MockAlertConfigRepository struct {
	mock.Mock
}

func (m *MockAlertConfigRepository) Create(ctx context.Context, config *entity.AlertConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockAlertConfigRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AlertConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AlertConfig), args.Error(1)
}

func (m *MockAlertConfigRepository) GetByBrandID(ctx context.Context, brandID uuid.UUID) (*entity.AlertConfig, error) {
	args := m.Called(ctx, brandID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AlertConfig), args.Error(1)
}

func (m *MockAlertConfigRepository) Update(ctx context.Context, config *entity.AlertConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockAlertConfigRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertConfigRepository) GetAllActive(ctx context.Context) ([]*entity.AlertConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlertConfig), args.Error(1)
}

func (m *MockAlertConfigRepository) ListAll(ctx context.Context) ([]*entity.AlertConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlertConfig), args.Error(1)
}

type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) Create(ctx context.Context, alert *entity.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) ListByBrandID(ctx context.Context, brandID uuid.UUID, limit, offset int) ([]entity.Alert, int64, error) {
	args := m.Called(ctx, brandID, limit, offset)
	return args.Get(0).([]entity.Alert), args.Get(1).(int64), args.Error(2)
}

func (m *MockAlertRepository) ListAll(ctx context.Context, limit, offset int) ([]entity.Alert, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]entity.Alert), args.Get(1).(int64), args.Error(2)
}

func (m *MockAlertRepository) GetLastFiredAt(ctx context.Context, configID uuid.UUID) (*time.Time, error) {
	args := m.Called(ctx, configID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(ctx context.Context, event *entity.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) List(ctx context.Context, eventType *string, limit, offset int) ([]entity.Event, int64, error) {
	args := m.Called(ctx, eventType, limit, offset)
	return args.Get(0).([]entity.Event), args.Get(1).(int64), args.Error(2)
}

type MockCooldownCache struct {
	mock.Mock
}

func (m *MockCooldownCache) TryLock(ctx context.Context, configID uuid.UUID, cooldownMinutes int) (bool, error) {
	args := m.Called(ctx, configID, cooldownMinutes)
	return args.Bool(0), args.Error(1)
}

type MockMentionCache struct {
	mock.Mock
}

func (m *MockMentionCache) GetList(ctx context.Context, filter entity.MentionFilter) ([]entity.Mention, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Mention), args.Get(1).(int64), args.Error(2)
}

func (m *MockMentionCache) SetList(ctx context.Context, filter entity.MentionFilter, items []entity.Mention, total int64) error {
	args := m.Called(ctx, filter, items, total)
	return args.Error(0)
}

func (m *MockMentionCache) InvalidateByBrand(ctx context.Context, brandID uuid.UUID) error {
	args := m.Called(ctx, brandID)
	return args.Error(0)
}

func (m *MockMentionCache) InvalidateAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockHealthChecker struct {
	mock.Mock
}

func (m *MockHealthChecker) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockHealthChecker) Name() string {
	args := m.Called()
	return args.String(0)
}

// MockDashboardRepository — мок для DashboardRepository.
type MockDashboardRepository struct {
	mock.Mock
}

func (m *MockDashboardRepository) GetBrandSentiment(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) (entity.SentimentCounts, error) {
	args := m.Called(ctx, brandID, dateFrom, dateTo)
	return args.Get(0).(entity.SentimentCounts), args.Error(1)
}

func (m *MockDashboardRepository) GetBrandSourceStats(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) ([]entity.SourceCount, error) {
	args := m.Called(ctx, brandID, dateFrom, dateTo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.SourceCount), args.Error(1)
}

func (m *MockDashboardRepository) GetBrandDailyStats(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) ([]entity.DailyCount, error) {
	args := m.Called(ctx, brandID, dateFrom, dateTo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.DailyCount), args.Error(1)
}

func (m *MockDashboardRepository) GetBrandAlertCount(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) (int64, error) {
	args := m.Called(ctx, brandID, dateFrom, dateTo)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDashboardRepository) GetOverallSentiment(ctx context.Context, dateFrom, dateTo string) (entity.SentimentCounts, error) {
	args := m.Called(ctx, dateFrom, dateTo)
	return args.Get(0).(entity.SentimentCounts), args.Error(1)
}

func (m *MockDashboardRepository) GetOverallDailyStats(ctx context.Context, dateFrom, dateTo string) ([]entity.DailyCount, error) {
	args := m.Called(ctx, dateFrom, dateTo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.DailyCount), args.Error(1)
}

func (m *MockDashboardRepository) GetAllBrandsSummary(ctx context.Context, dateFrom, dateTo string) ([]entity.BrandSummary, error) {
	args := m.Called(ctx, dateFrom, dateTo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.BrandSummary), args.Error(1)
}

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) Notify(ctx context.Context, message string) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}
