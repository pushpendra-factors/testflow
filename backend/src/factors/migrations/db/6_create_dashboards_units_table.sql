-- SEQUENCE: public.dashboard_units_id_seq

-- DROP SEQUENCE public.dashboard_units_id_seq;

CREATE SEQUENCE public.dashboard_units_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.dashboard_units_id_seq
    OWNER TO autometa;

-- SEQUENCE: public.dashboard_units_project_id_seq

-- DROP SEQUENCE public.dashboard_units_project_id_seq;

CREATE SEQUENCE public.dashboard_units_project_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.dashboard_units_project_id_seq
    OWNER TO autometa;

-- SEQUENCE: public.dashboard_units_dashboard_id_seq

-- DROP SEQUENCE public.dashboard_units_dashboard_id_seq;

CREATE SEQUENCE public.dashboard_units_dashboard_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.dashboard_units_dashboard_id_seq
    OWNER TO autometa;

-- Table: public.dashboard_units

-- DROP TABLE public.dashboard_units;

CREATE TABLE public.dashboard_units
(
    id bigint NOT NULL DEFAULT nextval('dashboard_units_id_seq'::regclass),
    project_id bigint NOT NULL DEFAULT nextval('dashboard_units_project_id_seq'::regclass),
    dashboard_id bigint NOT NULL DEFAULT nextval('dashboard_units_dashboard_id_seq'::regclass),
    title text COLLATE pg_catalog."default" NOT NULL,
    query jsonb NOT NULL,
    presentation character varying(5) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT dashboard_units_pkey PRIMARY KEY (id, project_id, dashboard_id),
    CONSTRAINT dashboard_units_project_id_dashboard_id_dashboards_project_id_i FOREIGN KEY (project_id, dashboard_id)
        REFERENCES public.dashboards (project_id, id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT dashboard_units_project_id_projects_id_foreign FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.dashboard_units
    OWNER to autometa;

