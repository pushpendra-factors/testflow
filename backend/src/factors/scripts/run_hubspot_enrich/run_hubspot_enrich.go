package main

import (
	C "factors/config"
	"factors/model/store"
	"flag"
	"fmt"
	"net/http"
	"time"

	IntHubspot "factors/integration/hubspot"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", "development", "")

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

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	dryRunSmartEvent := flag.Bool("dry_run_smart_event", false, "Dry run mode for smart event creation")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")
	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	ontableUserPropertiesWriteAllowedProjectIDs := flag.String("ontable_user_properties_allowed_projects",
		"", "List of projects to enable writing to on-table user_properties column.")
	deprecateUserPropertiesTableWriteProjectIDs := flag.String("deprecate_user_properties_table_write_projects",
		"", "List of projects to stop writing to user_properties table.")
	deprecateUserPropertiesTableReadProjectIDs := flag.String("deprecate_user_properties_table_read_projects",
		"", "List of projects for which user_properties table read to be deprecated.")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")

	flag.Parse()

	if *env != "development" && *env != "staging" && *env != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *env))
	}

	taskID := "hubspot_enrich_job"
	healthcheckPingID := C.HealthcheckHubspotEnrichPingID
	defer C.PingHealthcheckForPanic(taskID, *env, healthcheckPingID)

	// init DB, etcd
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
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		SentryDSN:           *sentryDSN,
		DryRunCRMSmartEvent: *dryRunSmartEvent,
		// List of project to enable on-table user_properties write on events and users table.
		OnTableUserPropertiesWriteAllowedProjects: *ontableUserPropertiesWriteAllowedProjectIDs,
		// List of projects to stop writing to user_properties table.
		DeprecateUserPropertiesTableWriteProjects: *deprecateUserPropertiesTableWriteProjectIDs,
		// List of projects to use on-table user_properties for read.
		DeprecateUserPropertiesTableReadProjects: *deprecateUserPropertiesTableReadProjectIDs,
		CacheSortedSet:                           *cacheSortedSet,
	}

	C.InitConf(config)
	C.InitSortedSetCache(config.CacheSortedSet)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *dbHost, "port": *dbPort}).Panic("Failed to initialize DB.")
	}
	db := C.GetServices().Db
	defer db.Close()

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	C.InitSmartEventMode(config.DryRunCRMSmartEvent)
	C.InitPropertiesTypeCache(*enablePropertyTypeFromDB, *propertiesTypeCacheSize, *whitelistedProjectIDPropertyTypeFromDB, *blacklistedProjectIDPropertyTypeFromDB)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	hubspotEnabledProjectSettings, errCode := store.GetStore().GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		log.Panic("No projects enabled hubspot integration.")
	}

	statusList := make([]IntHubspot.Status, 0, 0)
	var propertyDetailSyncStatus []IntHubspot.Status
	anyFailure := false
	for _, settings := range hubspotEnabledProjectSettings {
		if C.IsEnabledPropertyDetailByProjectID(settings.ProjectId) {
			log.Info(fmt.Sprintf("Starting sync property details for project %d", settings.ProjectId))

			failure, propertyDetailStatus := IntHubspot.SyncDatetimeAndNumericalProperties(settings.ProjectId, settings.APIKey)
			propertyDetailSyncStatus = append(propertyDetailSyncStatus, propertyDetailStatus...)
			if failure {
				anyFailure = true
			}

			log.Info(fmt.Sprintf("Synced property details for project %d", settings.ProjectId))
		}

		status, failure := IntHubspot.Sync(settings.ProjectId)
		if failure {
			anyFailure = true
		}

		statusList = append(statusList, status...)
	}

	syncStatus := map[string]interface{}{
		"document_sync":      statusList,
		"property_type_sync": propertyDetailSyncStatus,
	}

	if anyFailure {
		C.PingHealthcheckForFailure(healthcheckPingID, syncStatus)
		return
	}

	C.PingHealthcheckForSuccess(healthcheckPingID, syncStatus)
}
