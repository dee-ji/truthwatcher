package parser

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"truthwatcher/internal/evidence"
)

func TestRegistrySelectsSupportingParser(t *testing.T) {
	junos := fakeParser{name: "junos_version", platform: "junos", command: "show version"}
	iosxr := fakeParser{name: "iosxr_version", platform: "iosxr", command: "show version"}

	selected := NewRegistry(junos, iosxr).Select("junos", "show version")
	if selected.Name() != "junos_version" {
		t.Fatalf("selected parser = %q, want junos_version", selected.Name())
	}
}

func TestRegistrySelectionIsCaseInsensitiveViaParser(t *testing.T) {
	selected := NewRegistry(fakeParser{name: "junos_version", platform: "junos", command: "show version"}).
		Select(" JUNOS ", " SHOW VERSION ")

	if selected.Name() != "junos_version" {
		t.Fatalf("selected parser = %q, want junos_version", selected.Name())
	}
}

func TestRegistryFallsBackToNoopParser(t *testing.T) {
	selected := NewRegistry(fakeParser{name: "junos_version", platform: "junos", command: "show version"}).
		Select("iosxr", "show bgp summary")

	if selected.Name() != NoopParserName {
		t.Fatalf("selected parser = %q, want noop", selected.Name())
	}
}

func TestRegistryParseUsesSelectedParser(t *testing.T) {
	item := testEvidence("evidence-1", "show version")
	result, err := NewRegistry(fakeParser{name: "junos_version", platform: "junos", command: "show version"}).
		Parse(context.Background(), "junos", item)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if result.ParserName != "junos_version" {
		t.Fatalf("parser name = %q, want junos_version", result.ParserName)
	}
	if result.EvidenceID != item.ID {
		t.Fatalf("evidence id = %q, want %q", result.EvidenceID, item.ID)
	}
	if got, want := len(result.DeviceIdentities), 1; got != want {
		t.Fatalf("device identity count = %d, want %d", got, want)
	}
	if result.Facts[0].Name != "hostname" {
		t.Fatalf("fact name = %q, want hostname", result.Facts[0].Name)
	}
}

func TestNoopParserReturnsEmptyResultWithWarning(t *testing.T) {
	item := testEvidence("evidence-2", "show unsupported")
	result, err := NoopParser{}.Parse(context.Background(), item)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if result.ParserName != NoopParserName {
		t.Fatalf("parser name = %q, want noop", result.ParserName)
	}
	if result.EvidenceID != item.ID {
		t.Fatalf("evidence id = %q, want %q", result.EvidenceID, item.ID)
	}
	if len(result.DeviceIdentities) != 0 || len(result.Facts) != 0 || len(result.Relationships) != 0 {
		t.Fatal("noop parser returned model candidates")
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("warnings count = %d, want 1", len(result.Warnings))
	}
}

func TestNormalizedOutputShapes(t *testing.T) {
	result := Result{
		DeviceIdentities: []DeviceIdentity{{
			AssetRef: AssetRef{
				AssetType:   "device",
				IdentityKey: "device:serial:abc123",
				Confidence:  0.95,
				EvidenceID:  "evidence-1",
			},
			Hostname: "router1",
			Vendor:   "juniper",
			Model:    "mx480",
			Serial:   "abc123",
			Metadata: json.RawMessage(`{"source":"show version"}`),
		}},
		InventoryComponents: []InventoryComponent{{
			AssetRef: AssetRef{
				AssetType:   "chassis",
				IdentityKey: "chassis:serial:abc123",
				Confidence:  0.9,
			},
			ComponentType: "chassis",
			Name:          "chassis",
		}},
		Interfaces: []Interface{{
			AssetRef: AssetRef{
				AssetType:   "interface",
				IdentityKey: "interface:router1:xe-0/0/0",
				Confidence:  0.8,
			},
			Name:      "xe-0/0/0",
			OperState: "up",
		}},
		Neighbors: []Neighbor{{
			LocalInterfaceName:  "xe-0/0/0",
			RemoteSystemName:    "spine1",
			RemoteInterfaceName: "Ethernet1/1",
			Protocol:            "lldp",
			Confidence:          0.8,
		}},
		BGPPeers: []BGPPeer{{
			PeerAddress: "192.0.2.1",
			RemoteASN:   65001,
			State:       "established",
			Confidence:  0.75,
		}},
		Facts: []ParsedFact{{
			AssetIdentityKey: "device:serial:abc123",
			Name:             "hostname",
			Value:            json.RawMessage(`"router1"`),
			Source:           "parser",
			Confidence:       0.95,
		}},
		Relationships: []ParsedRelationship{{
			SourceIdentityKey: "interface:router1:xe-0/0/0",
			TargetIdentityKey: "interface:spine1:ethernet1/1",
			RelationshipType:  "lldp_neighbor_of",
			Confidence:        0.8,
			Metadata:          json.RawMessage(`{}`),
		}},
	}

	if len(result.DeviceIdentities) != 1 ||
		len(result.InventoryComponents) != 1 ||
		len(result.Interfaces) != 1 ||
		len(result.Neighbors) != 1 ||
		len(result.BGPPeers) != 1 ||
		len(result.Facts) != 1 ||
		len(result.Relationships) != 1 {
		t.Fatalf("normalized output shape is incomplete: %+v", result)
	}
}

type fakeParser struct {
	name     string
	platform string
	command  string
}

func (p fakeParser) Name() string {
	return p.name
}

func (p fakeParser) Supports(platform string, command string) bool {
	return strings.EqualFold(strings.TrimSpace(platform), p.platform) &&
		strings.EqualFold(strings.TrimSpace(command), p.command)
}

func (p fakeParser) Parse(ctx context.Context, item evidence.Evidence) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	return Result{
		ParserName: p.name,
		EvidenceID: item.ID,
		DeviceIdentities: []DeviceIdentity{{
			AssetRef: AssetRef{
				AssetType:   "device",
				IdentityKey: "device:hostname:router1",
				Confidence:  0.5,
				EvidenceID:  item.ID,
			},
			Hostname: "router1",
		}},
		Facts: []ParsedFact{{
			AssetIdentityKey: "device:hostname:router1",
			Name:             "hostname",
			Value:            json.RawMessage(`"router1"`),
			Source:           p.name,
			Confidence:       0.5,
			EvidenceID:       item.ID,
		}},
	}, nil
}

func testEvidence(id string, command string) evidence.Evidence {
	return evidence.Evidence{
		ID:             id,
		DiscoveryRunID: "run-1",
		Target:         "fixture://junos-mx",
		Method:         "fake",
		CommandOrAPI:   command,
		RawOutput:      "raw output",
		RawOutputHash:  evidence.HashRawOutput("raw output"),
		CollectedAt:    time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		Metadata:       json.RawMessage(`{}`),
	}
}
