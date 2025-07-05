package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cortex-client/pkg/client"
)

// mockPrometheusHandler returns an http.HandlerFunc that mocks the Prometheus API.
func mockPrometheusHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/query" && r.URL.Query().Get("query") == "up" {
			resp := struct {
				Status string      `json:"status"`
				Data   interface{} `json:"data"`
			}{
				Status: "success",
				Data:   map[string]interface{}{"resultType": "vector", "result": []interface{}{}},
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Errorf("failed to encode response: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}
}

func TestClientHitsMockPrometheus(t *testing.T) {
	ts := httptest.NewServer(mockPrometheusHandler(t))
	defer ts.Close()

	t.Run("query up returns success", func(t *testing.T) {
		backends := []string{ts.URL}
		query := "up"
		output, err := client.MergePrometheusQueries(backends, query)

		t.Logf("Client output:\n%s", string(output))

		if err != nil {
			t.Fatalf("Client returned error: %v\nOutput:\n%s", err, string(output))
		}

		if !strings.Contains(string(output), "success") {
			t.Fatalf("Expected 'success' in client output, got:\n%s", string(output))
		}
	})

	t.Log("Test passed: client received a successful response from the mock Prometheus API.")
}
