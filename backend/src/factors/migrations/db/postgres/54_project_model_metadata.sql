-- Table: public.project_model_metadata

CREATE TABLE public.project_model_metadata
(
    project_id bigint NOT NULL,
    model_id bigint NOT NULL,
    model_type text NOT NULL,
    start_time  bigint NOT NULL, 
    end_time bigint NOT NULL,
    chunks text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

ALTER TABLE public.project_model_metadata OWNER to postgres;

ALTER TABLE project_model_metadata ADD CONSTRAINT project_model_metadata_project_id_projects_id_foreign FOREIGN KEY (project_id) REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT;

CREATE INDEX project_model_metadata_updated_at ON project_model_metadata(updated_at ASC);
CREATE INDEX project_model_metadata_project_id ON project_model_metadata(project_id ASC);
CREATE UNIQUE INDEX project_model_metadata_project_id_stdate_enddate_unique_idx on project_model_metadata(project_id, start_time, end_time);
-- DROP TABLE public.project_model_metadata;

-- UP
ALTER TABLE project_model_metadata ADD COLUMN id TEXT NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE project_model_metadata ADD PRIMARY KEY (id, project_id);

-- DOWN
-- ALTER TABLE project_model_metadata DROP CONSTRAINT display_names_pkey;
-- ALTER TABLE project_model_metadata DROP COLUMN id;