CREATE TABLE IF NOT EXISTS clickable_elements (
    project_id bigint NOT NULL,
    id text NOT NULL,
    display_name text NOT NULL,
    element_type text,
    element_attributes json,
    click_count int NOT NULL,
    enabled boolean DEFAULT false,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id, display_name, element_type),
    KEY (project_id, display_name, element_type) USING CLUSTERED COLUMNSTORE,
    UNIQUE KEY(project_id, display_name, element_type) USING HASH
);