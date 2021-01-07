-- Add user_id column in hubspot_documents table
ALTER TABLE hubspot_documents ADD COLUMN user_id text;
-- ALTER TABLE hubspot_documents DROP COLUMN user_id;