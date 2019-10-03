-- UP
ALTER TABLE project_settings ADD COLUMN exclude_bot boolean;
UPDATE project_settings SET exclude_bot=true;

-- DOWN
-- ALTER TABLE project_settings DROP COLUMN exclude_bot RESTRICT;