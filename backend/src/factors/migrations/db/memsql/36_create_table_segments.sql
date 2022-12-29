CREATE TABLE IF NOT EXISTS segments(
    id text NOT NULL,
    project_id bigint NOT NULL,
    name text NOT NULL, 
    description text, 
    query json,
    type text,
    PRIMARY KEY (project_id, id),
    SHARD KEY (project_id, id)
);