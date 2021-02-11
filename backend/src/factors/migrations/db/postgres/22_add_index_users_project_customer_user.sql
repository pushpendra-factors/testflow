-- UP
CREATE INDEX users_project_id_customer_user_id_idx ON users(project_id, customer_user_id);

-- DOWN
-- DROP INDEX users_project_id_customer_user_id_idx;