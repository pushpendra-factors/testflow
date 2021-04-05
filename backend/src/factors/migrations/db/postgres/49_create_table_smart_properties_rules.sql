--UP
CREATE TABLE public.smart_properties_rules (
    id character varying(255) NOT NULL DEFAULT uuid_generate_v4(),
    project_id bigint NOT NULL,
    type_alias text,
    type bigint NOT NULL,
    description text,
    name text NOT NULL,
    rules jsonb NOT NULL,
    picked bool DEFAULT FALSE,
    is_deleted bool DEFAULT FALSE,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT smart_properties_rules_primary_key PRIMARY KEY (project_id, id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.smart_properties_rules OWNER to autometa;

--DOWN
-- drop table public.smart_properties_rules;