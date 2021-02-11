-- Table: public.project_billing_account_mappings

-- DROP TABLE public.project_billing_account_mappings;

CREATE TABLE public.project_billing_account_mappings
(
    project_id bigint NOT NULL ,
    billing_account_id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT project_billing_account_mappings_pkey PRIMARY KEY (project_id, billing_account_id),
    CONSTRAINT project_billing_account_mappings_billing_account_id_billing_acc FOREIGN KEY (billing_account_id)
        REFERENCES public.billing_accounts (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT project_billing_account_mappings_project_id_projects_id_foreign FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.project_billing_account_mappings
    OWNER to postgres;

-- Index: billing_account_id_project_id_idx

-- DROP INDEX public.billing_account_id_project_id_idx;

CREATE INDEX billing_account_id_project_id_idx
    ON public.project_billing_account_mappings USING btree
    (billing_account_id, project_id)
    TABLESPACE pg_default;