-- Table: public.property_details

CREATE TABLE public.property_details
(
    project_id bigint NOT NULL,
    event_name_id bigint null,
    "key" text  NOT NULL,
    "type" text  NOT NULL,
    entity integer  NOT NULL
);

ALTER TABLE public.property_details OWNER to postgres;

ALTER TABLE property_details ADD CONSTRAINT property_details_project_id_projects_id_foreign FOREIGN KEY (project_id) REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT;
ALTER TABLE property_details ADD CONSTRAINT property_details_event_name_id_event_name_id_foreign FOREIGN KEY (project_id, event_name_id) REFERENCES public.event_names (project_id, id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT;
CREATE UNIQUE INDEX property_details_project_id_event_name_id_key_unique_idx on property_details(project_id, event_name_id, "key");
-- DROP TABLE public.property_details;
