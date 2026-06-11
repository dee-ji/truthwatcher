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

func assetLabel(item assets.Asset) string {
	if strings.TrimSpace(item.IdentityKey) != "" {
		return item.IdentityKey
	}
	return item.ID
}
