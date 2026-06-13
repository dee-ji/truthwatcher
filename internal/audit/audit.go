package audit

import (
	"encoding/json"
	"strings"
	"time"
)

const DefaultInitiator = "unknown"

// DiscoveryAction records the minimum safety-relevant audit fields for one
// discovery command/API action. Persistence can later move this into a table;
// for now it is embedded in discovery/evidence metadata and API responses.
type DiscoveryAction struct {
	Initiator      string          `json:"initiator"`
	RequestID      string          `json:"request_id,omitempty"`
	DiscoveryRunID string          `json:"discovery_run_id"`
	Target         string          `json:"target"`
	Method         string          `json:"method"`
	Profile        string          `json:"profile"`
	Task           string          `json:"task"`
	CommandOrAPI   string          `json:"command_or_api"`
	Status         string          `json:"status"`
	EvidenceID     string          `json:"evidence_id,omitempty"`
	StartedAt      time.Time       `json:"started_at"`
	CompletedAt    time.Time       `json:"completed_at"`
	Context        json.RawMessage `json:"context,omitempty"`
}

func NormalizeInitiator(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultInitiator
	}
	return value
}
