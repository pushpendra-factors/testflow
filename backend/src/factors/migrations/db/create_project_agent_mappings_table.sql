-- Table: public.project_agent_mappings

-- DROP TABLE public.project_agent_mappings;

CREATE TABLE public.project_agent_mappings
(
    agent_uuid character varying(255) COLLATE pg_catalog."default" NOT NULL,
    project_id bigint NOT NULL DEFAULT nextval('project_agent_mappings_project_id_seq'::regclass),
    role bigint,
    invited_by character varying(255) COLLATE pg_catalog."default",
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT project_agent_mappings_pkey PRIMARY KEY (agent_uuid, project_id),
    CONSTRAINT project_agent_mappings_agent_uuid_agents_uuid_foreign FOREIGN KEY (agent_uuid)
        REFERENCES public.agents (uuid) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT project_agent_mappings_invited_by_agents_uuid_foreign FOREIGN KEY (invited_by)
        REFERENCES public.agents (uuid) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT project_agent_mappings_project_id_projects_id_foreign FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.project_agent_mappings
    OWNER to autometa;

-- Index: project_id_agent_uuid_idx

-- DROP INDEX public.project_id_agent_uuid_idx;

CREATE INDEX project_id_agent_uuid_idx
    ON public.project_agent_mappings USING btree
    (project_id, agent_uuid COLLATE pg_catalog."default")
    TABLESPACE pg_default;