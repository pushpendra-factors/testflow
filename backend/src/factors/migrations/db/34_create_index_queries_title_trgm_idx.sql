-- Create extension using superuser
-- CREATE EXTENSION pg_trgm;

--UP
CREATE INDEX queries_title_trgm_idx ON queries USING gin (title gin_trgm_ops);

--down
-- DROP INDEX queries_title_trgm_idx; 