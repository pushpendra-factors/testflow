ALTER TABLE linkedin_documents add column sync_status int default 0;
update linkedin_documents set sync_status = 2 where is_backfilled = TRUE;