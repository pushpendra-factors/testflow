-- UP
ALTER TABLE projects ADD COLUMN jobs_metadata JSONB;

-- DOWN
-- ALTER TABLE projects DROP COLUMN jobs_metadata;