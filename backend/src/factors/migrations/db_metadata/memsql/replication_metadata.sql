-- Database: MemSQL

-- UP
CREATE TABLE replication_metadata (
    table_name text PRIMARY KEY,
    last_run_at timestamp(6) NOT NULL,
    count bigint,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL
);

-- DOWN
DROP TABLE replication_metadata;