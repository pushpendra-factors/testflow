-- UP

CREATE TABLE public.scheduled_tasks (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),  -- Task id.
    job_id text NOT NULL,  -- Id for the parent task.
    project_id bigint NOT NULL,
    task_type text NOT NULL,
    task_status text NOT NULL,
    task_start_time bigint NOT NULL,
    task_end_time bigint,
    task_details jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.scheduled_tasks OWNER to postgres;

-- DOWN
-- DROP TABLE scheduled_tasks;