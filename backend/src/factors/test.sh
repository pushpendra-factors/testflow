cd $GOPATH/src/factors
mysql --default-auth mysql_native_password -h 127.0.0.1 --port 3306 -u root -pdbfactors123 << MY_QUERY
DROP DATABASE factors;
source migrations/db/memsql/1_create_schema.sql;
MY_QUERY
cd tests
echo "Running tests..."
go test ./... | grep FAIL: > failed.log
cat failed.log