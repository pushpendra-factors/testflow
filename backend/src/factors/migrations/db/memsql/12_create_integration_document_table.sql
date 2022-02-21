CREATE TABLE IF NOT EXISTS integration_documents (
    document_id text,
    project_id bigint,
    customer_account_id text,
    document_type int,
    timestamp bigint,
    source text,
    value JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id, document_id),
    KEY (updated_at) USING HASH,
    KEY (project_id, customer_account_id, document_id, document_type, source, timestamp)  USING CLUSTERED COLUMNSTORE
);
