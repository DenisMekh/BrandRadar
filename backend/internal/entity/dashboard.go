package entity

import "github.com/google/uuid"

// SentimentCounts — количество упоминаний по тональности.
type SentimentCounts struct {
	Positive int64
	Negative int64
	Neutral  int64
}

// Total возвращает общее кол-во упоминаний.
func (s SentimentCounts) Total() int64 {
	return s.Positive + s.Negative + s.Neutral
}

// SourceCount — кол-во упоминаний по одному источнику.
type SourceCount struct {
	Source string
	Count  int64
}

// DailyCount — разбивка по дню.
type DailyCount struct {
	Date     string // формат "2006-01-02"
	Positive int64
	Negative int64
	Neutral  int64
}

// BrandDashboard — агрегированная статистика по бренду.
type BrandDashboard struct {
	BrandID      uuid.UUID
	BrandName    string
	Sentiment    SentimentCounts
	BySources    []SourceCount
	ByDate       []DailyCount
	RecentAlerts int64
}

// BrandSummary — краткая статистика бренда.
type BrandSummary struct {
	BrandID   uuid.UUID
	BrandName string
	Sentiment SentimentCounts
}
