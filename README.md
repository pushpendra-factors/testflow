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

## etcd Setup (Homebrew):
```
brew install etcd
brew services start etcd
brew services stop etcd
brew services restart etcd
```
or run it in a new terminal tab using (logs are easily accessible using this approach)
```
etcd 
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
export PATH_TO_FACTORS=~/repos
export GOPATH=$PATH_TO_FACTORS/factors/backend
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
mkdir -p /usr/local/var/factors/config
mkdir -p /usr/local/var/factors/geolocation_data
export PATH_TO_FACTORS=~/repos
cp $PATH_TO_FACTORS/geolocation_data/GeoLite2-City.mmdb /usr/local/var/factors/geolocation_data
cp $GOPATH/src/factors/config/subdomain_login_config.json /usr/local/var/factors/config
./app
or
./app --env=development --api_http_port=8080 --etcd=localhost:2379 --db_host=localhost --db_port=5432 --db_user=autometa --db_name=autometa --db_pass=@ut0me7a --geo_loc_path=/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb --subdomain_enabled=true --subdomain_conf_path=/usr/local/var/factors/config/subdomain_login_config.json
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
export PATH_TO_FACTORS=~/repos
cd $PATH_TO_FACTORS/factors/frontend
npm install
npm run dev
```
* Frontend available at localhost:3000  (API assumed to be served from localhost:8080)
* Create production build by running `npm run build-prod` (Build will be on `dist` dir)
* Serve production build by running `npm run serve-prod` (Creates and serves a new production build).

## Pattern Server
* Setup
```
export PATH_TO_FACTORS=~/repos
export GOPATH=$PATH_TO_FACTORS/factors/backend
export ETCDCTL_API=3

mkdir -p /usr/local/var/factors/local_disk
mkdir -p /usr/local/var/factors/cloud_storage/metadata
touch /usr/local/var/factors/cloud_storage/metadata/version1.txt

etcdctl put /factors/metadata/project_version_key version1
```
* Build
```
go build -o $GOPATH/bin/pattern-app $GOPATH/src/factors/pattern_server/cmd/pattern-app.go
```
* Run
```
cd $GOPATH/bin
./pattern-app
Config can be passed using flags

./pattern-app --env=development --ip=127.0.0.1 --ps_rpc_port=8100 --etcd=localhost:2379 --disk_dir=/usr/local/var/factors/local_disk --bucket_name=/usr/local/var/factors/cloud_storage
```
## Bootstrapping sample data, Building and serving model.
* Start server on 8080.
* Using Localytics challenge data. (https://github.com/localytics/data-viz-challenge)  (https://medium.com/@aabraahaam/localytics-data-visualization-challengue-81ed409471e)

```
export PATH_TO_FACTORS=~/repos
cd $PATH_TO_FACTORS/factors/misc/ingest_events/src
export GOPATH=$PATH_TO_FACTORS/factors/misc/ingest_events
mkdir /usr/local/var/factors/localytics_data
git clone https://github.com/localytics/data-viz-challenge.git  /usr/local/var/factors/localytics_data

go run ingest_localytics_events.go --input_file=/usr/local/var/factors/localytics_data/data.json --server=http://localhost:8080
```

* Note \<projectId\> from the last line of the stdout of the script.
  
```
export PATH_TO_FACTORS=~/repos (path to github code)
cd ../../../backend/src/factors/scripts/
export GOPATH=$PATH_TO_FACTORS/factors/backend
go run run_pull_events.go --project_id=<projectId> --model_type=monthly --end_time=1396310326 --local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp 
--bucket_name=/usr/local/var/factors/cloud_storage
or
go run run_pull_events.go --project_id=<projectId> --start_time=1393632004 --end_time=1396310326 --local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp --bucket_name=/usr/local/var/factors/cloud_storage
```
* Note \<modelId\> from the last line of the stdout of the script.


* Check output file at /usr/local/var/factors/cloud_storage/projects/\<projectId\>models/\<modelId\>/events_<modelId>.txt

```
go run run_pattern_mine.go --env=development --etcd=localhost:2379 --local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp --s3_region=us-east-1 --s3=/usr/local/var/factors/cloud_storage --num_routines=3 --project_id=<projectId> --mode
or
go run run_pattern_mine.go --project_id=<projectId> --model_id=<modelId>
```

* Run sequential model build for all intervals.

```bash
# build for all projects
go run run_build_seq.go --etcd=localhost:2379 --local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp --bucket_name=/usr/local/var/factors/cloud_storage
```

```bash
# build for a specific project
go run run_build_seq.go --etcd=localhost:2379 --local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp --bucket_name=/usr/local/var/factors/cloud_storage --project_id=1
```


* Verify factors in action on Frontend.

## Setting up debugging in VSCode

* Open VSCode, Click `Debug` on top meu bar and choose `Add Configuration`.
* Choose `Go` on the environments list and it will open a file called `launch.json`.
* Copy and replace the content with below one and save. (Note: Change the `Program` path).

```
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "remotePath": "",
            "port": 2345,
            "host": "127.0.0.1",
            "program": "<path_to_gitub_code>/factors/backend/src/factors/app/app.go",
            "env": {},
            "args": [],
            "showLog": true
        }
    ]
}
```
* Choose a debug point on the app. 
* Click `Debug` and choose `Start Debugging`.
* Check the `DEBUG CONSOLE` on the bottom terminal pane.

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

### Setup, Build, Serve and Test

* Setup
```
export PATH_TO_FACTORS=~/repos (path to github code)
cd $PATH_TO_FACTORS/factors/sdk/javascript
npm install
```

* Build (development)
```
npm run build-dev
```

* Build (production)
```bash
# This build will not have tests.
npm run build-prod
```

* Build (Loader)
```
npm run build-loader
```

* Serve (development)
```
npm run serve-dev
```

* Dev Build will be served on http://localhost:8090/factors.dev using webpack.

* Serve (production)
```bash
# This build will not have tests. 
npm run serve-prod
```

* Prod build will be served on http://localhost:8090/factors.html using loader script.

* Validate SDK installation
```javascript
// Run below command on browser dev console.
factors.isInstalled();
```

* Tests.

```javascript
// Running all tests
factors.test.run() 

// Running specific suite
factors.test.runPrivateMethodsSuite()
factors.test.runPublicMethodsSuite()

// Running specific test method
factors.test.SuitePublicMethod.testIdentifyWithoutUserCookie()

```

## Setup and test token login

* Add below 2 entries to end of the file `/etc/hosts`
```
127.0.0.1       sample4ecom.factors-dev.ai
127.0.0.1       unauthorized.factors-dev.ai
```

* Copy subdomain_login_config.json to factors config.
```
cp  $GOPATH/src/factors/config/subdomain_login_config.json /usr/local/var/factors/config
```

* Map the ecommerce sample project's id to `sample4ecom`.

* Test valid subdomain login
```
curl -i -X GET http://sample4ecom.factors-dev.ai:8080/projects/<project_id_of_sample4ecom>/users
```

* Test invalid subdomain login
```
curl -i -X GET http://unauthorized.factors-dev.ai:8080/projects/1/users
```

