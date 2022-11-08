CREATE TABLE IF NOT EXISTS pathanalysis(
    id text NOT NULL,
    project_id bigint NOT NULL,
    title text,
    stat text,
    created_by text,
    quer json,
    created_on timestamp(6) NOT NULL,
    modified_on timestamp(6) NOT NULL,
    is_deleted boolean
);