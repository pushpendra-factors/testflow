ALTER TABLE hubspot_documents ADD COLUMN group_user_id text;


-- move company user_id to group_user_id
UPDATE hubspot_documents set group_user_id = user_id where project_id =595 and type =1 and user_id != '';