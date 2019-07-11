-- SEQUENCE: public.reports_id_seq

-- DROP SEQUENCE public.reports_id_seq;

CREATE SEQUENCE public.reports_id_seq;

ALTER SEQUENCE public.reports_id_seq
    OWNER TO postgres;

-- Table: public.reports

-- DROP TABLE public.reports;

CREATE TABLE public.reports
(
    id bigint NOT NULL DEFAULT nextval('reports_id_seq'::regclass),
    project_id bigint,
    dashboard_id bigint,
    dashboard_name text COLLATE pg_catalog."default",
    created_at timestamp with time zone,
    type text COLLATE pg_catalog."default",
    start_time bigint,
    end_time bigint,
    contents jsonb,
    invalid boolean,
    CONSTRAINT reports_pkey PRIMARY KEY (id),
    CONSTRAINT reports_project_id_dashboard_id_dashboards_project_id_id_foreig FOREIGN KEY (project_id, dashboard_id)
        REFERENCES public.dashboards (project_id, id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT reports_project_id_projects_id_foreign FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.reports
    OWNER to postgres;

-- Index: report_project_id_dashboard_id_type_st_et_unique_idx
CREATE INDEX report_project_id_dashboard_id_type_st_et_unique_idx
    ON public.reports USING btree
    (project_id, dashboard_id, type COLLATE pg_catalog."default", start_time, end_time)
    TABLESPACE pg_default;