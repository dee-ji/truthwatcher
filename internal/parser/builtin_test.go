package parser

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"truthwatcher/internal/evidence"
)

func TestBuiltInRegistrySelectsFixtureParsers(t *testing.T) {
	tests := []struct {
		platform string
		command  string
		parser   string
	}{
		{PlatformJunos, CommandShowVersion, "junos_show_version"},
		{PlatformJunos, CommandShowChassisHardware, "junos_show_chassis_hardware"},
		{PlatformJunos, CommandShowLLDPNeighbors, "junos_show_lldp_neighbors"},
		{PlatformJunos, CommandShowBGPSummary, "junos_show_bgp_summary"},
		{PlatformIOSXR, CommandShowVersion, "iosxr_show_version"},
		{PlatformIOSXR, CommandShowInventory, "iosxr_show_inventory"},
		{PlatformIOSXR, CommandShowLLDPNeighbors, "iosxr_show_lldp_neighbors"},
		{PlatformIOSXR, CommandShowBGPSummary, "iosxr_show_bgp_summary"},
	}

	registry := BuiltInRegistry()
	for _, tt := range tests {
		t.Run(tt.platform+" "+tt.command, func(t *testing.T) {
			selected := registry.Select(tt.platform, tt.command)
			if selected.Name() != tt.parser {
				t.Fatalf("parser = %q, want %q", selected.Name(), tt.parser)
			}
		})
	}
}

func TestJunosShowVersionParser(t *testing.T) {
	result := parseFixture(t, PlatformJunos, CommandShowVersion, "junos-mx", "show_version.txt")

	if got, want := result.ParserName, "junos_show_version"; got != want {
		t.Fatalf("parser = %q, want %q", got, want)
	}
	if got, want := len(result.DeviceIdentities), 1; got != want {
		t.Fatalf("device identity count = %d, want %d", got, want)
	}
	device := result.DeviceIdentities[0]
	if got, want := device.Hostname, "mx-edge-01"; got != want {
		t.Fatalf("hostname = %q, want %q", got, want)
	}
	if got, want := device.IdentityKey, "device:hostname:mx-edge-01"; got != want {
		t.Fatalf("identity key = %q, want %q", got, want)
	}
	if got, want := device.Vendor, "juniper"; got != want {
		t.Fatalf("vendor = %q, want %q", got, want)
	}
	if got, want := device.Model, "mx480"; got != want {
		t.Fatalf("model = %q, want %q", got, want)
	}
	if !hasFact(result.Facts, "software_version", `"22.4R3-S2.4"`) {
		t.Fatalf("software_version fact missing: %+v", result.Facts)
	}
}

func TestJunosShowChassisHardwareParser(t *testing.T) {
	result := parseFixture(t, PlatformJunos, CommandShowChassisHardware, "junos-mx", "show_chassis_hardware.txt")

	if got, want := len(result.InventoryComponents), 3; got != want {
		t.Fatalf("inventory count = %d, want %d", got, want)
	}
	if got, want := result.InventoryComponents[0].Serial, "JN1234ABCDEF"; got != want {
		t.Fatalf("chassis serial = %q, want %q", got, want)
	}
	if got, want := result.InventoryComponents[0].ComponentType, "chassis"; got != want {
		t.Fatalf("component type = %q, want %q", got, want)
	}
	if got, want := result.InventoryComponents[0].IdentityKey, "chassis:vendor_serial:juniper:jn1234abcdef"; got != want {
		t.Fatalf("identity key = %q, want %q", got, want)
	}
}

func TestJunosShowLLDPNeighborsParser(t *testing.T) {
	result := parseFixture(t, PlatformJunos, CommandShowLLDPNeighbors, "junos-mx", "show_lldp_neighbors.txt")

	if got, want := len(result.Neighbors), 2; got != want {
		t.Fatalf("neighbor count = %d, want %d", got, want)
	}
	if got, want := result.Neighbors[0].RemoteSystemName, "spine-01"; got != want {
		t.Fatalf("remote system = %q, want %q", got, want)
	}
	if got, want := len(result.Relationships), 2; got != want {
		t.Fatalf("relationship count = %d, want %d", got, want)
	}
	if got, want := result.Relationships[0].RelationshipType, "lldp_neighbor_of"; got != want {
		t.Fatalf("relationship type = %q, want %q", got, want)
	}
}

func TestIOSXRShowVersionParser(t *testing.T) {
	result := parseFixture(t, PlatformIOSXR, CommandShowVersion, "iosxr-asr", "show_version.txt")

	if got, want := len(result.DeviceIdentities), 1; got != want {
		t.Fatalf("device identity count = %d, want %d", got, want)
	}
	device := result.DeviceIdentities[0]
	if got, want := device.Hostname, "xr-edge-01"; got != want {
		t.Fatalf("hostname = %q, want %q", got, want)
	}
	if got, want := device.Vendor, "cisco"; got != want {
		t.Fatalf("vendor = %q, want %q", got, want)
	}
	if !hasFact(result.Facts, "software_version", `"7.9.2"`) {
		t.Fatalf("software_version fact missing: %+v", result.Facts)
	}
}

func TestIOSXRShowInventoryParser(t *testing.T) {
	result := parseFixture(t, PlatformIOSXR, CommandShowInventory, "iosxr-asr", "show_inventory.txt")

	if got, want := len(result.InventoryComponents), 2; got != want {
		t.Fatalf("inventory count = %d, want %d", got, want)
	}
	if got, want := result.InventoryComponents[0].Serial, "FOX1234ABCD"; got != want {
		t.Fatalf("chassis serial = %q, want %q", got, want)
	}
}

func TestIOSXRShowLLDPNeighborsParser(t *testing.T) {
	result := parseFixture(t, PlatformIOSXR, CommandShowLLDPNeighbors, "iosxr-asr", "show_lldp_neighbors.txt")

	if got, want := len(result.Neighbors), 2; got != want {
		t.Fatalf("neighbor count = %d, want %d", got, want)
	}
	if got, want := result.Neighbors[0].LocalInterfaceName, "Gi0/0/0/0"; got != want {
		t.Fatalf("local interface = %q, want %q", got, want)
	}
	if got, want := result.Neighbors[0].RemoteSystemName, "spine-01"; got != want {
		t.Fatalf("remote system = %q, want %q", got, want)
	}
}

func TestJunosShowBGPSummaryParser(t *testing.T) {
	result := parseFixture(t, PlatformJunos, CommandShowBGPSummary, "junos-mx", "show_bgp_summary.txt")

	if got, want := len(result.BGPPeers), 1; got != want {
		t.Fatalf("bgp peer count = %d, want %d", got, want)
	}
	peer := result.BGPPeers[0]
	if got, want := peer.PeerAddress, "192.0.2.1"; got != want {
		t.Fatalf("peer address = %q, want %q", got, want)
	}
	if got, want := peer.RemoteASN, uint32(65001); got != want {
		t.Fatalf("remote asn = %d, want %d", got, want)
	}
	if peer.AcceptedPrefixes == nil || *peer.AcceptedPrefixes != 12 {
		t.Fatalf("accepted prefixes = %#v, want 12", peer.AcceptedPrefixes)
	}
	if !hasFact(result.Facts, "bgp_peer_count", `3`) {
		t.Fatalf("bgp_peer_count fact missing: %+v", result.Facts)
	}
}

func TestIOSXRShowBGPSummaryParser(t *testing.T) {
	result := parseFixture(t, PlatformIOSXR, CommandShowBGPSummary, "iosxr-asr", "show_bgp_summary.txt")

	if got, want := len(result.BGPPeers), 1; got != want {
		t.Fatalf("bgp peer count = %d, want %d", got, want)
	}
	peer := result.BGPPeers[0]
	if got, want := peer.DeviceIdentityKey, "routing_context:router_id:198.51.100.10"; got != want {
		t.Fatalf("device identity key = %q, want %q", got, want)
	}
	if got, want := peer.PeerAddress, "192.0.2.2"; got != want {
		t.Fatalf("peer address = %q, want %q", got, want)
	}
	if got, want := peer.RemoteASN, uint32(65002); got != want {
		t.Fatalf("remote asn = %d, want %d", got, want)
	}
	if peer.AcceptedPrefixes == nil || *peer.AcceptedPrefixes != 18 {
		t.Fatalf("accepted prefixes = %#v, want 18", peer.AcceptedPrefixes)
	}
	if !hasFact(result.Facts, "bgp_router_id", `"198.51.100.10"`) {
		t.Fatalf("bgp_router_id fact missing: %+v", result.Facts)
	}
	if !hasFact(result.Facts, "bgp_local_as", `65000`) {
		t.Fatalf("bgp_local_as fact missing: %+v", result.Facts)
	}
}

func TestParserFailureKeepsEvidenceReference(t *testing.T) {
	item := fixtureEvidence("evidence-malformed", PlatformJunos, CommandShowVersion, "this does not include expected fields")

	result, err := BuiltInRegistry().Parse(context.Background(), PlatformJunos, item)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if result.EvidenceID != item.ID {
		t.Fatalf("evidence id = %q, want %q", result.EvidenceID, item.ID)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected parser warning for malformed input")
	}
	if len(result.DeviceIdentities) != 0 {
		t.Fatalf("device identities = %d, want 0", len(result.DeviceIdentities))
	}
}

func parseFixture(t *testing.T, platform string, command string, fixtureDir string, filename string) Result {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "fixtures", fixtureDir, filename))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	item := fixtureEvidence("evidence-"+filename, platform, command, string(raw))
	result, err := BuiltInRegistry().Parse(context.Background(), platform, item)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if result.EvidenceID != item.ID {
		t.Fatalf("evidence id = %q, want %q", result.EvidenceID, item.ID)
	}
	return result
}

func fixtureEvidence(id string, platform string, command string, raw string) evidence.Evidence {
	return evidence.Evidence{
		ID:             id,
		DiscoveryRunID: "run-1",
		Target:         "fixture://" + platform,
		Method:         "fake",
		CommandOrAPI:   command,
		RawOutput:      raw,
		RawOutputHash:  evidence.HashRawOutput(raw),
		CollectedAt:    time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		Metadata:       json.RawMessage(`{}`),
	}
}

func hasFact(facts []ParsedFact, name string, value string) bool {
	for _, fact := range facts {
		if fact.Name == name && string(fact.Value) == value {
			return true
		}
	}
	return false
}
