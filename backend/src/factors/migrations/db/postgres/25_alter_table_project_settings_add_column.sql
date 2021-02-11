--UP
ALTER TABLE project_settings
ADD COLUMN int_facebook_email text,
ADD COLUMN int_facebook_access_token text,
ADD COLUMN int_facebook_agent_uuid text,
ADD COLUMN int_facebook_user_id text,
ADD COLUMN int_facebook_ad_account text;
ALTER TABLE project_settings ADD CONSTRAINT project_settings_facebook_agent_uuid_foreign FOREIGN KEY (int_facebook_agent_uuid) REFERENCES agents (uuid);

--DOWN
-- ALTER TABLE project_settings
-- DROP COLUMN int_facebook_email,
-- DROP COLUMN int_facebook_access_token,
-- DROP COLUMN int_facebook_agent_uuid,
-- DROP COLUMN int_facebook_user_id,
-- DROP COLUMN int_facebook_ad_account;