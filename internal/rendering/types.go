package rendering

import (
	"context"
	"sort"
)

type DeviceConfigIR struct {
	Hostname     string
	Role         string
	BGPASN       int
	Services     []string
	TargetSites  []string
	TODOSections []string
}

func (ir *DeviceConfigIR) Normalize() {
	sort.Strings(ir.Services)
	sort.Strings(ir.TargetSites)
	sort.Strings(ir.TODOSections)
}

type Artifact struct {
	Vendor   string
	Format   string
	Filename string
	Contents string
	Metadata map[string]string
}

type Driver interface {
	Vendor() string
	Render(context.Context, DeviceConfigIR) (Artifact, error)
}
