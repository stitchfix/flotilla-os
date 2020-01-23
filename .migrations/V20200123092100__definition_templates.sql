--
-- Task Types
--
CREATE TABLE IF NOT EXISTS definition_template (
  id VARCHAR PRIMARY KEY,
  type VARCHAR NOT NULL,
  version INTEGER NOT NULL,
  schema JSONB NOT NULL,
  template TEXT NOT NULL,
  image VARCHAR NOT NULL,
  CONSTRAINT definition_template_type_version UNIQUE(type, version)
);

ALTER TABLE task_def
  DROP COLUMN task_type,
  ADD COLUMN template_id character varying REFERENCES definition_template(id),
  ADD COLUMN template_payload jsonb;

ALTER TABLE task
  DROP COLUMN task_type,
  ADD COLUMN template_id character varying REFERENCES definition_template(id),
  ADD COLUMN template_payload jsonb;
