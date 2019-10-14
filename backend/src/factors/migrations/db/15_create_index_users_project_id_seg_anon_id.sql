-- UP
CREATE INDEX users_project_id_segment_anonymous_id_idx ON users(project_id, segment_anonymous_id);

-- DOWN
-- DROP INDEX users_project_id_segment_anonymous_id_idx;