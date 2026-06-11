package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
)

type AssetReader interface {
	GetAsset(context.Context, string) (assets.Asset, error)
	ListAssets(context.Context) ([]assets.Asset, error)
	ListFactsByAsset(context.Context, string) ([]assets.Fact, error)
	ListRelationships(context.Context) ([]assets.Relationship, error)
}

type DiscoveryReader interface {
	GetDiscoveryRun(context.Context, string) (discovery.DiscoveryRun, error)
	ListDiscoveryRuns(context.Context) ([]discovery.DiscoveryRun, error)
}

type EvidenceReader interface {
	GetEvidence(context.Context, string) (evidence.Evidence, error)
	ListEvidenceByDiscoveryRun(context.Context, string) ([]evidence.Evidence, error)
}

type Service struct {
	assets    AssetReader
	discovery DiscoveryReader
	evidence  EvidenceReader
}

type Options struct {
	Assets    AssetReader
	Discovery DiscoveryReader
	Evidence  EvidenceReader
}

type Request struct {
	Message string `json:"message"`
}

type Response struct {
	Message   string         `json:"message"`
	Intent    string         `json:"intent"`
	ReadOnly  bool           `json:"read_only"`
	Actions   []string       `json:"actions"`
	Artifacts map[string]any `json:"artifacts,omitempty"`
}

func NewService(opts Options) Service {
	return Service{assets: opts.Assets, discovery: opts.Discovery, evidence: opts.Evidence}
}

func (s Service) Reply(ctx context.Context, req Request) (Response, error) {
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return Response{}, fmt.Errorf("message is required")
	}

	normalized := strings.ToLower(message)
	switch {
	case strings.Contains(normalized, "what do we know about"):
		return s.whatWeKnowAbout(ctx, message)
	case strings.Contains(normalized, "show") && strings.Contains(normalized, "neighbor"):
		return s.showNeighbors(ctx, message)
	case strings.Contains(normalized, "why") && strings.Contains(normalized, "believe") && strings.Contains(normalized, "exists"):
		return s.whyAssetExists(ctx, message)
	case strings.Contains(normalized, "what") && strings.Contains(normalized, "unknown"):
		return s.whatIsUnknown(ctx)
	case strings.Contains(normalized, "list") && strings.Contains(normalized, "asset"):
		return s.listKnownAssets(ctx)
	case strings.Contains(normalized, "evidence") && strings.Contains(normalized, "asset"):
		return s.explainAssetEvidence(ctx, message)
	case strings.Contains(normalized, "summarize") && strings.Contains(normalized, "discovery"):
		return s.summarizeDiscoveryRun(ctx, message)
	default:
		return Response{
			Message:  "I can answer deterministic read-only questions right now: list known assets, explain asset evidence, or summarize discovery run.",
			Intent:   "help",
			ReadOnly: true,
			Actions:  []string{"no network actions", "no discovery execution", "no external LLM call"},
		}, nil
	}
}

func (s Service) listKnownAssets(ctx context.Context) (Response, error) {
	if s.assets == nil {
		return Response{}, fmt.Errorf("asset repository is not configured")
	}
	items, err := s.assets.ListAssets(ctx)
	if err != nil {
		return Response{}, err
	}
	sort.Slice(items, func(i, j int) bool { return assetLabel(items[i]) < assetLabel(items[j]) })

	lines := []string{fmt.Sprintf("Known assets: %d", len(items))}
	for _, item := range items {
		vendor := "unknown vendor"
		if item.Vendor != nil && strings.TrimSpace(*item.Vendor) != "" {
			vendor = strings.TrimSpace(*item.Vendor)
		}
		lines = append(lines, fmt.Sprintf("- %s (%s, %s, state=%s, confidence=%.0f%%)", assetLabel(item), item.Type, vendor, item.State, item.Confidence*100))
	}

	return Response{
		Message:   strings.Join(lines, "\n"),
		Intent:    "list_known_assets",
		ReadOnly:  true,
		Actions:   []string{"read assets"},
		Artifacts: map[string]any{"assets": items},
	}, nil
}

func (s Service) explainAssetEvidence(ctx context.Context, message string) (Response, error) {
	if s.assets == nil {
		return Response{}, fmt.Errorf("asset repository is not configured")
	}
	if s.evidence == nil {
		return Response{}, fmt.Errorf("evidence repository is not configured")
	}

	asset, err := s.findAsset(ctx, message)
	if err != nil {
		return Response{}, err
	}
	facts, err := s.assets.ListFactsByAsset(ctx, asset.ID)
	if err != nil {
		return Response{}, err
	}
	relationships, err := s.assets.ListRelationships(ctx)
	if err != nil {
		return Response{}, err
	}

	evidenceIDs := map[string]struct{}{}
	for _, fact := range facts {
		if fact.EvidenceID != nil {
			evidenceIDs[*fact.EvidenceID] = struct{}{}
		}
	}
	for _, relationship := range relationships {
		if relationship.SourceAssetID != asset.ID && relationship.TargetAssetID != asset.ID {
			continue
		}
		if relationship.EvidenceID != nil {
			evidenceIDs[*relationship.EvidenceID] = struct{}{}
		}
	}

	ids := make([]string, 0, len(evidenceIDs))
	for id := range evidenceIDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	items := make([]evidence.Evidence, 0, len(ids))
	for _, id := range ids {
		item, err := s.evidence.GetEvidence(ctx, id)
		if err != nil {
			return Response{}, err
		}
		items = append(items, item)
	}

	lines := []string{
		fmt.Sprintf("Asset %s has %d facts and %d evidence records referenced by facts or relationships.", assetLabel(asset), len(facts), len(items)),
		"Evidence remains read-only and should be checked before trusting derived facts.",
	}
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("- %s via %s against %s at %s (hash %s)", item.CommandOrAPI, item.Method, item.Target, item.CollectedAt.Format("2006-01-02T15:04:05Z07:00"), item.RawOutputHash))
	}

	return Response{
		Message:  strings.Join(lines, "\n"),
		Intent:   "explain_asset_evidence",
		ReadOnly: true,
		Actions:  []string{"read asset", "read facts", "read relationships", "read evidence"},
		Artifacts: map[string]any{
			"asset":    asset,
			"facts":    facts,
			"evidence": items,
		},
	}, nil
}

func (s Service) whatWeKnowAbout(ctx context.Context, message string) (Response, error) {
	asset, ok, err := s.findAssetStrict(ctx, message)
	if err != nil {
		return Response{}, err
	}
	if !ok {
		return unknownResponse("what_we_know_about", "I do not know which asset you mean. No stored asset matched the question."), nil
	}

	facts, err := s.assets.ListFactsByAsset(ctx, asset.ID)
	if err != nil {
		return Response{}, err
	}
	relationships, err := s.assets.ListRelationships(ctx)
	if err != nil {
		return Response{}, err
	}
	related := relationshipsForAsset(relationships, asset.ID)

	lines := []string{
		fmt.Sprintf("Known asset: %s", assetLabel(asset)),
		fmt.Sprintf("- type: %s", valueOrUnknown(asset.Type)),
		fmt.Sprintf("- state: %s", valueOrUnknown(string(asset.State))),
		fmt.Sprintf("- confidence: %.0f%% (%s)", asset.Confidence*100, valueOrUnknown(asset.ConfidenceReason)),
	}
	if len(facts) == 0 {
		lines = append(lines, "- facts: unknown")
	} else {
		lines = append(lines, "Facts:")
		for _, fact := range facts {
			lines = append(lines, fmt.Sprintf("- %s=%s confidence=%.0f%% evidence=%s", fact.Name, factValue(fact.Value), fact.Confidence*100, evidenceRef(fact.EvidenceID)))
		}
	}
	if len(related) == 0 {
		lines = append(lines, "Relationships: unknown")
	} else {
		lines = append(lines, "Relationships:")
		for _, relationship := range related {
			lines = append(lines, fmt.Sprintf("- %s: %s -> %s confidence=%.0f%% evidence=%s", relationship.RelationshipType, relationship.SourceAssetID, relationship.TargetAssetID, relationship.Confidence*100, evidenceRef(relationship.EvidenceID)))
		}
	}

	return Response{
		Message:  strings.Join(lines, "\n"),
		Intent:   "what_we_know_about",
		ReadOnly: true,
		Actions:  []string{"read asset", "read facts", "read relationships"},
		Artifacts: map[string]any{
			"asset":         asset,
			"facts":         facts,
			"relationships": related,
		},
	}, nil
}

func (s Service) showNeighbors(ctx context.Context, message string) (Response, error) {
	asset, ok, err := s.findAssetStrict(ctx, message)
	if err != nil {
		return Response{}, err
	}
	if !ok {
		return unknownResponse("show_neighbors", "I do not know which asset you mean. No stored asset matched the question."), nil
	}

	relationships, err := s.assets.ListRelationships(ctx)
	if err != nil {
		return Response{}, err
	}
	related := relationshipsForAsset(relationships, asset.ID)
	if len(related) == 0 {
		return Response{
			Message:  fmt.Sprintf("Neighbors for %s are unknown. No stored relationships touch this asset.", assetLabel(asset)),
			Intent:   "show_neighbors",
			ReadOnly: true,
			Actions:  []string{"read relationships"},
			Artifacts: map[string]any{
				"asset":         asset,
				"relationships": []assets.Relationship{},
			},
		}, nil
	}

	lines := []string{fmt.Sprintf("Neighbors for %s:", assetLabel(asset))}
	neighbors := make([]assets.Asset, 0, len(related))
	for _, relationship := range related {
		neighborID := otherAssetID(relationship, asset.ID)
		neighbor, err := s.assets.GetAsset(ctx, neighborID)
		if err != nil {
			lines = append(lines, fmt.Sprintf("- %s via %s (asset details unavailable, evidence=%s)", neighborID, relationship.RelationshipType, evidenceRef(relationship.EvidenceID)))
			continue
		}
		neighbors = append(neighbors, neighbor)
		lines = append(lines, fmt.Sprintf("- %s via %s confidence=%.0f%% evidence=%s", assetLabel(neighbor), relationship.RelationshipType, relationship.Confidence*100, evidenceRef(relationship.EvidenceID)))
	}

	return Response{
		Message:  strings.Join(lines, "\n"),
		Intent:   "show_neighbors",
		ReadOnly: true,
		Actions:  []string{"read asset", "read relationships"},
		Artifacts: map[string]any{
			"asset":         asset,
			"neighbors":     neighbors,
			"relationships": related,
		},
	}, nil
}

func (s Service) whyAssetExists(ctx context.Context, message string) (Response, error) {
	asset, ok, err := s.findAssetStrict(ctx, message)
	if err != nil {
		return Response{}, err
	}
	if !ok {
		return unknownResponse("why_asset_exists", "I do not know that this asset exists. No stored asset matched the question."), nil
	}

	facts, err := s.assets.ListFactsByAsset(ctx, asset.ID)
	if err != nil {
		return Response{}, err
	}
	relationships, err := s.assets.ListRelationships(ctx)
	if err != nil {
		return Response{}, err
	}
	related := relationshipsForAsset(relationships, asset.ID)
	evidenceIDs := evidenceIDsFor(facts, related)

	lines := []string{
		fmt.Sprintf("We believe %s exists because it is stored as an asset with state=%s and confidence=%.0f%%.", assetLabel(asset), asset.State, asset.Confidence*100),
		fmt.Sprintf("Confidence reason: %s.", valueOrUnknown(asset.ConfidenceReason)),
	}
	if len(evidenceIDs) == 0 {
		lines = append(lines, "Evidence references: unknown. No fact or relationship evidence_id is currently linked to this asset.")
	} else {
		lines = append(lines, "Evidence references:")
		for _, id := range evidenceIDs {
			lines = append(lines, "- "+id)
		}
	}

	return Response{
		Message:  strings.Join(lines, "\n"),
		Intent:   "why_asset_exists",
		ReadOnly: true,
		Actions:  []string{"read asset", "read facts", "read relationships"},
		Artifacts: map[string]any{
			"asset":         asset,
			"facts":         facts,
			"relationships": related,
			"evidence_refs": evidenceIDs,
		},
	}, nil
}

func (s Service) whatIsUnknown(ctx context.Context) (Response, error) {
	if s.assets == nil {
		return Response{}, fmt.Errorf("asset repository is not configured")
	}
	items, err := s.assets.ListAssets(ctx)
	if err != nil {
		return Response{}, err
	}
	relationships, err := s.assets.ListRelationships(ctx)
	if err != nil {
		return Response{}, err
	}

	lines := []string{"Unknowns grounded in stored records:"}
	count := 0
	for _, item := range items {
		reasons := unknownAssetReasons(item)
		if len(reasons) == 0 {
			continue
		}
		count++
		lines = append(lines, fmt.Sprintf("- %s: %s", assetLabel(item), strings.Join(reasons, ", ")))
	}
	for _, relationship := range relationships {
		if relationship.State != assets.StateUnknown && relationship.State != assets.StateConflicting && relationship.EvidenceID != nil {
			continue
		}
		count++
		lines = append(lines, fmt.Sprintf("- relationship %s %s->%s: state=%s evidence=%s", relationship.RelationshipType, relationship.SourceAssetID, relationship.TargetAssetID, relationship.State, evidenceRef(relationship.EvidenceID)))
	}
	if count == 0 {
		lines = append(lines, "- No unknown or conflicting records are currently stored. That does not prove the network is complete.")
	}

	return Response{
		Message:  strings.Join(lines, "\n"),
		Intent:   "what_is_unknown",
		ReadOnly: true,
		Actions:  []string{"read assets", "read relationships"},
		Artifacts: map[string]any{
			"assets":        items,
			"relationships": relationships,
		},
	}, nil
}

func (s Service) summarizeDiscoveryRun(ctx context.Context, message string) (Response, error) {
	if s.discovery == nil {
		return Response{}, fmt.Errorf("discovery run repository is not configured")
	}
	if s.evidence == nil {
		return Response{}, fmt.Errorf("evidence repository is not configured")
	}

	run, err := s.findDiscoveryRun(ctx, message)
	if err != nil {
		return Response{}, err
	}
	evidenceItems, err := s.evidence.ListEvidenceByDiscoveryRun(ctx, run.ID)
	if err != nil {
		return Response{}, err
	}
	commands := make([]string, 0, len(evidenceItems))
	for _, item := range evidenceItems {
		commands = append(commands, item.CommandOrAPI)
	}
	sort.Strings(commands)

	seed := "{}"
	if len(run.SeedInput) > 0 && json.Valid(run.SeedInput) {
		seed = string(run.SeedInput)
	}
	messageText := fmt.Sprintf("Discovery run %s is %s with %d evidence records. Seed input: %s", run.ID, run.Status, len(evidenceItems), seed)
	if run.ErrorMessage != nil {
		messageText += fmt.Sprintf(" Error: %s", *run.ErrorMessage)
	}
	if len(commands) > 0 {
		messageText += "\nCommands/APIs observed:\n- " + strings.Join(commands, "\n- ")
	}

	return Response{
		Message:  messageText,
		Intent:   "summarize_discovery_run",
		ReadOnly: true,
		Actions:  []string{"read discovery run", "read discovery run evidence"},
		Artifacts: map[string]any{
			"discovery_run": run,
			"evidence":      evidenceItems,
		},
	}, nil
}

func (s Service) findAsset(ctx context.Context, message string) (assets.Asset, error) {
	items, err := s.assets.ListAssets(ctx)
	if err != nil {
		return assets.Asset{}, err
	}
	if len(items) == 0 {
		return assets.Asset{}, fmt.Errorf("no assets are available")
	}
	lower := strings.ToLower(message)
	for _, item := range items {
		if strings.Contains(lower, strings.ToLower(item.ID)) || strings.Contains(lower, strings.ToLower(item.IdentityKey)) {
			return item, nil
		}
	}
	sort.Slice(items, func(i, j int) bool { return assetLabel(items[i]) < assetLabel(items[j]) })
	return items[0], nil
}

func (s Service) findAssetStrict(ctx context.Context, message string) (assets.Asset, bool, error) {
	items, err := s.assets.ListAssets(ctx)
	if err != nil {
		return assets.Asset{}, false, err
	}
	if len(items) == 0 {
		return assets.Asset{}, false, nil
	}

	lower := strings.ToLower(message)
	for _, item := range items {
		if strings.Contains(lower, strings.ToLower(item.ID)) || strings.Contains(lower, strings.ToLower(item.IdentityKey)) {
			return item, true, nil
		}
		facts, err := s.assets.ListFactsByAsset(ctx, item.ID)
		if err != nil {
			return assets.Asset{}, false, err
		}
		for _, fact := range facts {
			if fact.Name == "hostname" && strings.Contains(lower, strings.ToLower(factValue(fact.Value))) {
				return item, true, nil
			}
		}
	}
	return assets.Asset{}, false, nil
}

func (s Service) findDiscoveryRun(ctx context.Context, message string) (discovery.DiscoveryRun, error) {
	runs, err := s.discovery.ListDiscoveryRuns(ctx)
	if err != nil {
		return discovery.DiscoveryRun{}, err
	}
	if len(runs) == 0 {
		return discovery.DiscoveryRun{}, fmt.Errorf("no discovery runs are available")
	}
	lower := strings.ToLower(message)
	for _, run := range runs {
		if strings.Contains(lower, strings.ToLower(run.ID)) {
			return run, nil
		}
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].CreatedAt.After(runs[j].CreatedAt) })
	return runs[0], nil
}

func unknownResponse(intent, message string) Response {
	return Response{
		Message:  message,
		Intent:   intent,
		ReadOnly: true,
		Actions:  []string{"read stored records"},
		Artifacts: map[string]any{
			"unknown": true,
		},
	}
}

func relationshipsForAsset(relationships []assets.Relationship, assetID string) []assets.Relationship {
	result := make([]assets.Relationship, 0)
	for _, relationship := range relationships {
		if relationship.SourceAssetID == assetID || relationship.TargetAssetID == assetID {
			result = append(result, relationship)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func otherAssetID(relationship assets.Relationship, assetID string) string {
	if relationship.SourceAssetID == assetID {
		return relationship.TargetAssetID
	}
	return relationship.SourceAssetID
}

func evidenceIDsFor(facts []assets.Fact, relationships []assets.Relationship) []string {
	seen := map[string]struct{}{}
	for _, fact := range facts {
		if fact.EvidenceID != nil {
			seen[*fact.EvidenceID] = struct{}{}
		}
	}
	for _, relationship := range relationships {
		if relationship.EvidenceID != nil {
			seen[*relationship.EvidenceID] = struct{}{}
		}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func evidenceRef(id *string) string {
	if id == nil || strings.TrimSpace(*id) == "" {
		return "unknown"
	}
	return *id
}

func factValue(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "unknown"
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return string(raw)
		}
		return string(encoded)
	}
}

func unknownAssetReasons(item assets.Asset) []string {
	var reasons []string
	if item.State == assets.StateUnknown || item.State == assets.StateConflicting {
		reasons = append(reasons, "state="+string(item.State))
	}
	if item.Confidence == 0 {
		reasons = append(reasons, "confidence unknown")
	}
	if item.Vendor == nil {
		reasons = append(reasons, "vendor unknown")
	}
	if item.Model == nil {
		reasons = append(reasons, "model unknown")
	}
	if item.Serial == nil {
		reasons = append(reasons, "serial unknown")
	}
	return reasons
}

func valueOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}

func assetLabel(item assets.Asset) string {
	if strings.TrimSpace(item.IdentityKey) != "" {
		return item.IdentityKey
	}
	return item.ID
}
