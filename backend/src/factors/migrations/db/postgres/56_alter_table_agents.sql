-- UP 
ALTER TABLE agents ADD COLUMN int_google_organic_refresh_token text;

-- DOWN 
-- ALTER TABLE agents DROP COLUMN int_google_organic_refresh_token;