
# Development Setup Using Docker

https://github.com/Slashbit-Technologies/factors/wiki/Development-server-for-branch-testing


## -------  Below setup on localhost is outdated. Use above Docker setup -------------


## Development Setup (Mac)

## Golang (Version 1.14.x)
https://golang.org/dl/ - go for default installation using installer. Not the custom installation using source.
https://golang.org/dl/go1.14.7.darwin-amd64.pkg - use this default installer for mac.

## Postgresql Setup (Homebrew):
```
brew install postgresql@9.6
brew tap homebrew/services
brew services start postgresql@9.6
brew services stop postgresql@9.6
brew services restart postgresql@9.6
```

Check and add the path to $PATH `/usr/local/Cellar/postgresql@9.6/9.6.20/bin` for postgres binaries.

## MemSQL/Singlestore Setup:

Please follow the instructions on the below doc for local setup.</br>
https://adjoining-bubbler-644.notion.site/MemSQL-Documentation-cb28eefa9bd340ed9f97725010e9fc5a

## Redis Setup (Homebrew):

* Install
```
brew install redis 
```

* Run
```
# This will start serving redis non-persistent instance.
cd $GOPATH/src; make serve-redis
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
```

### Create database using psql.
```
psql -U autometa postgres

postgres=> \conninfo
You are connected to database "postgres" as user "autometa" via socket in "/tmp" at port "5432".

postgres=> SELECT pg_catalog.set_config('search_path', '', false);
 set_config 
------------
 
(1 row)


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

### Setting up uuid extension and pg_trgm extension.

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

* Create extensions

```
dp-mac:factors dinesh$ psql -U <superuser> autometa
psql (10.4)
Type "help" for help.

autometa=# CREATE EXTENSION "uuid-ossp";
CREATE EXTENSION

autometa=# CREATE EXTENSION "pg_trgm";
CREATE EXTENSION
```

## Setup, Build and Run

* Setup
```
echo '127.0.0.1     factors-dev.com' | sudo tee -a /etc/hosts
mkdir ~/repos
cd ~/repos
git clone https://github.com/Slashbit-Technologies/factors.git
export PATH_TO_FACTORS=~/repos
export GOPATH=$PATH_TO_FACTORS/factors/backend
```

* Create tables

**Postgres**
```
cd $PATH_TO_FACTORS/factors/backend/src/factors/scripts/run_db_create
go run run_db_create.go
```

**MemSQL/Singlestore**

* Visit http://localhost:8040   
* Login with the local username and password.   
* Select SQL Editor on the left pan.    
* Paste the contents of the file on `<PATH_TO_REPO>/backend/src/factors/migrations/db/memsql/1_create_schema.sql` on the editor.    
* Select all and Click Run.   

Refer: https://adjoining-bubbler-644.notion.site/MemSQL-Documentation-cb28eefa9bd340ed9f97725010e9fc5a   

* Build
```
export PATH_TO_FACTORS=~/repos
cd $PATH_TO_FACTORS/factors/backend/src
make build-api
```

* Run 
```
export PATH_TO_FACTORS=~/repos
mkdir -p /usr/local/var/factors/config
mkdir -p /usr/local/var/factors/geolocation_data
mkdir -p /usr/local/var/factors/devicedetector_data

cp $PATH_TO_FACTORS/factors/geolocation_data/GeoLite2-City.mmdb /usr/local/var/factors/geolocation_data
cp -R $PATH_TO_FACTORS/factors/devicedetector_data/regexes /usr/local/var/factors/devicedetector_data

cd $PATH_TO_FACTORS/factors/backend/src
make serve-api

or

cd $GOPATH/bin
./app --env=development --api_domain=localhost:8080 --app_domain=localhost:3000  --api_http_port=8080 --etcd=localhost:2379 --db_host=localhost --db_port=5432 --db_user=autometa --db_name=autometa --db_pass=@ut0me7a --geo_loc_path=/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb --aws_region=us-east-1 --aws_key=dummy --aws_secret=dummy --email_sender=support@factors.ai
```

* Backend available at factors-dev.com:8080
* Api documentation for the app server will be available at http://factors-dev.com:8080/swagger/index.html
* To regenerate documentation after any changes, run `make build-api-doc`.

## Managing dependencies with go mod (Optional)

#### Adding new dependency
1. Import the required dependency in of the files inside backend src.
2. To install a specific version, use `go get -u <dependency>@<version>`
3. Run the following two commands:
    ```
    go mod tidy
    go mod vendor
    ```

#### Removing a dependency
1. Remove it from the import statement.
2. Run the following two commands:
    ```
    go mod tidy
    go mod vendor
    ```

## Frontend.
* Download and install Nodejs. https://nodejs.org/en/download/ (Node version 12 or 14.)
```
export PATH_TO_FACTORS=~/repos
cd $PATH_TO_FACTORS/factors/frontend-new
brew install node
npm install
npm run dev
```
* Frontend available at factors-dev.com:3000  (API assumed to be served from factors-dev.com:8080)
* Create production build by running `npm run build-prod` (Build will be on `dist` dir)
* Serve production build by running `npm run serve-prod` (Creates and serves a new production build).

## Pattern Server
* Setup
```
export PATH_TO_FACTORS=~/repos
export GOPATH=$PATH_TO_FACTORS/factors/backend

mkdir -p /usr/local/var/factors/local_disk
mkdir -p /usr/local/var/factors/cloud_storage/metadata
touch /usr/local/var/factors/cloud_storage/metadata/version1.txt

export ETCDCTL_API=3
etcdctl put /factors/metadata/project_version_key version1
```
* Build
```
cd $PATH_TO_FACTORS/factors/backend/src
make build-ps
```
* Run
```
make serve-ps

or

Config can be passed using flags
cd $GOPATH/bin
./pattern-app --env=development --ip=127.0.0.1 --ps_rpc_port=8100 --etcd=localhost:2379 --disk_dir=/usr/local/var/factors/local_disk --bucket_name=/usr/local/var/factors/cloud_storage --chunk_cache_size=5 --event_info_cache_size=10
```

### Data Server
* Build
```
cd $PATH_TO_FACTORS/factors/backend/src
make build-ds
```

* Run
```
make serve-ds

or

$GOPATH/bin/./data-service --env=development --port=8089 --db_host=localhost --db_port=5432 --db_user=autometa --db_name=autometa --db_pass=@ut0me7a --aws_region=us-east-1 --aws_key=dummy --aws_secret=dummy --email_sender=support@factors.ai --error_reporting_interval=300
```

### SDK Server
* Build
```
cd $PATH_TO_FACTORS/factors/backend/src
make build-sdk
```

* Run
```
make serve-sdk
```

### Workers
* Run SDK Request Worker
```
cd $PATH_TO_FACTORS/factors/backend/src/factors/workers/sdk_request
go run process.go
```

* Run Integration Request Worker
```
cd $PATH_TO_FACTORS/factors/backend/src/factors/workers/integration_request
go run process.go
```

### Add Session (Optional)
```
cd $PATH_TO_FACTORS/factors/backend/src/factors/scripts
go run add_session.go --project_ids="<LIST_OF_PROJECT_IDS>"
```

### Adwords Server (Optional)
* Install python3
```
brew install python3

# Note: If you have multiple versions of python. Make sure python refers to python3 and corresponding pip.
```

* Install dependencies
```
cd $PATH_TO_FACTORS/factors/integrations/adwords/service
pip install -r requirements.txt
# Use pip3 in case pip is not found
```

* Run
```
cd $PATH_TO_FACTORS/factors/integrations/service

python app.py --env development --developer_token <ADS_DEVELOPER_TOKEN> --oauth_secret  $(cat <GOOGLE_OAUTH_CLIENT_JSON_FILEPATH>)

or

python app.py --env development --port 8091 --host_url http://localhost:8091 --developer_token <ADS_DEVELOPER_TOKEN> --oauth_secret $(cat <GOOGLE_OAUTH_CLIENT_JSON_FILEPATH>) --api_host_url http://localhost:8080 ----app_host_url http://localhost:3000
```


## Bootstrapping sample data, Building and serving model.

### Ingesting Data into sdk.
* Start the sdk server on 8085.
* Using Localytics challenge data. (https://github.com/localytics/data-viz-challenge)  (https://medium.com/@aabraahaam/localytics-data-visualization-challengue-81ed409471e)

```
export PATH_TO_FACTORS=~/repos
cd $PATH_TO_FACTORS/factors/misc/ingest_events/src
export GOPATH=$PATH_TO_FACTORS/factors/misc/ingest_events
mkdir -p /usr/local/var/factors/localytics_data
git clone https://github.com/localytics/data-viz-challenge.git  /usr/local/var/factors/localytics_data --depth 1

# Create an account by signing into http://factors-dev.com:3000 and activate it by hitting url visibile in backend logs.
# After signing in, get the requestId and token from network request call to /projects.
go get github.com/sirupsen/logrus
go run ingest_localytics_events.go --input_file=/usr/local/var/factors/localytics_data/data.json --server=http://factors-dev.com:8085 --project_id=<project_id> --project_token=<project_token>

```

### Pull the events that needs to be processed into file, and build the models (Optional)
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
go run run_pattern_mine.go --env=development --etcd=localhost:2379 --local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp --s3_region=us-east-1 --s3=/usr/local/var/factors/cloud_storage --num_routines=3 --project_id=<projectId> --model_id=<modelId>
or
go run run_pattern_mine.go --project_id=<projectId> --model_id=<modelId>
```

### Run sequential model build for all intervals.
```bash
# build for all projects
go run run_build_seq.go --etcd=localhost:2379 --local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp --bucket_name=/usr/local/var/factors/cloud_storage
```

```bash
# build for a specific project
go run run_build_seq.go --project_ids=1,2 --model_type=all --look_back_days=30 --etcd=localhost:2379 --local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp --bucket_name=/usr/local/var/factors/cloud_storage
```
* The above look_back_days should be kept to longer value for development setup. For instance - 30000.

* Verify factors in action on Frontend.

### Weekly insights
 ```bash
go run run_weekly_insights.go --weekly_enabled --project_ids=34339 --k=1 --lookback=10
```

* Whatever project_id is supplied, create /projects/<id>/events/w/<date in yyyymmdd format>/events.txt in the following location /usr/local/var/factors/cloud_storage
* <id> - whatever projectid you are running the job with 
* date - create for past 2 sundays since WI always looks for 2 weeks
* events.txt - you can download from the cloudstorage in staging
* Whatever project you are running against, make sure that exist in your local db and there are some events/funnel queries. IF not, run some tests that creates queries and use that project for running WI
* Check output in /usr/local/var/factors/cloud_storage/projects/<id>/weeklyinsights/

### Numerical bucketing
 ```bash
go run run_numerical_bucketing.go --weekly_enabled --project_ids=34339 -lookback=10
```
* Same as weekly insights. Having a week's data is sufficient. Need not have queries/project

### Enabling BigQuery for a project (Optional)
* Make sure to have a project_id with some events. Refer above steps to ingest events data using `ingest_localytics_events.go` script if required.
* Enable archival and bigquery in project_settings for the project.
    ```
    UPDATE project_settings SET archive_enabled=true, bigquery_enabled=true where project_id=<project_id>;
    ```
* Creating events files in cloud_storage. Creates a single file from last pull timestamp till current timestamp.
    - Sample run. Use `--max_lookback_days` for archiving data older than 365 days.
    
        ```
        # Use --all=true to run for all enabled projects.
        cd ${PATH_TO_FACTORS}/factors/backend/src/factors
        go run scripts/run_archive_events.go --project_id=<project_id>
        ```
    - To regenrate events file on dev, delete entries from the scheduled_tasks table for the project_id`
        ```plsql
        DELETE FROM scheduled_tasks WHERE project_id = <project_id> AND task_type='EVENTS_ARCHIVAL';
        ```
* Pushing archive files to BigQuery.
    - Adding BigQuery configuration for a new project. Service account credentials json will be required.
        ```
        cd ${PATH_TO_FACTORS}/factors/backend/src/factors
        go run scripts/run_onboard_to_bigquery.go --project_id=<project_id> --bq_project_id=<bq_project_id> --bq_dataset=<bq_dataset> --bq_credentials_json=<credentials_json_filename>
        ```
    - For pushing data to bigquery.
        ```
        # Use --all=true to run for all enabled projects.
        cd ${PATH_TO_FACTORS}/factors/backend/src/factors
        go run scripts/run_push_to_bigquery.go --project_id=<project_id>
        ```
    - Big query configuration parameters:
        * `bq_project_id`: Project ID from Google Cloud Console. Must be an existing project. For testing, `factors-staging` can be used.
        * `bq_dataset`: Dataset name inside BigQuery. Script will create with the given name if not exists.
        * `bq_credentials_json`: Absolute path pointing to the credentials json file for the service account. This is to authenticate bigquery client and must be generated for the for project provided in `bq_project_id` with at least `BigQuery Data Owner` and `BigQuery Job User` roles assigned.
    - To repush old files to BigQuery (delete table from cloud console clean state is required without duplicate data), update `last_run_at` of `bigquery_documents` table.
    
        ```plsql
        UPDATE bigquery_documents SET last_run_at = 0 WHERE project_id=<project_id>;
        ```

## Running Adwords Sync (Optional)

* Setup
```
cd $PATH_TO_FACTORS/factors/integrations/adwords/scripts
pip install -r requirements.txt
```

* Run
```
python sync.py --developer_token=<ADS_DEVELOPER_TOKEN> --oauth_secret=$(cat <GOOGLE_OAUTH_CLIENT_JSON_FILEPATH>)

# Running sync for specific project.

python sync.py --developer_token=<ADS_DEVELOPER_TOKEN> --oauth_secret=$(cat <GOOGLE_OAUTH_CLIENT_JSON_FILEPATH>) --project_id=1
```

## Setting up debugging in VSCode (Optional)

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

## Debugging using LiteIDE and Delve (Optional)
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

* Add repo path to bash profile
```
export FACTORS_REPO=<path_to_factors>/factors
```

* Setup
```
cd $FACTORS_REPO/sdk/javascript
npm install
```

* Build (development)
```
cd $FACTORS_REPO/sdk/javascript
npm run build-dev
```

* Build (production)
```bash
# This build will not have tests.
cd $FACTORS_REPO/sdk/javascript
npm run build-prod
```

* Build (Loader)
```
cd $FACTORS_REPO/sdk/javascript
npm run build-loader
```

* Serve (development).
```
cd $FACTORS_REPO/sdk/javascript
npm run serve-dev
```

* Dev Build will be served on http://localhost:8090/factors.dev using webpack.

* Serve (production)
```bash
# This build will not have tests. 
cd $FACTORS_REPO/sdk/javascript
npm run serve-prod
```

* Prod build will be served on http://localhost:8090/factors.html using loader script.

* Validate SDK installation. Open browser console and run below command.
```javascript
factors
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

## AMP Page Setup

* Bring up sdk-server. Refer to [SDK Server](#sdk-server).
* Serve JS SDK webpack dev server.
```
cd $FACTORS_REPO/sdk/javascript
npm run serve-dev
```
* On file `FACTORS_REPO/sdk/javascript/factors_amp_test.html`, replace `<PROJECT-TOKEN>` with your local project's token.
* On browser, visit `http://localhost:8090/factors_amp_test.html` and check the browser console for request sent. 
