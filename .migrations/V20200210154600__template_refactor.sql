ALTER TABLE template
  DROP CONSTRAINT template_type_version
  RENAME COLUMN type to template_name
  ADD CONSTRAINT template_name_version UNIQUE(template_name, version);