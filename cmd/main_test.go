package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/cortex-client/pkg/client"
)

func stubMergePrometheusQueries(output string, err error) func(client.QueryData) ([]byte, error) {
	return func(_ client.QueryData) ([]byte, error) {
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
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, rOut)
		if err != nil {
			log.Fatalf("failed to read stdout: %v", err)
		}
		outC <- buf.String()
	}()
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, rErr)
		if err != nil {
			log.Fatalf("failed to read stderr: %v", err)
		}
		errC <- buf.String()
	}()
	f()
	closeWoutErr := wOut.Close()
	if closeWoutErr != nil {
		log.Fatalf("failed to close stdout: %v", closeWoutErr)
	}
	closeWerrErr := wErr.Close()
	if closeWerrErr != nil {
		log.Fatalf("failed to close stderr: %v", closeWerrErr)
	}
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
	file, err := os.CreateTemp("", "backends-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil {
			t.Fatalf("failed to remove temp file: %v", removeErr)
		}
	}()
	content := []byte("prometheus_backends:\n  - http://localhost:9090\n")
	if _, err := file.Write(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Fatalf("failed to close temp file: %v", closeErr)
		}
	}()
	out, _ := captureOutput(func() {
		code := RunCLIWithMergeFunc([]string{"--backends-file=" + file.Name(), "--query=up"}, stubMergePrometheusQueries("{\"status\":\"success\"}", nil))
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
