package tests

import (
	C "factors/config"
	"flag"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

var config *C.Configuration

func TestMain(m *testing.M) {
	env := flag.String("env", "development", "")
	port := flag.Int("port", 8100, "")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	enableDBConnectionPool2 := flag.Bool("enable_db_conn_pool2", false, "")
	memSQLHost2 := flag.String("memsql_host_2", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort2 := flag.Int("memsql_port_2", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser2 := flag.String("memsql_user_2", C.MemSQLDefaultDBParams.User, "")
	memSQLName2 := flag.String("memsql_name_2", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass2 := flag.String("memsql_pass_2", C.MemSQLDefaultDBParams.Password, "")

	allowedCampaignEnrichmentByProjectID := flag.String("allowed_campaign_enrichment_by_project_id", "*", "Campaign enrichment by project_id.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	geoLocFilePath := flag.String("geo_loc_path", "/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb", "")
	deviceDetectorPath := flag.String("dev_detect_path", "/usr/local/var/factors/devicedetector_data/regexes", "")

	apiDomain := flag.String("api_domain", "factors-dev.com:8080", "")
	appDomain := flag.String("app_domain", "factors-dev.com:3000", "")
	lookbackWindowForEventUserCache := flag.Int("lookback_window_event_user_cache",
		30, "look back window in cache for event/user cache")

	flag.Parse()

	config = &C.Configuration{
		AppName:       "development_test",
		Env:           *env,
		Port:          *port,
		EtcdEndpoints: strings.Split(*etcd, ","),
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
		},
		MemSQL2Info: C.DBConf{
			Host:     *memSQLHost2,
			Port:     *memSQLPort2,
			User:     *memSQLUser2,
			Name:     *memSQLName2,
			Password: *memSQLPass2,
		},
		PrimaryDatastore:            *primaryDatastore,
		RedisHost:                   *redisHost,
		RedisPort:                   *redisPort,
		RedisHostPersistent:         *redisHostPersistent,
		RedisPortPersistent:         *redisPortPersistent,
		GeolocationFile:             *geoLocFilePath,
		DeviceDetectorPath:          *deviceDetectorPath,
		APIDomain:                   *apiDomain,
		APPDomain:                   *appDomain,
		AllowSmartEventRuleCreation: true,

		// Test is not 100% backward compatible. Only some of the unit tests
		// have flag based cases for both backward and forward compatibility.
		// So kept on-table user_properties as primary.
		AllowedCampaignEnrichmentByProjectID:               *allowedCampaignEnrichmentByProjectID,
		UseOpportunityAssociationByProjectID:               "*",
		AllowedHubspotGroupsByProjectIDs:                   "*",
		UseSourcePropertyOverwriteByProjectIDs:             "*",
		AllowedSalesforceGroupsByProjectIDs:                "*",
		AllowSupportForUserPropertiesInIdentifyCall:        "*",
		SkipEventNameStepByProjectID:                       "*",
		SkipUserJoinInEventQueryByProjectID:                "*",
		EnableEventLevelEventProperties:                    "",
		LookbackWindowForEventUserCache:                    *lookbackWindowForEventUserCache,
		EnableOLTPQueriesMemSQLImprovements:                "*",
		CaptureSourceInUsersTable:                          "*",
		AllowSupportForSourceColumnInUsers:                 "*",
		MergeAmpIDAndSegmentIDWithUserIDByProjectID:        "*",
		AllowProfilesGroupSupport:                          "*",
		SessionBatchTransactionBatchSize:                   50,
		DisableCRMUniquenessConstraintsCheckByProjectID:    "*",
		EnableHubspotFormsEventsByProjectID:                "*",
		AllowEventsFunnelsGroupSupport:                     "*",
		UseHubspotBatchInsertByProjectID:                   "*",
		SalesforcePropertyLookBackTimeHr:                   48,
		SalesforceBatchInsertBatchSize:                     10,
		AllowHubspotEngagementsByProjectID:                 "*",
		HubspotPropertyLookBackLimit:                       1000,
		ClearbitEnabled:                                    1,
		SixSignalEnabled:                                   1,
		EnableOptimisedFilterOnProfileQuery:                true,
		EnableOptimisedFilterOnEventUserQuery:              true,
		AllowIdentificationOverwriteUsingSourceByProjectID: "*",
		RestrictReusingUsersByCustomerUserId:               "*",
		DBMaxAllowedPacket:                                 1048576000,
		AllowHubspotPastEventsEnrichmentByProjectID:        "*",
		AllowHubspotContactListInsertByProjectID:           "*",
		AllowedSalesforceActivityTasksByProjectIDs:         "*",
		AllowedSalesforceActivityEventsByProjectIDs:        "*",
		EnableUserLevelEventPullForAddSessionByProjectID:   "*",
		EventsPullMaxLimit:                                 50000,
		FormFillIdentificationAllowedProjects:              "*",
		EnableDBConnectionPool2:                            *enableDBConnectionPool2,
		EnableDomainsGroupByProjectID:                      "*",
		EnableSyncReferenceFieldsByProjectID:               "*",
		EnableUserDomainsGroupByProjectID:                  "*",
		EnableEventFiltersInSegments:                       true,
		AllAccountsProjectId:                               "*",
	}
	C.InitConf(config)

	// Setup.
	// Initialize configs and connections.
	if err := C.InitTestServer(config); err != nil {
		log.Fatal("Failed to initialize config and services.")
		os.Exit(1)
	}
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	C.InitFilemanager(path, *env, config)
	C.InitQueueClient(config.RedisHost, config.RedisPort)
	C.InitDuplicateQueueClient(config.RedisHost, config.RedisPort)

	C.InitPropertiesTypeCache(true, 10000, "*", "")
	if C.GetConfig().Env != C.DEVELOPMENT {
		log.Fatal("Environment is not Development.")
		os.Exit(1)
	}

	retCode := m.Run()
	os.Exit(retCode)
}
