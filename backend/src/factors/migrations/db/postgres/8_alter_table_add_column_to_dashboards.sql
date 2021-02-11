-- UP
ALTER TABLE dashboards ADD COLUMN units_position jsonb;

-- DOWN
-- ALTER TABLE dashboards DROP COLUMN units_position RESTRICT;