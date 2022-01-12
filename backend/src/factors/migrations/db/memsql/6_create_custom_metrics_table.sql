CREATE ROWSTORE TABLE IF NOT EXISTS custom_metrics(
    project_id bigint NOT NULL,
    id text NOT NULL,
    name text NOT NULL,
    description text,
    type_of_query int,  -- represents if kpi-profiles
    object_type text, -- represents if hubspot_contact ...
    transformations json,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id),
    UNIQUE KEY unique_custom_metrics_project_id_name_idx(project_id, name) USING HASH
)

