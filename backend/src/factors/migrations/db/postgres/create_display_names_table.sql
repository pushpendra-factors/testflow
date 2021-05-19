-- Table: public.display_names

CREATE TABLE public.display_names
(
    project_id bigint NOT NULL,
    event_name text NULL,
    property_name text NULL,
    entity_type integer NOT NULL, -- There are 4 types Event/EventProperty/UserProperty/Object Property(hubspot/salesforce)
    display_name text NOT NULL,
    tag text NOT NULL,
    group_name text NOT NULL,
    group_object_name text NOT NULL, -- This is to avoid having same property duplicated multiple times at event property/ user property level which actually is the same name ultimately
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

ALTER TABLE public.display_names OWNER to postgres;

ALTER TABLE display_names ADD CONSTRAINT display_names_project_id_projects_id_foreign FOREIGN KEY (project_id) REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT;
CREATE UNIQUE INDEX display_names_project_id_event_name_property_name_tag_unique_idx on display_names(project_id, event_name, property_name, tag);
CREATE UNIQUE INDEX display_names_project_id_object_group_entity_tag_unique_idx on display_names(project_id, entity_type, group_name, group_object_name, display_name);
CREATE INDEX display_names_updated_at ON display_names(updated_at ASC);
-- DROP TABLE public.display_names;