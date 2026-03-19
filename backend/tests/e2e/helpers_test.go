//go:build integration

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func doGET(path string) *http.Response {
	resp, err := httpClient.Get(baseURL + path)
	if err != nil {
		panic(fmt.Sprintf("doGET failed: %v", err))
	}
	return resp
}

func doPOST(path string, body interface{}) *http.Response {
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			panic(fmt.Sprintf("doPOST marshal failed: %v", err))
		}
		payload = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+path, payload)
	if err != nil {
		panic(fmt.Sprintf("doPOST request failed: %v", err))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("doPOST failed: %v", err))
	}
	return resp
}

func doPUT(path string, body interface{}) *http.Response {
	raw, err := json.Marshal(body)
	if err != nil {
		panic(fmt.Sprintf("doPUT marshal failed: %v", err))
	}

	req, err := http.NewRequest(http.MethodPut, baseURL+path, bytes.NewReader(raw))
	if err != nil {
		panic(fmt.Sprintf("doPUT request failed: %v", err))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("doPUT failed: %v", err))
	}
	return resp
}

func doDELETE(path string) *http.Response {
	req, err := http.NewRequest(http.MethodDelete, baseURL+path, nil)
	if err != nil {
		panic(fmt.Sprintf("doDELETE request failed: %v", err))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("doDELETE failed: %v", err))
	}
	return resp
}

func parseJSON(resp *http.Response) map[string]any {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("parseJSON read failed: %v", err))
	}
	if len(bodyBytes) == 0 {
		return map[string]any{}
	}

	var decoded map[string]any
	if err := json.Unmarshal(bodyBytes, &decoded); err != nil {
		panic(fmt.Sprintf("parseJSON decode failed: %v body=%s", err, string(bodyBytes)))
	}
	return decoded
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		bodyBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		t.Fatalf("unexpected status: got=%d want=%d body=%s", resp.StatusCode, expected, string(bodyBytes))
	}
}

func assertNoError(t *testing.T, body map[string]any) {
	t.Helper()
	if body["error"] != nil {
		t.Fatalf("expected no error, got: %#v", body["error"])
	}
}

func assertHasError(t *testing.T, body map[string]any, code string) {
	t.Helper()
	errObj, ok := body["error"].(map[string]any)
	if !ok || errObj == nil {
		t.Fatalf("expected error object, got: %#v", body["error"])
	}
	gotCode, _ := errObj["code"].(string)
	if gotCode != code {
		t.Fatalf("unexpected error code: got=%q want=%q", gotCode, code)
	}
}

func cleanupTable(t *testing.T, table string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	if err != nil {
		t.Fatalf("cleanup table %s failed: %v", table, err)
	}
}

func cleanupAllTables(t *testing.T) {
	t.Helper()
	cleanupTable(t, "events")
	cleanupTable(t, "alerts")
	cleanupTable(t, "alert_configs")
	cleanupTable(t, "sentiment_results")
	cleanupTable(t, "crawler_items")
	cleanupTable(t, "sources")
	cleanupTable(t, "brands")
}
