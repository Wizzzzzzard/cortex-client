package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cortex-client/pkg/client"
	"gopkg.in/yaml.v3"
)

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

// RunCLI runs the main CLI logic, returns exit code
func RunCLI(args []string) int {
	flags := flag.NewFlagSet("cortex-client", flag.ContinueOnError)
	backends := flags.String("backends", "", "Comma-separated list of Prometheus backend URLs")
	backendsFile := flags.String("backends-file", "", "Path to file with Prometheus backend URLs (one per line)")
	query := flags.String("query", "up", "Prometheus query string")
	if err := flags.Parse(args); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		return 2
	}

	var backendList []string

	if *backends != "" {
		backendList = append(backendList, client.SplitAndTrim(*backends)...)
	}

	if *backendsFile != "" {
		fileBackends, err := readBackendsFile(*backendsFile)
		if err != nil {
			fmt.Printf("Error reading backends file: %v\n", err)
			return 1
		}
		backendList = append(backendList, fileBackends...)
	}

	if len(backendList) == 0 {
		fmt.Println("Please provide at least one backend URL with --backends or --backends-file")
		return 1
	}

	b, err := client.MergePrometheusQueries(backendList, *query)
	if err != nil {
		fmt.Printf("Error merging queries: %v\n", err)
		return 1
	}
	fmt.Printf("Merged response:\n%s\n", string(b))
	return 0
}

func main() {
	os.Exit(RunCLI(os.Args[1:]))
}
