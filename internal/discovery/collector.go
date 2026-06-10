package discovery

import (
	"context"

	"truthwatcher/internal/policy"
)

// Collector gathers raw command/API outputs for approved discovery tasks.
type Collector interface {
	Collect(ctx context.Context, target string, profile Profile, tasks []policy.Task) ([]CollectedOutput, error)
}

// CollectedOutput is raw discovery evidence before persistence.
type CollectedOutput struct {
	Target      string
	Method      string
	Task        policy.Task
	Command     string
	RawOutput   string
	ParserHints []string
	ProfileName string
	Platform    string
	Vendor      string
}
