package entity

import (
	"time"

	"github.com/google/uuid"
)

// CrawlerItem — item из crawler для сохранения в БД.
type CrawlerItem struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Text        string     `json:"text" db:"text"`
	Link        string     `json:"link" db:"link"`
	PublishedAt time.Time  `json:"published_at" db:"published_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	SourceID    *uuid.UUID `json:"source_id" db:"source_id"`
}

// CrawlerItemText — облегчённая проекция crawler_item для кластеризации.
type CrawlerItemText struct {
	ID   uuid.UUID `db:"id"`
	Text string    `db:"text"`
}

// SentimentMLResult — результат ML анализа.
// store in db sentiment_ml result, itemId fkey, brandId fkey
type SentimentMLResult struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ItemID     uuid.UUID `json:"item_id" db:"item_id"`     // fkey -> CrawlerItem
	BrandID    uuid.UUID `json:"brand_id" db:"brand_id"`   // fkey -> Brand
	Sentiment  string    `json:"sentiment" db:"sentiment"` // positive/negative/neutral
	Confidence float64   `json:"confidence" db:"confidence"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
