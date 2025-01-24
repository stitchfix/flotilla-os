CREATE TYPE cluster_status AS ENUM ('active', 'maintenance', 'offline');
CREATE TYPE tier AS ENUM ('t0', 't1', 't2', 't3', 't4');

CREATE TABLE IF NOT EXISTS cluster_state (
    name VARCHAR PRIMARY KEY,
    status cluster_status NOT NULL DEFAULT 'active',
    status_reason VARCHAR,
    status_since TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    capabilities JSONB NOT NULL DEFAULT '[]'::jsonb,
    allowed_tiers JSONB NOT NULL DEFAULT '[]'::jsonb,
    region VARCHAR NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    namespace VARCHAR NOT NULL DEFAULT '',
    emr_virtual_cluster VARCHAR NOT NULL DEFAULT ''
);

CREATE INDEX ix_cluster_state_status ON cluster_state(status);