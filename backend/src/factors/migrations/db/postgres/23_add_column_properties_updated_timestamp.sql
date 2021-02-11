-- UP
ALTER TABLE events ADD properties_updated_timestamp bigint;
ALTER TABLE user_properties ADD updated_timestamp bigint;

--  DOWN
-- ALTER TABLE events DROP properties_updated_timestamp;
-- ALTER TABLE user_properties DROP updated_timestamp;