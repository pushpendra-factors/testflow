-- UP
ALTER TABLE project_settings ADD COLUMN int_hubspot_first_time_synced boolean NOT NULL DEFAULT 'false'
ALTER TABLE project_settings ADD COLUMN int_hubspot_portal_id int

-- DOWN
-- ALTER TABLE project_settings DROP COLUMN int_hubspot_first_time_synced
-- ALTER TABLE project_settings DROP COLUMN int_hubspot_portal_id
