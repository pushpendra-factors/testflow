-- SEQUENCE: public.billing_accounts_id_seq

-- DROP SEQUENCE public.billing_accounts_id_seq;

CREATE SEQUENCE public.billing_accounts_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

ALTER SEQUENCE public.billing_accounts_id_seq
    OWNER TO postgres;

-- Table: public.billing_accounts

-- DROP TABLE public.billing_accounts;

CREATE TABLE public.billing_accounts
(
    id bigint NOT NULL DEFAULT nextval('billing_accounts_id_seq'::regclass),
    plan_id bigint NOT NULL,
    agent_uuid text COLLATE pg_catalog."default" NOT NULL,
    organization_name text COLLATE pg_catalog."default",
    billing_address text COLLATE pg_catalog."default",
    pincode text COLLATE pg_catalog."default",
    phone_no text COLLATE pg_catalog."default",
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT billing_accounts_pkey PRIMARY KEY (id),
    CONSTRAINT billing_accounts_agent_uuid_agents_uuid_foreign FOREIGN KEY (agent_uuid)
        REFERENCES public.agents (uuid) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.billing_accounts
    OWNER to postgres;

-- Index: agent_uuid_unique_idx

-- DROP INDEX public.agent_uuid_unique_idx;

CREATE UNIQUE INDEX agent_uuid_unique_idx
    ON public.billing_accounts USING btree
    (agent_uuid COLLATE pg_catalog."default");