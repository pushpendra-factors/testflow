ALTER TABLE agents ADD COLUMN is_auth0_user boolean DEFAULT false;
ALTER TABLE agents ADD COLUMN value json;