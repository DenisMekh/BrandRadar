//go:build integration

package e2e

import (
	"net/http"
	"testing"
)

func TestEvent_List(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	brandResp := doPOST("/api/v1/brands", map[string]any{
		"name":       "EventBrand",
		"keywords":   []string{"event"},
		"exclusions": []string{},
		"risk_words": []string{},
	})
	assertStatus(t, brandResp, http.StatusCreated)
	_ = parseJSON(brandResp)

	sourceResp := doPOST("/api/v1/sources", map[string]any{
		"type": "telegram",
		"name": "Event source",
		"url":  "events_e2e",
	})
	assertStatus(t, sourceResp, http.StatusCreated)
	sourceBody := parseJSON(sourceResp)
	sourceID := sourceBody["data"].(map[string]any)["id"].(string)

	// Generates source_toggled event in usecase.
	toggleResp := doPOST("/api/v1/sources/"+sourceID+"/toggle", nil)
	assertStatus(t, toggleResp, http.StatusOK)
	_ = parseJSON(toggleResp)

	listResp := doGET("/api/v1/events")
	assertStatus(t, listResp, http.StatusOK)
	listBody := parseJSON(listResp)
	assertNoError(t, listBody)
	items := listBody["data"].(map[string]any)["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("expected non-empty events list")
	}

	filterResp := doGET("/api/v1/events?type=source_toggled")
	assertStatus(t, filterResp, http.StatusOK)
	filterBody := parseJSON(filterResp)
	assertNoError(t, filterBody)
	filterItems := filterBody["data"].(map[string]any)["items"].([]any)
	if len(filterItems) == 0 {
		t.Fatalf("expected non-empty source_toggled events")
	}
}
