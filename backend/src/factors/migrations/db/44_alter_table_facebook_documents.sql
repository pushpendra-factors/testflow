--UP
ALTER TABLE facebook_documents DROP COLUMN campaign_id, DROP COLUMN ad_set_id, DROP COLUMN ad_id;
ALTER TABLE facebook_documents ADD COLUMN campaign_id text, ADD COLUMN ad_set_id text, ADD COLUMN ad_id text;

--DOWN
-- ALTER TABLE facebook_documents DROP COLUMN campaign_id, DROP COLUMN ad_set_id, DROP COLUMN ad_id;