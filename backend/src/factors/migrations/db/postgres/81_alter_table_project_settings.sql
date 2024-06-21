ALTER TABLE project_settings ADD COLUMN sso_state int default 0;
ALTER TABLE project_settings DROP COLUMN saml_enabled;