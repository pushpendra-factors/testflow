package main

import (
	C "factors/config"
	H "factors/handler"
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

func syncWorker(projectID int64, wg *sync.WaitGroup, workerIndex, workerPerProject int, enrichStatus *EnrichStatus, salesforceProjectSettings *model.SalesforceProjectSettings, enrichPullLimit int, enrichRecordProcessLimit int) {
	defer wg.Done()

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "worder_index": workerIndex})
	logCtx.Info("Enrichment started for given project.")

	accessToken, instanceURL, err := IntSalesforce.GetAccessToken(salesforceProjectSettings, H.GetSalesforceRedirectURL())
	if err != nil || accessToken == "" || instanceURL == "" {
		log.WithField("project_id", salesforceProjectSettings.ProjectID).Errorf("Failed to get salesforce access token: %s", err)
		enrichStatus.AddEnrichStatus([]IntSalesforce.Status{{ProjectID: projectID, Message: err.Error()}}, false)
		return
	}

	dataPropertyByType, errCode := IntSalesforce.GetSalesforcePropertiesByDataType(projectID, model.SalesforceDataTypeDate, model.GetSalesforceAllowedObjects(projectID), accessToken, instanceURL)
	if errCode != http.StatusOK {
		log.WithField("project_id", salesforceProjectSettings.ProjectID).Error("Failed to get salesforce date properties.")
		enrichStatus.AddEnrichStatus([]IntSalesforce.Status{{ProjectID: projectID, Message: "Failed to get date properies"}}, true)
		return
	}

	status, hasFailure := IntSalesforce.Enrich(projectID, workerPerProject, dataPropertyByType, enrichPullLimit, enrichRecordProcessLimit)
	enrichStatus.AddEnrichStatus(status, hasFailure)
	logCtx.Info("Processing completed for given project.")
}

func allowProjectByProjectIDList(projectID int64, allProjects bool, allowedProjects, disabledProjects map[int64]bool) bool {
	return !disabledProjects[projectID] && (allProjects || allowedProjects[projectID])
}

func overrideLastSyncTimestampIfRequired(overrideSyncTimestamp int64, syncInfo map[string]int64) map[string]int64 {
	if overrideSyncTimestamp > 0 {

		for typ := range syncInfo {
			syncInfo[typ] = overrideSyncTimestamp
		}
	}

	return syncInfo
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
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
	sentryRollupSyncInSecs := flag.Int("sentry_rollup_sync_in_seconds", 300, "Enables to send errors to sentry in given interval.")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	factorsSixsignalAPIKey := flag.String("factors_sixsignal_api_key", "dummy", "")

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
	projectIDList := flag.String("project_ids", "*", "List of project_id to run for.")
	skippedOtpProjectIDs := flag.String("skipped_otp_project_ids", "", "List of project_id to be skip for otp job.")
	disabledProjectIDList := flag.String("disabled_project_ids", "", "List of project_ids to exclude.")
	enrichOnly := flag.Bool("enrich_only", false, "Run only enrichment.")
	allowedCampaignEnrichmentByProjectID := flag.String("allowed_campaign_enrichment_by_project_id", "", "Campaign enrichment by project_id.")
	useOpportunityAssociationByProjectID := flag.String("use_opportunity_association_by_project_id", "", "Use salesforce associations for opportunity stitching")
	numProjectRoutines := flag.Int("num_project_routines", 1, "Number of project level go routines to run in parallel.")
	useSourcePropertyOverwriteByProjectID := flag.String("use_source_property_overwrite_by_project_id", "", "")

	overrideEnrichHealthcheckPingID := flag.String("enrich_healthcheck_ping_id", "", "Override default enrich healthcheck ping id.")
	overrideSyncHealthcheckPingID := flag.String("sync_healthcheck_ping_id", "", "Override default sync healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	enableSalesforceGroupsByProjectIDs := flag.String("salesforce_groups_by_project_ids", "", "Enable salesforce groups by projects.")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "", "")
	restrictReusingUsersByCustomerUserId := flag.String("restrict_reusing_users_by_customer_user_id", "", "")
	disableCRMUniquenessConstraintsCheckByProjectID := flag.String("disable_crm_unique_constraint_check_by_project_id", "", "")
	numDocRoutines := flag.Int("num_unique_doc_routines", 1, "Number of unique document go routines per project")
	insertBatchSize := flag.Int("insert_batch_size", 1, "Number of unique document go routines per project")
	overrideLastSyncTimestamp := flag.Int64("override_last_sync_timestamp", 0, "Override last sync timestamp")
	clearbitEnabled := flag.Int("clearbit_enabled", 0, "To enable clearbit enrichment")
	sixsignalV1EnabledProjectIDs := flag.String("sixsignal_v1_enabled_projectIds", "", "To enable new sixsignal flow")
	useSalesforceV54APIByProjectID := flag.String("use_salesforce_v54_api_by_project_id", "", "Use v54 api for query salesforce data")
	allowIdentificationOverwriteUsingSourceByProjectID := flag.String("allow_identification_overwrite_using_source_by_project_id", "", "Allow identification overwrite based on request source.")
	IngestionTimezoneEnabledProjectIDs := flag.String("ingestion_timezone_enabled_projects", "", "List of projectIds whose ingestion timezone is enabled.")
	allowSalesforceActivityTaskByProjectID := flag.String("allowed_salesforce_activity_tasks_by_project_ids", "", "Allowed project id for salesforce activity - task")
	allowSalesforceActivityEventByProjectID := flag.String("allowed_salesforce_activity_events_by_project_ids", "", "Allowed project id for salesforce activity - event")
	disallowSalesforceActivityTaskByProjectID := flag.String("disallowed_salesforce_activity_tasks_by_project_ids", "", "Disallowed project id for salesforce activity - task")
	disallowSalesforceActivityEventByProjectID := flag.String("disallowed_salesforce_activity_events_by_project_ids", "", "Disallowed project id for salesforce activity - event")
	enableDomainsGroupByProjectID := flag.String("enable_domains_group_by_project_id", "", "")
	allowedSalesforceSyncDocTypes := flag.String("allowed_salesforce_doc_types_for_sync", "*", "")
	enableFieldsSyncByProjectID := flag.String("enable_fields_sync_by_project_ids", "", "Use FIELDS() for sync if Request Header is Too Large")
	enableUserDomainsGroupByProjectID := flag.String("enable_user_domains_group_by_project_id", "", "Allow domains group for users")
	enableSyncReferenceFieldsByProjectID := flag.String("enable_sync_reference_fields_by_project_id", "", "")
	enrichPullLimit := flag.Int("enrich_pull_limit", 0, "Limit number of records to be pull from db at a time")
	allowEmailDomainsByProjectID := flag.String("allow_email_domain_by_project_id", "", "Allow email domains for domain group")
	removeDisabledEventUserPropertiesByProjectId := flag.String("remove_disabled_event_user_properties",
		"", "List of projects to disable event user property population in events.")
	enrichRecordProcessLimit := flag.Int("enrich_record_process_limit", 0, "Limit number of records for enrichment at project level")
	disableOpportunityContactRolesByProjectID := flag.String("disable_opportunity_contact_roles_by_project_id", "", "")

	flag.Parse()
	defaultAppName := "salesforce_enrich"
	defaultEnrichHealthcheckPingID := C.HealthcheckSalesforceEnrichPingID
	enrichHealthcheckPingID := C.GetHealthcheckPingID(defaultEnrichHealthcheckPingID, *overrideEnrichHealthcheckPingID)
	defaultSyncHealthcheckPingID := C.HealthcheckSalesforceSyncPingID
	syncHealthcheckPingID := C.GetHealthcheckPingID(defaultSyncHealthcheckPingID, *overrideSyncHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	defer C.PingHealthcheckForPanic(appName, *env, enrichHealthcheckPingID)
	defer C.PingHealthcheckForPanic(appName, *env, syncHealthcheckPingID)

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
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:                                   *primaryDatastore,
		APIDomain:                                          *apiDomain,
		SentryDSN:                                          *sentryDSN,
		SentryRollupSyncInSecs:                             *sentryRollupSyncInSecs,
		SalesforceAppID:                                    *salesforceAppID,
		SalesforceAppSecret:                                *salesforceAppSecret,
		RedisHost:                                          *redisHost,
		RedisPort:                                          *redisPort,
		RedisHostPersistent:                                *redisHostPersistent,
		RedisPortPersistent:                                *redisPortPersistent,
		FactorsSixSignalAPIKey:                             *factorsSixsignalAPIKey,
		DryRunCRMSmartEvent:                                *dryRunSmartEvent,
		CacheSortedSet:                                     *cacheSortedSet,
		AllowedCampaignEnrichmentByProjectID:               *allowedCampaignEnrichmentByProjectID,
		UseOpportunityAssociationByProjectID:               *useOpportunityAssociationByProjectID,
		UseSourcePropertyOverwriteByProjectIDs:             *useSourcePropertyOverwriteByProjectID,
		SkippedOtpProjectIDs:                               *skippedOtpProjectIDs,
		AllowedSalesforceGroupsByProjectIDs:                *enableSalesforceGroupsByProjectIDs,
		CaptureSourceInUsersTable:                          *captureSourceInUsersTable,
		RestrictReusingUsersByCustomerUserId:               *restrictReusingUsersByCustomerUserId,
		DisableCRMUniquenessConstraintsCheckByProjectID:    *disableCRMUniquenessConstraintsCheckByProjectID,
		SalesforceBatchInsertBatchSize:                     *insertBatchSize,
		ClearbitEnabled:                                    *clearbitEnabled,
		SixSignalV1EnabledProjectIDs:                       *sixsignalV1EnabledProjectIDs,
		UseSalesforceV54APIByProjectID:                     *useSalesforceV54APIByProjectID,
		AllowIdentificationOverwriteUsingSourceByProjectID: *allowIdentificationOverwriteUsingSourceByProjectID,
		IngestionTimezoneEnabledProjectIDs:                 C.GetTokensFromStringListAsString(*IngestionTimezoneEnabledProjectIDs),
		AllowedSalesforceActivityTasksByProjectIDs:         *allowSalesforceActivityTaskByProjectID,
		AllowedSalesforceActivityEventsByProjectIDs:        *allowSalesforceActivityEventByProjectID,
		DisallowedSalesforceActivityTasksByProjectIDs:      *disallowSalesforceActivityTaskByProjectID,
		DisallowedSalesforceActivityEventsByProjectIDs:     *disallowSalesforceActivityEventByProjectID,
		EnableDomainsGroupByProjectID:                      *enableDomainsGroupByProjectID,
		AllowedSalesforceSyncDocTypes:                      *allowedSalesforceSyncDocTypes,
		EnableFieldsSyncByProjectID:                        *enableFieldsSyncByProjectID,
		EnableUserDomainsGroupByProjectID:                  *enableUserDomainsGroupByProjectID,
		EnableSyncReferenceFieldsByProjectID:               *enableSyncReferenceFieldsByProjectID,
		AllowEmailDomainsByProjectID:                       *allowEmailDomainsByProjectID,
		RemoveDisabledEventUserPropertiesByProjectID:       *removeDisabledEventUserPropertiesByProjectId,
		DisableOpportunityContactRolesByProjectID:          *disableOpportunityContactRolesByProjectID,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitSortedSetCache(config.CacheSortedSet)
	C.InitSalesforceConfig(config.SalesforceAppID, config.SalesforceAppSecret)
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
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

	syncInfo, status := store.GetStore().GetSalesforceSyncInfo()
	if status != http.StatusFound {
		log.Panicf("Failed to get salesforce syncinfo: %d", status)
	}

	allProjects, allowedProjects, disabledProjects := C.GetProjectsFromListWithAllProjectSupport(
		*projectIDList, *disabledProjectIDList)
	if !allProjects {
		log.WithField("projects", allowedProjects).Info("Running only for the given list of projects.")
	}

	var syncStatus salesforceSyncStatus
	var propertyDetailSyncStatus []IntSalesforce.Status
	anyFailure := false

	if !*enrichOnly {
		for pid, projectSettings := range syncInfo.ProjectSettings {
			if !allowProjectByProjectIDList(pid, allProjects, allowedProjects, disabledProjects) {
				continue
			}

			accessToken, instanceURL, err := IntSalesforce.GetAccessToken(projectSettings, H.GetSalesforceRedirectURL())
			if err != nil {
				log.WithField("project_id", pid).Errorf("Failed to get salesforce access token: %s", err)
				continue
			}

			syncInfo.LastSyncInfo[pid] = overrideLastSyncTimestampIfRequired(*overrideLastSyncTimestamp, syncInfo.LastSyncInfo[pid])

			objectStatus := IntSalesforce.SyncDocuments(projectSettings, syncInfo.LastSyncInfo[pid], accessToken)
			for i := range objectStatus {
				if objectStatus[i].Status != U.CRM_SYNC_STATUS_SUCCESS {
					syncStatus.Failures = append(syncStatus.Failures, objectStatus[i])
					anyFailure = true
				} else {
					syncStatus.Success = append(syncStatus.Success, objectStatus[i])
				}
			}

			if C.AllowSyncReferenceFields(pid) {
				log.Info(fmt.Sprintf("Starting sync reference fields for project %d", pid))

				accessToken, instanceURL, err := IntSalesforce.GetAccessToken(projectSettings, H.GetSalesforceRedirectURL())
				if err != nil {
					log.WithField("project_id", pid).Errorf("Failed to get salesforce access token for sync reference fields: %s", err)
					continue
				}

				failure := IntSalesforce.SyncReferenceField(pid, accessToken, instanceURL)
				if failure {
					anyFailure = true
				}

				log.Info(fmt.Sprintf("Synced reference fields for project %d", pid))
			}

			failure, propertyDetailSync := IntSalesforce.SyncDatetimeAndNumericalProperties(pid, accessToken, instanceURL)
			if failure {
				anyFailure = true
			}

			propertyDetailSyncStatus = append(propertyDetailSyncStatus, propertyDetailSync...)
		}

		if anyFailure {
			C.PingHealthcheckForFailure(syncHealthcheckPingID, syncStatus)
		} else {
			C.PingHealthcheckForSuccess(syncHealthcheckPingID, syncStatus)
			log.WithFields(log.Fields{"syncStatus": syncStatus}).Info("Sync Job completed.")
		}

		C.PingHealthcheckForSuccess(C.HealthcheckSalesforceSyncAlwaysSuccessPingID, nil)
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

		// Runs enrichment for list of project_ids as batch using go routines.
		batches := U.GetInt64ListAsBatch(allowedProjectIDs, *numProjectRoutines)
		enrichStatus := EnrichStatus{}
		workerIndex := 0
		for bi := range batches {
			batch := batches[bi]

			var wg sync.WaitGroup
			for pi := range batch {
				wg.Add(1)
				go syncWorker(batch[pi], &wg, workerIndex, *numDocRoutines, &enrichStatus, allowedSalesforceProjectSettings[batch[pi]], *enrichPullLimit, *enrichRecordProcessLimit)
				workerIndex++
			}
			wg.Wait()
		}

		jobStatus.EnrichStatus = enrichStatus.Status
		if enrichStatus.HasFailure {
			anyFailure = true
		}
		jobStatus.SyncStatus = syncStatus
		jobStatus.PropertyDetailStatus = propertyDetailSyncStatus

		if anyFailure {
			C.PingHealthcheckForFailure(enrichHealthcheckPingID, jobStatus)
			return
		}

		C.PingHealthcheckForSuccess(enrichHealthcheckPingID, jobStatus)
		log.WithFields(log.Fields{"jobStatus": jobStatus}).Info("Job completed.")
	}
}
