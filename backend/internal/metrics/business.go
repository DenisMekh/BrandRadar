package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	StatusCreated    = "created"
	StatusDuplicated = "duplicated"
	StatusSkipped    = "skipped"
	StatusError      = "error"
	StatusSuccess    = "success"
)

// Business описывает бизнес-метрики BrandRadar.
type Business interface {
	IncMentionsCreated(brandID, source string)
	IncAlertsFired(brandID string)
	ObserveIngestDuration(seconds float64)
	AddIngestItems(status string, value float64)
	SetSourcesActive(sourceType string, count float64)
	SetBrandsTotal(count float64)
	IncAPIErrors(endpoint, errorCode string)
}

type businessMetrics struct {
	mentionsCreatedTotal *prometheus.CounterVec
	alertsFiredTotal     *prometheus.CounterVec
	ingestDuration       prometheus.Histogram
	ingestItemsTotal     *prometheus.CounterVec
	sourcesActiveTotal   *prometheus.GaugeVec
	brandsTotal          prometheus.Gauge
	apiErrorsTotal       *prometheus.CounterVec
}

// NewBusiness создаёт и регистрирует бизнес-метрики.
func NewBusiness(registerer prometheus.Registerer) Business {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}

	mentionsCreatedTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "brandradar_mentions_created_total",
			Help: "Total number of mentions created by ingest flow.",
		},
		[]string{"brand_id", "source"},
	)
	alertsFiredTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "brandradar_alerts_fired_total",
			Help: "Total number of fired spike alerts.",
		},
		[]string{"brand_id"},
	)
	ingestDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "brandradar_ingest_duration_seconds",
			Help:    "Duration of ingest usecase execution in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)
	ingestItemsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "brandradar_ingest_items_total",
			Help: "Total number of ingest outcomes by status.",
		},
		[]string{"status"},
	)
	sourcesActiveTotal := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brandradar_sources_active_total",
			Help: "Number of active sources grouped by source type.",
		},
		[]string{"type"},
	)
	brandsTotal := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "brandradar_brands_total",
			Help: "Total number of brands.",
		},
	)
	apiErrorsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "brandradar_api_errors_total",
			Help: "Total number of API errors grouped by endpoint and error code.",
		},
		[]string{"endpoint", "error_code"},
	)
	registerer.MustRegister(
		mentionsCreatedTotal,
		alertsFiredTotal,
		ingestDuration,
		ingestItemsTotal,
		sourcesActiveTotal,
		brandsTotal,
		apiErrorsTotal,
	)

	return &businessMetrics{
		mentionsCreatedTotal: mentionsCreatedTotal,
		alertsFiredTotal:     alertsFiredTotal,
		ingestDuration:       ingestDuration,
		ingestItemsTotal:     ingestItemsTotal,
		sourcesActiveTotal:   sourcesActiveTotal,
		brandsTotal:          brandsTotal,
		apiErrorsTotal:       apiErrorsTotal,
	}
}

func (m *businessMetrics) IncMentionsCreated(brandID, source string) {
	m.mentionsCreatedTotal.WithLabelValues(brandID, source).Inc()
}

func (m *businessMetrics) IncAlertsFired(brandID string) {
	m.alertsFiredTotal.WithLabelValues(brandID).Inc()
}

func (m *businessMetrics) ObserveIngestDuration(seconds float64) {
	m.ingestDuration.Observe(seconds)
}

func (m *businessMetrics) AddIngestItems(status string, value float64) {
	if value <= 0 {
		return
	}
	m.ingestItemsTotal.WithLabelValues(status).Add(value)
}

func (m *businessMetrics) SetSourcesActive(sourceType string, count float64) {
	m.sourcesActiveTotal.WithLabelValues(sourceType).Set(count)
}

func (m *businessMetrics) SetBrandsTotal(count float64) {
	m.brandsTotal.Set(count)
}

func (m *businessMetrics) IncAPIErrors(endpoint, errorCode string) {
	m.apiErrorsTotal.WithLabelValues(endpoint, errorCode).Inc()
}

type noopBusiness struct{}

// NopBusiness возвращает no-op реализацию бизнес-метрик.
func NopBusiness() Business {
	return noopBusiness{}
}

func (noopBusiness) IncMentionsCreated(string, string) {}

func (noopBusiness) IncAlertsFired(string) {}

func (noopBusiness) ObserveIngestDuration(float64) {}

func (noopBusiness) AddIngestItems(string, float64) {}

func (noopBusiness) SetSourcesActive(string, float64) {}

func (noopBusiness) SetBrandsTotal(float64) {}

func (noopBusiness) IncAPIErrors(string, string) {}
