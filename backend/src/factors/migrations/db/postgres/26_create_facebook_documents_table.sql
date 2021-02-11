-- UP

CREATE TABLE public.facebook_documents (
    id text NOT NULL,
    project_id bigint NOT NULL,
    customer_ad_account_id text NOT NULL,
    platform text NOT NULL,
    type integer NOT NULL,
    timestamp bigint NOT NULL,
    value jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT facebook_documents_pkey PRIMARY KEY (project_id, customer_ad_account_id, platform, type, timestamp, id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.facebook_documents OWNER to postgres;

-- DOWN
-- DROP TABLE facebook_documents;