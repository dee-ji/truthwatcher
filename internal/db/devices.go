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
INSERT INTO devices (id, hostname, vendor, model, serial_number, management_ip, role, site)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, hostname, vendor, model, serial_number, management_ip, role, site, created_at, updated_at
`,
		id,
		params.Hostname,
		params.Vendor,
		params.Model,
		params.SerialNumber,
		params.ManagementIP,
		params.Role,
		params.Site,
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
SELECT id, hostname, vendor, model, serial_number, management_ip, role, site, created_at, updated_at
FROM devices
ORDER BY hostname ASC, id ASC
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
	var vendor sql.NullString
	var model sql.NullString
	var serialNumber sql.NullString
	var managementIP sql.NullString
	var role sql.NullString
	var site sql.NullString

	err := scanner.Scan(
		&item.ID,
		&item.Hostname,
		&vendor,
		&model,
		&serialNumber,
		&managementIP,
		&role,
		&site,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return devices.Device{}, devices.ErrNotFound
	}
	if err != nil {
		return devices.Device{}, err
	}

	item.Vendor = nullableString(vendor)
	item.Model = nullableString(model)
	item.SerialNumber = nullableString(serialNumber)
	item.ManagementIP = nullableString(managementIP)
	item.Role = nullableString(role)
	item.Site = nullableString(site)
	return item, nil
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
