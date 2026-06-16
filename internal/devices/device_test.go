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
		ID:                "device-1",
		Name:              params.Name,
		ManagementAddress: params.ManagementAddress,
		Platform:          params.Platform,
		Vendor:            params.Vendor,
		Model:             params.Model,
	}, nil
}

func (f *fakeRepository) ListDevices(ctx context.Context) ([]Device, error) {
	return f.devices, nil
}

func TestCreateDeviceValidatesName(t *testing.T) {
	_, err := NewService(&fakeRepository{}).CreateDevice(context.Background(), CreateDeviceParams{
		Name: " ",
	})
	if err == nil {
		t.Fatal("CreateDevice returned nil error for missing name")
	}
}

func TestCreateDeviceTrimsOptionalFields(t *testing.T) {
	repo := &fakeRepository{}
	address := " 192.0.2.10 "
	platform := " junos "
	empty := " "

	device, err := NewService(repo).CreateDevice(context.Background(), CreateDeviceParams{
		Name:              " mx-edge-01 ",
		ManagementAddress: &address,
		Platform:          &platform,
		Vendor:            &empty,
	})
	if err != nil {
		t.Fatalf("CreateDevice returned error: %v", err)
	}

	if device.Name != "mx-edge-01" {
		t.Fatalf("Name = %q, want mx-edge-01", device.Name)
	}
	if repo.createParams.ManagementAddress == nil || *repo.createParams.ManagementAddress != "192.0.2.10" {
		t.Fatalf("ManagementAddress = %v, want trimmed address", repo.createParams.ManagementAddress)
	}
	if repo.createParams.Platform == nil || *repo.createParams.Platform != "junos" {
		t.Fatalf("Platform = %v, want junos", repo.createParams.Platform)
	}
	if repo.createParams.Vendor != nil {
		t.Fatalf("Vendor = %v, want nil for blank optional field", repo.createParams.Vendor)
	}
}

func TestListDevicesUsesRepository(t *testing.T) {
	repo := &fakeRepository{devices: []Device{{ID: "device-1", Name: "mx-edge-01"}}}

	devices, err := NewService(repo).ListDevices(context.Background())
	if err != nil {
		t.Fatalf("ListDevices returned error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("len = %d, want 1", len(devices))
	}
	if devices[0].Name != "mx-edge-01" {
		t.Fatalf("device name = %q, want mx-edge-01", devices[0].Name)
	}
}
