package extensibility

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/policy"
)

var (
	ErrScriptRunnerDisabled = errors.New("script runner is disabled")
	ErrScriptNotAllowed     = errors.New("script is not allowlisted")
)

const defaultScriptTimeout = 10 * time.Second

// ScriptRunner executes a local BYO script as an importer. It is intentionally
// disabled by default and never shells out; the script path must be explicitly
// allowlisted by the local caller.
type ScriptRunner struct {
	enabled        bool
	allowedScripts map[string]struct{}
	timeout        time.Duration
	policy         policy.Engine
}

type ScriptRunnerOptions struct {
	Enabled        bool
	AllowedScripts []string
	Timeout        time.Duration
	Policy         policy.Engine
}

type ScriptInput struct {
	Target        string          `json:"target,omitempty"`
	Method        string          `json:"method,omitempty"`
	Profile       string          `json:"profile,omitempty"`
	Tasks         []policy.Task   `json:"tasks,omitempty"`
	CredentialRef string          `json:"credential_ref,omitempty"`
	Context       json.RawMessage `json:"context,omitempty"`
	DryRun        bool            `json:"dry_run"`
}

type ScriptOutput struct {
	Evidence   []EvidenceCandidate `json:"evidence,omitempty"`
	Candidates ScriptCandidates    `json:"candidates"`
	Warnings   []string            `json:"warnings,omitempty"`
}

type ScriptCandidates struct {
	Assets        []ScriptAssetCandidate        `json:"assets,omitempty"`
	Facts         []ScriptFactCandidate         `json:"facts,omitempty"`
	Relationships []ScriptRelationshipCandidate `json:"relationships,omitempty"`
}

type ScriptAssetCandidate struct {
	Type             string                 `json:"type"`
	IdentityKey      string                 `json:"identity_key"`
	Vendor           *string                `json:"vendor,omitempty"`
	Model            *string                `json:"model,omitempty"`
	Serial           *string                `json:"serial,omitempty"`
	SystemMAC        *string                `json:"system_mac,omitempty"`
	Confidence       float64                `json:"confidence"`
	ConfidenceReason string                 `json:"confidence_reason"`
	State            assets.ConfidenceState `json:"state"`
	Metadata         json.RawMessage        `json:"metadata,omitempty"`
}

type ScriptFactCandidate struct {
	AssetID          string                 `json:"asset_id"`
	Name             string                 `json:"name"`
	Value            json.RawMessage        `json:"value"`
	Source           string                 `json:"source"`
	Confidence       float64                `json:"confidence"`
	ConfidenceReason string                 `json:"confidence_reason"`
	State            assets.ConfidenceState `json:"state"`
	EvidenceID       *string                `json:"evidence_id,omitempty"`
}

type ScriptRelationshipCandidate struct {
	SourceAssetID    string                 `json:"source_asset_id"`
	TargetAssetID    string                 `json:"target_asset_id"`
	RelationshipType string                 `json:"relationship_type"`
	Confidence       float64                `json:"confidence"`
	ConfidenceReason string                 `json:"confidence_reason"`
	State            assets.ConfidenceState `json:"state"`
	EvidenceID       *string                `json:"evidence_id,omitempty"`
	Metadata         json.RawMessage        `json:"metadata,omitempty"`
}

func NewScriptRunner(opts ScriptRunnerOptions) ScriptRunner {
	engine := opts.Policy
	if len(engine.AllowedTasks()) == 0 {
		engine = policy.NewEngine()
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultScriptTimeout
	}
	allowed := make(map[string]struct{}, len(opts.AllowedScripts))
	for _, script := range opts.AllowedScripts {
		if normalized := normalizeScriptPath(script); normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}
	return ScriptRunner{
		enabled:        opts.Enabled,
		allowedScripts: allowed,
		timeout:        timeout,
		policy:         engine,
	}
}

func (ScriptRunner) Metadata() Metadata {
	return Metadata{
		Name:           "byo_script_runner",
		Kind:           KindImporter,
		Version:        "v1",
		ExternalSystem: "local_script",
		Capabilities:   []string{"json_stdin", "json_stdout", "explicit_local_opt_in"},
		ReadOnly:       true,
	}
}

func (r ScriptRunner) Import(ctx context.Context, request ImportRequest) (ImportResult, error) {
	if !r.enabled {
		return ImportResult{}, ErrScriptRunnerDisabled
	}
	script := normalizeScriptPath(request.Source)
	if script == "" {
		return ImportResult{}, fmt.Errorf("script path is required")
	}
	if _, ok := r.allowedScripts[script]; !ok {
		return ImportResult{}, fmt.Errorf("%w: %s", ErrScriptNotAllowed, script)
	}

	input, err := scriptInput(request)
	if err != nil {
		return ImportResult{}, err
	}
	if err := r.checkInput(input); err != nil {
		return ImportResult{}, err
	}

	payload, err := json.Marshal(input)
	if err != nil {
		return ImportResult{}, fmt.Errorf("encode script input: %w", err)
	}
	payload = append(payload, '\n')

	if err := ctx.Err(); err != nil {
		return ImportResult{}, err
	}
	runCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, script)
	cmd.Stdin = bytes.NewReader(payload)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if runCtx.Err() == context.DeadlineExceeded {
		return ImportResult{}, fmt.Errorf("script timed out after %s", r.timeout)
	}
	if err != nil {
		return ImportResult{}, fmt.Errorf("script exited unsuccessfully: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	var scriptOutput ScriptOutput
	decoder := json.NewDecoder(bytes.NewReader(output))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&scriptOutput); err != nil {
		return ImportResult{}, fmt.Errorf("decode script output: %w", err)
	}
	if err := r.checkOutput(scriptOutput); err != nil {
		return ImportResult{}, err
	}

	return ImportResult{
		Evidence:   scriptOutput.Evidence,
		Candidates: scriptOutput.Candidates.modelCandidates(),
		Warnings:   append([]string{"BYO script output is untrusted until persisted by the kernel"}, scriptOutput.Warnings...),
	}, nil
}

func (r ScriptRunner) checkInput(input ScriptInput) error {
	for _, task := range input.Tasks {
		if err := r.policy.CheckTask(task); err != nil {
			return err
		}
	}
	return nil
}

func (r ScriptRunner) checkOutput(output ScriptOutput) error {
	if len(output.Evidence) == 0 && len(output.Candidates.Assets) == 0 && len(output.Candidates.Facts) == 0 && len(output.Candidates.Relationships) == 0 {
		return fmt.Errorf("script output must include evidence or normalized candidates")
	}
	for _, item := range output.Evidence {
		if err := r.policy.CheckCommand(item.CommandOrAPI); err != nil {
			return fmt.Errorf("script evidence command_or_api %q failed policy: %w", item.CommandOrAPI, err)
		}
		if strings.TrimSpace(item.RawOutput) == "" {
			return fmt.Errorf("script evidence raw_output is required")
		}
	}
	return nil
}

func (c ScriptCandidates) modelCandidates() ModelCandidates {
	return ModelCandidates{
		Assets:        scriptAssetCandidates(c.Assets),
		Facts:         scriptFactCandidates(c.Facts),
		Relationships: scriptRelationshipCandidates(c.Relationships),
	}
}

func scriptAssetCandidates(items []ScriptAssetCandidate) []assets.CreateAssetParams {
	out := make([]assets.CreateAssetParams, 0, len(items))
	for _, item := range items {
		out = append(out, assets.CreateAssetParams{
			Type:             item.Type,
			IdentityKey:      item.IdentityKey,
			Vendor:           item.Vendor,
			Model:            item.Model,
			Serial:           item.Serial,
			SystemMAC:        item.SystemMAC,
			Confidence:       item.Confidence,
			ConfidenceReason: item.ConfidenceReason,
			State:            item.State,
			Metadata:         defaultJSON(item.Metadata),
		})
	}
	return out
}

func scriptFactCandidates(items []ScriptFactCandidate) []assets.CreateFactParams {
	out := make([]assets.CreateFactParams, 0, len(items))
	for _, item := range items {
		out = append(out, assets.CreateFactParams{
			AssetID:          item.AssetID,
			Name:             item.Name,
			Value:            defaultJSON(item.Value),
			Source:           item.Source,
			Confidence:       item.Confidence,
			ConfidenceReason: item.ConfidenceReason,
			State:            item.State,
			EvidenceID:       item.EvidenceID,
		})
	}
	return out
}

func scriptRelationshipCandidates(items []ScriptRelationshipCandidate) []assets.CreateRelationshipParams {
	out := make([]assets.CreateRelationshipParams, 0, len(items))
	for _, item := range items {
		out = append(out, assets.CreateRelationshipParams{
			SourceAssetID:    item.SourceAssetID,
			TargetAssetID:    item.TargetAssetID,
			RelationshipType: item.RelationshipType,
			Confidence:       item.Confidence,
			ConfidenceReason: item.ConfidenceReason,
			State:            item.State,
			EvidenceID:       item.EvidenceID,
			Metadata:         defaultJSON(item.Metadata),
		})
	}
	return out
}

func scriptInput(request ImportRequest) (ScriptInput, error) {
	input := ScriptInput{DryRun: request.DryRun, CredentialRef: strings.TrimSpace(request.CredentialRef)}
	if len(request.Scope) == 0 || strings.TrimSpace(string(request.Scope)) == "" {
		return input, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(request.Scope))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return ScriptInput{}, fmt.Errorf("decode script scope: %w", err)
	}
	input.DryRun = request.DryRun
	if input.CredentialRef == "" {
		input.CredentialRef = strings.TrimSpace(request.CredentialRef)
	}
	return input, nil
}

func normalizeScriptPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}
