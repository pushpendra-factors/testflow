package main

import (
	C "factors/config"
	H "factors/handler"
	IntSalesforce "factors/integration/salesforce"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type salesforceSyncStatus struct {
	Success  []IntSalesforce.ObjectStatus `json:"success"`
	Failures []IntSalesforce.ObjectStatus `json:"failures,omitempty"`
}

type salesforceJobStatus struct {
	SyncStatus           salesforceSyncStatus   `json:"sync_status"`
	EnrichStatus         []IntSalesforce.Status `json:"enrich_status"`
	PropertyDetailStatus []IntSalesforce.Status `json:"property_detail_status"`
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
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	salesforceAppID := flag.String("salesforce_app_id", "", "")
	salesforceAppSecret := flag.String("salesforce_app_secret", "", "")
	apiDomain := flag.String("api_domain", "factors-dev.com:8080", "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")
	dryRunSmartEvent := flag.Bool("dry_run_smart_event", false, "Dry run mode for smart event creation")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")
	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	blacklistEnrichmentByProjectID := flag.String("blacklist_enrichment_by_project_id", "", "Blacklist enrichment by project_id.")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")
	syncOnly := flag.Bool("sync_only", false, "Run only sync.")
	allowedCampaignEnrichmentByProjectID := flag.String("allowed_campaign_enrichment_by_project_id", "", "Campaign enrichment by project_id.")
	useOpportunityAssociationByProjectID := flag.String("use_opportunity_association_by_project_id", "", "Use salesforce associations for opportunity stitching")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	flag.Parse()

	defaultAppName := "salesforce_enrich"
	defaultHealthcheckPingID := C.HealthcheckSalesforceEnrichPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	if *env != "development" && *env != "staging" && *env != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *env))
	}

	if *salesforceAppID == "" || *salesforceAppSecret == "" {
		panic(fmt.Errorf("salesforce_app_id or salesforce_app_secret not recognised"))
	}

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
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
			// Todo: Remove UseSSL after enabling it by environment on all workloads.
			UseSSL:      *env == C.STAGING || *env == C.PRODUCTION,
			Certiifcate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:                     *primaryDatastore,
		APIDomain:                            *apiDomain,
		SentryDSN:                            *sentryDSN,
		SalesforceAppID:                      *salesforceAppID,
		SalesforceAppSecret:                  *salesforceAppSecret,
		RedisHost:                            *redisHost,
		RedisPort:                            *redisPort,
		RedisHostPersistent:                  *redisHostPersistent,
		RedisPortPersistent:                  *redisPortPersistent,
		DryRunCRMSmartEvent:                  *dryRunSmartEvent,
		CacheSortedSet:                       *cacheSortedSet,
		AllowedCampaignEnrichmentByProjectID: *allowedCampaignEnrichmentByProjectID,
		UseOpportunityAssociationByProjectID: *useOpportunityAssociationByProjectID,
	}

	C.InitConf(config)
	C.InitSortedSetCache(config.CacheSortedSet)
	C.InitSalesforceConfig(config.SalesforceAppID, config.SalesforceAppSecret)
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	C.InitSmartEventMode(config.DryRunCRMSmartEvent)
	C.InitPropertiesTypeCache(*enablePropertyTypeFromDB, *propertiesTypeCacheSize, *whitelistedProjectIDPropertyTypeFromDB, *blacklistedProjectIDPropertyTypeFromDB)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *dbHost, "port": *dbPort}).Panic("Failed to initialize DB.")
		os.Exit(0)
	}

	db := C.GetServices().Db
	defer db.Close()

	syncInfo, status := store.GetStore().GetSalesforceSyncInfo()
	if status != http.StatusFound {
		log.Panicf("Failed to get salesforce syncinfo: %d", status)
	}

	var syncStatus salesforceSyncStatus
	var propertyDetailSyncStatus []IntSalesforce.Status
	anyFailure := false
	for pid, projectSettings := range syncInfo.ProjectSettings {
		accessToken, err := IntSalesforce.GetAccessToken(projectSettings, H.GetSalesforceRedirectURL())
		if err != nil {
			log.WithField("project_id", pid).Errorf("Failed to get salesforce access token: %s", err)
			continue
		}

		objectStatus := IntSalesforce.SyncDocuments(projectSettings, syncInfo.LastSyncInfo[pid], accessToken)
		for i := range objectStatus {
			if objectStatus[i].Status != U.CRM_SYNC_STATUS_SUCCESS {
				syncStatus.Failures = append(syncStatus.Failures, objectStatus[i])
				anyFailure = true
			} else {
				syncStatus.Success = append(syncStatus.Success, objectStatus[i])
			}
		}

		failure, propertyDetailSyncStatus := IntSalesforce.SyncDatetimeAndNumericalProperties(pid, accessToken, projectSettings.InstanceURL)
		if failure {
			anyFailure = true
		}

		propertyDetailSyncStatus = append(propertyDetailSyncStatus, propertyDetailSyncStatus...)
	}

	var jobStatus salesforceJobStatus
	if !*syncOnly {
		projectIDs := strings.Split(*blacklistEnrichmentByProjectID, ",")
		blackListedProjectIDs := make(map[string]bool)
		for i := range projectIDs {
			blackListedProjectIDs[projectIDs[i]] = true
		}

		// salesforce enrich
		salesforceEnabledProjects, status := store.GetStore().GetAllSalesforceProjectSettings()
		if status != http.StatusFound {
			log.Panic("No projects enabled salesforce integration.")
		}

		statusList := make([]IntSalesforce.Status, 0, 0)

		for _, settings := range salesforceEnabledProjects {
			if _, exist := blackListedProjectIDs[fmt.Sprintf("%d", settings.ProjectID)]; exist {
				continue
			}

			status, failure := IntSalesforce.Enrich(settings.ProjectID)
			if failure {
				anyFailure = true
			}

			statusList = append(statusList, status...)
		}
		jobStatus.EnrichStatus = statusList
	}

	jobStatus.SyncStatus = syncStatus
	jobStatus.PropertyDetailStatus = propertyDetailSyncStatus

	if anyFailure {
		C.PingHealthcheckForFailure(healthcheckPingID, jobStatus)
		return
	}

	C.PingHealthcheckForSuccess(healthcheckPingID, jobStatus)
}
