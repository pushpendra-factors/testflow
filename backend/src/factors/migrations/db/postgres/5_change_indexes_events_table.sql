-- UP
CREATE INDEX project_id_event_name_id_user_id_timestamp_idx ON events (project_id, event_name_id, user_id, timestamp DESC);
DROP INDEX project_id_user_id_timestamp_idx;

-- DOWN
-- DROP INDEX project_id_event_name_id_user_id_timestamp_idx;
-- CREATE INDEX project_id_user_id_timestamp_idx ON events (project_id, user_id, timestamp DESC);