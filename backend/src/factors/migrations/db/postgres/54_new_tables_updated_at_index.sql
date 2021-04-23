-- UP
-- Add created_at, updated_at to property_details. Set to now() for existing records.
ALTER TABLE property_details ADD COLUMN updated_at timestamp with time zone;
ALTER TABLE property_details ADD COLUMN created_at timestamp with time zone;
UPDATE property_details set created_at = now(), updated_at = now();

-- Add index to newer tables.
CREATE INDEX adwords_documents_updated_at ON adwords_documents(updated_at ASC);
CREATE INDEX linkedin_documents_updated_at ON linkedin_documents(updated_at ASC);
CREATE INDEX property_details_updated_at ON property_details(updated_at ASC);
CREATE INDEX smart_properties_updated_at ON smart_properties(updated_at ASC);
CREATE INDEX smart_property_rules_updated_at ON smart_property_rules(updated_at ASC);


-- DOWN
ALTER TABLE property_details DROP COLUMN updated_at;
ALTER TABLE property_details DROP COLUMN created_at;

DROP INDEX adwords_documents_updated_at;
DROP INDEX linkedin_documents_updated_at;
DROP INDEX property_details_updated_at;
DROP INDEX smart_properties_updated_at;
DROP INDEX smart_property_rules_updated_at;