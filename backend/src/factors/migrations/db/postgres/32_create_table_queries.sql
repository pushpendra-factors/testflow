
--UP
CREATE TABLE public.queries (
    id SERIAL,
    project_id bigint NOT NULL,
    created_by character varying(255),
    title text ,
    query jsonb,
    type int,
    is_deleted boolean NOT NULL DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT queries_pkey PRIMARY KEY (project_id, id, type),

    CONSTRAINT queries_project_id_projects_id_foreign FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.queries OWNER to postgres;
-- DOWN
-- DROP TABLE public.queries;

-- Add query_id column in dashboard_units table

ALTER TABLE public.dashboard_units ADD COLUMN query_id bigint;

--DOWN
-- ALTER TABLE public.dashboard_units DROP COLUMN query_id;