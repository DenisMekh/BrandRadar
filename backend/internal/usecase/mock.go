package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/client"
	"prod-pobeda-2026/internal/entity"
)

type MockUseCase struct {
	mentionRepo       MentionRepository
	eventRepo         EventRepository
	brandRepo         BrandRepository
	sourceRepo        SourceRepository
	sentimentMLClient client.SentimentMLClient
}

func NewMockUseCase(
	mentionRepo MentionRepository,
	eventRepo EventRepository,
	brandRepo BrandRepository,
	sourceRepo SourceRepository,
	sentimentMLClient client.SentimentMLClient,
) *MockUseCase {
	return &MockUseCase{
		mentionRepo:       mentionRepo,
		eventRepo:         eventRepo,
		brandRepo:         brandRepo,
		sourceRepo:        sourceRepo,
		sentimentMLClient: sentimentMLClient,
	}
}

func (uc *MockUseCase) CreateMentionWithML(ctx context.Context, req CreateMockMentionRequest) (*entity.Mention, error) {
	if req.BrandID == uuid.Nil {
		return uc.createMentionsForAllBrands(ctx, req)
	}

	brand, err := uc.brandRepo.GetByID(ctx, req.BrandID)
	if err != nil {
		return nil, fmt.Errorf("MockUseCase.CreateMentionWithML: get brand: %w", err)
	}

	source, err := uc.sourceRepo.GetByID(ctx, req.SourceID)
	if err != nil {
		return nil, fmt.Errorf("MockUseCase.CreateMentionWithML: get source: %w", err)
	}

	return uc.createMentionForBrand(ctx, req, brand, source)
}

func (uc *MockUseCase) createMentionsForAllBrands(ctx context.Context, req CreateMockMentionRequest) (*entity.Mention, error) {
	brands, _, err := uc.brandRepo.List(ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("MockUseCase.createMentionsForAllBrands: list brands: %w", err)
	}

	source, err := uc.sourceRepo.GetByID(ctx, req.SourceID)
	if err != nil {
		return nil, fmt.Errorf("MockUseCase.createMentionsForAllBrands: get source: %w", err)
	}

	var createdMention *entity.Mention
	var createdCount int

	for _, brand := range brands {
		relevance, err := uc.sentimentMLClient.Relevance(req.Text, brand.Name, brand.Keywords)
		if err != nil {
			return nil, fmt.Errorf("MockUseCase.createMentionsForAllBrands: relevance ML failed for brand %s: %w", brand.ID, err)
		}

		if !relevance.IsRelevant {
			logrus.WithFields(logrus.Fields{
				"item":  req.Text,
				"brand": brand.ID,
			}).Info("mention is not relevant for brand, skipping (mock endpoint)")
			continue
		}

		sentiment, err := uc.sentimentMLClient.Sentiment(req.Text, brand.Name)
		if err != nil {
			return nil, fmt.Errorf("MockUseCase.createMentionsForAllBrands: sentiment ML failed for brand %s: %w", brand.ID, err)
		}

		mention := &entity.Mention{
			ID:           uuid.New(),
			BrandID:      brand.ID,
			SourceID:     req.SourceID,
			ExternalID:   fmt.Sprintf("%s-%s", req.ExternalID, brand.ID.String()[:8]),
			Title:        req.Title,
			Text:         req.Text,
			URL:          req.URL,
			Author:       req.Author,
			PublishedAt:  req.PublishedAt.UTC(),
			Status:       entity.MentionStatusNew,
			Deduplicated: false,
			Sentiment:    sentiment.Sentiment,
			ML: entity.MentionML{
				Label:      string(sentiment.Sentiment),
				Score:      sentiment.Confidence,
				IsRelevant: relevance.IsRelevant,
				SimilarIDs: []string{},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		if err := uc.mentionRepo.Create(ctx, mention); err != nil {
			logrus.Warnf("failed to create mention for brand %s: %v", brand.ID, err)
			continue
		}

		if err := uc.eventRepo.Create(ctx, &entity.Event{
			ID:         uuid.New(),
			Type:       entity.EventMentionCreated,
			Payload:    []byte(fmt.Sprintf(`{"mention_id":"%s","brand_id":"%s"}`, mention.ID.String(), brand.ID.String())),
			OccurredAt: time.Now().UTC(),
		}); err != nil {
			logrus.Warnf("failed to create mention event: %v", err)
		}

		logrus.WithFields(logrus.Fields{
			"mention_id":  mention.ID,
			"brand_id":    brand.ID,
			"source_id":   source.ID,
			"sentiment":   sentiment.Sentiment,
			"confidence":  sentiment.Confidence,
			"is_relevant": relevance.IsRelevant,
		}).Info("mock mention created with ML pipeline")

		createdMention = mention
		createdCount++
	}

	if createdCount == 0 {
		return nil, fmt.Errorf("no relevant brands found for mention")
	}

	logrus.Infof("created %d mock mentions for %d relevant brands", createdCount, len(brands))
	return createdMention, nil
}

func (uc *MockUseCase) createMentionForBrand(ctx context.Context, req CreateMockMentionRequest, brand *entity.Brand, source *entity.Source) (*entity.Mention, error) {
	relevance, err := uc.sentimentMLClient.Relevance(req.Text, brand.Name, brand.Keywords)
	if err != nil {
		return nil, fmt.Errorf("MockUseCase.createMentionForBrand: relevance ML failed: %w", err)
	}

	if !relevance.IsRelevant {
		logrus.WithFields(logrus.Fields{
			"item":  req.Text,
			"brand": brand.ID,
		}).Info("mention is not relevant according to ML, but creating anyway (mock endpoint)")
	}

	sentiment, err := uc.sentimentMLClient.Sentiment(req.Text, brand.Name)
	if err != nil {
		return nil, fmt.Errorf("MockUseCase.createMentionForBrand: sentiment ML failed: %w", err)
	}

	mention := &entity.Mention{
		ID:           uuid.New(),
		BrandID:      req.BrandID,
		SourceID:     req.SourceID,
		ExternalID:   req.ExternalID,
		Title:        req.Title,
		Text:         req.Text,
		URL:          req.URL,
		Author:       req.Author,
		PublishedAt:  req.PublishedAt.UTC(),
		Status:       entity.MentionStatusNew,
		Deduplicated: false,
		Sentiment:    sentiment.Sentiment,
		ML: entity.MentionML{
			Label:      string(sentiment.Sentiment),
			Score:      sentiment.Confidence,
			IsRelevant: relevance.IsRelevant,
			SimilarIDs: []string{},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := uc.mentionRepo.Create(ctx, mention); err != nil {
		return nil, fmt.Errorf("MockUseCase.CreateMentionWithML: create mention: %w", err)
	}

	if err := uc.eventRepo.Create(ctx, &entity.Event{
		ID:         uuid.New(),
		Type:       entity.EventMentionCreated,
		Payload:    []byte(fmt.Sprintf(`{"mention_id":"%s","brand_id":"%s"}`, mention.ID.String(), brand.ID.String())),
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		logrus.Warnf("failed to create mention event: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"mention_id":  mention.ID,
		"brand_id":    brand.ID,
		"source_id":   source.ID,
		"sentiment":   sentiment.Sentiment,
		"confidence":  sentiment.Confidence,
		"is_relevant": relevance.IsRelevant,
	}).Info("mock mention created with ML pipeline")

	return mention, nil
}

type CreateMockMentionRequest struct {
	BrandID     uuid.UUID `json:"brand_id"`
	SourceID    uuid.UUID `json:"source_id"`
	ExternalID  string    `json:"external_id"`
	Title       string    `json:"title"`
	Text        string    `json:"text"`
	URL         string    `json:"url"`
	Author      string    `json:"author"`
	PublishedAt time.Time `json:"published_at"`
}
