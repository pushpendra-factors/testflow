-- UP
ALTER TABLE events ADD COLUMN session_id text;

-- DOWN
-- ALTER TABLE events DROP COLUMN session_id RESTRICT;