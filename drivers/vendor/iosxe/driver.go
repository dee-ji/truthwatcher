package iosxe

import (
	"context"
	"fmt"

	"github.com/truthwatcher/truthwatcher/internal/rendering"
)

type Driver struct{}

func (Driver) Vendor() string { return "iosxe" }

func (Driver) Render(_ context.Context, ir rendering.DeviceConfigIR) (rendering.Artifact, error) {
	return rendering.Artifact{Vendor: "iosxe", Format: "text", Filename: ir.Hostname + ".cfg", Contents: fmt.Sprintf("! vendor=iosxe\nhostname %s\n", ir.Hostname)}, nil
}
