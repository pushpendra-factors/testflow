#!/bin/bash
touch /root/.pgpass
# Creating .pgpass file
# Postgres will pick password from this file
# hostname:port:database:username:password
echo $DB_HOST:$DB_PORT:$POSTGRES_DB:$POSTGRES_USER:$POSTGRES_PASSWORD > /root/.pgpass
chmod 600 /root/.pgpass
fname=$(date "+%Y-%b-%e-%H-%M-%S-backup.dump")
# http://zevross.com/blog/2014/06/11/use-postgresqls-custom-format-to-efficiently-backup-and-restore-tables/
# -Fc Indicates pg_dump to create backup in postgres custom format. It is compressed by defaut
pg_dump -Fc -U $POSTGRES_USER -h $DB_HOST $POSTGRES_DB > /root/$fname
gsutil cp /root/$fname gs://$(POSTGRES_BACKUP_BUCKET)/$fname