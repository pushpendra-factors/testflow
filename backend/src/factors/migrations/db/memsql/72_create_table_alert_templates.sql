CREATE TABLE IF NOT EXISTS alert_templates (
    v TEXT NOT NULL,
    id int NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title TEXT NOT NULL,
    alert json not null,
    template_constants json not null,
    is_deleted boolean not null DEFAULT false,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);