CREATE TABLE public.templates (
    project_id bigint NOT NULL,
    type int NOT NULL,
    thresholds jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT templates_primary_key PRIMARY KEY (project_id, type)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.templates OWNER to postgres;