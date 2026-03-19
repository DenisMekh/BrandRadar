package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

// EventResponse — событие в ответе API.
type EventResponse struct {
	ID         uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type       string          `json:"type" example:"spike_alert"`
	Payload    json.RawMessage `json:"payload" swaggertype:"string" example:"{\"brand_id\":\"550e8400-e29b-41d4-a716-446655440000\",\"count\":15}"`
	OccurredAt string          `json:"occurred_at" example:"2024-01-01T12:00:00Z"`
}
