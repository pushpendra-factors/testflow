-- UP
ALTER TABLE project_settings ADD COLUMN auto_form_capture boolean;
ALTER TABLE project_settings ALTER COLUMN auto_form_capture SET DEFAULT false;

-- DOWN
-- ALTER TABLE project_settings DROP COLUMN auto_form_capture RESTRICT;