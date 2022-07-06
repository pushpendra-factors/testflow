CREATE TABLE IF NOT EXISTS dashboard_templates(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    title text,
    description text,
    dashboard json,
    units json,
    is_deleted boolean DEFAULT false,
    similar_template_ids json,
    tags json,
);