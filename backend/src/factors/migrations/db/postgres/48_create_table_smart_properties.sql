-- UP

CREATE TABLE public.smart_properties (
    project_id bigint NOT NULL,
    source text NOT NULL,
    object_id text NOT NULL,
    object_type bigint NOT NULL,
    object_property jsonb NOT NULL,
    properties jsonb NOT NULL,
    rules_ref jsonb NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT smart_properties_primary_key PRIMARY KEY (project_id, object_id, object_type, source)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.smart_properties OWNER to postgres;

-- DOWN
-- DROP TABLE smart_properties;
