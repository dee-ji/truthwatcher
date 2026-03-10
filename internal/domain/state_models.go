package domain

import "time"

type ConfigSnapshot struct {
	ID          string    `json:"id"`
	DeviceID    string    `json:"device_id"`
	CapturedAt  time.Time `json:"captured_at"`
	ArtifactRef string    `json:"artifact_ref,omitempty"`
	Content     string    `json:"content"`
	Source      string    `json:"source,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type OperationalSnapshot struct {
	ID         string         `json:"id"`
	DeviceID   string         `json:"device_id"`
	CapturedAt time.Time      `json:"captured_at"`
	Content    map[string]any `json:"content"`
	Source     string         `json:"source,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

type DriftFinding struct {
	ID               string         `json:"id"`
	ReconcileRunID   string         `json:"reconcile_run_id,omitempty"`
	DeviceID         string         `json:"device_id"`
	Severity         string         `json:"severity"`
	Kind             string         `json:"kind"`
	Summary          string         `json:"summary"`
	IntendedArtifact string         `json:"intended_artifact,omitempty"`
	ActualSnapshotID string         `json:"actual_snapshot_id,omitempty"`
	Finding          map[string]any `json:"finding,omitempty"`
	Remediation      map[string]any `json:"remediation,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
}

type ReconcileRunRequest struct {
	IntentID string `json:"intent_id"`
	Actor    string `json:"actor,omitempty"`
}

type ReconcileRun struct {
	ID            string         `json:"id"`
	IntentID      string         `json:"intent_id"`
	Status        string         `json:"status"`
	Summary       string         `json:"summary,omitempty"`
	FindingsCount int            `json:"findings_count"`
	CreatedAt     time.Time      `json:"created_at"`
	CompletedAt   time.Time      `json:"completed_at,omitempty"`
	Findings      []DriftFinding `json:"findings,omitempty"`
}
