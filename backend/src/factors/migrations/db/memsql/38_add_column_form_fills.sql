-- UP
ALTER TABLE form_fills ADD COLUMN event_properties JSON;

-- DOWM
-- ALTER TABLE form_fills DROP COLUMN event_properties;