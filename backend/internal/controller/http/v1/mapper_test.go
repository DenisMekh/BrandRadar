package v1

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"prod-pobeda-2026/internal/entity"
)

func TestMapper_ToAlertConfigResponse(t *testing.T) {
	now := time.Now().UTC()
	cfg := &entity.AlertConfig{
		ID:                uuid.New(),
		BrandID:           uuid.New(),
		WindowMinutes:     60,
		CooldownMinutes:   30,
		SentimentFilter:   "negative",
		Enabled:           true,
		Percentile:        95.0,
		AnomalyWindowSize: 10,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	resp := toAlertConfigResponse(cfg)
	assert.Equal(t, cfg.ID.String(), resp.ID)
	assert.Equal(t, cfg.BrandID.String(), resp.BrandID)
	assert.Equal(t, 95.0, resp.Percentile)
	assert.Equal(t, 10, resp.AnomalyWindowSize)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
}

func TestMapper_ToMentionResponse(t *testing.T) {
	now := time.Now().UTC()
	m := &entity.Mention{
		ID:          uuid.New(),
		BrandID:     uuid.New(),
		SourceID:    uuid.New(),
		Text:        "text",
		URL:         "https://example.com",
		Sentiment:   entity.SENTIMENT_NEGATIVE,
		PublishedAt: now,
		CreatedAt:   now,
	}

	resp := toMentionResponse(m)
	assert.Equal(t, m.ID, resp.ID)
	assert.Equal(t, "negative", resp.Sentiment)
	assert.Equal(t, now.Format(time.RFC3339), resp.PublishedAt)
}

func TestMapper_ToEventAndAlertResponse(t *testing.T) {
	now := time.Now().UTC()
	e := &entity.Event{ID: uuid.New(), Type: entity.EventAlertFired, Payload: []byte(`{"ok":true}`), OccurredAt: now}
	a := &entity.Alert{
		ID:            uuid.New(),
		ConfigID:      uuid.New(),
		BrandID:       uuid.New(),
		MentionsCount: 8,
		WindowStart:   now.Add(-time.Hour),
		WindowEnd:     now,
		FiredAt:       now,
		CreatedAt:     now,
	}

	eventResp := toEventResponse(e)
	alertResp := toAlertResponse(a)
	assert.Equal(t, string(entity.EventAlertFired), eventResp.Type)
	assert.Equal(t, a.ID.String(), alertResp.ID)
	assert.Equal(t, 8, alertResp.MentionsCount)
}
