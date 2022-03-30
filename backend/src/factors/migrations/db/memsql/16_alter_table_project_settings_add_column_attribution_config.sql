-- Add attribution_config column in project_settings table

ALTER TABLE project_settings ADD COLUMN attribution_config JSON;
-- ALTER TABLE project_settings DROP COLUMN attribution_config;