-- UP

ALTER TABLE users
ADD COLUMN is_group_user boolean,
ADD COLUMN group_1_id text,
ADD COLUMN group_1_user_id text,
ADD COLUMN group_2_id text,
ADD COLUMN group_2_user_id text,
ADD COLUMN group_3_id text,
ADD COLUMN group_3_user_id text,
ADD COLUMN group_4_id text,
ADD COLUMN group_4_user_id text; 

CREATE ROWSTORE TABLE IF NOT EXISTS groups(
    project_id bigint NOT NULL,
    id int NOT NULL,
    name text NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, name),
    UNIQUE KEY (project_id,id)
);

-- DOWN

/*

ALTER TABLE users
DROP COLUMN is_group_user,
DROP COLUMN group_1_id,
DROP COLUMN group_1_user_id,
DROP COLUMN group_2_id,
DROP COLUMN group_2_user_id,
DROP COLUMN group_3_id,
DROP COLUMN group_3_user_id,
DROP COLUMN group_4_id,
DROP COLUMN group_4_user_id; 

DROP TABLE groups;
*/
