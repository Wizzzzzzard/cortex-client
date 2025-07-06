package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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
