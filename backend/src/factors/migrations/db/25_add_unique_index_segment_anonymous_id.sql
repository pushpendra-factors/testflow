-- UP
CREATE UNIQUE INDEX users_project_id_segment_anonymous_uidx ON users(project_id, segment_anonymous_id);
DROP INDEX users_project_id_segment_anonymous_id_idx;

-- DOWN
-- DROP INDEX users_project_id_segment_anonymous_idx;
-- CREATE INDEX users_project_id_segment_anonymous_id_idx ON users(project_id, segment_anonymous_id);
