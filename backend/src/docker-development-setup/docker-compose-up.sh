#!/bin/bash

#starting redis
docker-compose up redis

#starting postgres
docker-compose up postgres

#starting etcd
docker-compose up etcd

#starting api
docker-compose up api

#starting patternserver
docker-compose up patternserver

#starting sdkserver
docker-compose up sdkserver

#starting frontend
docker-compose up frontend

#building db and demo data
docker-compose up builddb

#checking api status (:need to add exit if failed)
curl -v -H "Cookie: _fuid=ZWNhMDNjMDUtYjlmOC00MmE4LWJlM2QtMTIzNTI2NTA4NmFm; factors-sidd=eyJhdSI6IjA1OGFhNjM3LTU4NDYtNDMwOS1iOTYxLTI2Nzk3MGU2YTkzOSIsInBmIjoiTVRZek5EYzVOekk0Tlh4dFZXeFdaWFozYWtaWmJXdGhiSEJHVW5oQ2RWVlRNalJOY1RaR1VUWjJRMk0zY0U5M2JqTlpTME0zYmxoelRVeEtibHA2TmtwTVVEZHpiMFptT1hCeGVHd3lRMDlXTXpacFYydDZRbWM0UFh4OF9lVXFJSnN3R2lQaDdtVHY0aERDcklTdk5jOXo1WTJ6a3FmTFlRbnBZZz09In0%3D" http://localhost:8080/status

#checking projects
curl -v -H "Cookie: _fuid=ZWNhMDNjMDUtYjlmOC00MmE4LWJlM2QtMTIzNTI2NTA4NmFm; factors-sidd=eyJhdSI6IjA1OGFhNjM3LTU4NDYtNDMwOS1iOTYxLTI2Nzk3MGU2YTkzOSIsInBmIjoiTVRZek5EYzVOekk0Tlh4dFZXeFdaWFozYWtaWmJXdGhiSEJHVW5oQ2RWVlRNalJOY1RaR1VUWjJRMk0zY0U5M2JqTlpTME0zYmxoelRVeEtibHA2TmtwTVVEZHpiMFptT1hCeGVHd3lRMDlXTXpacFYydDZRbWM0UFh4OF9lVXFJSnN3R2lQaDdtVHY0aERDcklTdk5jOXo1WTJ6a3FmTFlRbnBZZz09In0%3D" http://localhost:8080/projects

#running data-generator
docker-compose up datagen

#running session and rollupcache
docker-compose up session

docker-compose up rollupcache

#testing pattern server
#fetching models
curl -v -H "Cookie: _fuid=ZWNhMDNjMDUtYjlmOC00MmE4LWJlM2QtMTIzNTI2NTA4NmFm; factors-sidd=eyJhdSI6IjA1OGFhNjM3LTU4NDYtNDMwOS1iOTYxLTI2Nzk3MGU2YTkzOSIsInBmIjoiTVRZek5EYzVOekk0Tlh4dFZXeFdaWFozYWtaWmJXdGhiSEJHVW5oQ2RWVlRNalJOY1RaR1VUWjJRMk0zY0U5M2JqTlpTME0zYmxoelRVeEtibHA2TmtwTVVEZHpiMFptT1hCeGVHd3lRMDlXTXpacFYydDZRbWM0UFh4OF9lVXFJSnN3R2lQaDdtVHY0aERDcklTdk5jOXo1WTJ6a3FmTFlRbnBZZz09In0%3D" http://localhost:8080/projects/1/models

curl -v -d "{\"name\":\"\",\"rule\":{\"st_en\":{\"na\":null,\"pr\":null},\"en_en\":{\"na\":\"www.livspace.com/in/hire-a-designer\",\"pr\":[]},\"gpr\":[],\"vs\":true}}" -H "Cookie: _fuid=ZWNhMDNjMDUtYjlmOC00MmE4LWJlM2QtMTIzNTI2NTA4NmFm; factors-sidd=eyJhdSI6IjA1OGFhNjM3LTU4NDYtNDMwOS1iOTYxLTI2Nzk3MGU2YTkzOSIsInBmIjoiTVRZek5EYzVOekk0Tlh4dFZXeFdaWFozYWtaWmJXdGhiSEJHVW5oQ2RWVlRNalJOY1RaR1VUWjJRMk0zY0U5M2JqTlpTME0zYmxoelRVeEtibHA2TmtwTVVEZHpiMFptT1hCeGVHd3lRMDlXTXpacFYydDZRbWM0UFh4OF9lVXFJSnN3R2lQaDdtVHY0aERDcklTdk5jOXo1WTJ6a3FmTFlRbnBZZz09In0%3D" http://localhost:8080/projects/1/v1/factor -d 'type=singleevent' -d 'model_id=1596524994039'


