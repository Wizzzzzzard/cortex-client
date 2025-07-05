package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type PrometheusResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

// QueryPrometheus queries a single Prometheus backend
func QueryPrometheus(backendURL, query string) (*PrometheusResponse, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", strings.TrimRight(backendURL, "/"), query)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result PrometheusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MergePrometheusQueries queries all backends and merges the results
func MergePrometheusQueries(backends []string, query string) ([]byte, error) {
	var merged struct {
		Status string            `json:"status"`
		Data   []json.RawMessage `json:"data"`
	}
	merged.Status = "success"
	for _, backend := range backends {
		if backend == "" {
			continue
		}
		resp, err := QueryPrometheus(backend, query)
		if err != nil {
			continue
		}
		merged.Data = append(merged.Data, resp.Data)
	}
	return json.MarshalIndent(merged, "", "  ")
}

// SplitAndTrim splits a comma-separated string and trims spaces
func SplitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
