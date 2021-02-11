-- UP
CREATE SEQUENCE public.factors_tracked_events_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE public.factors_tracked_events (
    id bigint NOT NULL DEFAULT nextval('factors_tracked_events_id_seq'::regclass),
    project_id bigint,
    event_name_id bigint,
    "type" varchar(2) NOT NULL,
    created_by varchar(255),
    last_tracked_at timestamp with time zone,
    is_active boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT factors_tracked_events_pkey PRIMARY KEY (id),
    CONSTRAINT factors_tracked_events_project_id_projects_id_foreign FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT factors_tracked_events_event_name_id_event_name_id_foreign FOREIGN KEY (project_id, event_name_id)
        REFERENCES public.event_names (project_id, id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT factors_tracked_events_agent_uuid_created_by_foreign FOREIGN KEY (created_by)
        REFERENCES public.agents (uuid) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

CREATE UNIQUE INDEX projectid_eventnameid_unique_idx ON factors_tracked_events (project_id, event_name_id);

-- DOWN
-- DROP TABLE factors_tracked_events;