-- UP
ALTER TABLE dashboards ADD class text;

-- Update queries to set default as 'user_created' for all dashboards.
UPDATE dashboards SET class = 'user_created';
-- Change only existing website analytics enabled projects class to 'web'.
UPDATE dashboards SET class = 'web' WHERE name = 'Website Analytics' AND project_id IN (2, 216, 398);

-- DOWN
ALTER TABLE dashboards DROP COLUMN class;