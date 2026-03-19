package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"prod-pobeda-2026/internal/entity"
)

// BrandRepository — контракт хранилища брендов.
type BrandRepository interface {
	Create(ctx context.Context, brand *entity.Brand) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Brand, error)
	List(ctx context.Context, limit, offset int) ([]entity.Brand, int64, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, brand *entity.Brand) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SourceRepository — контракт хранилища источников.
type SourceRepository interface {
	Create(ctx context.Context, source *entity.Source) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Source, error)
	List(ctx context.Context, limit, offset int) ([]entity.Source, int64, error)
	CountActiveByType(ctx context.Context) (map[string]int64, error)
	ListActive(ctx context.Context) ([]entity.Source, error)
	Update(ctx context.Context, source *entity.Source) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// MentionFilter — алиас из entity для обратной совместимости.
type MentionFilter = entity.MentionFilter

// MentionRepository — контракт хранилища упоминаний.
type MentionRepository interface {
	Create(ctx context.Context, mention *entity.Mention) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Mention, error)
	List(ctx context.Context, filter MentionFilter) ([]entity.Mention, int64, error)
	// CountByBrandSince считает упоминания бренда с момента since (для spike detection).
	CountByBrandSince(ctx context.Context, brandID uuid.UUID, since time.Time, sentimentFilter string) (int64, error)
	// CountByBrandInRange считает упоминания бренда в конкретном временном диапазоне (для anomaly detection).
	CountByBrandInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time, sentimentFilter string) (int64, error)
	// GetAverageSentimentInRange возвращает средний сентимент бренда в диапазоне (для anomaly detection).
	// Sentiment кодируется: positive=1, neutral=0, negative=-1, умножается на confidence.
	GetAverageSentimentInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time) (float64, error)
	// GetWeightedSentimentPercentileInRange возвращает перцентиль взвешенного сентимента бренда в диапазоне.
	// Sentiment кодируется: positive=1, neutral=0, negative=-1, умножается на confidence от ML.
	GetWeightedSentimentPercentileInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time, percentile float64) (float64, error)
	// GetSimilarMentions возвращает до limit упоминаний того же кластера (по brand_id + cluster_label), исключая excludeID.
	GetSimilarMentions(ctx context.Context, brandID uuid.UUID, clusterLabel int, excludeID uuid.UUID, limit int) ([]entity.Mention, error)
}

// AlertConfigRepository — контракт хранилища настроек алертов.
type AlertConfigRepository interface {
	Create(ctx context.Context, config *entity.AlertConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AlertConfig, error)
	GetByBrandID(ctx context.Context, brandID uuid.UUID) (*entity.AlertConfig, error)
	Update(ctx context.Context, config *entity.AlertConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
	// GetAllActive возвращает все активные конфигурации алертов.
	GetAllActive(ctx context.Context) ([]*entity.AlertConfig, error)
	// ListAll возвращает все конфигурации алертов.
	ListAll(ctx context.Context) ([]*entity.AlertConfig, error)
}

// AlertRepository — контракт хранилища сработавших алертов.
type AlertRepository interface {
	Create(ctx context.Context, alert *entity.Alert) error
	ListByBrandID(ctx context.Context, brandID uuid.UUID, limit, offset int) ([]entity.Alert, int64, error)
	// ListAll возвращает все алерты с пагинацией (без фильтра по бренду).
	ListAll(ctx context.Context, limit, offset int) ([]entity.Alert, int64, error)
	// GetLastFiredAt возвращает время последнего сработавшего алерта для конфига.
	GetLastFiredAt(ctx context.Context, configID uuid.UUID) (*time.Time, error)
}

// EventRepository — контракт журнала событий (append-only).
type EventRepository interface {
	Create(ctx context.Context, event *entity.Event) error
	List(ctx context.Context, eventType *string, limit, offset int) ([]entity.Event, int64, error)
}

// CooldownCache — контракт кэша для cooldown алертов (Redis).
type CooldownCache interface {
	// TryLock пытается установить cooldown. Возвращает true если cooldown свободен и установлен.
	TryLock(ctx context.Context, configID uuid.UUID, cooldownMinutes int) (bool, error)
}

// MentionCache — контракт кэша списков упоминаний (Redis cache-aside).
type MentionCache interface {
	GetList(ctx context.Context, filter MentionFilter) ([]entity.Mention, int64, error)
	SetList(ctx context.Context, filter MentionFilter, items []entity.Mention, total int64) error
	InvalidateByBrand(ctx context.Context, brandID uuid.UUID) error
	InvalidateAll(ctx context.Context) error
}

// HealthChecker — контракт проверки доступности внешней зависимости.
type HealthChecker interface {
	Ping(ctx context.Context) error
	Name() string
}

// DashboardRepository — контракт для агрегированных запросов дашборда.
type DashboardRepository interface {
	GetBrandSentiment(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) (entity.SentimentCounts, error)
	GetBrandSourceStats(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) ([]entity.SourceCount, error)
	GetBrandDailyStats(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) ([]entity.DailyCount, error)
	GetBrandAlertCount(ctx context.Context, brandID uuid.UUID, dateFrom, dateTo string) (int64, error)
	GetOverallSentiment(ctx context.Context, dateFrom, dateTo string) (entity.SentimentCounts, error)
	GetOverallDailyStats(ctx context.Context, dateFrom, dateTo string) ([]entity.DailyCount, error)
	GetAllBrandsSummary(ctx context.Context, dateFrom, dateTo string) ([]entity.BrandSummary, error)
}

// AnalyticsRepository — контракт хранилища для analytics worker.
type AnalyticsRepository interface {
	BatchInsertItems(ctx context.Context, items []entity.CrawlerItem, results []entity.SentimentMLResult) error
}

// ClusteringRepository — контракт хранилища для clustering worker.
type ClusteringRepository interface {
	// GetItemsByBrandID возвращает все crawler_items, связанные с брендом через sentiment_results.
	GetItemsByBrandID(ctx context.Context, brandID uuid.UUID) ([]entity.CrawlerItemText, error)
	// UpdateClusterLabels проставляет cluster_label в sentiment_results для каждого item бренда.
	// nil в assignments означает noise (cluster_label = NULL).
	UpdateClusterLabels(ctx context.Context, brandID uuid.UUID, assignments map[uuid.UUID]*int) error
}
