CREATE SEQUENCE public.factors_task_details_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE public.task_details
(
    task_id bigint NOT NULL DEFAULT nextval('factors_task_details_id_seq'::regclass),
    task_name text NOT NULL,
    source text NULL,
    frequency integer NOT NULL,
    frequency_interval integer, -- There are 4 types hourly/daily/weekly/stateless
    skip_start_index integer,
    skip_end_index integer,
    offset_start_minutes integer, 
    recurrence boolean,
    metadata jsonb,
    is_project_enabled boolean,
    delay_alert_threshold_hours integer,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT task_details_primary_key PRIMARY KEY (task_id)
);

CREATE UNIQUE INDEX task_name_unique_idx on task_details(task_name);

ALTER TABLE public.task_details OWNER to postgres;

CREATE SEQUENCE public.factors_task_execution_details_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE public.task_execution_details
(
    execution_id bigint NOT NULL DEFAULT nextval('factors_task_execution_details_id_seq'::regclass),
    task_id bigint NOT NULL,
    project_id bigint NOT NULL,
    delta bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    metadata jsonb,
    is_completed boolean,
    CONSTRAINT task_execution_details_primary_key PRIMARY KEY (execution_id),
    CONSTRAINT task_execution_details_task_id_foreign FOREIGN KEY (task_id)
        REFERENCES public.task_details (task_id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
);

CREATE UNIQUE INDEX task_id_delta_unique_idx on task_execution_details(task_id, delta);

CREATE INDEX task_execution_details_task_id ON task_execution_details(task_id ASC);

ALTER TABLE public.task_execution_details OWNER to postgres;

CREATE SEQUENCE public.factors_task_execution_dependency_details_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    CACHE 1;

CREATE TABLE public.task_execution_dependency_details
(
    task_id bigint NOT NULL,
    dependent_task_id bigint NOT NULL,
    dependency_offset integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT task_execution_dependency_details_task_id_foreign FOREIGN KEY (task_id)
        REFERENCES public.task_details (task_id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,
    CONSTRAINT task_execution_dependency_details_dependent_task_id_foreign FOREIGN KEY (dependent_task_id)
        REFERENCES public.task_details (task_id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
);

CREATE UNIQUE INDEX task_id_dependent_task_id_unique_idx on task_execution_dependency_details(task_id, dependent_task_id);

CREATE INDEX task_execution_dependency_details_task_id ON task_execution_dependency_details(task_id ASC);

ALTER TABLE public.task_execution_dependency_details OWNER to postgres;

ALTER TABLE task_details ADD COLUMN id TEXT NOT NULL DEFAULT uuid_generate_v4();

ALTER TABLE task_execution_details ADD COLUMN id TEXT NOT NULL DEFAULT uuid_generate_v4();

ALTER TABLE task_execution_dependency_details ADD COLUMN id TEXT NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE task_execution_dependency_details ADD PRIMARY KEY (id);