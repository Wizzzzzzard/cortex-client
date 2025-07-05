package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// PrometheusResponse represents a generic response from Prometheus
type PrometheusResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

// queryPrometheus queries a single Prometheus backend
func queryPrometheus(backendURL, query string) (*PrometheusResponse, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", strings.TrimRight(backendURL, "/"), query)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result PrometheusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func main() {
	backends := flag.String("backends", "", "Comma-separated list of Prometheus backend URLs")
	backendsFile := flag.String("backends-file", "", "Path to file with Prometheus backend URLs (one per line)")
	query := flag.String("query", "up", "Prometheus query string")
	flag.Parse()

	var backendList []string

	if *backends != "" {
		backendList = append(backendList, splitAndTrim(*backends)...)
	}

	if *backendsFile != "" {
		fileBackends, err := readBackendsFile(*backendsFile)
		if err != nil {
			fmt.Printf("Error reading backends file: %v\n", err)
			os.Exit(1)
		}
		backendList = append(backendList, fileBackends...)
	}

	if len(backendList) == 0 {
		fmt.Println("Please provide at least one backend URL with --backends or --backends-file")
		os.Exit(1)
	}

	var merged struct {
		Status string          `json:"status"`
		Data   []json.RawMessage `json:"data"`
	}
	merged.Status = "success"

	for _, backend := range backendList {
		if backend == "" {
			continue
		}
		resp, err := queryPrometheus(backend, *query)
		if err != nil {
			fmt.Printf("Error querying %s: %v\n", backend, err)
			continue
		}
		merged.Data = append(merged.Data, resp.Data)
	}

	b, _ := json.MarshalIndent(merged, "", "  ")
	fmt.Printf("Merged response:\n%s\n", string(b))
}

func splitAndTrim(s string) []string {
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

// readBackendsFile reads a YAML file with prometheus_backends as a list
func readBackendsFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		PrometheusBackends []string `yaml:"prometheus_backends"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	return parsed.PrometheusBackends, nil
}

