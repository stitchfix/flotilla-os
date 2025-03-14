DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'cluster_status') THEN
        CREATE TYPE cluster_status AS ENUM ('active', 'maintenance', 'offline');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS cluster_state (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR NOT NULL,
    cluster_version VARCHAR NOT NULL DEFAULT '',
    status cluster_status NOT NULL DEFAULT 'active',
    status_reason VARCHAR,
    status_since TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    capabilities VARCHAR[] NOT NULL DEFAULT '{}',
    allowed_tiers VARCHAR[] NOT NULL DEFAULT '{}',
    region VARCHAR NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    namespace VARCHAR NOT NULL DEFAULT '',
    emr_virtual_cluster VARCHAR NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS ix_cluster_state_name ON cluster_state(name);
CREATE INDEX IF NOT EXISTS ix_cluster_state_status ON cluster_state(status);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1
        FROM information_schema.columns
        WHERE table_name='task' AND column_name='tier')
    THEN
ALTER TABLE task ADD COLUMN tier VARCHAR(20);
END IF;
END$$;