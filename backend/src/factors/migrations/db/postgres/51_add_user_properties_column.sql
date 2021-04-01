-- UP
ALTER TABLE users ADD COLUMN properties JSONB, ADD COLUMN properties_updated_timestamp bigint;
ALTER TABLE events ADD COLUMN user_properties JSONB;

-- DOWN
ALTER TABLE users DROP COLUMN properties, DROP COLUMN properties_updated_timestamp;
ALTER TABLE events DROP COLUMN user_properties;
