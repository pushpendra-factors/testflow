ALTER TABLE projects ADD COLUMN enable_billing boolean;
ALTER TABLE projects ADD COLUMN billing_account_id string;
ALTER TABLE projects ADD COLUMN billing_subscription_id string;
ALTER TABLE projects ADD COLUMN billing_last_synced_at timestamp(6);