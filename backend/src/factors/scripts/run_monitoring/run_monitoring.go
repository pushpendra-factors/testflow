package main

import (
	C "factors/config"
	"factors/integration"
	"factors/model/store/postgres"
	"factors/sdk"
	"factors/util"
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type SlowQueries struct {
	Runtime                  float64 `json:"runtime"`
	Query                    string  `json:"query"`
	Pid                      int64   `json:"pid"`
	Usename                  string  `json:"usename"`
	ApplicationName          string  `json:"application_name"`
	VacuumProgressPercentage float64 `json:"vacuum_progress_percentage,omitempty"`
	VacuumPhase              string  `json:"vacuum_phase,omitempty"`
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	slowQueriesThreshold := flag.Int("slow_queries_threshold", 50, "Threshold to report slow queries alert")
	sdkQueueThreshold := flag.Int("sdk_queue_threshold", 10000, "Threshold to report sdk queue size")
	integrationQueueThreshold := flag.Int("integration_queue_threshold", 1000, "Threshold to report integration queue size")
	delayedTaskThreshold := flag.Int("delayed_task_threshold", 1000, "Threshold to report delayed task size")

	killSlowQueries := flag.Bool("kill_slow_queries", false, "Kill slow queries. TO BE USED WITH CAUTION")

	flag.Parse()
	taskID := "monitoring_job"
	healthcheckPingID := C.HealthcheckMonitoringJobPingID
	defer C.PingHealthcheckForPanic(taskID, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName:            taskID,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  taskID,
		},
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
			AppName:  taskID,
		},
		PrimaryDatastore: *primaryDatastore,
		QueueRedisHost:   *queueRedisHost,
		QueueRedisPort:   *queueRedisPort,
	}

	C.InitConf(config.Env)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Panic("Failed to initalize db.")
	}
	db := C.GetServices().Db
	defer db.Close()

	err = C.InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Panic("Failed to initalize queue client.")
	}
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	sqlAdminSlowQueries, factorsSlowQueries, err := postgres.GetStore().RunMonitoringQuery()
	if err != nil {
		log.WithError(err).Panic("Failed to run monitoring query.")
	}

	if *killSlowQueries && len(factorsSlowQueries) > 0 {
		var killQuery string
		for _, slowQuery := range factorsSlowQueries {
			killQuery = killQuery + fmt.Sprintf("SELECT pg_cancel_backend(%d);", slowQuery.Pid)
		}
		_, err = db.Raw(killQuery).Rows()
		if err != nil {
			log.WithError(err).Panic("Failed to kill slow queries")
		}
		util.NotifyThroughSNS(taskID, *env, fmt.Sprintf("Killed %d slow queries. %v", len(factorsSlowQueries), factorsSlowQueries))
		return
	}

	if len(factorsSlowQueries) > *slowQueriesThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckDatabaseHealthPingID,
			fmt.Sprintf("Slow query count %d exceeds threshold of %d", len(factorsSlowQueries), *slowQueriesThreshold))
	}

	queueClient := C.GetServices().QueueClient
	delayedTaskCount, err := queueClient.GetBroker().GetDelayedTasksCount()
	if err != nil {
		log.WithError(err).Panic("Failed to get delayed task count from redis")
	}
	if delayedTaskCount > *delayedTaskThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Delayed task count %d exceeds threshold of %d", delayedTaskCount, *delayedTaskThreshold))
	}

	sdkQueueLength, err := queueClient.GetBroker().GetQueueLength(sdk.RequestQueue)
	if err != nil {
		log.WithError(err).Panic("Failed to get sdk_request_queue length")
	}
	if sdkQueueLength > *sdkQueueThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("SDK queue length %d exceeds threshold of %d", sdkQueueLength, *sdkQueueThreshold))
	}

	integrationQueueLength, err := queueClient.GetBroker().GetQueueLength(integration.RequestQueue)
	if err != nil {
		log.WithError(err).Panic("Failed to get integration_request_queue length")
	}
	if integrationQueueLength > *integrationQueueThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Integration queue length %d exceeds threshold of %d", integrationQueueLength, *integrationQueueThreshold))
	}

	tableSizes := postgres.GetStore().CollectTableSizes()

	slowQueriesStatus := map[string]interface{}{
		"factorsSlowQueries":       factorsSlowQueries[:util.MinInt(5, len(factorsSlowQueries))],
		"sqlAdminSlowQueries":      sqlAdminSlowQueries[:util.MinInt(5, len(sqlAdminSlowQueries))],
		"factorsSlowQueriesCount":  len(factorsSlowQueries),
		"sqlAdminSlowQueriesCount": len(sqlAdminSlowQueries),
		"delayedTaskCount":         delayedTaskCount,
		"sdkQueueLength":           sdkQueueLength,
		"integrationQueueLength":   integrationQueueLength,
		"tableSizes":               tableSizes,
	}
	C.PingHealthcheckForSuccess(healthcheckPingID, slowQueriesStatus)
}
