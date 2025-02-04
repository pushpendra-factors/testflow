-- UP
CREATE TABLE public.salesforce_documents (
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
    CONSTRAINT salesforce_documents_pkey PRIMARY KEY (project_id, id, type, timestamp)
) WITH (OIDS = FALSE) TABLESPACE pg_default;

ALTER TABLE
    public.salesforce_documents OWNER to postgres;

ALTER TABLE
    salesforce_documents
ADD
    CONSTRAINT salesforce_documents_foreign FOREIGN KEY (project_id) REFERENCES public.projects (id) MATCH SIMPLE ON UPDATE RESTRICT ON DELETE RESTRICT

-- DOWN
-- DROP TABLE salesforce_documents;