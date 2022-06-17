-- UP
CREATE INDEX id_idx ON users(id) USING HASH;

-- DOWN
-- DROP INDEX id_idx ON users;