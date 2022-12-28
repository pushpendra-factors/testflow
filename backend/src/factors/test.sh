cd $GOPATH/src/factors
mysql -h 127.0.0.1 --port 3306 -u root -pdbfactors123 << MY_QUERY
DROP DATABASE factors;
source migrations/db/memsql/1_create_schema.sql;
MY_QUERY
cd tests
go test ./... | grep FAIL: > fail.txt