-- Add settings column in dashboards table

ALTER TABLE dashboards ADD COLUMN settings jsonb;
-- ALTER TABLE dashboards DROP COLUMN settings;