package dto

// CreateAlertConfigRequest — запрос на создание конфигурации алерта.
type CreateAlertConfigRequest struct {
	BrandID           string  `json:"brand_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	WindowMinutes     int     `json:"window_minutes" binding:"required,min=1" example:"60"`
	CooldownMinutes   int     `json:"cooldown_minutes" binding:"required,min=1" example:"30"`
	SentimentFilter   string  `json:"sentiment_filter" example:"negative"`
	Percentile        float64 `json:"percentile" binding:"required,min=0,max=100" example:"95"`
	AnomalyWindowSize int     `json:"anomaly_window_size" binding:"required,min=3" example:"10"`
}

// UpdateAlertConfigRequest — запрос на обновление конфигурации алерта.
type UpdateAlertConfigRequest struct {
	WindowMinutes     *int     `json:"window_minutes" binding:"omitempty,min=1" example:"60"`
	CooldownMinutes   *int     `json:"cooldown_minutes" binding:"omitempty,min=1" example:"30"`
	SentimentFilter   *string  `json:"sentiment_filter" example:"negative"`
	Enabled           *bool    `json:"enabled" example:"true"`
	Percentile        *float64 `json:"percentile" binding:"omitempty,min=0,max=100" example:"95"`
	AnomalyWindowSize *int     `json:"anomaly_window_size" binding:"omitempty,min=3" example:"10"`
}

// AlertConfigResponse — конфигурация алерта в ответе API.
type AlertConfigResponse struct {
	ID                string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BrandID           string  `json:"brand_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	WindowMinutes     int     `json:"window_minutes" example:"60"`
	CooldownMinutes   int     `json:"cooldown_minutes" example:"30"`
	SentimentFilter   string  `json:"sentiment_filter" example:"negative"`
	Enabled           bool    `json:"enabled" example:"true"`
	Percentile        float64 `json:"percentile" example:"95"`
	AnomalyWindowSize int     `json:"anomaly_window_size" example:"10"`
	CreatedAt         string  `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt         string  `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}

// AlertResponse — сработавший алерт в ответе API.
type AlertResponse struct {
	ID            string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ConfigID      string `json:"config_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BrandID       string `json:"brand_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MentionsCount int    `json:"mentions_count" example:"15"`
	WindowStart   string `json:"window_start" example:"2024-01-01T11:00:00Z"`
	WindowEnd     string `json:"window_end" example:"2024-01-01T12:00:00Z"`
	FiredAt       string `json:"fired_at" example:"2024-01-01T12:00:00Z"`
	CreatedAt     string `json:"created_at" example:"2024-01-01T12:00:00Z"`
}
