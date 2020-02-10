CREATE TABLE template (
  template_id VARCHAR PRIMARY KEY,
  type VARCHAR NOT NULL,
  version INTEGER NOT NULL,
  schema JSONB NOT NULL,
  command_template TEXT NOT NULL,
  image VARCHAR NOT NULL,
  memory INTEGER NOT NULL,
  gpu INTEGER NOT NULL,
  cpu INTEGER NOT NULL,
  env JSONB,
  privileged BOOLEAN,
  adaptive_resource_allocation BOOLEAN,
  container_name VARCHAR NOT NULL,
  CONSTRAINT template_type_version UNIQUE(type, version)
);

ALTER TABLE task
  ADD COLUMN IF NOT EXISTS executable_request_custom JSONB;
