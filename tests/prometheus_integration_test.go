package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

type TargetsResponse struct {
	Status string `json:"status"`
}

func TestPrometheusAPI(t *testing.T) {
	// Check if containers are already running
	checkCmd := exec.Command("docker", "ps", "--filter", "ancestor=prom/prometheus:main", "--format", "{{.ID}}")
	output, err := checkCmd.Output()
	alreadyRunning := false
	if err == nil && len(output) > 0 {
		alreadyRunning = true
	}

	if !alreadyRunning {
		startCmd := exec.Command("../start-prometheus.sh", "start")
		startCmd.Stdout = io.Discard
		startCmd.Stderr = io.Discard
		if err := startCmd.Run(); err != nil {
			t.Fatalf("failed to start prometheus containers: %v", err)
		}
		// Ensure cleanup only if we started them
		defer func() {
			stopCmd := exec.Command("../start-prometheus.sh", "stop")
			stopCmd.Stdout = io.Discard
			stopCmd.Stderr = io.Discard
			_ = stopCmd.Run()
		}()
	}

	ports := []string{"9090", "9091"}
	for _, port := range ports {
		ready := false
		for i := 0; i < 20; i++ {
			resp, err := http.Get("http://localhost:" + port + "/api/v1/targets")
			if err == nil && resp.StatusCode == 200 {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				t.Logf("Prometheus API response on port %s: %s", port, string(body))
				var tr TargetsResponse
				json.Unmarshal(body, &tr)
				if tr.Status == "success" {
					ready = true
					break
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
		if !ready {
			t.Fatalf("Prometheus API did not become ready in time on port %s", port)
		}
	}
}
