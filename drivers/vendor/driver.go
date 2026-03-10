package vendor

import (
	"context"

	"github.com/truthwatcher/truthwatcher/internal/rendering"
)

// Renderer describes a vendor renderer implementation backed by vendor-neutral IR.
type Renderer interface {
	Vendor() string
	Render(context.Context, rendering.DeviceConfigIR) (rendering.Artifact, error)
}
