ALTER TABLE project_settings ADD COLUMN int_facebook_token_expiry bigint;
--Down
-- alter table project_settings drop column int_facebook_token_expiry;
