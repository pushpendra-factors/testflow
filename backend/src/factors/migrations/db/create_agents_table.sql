-- Table: public.agents

-- DROP TABLE public.agents;

CREATE TABLE public.agents
(
    uuid character varying(255) COLLATE pg_catalog."default" NOT NULL DEFAULT uuid_generate_v4(),
    first_name character varying(100) COLLATE pg_catalog."default",
    last_name character varying(100) COLLATE pg_catalog."default",
    email character varying(100) COLLATE pg_catalog."default",
    is_email_verified boolean,
    salt character varying(100) COLLATE pg_catalog."default",
    password character varying(100) COLLATE pg_catalog."default",
    password_created_at timestamp with time zone,
    invited_by uuid,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    is_deleted boolean,
    last_logged_in_at timestamp with time zone,
    login_count bigint,
    CONSTRAINT agents_pkey PRIMARY KEY (uuid)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.agents
    OWNER to autometa;

-- Index: agent_email_unique_idx

-- DROP INDEX public.agent_email_unique_idx;

CREATE UNIQUE INDEX agent_email_unique_idx
    ON public.agents USING btree
    (email COLLATE pg_catalog."default")
    TABLESPACE pg_default;