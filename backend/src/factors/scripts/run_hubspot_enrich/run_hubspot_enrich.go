package main

import (
	C "factors/config"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"

	IntHubspot "factors/integration/hubspot"

	log "github.com/sirupsen/logrus"
)

type SyncStatus struct {
	Status     []IntHubspot.Status
	HasFailure bool
	Lock       sync.Mutex
}

func (s *SyncStatus) AddSyncStatus(status []IntHubspot.Status, hasFailure bool) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	s.Status = append(s.Status, status...)
	if hasFailure {
		s.HasFailure = hasFailure
	}
}

func syncWorker(projectID uint64, wg *sync.WaitGroup, numDocRoutines int, syncStatus *SyncStatus) {
	defer wg.Done()

	status, hasFailure := IntHubspot.Sync(projectID, numDocRoutines)
	syncStatus.AddSyncStatus(status, hasFailure)
}

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
	memSQLCertificate := flag.String("memsql_cert", "", "")
	memSQLResourcePool := flag.String("memsql_resource_pool", "", "If provided, all the queries will run under the given resource pool")
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
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")

	projectIDList := flag.String("project_ids", "*", "List of project_id to run for.")
	disabledProjectIDList := flag.String("disabled_project_ids", "", "List of project_ids to exclude.")
	numProjectRoutines := flag.Int("num_project_routines", 1, "Number of project level go routines to run in parallel.")
	numDocRoutines := flag.Int("num_unique_doc_routines", 1, "Number of unique document go routines per project")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	flag.Parse()

	if *env != "development" && *env != "staging" && *env != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *env))
	}

	defaultAppName := "hubspot_enrich_job"
	defaultHealthcheckPingID := C.HealthcheckHubspotEnrichPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	// init DB, etcd
	config := &C.Configuration{
		AppName:            appName,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
		},
		MemSQLInfo: C.DBConf{
			Host:         *memSQLHost,
			Port:         *memSQLPort,
			User:         *memSQLUser,
			Name:         *memSQLName,
			Password:     *memSQLPass,
			Certificate:  *memSQLCertificate,
			ResourcePool: *memSQLResourcePool,
			AppName:      appName,
		},
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		SentryDSN:           *sentryDSN,
		DryRunCRMSmartEvent: *dryRunSmartEvent,
		CacheSortedSet:      *cacheSortedSet,
	}

	C.InitConf(config)
	C.InitSortedSetCache(config.CacheSortedSet)

	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 200, 100)
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

	var propertyDetailSyncStatus []IntHubspot.Status
	anyFailure := false

	allProjects, allowedProjects, disabledProjects := C.GetProjectsFromListWithAllProjectSupport(
		*projectIDList, *disabledProjectIDList)
	if !allProjects {
		log.WithField("projects", allowedProjects).Info("Running only for the given list of projects.")
	}

	if len(disabledProjects) > 0 {
		log.WithField("excluded_projects", disabledProjectIDList).Info("Running with exclusion of projects.")
	}

	projectIDs := make([]uint64, 0, 0)
	for _, settings := range hubspotEnabledProjectSettings {
		if exists := disabledProjects[settings.ProjectId]; exists {
			continue
		}

		if !allProjects {
			if _, exists := allowedProjects[settings.ProjectId]; !exists {
				continue
			}
		}

		if C.IsEnabledPropertyDetailByProjectID(settings.ProjectId) {
			log.Info(fmt.Sprintf("Starting sync property details for project %d", settings.ProjectId))

			failure, propertyDetailStatus := IntHubspot.SyncDatetimeAndNumericalProperties(settings.ProjectId, settings.APIKey)
			propertyDetailSyncStatus = append(propertyDetailSyncStatus, propertyDetailStatus...)
			if failure {
				anyFailure = true
			}

			log.Info(fmt.Sprintf("Synced property details for project %d", settings.ProjectId))
		}

		projectIDs = append(projectIDs, settings.ProjectId)
	}

	// Runs enrichment for list of project_ids as batch using go routines.
	batches := U.GetUint64ListAsBatch(projectIDs, *numProjectRoutines)
	syncStatus := SyncStatus{}
	for bi := range batches {
		batch := batches[bi]

		var wg sync.WaitGroup
		for pi := range batch {
			wg.Add(1)
			go syncWorker(batch[pi], &wg, *numDocRoutines, &syncStatus)
		}
		wg.Wait()
	}
	anyFailure = anyFailure || syncStatus.HasFailure

	jobStatus := map[string]interface{}{
		"document_sync":      syncStatus.Status,
		"property_type_sync": propertyDetailSyncStatus,
	}

	if anyFailure {
		C.PingHealthcheckForFailure(healthcheckPingID, jobStatus)
		return
	}

	C.PingHealthcheckForSuccess(healthcheckPingID, jobStatus)
}
