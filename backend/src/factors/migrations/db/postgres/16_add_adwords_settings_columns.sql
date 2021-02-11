-- UP
ALTER TABLE project_settings ADD int_adwords_enabled_agent_uuid text;
ALTER TABLE project_settings ADD int_adwords_customer_account_id text;
ALTER TABLE agents ADD int_adwords_refresh_token text;
ALTER TABLE project_settings ADD CONSTRAINT project_settings_adwords_agent_uuid_foreign FOREIGN KEY (int_adwords_enabled_agent_uuid) REFERENCES agents (uuid);

-- DOWN
-- ALTER TABLE project_settings DROP int_adwords_enabled_agent_id;
-- ALTER TABLE project_settings DROP int_adwords_customer_account_id;
-- ALTER TABLE agents ADD int_adwords_refresh_token text;

