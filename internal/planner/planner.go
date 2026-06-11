package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/policy"
)

type AssetReader interface {
	GetAsset(context.Context, string) (assets.Asset, error)
	ListAssets(context.Context) ([]assets.Asset, error)
	ListFactsByAsset(context.Context, string) ([]assets.Fact, error)
	ListRelationships(context.Context) ([]assets.Relationship, error)
}

type Service struct {
	assets AssetReader
	policy policy.Engine
}

type Options struct {
	Assets AssetReader
	Policy policy.Engine
}

type Request struct {
	SeedInput json.RawMessage `json:"seed_input"`
	Target    string          `json:"target"`
	Method    string          `json:"method"`
	Profile   string          `json:"profile"`
	Tasks     []policy.Task   `json:"tasks"`
}

type Plan struct {
	Steps            []Step         `json:"steps"`
	ApprovalRequired bool           `json:"approval_required"`
	ExecutionAllowed bool           `json:"execution_allowed"`
	Warnings         []string       `json:"warnings,omitempty"`
	SeedInput        map[string]any `json:"seed_input,omitempty"`
}

type Step struct {
	Target           string      `json:"target"`
	Method           string      `json:"method"`
	Profile          string      `json:"profile"`
	Task             policy.Task `json:"task"`
	Reason           string      `json:"reason"`
	ExpectedEvidence string      `json:"expected_evidence"`
	RiskLevel        string      `json:"risk_level"`
}

func NewService(opts Options) Service {
	engine := opts.Policy
	if len(engine.AllowedTasks()) == 0 {
		engine = policy.NewEngine()
	}
	return Service{assets: opts.Assets, policy: engine}
}

func (s Service) CreatePlan(ctx context.Context, req Request) (Plan, error) {
	if s.assets == nil {
		return Plan{}, fmt.Errorf("asset repository is required")
	}
	seed, err := parseSeedInput(req.SeedInput)
	if err != nil {
		return Plan{}, err
	}
	target := firstNonEmpty(req.Target, stringFromMap(seed, "target"))
	if strings.TrimSpace(target) == "" {
		return Plan{}, fmt.Errorf("target is required")
	}
	if !safeTarget(target) {
		return Plan{}, fmt.Errorf("target must be a single explicit host, asset, or fixture reference")
	}
	method := firstNonEmpty(req.Method, stringFromMap(seed, "method"), "ssh")
	if method != "ssh" && method != discovery.FakeMethod {
		return Plan{}, fmt.Errorf("method must be ssh or fake")
	}
	profileName := firstNonEmpty(req.Profile, stringFromMap(seed, "profile"))
	if profileName == "" {
		return Plan{}, fmt.Errorf("profile is required")
	}
	profile, ok := discovery.BuiltInProfile(profileName)
	if !ok {
		return Plan{}, fmt.Errorf("unknown discovery profile")
	}
	if err := profile.Validate(s.policy); err != nil {
		return Plan{}, err
	}

	asset, matched, err := s.findAsset(ctx, target)
	if err != nil {
		return Plan{}, err
	}
	tasks, err := s.recommendedTasks(ctx, asset, matched, req.Tasks)
	if err != nil {
		return Plan{}, err
	}

	steps := make([]Step, 0, len(tasks))
	for _, task := range tasks {
		commands, err := profile.CommandsForTask(task)
		if err != nil {
			return Plan{}, err
		}
		expected := expectedEvidence(task, commands)
		steps = append(steps, Step{
			Target:           target,
			Method:           method,
			Profile:          profile.Name,
			Task:             task,
			Reason:           reasonForTask(task, matched),
			ExpectedEvidence: expected,
			RiskLevel:        "low_read_only",
		})
	}

	warnings := []string{"human approval is required before any collector executes this plan", "planner does not guess credentials or expand target scope"}
	if !matched {
		warnings = append(warnings, "target is not yet represented by a stored asset; plan is limited to the explicit seed target")
	}

	return Plan{
		Steps:            steps,
		ApprovalRequired: true,
		ExecutionAllowed: false,
		Warnings:         warnings,
		SeedInput:        seed,
	}, nil
}

func (s Service) recommendedTasks(ctx context.Context, asset assets.Asset, matched bool, requested []policy.Task) ([]policy.Task, error) {
	if len(requested) > 0 {
		return s.validateTasks(requested)
	}
	if !matched {
		return []policy.Task{policy.TaskIdentifyDevice, policy.TaskGetInventory, policy.TaskGetNeighbors}, nil
	}

	facts, err := s.assets.ListFactsByAsset(ctx, asset.ID)
	if err != nil {
		return nil, err
	}
	relationships, err := s.assets.ListRelationships(ctx)
	if err != nil {
		return nil, err
	}

	seen := map[policy.Task]struct{}{}
	add := func(task policy.Task) { seen[task] = struct{}{} }
	if !hasFact(facts, "hostname") || asset.State == assets.StateUnknown || asset.Confidence == 0 {
		add(policy.TaskIdentifyDevice)
	}
	if asset.Model == nil || asset.Serial == nil {
		add(policy.TaskGetInventory)
	}
	if !touchesAnyRelationship(relationships, asset.ID) {
		add(policy.TaskGetNeighbors)
	}
	if !hasFact(facts, "bgp_summary") {
		add(policy.TaskGetBGPSummary)
	}
	if len(seen) == 0 {
		add(policy.TaskGetInterfaces)
	}

	tasks := make([]policy.Task, 0, len(seen))
	for task := range seen {
		tasks = append(tasks, task)
	}
	return s.validateTasks(tasks)
}

func (s Service) validateTasks(tasks []policy.Task) ([]policy.Task, error) {
	seen := map[policy.Task]struct{}{}
	out := make([]policy.Task, 0, len(tasks))
	for _, task := range tasks {
		if err := s.policy.CheckTask(task); err != nil {
			return nil, err
		}
		if _, ok := seen[task]; ok {
			continue
		}
		seen[task] = struct{}{}
		out = append(out, task)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

func (s Service) findAsset(ctx context.Context, target string) (assets.Asset, bool, error) {
	items, err := s.assets.ListAssets(ctx)
	if err != nil {
		return assets.Asset{}, false, err
	}
	needle := strings.ToLower(strings.TrimSpace(target))
	for _, item := range items {
		if strings.ToLower(item.ID) == needle || strings.ToLower(item.IdentityKey) == needle {
			return item, true, nil
		}
		facts, err := s.assets.ListFactsByAsset(ctx, item.ID)
		if err != nil {
			return assets.Asset{}, false, err
		}
		for _, fact := range facts {
			if fact.Name == "hostname" && strings.EqualFold(factValue(fact.Value), target) {
				return item, true, nil
			}
		}
	}
	return assets.Asset{}, false, nil
}

func parseSeedInput(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" {
		return map[string]any{}, nil
	}
	var seed map[string]any
	if err := json.Unmarshal(raw, &seed); err != nil {
		return nil, fmt.Errorf("seed_input must be a JSON object")
	}
	return seed, nil
}

func safeTarget(target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	blocked := []string{"/", "*", ",", " ", "\t", "\n", ".."}
	for _, item := range blocked {
		if strings.Contains(target, item) && !strings.HasPrefix(target, "fixture://") {
			return false
		}
	}
	if strings.HasPrefix(target, "fixture://") {
		name := strings.TrimPrefix(target, "fixture://")
		return name != "" && !strings.ContainsAny(name, "*, \t\n")
	}
	return true
}

func hasFact(facts []assets.Fact, name string) bool {
	for _, fact := range facts {
		if fact.Name == name {
			return true
		}
	}
	return false
}

func touchesAnyRelationship(relationships []assets.Relationship, assetID string) bool {
	for _, relationship := range relationships {
		if relationship.SourceAssetID == assetID || relationship.TargetAssetID == assetID {
			return true
		}
	}
	return false
}

func expectedEvidence(task policy.Task, commands []discovery.CommandMapping) string {
	labels := make([]string, 0, len(commands))
	for _, command := range commands {
		labels = append(labels, command.Command)
	}
	return fmt.Sprintf("raw output from %s for %s", strings.Join(labels, ", "), task)
}

func reasonForTask(task policy.Task, matched bool) string {
	if !matched {
		return "target is not represented in the graph yet; collect minimal identity and topology evidence for the explicit seed"
	}
	switch task {
	case policy.TaskIdentifyDevice:
		return "stored identity facts are incomplete or uncertain"
	case policy.TaskGetInventory:
		return "stored hardware model or serial data is incomplete"
	case policy.TaskGetNeighbors:
		return "stored graph has no relationships for this asset"
	case policy.TaskGetBGPSummary:
		return "stored graph lacks BGP summary facts for this asset"
	case policy.TaskGetInterfaces:
		return "asset already has core identity/topology records; interface evidence is the next safe enrichment"
	default:
		return "requested read-only task is allowlisted by policy"
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func stringFromMap(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func factValue(raw json.RawMessage) string {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}
