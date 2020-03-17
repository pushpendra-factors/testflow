#!/bin/bash

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    \c postgres;
    SELECT pg_catalog.set_config('search_path', '', false);
    ALTER ROLE autometa PASSWORD 'md5aa44dccdc9ac7b7a4a2a25e129e95784' SUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
    GRANT ALL PRIVILEGES ON DATABASE autometa TO autometa;
    \c autometa
    CREATE EXTENSION "uuid-ossp";
EOSQL