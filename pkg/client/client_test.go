package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestSplitAndTrim(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{"  a, b ,c  ", []string{"a", "b", "c"}},
		{"", nil},
		{" , , ", nil},
	}
	for _, c := range cases {
		result := SplitAndTrim(c.input)
		if !reflect.DeepEqual(result, c.expected) {
			t.Errorf("SplitAndTrim(%q) = %v, want %v", c.input, result, c.expected)
		}
	}
}

func TestReadBackendsFile(t *testing.T) {
	// Create a temp YAML file
	file, err := os.CreateTemp("", "backends-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil {
			t.Fatalf("failed to remove temp file: %v", removeErr)
		}
	}()
	content := []byte("prometheus_backends:\n  - http://localhost:9090\n  - http://localhost:9091\n")
	if _, err := file.Write(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Fatalf("failed to close temp file: %v", closeErr)
		}
	}()

	backends, err := ReadBackendFile(file.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(backends) != 2 || backends[0] != "http://localhost:9090" || backends[1] != "http://localhost:9091" {
		t.Errorf("unexpected backends: %v", backends)
	}
}

func TestReadBackendFile_Error(t *testing.T) {
	_, err := ReadBackendFile("nonexistent.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestQueryPrometheus_Error(t *testing.T) {
	_, err := QueryPrometheus("http://invalid:9999", "up")
	if err == nil {
		t.Error("expected error for invalid backend, got nil")
	}
}

func TestMergePrometheusQueries_EmptyBackends(t *testing.T) {
	testQueryData := QueryData{
		Query:    "up",
		Backends: []string{},
	}
	output, err := MergePrometheusQueries(testQueryData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var merged struct {
		Status string            `json:"status"`
		Data   []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(output, &merged); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if merged.Status != "success" {
		t.Errorf("expected status 'success', got %q", merged.Status)
	}
	if len(merged.Data) != 0 {
		t.Errorf("expected no data, got %d", len(merged.Data))
	}
}

func TestQueryPrometheus_MockServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()
	resp, err := QueryPrometheus(ts.URL, "up")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "success" {
		t.Errorf("expected status 'success', got %q", resp.Status)
	}
}

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
		testQueryData := QueryData{
			Query:    "up",
			Backends: []string{ts.URL},
		}
		output, err := MergePrometheusQueries(testQueryData)

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
