//go:build integration

package e2e

import (
	"net/http"
	"testing"
)

func TestBrand_CRUD(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	createResp := doPOST("/api/v1/brands", map[string]any{
		"name":       "TestBrand",
		"keywords":   []string{"test"},
		"exclusions": []string{},
		"risk_words": []string{},
	})
	assertStatus(t, createResp, http.StatusCreated)
	createBody := parseJSON(createResp)
	assertNoError(t, createBody)

	data := createBody["data"].(map[string]any)
	brandID := data["id"].(string)
	if brandID == "" {
		t.Fatalf("brand id is empty")
	}
	if data["name"] != "TestBrand" {
		t.Fatalf("unexpected brand name: %#v", data["name"])
	}

	getResp := doGET("/api/v1/brands/" + brandID)
	assertStatus(t, getResp, http.StatusOK)
	getBody := parseJSON(getResp)
	assertNoError(t, getBody)
	if getBody["data"].(map[string]any)["name"] != "TestBrand" {
		t.Fatalf("unexpected name in get response")
	}

	updateResp := doPUT("/api/v1/brands/"+brandID, map[string]any{
		"name":       "UpdatedBrand",
		"keywords":   []string{"test"},
		"exclusions": []string{},
		"risk_words": []string{},
	})
	assertStatus(t, updateResp, http.StatusOK)
	updateBody := parseJSON(updateResp)
	assertNoError(t, updateBody)
	if updateBody["data"].(map[string]any)["name"] != "UpdatedBrand" {
		t.Fatalf("unexpected updated name")
	}

	listResp := doGET("/api/v1/brands")
	assertStatus(t, listResp, http.StatusOK)
	listBody := parseJSON(listResp)
	assertNoError(t, listBody)
	items := listBody["data"].(map[string]any)["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("expected at least one brand in list")
	}

	deleteResp := doDELETE("/api/v1/brands/" + brandID)
	assertStatus(t, deleteResp, http.StatusNoContent)
	_ = deleteResp.Body.Close()

	getDeletedResp := doGET("/api/v1/brands/" + brandID)
	assertStatus(t, getDeletedResp, http.StatusNotFound)
	getDeletedBody := parseJSON(getDeletedResp)
	assertHasError(t, getDeletedBody, "NOT_FOUND")
}
