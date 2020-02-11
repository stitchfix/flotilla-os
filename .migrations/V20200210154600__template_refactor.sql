ALTER TABLE template DROP CONSTRAINT template_type_version;
ALTER TABLE template RENAME COLUMN type to template_name;
ALTER TABLE template ADD CONSTRAINT template_name_version UNIQUE(template_name, version);