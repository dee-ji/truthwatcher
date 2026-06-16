package devices

import (
	"context"
	"testing"
)

type fakeRepository struct {
	createParams CreateDeviceParams
	devices      []Device
}

func (f *fakeRepository) CreateDevice(ctx context.Context, params CreateDeviceParams) (Device, error) {
	f.createParams = params
	return Device{
		ID:           "device-1",
		Hostname:     params.Hostname,
		Vendor:       params.Vendor,
		Model:        params.Model,
		SerialNumber: params.SerialNumber,
		ManagementIP: params.ManagementIP,
		Role:         params.Role,
		Site:         params.Site,
	}, nil
}

func (f *fakeRepository) ListDevices(ctx context.Context) ([]Device, error) {
	return f.devices, nil
}

func TestCreateDeviceValidatesHostname(t *testing.T) {
	_, err := NewService(&fakeRepository{}).CreateDevice(context.Background(), CreateDeviceParams{
		Hostname: " ",
	})
	if err == nil {
		t.Fatal("CreateDevice returned nil error for missing hostname")
	}
}

func TestCreateDeviceTrimsOptionalFields(t *testing.T) {
	repo := &fakeRepository{}
	managementIP := " 192.0.2.10 "
	serialNumber := " ABC123 "
	role := " edge "
	empty := " "

	device, err := NewService(repo).CreateDevice(context.Background(), CreateDeviceParams{
		Hostname:     " mx-edge-01 ",
		ManagementIP: &managementIP,
		SerialNumber: &serialNumber,
		Role:         &role,
		Vendor:       &empty,
	})
	if err != nil {
		t.Fatalf("CreateDevice returned error: %v", err)
	}

	if device.Hostname != "mx-edge-01" {
		t.Fatalf("Hostname = %q, want mx-edge-01", device.Hostname)
	}
	if repo.createParams.ManagementIP == nil || *repo.createParams.ManagementIP != "192.0.2.10" {
		t.Fatalf("ManagementIP = %v, want trimmed address", repo.createParams.ManagementIP)
	}
	if repo.createParams.SerialNumber == nil || *repo.createParams.SerialNumber != "ABC123" {
		t.Fatalf("SerialNumber = %v, want ABC123", repo.createParams.SerialNumber)
	}
	if repo.createParams.Role == nil || *repo.createParams.Role != "edge" {
		t.Fatalf("Role = %v, want edge", repo.createParams.Role)
	}
	if repo.createParams.Vendor != nil {
		t.Fatalf("Vendor = %v, want nil for blank optional field", repo.createParams.Vendor)
	}
}

func TestListDevicesUsesRepository(t *testing.T) {
	repo := &fakeRepository{devices: []Device{{ID: "device-1", Hostname: "mx-edge-01"}}}

	devices, err := NewService(repo).ListDevices(context.Background())
	if err != nil {
		t.Fatalf("ListDevices returned error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("len = %d, want 1", len(devices))
	}
	if devices[0].Hostname != "mx-edge-01" {
		t.Fatalf("device hostname = %q, want mx-edge-01", devices[0].Hostname)
	}
}
