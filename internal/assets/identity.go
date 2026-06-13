package assets

import (
	"encoding/json"
	"strings"
)

type IdentityStrength string

const (
	IdentityStrengthStrong      IdentityStrength = "strong"
	IdentityStrengthProvisional IdentityStrength = "provisional"
	IdentityStrengthWeak        IdentityStrength = "weak"
)

type IdentityCandidate struct {
	AssetType   string           `json:"asset_type"`
	IdentityKey string           `json:"identity_key"`
	Strength    IdentityStrength `json:"strength"`
	Reason      string           `json:"reason"`
}

func IdentityCandidateForAsset(assetType string, vendor string, serial string, systemMAC string, hostname string, ipAddress string) IdentityCandidate {
	assetType = normalizeToken(assetType)
	if assetType == "" {
		assetType = "unknown"
	}
	vendor = normalizeToken(vendor)
	serial = normalizeToken(serial)
	systemMAC = normalizeToken(systemMAC)
	hostname = normalizeToken(hostname)
	ipAddress = strings.ToLower(strings.TrimSpace(ipAddress))

	switch {
	case vendor != "" && serial != "":
		return IdentityCandidate{
			AssetType:   assetType,
			IdentityKey: strings.Join([]string{assetType, "vendor_serial", vendor, serial}, ":"),
			Strength:    IdentityStrengthStrong,
			Reason:      "vendor plus serial is a durable hardware identity",
		}
	case systemMAC != "":
		return IdentityCandidate{
			AssetType:   assetType,
			IdentityKey: strings.Join([]string{assetType, "system_mac", systemMAC}, ":"),
			Strength:    IdentityStrengthStrong,
			Reason:      "system MAC is a durable hardware identity",
		}
	case serial != "":
		return IdentityCandidate{
			AssetType:   assetType,
			IdentityKey: strings.Join([]string{assetType, "serial", serial}, ":"),
			Strength:    IdentityStrengthStrong,
			Reason:      "serial is a durable hardware identity",
		}
	case hostname != "":
		return IdentityCandidate{
			AssetType:   assetType,
			IdentityKey: MakeIdentityKey(assetType, "hostname", hostname),
			Strength:    IdentityStrengthProvisional,
			Reason:      "hostname is not globally unique and may change",
		}
	case ipAddress != "":
		return IdentityCandidate{
			AssetType:   assetType,
			IdentityKey: MakeIdentityKey(assetType, "ip", ipAddress),
			Strength:    IdentityStrengthProvisional,
			Reason:      "IP address is not a durable asset identity",
		}
	default:
		return IdentityCandidate{
			AssetType:   assetType,
			IdentityKey: MakeIdentityKey(assetType, "unknown", "unknown"),
			Strength:    IdentityStrengthWeak,
			Reason:      "no durable identity evidence is available",
		}
	}
}

func IdentityCandidateFromKey(assetType string, identityKey string) IdentityCandidate {
	identityKey = NormalizeIdentityKey(identityKey)
	parts := strings.Split(identityKey, ":")
	if len(parts) > 0 && parts[0] != "" {
		assetType = parts[0]
	}
	source := ""
	if len(parts) > 1 {
		source = parts[1]
	}

	switch source {
	case "vendor_serial", "system_mac", "serial", "asset_tag", "external_id":
		return IdentityCandidate{
			AssetType:   normalizeToken(assetType),
			IdentityKey: identityKey,
			Strength:    IdentityStrengthStrong,
			Reason:      "identity key uses a durable identifier",
		}
	case "hostname", "ip", "name":
		return IdentityCandidate{
			AssetType:   normalizeToken(assetType),
			IdentityKey: identityKey,
			Strength:    IdentityStrengthProvisional,
			Reason:      source + " is not globally unique and may change",
		}
	default:
		return IdentityCandidate{
			AssetType:   normalizeToken(assetType),
			IdentityKey: identityKey,
			Strength:    IdentityStrengthWeak,
			Reason:      "identity key strength is unknown",
		}
	}
}

func IdentityCandidateForStoredAsset(item Asset) IdentityCandidate {
	vendor := stringValue(item.Vendor)
	serial := stringValue(item.Serial)
	systemMAC := stringValue(item.SystemMAC)
	candidate := IdentityCandidateForAsset(item.Type, vendor, serial, systemMAC, "", "")
	if candidate.Strength == IdentityStrengthStrong && candidate.IdentityKey == item.IdentityKey {
		return candidate
	}

	fromKey := IdentityCandidateFromKey(item.Type, item.IdentityKey)
	if fromKey.Strength == IdentityStrengthStrong {
		return fromKey
	}
	return fromKey
}

func AnnotateIdentityMetadata(metadata json.RawMessage, candidate IdentityCandidate) json.RawMessage {
	if strings.TrimSpace(string(metadata)) == "" {
		metadata = json.RawMessage(`{}`)
	}

	var payload map[string]any
	if err := json.Unmarshal(metadata, &payload); err != nil || payload == nil {
		payload = map[string]any{}
	}
	payload["identity_strength"] = candidate.Strength
	payload["identity_reason"] = candidate.Reason
	payload["identity_provisional"] = candidate.Strength != IdentityStrengthStrong

	encoded, err := json.Marshal(payload)
	if err != nil {
		return metadata
	}
	return encoded
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
