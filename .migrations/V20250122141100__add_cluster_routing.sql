CREATE TYPE cluster_status AS ENUM ('active', 'maintenance', 'offline');
CREATE TYPE tier AS ENUM ('t0', 't1', 't2', 't3', 't4');

CREATE TABLE cluster_state (
                               name VARCHAR PRIMARY KEY,
                               status cluster_status NOT NULL DEFAULT 'active',
                               status_reason VARCHAR,
                               status_since TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                               capabilities JSONB NOT NULL DEFAULT '[]'::jsonb,
                               allowed_tiers tier[] NOT NULL,
                               region VARCHAR NOT NULL,
                               updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX ix_cluster_state_status ON cluster_state(status);