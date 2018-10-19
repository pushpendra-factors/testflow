# Development Setup (Mac)

## Golang
https://golang.org/doc/install - go for default installation. Not custom installation.

## Postgresql Setup (Homebrew):
```
brew install postgres
brew tap homebrew/services
brew services start postgresql
brew services stop postgresql
brew services restart postgresql
```

### Create user and database.
```
createuser -d -e -P -r autometa
Enter password for new role:@ut0me7a
Enter it again:@ut0me7a
SELECT pg_catalog.set_config('search_path', '', false)
CREATE ROLE autometa PASSWORD 'md5aa44dccdc9ac7b7a4a2a25e129e95784' NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
```

### Create database using psql.
```
psql -U autometa postgres

postgres=> \conninfo
You are connected to database "postgres" as user "autometa" via socket in "/tmp" at port "5432".

postgres=> CREATE DATABASE autometa;
CREATE DATABASE

postgres=> GRANT ALL PRIVILEGES ON DATABASE autometa TO autometa;
GRANT

postgres=> \l
                                         List of databases
   Name    |     Owner     | Encoding |   Collate   |    Ctype    |        Access privileges        
-----------+---------------+----------+-------------+-------------+---------------------------------
 autometa  | autometa      | UTF8     | en_US.UTF-8 | en_US.UTF-8 | =Tc/autometa                   +
           |               |          |             |             | autometa=CTc/autometa
 postgres  | aravindmurthy | UTF8     | en_US.UTF-8 | en_US.UTF-8 | 
 template0 | aravindmurthy | UTF8     | en_US.UTF-8 | en_US.UTF-8 | =c/aravindmurthy               +
           |               |          |             |             | aravindmurthy=CTc/aravindmurthy
 template1 | aravindmurthy | UTF8     | en_US.UTF-8 | en_US.UTF-8 | =c/aravindmurthy               +
           |               |          |             |             | aravindmurthy=CTc/aravindmurthy
(4 rows)
```

### Connect to db using this command.
`psql -U autometa autometa`

### Setting up uuid extension.

* Get super user

```
dp-mac:factors dinesh$  psql -U autometa autometa
psql (10.5)
Type "help" for help.
autometa=> \du
                                   List of roles
 Role name   |                         Attributes                         | Member of
-------------+------------------------------------------------------------+-----------
 autometa    | Create role, Create DB                                     | {}
 <superuser> | Superuser, Create role, Create DB, Replication, Bypass RLS | {}

 autometa=> \q
```

* Create extension

```
dp-mac:factors dinesh$ psql -U <superuser> autometa
psql (10.4)
Type "help" for help.

autometa=# CREATE EXTENSION "uuid-ossp";
CREATE EXTENSION
```

### Create tables.
```
cd <path_to_githubcode>/factors/backend/src/scripts
export GOPATH=<path_to_githubcode>/factors/backend
go run run_db_create.go
```

## Backend Setup. Compilation, Package management using LiteIDE
* Download LiteIDE
* Unpack into Home Folder
* Move Bin to Applications Folder
* Open Project Folder in LiteIDE (<path_to_githubcode>/factors/backend)
* Go to "Build -> Build Configuration"
* Choose "Use Custom GOPATH". Remove "Inherit System GOPATH" and  "Inherit LITEIDE GOPATH"
* Add this line <path_to_githubcode>/factors/backend
* Dependencies managed using LiteIDe.
* Open File <path_to_githubcode>/factors/backend/src/app/app.go
* Build -> Get to add missing dependencies.
* Build -> CleanAll to remove existing binaries
* Build -> Install to create binary.
```
cd <path_to_githubcode>/factors/backend/bin
mkdir -p /tmp/factors/config
cp ../src/config/config.json /tmp/factors/config
./app --config_filepath=/tmp/factors/config/config.json
```
* Backend available at localhost:8080

## Frontend.
* Download and install Nodejs. https://nodejs.org/en/download/
```
cd <path_to_githubcode>/factors/frontend
npm install
npm serve dev
```
* Frontend available at localhost:3000  (API assumed to be served from localhost:8080)
* Use Atom IDE for editiong React Code.

## Bootstrapping sample data, Building and serving model.
* Start server on 8080.

```
cd <path_to_githubcode>/factors/misc/ingest_events/src
export GOPATH=<path_to_githubcode>/factors/misc/ingest_events
go run ingest_kasandr_events.go --input_file=/factors/sample_data/kasandr/sample_raw_data.csv --server=http://localhost:8080
```

* Note \<projectId\> from the last line of the stdout of the script.
  
```
cd ../../../backend/src/scripts/
export GOPATH=<path_to_githubcode>/factors/backend
mkdir -p /tmp/factors/output
go run run_pull_events.go --project_id=<projectId> --end_time="2016-06-20T00:00:00Z" --output_dir="/tmp/factors/output"
```

* Check output file at /tmp/factors/output/patterns-\<projectId\>-1466380800/events.txt

```
go run run_pattern_mine.go --input_file="/tmp/factors/output/patterns-<projectId>-1466380800/events.txt" --project_id=<projectId> --output_file="/tmp/factors/output/patterns-<projectId>-1466380800/patterns.txt"
```

* Change in /tmp/factors/config/config.json 
* "pattern_files": {} 
*    to
* "pattern_files": {"\<projectId\>": "/tmp/factors/output/patterns-\<projectId\>-1466380800/patterns.txt" }
* Restart server to see factors in action on Frontend.

## Debugging using LiteIDE and Delve
* cd ~/go
* GOPATH=~/go
* go get -u github.com/derekparker/delve/cmd/dlv
* cd src/github.com/derekparker/delve/
* make install
* On LiteIDE:  "Debug" -> "debugger/delve"
* Insert Breakpoints in IDE.
* "Build" -> "Debug"
