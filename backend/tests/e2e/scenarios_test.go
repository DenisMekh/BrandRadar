//go:build integration

package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenario_DashboardEndpoints(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	w := doPOST("/api/v1/brands", map[string]any{
		"name":       "DashTest",
		"keywords":   []string{"test"},
		"exclusions": []string{},
		"risk_words": []string{},
	})
	assertStatus(t, w, http.StatusCreated)

	brandResp := parseJSON(w)
	brandData := brandResp["data"].(map[string]interface{})
	brandID := brandData["id"].(string)

	w = doGET("/api/v1/dashboard")
	assertStatus(t, w, http.StatusOK)

	overallResp := parseJSON(w)
	assert.NotNil(t, overallResp["data"])
	assert.Nil(t, overallResp["error"])

	w = doGET("/api/v1/brands/" + brandID + "/dashboard")
	assertStatus(t, w, http.StatusOK)

	brandDashResp := parseJSON(w)
	data := brandDashResp["data"].(map[string]interface{})
	assert.Equal(t, brandID, data["brand_id"])
	assert.NotNil(t, data["sentiment"])
	assert.NotNil(t, data["by_source"])
	assert.NotNil(t, data["by_date"])

	w = doGET("/api/v1/dashboard?date_from=2026-01-01&date_to=2026-12-31")
	assertStatus(t, w, http.StatusOK)

	w = doGET("/api/v1/brands/00000000-0000-0000-0000-000000000099/dashboard")
	assertStatus(t, w, http.StatusNotFound)
}

func TestScenario_FullUserJourney(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	ctx := context.Background()

	var (
		brandID          string
		sourceIDTelegram string
		sourceIDWeb      string
		alertConfigID    string
		firstNegativeID  string
	)

	t.Run("1_create_brand", func(t *testing.T) {
		resp := doPOST("/api/v1/brands", map[string]any{
			"name":       "TestBrand",
			"keywords":   []string{"тест", "brand"},
			"exclusions": []string{},
			"risk_words": []string{},
		})
		assertStatus(t, resp, http.StatusCreated)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		data := mustMap(t, body["data"])
		brandID = mustString(t, data["id"])
		assert.Equal(t, "TestBrand", mustString(t, data["name"]))
	})

	t.Run("2_create_sources", func(t *testing.T) {
		respTG := doPOST("/api/v1/sources", map[string]any{
			"type": "telegram",
			"name": "Scenario Telegram",
			"url":  "scenario_tg",
		})
		assertStatus(t, respTG, http.StatusCreated)
		bodyTG := parseJSON(respTG)
		assertEnvelopeOK(t, bodyTG)
		sourceIDTelegram = mustString(t, mustMap(t, bodyTG["data"])["id"])

		respWeb := doPOST("/api/v1/sources", map[string]any{
			"type": "web",
			"name": "Scenario Web",
			"url":  "https://example.com/scenario",
		})
		assertStatus(t, respWeb, http.StatusCreated)
		bodyWeb := parseJSON(respWeb)
		assertEnvelopeOK(t, bodyWeb)
		sourceIDWeb = mustString(t, mustMap(t, bodyWeb["data"])["id"])
	})

	t.Run("3_configure_alert", func(t *testing.T) {
		resp := doPOST("/api/v1/alerts/config", map[string]any{
			"brand_id":            brandID,
			"window_minutes":      60,
			"cooldown_minutes":    5,
			"sentiment_filter":    "negative",
			"percentile":          95,
			"anomaly_window_size": 10,
		})
		assertStatus(t, resp, http.StatusCreated)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		alertConfigID = mustString(t, mustMap(t, body["data"])["id"])
		require.NotEmpty(t, alertConfigID)
	})

	t.Run("4_ingest_mentions", func(t *testing.T) {
		brandUUID := mustUUID(t, brandID)
		tgUUID := mustUUID(t, sourceIDTelegram)
		webUUID := mustUUID(t, sourceIDWeb)
		now := time.Now().UTC()

		firstNegativeID = insertMentionDirect(t, ctx, mentionSeed{
			BrandID:     brandUUID,
			SourceID:    tgUUID,
			Text:        "Тест бренд: негативная публикация 1",
			URL:         "https://example.com/neg-1",
			PublishedAt: now.Add(-10 * time.Minute),
			Sentiment:   "negative",
		}).String()

		_ = insertMentionDirect(t, ctx, mentionSeed{
			BrandID:     brandUUID,
			SourceID:    webUUID,
			Text:        "Тест бренд: негативная публикация 2",
			URL:         "https://example.com/neg-2",
			PublishedAt: now.Add(-9 * time.Minute),
			Sentiment:   "negative",
		})
		_ = insertMentionDirect(t, ctx, mentionSeed{
			BrandID:     brandUUID,
			SourceID:    tgUUID,
			Text:        "Тест бренд: негативная публикация 3",
			URL:         "https://example.com/neg-3",
			PublishedAt: now.Add(-8 * time.Minute),
			Sentiment:   "negative",
		})
		_ = insertMentionDirect(t, ctx, mentionSeed{
			BrandID:     brandUUID,
			SourceID:    webUUID,
			Text:        "Позитивный отзыв",
			URL:         "https://example.com/pos-1",
			PublishedAt: now.Add(-7 * time.Minute),
			Sentiment:   "positive",
		})
		_ = insertMentionDirect(t, ctx, mentionSeed{
			BrandID:     brandUUID,
			SourceID:    tgUUID,
			Text:        "Нейтральная публикация",
			URL:         "https://example.com/neu-1",
			PublishedAt: now.Add(-6 * time.Minute),
			Sentiment:   "neutral",
		})
	})

	t.Run("5_verify_count", func(t *testing.T) {
		listResp := doGET("/api/v1/mentions?brand_id=" + brandID)
		assertStatus(t, listResp, http.StatusOK)
		listBody := parseJSON(listResp)
		assertEnvelopeOK(t, listBody)
		total := paginatedTotal(t, listBody)
		assert.Equal(t, 5, total)
	})

	t.Run("6_check_sentiment_in_card", func(t *testing.T) {
		resp := doGET("/api/v1/mentions/" + firstNegativeID)
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		data := mustMap(t, body["data"])
		assert.Equal(t, "negative", mustString(t, data["sentiment"]))
	})

	t.Run("7_filter_negative", func(t *testing.T) {
		resp := doGET("/api/v1/mentions?brand_id=" + brandID + "&sentiment=negative")
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		items := paginatedItems(t, body)
		require.Len(t, items, 3)
		for _, item := range items {
			m := mustMap(t, item)
			assert.Equal(t, "negative", mustString(t, m["sentiment"]))
		}
	})

	t.Run("8_generate_spike_alert", func(t *testing.T) {
		err := insertAlertDirect(t, ctx, mustUUID(t, alertConfigID), mustUUID(t, brandID), 5)
		require.NoError(t, err)

		err = insertEventDirect(t, ctx, "spike_alert", map[string]any{
			"brand_id":       brandID,
			"mentions_count": 5,
			"type":           "spike",
		})
		require.NoError(t, err)

		resp := doGET("/api/v1/brands/" + brandID + "/alerts")
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		items := paginatedItems(t, body)
		require.NotEmpty(t, items)

		alertObj := mustMap(t, items[0])
		assert.Equal(t, brandID, mustString(t, alertObj["brand_id"]))
		assert.GreaterOrEqual(t, int(mustFloat64(t, alertObj["mentions_count"])), 1)
	})

	t.Run("9_check_events", func(t *testing.T) {
		resp := doGET("/api/v1/events")
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		items := paginatedItems(t, body)
		require.NotEmpty(t, items)
		found := false
		for _, item := range items {
			ev := mustMap(t, item)
			if mustString(t, ev["type"]) == "spike_alert" {
				found = true
				break
			}
		}
		assert.True(t, found, "expected at least one event with type=spike_alert")
	})

	t.Run("10_check_health", func(t *testing.T) {
		resp := doGET("/api/v1/health")
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		data := mustMap(t, body["data"])
		assert.Equal(t, "ok", mustString(t, data["status"]))
		deps := mustMap(t, data["dependencies"])
		assert.Equal(t, "ok", mustString(t, deps["postgres"]))
		assert.Equal(t, "ok", mustString(t, deps["redis"]))
	})
}

func TestScenario_BrandCRUDLifecycle(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	var brandID string

	t.Run("create", func(t *testing.T) {
		resp := doPOST("/api/v1/brands", map[string]any{
			"name":       "LifecycleBrand",
			"keywords":   []string{"life", "cycle"},
			"exclusions": []string{"skip"},
			"risk_words": []string{"risk"},
		})
		assertStatus(t, resp, http.StatusCreated)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		data := mustMap(t, body["data"])
		brandID = mustString(t, data["id"])
		assert.Equal(t, "LifecycleBrand", mustString(t, data["name"]))
	})

	t.Run("get", func(t *testing.T) {
		resp := doGET("/api/v1/brands/" + brandID)
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		data := mustMap(t, body["data"])
		assert.Equal(t, brandID, mustString(t, data["id"]))
		assert.Equal(t, "LifecycleBrand", mustString(t, data["name"]))
		assert.ElementsMatch(t, []string{"life", "cycle"}, toStringSlice(t, data["keywords"]))
		assert.ElementsMatch(t, []string{"skip"}, toStringSlice(t, data["exclusions"]))
		assert.ElementsMatch(t, []string{"risk"}, toStringSlice(t, data["risk_words"]))
	})

	t.Run("list", func(t *testing.T) {
		resp := doGET("/api/v1/brands")
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		items := paginatedItems(t, body)
		found := false
		for _, item := range items {
			m := mustMap(t, item)
			if mustString(t, m["id"]) == brandID {
				found = true
				break
			}
		}
		assert.True(t, found, "brand must be present in list")
	})

	t.Run("update", func(t *testing.T) {
		resp := doPUT("/api/v1/brands/"+brandID, map[string]any{
			"name":     "LifecycleBrandUpdated",
			"keywords": []string{"updated", "keywords"},
		})
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		data := mustMap(t, body["data"])
		assert.Equal(t, "LifecycleBrandUpdated", mustString(t, data["name"]))
	})

	t.Run("get_updated", func(t *testing.T) {
		resp := doGET("/api/v1/brands/" + brandID)
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		data := mustMap(t, body["data"])
		assert.Equal(t, "LifecycleBrandUpdated", mustString(t, data["name"]))
		assert.ElementsMatch(t, []string{"updated", "keywords"}, toStringSlice(t, data["keywords"]))
	})

	t.Run("delete", func(t *testing.T) {
		resp := doDELETE("/api/v1/brands/" + brandID)
		assertStatus(t, resp, http.StatusNoContent)
		_ = resp.Body.Close()
	})

	t.Run("get_deleted", func(t *testing.T) {
		resp := doGET("/api/v1/brands/" + brandID)
		assertStatus(t, resp, http.StatusNotFound)
		body := parseJSON(resp)
		assertHasError(t, body, "NOT_FOUND")
	})
}

func TestScenario_SourceManagement(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	var sourceID string

	t.Run("create_source", func(t *testing.T) {
		resp := doPOST("/api/v1/sources", map[string]any{
			"type": "telegram",
			"name": "ManagedSource",
			"url":  "managed_src",
		})
		assertStatus(t, resp, http.StatusCreated)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		sourceID = mustString(t, mustMap(t, body["data"])["id"])
	})

	t.Run("list_sources", func(t *testing.T) {
		resp := doGET("/api/v1/sources")
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		items := paginatedItems(t, body)
		found := false
		for _, item := range items {
			m := mustMap(t, item)
			if mustString(t, m["id"]) == sourceID {
				found = true
				break
			}
		}
		assert.True(t, found, "source must be present in list")
	})

	t.Run("toggle_disable", func(t *testing.T) {
		resp := doPOST("/api/v1/sources/"+sourceID+"/toggle", nil)
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		assert.Equal(t, "inactive", mustString(t, mustMap(t, body["data"])["status"]))
	})

	t.Run("verify_disabled", func(t *testing.T) {
		resp := doGET("/api/v1/sources/" + sourceID)
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		data := mustMap(t, body["data"])
		assert.Equal(t, "inactive", mustString(t, data["status"]))
		if val, ok := data["is_active"]; ok {
			assert.False(t, mustBool(t, val))
		}
	})

	t.Run("toggle_enable", func(t *testing.T) {
		resp := doPOST("/api/v1/sources/"+sourceID+"/toggle", nil)
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		assert.Equal(t, "active", mustString(t, mustMap(t, body["data"])["status"]))
	})

	t.Run("verify_enabled", func(t *testing.T) {
		resp := doGET("/api/v1/sources/" + sourceID)
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		data := mustMap(t, body["data"])
		assert.Equal(t, "active", mustString(t, data["status"]))
		if val, ok := data["is_active"]; ok {
			assert.True(t, mustBool(t, val))
		}
	})

	t.Run("delete_source", func(t *testing.T) {
		resp := doDELETE("/api/v1/sources/" + sourceID)
		assertStatus(t, resp, http.StatusNoContent)
		_ = resp.Body.Close()
	})

	t.Run("verify_deleted", func(t *testing.T) {
		resp := doGET("/api/v1/sources/" + sourceID)
		assertStatus(t, resp, http.StatusNotFound)
		body := parseJSON(resp)
		assertHasError(t, body, "NOT_FOUND")
	})
}

func TestScenario_MentionFilters(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	ctx := context.Background()

	var (
		brandA string
		brandB string
		srcA   string
		srcB   string
	)

	t.Run("seed_data", func(t *testing.T) {
		brandA = createBrandForScenario(t, "FilterBrandA")
		brandB = createBrandForScenario(t, "FilterBrandB")
		srcA = createSourceForScenario(t, "telegram", "FilterSourceA", "filter_src_a")
		srcB = createSourceForScenario(t, "web", "FilterSourceB", "https://example.com/filter-b")

		base := time.Now().UTC().Add(-2 * time.Hour)

		seeds := []mentionSeed{
			// brandA: 3 negative (srcA), 1 positive (srcB), 1 neutral (srcB)
			{BrandID: mustUUID(t, brandA), SourceID: mustUUID(t, srcA), Text: "brandA negative 1", URL: "https://a/1", PublishedAt: base.Add(1 * time.Minute), Sentiment: "negative"},
			{BrandID: mustUUID(t, brandA), SourceID: mustUUID(t, srcA), Text: "brandA negative 2", URL: "https://a/2", PublishedAt: base.Add(2 * time.Minute), Sentiment: "negative"},
			{BrandID: mustUUID(t, brandA), SourceID: mustUUID(t, srcB), Text: "brandA positive", URL: "https://a/3", PublishedAt: base.Add(3 * time.Minute), Sentiment: "positive"},
			{BrandID: mustUUID(t, brandA), SourceID: mustUUID(t, srcB), Text: "brandA neutral", URL: "https://a/4", PublishedAt: base.Add(4 * time.Minute), Sentiment: "neutral"},
			{BrandID: mustUUID(t, brandA), SourceID: mustUUID(t, srcA), Text: "brandA negative 3", URL: "https://a/5", PublishedAt: base.Add(5 * time.Minute), Sentiment: "negative"},
			// brandB: 3 negative (srcA/srcB), 1 positive (srcA), 1 neutral (srcB)
			{BrandID: mustUUID(t, brandB), SourceID: mustUUID(t, srcA), Text: "brandB negative 1", URL: "https://b/1", PublishedAt: base.Add(6 * time.Minute), Sentiment: "negative"},
			{BrandID: mustUUID(t, brandB), SourceID: mustUUID(t, srcA), Text: "brandB positive", URL: "https://b/2", PublishedAt: base.Add(7 * time.Minute), Sentiment: "positive"},
			{BrandID: mustUUID(t, brandB), SourceID: mustUUID(t, srcB), Text: "brandB neutral", URL: "https://b/3", PublishedAt: base.Add(8 * time.Minute), Sentiment: "neutral"},
			{BrandID: mustUUID(t, brandB), SourceID: mustUUID(t, srcB), Text: "brandB negative 2", URL: "https://b/4", PublishedAt: base.Add(9 * time.Minute), Sentiment: "negative"},
			{BrandID: mustUUID(t, brandB), SourceID: mustUUID(t, srcA), Text: "brandB negative 3", URL: "https://b/5", PublishedAt: base.Add(10 * time.Minute), Sentiment: "negative"},
		}

		for _, seed := range seeds {
			_ = insertMentionDirect(t, ctx, seed)
		}
	})

	t.Run("filter_by_brand", func(t *testing.T) {
		resp := doGET("/api/v1/mentions?brand_id=" + brandA)
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		items := paginatedItems(t, body)
		require.NotEmpty(t, items)
		for _, item := range items {
			m := mustMap(t, item)
			assert.Equal(t, brandA, mustString(t, m["brand_id"]))
		}
	})

	t.Run("filter_by_sentiment", func(t *testing.T) {
		resp := doGET("/api/v1/mentions?sentiment=negative")
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		items := paginatedItems(t, body)
		require.NotEmpty(t, items)
		for _, item := range items {
			m := mustMap(t, item)
			assert.Equal(t, "negative", mustString(t, m["sentiment"]))
		}
	})

	t.Run("filter_by_source", func(t *testing.T) {
		resp := doGET("/api/v1/mentions?source_id=" + srcA)
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		items := paginatedItems(t, body)
		require.NotEmpty(t, items)
		for _, item := range items {
			m := mustMap(t, item)
			src := mustMap(t, m["source"])
			assert.Equal(t, srcA, mustString(t, src["id"]))
		}
	})

	t.Run("filter_combined", func(t *testing.T) {
		// brandA + sentiment=negative → 3 items
		resp := doGET("/api/v1/mentions?brand_id=" + brandA + "&sentiment=negative")
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		items := paginatedItems(t, body)
		require.Len(t, items, 3)
		for _, item := range items {
			m := mustMap(t, item)
			assert.Equal(t, brandA, mustString(t, m["brand_id"]))
			assert.Equal(t, "negative", mustString(t, m["sentiment"]))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		resp := doGET("/api/v1/mentions?limit=2&offset=0")
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)

		items := paginatedItems(t, body)
		require.Len(t, items, 2)
		assert.Greater(t, paginatedTotal(t, body), 2)
	})

	t.Run("pagination_page2", func(t *testing.T) {
		page1Resp := doGET("/api/v1/mentions?limit=2&offset=0")
		assertStatus(t, page1Resp, http.StatusOK)
		page1Body := parseJSON(page1Resp)
		assertEnvelopeOK(t, page1Body)
		page1Items := paginatedItems(t, page1Body)
		require.Len(t, page1Items, 2)
		page1FirstID := mustString(t, mustMap(t, page1Items[0])["id"])

		page2Resp := doGET("/api/v1/mentions?limit=2&offset=2")
		assertStatus(t, page2Resp, http.StatusOK)
		page2Body := parseJSON(page2Resp)
		assertEnvelopeOK(t, page2Body)
		page2Items := paginatedItems(t, page2Body)
		require.Len(t, page2Items, 2)
		page2FirstID := mustString(t, mustMap(t, page2Items[0])["id"])

		assert.NotEqual(t, page1FirstID, page2FirstID, "page 2 should return next chunk")
	})
}

func TestScenario_AlertConfigLifecycle(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	var (
		brandID  string
		configID string
	)

	t.Run("create_brand", func(t *testing.T) {
		brandID = createBrandForScenario(t, "AlertLifecycleBrand")
	})

	t.Run("create_config", func(t *testing.T) {
		resp := doPOST("/api/v1/alerts/config", map[string]any{
			"brand_id":            brandID,
			"window_minutes":      60,
			"cooldown_minutes":    5,
			"sentiment_filter":    "negative",
			"percentile":          95,
			"anomaly_window_size": 10,
		})
		assertStatus(t, resp, http.StatusCreated)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		configID = mustString(t, mustMap(t, body["data"])["id"])
	})

	t.Run("get_config", func(t *testing.T) {
		resp := doGET("/api/v1/alerts/config/" + configID)
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		assert.Equal(t, configID, mustString(t, mustMap(t, body["data"])["id"]))
	})

	t.Run("list_by_brand", func(t *testing.T) {
		resp := doGET("/api/v1/brands/" + brandID + "/alerts/config")
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		assert.Equal(t, configID, mustString(t, mustMap(t, body["data"])["id"]))
	})

	t.Run("update_config", func(t *testing.T) {
		resp := doPUT("/api/v1/alerts/config/"+configID, map[string]any{
			"window_minutes": 120,
		})
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		assert.Equal(t, 120, int(mustFloat64(t, mustMap(t, body["data"])["window_minutes"])))
	})

	t.Run("verify_updated", func(t *testing.T) {
		resp := doGET("/api/v1/alerts/config/" + configID)
		assertStatus(t, resp, http.StatusOK)
		body := parseJSON(resp)
		assertEnvelopeOK(t, body)
		assert.Equal(t, 120, int(mustFloat64(t, mustMap(t, body["data"])["window_minutes"])))
	})

	t.Run("delete_config", func(t *testing.T) {
		resp := doDELETE("/api/v1/alerts/config/" + configID)
		assertStatus(t, resp, http.StatusNoContent)
		_ = resp.Body.Close()
	})

	t.Run("verify_deleted", func(t *testing.T) {
		resp := doGET("/api/v1/alerts/config/" + configID)
		assertStatus(t, resp, http.StatusNotFound)
		body := parseJSON(resp)
		assertHasError(t, body, "NOT_FOUND")
	})
}

// mentionSeed — данные для создания тестового упоминания напрямую в БД.
type mentionSeed struct {
	BrandID     uuid.UUID
	SourceID    uuid.UUID
	Text        string
	URL         string
	PublishedAt time.Time
	Sentiment   string
}

func createBrandForScenario(t *testing.T, name string) string {
	t.Helper()

	resp := doPOST("/api/v1/brands", map[string]any{
		"name":       name,
		"keywords":   []string{"scenario"},
		"exclusions": []string{},
		"risk_words": []string{},
	})
	assertStatus(t, resp, http.StatusCreated)
	body := parseJSON(resp)
	assertEnvelopeOK(t, body)
	return mustString(t, mustMap(t, body["data"])["id"])
}

func createSourceForScenario(t *testing.T, sourceType, name, url string) string {
	t.Helper()

	resp := doPOST("/api/v1/sources", map[string]any{
		"type": sourceType,
		"name": name,
		"url":  url,
	})
	assertStatus(t, resp, http.StatusCreated)
	body := parseJSON(resp)
	assertEnvelopeOK(t, body)
	return mustString(t, mustMap(t, body["data"])["id"])
}

// insertMentionDirect вставляет упоминание напрямую в crawler_items + sentiment_results.
// Возвращает ID записи в sentiment_results (= ID упоминания в API).
func insertMentionDirect(t *testing.T, ctx context.Context, seed mentionSeed) uuid.UUID {
	t.Helper()

	itemID := uuid.New()
	now := time.Now().UTC()

	_, err := pool.Exec(ctx, `
		INSERT INTO crawler_items (id, text, link, source_id, published_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, itemID, seed.Text, seed.URL, seed.SourceID, seed.PublishedAt, now)
	require.NoError(t, err)

	mentionID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO sentiment_results (id, item_id, brand_id, sentiment, confidence, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, mentionID, itemID, seed.BrandID, seed.Sentiment, 0.9, now)
	require.NoError(t, err)

	return mentionID
}

func insertAlertDirect(t *testing.T, ctx context.Context, configID, brandID uuid.UUID, mentionsCount int) error {
	t.Helper()

	now := time.Now().UTC()
	_, err := pool.Exec(ctx, `
		INSERT INTO alerts (
			id, config_id, brand_id, mentions_count, window_start, window_end, fired_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		uuid.New(),
		configID,
		brandID,
		mentionsCount,
		now.Add(-30*time.Minute),
		now,
		now,
	)
	return err
}

func insertEventDirect(t *testing.T, ctx context.Context, eventType string, payload map[string]any) error {
	t.Helper()

	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO events (id, type, payload, occurred_at)
		VALUES ($1, $2, $3, $4)
	`, uuid.New(), eventType, payloadRaw, time.Now().UTC())
	return err
}

func assertEnvelopeOK(t *testing.T, body map[string]any) {
	t.Helper()
	require.Contains(t, body, "data")
	require.Contains(t, body, "error")
	assert.Nil(t, body["error"])
}

func paginatedItems(t *testing.T, body map[string]any) []any {
	t.Helper()
	data := mustMap(t, body["data"])
	items, ok := data["items"].([]any)
	require.True(t, ok, "data.items must be an array")
	return items
}

func paginatedTotal(t *testing.T, body map[string]any) int {
	t.Helper()
	data := mustMap(t, body["data"])
	return int(mustFloat64(t, data["total"]))
}

func mustMap(t *testing.T, v any) map[string]any {
	t.Helper()
	m, ok := v.(map[string]any)
	require.True(t, ok, "expected map, got %T", v)
	return m
}

func mustString(t *testing.T, v any) string {
	t.Helper()
	s, ok := v.(string)
	require.True(t, ok, "expected string, got %T", v)
	return s
}

func mustBool(t *testing.T, v any) bool {
	t.Helper()
	b, ok := v.(bool)
	require.True(t, ok, "expected bool, got %T", v)
	return b
}

func mustFloat64(t *testing.T, v any) float64 {
	t.Helper()
	f, ok := v.(float64)
	require.True(t, ok, "expected float64, got %T", v)
	return f
}

func mustUUID(t *testing.T, raw string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(raw)
	require.NoError(t, err)
	return id
}

func toStringSlice(t *testing.T, v any) []string {
	t.Helper()
	rawItems, ok := v.([]any)
	require.True(t, ok, "expected []any, got %T", v)
	out := make([]string, 0, len(rawItems))
	for _, item := range rawItems {
		out = append(out, mustString(t, item))
	}
	return out
}
