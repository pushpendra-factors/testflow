--UP
ALTER TABLE google_organic_documents DROP CONSTRAINT google_organic_documents_primary_key;
ALTER TABLE google_organic_documents ADD CONSTRAINT google_organic_documents_primary_key PRIMARY KEY (project_id, url_prefix, timestamp, type, id);
