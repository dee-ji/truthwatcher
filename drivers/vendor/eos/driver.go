package eos

import (
	"context"
	"fmt"

	"github.com/truthwatcher/truthwatcher/internal/rendering"
)

type Driver struct{}

func (Driver) Vendor() string { return "eos" }

func (Driver) Render(_ context.Context, ir rendering.DeviceConfigIR) (rendering.Artifact, error) {
	return rendering.Artifact{Vendor: "eos", Format: "text", Filename: ir.Hostname + ".cfg", Contents: fmt.Sprintf("! vendor=eos\nhostname %s\n", ir.Hostname)}, nil
}
