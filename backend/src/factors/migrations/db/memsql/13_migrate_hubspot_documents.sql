-- drop existing hubspot_documents table
DROP TABLE hubspot_documents;

-- create new hubspot_documents table
CREATE TABLE IF NOT EXISTS hubspot_documents (
    id text,
	project_id bigint,
    type int,
    action int,
    timestamp bigint,
    value JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text,
    user_id text,
    group_user_id text,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id, type, id),
    KEY (project_id, type, action, id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (user_id) USING HASH,
    KEY (type) USING HASH,
    KEY (synced) USING HASH
    UNIQUE KEY project_id_id_type_action_timestamp_unique_idx(project_id, id, type,action,timestamp) USING HASH
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Unique (project_id, id, type, action, timestamp)
    -- Ref (project_id, user_id) -> users(project_id, id)
);