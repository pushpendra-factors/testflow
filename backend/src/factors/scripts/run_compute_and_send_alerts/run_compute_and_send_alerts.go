package main

import (
	C "factors/config"
	T "factors/task"
	taskWrapper "factors/task/task_wrapper"
	U "factors/util"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type message struct {
	AlertName     string
	AlertType     int
	Operator      string
	ActualValue   float64
	ComparedValue float64
	Value         float64
	DateRange     string
	ComparedTo    string
}

type dateRanges struct {
	from      int64
	to        int64
	prev_from int64
	prev_to   int64
}

//var timezoneString U.TimeZoneString

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
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")
	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	isWeeklyEnabled := flag.Bool("weekly_enabled", false, "")
	isMonthlyEnabled := flag.Bool("monthly_enabled", false, "")
	isQuarterlyEnabled := flag.Bool("quarterly_enabled", false, "")
	projectIdFlag := flag.String("project_id", "", "Comma separated list of project ids to run")
	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")
	enableDryRunAlerts := flag.Bool("dry_run_alerts", false, "")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	allowProfilesGroupSupport := flag.String("allow_profiles_group_support", "", "")
	enableOptimisedFilterOnProfileQuery := flag.Int("enable_optimised_filter_on_profile_query",
		0, "Enables filter optimisation logic for profiles query.")
	enableOptimisedFilterOnEventUserQuery := flag.Int("enable_optimised_filter_on_event_user_query",
		0, "Enables filter optimisation logic for events and users query.")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}
	defer U.NotifyOnPanic("Script#run_alerts", *env)
	appName := "run_alerts"
	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:                      *primaryDatastore,
		AWSKey:                                *awsAccessKeyId,
		AWSSecret:                             *awsSecretAccessKey,
		AWSRegion:                             *awsRegion,
		EmailSender:                           *factorsEmailSender,
		EnableDryRunAlerts:                    *enableDryRunAlerts,
		RedisHostPersistent:                   *redisHostPersistent,
		RedisPortPersistent:                   *redisPortPersistent,
		AllowProfilesGroupSupport:             *allowProfilesGroupSupport,
		EnableOptimisedFilterOnProfileQuery:   *enableOptimisedFilterOnProfileQuery != 0,
		EnableOptimisedFilterOnEventUserQuery: *enableOptimisedFilterOnEventUserQuery != 0,
	}
	defaultHealthcheckPingID := C.HealthcheckComputeAndSendAlertsPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	C.InitConf(config)
	C.InitSenderEmail(C.GetFactorsSenderEmail())
	C.InitMailClient(config.AWSKey, config.AWSSecret, config.AWSRegion)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()
	//Initialized configs

	query := "select DISTINCT(project_id) from alerts where is_deleted = false;"
	rows, err := db.Raw(query).Rows()
	if err != nil {
		log.Fatal(err)
	}
	allDistinctProjectid := make([]int64, 0)
	for rows.Next() {
		var projectId int64
		rows.Scan(&projectId)
		allDistinctProjectid = append(allDistinctProjectid, projectId)
	}
	runAllProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIdFlag, "")
	projectIdsArray := make([]int64, 0)
	if runAllProjects || *projectIdFlag == "" {
		projectIdsArray = append(projectIdsArray, allDistinctProjectid...)
	} else {
		for projectId := range projectIdsToRun {
			projectIdsArray = append(projectIdsArray, projectId)
		}
	}
	configs := make(map[string]interface{})

	if *isWeeklyEnabled {
		configs["modelType"] = T.ModelTypeWeek
		status := taskWrapper.TaskFuncWithProjectId("ComputeAndSendAlertWeekly", *lookback, projectIdsArray, T.ComputeAndSendAlerts, configs)
		log.Info(status)
		if status["err"] != nil {
			C.PingHealthcheckForFailure(healthcheckPingID, status)
		}
		C.PingHealthcheckForSuccess(healthcheckPingID, status)
	}

	if *isMonthlyEnabled {
		configs["modelType"] = T.ModelTypeMonth
		status := taskWrapper.TaskFuncWithProjectId("ComputeAndSendAlertMonthly", *lookback, projectIdsArray, T.ComputeAndSendAlerts, configs)
		log.Info(status)
		if status["err"] != nil {
			C.PingHealthcheckForFailure(healthcheckPingID, status)
		}
		C.PingHealthcheckForSuccess(healthcheckPingID, status)
	}

	if *isQuarterlyEnabled {
		configs["modelType"] = T.ModelTypeQuarter
		status := taskWrapper.TaskFuncWithProjectId("ComputeAndSendAlertQuarterly", *lookback, projectIdsArray, T.ComputeAndSendAlerts, configs)
		log.Info(status)
		if status["err"] != nil {
			C.PingHealthcheckForFailure(healthcheckPingID, status)
		}
		C.PingHealthcheckForSuccess(healthcheckPingID, status)
	}

}
