package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/client/sentiment_ml/v1/dto"
	"prod-pobeda-2026/internal/entity"
)

type client struct {
	httpClient *http.Client

	baseURL string
}

func NewSentimentMLClient(httpClient *http.Client, baseURL string) *client {
	return &client{
		httpClient: httpClient,

		baseURL: baseURL,
	}
}

func (c *client) HealthCheck() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("ml healthcheck: %w", err)
	}
	defer func(Body io.ReadCloser) {}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ml healthcheck failed: status %d", resp.StatusCode)
	}
	return nil
}

func (c *client) Sentiment(text, brandName string) (entity.SentimentMLOutput, error) {
	req := dto.SentimentRequest{
		Text: text,

		BrandName: brandName,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return entity.SentimentMLOutput{}, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/sentiment", "application/json", bytes.NewReader(body))
	if err != nil {
		return entity.SentimentMLOutput{}, fmt.Errorf("post request to %s: %w", c.baseURL, err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return entity.SentimentMLOutput{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result dto.SentimentResponse

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return entity.SentimentMLOutput{}, fmt.Errorf("decode response: %w", err)
	}

	var finalSentiment entity.Sentiment

	maxConfidence := 0.5 // Порог уверенности

	switch {

	case result.Scores.Positive > maxConfidence:

		finalSentiment = entity.SENTIMENT_POSITIVE

	case result.Scores.Negative > maxConfidence:

		finalSentiment = entity.SENTIMENT_NEGATIVE

	default:

		finalSentiment = entity.SENTIMENT_NEUTRAL

	}

	return entity.SentimentMLOutput{
		Confidence: result.Confidence,
		Sentiment:  finalSentiment,
	}, nil
}

func (c *client) Cluster(texts []string, minClusterSize int) (*dto.ClusterResponse, error) {
	req := dto.ClusterRequest{
		Texts:          texts,
		MinClusterSize: minClusterSize,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal cluster request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/clusterization", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("post cluster request to %s: %w", c.baseURL, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Warnf("failed to close cluster response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cluster unexpected status: %d", resp.StatusCode)
	}

	var result dto.ClusterResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode cluster response: %w", err)
	}

	return &result, nil
}

func (c *client) Relevance(text, companyName string, keywords []string) (entity.RelevanceMLOutput, error) {
	req := dto.RelevanceRequest{
		Text:     text,
		Brand:    companyName,
		Keywords: strings.Join(keywords, ";"),
	}

	body, err := json.Marshal(req)
	if err != nil {
		return entity.RelevanceMLOutput{}, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/relevance", "application/json", bytes.NewReader(body))
	if err != nil {
		return entity.RelevanceMLOutput{}, fmt.Errorf("post request to %s: %w", c.baseURL, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Warnf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {

		println("PROBLEM", c.baseURL)

		return entity.RelevanceMLOutput{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)

	}

	var result dto.RelevanceResponse

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return entity.RelevanceMLOutput{}, fmt.Errorf("decode response: %w", err)
	}

	return entity.RelevanceMLOutput{
		IsRelevant: result.IsRelevant,
		Confidence: result.Confidence,
		Label:      result.Label,
	}, nil
}
