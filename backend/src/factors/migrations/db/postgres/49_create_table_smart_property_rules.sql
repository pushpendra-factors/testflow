--UP
CREATE TABLE public.smart_property_rules (
    id character varying(255) NOT NULL DEFAULT uuid_generate_v4(),
    project_id bigint NOT NULL,
    type bigint NOT NULL,
    description text,
    name text NOT NULL,
    rules jsonb NOT NULL,
    evaluation_status int NOT NULL,
    is_deleted bool DEFAULT FALSE,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT smart_property_rules_primary_key PRIMARY KEY (project_id, id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.smart_property_rules OWNER to postgres;

--DOWN
-- drop table public.smart_property_rules;