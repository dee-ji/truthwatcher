package elsecall

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/truthwatcher/truthwatcher/drivers/junos"
	"github.com/truthwatcher/truthwatcher/internal/rendering"
)

var ErrUnsupportedVendor = errors.New("unsupported vendor")

type CompilerService struct {
	drivers map[string]rendering.Driver
}

func NewCompilerService() *CompilerService {
	j := junos.Driver{}
	return &CompilerService{drivers: map[string]rendering.Driver{j.Vendor(): j}}
}

func (s *CompilerService) BuildIntermediate(spec map[string]any) rendering.DeviceConfigIR {
	ir := rendering.DeviceConfigIR{Hostname: "unnamed-device", Role: "leaf"}
	if metadata, ok := spec["metadata"].(map[string]any); ok {
		if name, ok := metadata["name"].(string); ok && name != "" {
			ir.Hostname = name
		}
	}
	if routing, ok := spec["routing_intent"].(map[string]any); ok {
		if bgp, ok := routing["bgp"].(map[string]any); ok {
			switch asn := bgp["asn"].(type) {
			case int:
				ir.BGPASN = asn
			case float64:
				ir.BGPASN = int(asn)
			}
		}
	}
	if svc, ok := spec["desired_services"].([]any); ok {
		for _, item := range svc {
			if m, ok := item.(map[string]any); ok {
				if t, ok := m["type"].(string); ok && t != "" {
					ir.Services = append(ir.Services, t)
				}
			}
		}
	}
	if scope, ok := spec["target_scope"].(map[string]any); ok {
		if sites, ok := scope["sites"].([]any); ok {
			for _, site := range sites {
				if s, ok := site.(string); ok && s != "" {
					ir.TargetSites = append(ir.TargetSites, s)
				}
			}
		}
	}
	if _, ok := spec["interface_intent"]; ok {
		ir.TODOSections = append(ir.TODOSections, "interface_intent")
	}
	if ir.BGPASN == 0 {
		ir.BGPASN = 65000
	}
	ir.Normalize()
	return ir
}

func (s *CompilerService) RenderForVendor(ctx context.Context, vendor string, ir rendering.DeviceConfigIR) (rendering.Artifact, error) {
	d, ok := s.drivers[vendor]
	if !ok {
		keys := make([]string, 0, len(s.drivers))
		for k := range s.drivers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return rendering.Artifact{}, fmt.Errorf("%w: %s (supported: %v)", ErrUnsupportedVendor, vendor, keys)
	}
	return d.Render(ctx, ir)
}
