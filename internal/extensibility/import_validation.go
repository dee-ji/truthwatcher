package extensibility

import (
	"encoding/json"
	"fmt"
	"strings"

	"truthwatcher/internal/assets"
)

// ValidateImportResult checks local import candidates before the kernel decides
// whether to persist them. It does not upgrade imported records into observed
// truth; file data remains candidate context until explicitly reconciled.
func ValidateImportResult(result ImportResult) ([]string, error) {
	var warnings []string
	warnings = append(warnings, result.Warnings...)

	for index, item := range result.Candidates.Assets {
		if strings.TrimSpace(item.Type) == "" {
			return warnings, fmt.Errorf("asset candidate %d: asset type is required", index)
		}
		if strings.TrimSpace(item.IdentityKey) == "" {
			return warnings, fmt.Errorf("asset candidate %d: identity_key is required", index)
		}
		if item.Confidence < 0 || item.Confidence > 1 {
			return warnings, fmt.Errorf("asset candidate %d: confidence must be between 0 and 1", index)
		}
		if item.State != "" && !item.State.Valid() {
			return warnings, fmt.Errorf("asset candidate %d: invalid state %q", index, item.State)
		}
		if !validOptionalJSON(item.Metadata) {
			return warnings, fmt.Errorf("asset candidate %d: metadata must be valid JSON", index)
		}
		if item.State == assets.StateObserved {
			warnings = append(warnings, "imported observed assets are candidates only; this command does not persist them as observed proof")
		}
	}

	for index, item := range result.Candidates.Facts {
		if strings.TrimSpace(item.AssetID) == "" {
			return warnings, fmt.Errorf("fact candidate %d: asset_id is required", index)
		}
		if strings.TrimSpace(item.Name) == "" {
			return warnings, fmt.Errorf("fact candidate %d: fact name is required", index)
		}
		if strings.TrimSpace(item.Source) == "" {
			return warnings, fmt.Errorf("fact candidate %d: source is required", index)
		}
		if item.Confidence < 0 || item.Confidence > 1 {
			return warnings, fmt.Errorf("fact candidate %d: confidence must be between 0 and 1", index)
		}
		if item.State != "" && !item.State.Valid() {
			return warnings, fmt.Errorf("fact candidate %d: invalid state %q", index, item.State)
		}
		if !json.Valid(item.Value) {
			return warnings, fmt.Errorf("fact candidate %d: value must be valid JSON", index)
		}
		if item.State == assets.StateObserved {
			warnings = append(warnings, "imported observed facts are candidates only; this command does not persist them as observed proof")
		}
	}

	for index, item := range result.Candidates.Relationships {
		if strings.TrimSpace(item.SourceAssetID) == "" {
			return warnings, fmt.Errorf("relationship candidate %d: source_asset_id is required", index)
		}
		if strings.TrimSpace(item.TargetAssetID) == "" {
			return warnings, fmt.Errorf("relationship candidate %d: target_asset_id is required", index)
		}
		if strings.TrimSpace(item.RelationshipType) == "" {
			return warnings, fmt.Errorf("relationship candidate %d: relationship_type is required", index)
		}
		if item.Confidence < 0 || item.Confidence > 1 {
			return warnings, fmt.Errorf("relationship candidate %d: confidence must be between 0 and 1", index)
		}
		if item.State != "" && !item.State.Valid() {
			return warnings, fmt.Errorf("relationship candidate %d: invalid state %q", index, item.State)
		}
		if !validOptionalJSON(item.Metadata) {
			return warnings, fmt.Errorf("relationship candidate %d: metadata must be valid JSON", index)
		}
		if item.State == assets.StateObserved {
			warnings = append(warnings, "imported observed relationships are candidates only; this command does not persist them as observed proof")
		}
	}

	for index, item := range result.Evidence {
		if strings.TrimSpace(item.Target) == "" {
			return warnings, fmt.Errorf("evidence candidate %d: target is required", index)
		}
		if strings.TrimSpace(item.Method) == "" {
			return warnings, fmt.Errorf("evidence candidate %d: method is required", index)
		}
		if strings.TrimSpace(item.CommandOrAPI) == "" {
			return warnings, fmt.Errorf("evidence candidate %d: command_or_api is required", index)
		}
		if !validOptionalJSON(item.Metadata) {
			return warnings, fmt.Errorf("evidence candidate %d: metadata must be valid JSON", index)
		}
	}

	return dedupeWarnings(warnings), nil
}

func validOptionalJSON(value json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(value))
	return trimmed == "" || json.Valid(value)
}

func dedupeWarnings(warnings []string) []string {
	seen := make(map[string]struct{}, len(warnings))
	out := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		warning = strings.TrimSpace(warning)
		if warning == "" {
			continue
		}
		if _, ok := seen[warning]; ok {
			continue
		}
		seen[warning] = struct{}{}
		out = append(out, warning)
	}
	return out
}
