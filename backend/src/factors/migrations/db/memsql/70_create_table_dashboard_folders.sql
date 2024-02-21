CREATE TABLE IF NOT EXISTS dashboard_folders(
    id TEXT NOT NULL,
    name TEXT NOT NULL,
    project_id BIGINT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    is_default_folder BOOLEAN NOT NULL DEFAULT FALSE,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (id) USING HASH,
    SHARD KEY (id),
    PRIMARY KEY (id, project_id)
);