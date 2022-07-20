-- create clickable_elements table
CREATE TABLE IF NOT EXISTS clickable_elements (
    project_id bigint NOT NULL,
    id text NOT NULL,
    display_name text NOT NULL,
    element_type text,
    element_attributes json,
    click_count int NOT NULL,
    enabled boolean DEFAULT false,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    PRIMARY KEY(project_id, id)
);