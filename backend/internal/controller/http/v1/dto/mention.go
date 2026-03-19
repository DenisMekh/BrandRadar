package dto

import "github.com/google/uuid"

// SourceRef — краткая информация об источнике внутри упоминания.
type SourceRef struct {
	ID   string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name string `json:"name" example:"Медуза"`
	Type string `json:"type" example:"telegram"`
}

// SimilarMention — краткое представление похожего упоминания.
type SimilarMention struct {
	ID          string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Text        string    `json:"text" example:"Превью похожего упоминания..."`
	Source      SourceRef `json:"source"`
	Sentiment   string    `json:"sentiment" example:"neutral"`
	PublishedAt string    `json:"published_at" example:"2024-01-01T13:00:00Z"`
}

// MentionResponse — упоминание в ответе API.
type MentionResponse struct {
	ID              uuid.UUID        `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BrandID         uuid.UUID        `json:"brand_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Source          SourceRef        `json:"source"`
	Text            string           `json:"text" example:"Текст публикации"`
	URL             string           `json:"url" example:"https://t.me/channel/123"`
	Sentiment       string           `json:"sentiment" example:"positive"`
	PublishedAt     string           `json:"published_at" example:"2024-01-01T12:00:00Z"`
	CreatedAt       string           `json:"created_at" example:"2024-01-01T12:00:00Z"`
	SimilarMentions []SimilarMention `json:"similar_mentions"`
	SimilarCount    int              `json:"similar_count" example:"0"`
}
