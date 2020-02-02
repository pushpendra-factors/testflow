-- UP
ALTER TABLE project_settings ADD int_hubspot boolean NOT NULL DEFAULT 'false';
ALTER TABLE project_settings ADD int_hubspot_api_key text;

-- DOWN
-- ALTER TABLE project_settings DROP int_hubspot;
-- ALTER TABLE project_settings DROP int_hubspot_api_key;

