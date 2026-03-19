package entity

import (
	"time"

	"github.com/google/uuid"
)

type Sentiment string

const (
	SENTIMENT_POSITIVE Sentiment = "positive"
	SENTIMENT_NEGATIVE Sentiment = "negative"
	SENTIMENT_NEUTRAL  Sentiment = "neutral"
)

type SentimentMLOutput struct {
	Sentiment  Sentiment `json:"sentiment" db:"sentiment"`
	Confidence float64   `json:"confidence" db:"confidence"`
}

// SentimentToNumeric преобразует сентимент в числовое значение.
// positive = 1, neutral = 0, negative = -1
func SentimentToNumeric(s Sentiment) float64 {
	switch s {
	case SENTIMENT_POSITIVE:
		return 1.0
	case SENTIMENT_NEGATIVE:
		return -1.0
	case SENTIMENT_NEUTRAL:
		return 0.0
	default:
		return 0.0
	}
}

// MentionStatus — статус обработки упоминания.
type MentionStatus string

const (
	MentionStatusNew        MentionStatus = "new"
	MentionStatusProcessing MentionStatus = "processing"
	MentionStatusProcessed  MentionStatus = "processed"
	MentionStatusArchived   MentionStatus = "archived"
	MentionStatusDiscarded  MentionStatus = "discarded"
)

// validTransitions — допустимые переходы состояний для упоминаний.
var validTransitions = map[MentionStatus][]MentionStatus{
	MentionStatusNew:        {MentionStatusProcessing, MentionStatusDiscarded},
	MentionStatusProcessing: {MentionStatusProcessed},
	MentionStatusProcessed:  {MentionStatusArchived},
}

// Mention — упоминание бренда, собранное из crawler_items + sentiment_results + sources.
type Mention struct {
	ID           uuid.UUID     `json:"id" db:"id"`
	BrandID      uuid.UUID     `json:"brand_id" db:"brand_id"`
	SourceID     uuid.UUID     `json:"source_id" db:"source_id"`
	ExternalID   string        `json:"external_id" db:"external_id"`
	SourceName   string        `json:"source_name" db:"source_name"`
	SourceType   string        `json:"source_type" db:"source_type"`
	Title        string        `json:"title" db:"title"`
	Text         string        `json:"text" db:"text"`
	URL          string        `json:"url" db:"url"`
	Author       string        `json:"author" db:"author"`
	Sentiment    Sentiment     `json:"sentiment" db:"sentiment"`
	Status       MentionStatus `json:"status" db:"status"`
	Deduplicated bool          `json:"deduplicated" db:"deduplicated"`
	ML           MentionML     `json:"ml" db:"-"`
	PublishedAt  time.Time     `json:"published_at" db:"published_at"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
	ClusterLabel *int          `json:"-" db:"-"` // cluster_label из sentiment_results
	Similar      []Mention     `json:"-" db:"-"` // похожие упоминания того же кластера
}

// MentionML — ML-поля упоминания.
type MentionML struct {
	Label      string   `json:"label" db:"ml_label"`
	Score      float64  `json:"score" db:"ml_score"`
	IsRelevant bool     `json:"is_relevant" db:"ml_is_relevant"`
	SimilarIDs []string `json:"similar_ids" db:"ml_similar_ids"`
}

// MentionFilter — фильтры для поиска упоминаний в ленте.
type MentionFilter struct {
	BrandID   uuid.UUID `json:"brand_id"`
	SourceID  uuid.UUID `json:"source_id"`
	Sentiment string    `json:"sentiment"`
	Search    string    `json:"search"`
	DateFrom  string    `json:"date_from"`
	DateTo    string    `json:"date_to"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}
