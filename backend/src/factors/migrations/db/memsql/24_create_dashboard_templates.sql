CREATE TABLE IF NOT EXISTS dashboard_templates(
    id text NOT NULL,
    title text,
    description text,
    dashboard json,
    units json,
    is_deleted boolean DEFAULT false,
    similar_template_ids json,
    tags json,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
    KEY (id) USING HASH,
    SHARD KEY (id),
);