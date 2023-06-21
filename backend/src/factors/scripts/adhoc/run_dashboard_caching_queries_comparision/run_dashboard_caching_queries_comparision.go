package main

import (
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	// overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	projectIDFlag := flag.String("project_id", "", "Comma separated project ids to run for. * to run for all")
	flag.Parse()

	config := &C.Configuration{
		AppName: *overrideAppName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     *overrideAppName,
		},
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		SentryDSN:           *sentryDSN,
	}
	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize DB")
	}
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	C.InitSentryLogging(config.SentryDSN, config.AppName)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	projectIDList := store.GetStore().GetProjectsToRunForIncludeExcludeString(*projectIDFlag, "")

	for _, projectID := range projectIDList {
		dashboardUnits, errCode := store.GetStore().GetDashboardUnitsForProjectID(projectID)

		if errCode != http.StatusFound || len(dashboardUnits) == 0 {
			log.WithField("projectID", projectID).
				Warn("Failed to get dashboard unit.")
			continue
		}

		for _, dashboardUnit := range dashboardUnits {
			queryClass, queryInfo, errMsg := store.GetStore().GetQueryAndClassFromDashboardUnit(&dashboardUnit)
			if errMsg != "" {
				log.WithField("projectID", projectID).WithField("dashboardUnit", dashboardUnit.ID).
					Error("failed to get query class")
				continue
			}

			if queryClass == "events" {
				log.WithField("projectID", projectID).WithField("dashboardUnit", dashboardUnit.ID).
					Error("Starting for dashboard unit")
				queryJSON := queryInfo.Query
				var query model.QueryGroup
				U.DecodePostgresJsonbToStructType(&queryJSON, &query)
				lengthOfGBPFromDB := len(query.Queries[0].GroupByProperties)
				if lengthOfGBPFromDB != 0 {

					keys := redisScanAndGetDashboardProjectKeys(projectID, dashboardUnit.ID)

					for _, key := range keys {
						response, err := redisGET(key)
						log.WithField("projectID", projectID).WithField("dashboardUnit.ID", dashboardUnit.ID).Warn(err)

						var cacheResult model.DashboardCacheResult
						err = json.Unmarshal([]byte(response), &cacheResult)
						if err != nil {
							log.WithField("projectID", projectID).WithField("dashboardUnit.ID", dashboardUnit.ID).
								Warn("Failed during unmarshal cache Result.")
						}

						// eventResult := cacheResult.Result.(model.ResultGroup)
						// lenOfGBTFromResult := len(eventResult.Results[0].Meta.Query.GroupByProperties)
						eventResult := cacheResult.Result.(map[string]interface{})
						resultGroup := eventResult["result_group"].([]interface{})
						singleResult := resultGroup[0].(map[string]interface{})
						meta := singleResult["meta"].(map[string]interface{})
						query := meta["query"].(map[string]interface{})
						gbp := query["gbp"].([]interface{})
						lenOfGBTFromResult := len(gbp)
						if lengthOfGBPFromDB != lenOfGBTFromResult {
							log.WithField("key", key).WithField("projectID", projectID).WithField("dashboardUnit", dashboardUnit.ID).
								Warn(" There is a difference in queries.")
						}
					}
				}
			}
		}
	}
}

func redisScanAndGetDashboardProjectKeys(projectID int64, dashboardUnitID int64) []string {
	redisConn := C.GetCacheRedisPersistentConnection()
	defer redisConn.Close()
	cursor := 0
	pattern := fmt.Sprintf("dashboard:query:*pid:%d*:duid:%d*", projectID, dashboardUnitID)
	keys := []string{}

	for {
		res, err := redis.Values(redisConn.Do("SCAN", cursor, "MATCH", pattern, "COUNT", 1000))
		if err != nil {
			log.WithError(err).Error("scan failed")
		}

		cursor, _ = redis.Int(res[0], nil)
		k, _ := redis.Strings(res[1], nil)
		keys = append(keys, k...)

		if cursor == 0 {
			break
		}
	}
	return keys
}

func redisGET(input string) (string, error) {
	if input == "" {
		return "", errors.New("key not found")
	}

	redisConn := C.GetCacheRedisPersistentConnection()
	defer redisConn.Close()

	return redis.String(redisConn.Do("GET", input))
}
