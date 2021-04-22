-- UP
CREATE INDEX adwords_documents_updated_at ON projects(updated_at ASC);
CREATE INDEX linkedin_documents_updated_at ON projects(updated_at ASC);
CREATE INDEX property_details_updated_at ON projects(updated_at ASC);
CREATE INDEX smart_properties_updated_at ON projects(updated_at ASC);
CREATE INDEX smart_property_rules_updated_at ON projects(updated_at ASC);


-- DOWN
DROP INDEX adwords_documents_updated_at;
DROP INDEX linkedin_documents_updated_at;
DROP INDEX property_details_updated_at;
DROP INDEX smart_properties_updated_at;
DROP INDEX smart_property_rules_updated_at;