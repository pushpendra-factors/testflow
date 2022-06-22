package main

import (
	"flag"
	"time"

	C "factors/config"
	"factors/model/store"

	leadSquaredEnrich "factors/task/lead_squared_enrich"
	taskWrapper "factors/task/task_wrapper"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

func main() {

	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production.")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")
	useSourcePropertyOverwriteByProjectID := flag.String("use_source_property_overwrite_by_project_id", "", "")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "", "")
	restrictReusingUsersByCustomerUserId := flag.String("restrict_reusing_users_by_customer_user_id", "", "")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")
	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	lookback := flag.Int("lookback", 1, "lookback for job")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	numDocRoutines := flag.Int("num_unique_doc_routines", 1, "Number of unique document go routines per project")
	minSyncTimestamp := flag.Int64("min_sync_timestamp", 0, "Min timstamp from where to process records")
	projectIDList := flag.String("project_ids", "*", "List of project_id to run for.")
	disabledProjectIDList := flag.String("disabled_project_ids", "", "List of project_ids to exclude.")
	flag.Parse()

	appName := "lead_squared_enrich"
	defaultHealthcheckPingID := C.HealthcheckLeadSquaredEnrichPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	defer C.PingHealthcheckForPanic(appName, *envFlag, healthcheckPingID)

	config := &C.Configuration{
		Env: *envFlag,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:                       *primaryDatastore,
		RedisHost:                              *redisHost,
		RedisPort:                              *redisPort,
		RedisHostPersistent:                    *redisHostPersistent,
		RedisPortPersistent:                    *redisPortPersistent,
		SentryDSN:                              *sentryDSN,
		CacheSortedSet:                         *cacheSortedSet,
		UseSourcePropertyOverwriteByProjectIDs: *useSourcePropertyOverwriteByProjectID,
		CaptureSourceInUsersTable:              *captureSourceInUsersTable,
		RestrictReusingUsersByCustomerUserId:   *restrictReusingUsersByCustomerUserId,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}

	err = C.InitDBWithMaxIdleAndMaxOpenConn(*config, 200, 100)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *envFlag,
			"host": *memSQLHost, "port": *memSQLPort}).Panic("Failed to initialize DB.")
	}
	db := C.GetServices().Db
	defer db.Close()

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)
	C.InitPropertiesTypeCache(*enablePropertyTypeFromDB, *propertiesTypeCacheSize, *whitelistedProjectIDPropertyTypeFromDB, *blacklistedProjectIDPropertyTypeFromDB)

	projectIdsArray := make([]uint64, 0)
	mappings, err := store.GetStore().GetAllLeadSquaredEnabledProjects()
	if err != nil {
		C.PingHealthcheckForFailure(healthcheckPingID, "Failed to get LeadSquared Projects")
		return
	}

	allProjects, allowedProjects, disabledProjects := C.GetProjectsFromListWithAllProjectSupport(
		*projectIDList, *disabledProjectIDList)
	if !allProjects {
		log.WithField("projects", allowedProjects).Info("Running only for the given list of projects.")
	}

	for projectID := range mappings {
		if exists := disabledProjects[projectID]; exists {
			continue
		}

		if !allProjects {
			if _, exists := allowedProjects[projectID]; !exists {
				continue
			}
		}

		projectIdsArray = append(projectIdsArray, projectID)
	}

	configs := make(map[string]interface{})
	configs["document_routines"] = *numDocRoutines
	configs["min_sync_timestamp"] = *minSyncTimestamp
	status := taskWrapper.TaskFuncWithProjectId(appName, *lookback, projectIdsArray, leadSquaredEnrich.RunLeadSquaredEnrich, configs)

	anyFailure := false
	for _, valInt := range status {
		var projectEnrichStatus map[string]leadSquaredEnrich.EnrichStatus
		var ok bool
		if projectEnrichStatus, ok = valInt.(map[string]leadSquaredEnrich.EnrichStatus); !ok {
			continue
		}

		for commonStatus := range projectEnrichStatus {
			if commonStatus == U.CRM_SYNC_STATUS_FAILURES {
				anyFailure = true
			}
		}
	}

	log.Info(status)
	if anyFailure {
		C.PingHealthcheckForFailure(healthcheckPingID, status)
		return
	}
	C.PingHealthcheckForSuccess(healthcheckPingID, status)

}
