CREATE TABLE IF NOT EXISTS cache_db (
    k TEXT NOT NULL,
    v TEXT,
    project_id BIGINT NOT NULL,
    expiry_in_secs INT NOT NULL,
    expires_at INT NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,    
    PRIMARY KEY (k),
    SHARD KEY (k),
    KEY (expires_at) USING CLUSTERED COLUMNSTORE
);

