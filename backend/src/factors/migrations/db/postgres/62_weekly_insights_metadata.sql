-- Table: public.weekly_insights_metadata

CREATE TABLE public.weekly_insights_metadata
(
    id TEXT NOT NULL DEFAULT uuid_generate_v4(),
    project_id bigint NOT NULL,
    dashboard_unit_id bigint NOT NULL,
    insight_type text NOT NULL,
    base_start_time  bigint NOT NULL, 
    base_end_time bigint NOT NULL,
    comparison_start_time  bigint NOT NULL, 
    comparison_end_time bigint NOT NULL,
    insight_id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT weekly_insights_metadata_primary_key PRIMARY KEY (id, project_id)
);

ALTER TABLE public.weekly_insights_metadata OWNER to postgres;

ALTER TABLE weekly_insights_metadata ADD CONSTRAINT weekly_insights_metadata_project_id_projects_id_foreign FOREIGN KEY (project_id) REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT;

ALTER TABLE weekly_insights_metadata ADD CONSTRAINT weekly_insights_metadata_dashboard_unit_id_projects_id_foreign FOREIGN KEY (dashboard_unit_id) REFERENCES public.dashboard_unit (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT;

CREATE INDEX weekly_insights_metadata_updated_at ON weekly_insights_metadata(updated_at ASC);
CREATE INDEX weekly_insights_metadata_project_id ON weekly_insights_metadata(project_id ASC);
CREATE UNIQUE INDEX weekly_insights_metadata_project_id_stdate_enddate_unique_idx on weekly_insights_metadata(project_id, dashboard_unit_id, base_start_time, base_end_time, comparison_start_time, comparison_end_time);

-- DROP TABLE public.weekly_insights_metadata;
-- DOWN
-- ALTER TABLE weekly_insights_metadata DROP CONSTRAINT display_names_pkey;
-- ALTER TABLE weekly_insights_metadata DROP COLUMN id;