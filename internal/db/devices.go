package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"truthwatcher/internal/devices"
	"truthwatcher/internal/discovery"
)

type DeviceRepository struct {
	db *sql.DB
}

func NewDeviceRepository(conn *sql.DB) DeviceRepository {
	return DeviceRepository{db: conn}
}

func (r DeviceRepository) CreateDevice(ctx context.Context, params devices.CreateDeviceParams) (devices.Device, error) {
	if r.db == nil {
		return devices.Device{}, fmt.Errorf("database is required")
	}

	id, err := discovery.NewID()
	if err != nil {
		return devices.Device{}, err
	}

	result, err := scanDevice(r.db.QueryRowContext(ctx, `
INSERT INTO devices (id, name, management_address, platform, vendor, model)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, management_address, platform, vendor, model, created_at, updated_at
`,
		id,
		params.Name,
		params.ManagementAddress,
		params.Platform,
		params.Vendor,
		params.Model,
	))
	if err != nil {
		return devices.Device{}, fmt.Errorf("create device: %w", err)
	}

	return result, nil
}

func (r DeviceRepository) ListDevices(ctx context.Context) ([]devices.Device, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, name, management_address, platform, vendor, model, created_at, updated_at
FROM devices
ORDER BY name ASC, id ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var results []devices.Device
	for rows.Next() {
		item, err := scanDevice(rows)
		if err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read devices: %w", err)
	}

	return results, nil
}

type deviceScanner interface {
	Scan(dest ...any) error
}

func scanDevice(scanner deviceScanner) (devices.Device, error) {
	var item devices.Device
	var managementAddress sql.NullString
	var platform sql.NullString
	var vendor sql.NullString
	var model sql.NullString

	err := scanner.Scan(
		&item.ID,
		&item.Name,
		&managementAddress,
		&platform,
		&vendor,
		&model,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return devices.Device{}, devices.ErrNotFound
	}
	if err != nil {
		return devices.Device{}, err
	}

	item.ManagementAddress = nullableString(managementAddress)
	item.Platform = nullableString(platform)
	item.Vendor = nullableString(vendor)
	item.Model = nullableString(model)
	return item, nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
