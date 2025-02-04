
CREATE TABLE IF NOT EXISTS dash_query_results (
    id text,
    project_id bigint,
    dashboard_id bigint,
    dashboard_unit_id bigint,
    query_id bigint,
    from_t bigint,
    to_t bigint,
    result LONGBLOB,
    computed_at bigint,
    is_deleted boolean NOT NULL DEFAULT FALSE,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    KEY (project_id, id) USING CLUSTERED COLUMNSTORE,
    PRIMARY KEY (project_id, query_id, id)
    );


