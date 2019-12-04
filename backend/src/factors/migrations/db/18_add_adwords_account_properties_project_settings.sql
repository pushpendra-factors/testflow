-- UP
ALTER TABLE project_settings ADD COLUMN int_adwords_customer_account_properties jsonb;

-- DOWN
-- ALTER TABLE project_settings DROP COLUMN int_adwords_customer_account_properties RESTRICT;