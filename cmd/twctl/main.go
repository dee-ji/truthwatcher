package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsageAndExit()
	}
	api := os.Getenv("TW_API")
	if api == "" {
		api = "http://localhost:8080"
	}

	switch {
	case len(os.Args) >= 4 && os.Args[1] == "intent" && os.Args[2] == "validate":
		runIntentValidate(os.Args[3])
	case len(os.Args) >= 4 && os.Args[1] == "render" && os.Args[2] == "preview":
		vendor := "junos"
		if len(os.Args) >= 5 && strings.HasPrefix(os.Args[4], "--vendor=") {
			vendor = strings.TrimPrefix(os.Args[4], "--vendor=")
		}
		runRenderPreview(api, os.Args[3], vendor)
	case len(os.Args) >= 4 && os.Args[1] == "deploy" && os.Args[2] == "create":
		runDeployCreate(api, os.Args[3], os.Args[4:])
	case len(os.Args) >= 4 && os.Args[1] == "deploy" && os.Args[2] == "get":
		runDeployGet(api, os.Args[3])

	case len(os.Args) >= 4 && os.Args[1] == "state" && os.Args[2] == "compare":
		runStateCompare(api, os.Args[3])
	default:
		printUsageAndExit()
	}
}

func printUsageAndExit() {
	fmt.Println("usage: twctl intent validate <file>|twctl render preview <id> [--vendor=<name>]|twctl deploy create <intent-id> --idempotency-key=<key> [--target=<device>] [--mode=dry-run]|twctl deploy get <deployment-id>|twctl state compare <intent-id>")
	os.Exit(1)
}

func runIntentValidate(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	text := string(content)
	if !strings.Contains(text, "metadata:") || !strings.Contains(text, "name:") {
		fmt.Println("invalid intent: metadata.name is required")
		os.Exit(1)
	}
	fmt.Println("intent valid")
}

func runRenderPreview(api, id, vendor string) {
	body, _ := json.Marshal(map[string]string{"vendor": vendor})
	resp, err := http.Post(api+"/api/v1/intents/"+id+"/compile", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	getResp, err := http.Get(api + "/api/v1/intents/" + id)
	if err != nil {
		panic(err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode >= 300 {
		b, _ := io.ReadAll(getResp.Body)
		fmt.Println(string(b))
		os.Exit(1)
	}
	var out struct {
		Artifacts []struct {
			Vendor   string            `json:"vendor"`
			Artifact string            `json:"artifact"`
			Metadata map[string]string `json:"metadata"`
		} `json:"artifacts"`
	}
	_ = json.NewDecoder(getResp.Body).Decode(&out)
	if len(out.Artifacts) == 0 {
		fmt.Println("no artifacts generated")
		return
	}
	fmt.Printf("vendor=%s metadata=%v\n%s\n", out.Artifacts[0].Vendor, out.Artifacts[0].Metadata, out.Artifacts[0].Artifact)
}

func runDeployCreate(api, intentID string, args []string) {
	mode := "dry-run"
	idempotencyKey := ""
	targets := []string{}
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--idempotency-key="):
			idempotencyKey = strings.TrimPrefix(arg, "--idempotency-key=")
		case strings.HasPrefix(arg, "--target="):
			targets = append(targets, strings.TrimPrefix(arg, "--target="))
		case strings.HasPrefix(arg, "--mode="):
			mode = strings.TrimPrefix(arg, "--mode=")
		}
	}
	if idempotencyKey == "" {
		fmt.Println("--idempotency-key is required")
		os.Exit(1)
	}
	body, _ := json.Marshal(map[string]any{
		"intent_id":       intentID,
		"idempotency_key": idempotencyKey,
		"mode":            mode,
		"targets":         targets,
	})
	resp, err := http.Post(api+"/api/v1/deployments", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		fmt.Printf("deployment create failed: %s\n", string(payload))
		os.Exit(1)
	}
	fmt.Println(string(payload))
}

func runDeployGet(api, deploymentID string) {
	resp, err := http.Get(api + "/api/v1/deployments/" + deploymentID)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		fmt.Printf("deployment get failed: %s\n", string(payload))
		os.Exit(1)
	}
	fmt.Println(string(payload))
}

func runStateCompare(api, intentID string) {
	body, _ := json.Marshal(map[string]any{"intent_id": intentID, "actor": "twctl"})
	resp, err := http.Post(api+"/api/v1/reconcile/runs", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		fmt.Printf("state compare failed: %s\n", string(payload))
		os.Exit(1)
	}
	fmt.Println(string(payload))
}
