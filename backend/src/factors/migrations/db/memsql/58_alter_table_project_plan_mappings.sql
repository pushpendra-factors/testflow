ALTER TABLE project_plan_mappings ADD COLUMN billing_plan_id text;
ALTER TABLE project_plan_mappings ADD COLUMN billing_addons json;
ALTER TABLE project_plan_mappings ADD COLUMN billing_last_synced_at timestamp(6);