CREATE INDEX IF NOT EXISTS ix_task_executable_id ON task(executable_id);
CREATE INDEX IF NOT EXISTS ix_task_executable_id_started_at_desc ON task(executable_id, started_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS ix_task_executable_id_started_at_desc_engine ON task(executable_id, started_at DESC NULLS LAST, engine);
