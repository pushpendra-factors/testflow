-- UP
CREATE INDEX source_idx ON users(source) USING HASH;


-- DOWN
-- DROP INDEX source_idx ON users;