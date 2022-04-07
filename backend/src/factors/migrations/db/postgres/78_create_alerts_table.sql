CREATE TABLE IF NOT EXISTS alerts(
    id text NOT NULL,
    project_id bigint NOT NULL,
    alert_name text,
    created_by text,
    alert_type int,
    alert_description json,
    alert_configuration json,
    last_alert_sent bool,
    last_run_time timestamp(6),
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    is_deleted boolean,
    CONSTRAINT alerts_primary_key PRIMARY KEY (id),
    CONSTRAINT alerts_project_id_projects_id_foreign FOREIGN KEY (project_id) 
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
);

ALTER TABLE public.feedbacks OWNER to postgres;