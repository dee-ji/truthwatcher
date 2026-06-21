package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"truthwatcher/internal/api"
)

func main() {
	out := flag.String("out", "docs/openapi.json", "path to write generated OpenAPI JSON")
	version := flag.String("version", "dev", "API version to write into the OpenAPI document")
	flag.Parse()

	data, err := json.MarshalIndent(api.OpenAPISpec(*version), "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal openapi spec: %v\n", err)
		os.Exit(1)
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "create output directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write openapi spec: %v\n", err)
		os.Exit(1)
	}
}
