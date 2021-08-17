-- UP
ALTER TABLE event_names ALTER COLUMN id TYPE bigint USING id::bigint;
ALTER TABLE events ALTER COLUMN event_name_id TYPE bigint USING event_name_id::bigint;

-- DOWN
-- ALTER TABLE event_names ALTER COLUMN id TYPE text USING id::text;
-- ALTER TABLE events ALTER COLUMN event_name_id TYPE text USING event_name_id::text;
