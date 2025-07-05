package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/cortex-client/pkg/client"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TargetsResponse struct {
	Status string `json:"status"`
}

func TestPrometheusAPI(t *testing.T) {
	ctx := context.Background()

	proms := []struct {
		name string
		port string
	}{
		{"prometheus1", "9090"},
		{"prometheus2", "9091"},
	}

	containers := make([]testcontainers.Container, 0, len(proms))
	backendURLs := make([]string, 0, len(proms))

	for _, prom := range proms {
		portSpec := nat.Port(prom.port + "/tcp")
		// Use a static prometheus.yml from test resources, one per port
		configFile := fmt.Sprintf("prometheus_%s.yml", prom.port)
		configPath, err := filepath.Abs(filepath.Join("..", "resources", configFile))
		if err != nil {
			t.Fatalf("failed to get absolute path for %s: %v", configFile, err)
		}

		cmd := []string{"--config.file=/etc/prometheus/prometheus.yml"}
		if prom.port != "9090" {
			cmd = append(cmd, fmt.Sprintf("--web.listen-address=:%s", prom.port))
		}

		req := testcontainers.ContainerRequest{
			Image:        "prom/prometheus:main",
			ExposedPorts: []string{string(portSpec)},
			WaitingFor:   wait.ForHTTP("/api/v1/targets").WithPort(portSpec),
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      configPath,
					ContainerFilePath: "/etc/prometheus/prometheus.yml",
					FileMode:          0o644,
				},
			},
			Cmd: cmd,
		}
		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			t.Fatalf("failed to start prometheus container %s: %v", prom.name, err)
		}
		containers = append(containers, container)

		host, err := container.Host(ctx)
		if err != nil {
			t.Fatalf("failed to get host for %s: %v", prom.name, err)
		}
		mappedPort, err := container.MappedPort(ctx, portSpec)
		if err != nil {
			t.Fatalf("failed to get mapped port for %s: %v", prom.name, err)
		}
		backendURLs = append(backendURLs, fmt.Sprintf("http://%s:%s", host, mappedPort.Port()))
	}
	defer func() {
		for _, c := range containers {
			_ = c.Terminate(ctx)
		}
	}()

	// Wait a bit for Prometheus to be fully ready
	time.Sleep(2 * time.Second)

	output, err := client.MergePrometheusQueries(backendURLs, "up")
	if err != nil {
		t.Fatalf("Error querying Prometheus containers: %v", err)
	}

	var merged struct {
		Status string            `json:"status"`
		Data   []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(output, &merged); err != nil {
		t.Fatalf("Failed to unmarshal merged response: %v", err)
	}

	if merged.Status != "success" {
		t.Fatalf("Expected status 'success', got: %s", merged.Status)
	}

	if len(merged.Data) != 2 {
		t.Fatalf("Expected data from 2 containers, got: %d", len(merged.Data))
	}

	t.Logf("Test passed: received merged response from both Prometheus containers. Output:\n%s", string(output))
}
