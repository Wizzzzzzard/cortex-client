package tests

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

type TargetsResponse struct {
	Status string `json:"status"`
}

func TestPrometheusAPI(t *testing.T) {
	promPath := "../prometheus"
	configs := []struct {
		configPath string
		dataPath   string
		port       string
	}{
		{"../prometheus1/prometheus.yml", "../prometheus1/data", "9090"},
		{"../prometheus2/prometheus.yml", "../prometheus2/data", "9091"},
	}

	cmds := make([]*exec.Cmd, 0, len(configs))
	for _, c := range configs {
		cmd := exec.Command(promPath,
			"--config.file="+c.configPath,
			"--storage.tsdb.path="+c.dataPath,
			"--web.listen-address=:"+c.port,
		)
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start prometheus on port %s: %v", c.port, err)
		}
		cmds = append(cmds, cmd)
		defer cmd.Process.Kill()
	}

	for _, c := range configs {
		ready := false
		for i := 0; i < 20; i++ {
			resp, err := http.Get("http://localhost:" + c.port + "/api/v1/targets")
			if err == nil && resp.StatusCode == 200 {
				body, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				t.Logf("Prometheus API response on port %s: %s", c.port, string(body))
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
			t.Fatalf("Prometheus API did not become ready in time on port %s", c.port)
		}
	}
}
