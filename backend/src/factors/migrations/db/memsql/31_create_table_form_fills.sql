CREATE TABLE IF NOT EXISTS form_fills(
    project_id bigint NOT NULL,
    id text NOT NULL,
    form_id text NOT NULL,
    value text,
    field_id text NOT NULL,
    first_updated_time bigint,
    last_updated_time bigint,
    time_spent_on_field bigint,
    created_at timestamp(6),
    updated_at timestamp(6),
    PRIMARY KEY (project_id, form_id, id),
    SHARD KEY (project_id, form_id)
);