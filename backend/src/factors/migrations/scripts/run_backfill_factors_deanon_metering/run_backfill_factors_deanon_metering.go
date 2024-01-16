package main

import (
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"math"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	projectIds := flag.String("project_ids", "", "List or projects ids to backfill factors deanon metering.")
	env := flag.String("env", C.DEVELOPMENT, "")

	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	flag.Parse()

	if *env != C.DEVELOPMENT &&
		*env != C.STAGING &&
		*env != C.PRODUCTION {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	config := &C.Configuration{
		AppName: "backfill_factors_deanon_metering",
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Certificate: *memSQLCertificate,
			IsPSCHost:   *isPSCHost,
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
		},
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		PrimaryDatastore:    *primaryDatastore,
	}

	C.InitConf(config)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db.")
	}

	projectIdList := C.GetTokensFromStringListAsUint64(*projectIds)

	for _, projectId := range projectIdList {

		start := time.Now()
		logCtx := log.WithField("project_id", projectId)

		timeZone, statusCode := store.GetStore().GetTimezoneForProject(projectId)
		if statusCode != http.StatusFound {
			logCtx.Error("Failed fetching timezone")
		}
		startTimeStamp, err := U.GetTimestampFromDateTimeAndTimezone("2024-01-01 00:00:00", timeZone)
		if err != nil {
			logCtx.Error(err)
		}

		endTimeStamp := 1704457800 //2024-01-05 18:00:00 IST

		qParams := make([]interface{}, 0, 0)

		qStmnt := "SELECT DISTINCT(JSON_EXTRACT_STRING(events.user_properties,?)) FROM events WHERE project_id=? AND timestamp >= ? AND timestamp <= ?;"
		qParams = append(qParams, U.SIX_SIGNAL_DOMAIN, projectId, startTimeStamp, endTimeStamp)

		rows, tx, err, reqID := store.GetStore().ExecQueryWithContext(qStmnt, qParams)
		if err != nil {
			logCtx.WithError(err).Fatal("Failed to execute query.")
		}

		resultHeaders, resultRows, err := U.DBReadRows(rows, tx, reqID)
		if err != nil {
			logCtx.WithError(err).Fatal("Failed to read rows.")
		}

		logCtx.Info(resultHeaders)
		logCtx.Info("Number of unique domains fetched: ", len(resultRows))
		count := 0
		for _, row := range resultRows {

			domain := fmt.Sprintf("%v", row[0])
			isAdded, err := BackfillSixSignalMonthlyUniqueEnrichmentCount(projectId, domain, timeZone)
			if err != nil {
				logCtx.Error("SetSixSignalMonthlyUniqueEnrichmentCount Failed.")
			} else if isAdded {
				count = count + 1
			}

		}
		logCtx.Info("Successful unique domain meter for projectid :", projectId, " is: ", count)
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		logCtx.Info("Total Time taken: ", latency)

	}
}

func BackfillSixSignalMonthlyUniqueEnrichmentCount(projectId int64, domain string, timeZone U.TimeZoneString) (bool, error) {

	monthYear := "January2024"
	key, err := model.GetSixSignalMonthlyUniqueEnrichmentKey(projectId, monthYear)
	if err != nil {
		return false, err
	}

	isAdded, err := cacheRedis.PFAddPersistent(key, domain, 0)
	return isAdded, err
}
