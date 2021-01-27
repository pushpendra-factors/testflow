-- UP
CREATE INDEX hubspot_documents_project_id_type_timestamp_user_id_idx ON hubspot_documents(project_id, type, timestamp DESC, user_id);
CREATE INDEX salesforce_documents_project_id_type_timestamp_user_id_idx ON salesforce_documents(project_id, type, timestamp DESC, user_id);

-- DOWN
-- DROP INDEX hubspot_documents_project_id_type_user_id_timestamp_idx;
-- DROP INDEX salesforce_documents_project_id_type_user_id_timestamp_idx;