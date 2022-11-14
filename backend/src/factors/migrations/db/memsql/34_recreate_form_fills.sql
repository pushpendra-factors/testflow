-- UP
DROP table form_fills;
CREATE TABLE IF NOT EXISTS form_fills(
    project_id bigint NOT NULL,
    id text NOT NULL,
    user_id text NOT NULL,
    form_id text NOT NULL,
    field_id text NOT NULL,
    value text,
    created_at timestamp(6),
    updated_at timestamp(6),
    PRIMARY KEY (project_id, user_id, form_id, id),
    SHARD KEY (project_id, user_id, form_id)
);

