CREATE TABLE IF NOT EXISTS widget_groups (
    project_id BIGINT NOT NULL,
    id TEXT NOT NULL,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    is_non_comparable boolean DEFAULT false,
    widgets json,
    widgets_added boolean NOT NULL DEFAULT FALSE,
    created_at timestamp(6) DEFAULT '1970-01-01 00:00:00',
    updated_at timestamp(6) DEFAULT '1970-01-01 00:00:00',
    SHARD KEY (project_id),
    KEY (project_id, id) USING HASH,
    UNIQUE KEY unique_widget_groups_project_id_name_idx(project_id, display_name) USING HASH
)
