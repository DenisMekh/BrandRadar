package anomalies

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AlertChecker — интерфейс для проверки алертов.

type AlertChecker interface {
	CheckAndFire(ctx context.Context, brandID uuid.UUID) error
}

// MentionCounter — интерфейс для подсчёта упоминаний и взвешенного сентимента.

type MentionCounter interface {
	CountByBrandSince(ctx context.Context, brandID uuid.UUID, since time.Time, sentimentFilter string) (int64, error)

	CountByBrandInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time, sentimentFilter string) (int64, error)

	// GetWeightedSentimentPercentileInRange возвращает перцентиль взвешенного сентимента (sentiment * confidence) за период.

	GetWeightedSentimentPercentileInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time, percentile float64) (float64, error)
}

// AlertConfigProvider — интерфейс для получения конфигов алертов.

type AlertConfigProvider interface {
	GetByBrandID(ctx context.Context, brandID uuid.UUID) (*AlertConfig, error)

	GetAllActive(ctx context.Context) ([]*AlertConfig, error)
}

// AlertConfig — конфигурация алерта.

type AlertConfig struct {
	ID uuid.UUID

	BrandID uuid.UUID

	WindowMinutes int

	CooldownMinutes int

	SentimentFilter string

	Enabled bool

	Percentile float64

	AnomalyWindowSize int
}

// Worker — воркер для проверки аномалий в упоминаниях брендов.

type Worker struct {
	configProvider AlertConfigProvider

	mentionCounter MentionCounter

	alertChecker AlertChecker

	checkInterval time.Duration

	mu sync.RWMutex

	running bool

	stopCh chan struct{}
}

// NewWorker создаёт новый экземпляр воркера.

func NewWorker(
	configProvider AlertConfigProvider,

	mentionCounter MentionCounter,

	alertChecker AlertChecker,

	checkInterval time.Duration,
) *Worker {
	return &Worker{
		configProvider: configProvider,

		mentionCounter: mentionCounter,

		alertChecker: alertChecker,

		checkInterval: checkInterval,

		stopCh: make(chan struct{}),
	}
}

// Start запускает воркер в горутине.

func (w *Worker) Start(ctx context.Context) {
	w.mu.Lock()

	if w.running {

		w.mu.Unlock()

		return

	}

	w.running = true

	w.mu.Unlock()

	go w.run(ctx)
}

// Stop останавливает воркер.

func (w *Worker) Stop() {
	w.mu.Lock()

	defer w.mu.Unlock()

	if !w.running {
		return
	}

	w.running = false

	close(w.stopCh)

	w.stopCh = make(chan struct{})
}

// IsRunning возвращает статус воркера.

func (w *Worker) IsRunning() bool {
	w.mu.RLock()

	defer w.mu.RUnlock()

	return w.running
}

func (w *Worker) run(ctx context.Context) {
	ticker := time.NewTicker(w.checkInterval)

	defer ticker.Stop()

	logrus.Info("anomaly worker started")

	for {
		select {

		case <-ctx.Done():

			logrus.Info("anomaly worker stopped by context")

			return

		case <-w.stopCh:

			logrus.Info("anomaly worker stopped")

			return

		case <-ticker.C:

			w.checkAllBrands(ctx)

		}
	}
}

func (w *Worker) checkAllBrands(ctx context.Context) {
	logrus.Debug("anomaly worker: checking all brands")

	// Получаем все активные конфиги алертов

	configs, err := w.configProvider.GetAllActive(ctx)
	if err != nil {

		logrus.Errorf("anomaly worker: failed to get active configs: %v", err)

		return

	}

	logrus.Debugf("anomaly worker: found %d active configs", len(configs))

	// Проверяем каждый бренд на аномалии

	for _, cfg := range configs {

		if !cfg.Enabled {
			continue
		}

		if err := w.CheckBrand(ctx, cfg.BrandID); err != nil {
			logrus.Errorf("anomaly worker: check brand %s: %v", cfg.BrandID, err)
		}

	}
}

// CheckBrand проверяет конкретный бренд на аномалии.

func (w *Worker) CheckBrand(ctx context.Context, brandID uuid.UUID) error {
	cfg, err := w.configProvider.GetByBrandID(ctx, brandID)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	if !cfg.Enabled {
		return nil
	}

	// Собираем данные за период для анализа

	now := time.Now().UTC()

	totalWindow := time.Duration(cfg.WindowMinutes*cfg.AnomalyWindowSize) * time.Minute

	windowStart := now.Add(-totalWindow)

	counts, err := w.collectDataPoints(ctx, brandID, cfg, windowStart, now)
	if err != nil {
		return fmt.Errorf("collect data points: %w", err)
	}

	if len(counts) < cfg.AnomalyWindowSize {

		logrus.Debugf("not enough data points for brand %s: got %d, need %d", brandID, len(counts), cfg.AnomalyWindowSize)

		return nil

	}

	// Анализируем последнюю точку на аномалию

	results := Analyze(counts, cfg.AnomalyWindowSize-1)

	if len(results) == 0 {
		return nil
	}

	lastResult := results[len(results)-1]

	if lastResult.IsAnom {

		logrus.Warnf("anomaly detected for brand %s: value=%.2f, interval=[%.2f, %.2f], z=%.2f",

			brandID, lastResult.Value, lastResult.Interval.Lower, lastResult.Interval.Upper, lastResult.Z)

		// Фирим алерт

		return w.alertChecker.CheckAndFire(ctx, brandID)

	}

	logrus.Debugf("no anomaly for brand %s: value=%.2f, interval=[%.2f, %.2f]",

		brandID, lastResult.Value, lastResult.Interval.Lower, lastResult.Interval.Upper)

	return nil
}

// collectDataPoints собирает точки данных для анализа.

// Для каждого окна вычисляет перцентиль взвешенного сентимента (sentiment * confidence).

// Sentiment кодируется: positive=1, neutral=0, negative=-1.

func (w *Worker) collectDataPoints(
	ctx context.Context,

	brandID uuid.UUID,

	cfg *AlertConfig,

	start, end time.Time,
) ([]float64, error) {
	interval := time.Duration(cfg.WindowMinutes) * time.Minute

	points := make([]float64, 0, cfg.AnomalyWindowSize)

	// Собираем точки от конца к началу (от текущего окна к прошлому)

	for t := end; t.After(start); t = t.Add(-interval) {

		windowEnd := t

		windowStart := t.Add(-interval)

		// Debug: выводим границы окна

		logrus.Debugf("collectDataPoints: window=[%s, %s], brandID=%s",

			windowStart.Format(time.RFC3339), windowEnd.Format(time.RFC3339), brandID)

		// Получаем перцентиль взвешенного сентимента за окно

		percentileValue, err := w.mentionCounter.GetWeightedSentimentPercentileInRange(ctx, brandID, windowStart, windowEnd, cfg.Percentile)
		if err != nil {
			return nil, fmt.Errorf("get weighted sentiment percentile in range [%s, %s]: %w", windowStart, windowEnd, err)
		}

		points = append(points, percentileValue)

	}

	// Реверсируем, чтобы точки шли от старых к новым

	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return points, nil
}
