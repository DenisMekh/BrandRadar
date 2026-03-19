package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/anomalies"
	"prod-pobeda-2026/internal/entity"
	"prod-pobeda-2026/internal/metrics"
	"prod-pobeda-2026/internal/notify"
)

// AlertUseCase — бизнес-логика spike-алертов.
type AlertUseCase struct {
	configRepo    AlertConfigRepository
	alertRepo     AlertRepository
	mentionRepo   MentionRepository
	brandRepo     BrandRepository
	cooldown      CooldownCache
	eventRepo     EventRepository
	metrics       metrics.Business
	anomalyWorker *anomalies.Worker
	notifier      notify.Notifier
}

// NewAlertUseCase — конструктор usecase алертов.
func NewAlertUseCase(
	configRepo AlertConfigRepository,
	alertRepo AlertRepository,
	mentionRepo MentionRepository,
	brandRepo BrandRepository,
	cooldown CooldownCache,
	eventRepo EventRepository,
	checkInterval time.Duration,
	notifier notify.Notifier,
	businessMetrics ...metrics.Business,
) *AlertUseCase {
	m := metrics.NopBusiness()
	if len(businessMetrics) > 0 && businessMetrics[0] != nil {
		m = businessMetrics[0]
	}

	uc := &AlertUseCase{
		configRepo:  configRepo,
		alertRepo:   alertRepo,
		mentionRepo: mentionRepo,
		brandRepo:   brandRepo,
		cooldown:    cooldown,
		eventRepo:   eventRepo,
		metrics:     m,
		notifier:    notifier,
	}

	// Создаём и запускаем воркер аномалий
	uc.anomalyWorker = anomalies.NewWorker(
		&alertConfigProviderAdapter{repo: configRepo},
		&mentionCounterAdapter{repo: mentionRepo},
		&alertCheckerAdapter{uc: uc},
		checkInterval,
	)

	return uc
}

// CreateConfig создаёт конфигурацию spike-алерта.
func (uc *AlertUseCase) CreateConfig(ctx context.Context, cfg *entity.AlertConfig) error {
	if cfg.BrandID == uuid.Nil {
		return fmt.Errorf("AlertUseCase.CreateConfig: brand_id is nil: %w", entity.ErrValidation)
	}
	if cfg.WindowMinutes <= 0 {
		return fmt.Errorf("AlertUseCase.CreateConfig: window_minutes must be > 0: %w", entity.ErrValidation)
	}
	if cfg.CooldownMinutes <= 0 {
		return fmt.Errorf("AlertUseCase.CreateConfig: cooldown_minutes must be > 0: %w", entity.ErrValidation)
	}

	cfg.ID = uuid.New()
	cfg.Enabled = true
	now := time.Now().UTC()
	cfg.CreatedAt = now
	cfg.UpdatedAt = now

	if err := uc.configRepo.Create(ctx, cfg); err != nil {
		return fmt.Errorf("AlertUseCase.CreateConfig: %w", err)
	}

	logrus.Infof("alert config created: brand_id=%s", cfg.BrandID.String())
	return nil
}

// GetConfig возвращает конфигурацию алерта по идентификатору.
func (uc *AlertUseCase) GetConfig(ctx context.Context, id uuid.UUID) (*entity.AlertConfig, error) {
	cfg, err := uc.configRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("AlertUseCase.GetConfig: %w", err)
	}
	return cfg, nil
}

// GetConfigByBrand возвращает конфигурацию алерта по идентификатору бренда.
func (uc *AlertUseCase) GetConfigByBrand(ctx context.Context, brandID uuid.UUID) (*entity.AlertConfig, error) {
	cfg, err := uc.configRepo.GetByBrandID(ctx, brandID)
	if err != nil {
		return nil, fmt.Errorf("AlertUseCase.GetConfigByBrand: %w", err)
	}
	return cfg, nil
}

// UpdateConfig обновляет конфигурацию алерта.
func (uc *AlertUseCase) UpdateConfig(ctx context.Context, cfg *entity.AlertConfig) error {
	if cfg.WindowMinutes <= 0 {
		return fmt.Errorf("AlertUseCase.UpdateConfig: window_minutes must be > 0: %w", entity.ErrValidation)
	}
	if cfg.CooldownMinutes <= 0 {
		return fmt.Errorf("AlertUseCase.UpdateConfig: cooldown_minutes must be > 0: %w", entity.ErrValidation)
	}

	cfg.UpdatedAt = time.Now().UTC()
	if err := uc.configRepo.Update(ctx, cfg); err != nil {
		return fmt.Errorf("AlertUseCase.UpdateConfig: %w", err)
	}
	return nil
}

// DeleteConfig удаляет конфигурацию алерта.
func (uc *AlertUseCase) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	if err := uc.configRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("AlertUseCase.DeleteConfig: %w", err)
	}
	return nil
}

// ListAllConfigs возвращает все конфигурации алертов.
func (uc *AlertUseCase) ListAllConfigs(ctx context.Context) ([]*entity.AlertConfig, error) {
	configs, err := uc.configRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("AlertUseCase.ListAllConfigs: %w", err)
	}
	return configs, nil
}

// CheckSpike проверяет наличие аномалии с помощью confidence interval анализа.
func (uc *AlertUseCase) CheckSpike(ctx context.Context, brandID uuid.UUID) error {
	return uc.anomalyWorker.CheckBrand(ctx, brandID)
}

// CheckAndFire создаёт алерт при обнаружении аномалии.
func (uc *AlertUseCase) CheckAndFire(ctx context.Context, brandID uuid.UUID) error {
	cfg, err := uc.configRepo.GetByBrandID(ctx, brandID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("AlertUseCase.CheckAndFire: get config: %w", err)
	}
	if !cfg.Enabled {
		return nil
	}

	now := time.Now().UTC()
	windowStart := now.Add(-time.Duration(cfg.WindowMinutes) * time.Minute)

	count, err := uc.mentionRepo.CountByBrandSince(ctx, brandID, windowStart, cfg.SentimentFilter)
	if err != nil {
		return fmt.Errorf("AlertUseCase.CheckAndFire: count: %w", err)
	}

	acquired, err := uc.cooldown.TryLock(ctx, cfg.ID, cfg.CooldownMinutes)
	if err != nil {
		return fmt.Errorf("AlertUseCase.CheckAndFire: cooldown: %w", err)
	}
	if !acquired {
		logrus.Debugf("spike cooldown active: brand_id=%s", brandID.String())
		return nil
	}

	alert := &entity.Alert{
		ID:            uuid.New(),
		ConfigID:      cfg.ID,
		BrandID:       brandID,
		MentionsCount: int(count),
		WindowStart:   windowStart,
		WindowEnd:     now,
		FiredAt:       now,
		CreatedAt:     now,
	}

	if err := uc.alertRepo.Create(ctx, alert); err != nil {
		return fmt.Errorf("AlertUseCase.CheckAndFire: create alert: %w", err)
	}
	uc.metrics.IncAlertsFired(brandID.String())

	payload, err := json.Marshal(map[string]any{
		"alert_id":       alert.ID.String(),
		"brand_id":       brandID.String(),
		"mentions_count": count,
	})
	if err != nil {
		logrus.Warnf("failed to serialize alert event: %v", err)
		payload = []byte("{}")
	}
	if err := uc.eventRepo.Create(ctx, &entity.Event{
		ID:         uuid.New(),
		Type:       entity.EventAlertFired,
		Payload:    payload,
		OccurredAt: now,
	}); err != nil {
		logrus.Warnf("failed to write alert event: %v", err)
	}

	logrus.Warnf("spike alert fired: brand_id=%s, count=%d", brandID.String(), count)

	brandName := brandID.String()
	if brand, err := uc.brandRepo.GetByID(ctx, brandID); err == nil {
		brandName = brand.Name
	}
	msg := fmt.Sprintf("🚨 Spike alert: <b>%s</b>", brandName)
	if uc.notifier != nil {
		if notifyErr := uc.notifier.Notify(ctx, msg); notifyErr != nil {
			logrus.Warnf("alert notify failed: %v", notifyErr)
		}
	}

	return nil
}

// StartAnomalyWorker запускает воркер проверки аномалий.
func (uc *AlertUseCase) StartAnomalyWorker(ctx context.Context) {
	uc.anomalyWorker.Start(ctx)
}

// StopAnomalyWorker останавливает воркер проверки аномалий.
func (uc *AlertUseCase) StopAnomalyWorker() {
	uc.anomalyWorker.Stop()
}

// alertConfigProviderAdapter адаптирует AlertConfigRepository для anomalies.AlertConfigProvider.
type alertConfigProviderAdapter struct {
	repo AlertConfigRepository
}

func (a *alertConfigProviderAdapter) GetByBrandID(ctx context.Context, brandID uuid.UUID) (*anomalies.AlertConfig, error) {
	cfg, err := a.repo.GetByBrandID(ctx, brandID)
	if err != nil {
		return nil, err
	}
	return &anomalies.AlertConfig{
		ID:                cfg.ID,
		BrandID:           cfg.BrandID,
		WindowMinutes:     cfg.WindowMinutes,
		CooldownMinutes:   cfg.CooldownMinutes,
		SentimentFilter:   cfg.SentimentFilter,
		Enabled:           cfg.Enabled,
		Percentile:        cfg.Percentile,
		AnomalyWindowSize: cfg.AnomalyWindowSize,
	}, nil
}

func (a *alertConfigProviderAdapter) GetAllActive(ctx context.Context) ([]*anomalies.AlertConfig, error) {
	configs, err := a.repo.GetAllActive(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*anomalies.AlertConfig, 0, len(configs))
	for _, cfg := range configs {
		result = append(result, &anomalies.AlertConfig{
			ID:                cfg.ID,
			BrandID:           cfg.BrandID,
			WindowMinutes:     cfg.WindowMinutes,
			CooldownMinutes:   cfg.CooldownMinutes,
			SentimentFilter:   cfg.SentimentFilter,
			Enabled:           cfg.Enabled,
			Percentile:        cfg.Percentile,
			AnomalyWindowSize: cfg.AnomalyWindowSize,
		})
	}
	return result, nil
}

// mentionCounterAdapter адаптирует MentionRepository для anomalies.MentionCounter.
type mentionCounterAdapter struct {
	repo MentionRepository
}

func (a *mentionCounterAdapter) CountByBrandSince(ctx context.Context, brandID uuid.UUID, since time.Time, sentimentFilter string) (int64, error) {
	return a.repo.CountByBrandSince(ctx, brandID, since, sentimentFilter)
}

func (a *mentionCounterAdapter) CountByBrandInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time, sentimentFilter string) (int64, error) {
	return a.repo.CountByBrandInRange(ctx, brandID, from, to, sentimentFilter)
}

func (a *mentionCounterAdapter) GetWeightedSentimentPercentileInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time, percentile float64) (float64, error) {
	return a.repo.GetWeightedSentimentPercentileInRange(ctx, brandID, from, to, percentile)
}

// alertCheckerAdapter адаптирует AlertUseCase для anomalies.AlertChecker.
type alertCheckerAdapter struct {
	uc *AlertUseCase
}

func (a *alertCheckerAdapter) CheckAndFire(ctx context.Context, brandID uuid.UUID) error {
	return a.uc.CheckAndFire(ctx, brandID)
}

// ListAllAlerts возвращает все сработавшие алерты (без фильтра по бренду).
func (uc *AlertUseCase) ListAllAlerts(ctx context.Context, limit, offset int) ([]entity.Alert, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := uc.alertRepo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("AlertUseCase.ListAllAlerts: %w", err)
	}
	return items, int(total), nil
}

// ListAlerts возвращает список сработавших алертов бренда.
func (uc *AlertUseCase) ListAlerts(ctx context.Context, brandID uuid.UUID, limit, offset int) ([]entity.Alert, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := uc.alertRepo.ListByBrandID(ctx, brandID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("AlertUseCase.ListAlerts: %w", err)
	}
	return items, int(total), nil
}
