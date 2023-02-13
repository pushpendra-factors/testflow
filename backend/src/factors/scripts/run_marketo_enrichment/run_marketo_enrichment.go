package main

import (
	C "factors/config"
	enrichment "factors/crm_enrichment"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", "development", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
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

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")

	projectIDList := flag.String("project_ids", "*", "List of project_id to run for.")
	disabledProjectIDList := flag.String("disabled_project_ids", "", "List of project_ids to exclude.")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	useSourcePropertyOverwriteByProjectID := flag.String("use_source_property_overwrite_by_project_id", "", "")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "", "")
	restrictReusingUsersByCustomerUserId := flag.String("restrict_reusing_users_by_customer_user_id", "", "")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")
	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	numDocRoutines := flag.Int("num_unique_doc_routines", 1, "Number of unique document go routines per project")
	minSyncTimestamp := flag.Int64("min_sync_timestamp", 0, "Min timstamp from where to process records")
	clearbitEnabled := flag.Int("clearbit_enabled", 0, "To enable clearbit enrichment")
	sixSignalEnabled := flag.Int("six_signal_enabled", 0, "To enable sixSignal enrichment")
	IngestionTimezoneEnabledProjectIDs := flag.String("ingestion_timezone_enabled_projects", "", "List of projectIds whose ingestion timezone is enabled.")

	recordProcessLimit := flag.Int("record_process_limit", 0, "Adding limit for processing records") // By default, pull all records

	flag.Parse()
	if *env != "development" && *env != "staging" && *env != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *env))
	}

	defaultAppName := "marketo_enrich_job"
	defaultHealthcheckPingID := C.HealthcheckMarketoEnrichmentPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	// init DB, etcd
	config := &C.Configuration{
		AppName:            appName,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
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
		ClearbitEnabled:                        *clearbitEnabled,
		SixSignalEnabled:                       *sixSignalEnabled,
		IngestionTimezoneEnabledProjectIDs:     C.GetTokensFromStringListAsString(*IngestionTimezoneEnabledProjectIDs),
	}

	C.InitConf(config)
	C.InitSortedSetCache(config.CacheSortedSet)

	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 200, 100)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
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

	anyFailure := false

	sourceObjectTypeAndAlias, err := model.GetMarketoTypeToAliasMap(model.MarketoDocumentTypeAlias)
	if err != nil {
		log.WithError(err).Panic("Failed to get type alias map.")
	}

	userTypes := map[int]bool{
		model.MarketoDocumentTypeAlias[model.MARKETO_TYPE_NAME_LEAD]: true,
	}

	activityTypes := map[int]bool{
		model.MarketoDocumentTypeAlias[model.MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP]: true,
	}

	sourceConfig, err := enrichment.NewCRMEnrichmentConfig(U.CRM_SOURCE_NAME_MARKETO, sourceObjectTypeAndAlias, userTypes, nil, activityTypes, *recordProcessLimit)
	if err != nil {
		log.WithError(err).Error("Failed to create new crm enrichment config.")
		anyFailure = true
	}

	allProjects, allowedProjects, disabledProjects := C.GetProjectsFromListWithAllProjectSupport(
		*projectIDList, *disabledProjectIDList)
	if !allProjects {
		log.WithField("projects", allowedProjects).Info("Running only for the given list of projects.")
	}

	if len(disabledProjects) > 0 {
		log.WithField("excluded_projects", disabledProjectIDList).Info("Running with exclusion of projects.")
	}

	fivetranIntegrations, err := store.GetStore().GetAllActiveFiveTranMappingByIntegration(model.MarketoIntegration)
	if err != nil {
		log.WithError(err).Error("Failed to get marketo projects for enrichment.")
		anyFailure = true
	}

	propertySyncStatus := make(map[int64][]enrichment.EnrichStatus)
	for i := range fivetranIntegrations {
		projectID := fivetranIntegrations[i].ProjectID

		if exists := disabledProjects[projectID]; exists {
			continue
		}

		if !allProjects {
			if _, exists := allowedProjects[projectID]; !exists {
				continue
			}
		}

		propertyEnrichStatus := enrichment.SyncProperties(projectID, sourceConfig)
		propertySyncStatus[projectID] = propertyEnrichStatus
	}

	enrichStatus := make(map[int64][]enrichment.EnrichStatus)
	for i := range fivetranIntegrations {
		projectID := fivetranIntegrations[i].ProjectID

		if exists := disabledProjects[projectID]; exists {
			continue
		}

		if !allProjects {
			if _, exists := allowedProjects[projectID]; !exists {
				continue
			}
		}

		status := enrichment.Enrich(projectID, sourceConfig, *numDocRoutines, *minSyncTimestamp)
		enrichStatus[projectID] = status
		for _, tableStatus := range status {
			if tableStatus.Status == U.CRM_SYNC_STATUS_FAILURES {
				anyFailure = true
			}
		}
	}

	overAllSyncStatus := map[string]interface{}{
		"property_sync": propertySyncStatus,
		"enrich":        enrichStatus,
	}

	if anyFailure {
		C.PingHealthcheckForFailure(healthcheckPingID, overAllSyncStatus)
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, overAllSyncStatus)
	}
}
