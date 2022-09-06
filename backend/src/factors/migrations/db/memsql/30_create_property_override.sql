CREATE TABLE IF NOT EXISTS property_overrides (
    project_id bigint NOT NULL,
    property_name text NOT NULL,
    override_type int NOT NULL,
    entity integer NOT NULL, 
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id)
);