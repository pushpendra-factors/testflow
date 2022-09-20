CREATE TABLE IF NOT EXISTS otp_rules(
    id text NOT NULL,
    project_id bigint NOT NULL,
    rule_type text,
    crm_type text,
    touch_point_time_ref text,
    filters json,
    properties_map json,
    is_deleted boolean DEFAULT false,
    created_by text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (id) USING HASH,
    SHARD KEY (id)
);