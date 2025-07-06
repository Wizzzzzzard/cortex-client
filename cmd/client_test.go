package main

import (
	"bytes"
	"io"
	"os"
	"strings"
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

func stubMergePrometheusQueries(output string, err error) func([]string, string) ([]byte, error) {
	return func(_ []string, _ string) ([]byte, error) {
		if err != nil {
			return nil, err
		}
		return []byte(output), nil
	}
}

func captureOutput(f func()) (string, string) {
	oldOut, oldErr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr
	defer func() {
		os.Stdout, os.Stderr = oldOut, oldErr
	}()
	outC := make(chan string)
	errC := make(chan string)
	go func() { var buf bytes.Buffer; io.Copy(&buf, rOut); outC <- buf.String() }()
	go func() { var buf bytes.Buffer; io.Copy(&buf, rErr); errC <- buf.String() }()
	f()
	wOut.Close(); wErr.Close()
	out, err := <-outC, <-errC
	return out, err
}

func TestRunCLI_NoBackends(t *testing.T) {
	out, _ := captureOutput(func() {
		code := RunCLIWithMergeFunc([]string{"--query=up"}, stubMergePrometheusQueries("", nil))
		if code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	})
	if !strings.Contains(out, "Please provide at least one backend") {
		t.Errorf("expected error message for missing backends, got: %s", out)
	}
}

func TestRunCLI_InvalidBackendsFile(t *testing.T) {
	out, _ := captureOutput(func() {
		code := RunCLIWithMergeFunc([]string{"--backends-file=nonexistent.yaml"}, stubMergePrometheusQueries("", nil))
		if code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	})
	if !strings.Contains(out, "Error reading backends file") {
		t.Errorf("expected error message for invalid file, got: %s", out)
	}
}

func TestRunCLI_ValidBackendsFlag(t *testing.T) {
	out, _ := captureOutput(func() {
		code := RunCLIWithMergeFunc([]string{"--backends=http://localhost:9090", "--query=up"}, stubMergePrometheusQueries("{\"status\":\"success\"}", nil))
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})
	if !strings.Contains(out, "Merged response") || !strings.Contains(out, "success") {
		t.Errorf("expected merged response with success, got: %s", out)
	}
}

func TestRunCLI_ValidBackendsFile(t *testing.T) {
	f, err := os.CreateTemp("", "backends-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	content := []byte("prometheus_backends:\n  - http://localhost:9090\n")
	if _, err := f.Write(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	out, _ := captureOutput(func() {
		code := RunCLIWithMergeFunc([]string{"--backends-file=" + f.Name(), "--query=up"}, stubMergePrometheusQueries("{\"status\":\"success\"}", nil))
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})
	if !strings.Contains(out, "Merged response") || !strings.Contains(out, "success") {
		t.Errorf("expected merged response with success, got: %s", out)
	}
}

func TestRunCLI_InvalidFlag(t *testing.T) {
	out, _ := captureOutput(func() {
		code := RunCLIWithMergeFunc([]string{"--notaflag"}, stubMergePrometheusQueries("", nil))
		if code != 2 {
			t.Errorf("expected exit code 2, got %d", code)
		}
	})
	if !strings.Contains(out, "Error parsing flags") {
		t.Errorf("expected flag parsing error, got: %s", out)
	}
}
