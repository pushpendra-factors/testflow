-- UP
ALTER TABLE display_names ADD COLUMN id TEXT NOT NULL DEFAULT uuid_generate_v4();
ALTER TABLE display_names ADD PRIMARY KEY (id, project_id);

-- DOWN
ALTER TABLE display_names DROP CONSTRAINT display_names_pkey;
ALTER TABLE display_names DROP COLUMN id;