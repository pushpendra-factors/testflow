-- UP

-- Drop constraint on one table to enable changing type. Change type to text for uuid.
ALTER TABLE project_billing_account_mappings DROP CONSTRAINT project_billing_account_mappings_billing_account_id_billing_acc;
ALTER TABLE project_billing_account_mappings ALTER COLUMN billing_account_id TYPE text;
ALTER TABLE billing_accounts ALTER COLUMN id TYPE text;

ALTER TABLE billing_accounts ALTER COLUMN id SET DEFAULT uuid_generate_v4();
ALTER TABLE project_billing_account_mappings ALTER COLUMN billing_account_id SET DEFAULT NULL;

-- Temporary table for billing_account_ids.
CREATE TABLE billing_account_id_to_uuid (
    id bigint PRIMARY KEY NOT NULL,
    uuid text NOT NULL DEFAULT uuid_generate_v4()
);
INSERT INTO billing_account_id_to_uuid (id) SELECT id::bigint FROM billing_accounts;

-- Update old bigint ids with uuid.
UPDATE project_billing_account_mappings SET billing_account_id = (SELECT uuid FROM billing_account_id_to_uuid WHERE id = billing_account_id::bigint);
UPDATE billing_accounts ba SET id = (SELECT uuid FROM billing_account_id_to_uuid bamap WHERE bamap.id = ba.id::bigint);

-- Put back the dropped constraint.
ALTER TABLE project_billing_account_mappings ADD CONSTRAINT project_billing_account_mappings_billing_account_id_billing_acc FOREIGN KEY (billing_account_id)
    REFERENCES public.billing_accounts (id) MATCH SIMPLE ON UPDATE RESTRICT ON DELETE RESTRICT;

-- WARN: Drop later after verifying from app that nothing is breaking.
-- Drop the id <> uuid mapping table.
DROP TABLE billing_account_id_to_uuid;
DROP SEQUENCE billing_accounts_id_seq;
DROP SEQUENCE project_billing_account_mappings_billing_account_id_seq;





-- DOWN
-- Use above created mapping table to revert back to old ids.
ALTER TABLE project_billing_account_mappings DROP CONSTRAINT project_billing_account_mappings_billing_account_id_billing_acc;

UPDATE project_billing_account_mappings SET billing_account_id = (SELECT id::text FROM billing_account_id_to_uuid WHERE uuid = billing_account_id);
UPDATE billing_accounts ba SET id = (SELECT id::text FROM billing_account_id_to_uuid bamap WHERE uuid = ba.id);

ALTER TABLE project_billing_account_mappings ALTER COLUMN billing_account_id TYPE bigint USING billing_account_id::bigint;
ALTER TABLE billing_accounts ALTER COLUMN id TYPE bigint USING id::bigint;

ALTER TABLE project_billing_account_mappings ADD CONSTRAINT project_billing_account_mappings_billing_account_id_billing_acc FOREIGN KEY (billing_account_id)
    REFERENCES public.billing_accounts (id) MATCH SIMPLE ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE billing_accounts ALTER COLUMN id SET DEFAULT nextval('billing_accounts_id_seq');
ALTER TABLE project_billing_account_mappings ALTER COLUMN billing_account_id SET DEFAULT nextval('project_billing_account_mappings_billing_account_id_seq');