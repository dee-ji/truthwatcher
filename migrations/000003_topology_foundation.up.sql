ALTER TABLE platforms ALTER COLUMN vendor_id SET NOT NULL;

ALTER TABLE interfaces
  ADD CONSTRAINT uq_interfaces_device_name UNIQUE (device_id, name);

ALTER TABLE links
  ALTER COLUMN a_interface_id SET NOT NULL,
  ALTER COLUMN z_interface_id SET NOT NULL,
  ADD CONSTRAINT chk_links_distinct_endpoints CHECK (a_interface_id <> z_interface_id);

CREATE INDEX IF NOT EXISTS idx_devices_vendor ON devices(vendor_id);
CREATE INDEX IF NOT EXISTS idx_devices_platform ON devices(platform_id);
CREATE INDEX IF NOT EXISTS idx_interfaces_device ON interfaces(device_id);
CREATE INDEX IF NOT EXISTS idx_links_a_interface ON links(a_interface_id);
CREATE INDEX IF NOT EXISTS idx_links_z_interface ON links(z_interface_id);
