-- create crm_users table
CREATE TABLE public.crm_users (
    id text NOT NULL,
    project_id bigint NOT NULL,
    source int NOT NULL,
    type int NOT NULL,
    timestamp bigint NOT NULL,
    email text,
    phone text,
    action int NOT NULL,
    metadata JSON,
    properties JSON NOT NULL,
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text,
    user_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    CONSTRAINT crm_users_pkey PRIMARY KEY (project_id, source, id, type, action, timestamp)
);

ALTER TABLE
    public.crm_users
ADD
    CONSTRAINT crm_users_foreign FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


-- create crm_groups table
CREATE TABLE public.crm_groups (
    id text NOT NULL,
    project_id bigint NOT NULL,
    source int NOT NULL,
    type int NOT NULL,
    timestamp bigint NOT NULL,
    action int NOT NULL,
    metadata JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4' NOT NULL,
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text,
    user_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    CONSTRAINT crm_groups_pkey PRIMARY KEY (project_id, source, id, type, action, timestamp)

);

ALTER TABLE
    public.crm_groups
ADD
    CONSTRAINT crm_groups_foreign FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


-- create new crm_relationships table
CREATE TABLE public.crm_relationships (
    id text NOT NULL uuid_generate_v4(),
    project_id bigint NOT NULL,
    source int NOT NULL,
    from_type int NOT NULL,
    from_id text NOT NULL,
    to_type int NOT NULL,
    to_id text NOT NULL,
    timestamp bigint NOT NULL,
    external_relationship_name text,
    external_relationship_id text,
    properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4' NOT NULL,
    skip_process boolean NOT NULL DEFAULT FALSE,
    synced boolean NOT NULL DEFAULT FALSE,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    CONSTRAINT crm_relationships_pkey PRIMARY KEY (project_id, source, from_type, from_id, to_type, to_id)
);

ALTER TABLE
    public.crm_relationships
ADD
    CONSTRAINT crm_relationships_foreign FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


-- create new crm_activities table
CREATE TABLE public.crm_activities (
    id text uuid_generate_v4(),
    project_id bigint NOT NULL,
    source int NOT NULL,
    name text NOT NULL,
    type int NOT NULL,
    actor_type int NOT NULL,
    actor_id text NOT NULL,
    timestamp bigint NOT NULL,
    action int NOT NULL,
    properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4' NOT NULL,
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text,
    user_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    CONSTRAINT crm_activities_pkey PRIMARY KEY (project_id, source, id, type, actor_type, actor_id, timestamp)
);
ALTER TABLE
    public.crm_activities
ADD
    CONSTRAINT crm_activities_foreign FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE RESTRICT ON DELETE RESTRICT;
