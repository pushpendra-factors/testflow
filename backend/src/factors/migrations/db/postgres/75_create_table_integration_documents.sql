CREATE TABLE public.fivetran_mappings(
    project_id bigint NOT NULL,
    id text NOT NULL,
    integration text NOT NULL,
    connector_id text NOT NULL,
    schema_id text NOT NULL,
    status boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

ALTER TABLE public.fivetran_mappings OWNER to postgres;

CREATE TABLE public.integration_documents (
    document_id text,
    project_id bigint,
    customer_account_id text,
    document_type int,
    timestamp bigint,
    source text,
    value jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

ALTER TABLE public.integration_documents OWNER to postgres;
