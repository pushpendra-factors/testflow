CREATE TABLE IF NOT EXISTS workflows (
    id TEXT NOT NULL,
    project_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    alert_body JSON,
    created_at timestamp(6) DEFAULT '1970-01-01 00:00:00',
    updated_at timestamp(6) DEFAULT '1970-01-01 00:00:00',
    created_by TEXT,
    is_deleted BOOLEAN DEFAULT FALSE,
    SHARD KEY (project_id),
    KEY (project_id, id) USING HASH,
);