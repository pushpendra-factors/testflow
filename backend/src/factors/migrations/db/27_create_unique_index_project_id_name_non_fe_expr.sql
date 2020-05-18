-- UP
CREATE UNIQUE INDEX event_names_project_id_name_type_not_fe_uindx ON event_names(project_id, name, type) WHERE type != 'FE';

-- DOWN
-- DROP INDEX event_names_project_id_name_type_not_fe_uindx;