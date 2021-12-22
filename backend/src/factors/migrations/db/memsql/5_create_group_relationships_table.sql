CREATE TABLE IF NOT EXISTS group_relationships(
    project_id bigint NOT NULL,
    left_group_name_id int NOT NULL,
    left_group_user_id text NOT NULL,
    right_group_name_id int NOT NULL,
    right_group_user_id text NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (left_group_user_id),
    KEY (project_id, left_group_user_id) USING CLUSTERED COLUMNSTORE,
    UNIQUE KEY(project_id, left_group_user_id,right_group_user_id) USING HASH
);