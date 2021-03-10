CREATE TABLE IF NOT EXISTS property_details (
    project_id bigint NOT NULL,
    event_name_id bigint NULL,
    `key` text NOT NULL,
    `type` text NOT NULL,
    entity integer NOT NULL,
    SHARD KEY (project_id),
    UNIQUE KEY property_details_project_id_event_name_id_key_unique_idx(project_id, event_name_id,`key`)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref.(project_id,event_name_id) -> event_names(project_id,id)
);