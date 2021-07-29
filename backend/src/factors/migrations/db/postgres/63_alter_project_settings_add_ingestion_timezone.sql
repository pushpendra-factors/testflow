-- UP 
ALTER TABLE project_settings ADD COLUMN int_google_ingestion_timezone text;

-- DOWN 
-- ALTER TABLE project_settings DROP COLUMN int_google_ingestion_timezone;