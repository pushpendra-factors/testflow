ALTER TABLE project_settings ADD COLUMN int_client_demandbase boolean DEFAULT false;
ALTER TABLE project_settings ADD COLUMN client_demandbase_key text;