CREATE TABLE IF NOT EXISTS project_plan_mapping (
    project_id bigint,
    plan_id bigint NOT NULL,
    add_ons json,
    last_renewed_on timestamp(6),
    PRIMARY KEY (project_id),
    SHARD KEY (project_id)
);