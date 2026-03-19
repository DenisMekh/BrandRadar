package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateMockMentionRequest — запрос на создание тестового упоминания через mock endpoint.

// Все поля можно конфигурировать, включая дату публикации.

// Упоминание проходит стандартный ML pipeline (relevance + sentiment).

type CreateMockMentionRequest struct {
	BrandID uuid.UUID `json:"brand_id" binding:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`

	SourceID uuid.UUID `json:"source_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`

	ExternalID string `json:"external_id" binding:"required" example:"mock-12345"`

	Title string `json:"title" example:"Заголовок публикации"`

	Text string `json:"text" binding:"required" example:"Текст публикации о бренде"`

	URL string `json:"url" example:"https://t.me/channel/123"`

	Author string `json:"author" example:"@user"`

	PublishedAt time.Time `json:"published_at" binding:"required" example:"2024-01-01T12:00:00Z"`
}
