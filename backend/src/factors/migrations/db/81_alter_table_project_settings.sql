ALTER TABLE project_settings ADD COLUMN sso_state int default 1;
ALTER TABLE project_settings DROP COLUMN saml_enabled;