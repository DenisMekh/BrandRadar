package v1

import (
	"time"

	"github.com/google/uuid"

	"prod-pobeda-2026/internal/controller/http/v1/dto"
	"prod-pobeda-2026/internal/entity"
)

func toBrandResponse(b *entity.Brand) dto.BrandResponse {
	return dto.BrandResponse{
		ID:         b.ID.String(),
		Name:       b.Name,
		Keywords:   b.Keywords,
		Exclusions: b.Exclusions,
		RiskWords:  b.RiskWords,
		CreatedAt:  b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  b.UpdatedAt.Format(time.RFC3339),
	}
}

func toSourceResponse(s *entity.Source) dto.SourceResponse {
	return dto.SourceResponse{
		ID:        s.ID.String(),
		Type:      s.Type,
		Name:      s.Name,
		URL:       s.URL,
		Status:    string(s.Status),
		CreatedAt: s.CreatedAt.Format(time.RFC3339),
		UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
	}
}

func toMentionResponse(m *entity.Mention) dto.MentionResponse {
	var sourceRef dto.SourceRef
	if m.SourceID != uuid.Nil {
		sourceRef = dto.SourceRef{
			ID:   m.SourceID.String(),
			Name: m.SourceName,
			Type: m.SourceType,
		}
	}

	similar := make([]dto.SimilarMention, 0, len(m.Similar))
	for _, s := range m.Similar {
		sim := dto.SimilarMention{
			ID:          s.ID.String(),
			Text:        s.Text,
			Sentiment:   string(s.Sentiment),
			PublishedAt: s.PublishedAt.Format(time.RFC3339),
		}
		if s.SourceID != uuid.Nil {
			sim.Source = dto.SourceRef{
				ID:   s.SourceID.String(),
				Name: s.SourceName,
				Type: s.SourceType,
			}
		}
		similar = append(similar, sim)
	}

	return dto.MentionResponse{
		ID:              m.ID,
		BrandID:         m.BrandID,
		Source:          sourceRef,
		Text:            m.Text,
		URL:             m.URL,
		Sentiment:       string(m.Sentiment),
		PublishedAt:     m.PublishedAt.Format(time.RFC3339),
		CreatedAt:       m.CreatedAt.Format(time.RFC3339),
		SimilarMentions: similar,
		SimilarCount:    len(similar),
	}
}

func toAlertConfigResponse(cfg *entity.AlertConfig) dto.AlertConfigResponse {
	return dto.AlertConfigResponse{
		ID:                cfg.ID.String(),
		BrandID:           cfg.BrandID.String(),
		WindowMinutes:     cfg.WindowMinutes,
		CooldownMinutes:   cfg.CooldownMinutes,
		SentimentFilter:   cfg.SentimentFilter,
		Enabled:           cfg.Enabled,
		Percentile:        cfg.Percentile,
		AnomalyWindowSize: cfg.AnomalyWindowSize,
		CreatedAt:         cfg.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         cfg.UpdatedAt.Format(time.RFC3339),
	}
}

func toAlertResponse(a *entity.Alert) dto.AlertResponse {
	return dto.AlertResponse{
		ID:            a.ID.String(),
		ConfigID:      a.ConfigID.String(),
		BrandID:       a.BrandID.String(),
		MentionsCount: a.MentionsCount,
		WindowStart:   a.WindowStart.Format(time.RFC3339),
		WindowEnd:     a.WindowEnd.Format(time.RFC3339),
		FiredAt:       a.FiredAt.Format(time.RFC3339),
		CreatedAt:     a.CreatedAt.Format(time.RFC3339),
	}
}

func toEventResponse(e *entity.Event) dto.EventResponse {
	return dto.EventResponse{
		ID:         e.ID,
		Type:       string(e.Type),
		Payload:    e.Payload,
		OccurredAt: e.OccurredAt.Format(time.RFC3339),
	}
}

func toSentimentStats(s entity.SentimentCounts) dto.SentimentStats {
	return dto.SentimentStats{
		Positive: s.Positive,
		Negative: s.Negative,
		Neutral:  s.Neutral,
	}
}

func toBrandDashboardResponse(d *entity.BrandDashboard) dto.BrandDashboardResponse {
	sources := make([]dto.SourceStats, len(d.BySources))
	for i, s := range d.BySources {
		sources[i] = dto.SourceStats{Source: s.Source, Count: s.Count}
	}

	daily := make([]dto.DailyStats, len(d.ByDate))
	for i, day := range d.ByDate {
		daily[i] = dto.DailyStats{
			Date:     day.Date,
			Positive: day.Positive,
			Negative: day.Negative,
			Neutral:  day.Neutral,
			Total:    day.Positive + day.Negative + day.Neutral,
		}
	}

	return dto.BrandDashboardResponse{
		BrandID:       d.BrandID,
		BrandName:     d.BrandName,
		TotalMentions: d.Sentiment.Total(),
		Sentiment:     toSentimentStats(d.Sentiment),
		BySources:     sources,
		ByDate:        daily,
		RecentAlerts:  d.RecentAlerts,
	}
}

func toOverallDashboardResponse(s *entity.SentimentCounts, brands []entity.BrandSummary, daily []entity.DailyCount) dto.OverallDashboardResponse {
	brandDTOs := make([]dto.BrandSummary, len(brands))
	for i, b := range brands {
		brandDTOs[i] = dto.BrandSummary{
			BrandID:       b.BrandID,
			BrandName:     b.BrandName,
			TotalMentions: b.Sentiment.Total(),
			Sentiment:     toSentimentStats(b.Sentiment),
		}
	}

	dailyDTOs := make([]dto.DailyStats, len(daily))
	for i, d := range daily {
		dailyDTOs[i] = dto.DailyStats{
			Date:     d.Date,
			Positive: d.Positive,
			Negative: d.Negative,
			Neutral:  d.Neutral,
			Total:    d.Positive + d.Negative + d.Neutral,
		}
	}

	return dto.OverallDashboardResponse{
		TotalMentions: s.Total(),
		Sentiment:     toSentimentStats(*s),
		Brands:        brandDTOs,
		ByDate:        dailyDTOs,
	}
}
