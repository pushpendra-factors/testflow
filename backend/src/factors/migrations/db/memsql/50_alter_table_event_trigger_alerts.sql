ALTER TABLE event_trigger_alerts ADD COLUMN internal_status TEXT;
UPDATE event_trigger_alerts SET internal_status = 'active';