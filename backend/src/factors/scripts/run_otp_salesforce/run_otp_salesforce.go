package main

import (
	C "factors/config"
	IntSalesforce "factors/integration/salesforce"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
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

type EnrichStatus struct {
	Status     []IntSalesforce.Status
	HasFailure bool
	Lock       sync.Mutex
}

func (es *EnrichStatus) AddEnrichStatus(status []IntSalesforce.Status, hasFailure bool) {
	es.Lock.Lock()
	defer es.Lock.Unlock()

	es.Status = append(es.Status, status...)
	if hasFailure {
		es.HasFailure = hasFailure
	}
}

func RunOTPSalesForceForProjects(configs map[string]interface{}) (map[string]interface{}, bool) {

	projectIDList := configs["project_ids"].(string)
	disabledProjectIDList := configs["disabled_project_ids"].(string)
	defaultHealthcheckPingID := configs["health_check_ping_id"].(string)
	overrideHealthcheckPingID := configs["override_healthcheck_ping_id"].(string)
	numProjectRoutines := configs["num_project_routines"].(int)

	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, overrideHealthcheckPingID)

	SalesforceEnabledProjectSettings, errCode := store.GetStore().GetAllSalesforceProjectSettings()
	if errCode != http.StatusFound {
		log.Panic("No projects enabled Salesforce integration.")
	}

	anyFailure := false
	panicError := true
	jobStatus := make(map[string]interface{})
	defer func() {
		if panicError || anyFailure {
			C.PingHealthcheckForFailure(healthcheckPingID, jobStatus)
		} else {
			C.PingHealthcheckForSuccess(healthcheckPingID, jobStatus)
		}
	}()
	allProjects, allowedProjects, disabledProjects := C.GetProjectsFromListWithAllProjectSupport(
		projectIDList, disabledProjectIDList)
	if !allProjects {
		log.WithField("projects", allowedProjects).Info("Running only for the given list of projects.")
	}
	if len(disabledProjects) > 0 {
		log.WithField("excluded_projects", disabledProjectIDList).Info("Running with exclusion of projects.")
	}
	projectIDs := make([]int64, 0, 0)
	for _, settings := range SalesforceEnabledProjectSettings {
		if exists := disabledProjects[settings.ProjectID]; exists {
			continue
		}
		if !allProjects {
			if _, exists := allowedProjects[settings.ProjectID]; !exists {
				continue
			}
		}
		projectIDs = append(projectIDs, settings.ProjectID)
		log.WithFields(log.Fields{"projects": projectIDs}).Info("all project list")
	}
	// Runs enrichment for list of project_ids as batch using go routines.
	batches := U.GetInt64ListAsBatch(projectIDs, numProjectRoutines)
	log.WithFields(log.Fields{"project_batches": batches}).Info("Running for batches.")

	for bi := range batches {
		batch := batches[bi]
		var wg sync.WaitGroup
		for pi := range batch {

			wg.Add(1)
			go IntSalesforce.WorkerForSfOtp(batch[pi], &wg)
		}
		wg.Wait()
	}

	panicError = false
	return jobStatus, true

}

func allowProjectByProjectIDList(projectID int64, allProjects bool, allowedProjects, disabledProjects map[int64]bool) bool {
	return !disabledProjects[projectID] && (allProjects || allowedProjects[projectID])
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

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
	projectIDList := flag.String("project_ids", "*", "List of project_id to run for.")
	disabledProjectIDList := flag.String("disabled_project_ids", "", "List of project_ids to exclude.")
	//enrichOnly := flag.Bool("enrich_only", false, "Run only enrichment.")
	allowedCampaignEnrichmentByProjectID := flag.String("allowed_campaign_enrichment_by_project_id", "", "Campaign enrichment by project_id.")
	useOpportunityAssociationByProjectID := flag.String("use_opportunity_association_by_project_id", "", "Use salesforce associations for opportunity stitching")
	numProjectRoutines := flag.Int("num_project_routines", 1, "Number of project level go routines to run in parallel.")
	useSourcePropertyOverwriteByProjectID := flag.String("use_source_property_overwrite_by_project_id", "", "")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	enableSalesforceGroupsByProjectIDs := flag.String("salesforce_groups_by_project_ids", "", "Enable salesforce groups by projects.")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "", "")
	restrictReusingUsersByCustomerUserId := flag.String("restrict_reusing_users_by_customer_user_id", "", "")
	disableCRMUniquenessConstraintsCheckByProjectID := flag.String("disable_crm_unique_constraint_check_by_project_id", "", "")
	numDocRoutines := flag.Int("num_unique_doc_routines", 1, "Number of unique document go routines per project")
	insertBatchSize := flag.Int("insert_batch_size", 1, "Number of unique document go routines per project")
	//overrideLastSyncTimestamp := flag.Int64("override_last_sync_timestamp", 0, "Override last sync timestamp")
	clearbitEnabled := flag.Int("clearbit_enabled", 0, "To enable clearbit enrichment")
	sixSignalEnabled := flag.Int("six_signal_enabled", 0, "To enable sixSignal enrichment")
	useSalesforceV54APIByProjectID := flag.String("use_salesforce_v54_api_by_project_id", "", "Use v54 api for query salesforce data")
	allowIdentificationOverwriteUsingSourceByProjectID := flag.String("allow_identification_overwrite_using_source_by_project_id", "", "Allow identification overwrite based on request source.")
	IngestionTimezoneEnabledProjectIDs := flag.String("ingestion_timezone_enabled_projects", "", "List of projectIds whose ingestion timezone is enabled.")
	allowSalesforceActivityTaskByProjectID := flag.String("allowed_salesforce_activity_tasks_by_project_ids", "", "Allowed project id for salesforce activity - task")
	allowSalesforceActivityEventByProjectID := flag.String("allowed_salesforce_activity_events_by_project_ids", "", "Allowed project id for salesforce activity - event")
	disallowSalesforceActivityTaskByProjectID := flag.String("disallowed_salesforce_activity_tasks_by_project_ids", "", "Disallowed project id for salesforce activity - task")
	disallowSalesforceActivityEventByProjectID := flag.String("disallowed_salesforce_activity_events_by_project_ids", "", "Disallowed project id for salesforce activity - event")

	flag.Parse()
	defaultAppName := "otp_salesforce_job"

	defaultHealthcheckPingID := C.HealthcheckOTPSalesforcePingID
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
		APIDomain:                              *apiDomain,
		SentryDSN:                              *sentryDSN,
		SalesforceAppID:                        *salesforceAppID,
		SalesforceAppSecret:                    *salesforceAppSecret,
		RedisHost:                              *redisHost,
		RedisPort:                              *redisPort,
		RedisHostPersistent:                    *redisHostPersistent,
		RedisPortPersistent:                    *redisPortPersistent,
		DryRunCRMSmartEvent:                    *dryRunSmartEvent,
		CacheSortedSet:                         *cacheSortedSet,
		AllowedCampaignEnrichmentByProjectID:   *allowedCampaignEnrichmentByProjectID,
		UseOpportunityAssociationByProjectID:   *useOpportunityAssociationByProjectID,
		UseSourcePropertyOverwriteByProjectIDs: *useSourcePropertyOverwriteByProjectID,
		AllowedSalesforceGroupsByProjectIDs:    *enableSalesforceGroupsByProjectIDs,
		CaptureSourceInUsersTable:              *captureSourceInUsersTable,
		RestrictReusingUsersByCustomerUserId:   *restrictReusingUsersByCustomerUserId,
		DisableCRMUniquenessConstraintsCheckByProjectID:    *disableCRMUniquenessConstraintsCheckByProjectID,
		SalesforceBatchInsertBatchSize:                     *insertBatchSize,
		ClearbitEnabled:                                    *clearbitEnabled,
		SixSignalEnabled:                                   *sixSignalEnabled,
		UseSalesforceV54APIByProjectID:                     *useSalesforceV54APIByProjectID,
		AllowIdentificationOverwriteUsingSourceByProjectID: *allowIdentificationOverwriteUsingSourceByProjectID,
		IngestionTimezoneEnabledProjectIDs:                 C.GetTokensFromStringListAsString(*IngestionTimezoneEnabledProjectIDs),
		AllowedSalesforceActivityTasksByProjectIDs:         *allowSalesforceActivityTaskByProjectID,
		AllowedSalesforceActivityEventsByProjectIDs:        *allowSalesforceActivityEventByProjectID,
		DisallowedSalesforceActivityTasksByProjectIDs:      *disallowSalesforceActivityTaskByProjectID,
		DisallowedSalesforceActivityEventsByProjectIDs:     *disallowSalesforceActivityEventByProjectID,
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
			"host": *memSQLHost, "port": *memSQLPort}).Panic("Failed to initialize DB.")
		os.Exit(0)
	}

	db := C.GetServices().Db
	defer db.Close()

	allProjects, allowedProjects, disabledProjects := C.GetProjectsFromListWithAllProjectSupport(
		*projectIDList, *disabledProjectIDList)
	if !allProjects {
		log.WithField("projects", allowedProjects).Info("Running only for the given list of projects.")
	}

	anyFailure := false

	var jobStatus salesforceJobStatus

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

	allowedProjectIDs := make([]int64, 0)
	allowedSalesforceProjectSettings := make(map[int64]*model.SalesforceProjectSettings)
	for i := range salesforceEnabledProjects {
		if !allowProjectByProjectIDList(salesforceEnabledProjects[i].ProjectID, allProjects, allowedProjects, disabledProjects) {
			continue
		}

		if _, exist := blackListedProjectIDs[fmt.Sprintf("%d", salesforceEnabledProjects[i].ProjectID)]; exist {
			continue
		}

		allowedSalesforceProjectSettings[salesforceEnabledProjects[i].ProjectID] = &salesforceEnabledProjects[i]
		allowedProjectIDs = append(allowedProjectIDs, salesforceEnabledProjects[i].ProjectID)
	}

	configsEnrich := make(map[string]interface{})
	configsEnrich["project_ids"] = *projectIDList
	configsEnrich["disabled_project_ids"] = *disabledProjectIDList
	configsEnrich["num_unique_doc_routines"] = *numDocRoutines
	configsEnrich["health_check_ping_id"] = defaultHealthcheckPingID
	configsEnrich["override_healthcheck_ping_id"] = *overrideHealthcheckPingID
	configsEnrich["num_project_routines"] = *numProjectRoutines

	configsDistributer := make(map[string]interface{})
	configsDistributer["health_check_ping_id"] = ""

	var notifyMessage string

	RunOTPSalesForceForProjects(configsEnrich)

	notifyMessage = fmt.Sprintf("enrichment for otp hubspot successful for %s - %s projects.", *projectIDList, *disabledProjectIDList)

	if anyFailure {
		C.PingHealthcheckForFailure(healthcheckPingID, jobStatus)
		return
	}

	C.PingHealthcheckForSuccess(healthcheckPingID, notifyMessage)
}
