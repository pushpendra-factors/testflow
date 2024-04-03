CREATE TABLE IF NOT EXISTS widget_groups (
    project_id BIGINT NOT NULL,
    id TEXT NOT NULL,
    display_name TEXT NOT NULL,
    widgets json,
    widgets_added boolean NOT NULL DEFAULT FALSE,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    KEY (project_id, id) USING HASH,
    UNIQUE KEY unique_widget_groups_project_id_name_idx(project_id, display_name) USING HASH
);