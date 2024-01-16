CREATE TABLE IF NOT EXISTS slack_users_list(
    project_id BIGINT NOT NULL, 
    agent_id TEXT NOT NULL,
    users_list JSON,
    last_sync_time TIMESTAMP(6)
)