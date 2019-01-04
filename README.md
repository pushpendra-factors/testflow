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
mkdir -p /tmp/factors/geolocation_data
cp <path_to_github_code>/geolocation_data/GeoLite2-City.mmdb /tmp/factors/geolocation_data
cp $GOPATH/src/factors/config/config.json /tmp/factors/config
cp $GOPATH/src/factors/config/subdomain_login_config.json /tmp/factors/config
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

## Pattern Server
* Setup
```
export GOPATH=<path_to_githubcode>/factors/backend
export ETCDCTL_API=3
etcdctl put /factors/metadata/project_version_key version1
mkdir -p /tmp/factors-dev/metadata
mkdir -p /tmp/factors/metadata
cp $GOPATH/src/factors/patternserver/cmd/project_data.txt /tmp/factors/metadata/version1.txt
cp /tmp/factors/metadata/version1.txt /tmp/factors-dev/metadata/version1.txt
```
* Build
```
go build -o $GOPATH/bin/pattern-app $GOPATH/src/factors/patternserver/cmd/pattern-app.go
```
* Run
```
cd $GOPATH/bin
./pattern-app
Config can be passed using flags

./pattern-app --env=development --ip=127.0.0.1 --port=8100 --etcd=localhost:2379 --disk_dir=/tmp/factors --s3=/tmp/factors-dev --s3_region=us-east-1
```
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
go run run_pull_events.go --project_id=<projectId> --end_time=1465948361 --disk_dir=/tmp/factors --bucket_name=/tmp/factors-dev
* Note \<modelId\> from the last line of the stdout of the script.
```

* Check output file at /tmp/factors/projects/\<projectId\>models/\<modelId\>/events_<modelId>.txt

```
go run run_pattern_mine.go --env=development --etcd=localhost:2379 --disk_dir=/tmp/factors --s3_region=us-east-1 --s3=/tmp/factors-dev --num_routines=3 --project_id=<projectId> --model_id=<modelId>
or
go run run_pattern_mine.go --project_id=<projectId> --model_id=<modelId>
```

* Change in /tmp/factors/config/config.json 
* "project_model_mapping": {} 
*    to
* "project_model_mapping": {"\<projectId\>": "\<modelId\>"}
* Restart server to see factors in action on Frontend.

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
cd <path_to_github_code>/factors/sdk/javascript
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

// Running a specific test
factors.test.Suite.TEST_NAME
```

## Setup and test token login

* Add below 2 entries to end of the file `/etc/hosts`
```
127.0.0.1       sample4ecom.factors-dev.ai
127.0.0.1       unauthorized.factors-dev.ai
```

* Copy subdomain_login_config.json to tmp.
```
cp  $GOPATH/src/factors/config/subdomain_login_config.json /tmp/factors/config
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

# Deployment

## Setup docker - For building image

* Download and run docker installer from [link]: https://hub.docker.com/editions/community/docker-ce-desktop-mac

* Test installation
```
$ docker run hello-world

Unable to find image 'hello-world:latest' locally
latest: Pulling from library/hello-world
ca4f61b1923c: Pull complete
Digest: sha256:ca0eeb6fb05351dfc8759c20733c91def84cb8007aa89a5bf606bc8b315b9fc7
Status: Downloaded newer image for hello-world:latest

Hello from Docker!
This message shows that your installation appears to be working correctly.
...
```

## Pattern server

*  Building image with tag `factors/pattern-server:latest` (For an increamental version, use `-t factors/pattern-server:v\`date +%s\``).
```
cd $GOPATH/src
docker build -f ./Dockerfile.patternserver -t factors/pattern-server:latest .

$ docker image ls
REPOSITORY               TAG                 IMAGE ID            CREATED             SIZE
factors/pattern-server   latest              fc194dc8343e        10 seconds ago      31.6MB
```

* Run image
```
docker run -p 8100:8100 --env=test factors/pattern-server:latest
```

* Debug - opens an interactive `sh` (not `bash`).
```
docker run -it --entrypoint="/bin/sh" factors/pattern-server:latest
```

## (optional) Local minikube and kubectl installation

* Install minikube
```
brew cask install virtualbox minikube
```

* Create and start cluster
```
$ minikube start

Starting local Kubernetes v1.12.4 cluster...
Starting VM...
Downloading Minikube ISO
 178.88 MB / 178.88 MB [============================================] 100.00% 0s
Getting VM IP address...
Moving files into cluster...
Downloading kubeadm v1.12.4
Downloading kubelet v1.12.4
Finished Downloading kubeadm v1.12.4
Finished Downloading kubelet v1.12.4
Setting up certs...
Connecting to cluster...
Setting up kubeconfig...
Stopping extra container runtimes...
Starting cluster components...
Verifying kubelet health ...
Verifying apiserver health ...Kubectl is now configured to use the cluster.
Loading cached images from config file.

Everything looks great. Please enjoy minikube!
```

* Stop cluster
```
minikube stop
```

* Dashboard
```
minikube dashboard
```

* Switching context to minikube cluster (using kubectl with multiple clusters).
```
kubectl config get-contexts
kubectl config use-context minikube
```


