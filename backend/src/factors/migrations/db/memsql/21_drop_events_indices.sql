-- UP
DROP INDEX event_name_id_idx ON events;
DROP INDEX project_id_idx ON events;
DROP INDEX timestamp_idx ON events;

-- DOWN
-- CREATE INDEX event_name_id_idx ON events(event_name_id) USING HASH;
-- CREATE INDEX project_id_idx ON events(project_id) USING HASH;
-- CREATE INDEX timestamp_idx ON events(timestamp) USING HASH;