CREATE TABLE IF NOT EXISTS project_plan_mappings (
    project_id bigint,
    plan_id bigint NOT NULL,
    over_write json,
    last_renewed_on timestamp(6),
    PRIMARY KEY (project_id),
    SHARD KEY (project_id)
);