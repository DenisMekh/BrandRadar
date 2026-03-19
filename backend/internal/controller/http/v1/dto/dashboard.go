package dto

import "github.com/google/uuid"

// SentimentStats — количество упоминаний по тональности.
type SentimentStats struct {
	Positive int64 `json:"positive" example:"120"`
	Negative int64 `json:"negative" example:"89"`
	Neutral  int64 `json:"neutral" example:"133"`
}

// SourceStats — количество упоминаний по источнику.
type SourceStats struct {
	Source string `json:"source" example:"telegram"`
	Count  int64  `json:"count" example:"150"`
}

// DailyStats — разбивка по дням с тональностью.
type DailyStats struct {
	Date     string `json:"date" example:"2026-03-14"`
	Positive int64  `json:"positive" example:"40"`
	Negative int64  `json:"negative" example:"30"`
	Neutral  int64  `json:"neutral" example:"50"`
	Total    int64  `json:"total" example:"120"`
}

// BrandDashboardResponse — статистика по одному бренду.
type BrandDashboardResponse struct {
	BrandID       uuid.UUID      `json:"brand_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BrandName     string         `json:"brand_name" example:"Tesla"`
	TotalMentions int64          `json:"total_mentions" example:"342"`
	Sentiment     SentimentStats `json:"sentiment"`
	BySources     []SourceStats  `json:"by_source"`
	ByDate        []DailyStats   `json:"by_date"`
	RecentAlerts  int64          `json:"recent_alerts" example:"3"`
}

// BrandSummary — краткая статистика бренда для общего дашборда.
type BrandSummary struct {
	BrandID       uuid.UUID      `json:"brand_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BrandName     string         `json:"brand_name" example:"Tesla"`
	TotalMentions int64          `json:"total_mentions" example:"342"`
	Sentiment     SentimentStats `json:"sentiment"`
}

// OverallDashboardResponse — общий дашборд по всем брендам.
type OverallDashboardResponse struct {
	TotalMentions int64          `json:"total_mentions" example:"1200"`
	Sentiment     SentimentStats `json:"sentiment"`
	Brands        []BrandSummary `json:"brands"`
	ByDate        []DailyStats   `json:"by_date"`
}
