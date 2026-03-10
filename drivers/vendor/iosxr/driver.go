package iosxr

import (
	"context"
	"fmt"

	"github.com/truthwatcher/truthwatcher/internal/rendering"
)

type Driver struct{}

func (Driver) Vendor() string { return "iosxr" }

func (Driver) Render(_ context.Context, ir rendering.DeviceConfigIR) (rendering.Artifact, error) {
	return rendering.Artifact{Vendor: "iosxr", Format: "text", Filename: ir.Hostname + ".cfg", Contents: fmt.Sprintf("! vendor=iosxr\nhostname %s\n", ir.Hostname)}, nil
}
