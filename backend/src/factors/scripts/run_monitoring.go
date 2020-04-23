package main

import (
	C "factors/config"
	"factors/integration"
	"factors/sdk"
	"factors/util"
	"flag"

	log "github.com/sirupsen/logrus"
)

type SlowQueries struct {
	Runtime int64  `json:"runtime"`
	Query   string `json:"query"`
	Pid     int64  `json:"pid"`
}

const taskID = "Task#Monitoring"

func main() {
	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")

	flag.Parse()

	config := &C.Configuration{
		AppName: "monitoring",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		QueueRedisHost: *queueRedisHost,
		QueueRedisPort: *queueRedisPort,
	}

	C.InitConf(config.Env)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.Fatal("Failed to run slow queries. Init failed.")
	}

	err = C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).Fatal("Failed to initalize db.")
	}

	err = C.InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Fatal("Failed to initalize queue client.")
	}

	slowQueries := make([]SlowQueries, 0, 0)

	db := C.GetServices().Db
	defer db.Close()

	queryStr := `SELECT EXTRACT(epoch from (now() - query_start)) as runtime,query, pid FROM  pg_stat_activity` +
		` WHERE EXTRACT(epoch from (now() - query_start)) > 120 AND state = 'active' AND query NOT ILIKE '%pg_stat_activity%'` +
		` ORDER BY runtime DESC LIMIT 10`
	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get slow queries from pg_stat_activity")
	}

	for rows.Next() {
		var slowQuery SlowQueries
		if err := db.ScanRows(rows, &slowQuery); err != nil {
			log.WithError(err).Error("Failed to scan slow queries from db.")
		}

		slowQueries = append(slowQueries, slowQuery)
	}

	queueClient := C.GetServices().QueueClient
	delayedTaskCount, err := queueClient.GetBroker().GetDelayedTasksCount()
	if err != nil {
		log.WithError(err).Error("Failed to get delayed task count from redis")
	}

	sdkQueueLength, err := queueClient.GetBroker().GetQueueLength(sdk.RequestQueue)
	if err != nil {
		log.WithError(err).Error("Failed to get sdk_request_queue length")
	}

	integrationQueueLength, err := queueClient.GetBroker().GetQueueLength(integration.RequestQueue)
	if err != nil {
		log.WithError(err).Error("Failed to get integration_request_queue length")
	}

	slowQueriesStatus := map[string]interface{}{
		"slowQueries":            slowQueries,
		"delayedTaskCount":       delayedTaskCount,
		"sdkQueueLength":         sdkQueueLength,
		"integrationQueueLength": integrationQueueLength,
	}

	if *env == "development" {
		log.Info(slowQueriesStatus)
	} else {
		if len(slowQueries) > 0 || delayedTaskCount > 1000 ||
			sdkQueueLength > 1000 || integrationQueueLength > 1000 {
			if err := util.NotifyThroughSNS(taskID, *env, slowQueriesStatus); err != nil {
				log.WithError(err).Error("Failed to notify slow queries status.")
			} else {
				log.Info("Notified slow queries status.")
			}
		}
	}
}
