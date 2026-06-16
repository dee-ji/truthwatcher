DROP INDEX IF EXISTS devices_name_unique_idx;
DROP INDEX IF EXISTS devices_management_address_idx;
DROP INDEX IF EXISTS devices_platform_idx;

ALTER TABLE devices RENAME COLUMN name TO hostname;
ALTER TABLE devices RENAME COLUMN management_address TO management_ip;
ALTER TABLE devices DROP COLUMN IF EXISTS platform;
ALTER TABLE devices ADD COLUMN serial_number TEXT;
ALTER TABLE devices ADD COLUMN role TEXT;
ALTER TABLE devices ADD COLUMN site TEXT;

CREATE UNIQUE INDEX devices_hostname_unique_idx ON devices (lower(hostname));
CREATE INDEX devices_management_ip_idx ON devices (management_ip) WHERE management_ip IS NOT NULL AND management_ip <> '';
CREATE INDEX devices_serial_number_idx ON devices (serial_number) WHERE serial_number IS NOT NULL AND serial_number <> '';
CREATE INDEX devices_role_idx ON devices (role) WHERE role IS NOT NULL AND role <> '';
CREATE INDEX devices_site_idx ON devices (site) WHERE site IS NOT NULL AND site <> '';
