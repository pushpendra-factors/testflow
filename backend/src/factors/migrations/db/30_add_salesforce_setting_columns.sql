-- UP
ALTER TABLE project_settings ADD int_salesforce_enabled_agent_uuid text;
ALTER TABLE agents ADD int_salesforce_refresh_token text;
ALTER TABLE agents ADD int_salesforce_instance_url text;
ALTER TABLE project_settings ADD CONSTRAINT project_settings_salesforce_enabled_agent_uuid_foreign FOREIGN KEY (int_salesforce_enabled_agent_uuid) REFERENCES agents (uuid);


-- DOWN
-- ALTER TABLE project_settings DROP int_salesforce_enabled_agent_uuid;
-- ALTER TABLE agents DROP int_salesforce_refresh_token text;
-- ALTER TABLE agents DROP int_salesforce_instance_url text;
-- ALTER TABLE project_settings DROP CONSTRAINT project_settings_salesforce_enabled_agent_uuid_foreign;

