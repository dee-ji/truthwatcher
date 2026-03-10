package junos

import (
	"context"
	"testing"

	"github.com/truthwatcher/truthwatcher/internal/rendering"
)

func TestRenderDeterministicSetOutput(t *testing.T) {
	ir := rendering.DeviceConfigIR{
		Hostname:     "leaf-1",
		Role:         "leaf",
		BGPASN:       65123,
		Services:     []string{"l3-underlay"},
		TargetSites:  []string{"dc1"},
		TODOSections: []string{"interface_intent"},
	}
	out, err := Driver{}.Render(context.Background(), ir)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	want := "set system host-name leaf-1\n" +
		"set routing-options autonomous-system 65123\n" +
		"## TODO(truthwatcher): service \"l3-underlay\" is not yet rendered for junos\n" +
		"## TODO(truthwatcher): section \"interface_intent\" is not yet rendered for junos\n"
	if out.Contents != want {
		t.Fatalf("unexpected output:\n%s", out.Contents)
	}
	if out.Format != "set" || out.Vendor != "junos" || out.Filename != "leaf-1.set" {
		t.Fatalf("unexpected metadata: %#v", out)
	}
}
