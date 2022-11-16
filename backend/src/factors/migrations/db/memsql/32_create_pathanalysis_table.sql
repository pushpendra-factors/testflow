CREATE TABLE IF NOT EXISTS pathanalysis(
    id text NOT NULL,
    project_id bigint NOT NULL,
    title text,
    status text,
    created_by text,
    query json,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    is_deleted boolean NOT NULL DEFAULT FALSE,
    PRIMARY KEY (project_id, id)
);