CREATE TABLE IF NOT EXISTS property_mappings (
    id text NOT NULL,
    project_id bigint NOT NULL,
    name text NOT NULL, 
    display_name text NOT NULL,
    section_bit_map int NOT NULL,
    data_type text NOT NULL,
    properties json,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    is_deleted boolean NOT NULL DEFAULT false,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id),
    UNIQUE KEY unique_property_mappings_project_id_name_idx(project_id, name) USING HASH
);