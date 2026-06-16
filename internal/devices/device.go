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
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	ManagementAddress *string   `json:"management_address,omitempty"`
	Platform          *string   `json:"platform,omitempty"`
	Vendor            *string   `json:"vendor,omitempty"`
	Model             *string   `json:"model,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CreateDeviceParams struct {
	Name              string
	ManagementAddress *string
	Platform          *string
	Vendor            *string
	Model             *string
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

	params.Name = strings.TrimSpace(params.Name)
	if params.Name == "" {
		return Device{}, fmt.Errorf("device name is required")
	}

	params.ManagementAddress = cleanOptional(params.ManagementAddress)
	params.Platform = cleanOptional(params.Platform)
	params.Vendor = cleanOptional(params.Vendor)
	params.Model = cleanOptional(params.Model)

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
