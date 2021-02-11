-- UP
ALTER TABLE events ADD COLUMN customer_event_id varchar(500);
CREATE UNIQUE INDEX project_id_customer_event_id_unique_idx ON events (project_id, customer_event_id);

-- DOWN
-- DROP INDEX project_id_customer_event_id_unique_idx;
-- ALTER TABLE events DROP COLUMN customer_event_id RESTRICT;