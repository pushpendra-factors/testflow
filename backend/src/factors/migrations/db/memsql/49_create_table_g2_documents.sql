CREATE TABLE IF NOT EXISTS g2_documents (
    id text NOT NULL,
    project_id bigint NOT NULL,
    type int NOT NULL,
    timestamp bigint NOT NULL,
    value JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    synced boolean NOT NULL default FALSE,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id),
    KEY (project_id, type, timestamp) USING CLUSTERED COLUMNSTORE
);