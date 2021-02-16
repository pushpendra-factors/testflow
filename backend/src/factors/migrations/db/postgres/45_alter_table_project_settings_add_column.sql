--UP
ALTER TABLE project_settings
ADD COLUMN int_linkedin_ad_account text,
ADD COLUMN int_linkedin_access_token text,
ADD COLUMN int_linkedin_access_token_expiry bigint,
ADD COLUMN int_linkedin_refresh_token text,
ADD COLUMN int_linkedin_refresh_token_expiry bigint,
ADD COLUMN int_linkedin_agent_uuid text;
ALTER TABLE project_settings ADD CONSTRAINT project_settings_linkedin_agent_uuid_foreign FOREIGN KEY (int_linkedin_agent_uuid) REFERENCES agents (uuid);

--DOWN
-- ALTER TABLE project_settings
-- DROP COLUMN int_linkedin_ad_account,
-- DROP COLUMN int_linkedin_access_token,
-- DROP COLUMN int_linkedin_agent_uuid,
-- DROP COLUMN int_linkedin_access_token_expiry,
-- DROP COLUMN int_linkedin_refresh_token_expiry,
-- DROP COLUMN int_linkedin_refresh_token;