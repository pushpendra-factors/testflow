-- Database: Postgres

-- UP
CREATE TABLE public.user_properties_migration_metadata (
    table_name text PRIMARY KEY,
    last_run_at timestamp with time zone,
    count bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.user_properties_migration_metadata OWNER to postgres;

-- DOWN
DROP TABLE public.user_properties_migration_metadata;