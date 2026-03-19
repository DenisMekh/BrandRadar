//go:build integration

package e2e

import (
	"net/http"
	"testing"
)

func TestSource_CRUD(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	brandResp := doPOST("/api/v1/brands", map[string]any{
		"name":       "BrandForSource",
		"keywords":   []string{"source"},
		"exclusions": []string{},
		"risk_words": []string{},
	})
	assertStatus(t, brandResp, http.StatusCreated)
	_ = parseJSON(brandResp)

	createResp := doPOST("/api/v1/sources", map[string]any{
		"type": "telegram",
		"name": "E2E source",
		"url":  "e2etest",
	})
	assertStatus(t, createResp, http.StatusCreated)
	createBody := parseJSON(createResp)
	assertNoError(t, createBody)
	sourceID := createBody["data"].(map[string]any)["id"].(string)

	getResp := doGET("/api/v1/sources/" + sourceID)
	assertStatus(t, getResp, http.StatusOK)
	getBody := parseJSON(getResp)
	assertNoError(t, getBody)

	listResp := doGET("/api/v1/sources")
	assertStatus(t, listResp, http.StatusOK)
	listBody := parseJSON(listResp)
	assertNoError(t, listBody)
	items := listBody["data"].(map[string]any)["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("expected at least one source")
	}

	toggleResp := doPOST("/api/v1/sources/"+sourceID+"/toggle", nil)
	assertStatus(t, toggleResp, http.StatusOK)
	toggleBody := parseJSON(toggleResp)
	assertNoError(t, toggleBody)
	status := toggleBody["data"].(map[string]any)["status"].(string)
	if status != "inactive" {
		t.Fatalf("expected toggled status inactive, got=%s", status)
	}

	deleteResp := doDELETE("/api/v1/sources/" + sourceID)
	assertStatus(t, deleteResp, http.StatusNoContent)
	_ = deleteResp.Body.Close()

	getDeletedResp := doGET("/api/v1/sources/" + sourceID)
	assertStatus(t, getDeletedResp, http.StatusNotFound)
	getDeletedBody := parseJSON(getDeletedResp)
	assertHasError(t, getDeletedBody, "NOT_FOUND")
}
