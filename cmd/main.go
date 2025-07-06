package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cortex-client/pkg/client"
)

// RunCLI runs the main CLI logic, returns exit code
func RunCLI(args []string) int {
	return RunCLIWithMergeFunc(args, client.MergePrometheusQueries)
}

// RunCLIWithMergeFunc allows injecting a merge function for testing
func RunCLIWithMergeFunc(args []string, mergeFunc func(client.QueryData) ([]byte, error)) int {
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
		fileBackends, err := client.ReadBackendFile(*backendsFile)
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

	queryData := client.QueryData{
		Query:    *query,
		Backends: backendList,
	}

	b, err := mergeFunc(queryData)
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
