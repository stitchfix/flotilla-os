ALTER TABLE template
  RENAME COLUMN type to template_name
  DROP CONSTRAINT template_type_version
  ADD CONSTRAINT template_name_version UNIQUE(template_name, version);