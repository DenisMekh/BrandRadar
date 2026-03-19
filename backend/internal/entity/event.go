package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventType — тип события в журнале.
type EventType string

const (
	EventMentionCreated    EventType = "mention_created"
	EventMentionDuplicated EventType = "mention_duplicated"
	EventSpikeDetected     EventType = "spike_detected"
	EventAlertFired        EventType = "alert_fired"
	EventCollectorStarted  EventType = "collector_started"
	EventCollectorStopped  EventType = "collector_stopped"
	EventCollectorFailed   EventType = "collector_failed"
	EventSourceToggled     EventType = "source_toggled"
)

// Event — запись в журнале событий
type Event struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	Type       EventType       `json:"type" db:"type"`
	Payload    json.RawMessage `json:"payload" db:"payload"`
	OccurredAt time.Time       `json:"occurred_at" db:"occurred_at"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}
