-- Table: reports

-- UP
ALTER TABLE reports ADD COLUMN updated_at timestamp with time zone;
-- Fill with initial values required.
UPDATE reports SET updated_at=created_at WHERE created_at IS NOT NULL AND updated_at IS NULL;

-- DOWN
-- ALTER TABLE reports DROP COLUMN updated_at;
