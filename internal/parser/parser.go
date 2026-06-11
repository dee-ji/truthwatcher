package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"truthwatcher/internal/evidence"
)

const NoopParserName = "noop"

// Parser converts raw evidence into normalized model candidates.
type Parser interface {
	Name() string
	Supports(platform string, command string) bool
	Parse(context.Context, evidence.Evidence) (Result, error)
}

type Result struct {
	ParserName          string               `json:"parser_name"`
	EvidenceID          string               `json:"evidence_id"`
	DeviceIdentities    []DeviceIdentity     `json:"device_identities,omitempty"`
	InventoryComponents []InventoryComponent `json:"inventory_components,omitempty"`
	Interfaces          []Interface          `json:"interfaces,omitempty"`
	Neighbors           []Neighbor           `json:"neighbors,omitempty"`
	BGPPeers            []BGPPeer            `json:"bgp_peers,omitempty"`
	Facts               []ParsedFact         `json:"facts,omitempty"`
	Relationships       []ParsedRelationship `json:"relationships,omitempty"`
	Warnings            []string             `json:"warnings,omitempty"`
}

// DeviceIdentity describes a candidate device asset identity.
type DeviceIdentity struct {
	AssetRef
	Hostname  string          `json:"hostname,omitempty"`
	Vendor    string          `json:"vendor,omitempty"`
	Model     string          `json:"model,omitempty"`
	Serial    string          `json:"serial,omitempty"`
	SystemMAC string          `json:"system_mac,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

// InventoryComponent describes a chassis, card, port, optic, or similar asset.
type InventoryComponent struct {
	AssetRef
	ParentIdentityKey string          `json:"parent_identity_key,omitempty"`
	ComponentType     string          `json:"component_type"`
	Name              string          `json:"name,omitempty"`
	Vendor            string          `json:"vendor,omitempty"`
	Model             string          `json:"model,omitempty"`
	Serial            string          `json:"serial,omitempty"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

type Interface struct {
	AssetRef
	DeviceIdentityKey string          `json:"device_identity_key,omitempty"`
	Name              string          `json:"name"`
	AdminState        string          `json:"admin_state,omitempty"`
	OperState         string          `json:"oper_state,omitempty"`
	MACAddress        string          `json:"mac_address,omitempty"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

type Neighbor struct {
	LocalInterfaceIdentityKey string          `json:"local_interface_identity_key,omitempty"`
	LocalInterfaceName        string          `json:"local_interface_name,omitempty"`
	RemoteIdentityKey         string          `json:"remote_identity_key,omitempty"`
	RemoteSystemName          string          `json:"remote_system_name,omitempty"`
	RemoteInterfaceName       string          `json:"remote_interface_name,omitempty"`
	Protocol                  string          `json:"protocol,omitempty"`
	Confidence                float64         `json:"confidence"`
	Metadata                  json.RawMessage `json:"metadata,omitempty"`
}

type BGPPeer struct {
	DeviceIdentityKey string          `json:"device_identity_key,omitempty"`
	PeerAddress       string          `json:"peer_address"`
	RemoteASN         uint32          `json:"remote_asn,omitempty"`
	State             string          `json:"state,omitempty"`
	AcceptedPrefixes  *int            `json:"accepted_prefixes,omitempty"`
	Confidence        float64         `json:"confidence"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

// AssetRef points to an asset candidate before database IDs exist.
type AssetRef struct {
	AssetType   string  `json:"asset_type"`
	IdentityKey string  `json:"identity_key"`
	Confidence  float64 `json:"confidence"`
	EvidenceID  string  `json:"evidence_id,omitempty"`
}

// ParsedFact mirrors the stable fact model using identity keys before asset IDs.
type ParsedFact struct {
	AssetIdentityKey string          `json:"asset_identity_key"`
	Name             string          `json:"name"`
	Value            json.RawMessage `json:"value"`
	Source           string          `json:"source"`
	Confidence       float64         `json:"confidence"`
	EvidenceID       string          `json:"evidence_id,omitempty"`
}

// ParsedRelationship mirrors the stable relationship model using identity keys before asset IDs.
type ParsedRelationship struct {
	SourceIdentityKey string          `json:"source_identity_key"`
	TargetIdentityKey string          `json:"target_identity_key"`
	RelationshipType  string          `json:"relationship_type"`
	Confidence        float64         `json:"confidence"`
	EvidenceID        string          `json:"evidence_id,omitempty"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

type Registry struct {
	parsers  []Parser
	fallback Parser
}

func NewRegistry(parsers ...Parser) Registry {
	copied := append([]Parser(nil), parsers...)
	return Registry{
		parsers:  copied,
		fallback: NoopParser{},
	}
}

func (r Registry) Select(platform string, command string) Parser {
	for _, parser := range r.parsers {
		if parser == nil {
			continue
		}
		if parser.Supports(platform, command) {
			return parser
		}
	}
	if r.fallback == nil {
		return NoopParser{}
	}
	return r.fallback
}

func (r Registry) Parse(ctx context.Context, platform string, item evidence.Evidence) (Result, error) {
	selected := r.Select(platform, item.CommandOrAPI)
	result, err := selected.Parse(ctx, item)
	if err != nil {
		return Result{}, err
	}
	if result.ParserName == "" {
		result.ParserName = selected.Name()
	}
	if result.EvidenceID == "" {
		result.EvidenceID = item.ID
	}
	return result, nil
}

// NoopParser intentionally returns no model candidates for unsupported evidence.
type NoopParser struct{}

func (NoopParser) Name() string {
	return NoopParserName
}

func (NoopParser) Supports(platform string, command string) bool {
	return true
}

func (NoopParser) Parse(ctx context.Context, item evidence.Evidence) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	return Result{
		ParserName: NoopParserName,
		EvidenceID: item.ID,
		Warnings: []string{
			fmt.Sprintf("no parser registered for command %q", strings.TrimSpace(item.CommandOrAPI)),
		},
	}, nil
}
