package entity

import (
	"time"

	"github.com/google/uuid"
)

// AlertConfig — настройки spike-алерта для бренда.
type AlertConfig struct {
	ID                uuid.UUID `json:"id" db:"id"`
	BrandID           uuid.UUID `json:"brand_id" db:"brand_id"`
	WindowMinutes     int       `json:"window_minutes" db:"window_minutes"`
	CooldownMinutes   int       `json:"cooldown_minutes" db:"cooldown_minutes"`
	SentimentFilter   string    `json:"sentiment_filter" db:"sentiment_filter"`
	Enabled           bool      `json:"enabled" db:"enabled"`
	Percentile        float64   `json:"percentile" db:"percentile"`                   // перцентиль для анализа аномалий (0-100)
	AnomalyWindowSize int       `json:"anomaly_window_size" db:"anomaly_window_size"` // размер окна для анализа (кол-во точек)
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Alert — сработавший spike-алерт.
type Alert struct {
	ID            uuid.UUID `json:"id" db:"id"`
	ConfigID      uuid.UUID `json:"config_id" db:"config_id"`
	BrandID       uuid.UUID `json:"brand_id" db:"brand_id"`
	MentionsCount int       `json:"mentions_count" db:"mentions_count"`
	WindowStart   time.Time `json:"window_start" db:"window_start"`
	WindowEnd     time.Time `json:"window_end" db:"window_end"`
	FiredAt       time.Time `json:"fired_at" db:"fired_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
