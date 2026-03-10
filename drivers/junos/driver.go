package junos

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/truthwatcher/truthwatcher/internal/rendering"
)

type Driver struct{}

func (Driver) Vendor() string { return "junos" }

func (Driver) Render(_ context.Context, ir rendering.DeviceConfigIR) (rendering.Artifact, error) {
	lines := []string{
		fmt.Sprintf("set system host-name %s", ir.Hostname),
		fmt.Sprintf("set routing-options autonomous-system %d", ir.BGPASN),
	}
	for _, svc := range ir.Services {
		lines = append(lines, fmt.Sprintf("## TODO(truthwatcher): service %q is not yet rendered for junos", svc))
	}
	for _, section := range ir.TODOSections {
		lines = append(lines, fmt.Sprintf("## TODO(truthwatcher): section %q is not yet rendered for junos", section))
	}
	sort.Strings(ir.TargetSites)
	metadata := map[string]string{
		"hostname":     ir.Hostname,
		"role":         ir.Role,
		"target_sites": strings.Join(ir.TargetSites, ","),
	}
	return rendering.Artifact{
		Vendor:   "junos",
		Format:   "set",
		Filename: fmt.Sprintf("%s.set", ir.Hostname),
		Contents: strings.Join(lines, "\n") + "\n",
		Metadata: metadata,
	}, nil
}
