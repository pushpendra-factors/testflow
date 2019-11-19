-- UP

CREATE TABLE public.adwords_documents (
    project_id bigint NOT NULL,
    customer_account_id text NOT NULL,
    type integer NOT NULL,
    "timestamp" bigint NOT NULL,
    value jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.agents OWNER to postgres;

-- DOWN
-- DROP TABLE adwords_documents;