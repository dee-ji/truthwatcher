package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/evidence"
)

const (
	PlatformJunos = "junos"
	PlatformIOSXR = "iosxr"

	CommandShowVersion         = "show version"
	CommandShowChassisHardware = "show chassis hardware"
	CommandShowInventory       = "show inventory"
	CommandShowLLDPNeighbors   = "show lldp neighbors"
)

// BuiltInRegistry returns the first fixture-driven parser set.
func BuiltInRegistry() Registry {
	return NewRegistry(
		commandParser{name: "junos_show_version", platform: PlatformJunos, command: CommandShowVersion, parse: parseJunosShowVersion},
		commandParser{name: "junos_show_chassis_hardware", platform: PlatformJunos, command: CommandShowChassisHardware, parse: parseJunosShowChassisHardware},
		commandParser{name: "junos_show_lldp_neighbors", platform: PlatformJunos, command: CommandShowLLDPNeighbors, parse: parseJunosShowLLDPNeighbors},
		commandParser{name: "iosxr_show_version", platform: PlatformIOSXR, command: CommandShowVersion, parse: parseIOSXRShowVersion},
		commandParser{name: "iosxr_show_inventory", platform: PlatformIOSXR, command: CommandShowInventory, parse: parseIOSXRShowInventory},
		commandParser{name: "iosxr_show_lldp_neighbors", platform: PlatformIOSXR, command: CommandShowLLDPNeighbors, parse: parseIOSXRShowLLDPNeighbors},
	)
}

type commandParser struct {
	name     string
	platform string
	command  string
	parse    func(context.Context, evidence.Evidence, string) (Result, error)
}

func (p commandParser) Name() string {
	return p.name
}

func (p commandParser) Supports(platform string, command string) bool {
	return sameToken(platform, p.platform) && sameToken(command, p.command)
}

func (p commandParser) Parse(ctx context.Context, item evidence.Evidence) (Result, error) {
	if p.parse == nil {
		return Result{}, fmt.Errorf("parser %q has no parse function", p.name)
	}
	return p.parse(ctx, item, p.name)
}

func parseJunosShowVersion(ctx context.Context, item evidence.Evidence, parserName string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	fields := parseColonFields(item.RawOutput)
	hostname := fields["hostname"]
	model := fields["model"]
	version := fields["junos"]

	result := baseResult(parserName, item.ID)
	if hostname == "" {
		result.Warnings = append(result.Warnings, "hostname not found in show version")
		return result, nil
	}

	identityKey := assets.MakeIdentityKey("device", "hostname", hostname)
	result.DeviceIdentities = append(result.DeviceIdentities, DeviceIdentity{
		AssetRef: AssetRef{
			AssetType:   "device",
			IdentityKey: identityKey,
			Confidence:  0.55,
			EvidenceID:  item.ID,
		},
		Hostname: hostname,
		Vendor:   "juniper",
		Model:    model,
		Metadata: mustJSON(map[string]string{
			"platform": "junos",
			"command":  item.CommandOrAPI,
		}),
	})
	result.Facts = append(result.Facts,
		stringFact(identityKey, "hostname", hostname, parserName, 0.8, item.ID),
		stringFact(identityKey, "platform", "junos", parserName, 0.8, item.ID),
	)
	if model != "" {
		result.Facts = append(result.Facts, stringFact(identityKey, "model", model, parserName, 0.7, item.ID))
	}
	if version != "" {
		result.Facts = append(result.Facts, stringFact(identityKey, "software_version", version, parserName, 0.7, item.ID))
	}
	return result, nil
}

func parseIOSXRShowVersion(ctx context.Context, item evidence.Evidence, parserName string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	result := baseResult(parserName, item.ID)
	hostname := parseColonFields(item.RawOutput)["hostname"]
	model := firstContainingLine(item.RawOutput, "Cisco ASR")
	version := parseAfter(item.RawOutput, "Cisco IOS XR Software, Version ")
	if hostname == "" {
		result.Warnings = append(result.Warnings, "hostname not found in show version")
		return result, nil
	}

	identityKey := assets.MakeIdentityKey("device", "hostname", hostname)
	result.DeviceIdentities = append(result.DeviceIdentities, DeviceIdentity{
		AssetRef: AssetRef{
			AssetType:   "device",
			IdentityKey: identityKey,
			Confidence:  0.55,
			EvidenceID:  item.ID,
		},
		Hostname: hostname,
		Vendor:   "cisco",
		Model:    model,
		Metadata: mustJSON(map[string]string{
			"platform": "iosxr",
			"command":  item.CommandOrAPI,
		}),
	})
	result.Facts = append(result.Facts,
		stringFact(identityKey, "hostname", hostname, parserName, 0.8, item.ID),
		stringFact(identityKey, "platform", "iosxr", parserName, 0.8, item.ID),
	)
	if model != "" {
		result.Facts = append(result.Facts, stringFact(identityKey, "model", model, parserName, 0.65, item.ID))
	}
	if version != "" {
		result.Facts = append(result.Facts, stringFact(identityKey, "software_version", version, parserName, 0.75, item.ID))
	}
	return result, nil
}

func parseJunosShowChassisHardware(ctx context.Context, item evidence.Evidence, parserName string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	result := baseResult(parserName, item.ID)
	for _, line := range strings.Split(item.RawOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Hardware inventory") || strings.HasPrefix(line, "Item ") || strings.HasPrefix(line, "PIC ") {
			continue
		}

		if strings.HasPrefix(line, "Chassis") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				result.InventoryComponents = append(result.InventoryComponents, component("chassis", "chassis", "juniper", parts[2], parts[1], item.ID, 0.85))
			}
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 7 {
			result.Warnings = append(result.Warnings, "skipped inventory line: "+line)
			continue
		}
		name := strings.Join(parts[0:2], " ")
		serial := parts[5]
		model := strings.Join(parts[6:], " ")
		componentType := strings.ToLower(parts[0])
		result.InventoryComponents = append(result.InventoryComponents, component(componentType, name, "juniper", model, serial, item.ID, 0.75))
	}
	if len(result.InventoryComponents) == 0 {
		result.Warnings = append(result.Warnings, "no inventory components parsed")
	}
	return result, nil
}

func parseIOSXRShowInventory(ctx context.Context, item evidence.Evidence, parserName string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	result := baseResult(parserName, item.ID)
	blocks := strings.Split(item.RawOutput, "\n\n")
	for _, block := range blocks {
		name := regexpFind(block, `NAME:\s*"([^"]+)"`)
		model := regexpFind(block, `PID:\s*([^,\s]+)`)
		serial := regexpFind(block, `SN:\s*([A-Za-z0-9_-]+)`)
		if name == "" || serial == "" {
			result.Warnings = append(result.Warnings, "skipped inventory block")
			continue
		}
		componentType := "component"
		if strings.Contains(strings.ToLower(name), "rack") {
			componentType = "chassis"
		}
		result.InventoryComponents = append(result.InventoryComponents, component(componentType, name, "cisco", model, serial, item.ID, 0.8))
	}
	if len(result.InventoryComponents) == 0 {
		result.Warnings = append(result.Warnings, "no inventory components parsed")
	}
	return result, nil
}

func parseJunosShowLLDPNeighbors(ctx context.Context, item evidence.Evidence, parserName string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	result := baseResult(parserName, item.ID)
	for _, line := range strings.Split(item.RawOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Local Interface") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 5 {
			result.Warnings = append(result.Warnings, "skipped lldp line: "+line)
			continue
		}
		result.Neighbors = append(result.Neighbors, Neighbor{
			LocalInterfaceName:  parts[0],
			RemoteIdentityKey:   assets.MakeIdentityKey("device", "hostname", parts[4]),
			RemoteSystemName:    parts[4],
			RemoteInterfaceName: parts[3],
			Protocol:            "lldp",
			Confidence:          0.75,
			Metadata:            mustJSON(map[string]string{"remote_chassis_id": parts[2]}),
		})
		result.Relationships = append(result.Relationships, ParsedRelationship{
			SourceIdentityKey: assets.MakeIdentityKey("interface", "name", parts[0]),
			TargetIdentityKey: assets.MakeIdentityKey("device", "hostname", parts[4]),
			RelationshipType:  "lldp_neighbor_of",
			Confidence:        0.75,
			EvidenceID:        item.ID,
			Metadata:          mustJSON(map[string]string{"remote_interface": parts[3]}),
		})
	}
	if len(result.Neighbors) == 0 {
		result.Warnings = append(result.Warnings, "no lldp neighbors parsed")
	}
	return result, nil
}

func parseIOSXRShowLLDPNeighbors(ctx context.Context, item evidence.Evidence, parserName string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	result := baseResult(parserName, item.ID)
	for _, line := range strings.Split(item.RawOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Capability") || strings.HasPrefix(line, "(") || strings.HasPrefix(line, "Device ID") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 5 {
			result.Warnings = append(result.Warnings, "skipped lldp line: "+line)
			continue
		}
		result.Neighbors = append(result.Neighbors, Neighbor{
			LocalInterfaceName:  parts[1],
			RemoteIdentityKey:   assets.MakeIdentityKey("device", "hostname", parts[0]),
			RemoteSystemName:    parts[0],
			RemoteInterfaceName: parts[4],
			Protocol:            "lldp",
			Confidence:          0.75,
			Metadata:            mustJSON(map[string]string{"capability": parts[3]}),
		})
		result.Relationships = append(result.Relationships, ParsedRelationship{
			SourceIdentityKey: assets.MakeIdentityKey("interface", "name", parts[1]),
			TargetIdentityKey: assets.MakeIdentityKey("device", "hostname", parts[0]),
			RelationshipType:  "lldp_neighbor_of",
			Confidence:        0.75,
			EvidenceID:        item.ID,
			Metadata:          mustJSON(map[string]string{"remote_interface": parts[4]}),
		})
	}
	if len(result.Neighbors) == 0 {
		result.Warnings = append(result.Warnings, "no lldp neighbors parsed")
	}
	return result, nil
}

func baseResult(parserName string, evidenceID string) Result {
	return Result{ParserName: parserName, EvidenceID: evidenceID}
}

func component(componentType string, name string, vendor string, model string, serial string, evidenceID string, confidence float64) InventoryComponent {
	assetType := strings.ToLower(strings.TrimSpace(componentType))
	if assetType == "" {
		assetType = "component"
	}
	identityKey := assets.MakeIdentityKey(assetType, "serial", serial)
	return InventoryComponent{
		AssetRef: AssetRef{
			AssetType:   assetType,
			IdentityKey: identityKey,
			Confidence:  confidence,
			EvidenceID:  evidenceID,
		},
		ComponentType: assetType,
		Name:          strings.TrimSpace(name),
		Vendor:        vendor,
		Model:         strings.TrimSpace(model),
		Serial:        strings.TrimSpace(serial),
		Metadata:      mustJSON(map[string]string{"source": "inventory"}),
	}
}

func stringFact(assetIdentityKey string, name string, value string, source string, confidence float64, evidenceID string) ParsedFact {
	encoded, _ := json.Marshal(value)
	return ParsedFact{
		AssetIdentityKey: assetIdentityKey,
		Name:             name,
		Value:            encoded,
		Source:           source,
		Confidence:       confidence,
		EvidenceID:       evidenceID,
	}
}

func parseColonFields(raw string) map[string]string {
	result := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		result[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}
	return result
}

func parseAfter(raw string, prefix string) string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func firstContainingLine(raw string, needle string) string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, needle) {
			return line
		}
	}
	return ""
}

func regexpFind(raw string, pattern string) string {
	matches := regexp.MustCompile(pattern).FindStringSubmatch(raw)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

func sameToken(a string, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func mustJSON(value any) json.RawMessage {
	encoded, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return encoded
}
