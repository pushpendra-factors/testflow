CREATE TABLE public.content_groups(
    id text NOT NULL DEFAULT uuid_generate_v4(),
    project_id bigint NOT NULL,
    content_group_name text,
    content_group_description text,
    rule jsonb NOT NULL,
    created_by varchar(255),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    is_deleted boolean,
    CONSTRAINT content_groups_primary_key PRIMARY KEY (id, project_id)
);

ALTER TABLE public.content_group OWNER to postgres;

ALTER TABLE content_groups ADD CONSTRAINT content_group_project_id_projects_id_foreign FOREIGN KEY (project_id) 
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT;

ALTER TABLE content_groups CONSTRAINT content_group_agent_uuid_created_by_foreign FOREIGN KEY (created_by)
        REFERENCES public.agents (uuid) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT

CREATE INDEX content_group_updated_at ON content_groups(updated_at ASC);
CREATE INDEX content_group_project_id ON content_groups(project_id ASC);
CREATE UNIQUE INDEX content_group_project_id_name_value_unique_idx on content_groups(project_id, content_group_name);
