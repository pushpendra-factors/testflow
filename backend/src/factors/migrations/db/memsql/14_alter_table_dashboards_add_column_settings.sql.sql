-- Add settings column in dashboards table

ALTER TABLE dashboards ADD COLUMN settings JSON;
-- ALTER TABLE dashboards DROP COLUMN settings;