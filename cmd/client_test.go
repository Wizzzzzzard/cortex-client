package main

import (
	"os"
	"testing"
)

func TestReadBackendsFile(t *testing.T) {
	// Create a temp YAML file
	f, err := os.CreateTemp("", "backends-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	content := []byte("prometheus_backends:\n  - http://localhost:9090\n  - http://localhost:9091\n")
	if _, err := f.Write(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()

	backends, err := readBackendsFile(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(backends) != 2 || backends[0] != "http://localhost:9090" || backends[1] != "http://localhost:9091" {
		t.Errorf("unexpected backends: %v", backends)
	}
}

func TestReadBackendsFile_Error(t *testing.T) {
	_, err := readBackendsFile("nonexistent.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}
