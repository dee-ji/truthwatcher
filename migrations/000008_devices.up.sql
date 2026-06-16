CREATE TABLE devices (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    management_address TEXT,
    platform TEXT,
    vendor TEXT,
    model TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX devices_name_unique_idx ON devices (lower(name));
CREATE INDEX devices_management_address_idx ON devices (management_address) WHERE management_address IS NOT NULL AND management_address <> '';
CREATE INDEX devices_platform_idx ON devices (platform) WHERE platform IS NOT NULL AND platform <> '';
CREATE INDEX devices_vendor_idx ON devices (vendor) WHERE vendor IS NOT NULL AND vendor <> '';
