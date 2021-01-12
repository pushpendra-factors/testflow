-- UP
CREATE INDEX hubspot_documents_project_id_type_user_id_timestamp_idx ON hubspot_documents(project_id, type, user_id, timestamp DESC);
CREATE INDEX salesforce_documents_project_id_type_user_id_timestamp_idx ON salesforce_documents(project_id, type, user_id, timestamp DESC);

-- DOWN
-- DROP INDEX hubspot_documents_project_id_type_user_id_timestamp_idx;
-- DROP INDEX salesforce_documents_project_id_type_user_id_timestamp_idx;