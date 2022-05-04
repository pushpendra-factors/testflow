#!/bin/bash

#starting redis
docker-compose build redis

#starting etcd
docker-compose build etcd

#starting api
docker-compose build api

#starting patternserver
docker-compose build patternserver

#starting sdkserver
docker-compose build sdkserver

#starting frontend
docker-compose build frontend

#building db and demo data
docker-compose build builddb

#running data-generator
docker-compose build datagen

#running session and rollupcache
docker-compose build session

docker-compose build rollupcache

docker-compose build documentdatagen
