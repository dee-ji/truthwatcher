package topology

import (
	"context"
	"testing"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

func TestInMemoryRepositoryAndAdjacency(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)
	snapshot := domain.TopologySnapshot{
		Vendors:   []domain.Vendor{{ID: "v1", Name: "eos"}},
		Platforms: []domain.Platform{{ID: "p1", VendorID: "v1", Name: "7050"}},
		Sites:     []domain.Site{{ID: "s1", Name: "dc1"}},
		Devices: []domain.Device{
			{ID: "d1", Hostname: "leaf1", VendorID: "v1", PlatformID: "p1", SiteID: "s1"},
			{ID: "d2", Hostname: "spine1", VendorID: "v1", PlatformID: "p1", SiteID: "s1"},
		},
		Interfaces: []domain.Interface{{ID: "i1", DeviceID: "d1", Name: "eth1"}, {ID: "i2", DeviceID: "d2", Name: "eth1"}},
		Links:      []domain.Link{{ID: "l1", AInterfaceID: "i1", ZInterfaceID: "i2"}},
	}
	if err := svc.Import(context.Background(), snapshot); err != nil {
		t.Fatal(err)
	}
	devices, err := svc.Devices(context.Background(), DeviceFilter{Site: "dc1"})
	if err != nil || len(devices) != 2 {
		t.Fatalf("expected 2 devices: err=%v len=%d", err, len(devices))
	}
	adj, err := svc.AdjacentDeviceIDs(context.Background(), "d1")
	if err != nil || len(adj) != 1 || adj[0] != "d2" {
		t.Fatalf("unexpected adjacency: %v err=%v", adj, err)
	}
	detail, err := svc.Device(context.Background(), "d1")
	if err != nil || len(detail.Interfaces) != 1 || len(detail.Links) != 1 {
		t.Fatalf("unexpected detail: %+v err=%v", detail, err)
	}
}
