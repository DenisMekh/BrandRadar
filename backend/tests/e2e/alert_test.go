//go:build integration

package e2e

import (
	"net/http"
	"testing"
)

func TestAlertConfig_CRUD(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	brandResp := doPOST("/api/v1/brands", map[string]any{
		"name":       "AlertBrand",
		"keywords":   []string{"alert"},
		"exclusions": []string{},
		"risk_words": []string{},
	})
	assertStatus(t, brandResp, http.StatusCreated)
	brandBody := parseJSON(brandResp)
	brandID := brandBody["data"].(map[string]any)["id"].(string)

	createResp := doPOST("/api/v1/alerts/config", map[string]any{
		"brand_id":            brandID,
		"window_minutes":      60,
		"cooldown_minutes":    30,
		"sentiment_filter":    "negative",
		"percentile":          95,
		"anomaly_window_size": 10,
	})
	assertStatus(t, createResp, http.StatusCreated)
	createBody := parseJSON(createResp)
	assertNoError(t, createBody)
	configID := createBody["data"].(map[string]any)["id"].(string)

	getResp := doGET("/api/v1/alerts/config/" + configID)
	assertStatus(t, getResp, http.StatusOK)
	getBody := parseJSON(getResp)
	assertNoError(t, getBody)

	updateResp := doPUT("/api/v1/alerts/config/"+configID, map[string]any{
		"window_minutes": 120,
	})
	assertStatus(t, updateResp, http.StatusOK)
	updateBody := parseJSON(updateResp)
	assertNoError(t, updateBody)
	if int(updateBody["data"].(map[string]any)["window_minutes"].(float64)) != 120 {
		t.Fatalf("expected window_minutes=120")
	}

	byBrandResp := doGET("/api/v1/brands/" + brandID + "/alerts/config")
	assertStatus(t, byBrandResp, http.StatusOK)
	byBrandBody := parseJSON(byBrandResp)
	assertNoError(t, byBrandBody)
	if byBrandBody["data"].(map[string]any)["id"].(string) != configID {
		t.Fatalf("unexpected alert config by brand")
	}

	deleteResp := doDELETE("/api/v1/alerts/config/" + configID)
	assertStatus(t, deleteResp, http.StatusNoContent)
	_ = deleteResp.Body.Close()
}

func TestAlertConfig_InvalidWindowMinutes(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	brandResp := doPOST("/api/v1/brands", map[string]any{
		"name":       "AlertInvalidBrand",
		"keywords":   []string{"alert"},
		"exclusions": []string{},
		"risk_words": []string{},
	})
	assertStatus(t, brandResp, http.StatusCreated)
	brandBody := parseJSON(brandResp)
	brandID := brandBody["data"].(map[string]any)["id"].(string)

	resp := doPOST("/api/v1/alerts/config", map[string]any{
		"brand_id":            brandID,
		"window_minutes":      -1,
		"cooldown_minutes":    30,
		"percentile":          95,
		"anomaly_window_size": 10,
	})
	assertStatus(t, resp, http.StatusBadRequest)
	body := parseJSON(resp)
	assertHasError(t, body, "VALIDATION_ERROR")
}
