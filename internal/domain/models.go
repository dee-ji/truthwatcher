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

type Vendor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Platform struct {
	ID       string `json:"id"`
	VendorID string `json:"vendor_id"`
	Name     string `json:"name"`
}

type Site struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Device struct {
	ID         string `json:"id"`
	Hostname   string `json:"hostname"`
	Vendor     string `json:"vendor"`
	Platform   string `json:"platform,omitempty"`
	Site       string `json:"site,omitempty"`
	VendorID   string `json:"vendor_id,omitempty"`
	PlatformID string `json:"platform_id,omitempty"`
	SiteID     string `json:"site_id,omitempty"`
}

type Interface struct {
	ID       string `json:"id"`
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}

type Link struct {
	ID             string `json:"id"`
	AInterfaceID   string `json:"a_interface_id"`
	ZInterfaceID   string `json:"z_interface_id"`
	FromDeviceID   string `json:"from_device_id,omitempty"`
	ToDeviceID     string `json:"to_device_id,omitempty"`
	FromDevice     string `json:"from_device,omitempty"`
	ToDevice       string `json:"to_device,omitempty"`
	AInterfaceName string `json:"a_interface_name,omitempty"`
	ZInterfaceName string `json:"z_interface_name,omitempty"`
}

type DeviceDetail struct {
	Device            Device      `json:"device"`
	Interfaces        []Interface `json:"interfaces"`
	AdjacentDeviceIDs []string    `json:"adjacent_device_ids"`
	Links             []Link      `json:"links"`
}

type TopologySnapshot struct {
	Vendors    []Vendor    `json:"vendors" yaml:"vendors"`
	Platforms  []Platform  `json:"platforms" yaml:"platforms"`
	Sites      []Site      `json:"sites" yaml:"sites"`
	Devices    []Device    `json:"devices" yaml:"devices"`
	Interfaces []Interface `json:"interfaces" yaml:"interfaces"`
	Links      []Link      `json:"links" yaml:"links"`
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
