package domain

import "time"

type Intent struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Revision    int                    `json:"revision"`
	CreatedAt   time.Time              `json:"created_at"`
	Spec        map[string]any         `json:"spec,omitempty"`
	Artifacts   []CompiledArtifactView `json:"artifacts,omitempty"`
}

type CompiledArtifactView struct {
	Vendor    string            `json:"vendor"`
	Format    string            `json:"format"`
	Artifact  string            `json:"artifact"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type Deployment struct {
	ID             string    `json:"id"`
	IntentID       string    `json:"intent_id"`
	Status         string    `json:"status"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}

type Device struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	Vendor   string `json:"vendor"`
}

type Link struct {
	ID         string `json:"id"`
	FromDevice string `json:"from_device"`
	ToDevice   string `json:"to_device"`
}

type AuditEvent struct {
	ID        string         `json:"id"`
	Actor     string         `json:"actor"`
	Action    string         `json:"action"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type ReconcileRun struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
