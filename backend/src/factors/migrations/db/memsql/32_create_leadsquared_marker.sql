CREATE TABLE IF NOT EXISTS leadsquared_markers(
    project_id bigint NOT NULL,
    delta bigint NOT NULL,
    document text NOT NULL,
    no_of_retries int,
    index_number int,
    tag text,
    created_at timestamp(6),
    updated_at timestamp(6)
);