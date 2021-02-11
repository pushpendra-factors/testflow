-- UP
CREATE INDEX project_id_event_name_id_timestamp_idx ON events(project_id, event_name_id, timestamp DESC);
CREATE INDEX project_id_event_name_id_user_id_idx ON events(project_id, event_name_id, user_id);
DROP INDEX project_id_event_name_id_user_id_timestamp_idx;

-- DOWN
-- DROP INDEX project_id_event_name_id_timestamp_idx;
-- DROP INDEX project_id_event_name_id_user_id_idx;
-- CREATE INDEX project_id_event_name_id_user_id_timestamp_idx ON events (project_id, event_name_id, user_id, timestamp DESC);
