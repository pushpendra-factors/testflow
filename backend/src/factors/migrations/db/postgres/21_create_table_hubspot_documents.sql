-- UP

CREATE TABLE public.hubspot_documents (
    id text NOT NULL,
    project_id bigint NOT NULL,
    type integer NOT NULL,
    action integer NOT NULL,
    "timestamp" bigint NOT NULL,
    value jsonb,
    synced boolean NOT NULL DEFAULT false,
    sync_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT hubspot_documents_pkey PRIMARY KEY (project_id, id, type, action, timestamp)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.hubspot_documents OWNER to postgres;

-- DOWN
-- DROP TABLE hubspot_documents;