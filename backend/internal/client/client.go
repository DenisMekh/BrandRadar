package client

import (
	"prod-pobeda-2026/internal/client/sentiment_ml/v1/dto"
	"prod-pobeda-2026/internal/entity"
)

type SentimentMLClient interface {
	HealthCheck() error
	Sentiment(text, brandName string) (entity.SentimentMLOutput, error)
	Relevance(text, brandName string, keywords []string) (entity.RelevanceMLOutput, error)
	Cluster(texts []string, minClusterSize int) (*dto.ClusterResponse, error)
}
