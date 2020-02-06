ALTER TABLE task
  ADD COLUMN executable_id VARCHAR,
  ADD COLUMN executable_type VARCHAR DEFAULT 'task_definition';