-- UP
ALTER TABLE events ADD properties_updated_timestamp bigint NOT NULL DEFAULT 0;
ALTER TABLE user_properties ADD updated_timestamp bigint NOT NULL DEFAULT 0;

--  DOWN
-- ALTER TABLE events DROP properties_updated_timestamp;
-- ALTER TABLE user_properties DROP updated_timestamp;
