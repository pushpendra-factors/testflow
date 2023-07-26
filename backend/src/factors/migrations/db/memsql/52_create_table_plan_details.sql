CREATE TABLE IF NOT EXISTS plan_details (
    id bigint auto_increment,
    name text,
    mtu_limit bigint,
    feature_list json,
    SHARD KEY (id),
    PRIMARY KEY (id)
);