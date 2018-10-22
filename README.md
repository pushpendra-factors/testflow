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

## Setup, Build and Run

* Setup
```
git clone https://github.com/Slashbit-Technologies/factors.git
export GOPATH=<path_to_githubcode>/factors/backend
```

* Create tables
```
cd $GOPATH/factors/backend/src/factors/scripts
go run run_db_create.go
```

* Install godep (mac) `brew install dep`.
* Install all dependencies of the project (Optional, as we have checked in vendor libs already).
```
cd $GOPATH/src/factors
dep ensure
```
* Build
```
go build -o $GOPATH/bin/app $GOPATH/src/factors/app/app.go
```

* Run 
```
cd $GOPATH/bin
mkdir -p /tmp/factors/config
cp $GOPATH/src/factors/config/config.json /tmp/factors/config
./app --config_filepath=/tmp/factors/config/config.json
```

* Backend available at localhost:8080

## Managing dependencies with godep

* Adding new dependency
```
dep ensure -add github.com/<path_to_repo>
```
* Removing a dependency - Remove it from import and run `dep ensure`

## Frontend.
* Download and install Nodejs. https://nodejs.org/en/download/
```
cd <path_to_githubcode>/factors/frontend
npm install
npm run dev
```
* Frontend available at localhost:3000  (API assumed to be served from localhost:8080)
* Use Atom IDE for editiong React Code.

## Bootstrapping sample data, Building and serving model.
* Start server on 8080.

```
cd <path_to_githubcode>/factors/misc/ingest_events/src
export GOPATH=<path_to_githubcode>/factors/misc/ingest_events
go run ingest_kasandr_events.go --input_file=<path_to_githubcode>/factors/sample_data/kasandr/sample_raw_data.csv --server=http://localhost:8080
```

* Note \<projectId\> from the last line of the stdout of the script.
  
```
cd ../../../backend/src/factors/scripts/
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

## Javascript SDK

### Setup, Build and Serve

* Setup
```
cd <path_to_github_code>/factors/sdk/javascript
npm install
```

* Build (development)
```
npm run build-dev
```

* Build (production)
```
npm run build-prod
```

* Serving development build
```
npm run serve-dev
```

* Build will be served on http://localhost:8090/bundle-dev

* Open developer console and run.

```javascript
factors.isInstalled()

// "Factors sdk v0.1 is installed!"
```



