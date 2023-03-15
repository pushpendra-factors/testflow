CREATE TABLE IF NOT EXISTS display_name_labels (
    project_id bigint NOT NULL,
    id text NOT NULL,
    source text NOT NULL,
    property_key text NOT NULL,
    value text NOT NULL,
    label text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (project_id, source, property_key, value, id) USING CLUSTERED COLUMNSTORE,
    SHARD KEY (project_id, source, id),
    UNIQUE KEY(project_id, source, id, property_key, value) USING HASH
);