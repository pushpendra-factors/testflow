CREATE TABLE public.leadgen_settings (
    project_id bigint NOT NULL,
    source int NOT NULL,
    source_property text NOT NULL,
    spreadsheet_id text,
    sheet_name text,
    row_read bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT leadgen_settings_pkey PRIMARY KEY (project_id, source)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.leadgen_settings OWNER to postgres;