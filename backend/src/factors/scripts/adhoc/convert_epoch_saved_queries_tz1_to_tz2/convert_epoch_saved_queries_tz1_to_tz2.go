package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {
	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production")
	projectIDFlag := flag.Int("project_id", 0, "Single project id to run for.")
	currentTimzone := flag.String("current_timezone", "", "current timezone to be input.")
	nextTimezone := flag.String("next_timezone", "", "next timezone to be input.")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	runningForMemsql := flag.Int("running_for_memsql", 0, "Disable routines for memsql.")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	flag.Parse()

	taskID := "saved_queries_timezone_change"
	healthcheckPingID := C.HealthcheckSavedQueriesTimezoneChangePingID
	defer C.PingHealthcheckForPanic(taskID, *envFlag, healthcheckPingID)
	logCtx := log.WithFields(log.Fields{"Prefix": taskID})

	if *envFlag != C.DEVELOPMENT && *envFlag != C.STAGING && *envFlag != C.PRODUCTION {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	}
	_, err1 := time.LoadLocation(*currentTimzone)
	if err1 != nil {
		panic(fmt.Errorf("current Timezone is not a correct one: %s", *currentTimzone))
	}
	_, err2 := time.LoadLocation(*nextTimezone)
	if err2 != nil {
		panic(fmt.Errorf("next Timezone is not a correct one: %s", *nextTimezone))
	}

	config := &C.Configuration{
		AppName: taskID,
		Env:     *envFlag,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		PrimaryDatastore:   *primaryDatastore,
		RedisHost:          *redisHost,
		RedisPort:          *redisPort,
		SentryDSN:          *sentryDSN,
		IsRunningForMemsql: *runningForMemsql,
	}
	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		logCtx.WithError(err).Panic("Failed to initialize DB")
	}
	C.KillDBQueriesOnExit()
	C.InitRedisPersistent(config.RedisHost, config.RedisPort)

	C.InitSentryLogging(config.SentryDSN, config.AppName)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	if *primaryDatastore != "memsql" {
		panic("Input primary Datastore is not memsql.")
	}
	db := C.GetServices().Db
	failures := make([]interface{}, 0)

	savedQueries, statusCode := getSavedQueries(int64(*projectIDFlag))
	if statusCode != http.StatusFound {
		log.WithField("statusCode", statusCode).Warn("Failed in getting saved queries.")
		C.PingHealthcheckForFailure(C.HealthcheckSavedQueriesTimezoneChangePingID, "Failed while getting Saved queries.")
		panic(fmt.Errorf("stopping after saved Queries fetch"))
	}
	log.WithField("Count of savedQueries fetched", len(savedQueries)).Warn("Count of saved Queries.")
	count := 0
	for _, savedQuery := range savedQueries {
		transformedSavedQuery, errMsg := convertSavedQuery(savedQuery, *currentTimzone, *nextTimezone)
		if errMsg != "" {
			errObject := map[string]interface{}{
				"query":  savedQuery,
				"errMsg": errMsg,
			}
			failures = append(failures, errObject)
			continue
		} else {
			transformedSavedQuery.UpdatedAt = gorm.NowFunc()
			transformedSavedQuery.Converted = true
			if err := db.Save(&transformedSavedQuery).Error; err != nil {
				log.WithField("err", err).WithField("saved Query failed", savedQuery).Warn("Failed in converting savedQuery")
			} else {
				count++
			}
		}
	}
	if len(failures) != 0 {
		C.PingHealthcheckForFailure(C.HealthcheckSavedQueriesTimezoneChangePingID, failures)
	} else {
		C.PingHealthcheckForSuccess(C.HealthcheckSavedQueriesTimezoneChangePingID, "")
	}
	log.WithField("Count of successfully saved - savedQueries.", count).Warn("Count of saved Queries.")
}

func getSavedQueries(projectID int64) ([]model.Queries, int) {
	storeSelected := store.GetStore()
	return storeSelected.GetAllNonConvertedQueries(projectID)
}

func convertSavedQuery(savedQuery model.Queries, currentTimezone string, nextTimezone string) (model.Queries, string) {
	storeSelected := store.GetStore()
	var dupBaseQuery model.Queries
	U.DeepCopy(&savedQuery, &dupBaseQuery)

	queryClass, errMsg := storeSelected.GetQueryClassFromQueries(savedQuery)
	if errMsg != "" {
		return model.Queries{}, errMsg
	}
	baseQuery, err := model.DecodeQueryForClass(savedQuery.Query, queryClass)
	if err != nil {
		return model.Queries{}, err.Error()
	}
	baseQuery.ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone)
	decodedQuery, err2 := U.EncodeStructTypeToPostgresJsonb(baseQuery)
	if err2 != nil {
		return model.Queries{}, err.Error()
	}
	dupBaseQuery.Query = *decodedQuery
	return dupBaseQuery, ""
}
