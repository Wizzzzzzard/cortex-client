package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

type PrometheusResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

type QueryData struct {
	Query    string
	Backends []string
}

type PrometheusQueryJob struct {
	BackendURL string
	Query      string
}

func prometheusQueryWorker(jobs <-chan PrometheusQueryJob, results chan<- *PrometheusResponse, wg *sync.WaitGroup) {
	for job := range jobs {
		resp, err := QueryPrometheus(job.BackendURL, job.Query)
		if err != nil {
			log.Printf("error querying backend %s: %v", job.BackendURL, err)
			continue
		}
		results <- resp
		wg.Done()
	}
}

// ReadBackendFile reads a YAML file with prometheus_backends as a list
func ReadBackendFile(path string) ([]string, error) {
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

// QueryPrometheus queries a single Prometheus backend
func QueryPrometheus(backendURL, query string) (*PrometheusResponse, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", strings.TrimRight(backendURL, "/"), query)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// Optionally log or handle the error, e.g.:
			fmt.Printf("error closing response body: %v\n", cerr)
		}
	}()
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
func MergePrometheusQueries(data QueryData) ([]byte, error) {
	var merged struct {
		Status string            `json:"status"`
		Data   []json.RawMessage `json:"data"`
	}

	merged.Status = "success"
	numWorkers := 5

	jobs := make(chan PrometheusQueryJob, len(data.Backends))
	results := make(chan *PrometheusResponse, len(data.Backends))
	var wg sync.WaitGroup

	for range numWorkers {
		go prometheusQueryWorker(jobs, results, &wg)
	}

	wg.Add(len(data.Backends))
	for _, backend := range data.Backends {
		if backend == "" {
			continue
		}
		jobs <- PrometheusQueryJob{
			BackendURL: backend,
			Query:      data.Query,
		}
	}
	close(jobs)
	wg.Wait()

	for range data.Backends {
		resp := <-results
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
