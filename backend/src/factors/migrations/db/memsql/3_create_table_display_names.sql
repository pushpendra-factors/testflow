CREATE TABLE IF NOT EXISTS display_names (
    project_id bigint NOT NULL,
    event_name text NULL,
    property_name text NULL,
    entity_type integer NOT NULL,
    display_name text NOT NULL,
    tag text NOT NULL,
    group_name text NOT NULL,
    group_object_name text NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    UNIQUE KEY  display_names_project_id_event_name_property_name_tag_unique_idx(project_id, event_name, property_name, tag),
    UNIQUE KEY  display_names_project_id_object_group_entity_tag_unique_idx(project_id, group_name, entity_type, group_object_name, display_name)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
); 