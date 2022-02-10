CREATE INDEX join_timestamp ON users(join_timestamp) using HASH;
CREATE INDEX is_group_user ON users(is_group_user) using HASH;
CREATE INDEX group_1_id ON users(group_1_id) using HASH;
CREATE INDEX group_2_id ON users(group_2_id) using HASH;
CREATE INDEX group_3_id ON users(group_3_id) using HASH;
CREATE INDEX group_4_id ON users(group_4_id) using HASH;