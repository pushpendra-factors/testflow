CREATE TABLE public.feedbacks(
    id text NOT NULL DEFAULT uuid_generate_v4(),
    project_id bigint NOT NULL,
    feature text NOT NULL,
    property jsonb,
    vote_type int NOT NULL,
    created_by text ,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT feedbacks_pkey PRIMARY KEY(id),
    CONSTRAINT feedbacks_project_id_projects_id_foreign FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.feedbacks OWNER to postgres;

-- INDEX
CREATE INDEX feedbacks_updated_at ON feedbacks(updated_at ASC);