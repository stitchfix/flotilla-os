ALTER TABLE task_def ADD COLUMN IF NOT EXISTS requires_docker BOOLEAN;
ALTER TABLE task ADD COLUMN IF NOT EXISTS ephemeral_storage BOOLEAN;
