-- Add campaign_id column in facebook_documents table
ALTER TABLE facebook_documents ADD COLUMN campaign_id bigint;
-- ALTER TABLE facebook_documents DROP COLUMN campaign_id;

-- Add ad_set_id column in facebook_documents table
ALTER TABLE facebook_documents ADD COLUMN ad_set_id bigint;
-- ALTER TABLE facebook_documents DROP COLUMN ad_set_id;

-- Add ad_id column in facebook_documents table
ALTER TABLE facebook_documents ADD COLUMN ad_id bigint;
-- ALTER TABLE facebook_documents DROP COLUMN ad_id;