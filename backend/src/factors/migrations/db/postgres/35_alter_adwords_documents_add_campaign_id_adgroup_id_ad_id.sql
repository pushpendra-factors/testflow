-- Add campaign_id column in adwords_documents table
ALTER TABLE adwords_documents ADD COLUMN campaign_id bigint;
-- ALTER TABLE adwords_documents DROP COLUMN campaign_id;

-- Add ad_group_id column in adwords_documents table
ALTER TABLE adwords_documents ADD COLUMN ad_group_id bigint;
-- ALTER TABLE adwords_documents DROP COLUMN ad_group_id;

-- Add ad_id column in adwords_documents table
ALTER TABLE adwords_documents ADD COLUMN ad_id bigint;
-- ALTER TABLE adwords_documents DROP COLUMN ad_id;


-- Add keyword_id column in adwords_documents table
ALTER TABLE adwords_documents ADD COLUMN keyword_id bigint;
-- ALTER TABLE adwords_documents DROP COLUMN keyword_id;
