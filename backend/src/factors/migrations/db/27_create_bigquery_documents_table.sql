-- UP

CREATE TABLE public.bigquery_settings (
    id text NOT NULL,
    project_id bigint NOT NULL,
    bq_project_id text NOT NULL,
    bq_dataset_name text NOT NULL,
    bq_credentials_json text NOT NULL,
    last_run_at bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT bigquery_settings_pkey PRIMARY KEY (id, project_id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.bigquery_settings OWNER to postgres;

-- Alter project settings to add archival related columns.
ALTER TABLE project_settings
ADD COLUMN archive_enabled boolean DEFAULT FALSE,
ADD COLUMN bigquery_enabled boolean DEFAULT FALSE;

-- DOWN
-- DROP TABLE bigquery_settings;
-- ALTER TABLE project_settings
-- DROP COLUMN archive_enabled,
-- DROP COLUMN bigquery_enabled;