//go:build integration

package e2e

import (
	"net/http"
	"testing"
)

func TestHealth_AllUp(t *testing.T) {
	t.Cleanup(func() { cleanupAllTables(t) })

	resp := doGET("/health")
	assertStatus(t, resp, http.StatusOK)
	body := parseJSON(resp)
	assertNoError(t, body)

	data := body["data"].(map[string]any)
	if data["status"] != "ok" {
		t.Fatalf("expected status=ok, got=%v", data["status"])
	}
	deps := data["dependencies"].(map[string]any)
	if deps["postgres"] != "ok" {
		t.Fatalf("postgres dependency must be ok")
	}
	if deps["redis"] != "ok" {
		t.Fatalf("redis dependency must be ok")
	}
	if data["version"] != "1.0.0" {
		t.Fatalf("expected version 1.0.0")
	}
}
