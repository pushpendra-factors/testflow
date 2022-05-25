CREATE TABLE IF NOT EXISTS crm_settings (
    project_id bigint NOT NULL,
    hubspot_enrich_heavy boolean NOT NULL DEFAULT FALSE,
    PRIMARY KEY (project_id),
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
);