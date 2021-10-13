ALTER TABLE users CHANGE properties properties_backup;
ALTER TABLE users ADD COLUMN properties TEXT;
ALTER TABLE users ADD COLUMN properties_json AS properties PERSISTED JSON;
UPDATE users SET properties = properties_backup;
ALTER TABLE users DROP COLUMN properties_backup;
