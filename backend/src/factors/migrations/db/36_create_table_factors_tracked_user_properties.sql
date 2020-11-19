-- UP
CREATE SEQUENCE public.factors_tracked_user_properties_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE public.factors_tracked_user_properties (
    id bigint NOT NULL DEFAULT nextval('factors_tracked_user_properties_id_seq'::regclass),
    project_id bigint,
    user_property_name text,
    "type" varchar(2) NOT NULL,
    created_by varchar(255),
    last_tracked_at timestamp with time zone,
    is_active boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT factors_tracked_user_properties_pkey PRIMARY KEY (id),
    CONSTRAINT factors_tracked_user_properties_project_id_projects_id_foreign FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT factors_tracked_user_properties_agent_uuid_created_by_foreign FOREIGN KEY (created_by)
        REFERENCES public.agents (uuid) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

CREATE UNIQUE INDEX projectid_userpropertyname_unique_idx ON factors_tracked_user_properties (project_id, user_property_name);

-- DOWN
-- DROP TABLE factors_tracked_user_properties;