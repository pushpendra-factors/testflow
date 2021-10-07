CREATE TABLE public.groups
(
    project_id bigint,
    id int not null,
    name text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
ALTER TABLE groups ADD PRIMARY KEY (project_id,name);

CREATE UNIQUE INDEX project_id_id on groups(project_id, id);


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