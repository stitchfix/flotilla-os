DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tier') THEN
CREATE TYPE tier AS ENUM ('Tier1', 'Tier2', 'Tier3', 'Tier4');
END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'cluster_status') THEN
CREATE TYPE cluster_status AS ENUM ('active', 'maintenance', 'offline');
END IF;
END$$;

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

CREATE INDEX IF NOT EXISTS ix_cluster_state_status ON cluster_state(status);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1
        FROM information_schema.columns
        WHERE table_name='task' AND column_name='tier')
    THEN
        ALTER TABLE task ADD COLUMN tier tier DEFAULT 'Tier4';
    END IF;
END$$;