create table group_relationships(
    project_id bigint,
    left_group_name_id int,
    left_group_user_id text,
    right_group_name_id int,
    right_group_user_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE UNIQUE INDEX project_id_left_group_user_id_right_group_user_id_unique_idx 
on group_relationships(project_id, left_group_user_id,right_group_user_id);