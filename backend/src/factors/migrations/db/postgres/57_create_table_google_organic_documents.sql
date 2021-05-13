-- UP

CREATE TABLE public.google_organic_documents (
    id text NOT NULL,
    project_id bigint NOT NULL,
    url_prefix text NOT NULL,
    timestamp bigint NOT NULL,
    value jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT google_organic_documents_primary_key PRIMARY KEY (id, project_id, url_prefix, timestamp)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.google_organic_documents OWNER to postgres;

-- DOWN
-- DROP TABLE google_organic_documents;