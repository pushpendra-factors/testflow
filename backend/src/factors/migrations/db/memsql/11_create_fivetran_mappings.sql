CREATE ROWSTORE TABLE IF NOT EXISTS fivetran_mappings(
    project_id bigint NOT NULL,
    id text NOT NULL,
    integration text NOT NULL,
    connector_id text NOT NULL,
    schema_id text NOT NULL,
    accounts text NOT NULL,
    status boolean,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id)
);

