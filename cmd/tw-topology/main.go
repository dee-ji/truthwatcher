package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	api := os.Getenv("TW_API")
	if api == "" {
		api = "http://localhost:8080"
	}
	switch os.Args[1] {
	case "import":
		runImport(api, flagValue("--file"))
	case "export":
		runExport(api, flagValue("--out"))
	case "query":
		if len(os.Args) < 3 || os.Args[2] != "adjacency" {
			usage()
			os.Exit(1)
		}
		runQueryAdjacency(api, flagValue("--device-id"))
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("usage: tw-topology import --file=<fixture.{json|yaml|yml}> | tw-topology export [--out=<path>] | tw-topology query adjacency --device-id=<id>")
}

func flagValue(key string) string {
	for _, arg := range os.Args[2:] {
		if strings.HasPrefix(arg, key+"=") {
			return strings.TrimPrefix(arg, key+"=")
		}
	}
	return ""
}

func decodeFixture(path string) (domain.TopologySnapshot, error) {
	var out domain.TopologySnapshot
	b, err := os.ReadFile(path)
	if err != nil {
		return out, err
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".yaml" || ext == ".yml" {
		return parseSimpleTopologyYAML(b)
	}
	err = json.Unmarshal(b, &out)
	return out, err
}

func parseSimpleTopologyYAML(b []byte) (domain.TopologySnapshot, error) {
	var out domain.TopologySnapshot
	scanner := bufio.NewScanner(bytes.NewReader(b))
	section := ""
	item := map[string]string{}
	flush := func() {
		if len(item) == 0 {
			return
		}
		switch section {
		case "vendors":
			out.Vendors = append(out.Vendors, domain.Vendor{ID: item["id"], Name: item["name"]})
		case "platforms":
			out.Platforms = append(out.Platforms, domain.Platform{ID: item["id"], VendorID: item["vendor_id"], Name: item["name"]})
		case "sites":
			out.Sites = append(out.Sites, domain.Site{ID: item["id"], Name: item["name"]})
		case "devices":
			out.Devices = append(out.Devices, domain.Device{ID: item["id"], Hostname: item["hostname"], VendorID: item["vendor_id"], PlatformID: item["platform_id"], SiteID: item["site_id"]})
		case "interfaces":
			out.Interfaces = append(out.Interfaces, domain.Interface{ID: item["id"], DeviceID: item["device_id"], Name: item["name"]})
		case "links":
			out.Links = append(out.Links, domain.Link{ID: item["id"], AInterfaceID: item["a_interface_id"], ZInterfaceID: item["z_interface_id"]})
		}
		item = map[string]string{}
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, ":") && !strings.HasPrefix(line, "-") {
			flush()
			section = strings.TrimSuffix(line, ":")
			continue
		}
		if strings.HasPrefix(line, "-") {
			flush()
			line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
			if line == "" {
				continue
			}
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		item[strings.TrimSpace(parts[0])] = strings.TrimSpace(strings.Trim(parts[1], "\"'"))
	}
	flush()
	return out, scanner.Err()
}

func runImport(api, path string) {
	if path == "" {
		panic("--file is required")
	}
	snap, err := decodeFixture(path)
	if err != nil {
		panic(err)
	}
	body, _ := json.Marshal(snap)
	resp, err := http.Post(api+"/api/v1/topology/import", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		panic(string(b))
	}
	fmt.Printf("imported %d devices and %d links\n", len(snap.Devices), len(snap.Links))
}

func runExport(api, outPath string) {
	resp, err := http.Get(api + "/api/v1/topology/export")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		panic(string(b))
	}
	b, _ := io.ReadAll(resp.Body)
	if outPath == "" {
		fmt.Println(string(b))
		return
	}
	if err := os.WriteFile(outPath, b, 0o644); err != nil {
		panic(err)
	}
	fmt.Println("wrote", outPath)
}

func runQueryAdjacency(api, deviceID string) {
	if deviceID == "" {
		panic("--device-id is required")
	}
	resp, err := http.Get(api + "/api/v1/topology/query/adjacency?device_id=" + deviceID)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		panic(string(b))
	}
	fmt.Println(string(b))
}
