package iosxr

import (
	"context"
	"testing"

	"github.com/truthwatcher/truthwatcher/internal/rendering"
)

func TestRender(t *testing.T) {
	out, err := Driver{}.Render(context.Background(), rendering.DeviceConfigIR{Hostname: "leaf-1"})
	if err != nil || out.Contents == "" {
		t.Fatalf("expected output")
	}
}
