-- UP 
ALTER TABLE project_settings ADD COLUMN int_facebook_ingestion_timezone text;

-- DOWN 
-- ALTER TABLE project_settings DROP COLUMN int_facebook_ingestion_timezone;