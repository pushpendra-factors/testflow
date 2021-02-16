-- UP

CREATE TABLE public.linkedin_documents (
    id text NOT NULL,

    project_id bigint NOT NULL,
    customer_ad_account_id text NOT NULL,
    type integer NOT NULL,
    timestamp bigint NOT NULL,
    value jsonb,
    creative_id text,
    campaign_group_id text,
    campaign_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT linkedin_documents_pkey PRIMARY KEY (project_id, customer_ad_account_id,type, timestamp, id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.linkedin_documents OWNER to postgres;

-- DOWN
-- DROP TABLE linkedin_documents;