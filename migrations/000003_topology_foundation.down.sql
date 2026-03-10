DROP INDEX IF EXISTS idx_links_z_interface;
DROP INDEX IF EXISTS idx_links_a_interface;
DROP INDEX IF EXISTS idx_interfaces_device;
DROP INDEX IF EXISTS idx_devices_platform;
DROP INDEX IF EXISTS idx_devices_vendor;

ALTER TABLE links DROP CONSTRAINT IF EXISTS chk_links_distinct_endpoints;
ALTER TABLE links ALTER COLUMN z_interface_id DROP NOT NULL;
ALTER TABLE links ALTER COLUMN a_interface_id DROP NOT NULL;

ALTER TABLE interfaces DROP CONSTRAINT IF EXISTS uq_interfaces_device_name;

ALTER TABLE platforms ALTER COLUMN vendor_id DROP NOT NULL;
