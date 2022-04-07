-- create crm_users table
CREATE TABLE IF NOT EXISTS crm_users (
    id text NOT NULL,
    project_id bigint NOT NULL,
    source int NOT NULL,
    type int NOT NULL ,
    timestamp bigint NOT NULL,
    email text,
    phone text,
    action int NOT NULL,
    metadata JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4' NOT NULL,
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text,
    user_id text,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id, source, type, id),
    KEY (project_id, source, type, action, id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (synced) USING HASH,
    UNIQUE KEY project_id_source_id_type_timestamp_unique_idx(project_id,source, id, type, action, timestamp) USING HASH
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Unique (project_id,source, id, type, action, timestamp)
    -- Ref (project_id, user_id) -> users(project_id, id)
);


-- create crm_groups table
CREATE TABLE IF NOT EXISTS crm_groups (
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
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id, source, type, id),
    KEY (project_id, source, type, action, id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (synced) USING HASH,
    UNIQUE KEY project_id_source_id_type_timestamp_unique_idx(project_id,source, id, type, action, timestamp) USING HASH
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Unique (project_id,source, id, type, action, timestamp)
    -- Ref (project_id, user_id) -> users(project_id, id)
); 


-- create new crm_relationships table
CREATE TABLE IF NOT EXISTS crm_relationships (
    id text NOT NULL,
    project_id bigint NOT NULL,
    source int NOT NULL,
    from_type int NOT NULL,
    from_id text NOT NULL,
    to_type int NOT NULL,
    to_id text NOT NULL,
    timestamp bigint NOT NULL,
    external_relationship_name text,
    external_relationship_id text,
    properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    skip_process  boolean NOT NULL DEFAULT FALSE,
    synced boolean NOT NULL DEFAULT FALSE,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id, source, from_type,from_id),
    KEY (project_id, source, from_type, to_type, from_id, to_id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (synced) USING HASH,
    UNIQUE KEY project_id_source_id_type_timestamp_unique_idx(project_id, source, from_type, from_id, to_type, to_id) USING HASH
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Unique (project_id, source, from_type, from_id, to_type, to_id)
);

-- create new crm_activities table
CREATE TABLE IF NOT EXISTS crm_activities (
    id text NOT NULL,
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
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id, source, type, id),
    KEY (project_id, source, type, action, id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (user_id) USING HASH,
    KEY (synced) USING HASH,
    UNIQUE KEY project_id_source_id_type_timestamp_unique_idx(project_id,source, id, type, actor_type, actor_id, timestamp) USING HASH
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Unique (project_id,source, id, type, actor_type, actor_id, timestamp)
    -- Ref (project_id, user_id) -> users(project_id, id)
);