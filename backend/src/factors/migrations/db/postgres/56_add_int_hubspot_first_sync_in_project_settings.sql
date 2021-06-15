-- UP
ALTER TABLE project_settings ADD COLUMN int_hubspot_first_time_synced boolean NOT NULL DEFAULT 'false'
ALTER TABLE project_settings ADD COLUMN int_hubspot_portal_id int
ALTER TABLE project_settings add column int_hubspot_sync_info jsonb

-- DOWN
-- ALTER TABLE project_settings DROP COLUMN int_hubspot_first_time_synced
-- ALTER TABLE project_settings DROP COLUMN int_hubspot_portal_id
-- ALTER TABLE project_settings add column int_hubspot_sync_info jsonb
