CREATE TABLE IF NOT EXISTS segment_folders (
    id  bigint NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name text NOT NULL,
    project_id bigint(20),
    folder_type text,
    created_at timestamp not null,
    updated_at timestamp not null,
    KEY (project_id) USING CLUSTERED COLUMNSTORE
);
