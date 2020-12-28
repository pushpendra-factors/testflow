-- Add user_id column in salesforce_documents table
ALTER TABLE salesforce_documents ADD COLUMN user_id text;
-- ALTER TABLE salesforce_documents DROP COLUMN user_id;