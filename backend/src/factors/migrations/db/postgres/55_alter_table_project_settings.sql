-- UP
ALTER TABLE project_settings ADD COLUMN int_google_organic_enabled_agent_uuid text;
ALTER TABLE project_settings ADD COLUMN int_google_organic_url_prefixes text;

-- DOWN
-- ALTER TABLE project_settings DROP COLUMN int_google_organic_enabled_agent_uuid;
-- ALTER TABLE project_settings DROP COLUMN int_google_organic_url_prefixes;