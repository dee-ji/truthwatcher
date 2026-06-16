DROP INDEX IF EXISTS devices_site_idx;
DROP INDEX IF EXISTS devices_role_idx;
DROP INDEX IF EXISTS devices_serial_number_idx;
DROP INDEX IF EXISTS devices_management_ip_idx;
DROP INDEX IF EXISTS devices_hostname_unique_idx;

ALTER TABLE devices ADD COLUMN platform TEXT;
ALTER TABLE devices DROP COLUMN IF EXISTS site;
ALTER TABLE devices DROP COLUMN IF EXISTS role;
ALTER TABLE devices DROP COLUMN IF EXISTS serial_number;
ALTER TABLE devices RENAME COLUMN management_ip TO management_address;
ALTER TABLE devices RENAME COLUMN hostname TO name;

CREATE UNIQUE INDEX devices_name_unique_idx ON devices (lower(name));
CREATE INDEX devices_management_address_idx ON devices (management_address) WHERE management_address IS NOT NULL AND management_address <> '';
CREATE INDEX devices_platform_idx ON devices (platform) WHERE platform IS NOT NULL AND platform <> '';
