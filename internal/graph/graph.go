package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"truthwatcher/internal/assets"
)

type AssetReader interface {
	GetAsset(context.Context, string) (assets.Asset, error)
	ListFactsByAsset(context.Context, string) ([]assets.Fact, error)
	ListRelationships(context.Context) ([]assets.Relationship, error)
}

type Service struct {
	reader AssetReader
}

func NewService(reader AssetReader) Service {
	return Service{reader: reader}
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Node struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Label            string                 `json:"label"`
	IdentityKey      string                 `json:"identity_key"`
	Vendor           *string                `json:"vendor,omitempty"`
	Model            *string                `json:"model,omitempty"`
	Serial           *string                `json:"serial,omitempty"`
	SystemMAC        *string                `json:"system_mac,omitempty"`
	Confidence       float64                `json:"confidence"`
	ConfidenceReason string                 `json:"confidence_reason"`
	State            assets.ConfidenceState `json:"state"`
	Metadata         json.RawMessage        `json:"metadata"`
	Facts            []assets.Fact          `json:"facts,omitempty"`
}

type Edge struct {
	ID               string                 `json:"id"`
	Source           string                 `json:"source"`
	Target           string                 `json:"target"`
	RelationshipType string                 `json:"relationship_type"`
	Confidence       float64                `json:"confidence"`
	ConfidenceReason string                 `json:"confidence_reason"`
	State            assets.ConfidenceState `json:"state"`
	EvidenceID       *string                `json:"evidence_id,omitempty"`
	Metadata         json.RawMessage        `json:"metadata"`
}

type PathCandidate struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func (s Service) GetAssetGraph(ctx context.Context, assetID string) (Graph, error) {
	return s.GetAssetGraphWithDepth(ctx, assetID, 1)
}

func (s Service) GetAssetGraphWithDepth(ctx context.Context, assetID string, maxDepth int) (Graph, error) {
	if strings.TrimSpace(assetID) == "" {
		return Graph{}, fmt.Errorf("asset_id is required")
	}
	if s.reader == nil {
		return Graph{}, fmt.Errorf("asset reader is required")
	}

	if maxDepth < 1 {
		maxDepth = 1
	}
	if maxDepth > 2 {
		maxDepth = 2
	}

	root, err := s.reader.GetAsset(ctx, assetID)
	if err != nil {
		return Graph{}, err
	}
	facts, err := s.reader.ListFactsByAsset(ctx, assetID)
	if err != nil {
		return Graph{}, err
	}

	graph := Graph{Nodes: []Node{nodeFromAsset(root, facts)}}
	seenNodes := map[string]struct{}{root.ID: {}}
	seenEdges := map[string]struct{}{}
	frontier := []string{root.ID}

	relationships, err := s.reader.ListRelationships(ctx)
	if err != nil {
		return Graph{}, err
	}
	for depth := 0; depth < maxDepth; depth++ {
		nextFrontier := []string{}
		frontierSet := make(map[string]struct{}, len(frontier))
		for _, id := range frontier {
			frontierSet[id] = struct{}{}
		}

		for _, relationship := range relationships {
			_, sourceInFrontier := frontierSet[relationship.SourceAssetID]
			_, targetInFrontier := frontierSet[relationship.TargetAssetID]
			if !sourceInFrontier && !targetInFrontier {
				continue
			}
			if _, ok := seenEdges[relationship.ID]; !ok {
				graph.Edges = append(graph.Edges, edgeFromRelationship(relationship))
				seenEdges[relationship.ID] = struct{}{}
			}

			for _, neighborID := range []string{relationship.SourceAssetID, relationship.TargetAssetID} {
				if _, ok := seenNodes[neighborID]; ok {
					continue
				}
				neighbor, err := s.reader.GetAsset(ctx, neighborID)
				if err != nil {
					return Graph{}, err
				}
				neighborFacts, err := s.reader.ListFactsByAsset(ctx, neighborID)
				if err != nil {
					return Graph{}, err
				}
				graph.Nodes = append(graph.Nodes, nodeFromAsset(neighbor, neighborFacts))
				seenNodes[neighborID] = struct{}{}
				nextFrontier = append(nextFrontier, neighborID)
			}
		}
		frontier = nextFrontier
		if len(frontier) == 0 {
			break
		}
	}

	return graph, nil
}

func (s Service) GetNeighbors(ctx context.Context, assetID string) (Graph, error) {
	return s.GetAssetGraph(ctx, assetID)
}

func (s Service) PathCandidates(ctx context.Context, assetID string) ([]PathCandidate, error) {
	graph, err := s.GetAssetGraph(ctx, assetID)
	if err != nil {
		return nil, err
	}

	nodesByID := make(map[string]Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodesByID[node.ID] = node
	}

	candidates := make([]PathCandidate, 0, len(graph.Edges))
	for _, edge := range graph.Edges {
		source, ok := nodesByID[edge.Source]
		if !ok {
			continue
		}
		target, ok := nodesByID[edge.Target]
		if !ok {
			continue
		}
		candidates = append(candidates, PathCandidate{
			Nodes: []Node{source, target},
			Edges: []Edge{edge},
		})
	}
	return candidates, nil
}

func nodeFromAsset(asset assets.Asset, facts []assets.Fact) Node {
	return Node{
		ID:               asset.ID,
		Type:             asset.Type,
		Label:            labelForAsset(asset, facts),
		IdentityKey:      asset.IdentityKey,
		Vendor:           asset.Vendor,
		Model:            asset.Model,
		Serial:           asset.Serial,
		SystemMAC:        asset.SystemMAC,
		Confidence:       asset.Confidence,
		ConfidenceReason: asset.ConfidenceReason,
		State:            asset.State,
		Metadata:         asset.Metadata,
		Facts:            facts,
	}
}

func edgeFromRelationship(relationship assets.Relationship) Edge {
	return Edge{
		ID:               relationship.ID,
		Source:           relationship.SourceAssetID,
		Target:           relationship.TargetAssetID,
		RelationshipType: relationship.RelationshipType,
		Confidence:       relationship.Confidence,
		ConfidenceReason: relationship.ConfidenceReason,
		State:            relationship.State,
		EvidenceID:       relationship.EvidenceID,
		Metadata:         relationship.Metadata,
	}
}

func touchesAsset(relationship assets.Relationship, assetID string) bool {
	return relationship.SourceAssetID == assetID || relationship.TargetAssetID == assetID
}

func otherAssetID(relationship assets.Relationship, assetID string) string {
	if relationship.SourceAssetID == assetID {
		return relationship.TargetAssetID
	}
	return relationship.SourceAssetID
}

func labelForAsset(asset assets.Asset, facts []assets.Fact) string {
	for _, fact := range facts {
		if fact.Name != "hostname" {
			continue
		}
		var hostname string
		if err := json.Unmarshal(fact.Value, &hostname); err == nil && strings.TrimSpace(hostname) != "" {
			return strings.TrimSpace(hostname)
		}
	}
	if strings.TrimSpace(asset.IdentityKey) != "" {
		return asset.IdentityKey
	}
	return asset.ID
}
