navigate to factors/backend/src/docker-development-setup
./docker-compose-build.sh
docker-compose up
After all the services are up - api/patternserver/frontend/redis/postgres/sdkserver - you can checked this with -> docker ps
Open a new terminal and run the following 4 one by one in the following order
docker-compose up datagen
docker-compose up documentdatagen
docker-compose up session
docker-compose up rollupcache
Then http://factors-dev.com:3000/ should give you the UI local
Username: test@factors.ai Password: Factors123


#################################
Hardcoded model id - 1596524994039
Harcoded projectid - 1
Harcoded agent - 058aa637-5846-4309-b961-267970e6a939
Chunk to be tested - factors/backend/src/docker-development-setup/samplemodeldata - place it here/rebuild/redeploy-patternserver
HardcodedEmail- test@factors.ai
HardcodedPassword- Factors123

#################################
Docker-Commands
docker images -a
docker rmi -f $(docker images -a -q)
docker container list
docker container prune
docker image prune
docker exec -it <containerid> bash
docker-compose logs -f <name of the service>
