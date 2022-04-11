CREATE TABLE public.shareable_urls (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    query_id text NOT NULL,
    entity_type integer NOT NULL,
    share_type integer NOT NULL,
    entity_id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    is_deleted boolean NOT NULL DEFAULT false,
    expires_at bigint,
    project_id bigint NOT NULL,
    created_by uuid NOT NULL,
    CONSTRAINT shareable_urls_primary_key PRIMARY KEY (id)
);

ALTER TABLE public.shareable_urls OWNER to postgres;

CREATE TABLE public.shareable_url_audits (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    project_id bigint NOT NULL,
    share_id uuid NOT NULL,
    query_id text,
    entity_id bigint,
    entity_type integer,
    share_type integer,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    is_deleted boolean NOT NULL DEFAULT false,
    expires_at bigint,
    accessed_by uuid,
    CONSTRAINT shareable_url_audits_primary_key PRIMARY KEY (id)
);

ALTER TABLE public.shareable_url_audits OWNER to postgres;

-- DROP TABLE public.shareable_urls;
-- DROP TABLE public.shareable_url_audits;