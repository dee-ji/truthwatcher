package devices

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("device not found")

type Device struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	Vendor       *string   `json:"vendor,omitempty"`
	Model        *string   `json:"model,omitempty"`
	SerialNumber *string   `json:"serial_number,omitempty"`
	ManagementIP *string   `json:"management_ip,omitempty"`
	Role         *string   `json:"role,omitempty"`
	Site         *string   `json:"site,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateDeviceParams struct {
	Hostname     string
	Vendor       *string
	Model        *string
	SerialNumber *string
	ManagementIP *string
	Role         *string
	Site         *string
}

type Repository interface {
	CreateDevice(context.Context, CreateDeviceParams) (Device, error)
	ListDevices(context.Context) ([]Device, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) CreateDevice(ctx context.Context, params CreateDeviceParams) (Device, error) {
	if s.repo == nil {
		return Device{}, fmt.Errorf("device repository is required")
	}

	params.Hostname = strings.TrimSpace(params.Hostname)
	if params.Hostname == "" {
		return Device{}, fmt.Errorf("device hostname is required")
	}

	params.Vendor = cleanOptional(params.Vendor)
	params.Model = cleanOptional(params.Model)
	params.SerialNumber = cleanOptional(params.SerialNumber)
	params.ManagementIP = cleanOptional(params.ManagementIP)
	params.Role = cleanOptional(params.Role)
	params.Site = cleanOptional(params.Site)

	return s.repo.CreateDevice(ctx, params)
}

func (s Service) ListDevices(ctx context.Context) ([]Device, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("device repository is required")
	}
	return s.repo.ListDevices(ctx)
}

func cleanOptional(value *string) *string {
	if value == nil {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil
	}
	return &cleaned
}
