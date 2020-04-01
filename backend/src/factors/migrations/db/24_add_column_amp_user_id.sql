-- UP
ALTER TABLE users ADD amp_user_id text;
CREATE UNIQUE INDEX users_project_id_amp_user_idx ON users(project_id, amp_user_id);

-- DOWN
-- ALTER TABLE users DROP amp_user_id;
-- DROP INDEX users_project_id_amp_user_idx;