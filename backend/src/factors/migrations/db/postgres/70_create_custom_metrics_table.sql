CREATE TABLE IF NOT EXISTS custom_metrics (
    project_id bigint NOT NULL,
    id character varying(255) NOT NULL DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    description text,
    type_of_query int,
    object_type text, -- represents if hubspot_contact ...
    transformations json,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT custom_metrics_primary_key PRIMARY KEY (project_id, id)
);

CREATE UNIQUE INDEX custom_metrics_project_id_name_unique_idx on custom_metrics(project_id, id);