DROP INDEX project_id_user_id_timestamp_idx;
CREATE INDEX project_id_user_id_timestamp_idx ON events (project_id, user_id, timestamp DESC);