//go:build integration

package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"prod-pobeda-2026/internal/entity"
)

func TestMention_ListWithFilters(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	brandID, sourceID := createBrandAndSourceForMentions(t, "MentionBrand")

	sentiments := []entity.Sentiment{
		entity.SENTIMENT_NEGATIVE,
		entity.SENTIMENT_POSITIVE,
		entity.SENTIMENT_NEGATIVE,
		entity.SENTIMENT_NEUTRAL,
		entity.SENTIMENT_NEGATIVE,
	}
	for i := 0; i < 5; i++ {
		err := mentionRepo.Create(context.Background(), &entity.Mention{
			ID:          uuid.New(),
			BrandID:     brandID,
			SourceID:    sourceID,
			Text:        "text",
			URL:         "https://example.com",
			Sentiment:   sentiments[i],
			PublishedAt: time.Now().UTC().Add(time.Duration(i) * time.Minute),
			CreatedAt:   time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("failed to create mention %d: %v", i, err)
		}
	}

	filterResp := doGET("/api/v1/mentions?brand_id=" + brandID.String() + "&sentiment=negative")
	assertStatus(t, filterResp, http.StatusOK)
	filterBody := parseJSON(filterResp)
	assertNoError(t, filterBody)
	filterItems := filterBody["data"].(map[string]any)["items"].([]any)
	if len(filterItems) != 3 {
		t.Fatalf("expected 3 negative mentions, got=%d", len(filterItems))
	}

	pageResp := doGET("/api/v1/mentions?brand_id=" + brandID.String() + "&limit=2&offset=0")
	assertStatus(t, pageResp, http.StatusOK)
	pageBody := parseJSON(pageResp)
	assertNoError(t, pageBody)
	pageData := pageBody["data"].(map[string]any)
	if len(pageData["items"].([]any)) != 2 {
		t.Fatalf("expected 2 items in page")
	}
	total := int(pageData["total"].(float64))
	if total != 5 {
		t.Fatalf("expected total=5, got=%d", total)
	}
}

func TestMention_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	brandID, sourceID := createBrandAndSourceForMentions(t, "MentionByIDBrand")
	mentionID := uuid.New()

	err := mentionRepo.Create(context.Background(), &entity.Mention{
		ID:          mentionID,
		BrandID:     brandID,
		SourceID:    sourceID,
		Text:        "single-text",
		URL:         "https://example.com/single",
		Sentiment:   entity.SENTIMENT_NEGATIVE,
		PublishedAt: time.Now().UTC(),
		CreatedAt:   time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("failed to create mention: %v", err)
	}

	resp := doGET("/api/v1/mentions/" + mentionID.String())
	assertStatus(t, resp, http.StatusOK)
	body := parseJSON(resp)
	assertNoError(t, body)

	data, _ := body["data"].(map[string]any)
	if data["sentiment"] == "" {
		t.Fatalf("expected sentiment to be set")
	}
	if _, ok := data["published_at"]; !ok {
		t.Fatalf("expected published_at")
	}
}

func TestMention_NotFound(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	resp := doGET("/api/v1/mentions/00000000-0000-0000-0000-000000000099")
	assertStatus(t, resp, http.StatusNotFound)
	body := parseJSON(resp)
	assertHasError(t, body, "NOT_FOUND")
}

func createBrandAndSourceForMentions(t *testing.T, brandName string) (uuid.UUID, uuid.UUID) {
	t.Helper()

	brandResp := doPOST("/api/v1/brands", map[string]any{
		"name":       brandName,
		"keywords":   []string{"mentions"},
		"exclusions": []string{},
		"risk_words": []string{},
	})
	assertStatus(t, brandResp, http.StatusCreated)
	brandBody := parseJSON(brandResp)
	brandID := uuid.MustParse(brandBody["data"].(map[string]any)["id"].(string))

	sourceResp := doPOST("/api/v1/sources", map[string]any{
		"type": "telegram",
		"name": "Mentions source",
		"url":  "mentions_e2e",
	})
	assertStatus(t, sourceResp, http.StatusCreated)
	sourceBody := parseJSON(sourceResp)
	sourceID := uuid.MustParse(sourceBody["data"].(map[string]any)["id"].(string))

	return brandID, sourceID
}
