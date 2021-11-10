# Development Setup (Mac)

## Install golang, docker, docker-compose

## Local setup
navigate to factors/backend/src/docker-development-setup.
```
./docker-compose-build.sh
docker-compose up
```
After all the services are up - api/patternserver/frontend/redis/postgres/sdkserver - you can checked this with docker ps.
Open a new terminal and run the following 4 one by one in the following order.
```
docker-compose up datagen
docker-compose up documentdatagen
docker-compose up session
docker-compose up rollupcache
```
```
Then http://factors-dev.com:3000/ should give you the UI local
Username: test@factors.ai Password: Factors123
```

## To run datagen for past date
```
export SEED_DATE="2021-11-07T00-00-00.000"
Then run docker-compose up datagen
```

### To get memsql hosted in your box
```
export MEMSQL_LICENSE_KEY=""
export ROOT_PASSWORD=""
```

## Project and user details
```
Hardcoded model id - 1596524994039.
Harcoded projectid - 1.
Harcoded agent - 058aa637-5846-4309-b961-267970e6a939.
Chunk to be tested - factors/backend/src/docker-development-setup/.samplemodeldata - place it here/rebuild/redeploy-patternserver.
HardcodedEmail- test@factors.ai.
HardcodedPassword- Factors123.
```

## Docker commands
```
docker images -a.
docker rmi -f $(docker images -a -q). - To remove all the images
docker rm -f. To remove a particular container
docker container list. - To list all active containers
docker container list --all. - To list all active/inactive containers
docker container prune. - to delete dangling containers
docker image prune.
docker exec -it <containerid> bash. - to get inside the container
docker-compose logs -f <name of the service>.
```
