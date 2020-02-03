CREATE TABLE IF NOT EXISTS template_definition (
  id VARCHAR PRIMARY KEY,
  type VARCHAR NOT NULL,
  version INTEGER NOT NULL,
  schema JSONB NOT NULL,
  command_template TEXT NOT NULL,
  definition_id REFERENCES task_def(definition_id),
  CONSTRAINT definition_template_type_version UNIQUE(type, version)
);

CREATE TABLE IF NOT EXISTS template_run (
  template_definition_id VARCHAR REFERENCES template_definition(id),
  run_id VARCHAR REFERENCES task(run_id),
  template_arguments JSONB
);

ALTER TABLE task_def DROP COLUMN task_type;
ALTER TABLE task DROP COLUMN task_type;
