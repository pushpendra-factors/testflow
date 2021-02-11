-- SEQUENCE: public.dashboards_id_seq

-- DROP SEQUENCE public.dashboards_id_seq;

CREATE SEQUENCE public.dashboards_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.dashboards_id_seq
    OWNER TO autometa;

-- SEQUENCE: public.dashboards_project_id_seq

-- DROP SEQUENCE public.dashboards_project_id_seq;

CREATE SEQUENCE public.dashboards_project_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.dashboards_project_id_seq
    OWNER TO autometa;

-- Table: public.dashboards

-- DROP TABLE public.dashboards;

CREATE TABLE public.dashboards
(
    id bigint NOT NULL DEFAULT nextval('dashboards_id_seq'::regclass),
    project_id bigint NOT NULL DEFAULT nextval('dashboards_project_id_seq'::regclass),
    agent_uuid text COLLATE pg_catalog."default" NOT NULL,
    name text COLLATE pg_catalog."default" NOT NULL,
    type character varying(5) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT dashboards_pkey PRIMARY KEY (id, project_id, agent_uuid),
    CONSTRAINT dashboards_project_id_projects_id_foreign FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.dashboards
    OWNER to autometa;

-- Index: project_id_id_unique_idx

-- DROP INDEX public.project_id_id_unique_idx;

CREATE UNIQUE INDEX project_id_id_unique_idx
    ON public.dashboards USING btree
    (project_id, id)
    TABLESPACE pg_default;