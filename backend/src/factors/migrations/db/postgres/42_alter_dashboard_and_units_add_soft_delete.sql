-- UP
ALTER TABLE dashboards ADD COLUMN is_deleted boolean DEFAULT FALSE;
ALTER TABLE dashboard_units ADD COLUMN is_deleted boolean DEFAULT FALSE;
ALTER TABLE reports ADD COLUMN is_deleted boolean DEFAULT FALSE;

-- DOWN
-- ALTER TABLE dashboards DROP COLUMN is_deleted;
-- ALTER TABLE dashboard_units DROP COLUMN is_deleted;
-- ALTER TABLE reports DROP COLUMN is_deleted;