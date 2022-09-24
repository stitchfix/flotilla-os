ALTER TABLE task ADD COLUMN IF NOT EXISTS idempotence_key varchar;
