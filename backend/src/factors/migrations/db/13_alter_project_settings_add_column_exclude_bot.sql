-- UP
ALTER TABLE project_settings ADD COLUMN exclude_bot boolean;

UPDATE project_settings SET exclude_bot=true;
-- ADDED NOT NULL AND DEFAULT
ALTER TABLE project_settings ALTER COLUMN exclude_bot SET NOT NULL;
ALTER TABLE project_settings ALTER COLUMN exclude_bot SET DEFAULT false;

-- DOWN
-- ALTER TABLE project_settings DROP COLUMN exclude_bot RESTRICT;