CREATE TABLE IF NOT EXISTS data_availabilities (
    project_id bigint NOT NULL,
    integration text,
    latest_data_timestamp bigint,
    last_polled timestamp(6) NOT NULL, 
    source text,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    UNIQUE KEY project_id_integration_unique_idx(project_id,integration) USING HASH
);