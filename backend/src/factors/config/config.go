package config

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/chargebee/chargebee-go/v3"
	"github.com/getsentry/sentry-go"
	"github.com/go-sql-driver/mysql"

	"github.com/coreos/etcd/mvcc/mvccpb"

	"factors/filestore"
	"factors/vendor_custom/machinery/v1"
	machineryConfig "factors/vendor_custom/machinery/v1/config"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/evalphobia/logrus_sentry"
	D "github.com/gamebtc/devicedetector"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	geoip2 "github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"

	"factors/metrics"
	U "factors/util"

	"factors/services/error_collector"
	serviceEtcd "factors/services/etcd"
	"factors/services/mailer"
	serviceSes "factors/services/ses"

	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"

	cache "github.com/hashicorp/golang-lru"
)

const DEVELOPMENT = "development"
const TEST = "test"
const STAGING = "staging"
const PRODUCTION = "production"

const SENTRY_APP_KEY = "AppName"
const SENTRY_OCCURRENCE_KEY = "occurences"

// Warning: Any changes to the cookie name has to be
// in sync with other services which uses the cookie.
const FactorsSessionCookieName = "factors-sid"

const FactorsAuth0StateCookieName = "factors-auth0-state"

// URL for loading SDK on client side.
const SDKAssetsURL = "https://app.factors.ai/assets/factors.js"

// Datastore specific constants.
const (
	DatastoreTypeMemSQL = "memsql"
)

// MemSQLDefaultDBParams Default connection params for Postgres.
var MemSQLDefaultDBParams = DBConf{
	Host:      "localhost",
	Port:      3306,
	User:      "root",
	Name:      "factors",
	Password:  "dbfactors123",
	IsPSCHost: 0,
}

// PostgresDefaultDBParams Default connection params for MemSQL.
var PostgresDefaultDBParams = DBConf{
	Host:     "localhost",
	Port:     5432,
	User:     "autometa",
	Name:     "autometa",
	Password: "@ut0me7a",
}

type DBConf struct {
	Host        string
	IsPSCHost   int
	Port        int
	User        string
	Name        string
	Password    string
	AppName     string
	UseSSL      bool
	Certificate string

	// Pooling
	MaxOpenConnections     int
	MaxIdleConnections     int
	UseExactConnFromConfig bool
}

type Auth0Conf struct {
	Domain       string
	ClientId     string
	ClientSecret string
	CallbackUrl  string
}

type Configuration struct {
	GCPProjectID                                   string
	GCPProjectLocation                             string
	AppName                                        string
	Env                                            string
	Port                                           int
	DBInfo                                         DBConf
	MemSQLInfo                                     DBConf
	MemSQL2Info                                    DBConf
	Auth0Info                                      Auth0Conf
	SessionStore                                   string
	SessionStoreSecret                             string
	RedisHost                                      string
	RedisPort                                      int
	RedisHostPersistent                            string
	RedisPortPersistent                            int
	QueueRedisHost                                 string
	QueueRedisPort                                 int
	DuplicateQueueRedisHost                        string
	DuplicateQueueRedisPort                        int
	EnableSDKAndIntegrationRequestQueueDuplication bool
	EtcdEndpoints                                  []string
	GeolocationFile                                string
	DeviceDetectorPath                             string
	FactorsSixSignalAPIKey                         string
	APIDomain                                      string
	APPDomain                                      string
	APPOldDomain                                   string
	AWSRegion                                      string
	AWSKey                                         string
	AWSSecret                                      string
	Cookiename                                     string
	Auth0StateName                                 string
	EmailSender                                    string
	ErrorReportingInterval                         int
	AdminLoginEmail                                string
	AdminLoginToken                                string
	FacebookAppID                                  string
	FacebookAppSecret                              string
	LinkedinClientID                               string
	LinkedinClientSecret                           string
	SalesforceAppID                                string
	SalesforceAppSecret                            string
	SentryDSN                                      string
	LoginTokenMap                                  map[string]string
	SkipTrackProjectIds                            []int64
	SDKRequestQueueProjectTokens                   []string
	SegmentRequestQueueProjectTokens               []string
	UseDefaultProjectSettingForSDK                 bool
	BlockedSDKRequestProjectTokens                 []string
	BlockedIPProjectTokens                         string
	// Usage: 	"--cache_look_up_range_projects", "1:20140307"
	CacheLookUpRangeProjects                map[int64]time.Time // Usually cache look up is for past 30 days. If certain projects need override, then this is used
	LookbackWindowForEventUserCache         int
	ActiveFactorsGoalsLimit                 int
	ActiveFactorsTrackedEventsLimit         int
	ActiveFactorsTrackedUserPropertiesLimit int
	DryRunCRMSmartEvent                     bool
	DryRunSmartProperties                   bool
	IsBeamPipeline                          bool
	AllowSmartEventRuleCreation             bool
	// non exported field, only access through function
	propertiesTypeCacheSize                int
	enablePropertyTypeFromDB               bool
	whitelistedProjectIDPropertyTypeFromDB string
	blacklistedProjectIDPropertyTypeFromDB string
	CacheSortedSet                         bool
	ProjectAnalyticsWhitelistedUUIds       []string
	CustomerEnabledProjectsLastComputed    []int64
	SkippedOtpProjectIDs                   string
	PrimaryDatastore                       string
	// Segment Marker lookback
	UseLookbackSegmentMarker           bool
	LookbackSegmentMarker              int
	AllowedGoRoutines                  int
	ProcessOnlyAccountSegments         bool
	RunAllAccountsMarkerProjectIDs     string
	DisableAllAccountsMarkerProjectIDs string
	UseOptimisedEventsQueryProjectIDs  string
	LatestSegmentsMarkerRun            string
	RunForAllAccountsInHours           int
	JobRunCountPerDayMarker            int
	DomainsLimitAllRun                 int
	BatchSizeDomains                   int
	// Flag for enabling only the /mql routes for secondary env testing.
	EnableMQLAPI bool
	// Flags to disable DB and Redis writes when enabled.
	// Added as pointer to prevent accidental writes from
	// other services while testing.
	DisableDBWrites                                      *bool
	DisableQueryCache                                    *bool
	AllowedCampaignEnrichmentByProjectID                 string
	UseOpportunityAssociationByProjectID                 string
	AllowChannelGroupingForProjectIDs                    string
	CloudManager                                         filestore.FileManager
	SegmentExcludedCustomerIDByProject                   map[int64]string // map[project_id]customer_user_id
	AttributionDebug                                     int
	AttributionCommonFlow                                string
	AttributionDBCacheLookup                             string
	AttributionDebugKPI                                  string
	DisableDashboardQueryDBExecution                     bool
	AllowedHubspotGroupsByProjectIDs                     string
	EnableFilterOptimisation                             bool
	FilterPropertiesStartTimestamp                       int64
	OnlyAttributionDashboardCaching                      int
	SkipAttributionDashboardCaching                      int
	IsRunningForMemsql                                   int
	IsHourlyRunEnabled                                   int
	UseSourcePropertyOverwriteByProjectIDs               string
	AllowedSalesforceGroupsByProjectIDs                  string
	DevBox                                               bool
	AllowSupportForUserPropertiesInIdentifyCall          string
	SkipEventNameStepByProjectID                         string
	SkipUserJoinInEventQueryByProjectID                  string
	EnableEventLevelEventProperties                      string
	EnableOLTPQueriesMemSQLImprovements                  string
	CaptureSourceInUsersTable                            string
	AllowSupportForSourceColumnInUsers                   string
	AllowSupportForV1AvgKPIComputation                   string
	ResourcePoolForAnalytics                             string
	RestrictReusingUsersByCustomerUserId                 string
	HubspotAPIOnboardingHAPIKey                          string
	HubspotAPIOnboardingPrivateAccessToken               string
	MailModoOnboardingAPIKey                             string
	MailModoOnboardingURL1                               string
	MailModoOnboardingURL2                               string
	SlackOnboardingWebhookURL                            string
	AllowProfilesGroupSupport                            string
	WebsiteAggregationTestEnabledProjects                string
	DebugEnabled                                         bool
	MergeAmpIDAndSegmentIDWithUserIDByProjectID          string
	SessionBatchTransactionBatchSize                     int
	FivetranGroupId                                      string
	FivetranLicenseKey                                   string
	DisableCRMUniquenessConstraintsCheckByProjectID      string
	SkipDashboardCachingAnalytics                        int
	AllowEventsFunnelsGroupSupport                       string
	MonitoringAPIToken                                   string
	DelayedTaskThreshold                                 int
	SdkQueueThreshold                                    int
	UseQueueRedis                                        bool
	IntegrationQueueThreshold                            int
	UsageBasedDashboardCaching                           int
	OnlyKPICaching                                       int
	SkipKPICaching                                       int
	EnableBingAdsAttribution                             bool
	EnableHubspotFormsEventsByProjectID                  string
	DisableHubspotNonMarketingContactsByProjectID        string
	RemoveDisabledEventUserPropertiesByProjectID         string
	HubspotBatchInsertBatchSize                          int
	UseHubspotBatchInsertByProjectID                     string
	SalesforcePropertyLookBackTimeHr                     int
	SalesforceBatchInsertBatchSize                       int
	AllowHubspotEngagementsByProjectID                   string
	HubspotPropertyLookBackLimit                         int
	EnableSlowDBQueryLogging                             bool
	SlackAppClientID                                     string
	SlackAppClientSecret                                 string
	EnableDryRunAlerts                                   bool
	DataAvailabilityExpiry                               int
	UseSalesforceV54APIByProjectID                       string
	EnableOptimisedFilterOnProfileQuery                  bool
	HubspotAppID                                         string
	HubspotAppSecret                                     string
	EnableOptimisedFilterOnEventUserQuery                bool
	EnableEmailBlockingFlag                              bool
	EnableIPBlockingFlag                                 bool
	BlockedEmailList                                     []string
	BlockedIPList                                        []string
	BlockedEmailDomainList                               []string
	AllAccountsProjectId                                 string
	TimelinesTablePropsQueryOpt                          string
	MarkerPreviewAllAccountsProjectId                    string
	BatchSizePreviewDomain                               int
	AccountsToProcessForPreview                          int
	NumberOfRunsForPreview                               int
	AccountLimitPreviewListing                           int
	UseMarkerByProjectID                                 string
	EnableNewAllAccountsByProjectID                      string
	DBMaxAllowedPacket                                   int64
	AllowIdentificationOverwriteUsingSourceByProjectID   string
	AllowHubspotPastEventsEnrichmentByProjectID          string
	AllowHubspotContactListInsertByProjectID             string
	IngestionTimezoneEnabledProjectIDs                   []string
	LinkedinMemberCompanyConfigProjectIDs                []string
	AllowedSalesforceActivityTasksByProjectIDs           string
	AllowedSalesforceActivityEventsByProjectIDs          string
	DisallowedSalesforceActivityTasksByProjectIDs        string
	DisallowedSalesforceActivityEventsByProjectIDs       string
	EnableUserLevelEventPullForAddSessionByProjectID     string
	EventsPullMaxLimit                                   int
	EnableFeatureGates                                   bool
	EnableDBConnectionPool2                              bool
	FormFillIdentificationAllowedProjects                string
	EmailUTMParameterAllowedProjects                     string
	EnableEventFiltersInSegments                         bool
	UseSentryRollup                                      bool
	SentryRollupSyncInSecs                               int
	EnableSixSignalGroupByProjectID                      string
	EnableDebuggingForIP                                 bool
	TeamsAppTenantID                                     string
	TeamsAppClientID                                     string
	TeamsAppClientSecret                                 string
	TeamsApplicationID                                   string
	EnableDomainsGroupByProjectID                        string
	DisableUpdateNextSessionTimestamp                    int
	EnableSyncReferenceFieldsByProjectID                 string
	StartTimestampForWeekMonth                           int64
	CacheForLongerExpiryProjects                         string
	CacheOnlyDashboards                                  string
	CacheOnlyDashboardUnits                              string
	AllowedSalesforceSyncDocTypes                        string
	CustomDateStart                                      int64
	CustomDateEnd                                        int64
	EnableFieldsSyncByProjectID                          string
	EnableUserDomainsGroupByProjectID                    string
	UseHubspotCompaniesV3APIByProjectID                  string
	UseHubspotEngagementsV3APIByProjectID                string
	AllowEventAnalyticsGroupsByProjectID                 string
	OtpKeyWithQueryCheckEnabled                          bool
	EnableFeatureGatesV2                                 bool
	AllowEmailDomainsByProjectID                         string
	UseHubspotDealsV3APIByProjectID                      string
	EnableScoringByProjectID                             string
	DeviceServiceURL                                     string
	EnableDeviceServiceByProjectID                       string
	DisableOpportunityContactRolesByProjectID            string
	ExcludeBotIPV4AddressByRange                         string
	SlackInternalAlertWebhookUrl                         string
	UseHashIDForCRMGroupUserByProject                    string
	MoveHubspotCompanyAssocationFlowToContactByPojectID  string
	ExplainV3QueryBuilder                                bool
	ChargebeeApiKey                                      string
	ChargebeeSiteName                                    string
	UserPropertyUpdateOptProjects                        string
	AssociateDealToDomainByProjectID                     string
	EnableSyncTriesFlag                                  bool
	ClearbitProvisionAccountAPIKey                       string
	SalesforceSkipLeadUpdatesProcessingByProjectID       string
	SalesforceAllowOpportunityOverrideCreateCreatedEvent string
	AggrEventPropertyValuesCacheByProjectID              string
	ParagonTokenSigningKey                               string
	ParagonProjectID                                     string
	MailModoTriggerCampaignAPIKey                        string
	AddCRMObjectURLPropertyByProjectID                   string
	EnableTotalSessionPropertiesV2ByProjectID            string
	SkipKpiResultValidation                              bool
	EnableDomainWebsitePropertiesByProjectID             string
	HubspotEnrichSkipContactUpdatesByProjectID           string
	EnableSalesforceDeletedRecordByProjectID             string
	EnableEnrichmentDebugLogsByProjectID                 string
	EnableCacheDBWriteProjects                           string
	EnableCacheDBReadProjects                            string
	SkipSalesforceLeadEnrichmentByProjectID              string
	SalesforceEnrichOnlyObjects                          string
	EnableSalesforceRelationshipPullByProjectID          string
	SixSignalV3ProjectIds                                string
	ChatDebug                                            int
}

type Services struct {
	Db                   *gorm.DB
	DBContext            *context.Context
	DBContextCancel      *context.CancelFunc
	Db2                  *gorm.DB
	DBContext2           *context.Context
	DBContextCancel2     *context.CancelFunc
	GeoLocation          *geoip2.Reader
	Etcd                 *serviceEtcd.EtcdClient
	Redis                *redis.Pool
	RedisPeristent       *redis.Pool
	QueueClient          *machinery.Server
	DuplicateQueueClient *machinery.Server
	patternServersLock   sync.RWMutex
	patternServers       map[string]string
	Mailer               mailer.Mailer
	ErrorCollector       *error_collector.Collector
	DeviceDetector       *D.DeviceDetector
	SentryHook           *logrus_sentry.SentryHook
	MetricsExporter      *stackdriver.Exporter
	logErrorsLock        sync.RWMutex
	logErrors            map[string]*SentryInfo
	QueueRedis           *redis.Pool
}

// Healthchecks.io ping IDs for monitoring. Can be used anywhere in code to report error on job.
// Use flag --healthcheck_ping_id to override default ping_id for duplicate/special jobs.
const (
	// Cron ping IDs.
	HealthcheckAddSessionPingID                       = "8da15fff-15f0-4410-9efc-524f624fd388"
	HealthcheckArchiveEventsPingID                    = "b2d0f1df-901e-4113-bb45-eed95539790b"
	HealthcheckBigqueryUploadPingID                   = "03e0fba3-d660-4679-8595-29b6cd04e87c"
	HealthcheckCleanupEventUserCachePingID            = "85e21b5c-5503-4172-af40-de918741a4d1"
	HealthcheckDashboardCachingPingID                 = "72e5eadc-b46e-45ca-ba78-29819532307d"
	HealthcheckDashboardDBAttributionPingID           = "5dfce175-273d-4096-ad70-578270621e8c"
	HealthcheckDashboardDBAttributionHourlyPingID     = "8f4dd14f-642f-4eed-b4aa-2f971c3bbbc0"
	HealthcheckHubspotEnrichPingID                    = "6f522e60-6bf8-4aea-99fe-f5a1c68a00e7"
	HealthcheckHubspotFirstTimeEnrichPingID           = "6636c652-cd5c-47ab-b996-0d2fb5aed0c6"
	HealthcheckOTPHubspotPingID                       = "c937adf4-a54d-4ee8-8ec9-3e5ed8a83e42"
	HealthcheckOTPSalesforcePingID                    = "4ffe2119-64bc-41d4-bc5a-64bea0f121b7"
	HealthcheckMonitoringJobPingID                    = "18db44be-c193-4f11-84e5-5ff144e272e9"
	HealthcheckSalesforceEnrichPingID                 = "e56175aa-3407-4595-bb94-d8325952b224"
	HealthcheckSalesforceSyncPingID                   = "c5434535-ea40-42b8-8e1f-243ccf88fef3"
	HealthcheckSalesforceSyncAlwaysSuccessPingID      = "7ad8b8dc-c820-4059-9c07-53fccd320224"
	HealthcheckYourstoryAddPropertiesPingID           = "acf7faab-c56f-415e-aa10-ca2aa9246172"
	HealthCheckSmartPropertiesPingID                  = "ead84671-b84c-481b-bfa5-59403d626652"
	HealthCheckSmartPropertiesDupPingID               = "d2b55241-52d8-4cc5-a49c-5b57f6a96642"
	HealthcheckBeamDashboardCachingPingID             = "ecb259b9-4ff8-4825-b989-81d47bd34d93"
	HealthcheckBeamDashboardCachingNowPingID          = "be2f00de-57e1-401b-b2c9-9df305c3f528"
	HealthcheckMonitoringJobMemSQLPingID              = "de2b64d4-952e-47ca-ac70-1bf9d8e1587e"
	HealthcheckSavedQueriesTimezoneChangePingID       = "42f96466-c467-44cc-899d-7e55b8a1aa4e"
	HealthcheckLeadgenInsertionJobPingID              = "830c0112-fc71-4257-b265-b3732f03115a"
	HealthcheckBingAdsIntegrationPingID               = "33f862b1-453a-4352-b209-945b38ed1902"
	HealthcheckMarketoEnrichmentPingID                = "72709dcf-3e85-476d-9c81-8eb4be639368"
	HealthcheckMarketoIntegrationPingID               = "58653751-9db8-4251-9048-6c999a139138"
	HealthcheckLeadSquaredIntegrationSyncPingID       = "9b0c317c-ea4f-4113-afe4-a2de84f6c33f"
	HealthcheckLeadSquaredPullEventsPingID            = "5766cfd3-1108-4633-b9e7-5a818748d1bf"
	HealthcheckLeadSquaredEnrichPingID                = "83fdec06-a2a0-4fdb-ba9e-97c9bb730e23"
	HealthcheckAdsImportPingID                        = "c392e3b4-4883-47ae-b5ff-63743d5d0c78"
	HealthcheckComputeAndSendAlertsPingID             = "8345e798-1622-4881-942e-99fdd638ddf0"
	HealthcheckMailWIPingID                           = "950b628b-d623-4666-be39-952516e543c0"
	HealthCheckWeeklyInsightsPingID                   = "56c0ba58-3d1f-4408-924a-3e6588d4ad5c"
	HealthcheckPatternMinePingID                      = "04e9ba3d-5b07-4325-ad28-6ac7cf15971b"
	HealthcheckPullEventsPingID                       = "088cc760-f350-4eb1-bbb6-c2bbde66b530"
	HealthcheckPathAnalysisPingID                     = "9f71b930-9233-4e58-9935-5de0434d8fa8"
	HealthCheckPreBuiltCustomKPIPingID                = "9e5ac799-e15f-4f44-86b0-4be88379f486"
	HealthCheckAnalyzeJobPingID                       = "3d1bd82d-e036-4433-a794-1042a7f29976"
	HealthCheckSixSignalReportPingID                  = "2508c4c3-b941-40bb-8f2b-a59e4bedf3e5"
	HealthcheckCurrencyUploadPingID                   = "29defb4f-c95e-4895-a515-591fb7c216f7"
	HealthcheckEventTriggerAlertPingID                = "352760ec-66a2-4b5f-b52e-e2c3f434a567"
	HealthcheckEventTriggerAlertForHighPriorityPingID = "081f8a7e-a3cf-4aaa-b67e-ab1c729fb3a7"
	HealthcheckLinkedinGroupUserPingID                = "a8b221cd-6f14-4c9c-8ae7-cd26f585868b"
	HeathCheckG2ETLPingID                             = "4ccbf168-5175-4e08-84e6-7a6ce58bcb08"
	HeathCheckG2EnrichmentPingID                      = "3b240e93-e130-4ea6-b698-5d5d0ea0a83f"
	HealthcheckAccScoringJobPingID                    = "3f93c58e-708c-413e-abc4-0e112ae07260"
	HealthcheckEventCubeAggregationPingID             = "a9ebad9d-4d78-4ea3-9e92-002188102cdd"
	HealthCheckClearbitAccountProvisioningJobPingID   = "a18e152a-7978-4d9f-aa4a-da5121823203"
	HealthcheckSlackUsersSyncJobPingID                = ""

	// Other services ping IDs. Only reported when alert conditions are met, not periodically.
	// Once an alert is triggered, ping manually from Healthchecks UI after fixing.
	HealthcheckDatabaseHealthPingID       = "8464d06b-418b-42d2-9201-b01dc744d283"
	HealthcheckDatabaseHealthMemSQLPingID = "763baa99-61bf-4721-b293-e62eb1027987"
	HealthcheckSDKHealthPingID            = "bb2c4757-9fa4-48eb-bd08-42a16996a61b"
	HealthcheckFactorsDeanonAlertPingID   = "7ea00068-f971-4d03-b3b7-115f3708178b"
)

func PingHealthCheckBasedOnStatus(status map[string]interface{}, healthcheckPingID string) bool {
	errorMap := make(map[string]map[string]interface{})
	isSuccess := true
	for reason, message := range status {
		if message == false {
			errorMap[reason] = make(map[string]interface{})
			deltaStatus := make(map[string]interface{})
			switch x := status[reason[6:]].(type) {
			case map[string]interface{}:
				deltaStatus = x
			case string:
				errorMap[reason]["error"] = x
			}
			for key, val := range deltaStatus {
				if strings.Contains(key, "error") {
					errorMap[reason][key] = val
					if strings.HasPrefix(val.(string), "invalid end timestamp") {
						continue
					}
					isSuccess = false
				}
			}
		}
	}
	if isSuccess {
		PingHealthcheckForSuccess(healthcheckPingID, status)
	} else {
		PingHealthcheckForFailure(healthcheckPingID, errorMap)
	}
	return isSuccess
}

func (service *Services) GetPatternServerAddresses() []string {
	service.patternServersLock.RLock()
	defer service.patternServersLock.RUnlock()

	ps := make([]string, 0, 0)
	for _, addr := range service.patternServers {
		ps = append(ps, addr)
	}
	return ps
}

func (service *Services) addPatternServer(key, addr string) {
	log.Infof("Add Pattern Server Key:%s, addr: %s", key, addr)
	service.patternServersLock.Lock()
	defer service.patternServersLock.Unlock()

	service.patternServers[key] = addr
}

func (service *Services) removePatternServer(key string) {
	log.Infof("Remove Pattern Server Key: %s", key)
	service.patternServersLock.Lock()
	defer service.patternServersLock.Unlock()

	delete(services.patternServers, key)
}

var configuration *Configuration
var services *Services = nil

// PropertiesTypeCache common cache with reset date
type PropertiesTypeCache struct {
	Cache         *cache.Cache `json:"cache"`
	LastResetDate string       `json:"last_reset_date"`
}

var propertiesTypeCache *PropertiesTypeCache

// InitPropertiesTypeCache initialze properties type LRU cache by fixed size
func InitPropertiesTypeCache(enablePropertyTypeFromDB bool, propertiesTypeCacheSize int, whitelistedProjectIDPropertyTypeFromDB, blacklistedProjectIDPropertyTypeFromDB string) {
	if !enablePropertyTypeFromDB || propertiesTypeCacheSize <= 0 || propertiesTypeCache != nil {
		return
	}

	if (blacklistedProjectIDPropertyTypeFromDB == "" && whitelistedProjectIDPropertyTypeFromDB == "") ||
		(blacklistedProjectIDPropertyTypeFromDB != "" && whitelistedProjectIDPropertyTypeFromDB != "") {
		return
	}

	pCache, err := cache.New(propertiesTypeCacheSize)
	if err != nil {
		log.WithError(err).WithField("PropertiesTypeCacheSize",
			propertiesTypeCacheSize).Fatal("Failed to initialize properties_type cache size.")
		return
	}
	propertiesTypeCache = &PropertiesTypeCache{
		Cache: pCache,
	}

	if blacklistedProjectIDPropertyTypeFromDB != "" {
		configuration.blacklistedProjectIDPropertyTypeFromDB = blacklistedProjectIDPropertyTypeFromDB
	} else {
		configuration.whitelistedProjectIDPropertyTypeFromDB = whitelistedProjectIDPropertyTypeFromDB
	}

	configuration.enablePropertyTypeFromDB = enablePropertyTypeFromDB

	propertiesTypeCache.LastResetDate = U.GetDateOnlyFromTimestampZ(U.TimeNowUnix())
	log.Info("Properties_type cache initialized.")
}

func IsDBConnectionPool2Enabled() bool {
	return configuration.EnableDBConnectionPool2
}

func IsAllowedCampaignEnrichementByProjectID(projectID int64) bool {
	if configuration.AllowedCampaignEnrichmentByProjectID == "" {
		return false
	}

	if configuration.AllowedCampaignEnrichmentByProjectID == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.AllowedCampaignEnrichmentByProjectID, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false

}

func IsAllowedHubspotGroupsByProjectID(projectID int64) bool {
	if configuration.AllowedHubspotGroupsByProjectIDs == "" {
		return false
	}

	if configuration.AllowedHubspotGroupsByProjectIDs == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.AllowedHubspotGroupsByProjectIDs, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}

func IsAllowedSalesforceGroupsByProjectID(projectID int64) bool {
	if configuration.AllowedSalesforceGroupsByProjectIDs == "" {
		return false
	}

	if configuration.AllowedSalesforceGroupsByProjectIDs == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.AllowedSalesforceGroupsByProjectIDs, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}

func SkipEventNameStepByProjectID(projectID int64) bool {
	if configuration.SkipEventNameStepByProjectID == "" {
		return false
	}

	if configuration.SkipEventNameStepByProjectID == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.SkipEventNameStepByProjectID, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}

func SkipUserJoinInEventQueryByProjectID(projectID int64) bool {
	if configuration.SkipUserJoinInEventQueryByProjectID == "" {
		return false
	}

	if configuration.SkipUserJoinInEventQueryByProjectID == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.SkipUserJoinInEventQueryByProjectID, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}

// GetPropertiesTypeCache returns PropertiesTypeCache instance
func GetPropertiesTypeCache() *PropertiesTypeCache {
	return propertiesTypeCache
}

// ResetPropertyDetailsCacheByDate reset PropertiesTypeCache with date
func ResetPropertyDetailsCacheByDate(timestamp int64) {
	date := U.GetDateOnlyFromTimestampZ(timestamp)
	propertiesTypeCache.Cache.Purge()
	propertiesTypeCache.LastResetDate = date
}

// IsEnabledPropertyDetailFromDB should allow property type check from DB.
func IsEnabledPropertyDetailFromDB() bool {
	return configuration.enablePropertyTypeFromDB
}

// UseLookbackForSegmentMarker allow next run from LookbackSegmentMarker flag
func UseLookbackForSegmentMarker() bool {
	return configuration.UseLookbackSegmentMarker
}

func LookbackForSegmentMarker() int {
	return configuration.LookbackSegmentMarker
}

// Run marker for all accounts if marker_last_run_all_accounts is greater than given hours
func TimeRangeForAllDomains() int {
	return configuration.RunForAllAccountsInHours
}
func MarkerJobRunCountPerDay() int {
	return configuration.JobRunCountPerDayMarker
}
func MarkerDomainLimitForAllRun() int {
	return configuration.DomainsLimitAllRun
}
func BatchSizeSegmentMarker() int {
	return configuration.BatchSizeDomains
}
func ProcessOnlyAllAccountsSegments() bool {
	return configuration.ProcessOnlyAccountSegments
}

func AllowedGoRoutinesSegmentMarker() int {
	return configuration.AllowedGoRoutines
}

// IsEnabledEventsFilterInSegments should allow event properties to be added in the query.
func IsEnabledEventsFilterInSegments() bool {
	return configuration.EnableEventFiltersInSegments
}

// IsEnabledPropertyDetailByProjectID enabled project_id for property type check from DB
func IsEnabledPropertyDetailByProjectID(projectID int64) bool {
	if projectID == 0 || !IsEnabledPropertyDetailFromDB() {
		return false
	}

	projectIDstr := fmt.Sprintf("%d", projectID)

	if configuration.whitelistedProjectIDPropertyTypeFromDB == "*" {
		return true
	}

	if configuration.whitelistedProjectIDPropertyTypeFromDB != "" {
		projectIDs := strings.Split(configuration.whitelistedProjectIDPropertyTypeFromDB, ",")
		for i := range projectIDs {
			if projectIDs[i] == projectIDstr {
				return true
			}
		}
	}

	if configuration.blacklistedProjectIDPropertyTypeFromDB != "" {
		projectIDs := strings.Split(configuration.blacklistedProjectIDPropertyTypeFromDB, ",")
		for i := range projectIDs {
			if projectIDs[i] == projectIDstr {
				return false
			}
		}
	}

	return false
}
func initLogging(collector *error_collector.Collector) {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	if IsDevelopment() {
		log.SetLevel(log.DebugLevel)
	}

	log.SetReportCaller(true)

	if collector != nil {
		hook := &U.Hook{C: services.ErrorCollector}
		log.AddHook(hook)
	}
}

func initGeoLocationService(geoLocationFile string) {
	if geoLocationFile == "" {
		log.WithField("geo_location_file",
			geoLocationFile).Fatal("Invalid geolocation file.")
	}

	if services == nil {
		services = &Services{}
	}

	// Ref: https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz
	geolocation, err := geoip2.Open(geoLocationFile)
	if err != nil {
		log.WithError(err).WithField("GeolocationFilePath",
			geoLocationFile).Fatal("Failed to initialize geolocation service.")
	}

	log.Info("Geolocation service intialized.")
	services.GeoLocation = geolocation
}
func initDeviceDetectorPath(deviceDetectorPath string) {
	if deviceDetectorPath == "" {
		log.WithField("dev_detect_path",
			deviceDetectorPath).Fatal("Invalid device detector path.")
	}
	if services == nil {
		services = &Services{}
	}
	deviceDetector, err := D.NewDeviceDetector(deviceDetectorPath)
	if err != nil {
		log.WithError(err).WithField("DeviceDetectorPath",
			deviceDetectorPath).Fatal("Failed to initialize device detector service.")
	}

	log.Info("Device Detector Path service intialized.")
	services.DeviceDetector = deviceDetector
}

func initAppServerServices(config *Configuration) error {
	services = &Services{patternServers: make(map[string]string)}

	err := InitDB(*config)
	if err != nil {
		return err
	}

	initCookieInfo(configuration.Env)

	InitRedis(config.RedisHost, config.RedisPort)
	InitQueueRedis(config.QueueRedisHost, config.QueueRedisPort)

	InitMailClient(config.AWSKey, config.AWSSecret, config.AWSRegion)
	InitSentryLogging(config.SentryDSN, config.AppName)
	InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)

	initGeoLocationService(config.GeolocationFile)
	initDeviceDetectorPath(config.DeviceDetectorPath)

	InitChargebeeObject(config.ChargebeeApiKey, config.ChargebeeSiteName)

	// Etcd is now an optional service.
	// Any failure should still continue instead of crashing.
	err = InitEtcd(config.EtcdEndpoints)
	if err == nil {
		regPatternServers, err := GetServices().Etcd.DiscoverPatternServers()
		if err != nil && err != serviceEtcd.NotFound {
			log.WithError(err).Errorln("Falied to initialize discover pattern servers")
			return err
		}

		for _, ps := range regPatternServers {
			services.addPatternServer(ps.Key, ps.Value)
		}

		go func() {
			psUpdateChannel := GetServices().Etcd.Watch(serviceEtcd.PatternServerPrefix, clientv3.WithPrefix())
			watchPatternServers(psUpdateChannel)
		}()
	} else {
		log.WithError(err).Error("Failed to intialise etcd. Skipping.")
	}

	return nil
}

func initCookieInfo(env string) {
	// Warning: Any changes to the cookie name has to be
	// in sync with other services which uses the cookie.

	cookieName := fmt.Sprintf("%s%s", FactorsSessionCookieName, "d")
	stateCookieName := fmt.Sprintf("%s%s", FactorsAuth0StateCookieName, "d")
	if env == STAGING {
		cookieName = fmt.Sprintf("%s%s", FactorsSessionCookieName, "s")
		stateCookieName = fmt.Sprintf("%s%s", FactorsAuth0StateCookieName, "s")
	} else if env == PRODUCTION {
		cookieName = FactorsSessionCookieName
		stateCookieName = FactorsAuth0StateCookieName
	}

	configuration.Cookiename = cookieName
	configuration.Auth0StateName = stateCookieName

	log.Info("Initialised cookie info.")
}

func InitConf(c *Configuration) {
	if IsConfigInitialized() {
		log.Info("Configuration alreay initialised.")
		return
	}

	log.SetFormatter(&log.JSONFormatter{})
	if c == nil {
		log.Fatal("Invalid configuration.")
	}

	if c.Env == "" {
		log.WithField("config", c).
			Fatal("Environment not provided on config intialization.")
	}

	// default config.
	c.UseSentryRollup = true
	if c.SentryRollupSyncInSecs == 0 {
		c.SentryRollupSyncInSecs = 300
	}

	log.WithField("config", c).Info("Configuration Initialized.")
	configuration = c
}

func IsConfigInitialized() bool {
	return configuration != nil && configuration.Env != ""
}

func InitSortedSetCache(cacheSortedSet bool) {
	configuration.CacheSortedSet = cacheSortedSet
}

func InitSalesforceConfig(salesforceAppId, salesforceAppSecret string) {
	configuration.SalesforceAppID = salesforceAppId
	configuration.SalesforceAppSecret = salesforceAppSecret
}

func InitEtcd(EtcdEndpoints []string) error {
	etcdClient, err := serviceEtcd.New(EtcdEndpoints)
	if err != nil {
		log.WithError(err).Errorln("Falied to initialize etcd client")
		return err
	}
	log.Infof("ETCD Service Initialized with endpoints: %v", EtcdEndpoints)
	services.Etcd = etcdClient
	configuration.EtcdEndpoints = EtcdEndpoints
	return nil
}

func InitDBWithMaxIdleAndMaxOpenConn(config Configuration,
	maxOpenConns, maxIdleConns int) error {
	if IsDBConnectionPool2Enabled() {
		InitMemSQLDBWithMaxIdleAndMaxOpenConn(config.MemSQL2Info, maxOpenConns, maxIdleConns, true)
	}

	return InitMemSQLDBWithMaxIdleAndMaxOpenConn(config.MemSQLInfo, maxOpenConns, maxIdleConns, false)
}

func GetMemSQLDSNString(dbConf *DBConf) string {
	if dbConf.User == "" || dbConf.Password == "" ||
		dbConf.Name == "" || dbConf.Host == "" ||
		dbConf.Port == 0 {

		log.WithField("db_config", dbConf).Fatal("Invalid memsql db config.")
	}

	memsqlDBConfig := mysql.Config{
		User:                 dbConf.User,
		Passwd:               dbConf.Password,
		Addr:                 fmt.Sprintf("%s:%d", dbConf.Host, dbConf.Port),
		Net:                  "tcp",
		DBName:               dbConf.Name,
		Loc:                  time.Local, // Todo: Use UTC timezone.
		AllowNativePasswords: true,
		ParseTime:            true,
		Params:               map[string]string{"charset": "utf8mb4"},
	}

	if dbConf.UseSSL {
		if dbConf.Certificate == "" {
			log.Fatal("Enable use_ssl but certificate not given.")
		}

		const tlsConfigname = "custom"

		// Register certificate.
		rootCertPool := x509.NewCertPool()
		if ok := rootCertPool.AppendCertsFromPEM([]byte(dbConf.Certificate)); !ok {
			log.Fatal("Failed to add certificate for memsql connection.")
		}
		mysql.RegisterTLSConfig(tlsConfigname, &tls.Config{RootCAs: rootCertPool})

		// use the registered certificate.
		memsqlDBConfig.TLSConfig = tlsConfigname

		log.Info("Using SSL for MemSQL connections.")
	}

	return memsqlDBConfig.FormatDSN()
}

func UseResourcePoolForAnalytics() (bool, string) {
	return configuration.ResourcePoolForAnalytics != "", configuration.ResourcePoolForAnalytics
}

func SetMemSQLResourcePoolQueryCallbackUsingSQLTx(db *sql.Tx, pool string) {
	logCtx := log.WithField("memsql_user", configuration.MemSQLInfo.User)

	// Use olap_pool only for production environment.
	if !IsProduction() {
		return
	}

	if configuration.PrimaryDatastore != DatastoreTypeMemSQL {
		return
	}

	_, err := db.Exec("SET resource_pool = ?", pool)
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to set resource pool before query.")
		return
	}
}

func isValidMemSQLResourcePool(resourcePool string) bool {
	if resourcePool == "" {
		return true
	}

	// Keeping it flexible for develpment.
	if IsDevelopment() {
		return true
	}

	var availablePools []string
	if IsProduction() {
		availablePools = []string{
			"soft_cpu_50",
			"timeout_5m",
			"soft_cpu_45_mem_50_tout_15m",
		}

	} else if IsStaging() {
		availablePools = []string{
			"soft_cpu_50",
			"soft_cpu_30",
			"soft_cpu_15",

			"timeout_1m",
			"timeout_5m",
			"timeout_10m",
		}
	}

	exists, _, _ := U.StringIn(availablePools, resourcePool)
	return exists
}

func InitMemSQLDBWithMaxIdleAndMaxOpenConn(dbConf DBConf, maxOpenConns, maxIdleConns int, isDb2 bool) error {
	if services == nil {
		services = &Services{}
	}

	// SSL enabled for staging and production,
	// if Private Service Connect is not enabled.
	isPSCHost := dbConf.IsPSCHost > 0
	dbConf.UseSSL = (IsStaging() || IsProduction()) && !isPSCHost

	memSQLDB, err := gorm.Open("mysql", GetMemSQLDSNString(&dbConf))
	if err != nil {
		log.WithError(err).Error("Failed connecting to memsql.")
	}
	memSQLDB.LogMode(IsDevelopment())

	// Removes emoji and cleans up string and postgres.Jsonb columns.
	memSQLDB.Callback().Create().Before("gorm:create").Register("cleanup", U.GormCleanupCallback)
	memSQLDB.Callback().Create().Before("gorm:create").Register("default", U.GormDefault)

	memSQLDB.Callback().Create().Before("gorm:update").Register("cleanup", U.GormCleanupCallback)
	var info DBConf
	if IsDBConnectionPool2Enabled() {
		info = configuration.MemSQL2Info
	} else {
		info = configuration.MemSQLInfo
	}

	// info=
	if info.UseExactConnFromConfig {
		// Use connection configuration from flag.
		maxOpenConns = info.MaxOpenConnections
		maxIdleConns = info.MaxIdleConnections
	} else {
		// Using same no.of connections for both max_open and
		// max_idle (greatest among two) as a workaround to
		// avoid connection timout error, while adding new
		// connection to the pool.
		// dial tcp 34.82.234.136:3306: connect: connection timed out
		connections := maxOpenConns
		if maxIdleConns > connections {
			connections = maxIdleConns
		}
		log.Warnf("Using %d connections for both max_idle and max_open for memsql.", connections)

		maxOpenConns = connections
		maxIdleConns = connections
	}
	logCtx := log.WithField("max_open_connections", maxOpenConns).
		WithField("max_idle_connections", maxIdleConns)

	if maxOpenConns == 0 {
		logCtx.Fatal("Invalid max_open_connections. Should be greater than zero.")
	}
	if maxIdleConns == 0 {
		logCtx.Warn("Max idle connections configured as zero.")
	}

	memSQLDB.DB().SetMaxOpenConns(maxOpenConns)
	memSQLDB.DB().SetMaxIdleConns(maxIdleConns)

	logCtx.Info("MemSQL DB Service initialized")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// initiates corresponding service.
	if isDb2 {
		services.Db2 = memSQLDB
		services.DBContext2 = &ctx
		services.DBContextCancel2 = &cancel
	} else {
		services.Db = memSQLDB
		services.DBContext = &ctx
		services.DBContextCancel = &cancel
	}

	configuration.DBInfo = dbConf

	return nil
}

// UseMemSQLDatabaseStore Returns true if memsql is configured as primary datastore.
func UseMemSQLDatabaseStore() bool {
	return GetPrimaryDatastore() == DatastoreTypeMemSQL
}

// GetPrimaryDatastore Returns memsql only if set in config. Defaults to postgres.
func GetPrimaryDatastore() string {
	return DatastoreTypeMemSQL
}

func IsDatastoreMemSQL() bool {
	return GetConfig().PrimaryDatastore == DatastoreTypeMemSQL
}

// GetRoutesURLPrefix Prefix for urls supported on memsql. Returns /mql if enabled.
func GetRoutesURLPrefix() string {
	if EnableMQLAPI() {
		return "/mql"
	}
	return ""
}

// DisableMemSQLDBWrites If DB writes are disabled. Defaults to true unless specified explicitly.
func DisableDBWrites() bool {
	if GetConfig().Env == DEVELOPMENT || GetConfig().Env == TEST {
		return false
	}

	if GetConfig().DisableDBWrites != nil {
		return *GetConfig().DisableDBWrites
	}
	return true
}

// DisableMemSQLQueryCache If dashboard and query cache to be disabled. Defaults to false unless specified explicitly.
func DisableQueryCache() bool {
	if GetConfig().Env == DEVELOPMENT || GetConfig().Env == TEST {
		return false
	}

	if GetConfig().DisableQueryCache != nil {
		return *GetConfig().DisableQueryCache
	}
	return false
}

func NewRequestBuilderWithPrefix(methodType, URL string) *U.RequestBuilder {
	return U.NewRequestBuilder(methodType, GetRoutesURLPrefix()+URL)
}

// KillDBQueriesOnExit Uses context to kill any running queries when kill signal is received.
func KillDBQueriesOnExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		select {
		case <-c:
			if GetServices().DBContext != nil && GetServices().DBContextCancel != nil {
				(*GetServices().DBContextCancel)()
			}

			if GetServices().DBContext2 != nil && GetServices().DBContextCancel2 != nil {
				(*GetServices().DBContextCancel2)()
			}

			signal.Stop(c)
		}
	}()
}

func InitDB(config Configuration) error {
	if !IsConfigInitialized() {
		log.Fatal("Config not initialised on InitDB.")
	}

	// default configuration.
	return InitDBWithMaxIdleAndMaxOpenConn(config, 50, 10)
}

func InitRedisPersistent(host string, port int) {
	initRedisConnection(host, port, true, false, 300, 100)
}

func InitFilemanager(bucketName string, env string, config *Configuration) {
	if env == "development" {
		config.CloudManager = serviceDisk.New(bucketName)
	} else {

		log.WithField("bucket_name", bucketName).Info("Initialised file manager.")

		var err error
		config.CloudManager, err = serviceGCS.New(bucketName)
		if err != nil {
			log.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
	}
}

func InitRedis(host string, port int) {
	initRedisConnection(host, port, false, false, 300, 100)
}

func InitQueueRedis(host string, port int) {
	initRedisConnection(host, port, false, true, 300, 100)
}

// InitRedisConnection Init redis with custom requirements.
func InitRedisConnection(host string, port int, persistent bool, maxActive, maxIdle int) {
	initRedisConnection(host, port, persistent, false, maxActive, maxIdle)
}

func initRedisConnection(host string, port int, persistent bool, queue bool, maxActive, maxIdle int) {
	if host == "" || port == 0 {
		log.WithField("host", host).WithField("port", port).Fatal(
			"Invalid redis host or port.")
	}

	if services == nil {
		services = &Services{}
	}

	conn := fmt.Sprintf("%s:%d", host, port)
	redisPool := &redis.Pool{
		MaxActive: maxActive,
		MaxIdle:   maxIdle,
		// IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", conn)
			if err != nil {
				// do not panic. connection dial would be called
				// on pool refill too.
				log.WithError(err).Error("Redis connection dial error.")
				return nil, err
			}

			return c, err
		},

		// Tests connection before idle connection being reused.
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}

			_, err := c.Do("PING")
			if err != nil {
				log.WithError(err).Error("Redis connection test on borrow error.")
			}

			return err
		},
	}

	log.Info("Redis Service initialized.")
	if queue {
		configuration.QueueRedisHost = host
		configuration.QueueRedisPort = port
		services.QueueRedis = redisPool
	} else if persistent {
		configuration.RedisHostPersistent = host
		configuration.RedisPortPersistent = port
		services.RedisPeristent = redisPool
	} else {
		configuration.RedisHost = host
		configuration.RedisPort = port
		services.Redis = redisPool
	}
}

func initQueueClientWithRedis(redisHost string, redisPort int) (*machinery.Server, error) {
	if services == nil {
		services = &Services{}
	}

	if redisHost == "" || redisPort == 0 {
		return nil, fmt.Errorf("invalid queue redis host %s port %d", redisHost, redisPort)
	}

	// format: redis://[password@]host[port][/db_num]
	// Todo: Add password support for other environments.
	redisConnectionString := fmt.Sprintf("redis://%s:%d", redisHost, redisPort)

	config := &machineryConfig.Config{
		Broker: redisConnectionString,
		// No default queue configured. Queue name is decided conditionaly
		// and given on sendTask (enqueue) as routing_key and
		// on customer worker (dequeue) as queue_name.
		// DefaultQueue: "default_queue"
		Redis: &machineryConfig.RedisConfig{
			MaxActive: 300,
			MaxIdle:   100,
		},
		// Result Backend creates individual keys for each task
		// with the state after processing. Expiring the keys in 2 mins.
		// Retry is not using or affected by this, It is using a
		// seperate internal queue.
		ResultBackend:   redisConnectionString,
		ResultsExpireIn: 2 * 60,
		Debug:           !IsProduction(),
	}

	return machinery.NewServer(config)
}

func InitQueueClient(redisHost string, redisPort int) error {
	client, err := initQueueClientWithRedis(redisHost, redisPort)
	if err != nil {
		return err
	}

	services.QueueClient = client

	return nil
}

// InitDuplicateQueueClient - Initializes queue client with duplicate
// queue's redis host and port.
func InitDuplicateQueueClient(redisHost string, redisPort int) error {
	client, err := initQueueClientWithRedis(redisHost, redisPort)
	if err != nil {
		return err
	}

	services.DuplicateQueueClient = client

	return nil
}

// isQueueDuplicationEnabled - Conditions for enabling the queue duplication.
func IsQueueDuplicationEnabled() bool {
	return configuration.EnableSDKAndIntegrationRequestQueueDuplication
}

// InitMetricsExporter Initialized Opencensus metrics exporter to collect metrics.
func InitMetricsExporter(env, appName, projectID, projectLocation string) {
	if services == nil {
		services = &Services{}
	}
	if env == "" || appName == "" || projectID == "" || projectLocation == "" {
		return
	}
	services.MetricsExporter = metrics.InitMetrics(env, appName, projectID, projectLocation)
}

// InitSmartEventMode initializes smart event mode
func InitSmartEventMode(mode bool) {
	configuration.DryRunCRMSmartEvent = mode
}

// initializes smart properties mode
func InitSmartPropertiesMode(mode bool) {
	configuration.DryRunSmartProperties = mode
}

// init chargebee
func InitChargebeeObject(apiKey, siteName string) {
	chargebee.Configure(apiKey, siteName)
}

// SetIsBeamPipeline Sets variable to indicate that the job is running from a beam pipeline.
func SetIsBeamPipeline() {
	configuration.IsBeamPipeline = true
}

// IsBeamPipeline Returns is the beam pipeline variable is set.
func IsBeamPipeline() bool {
	return configuration.IsBeamPipeline
}

type SentryInfo struct {
	logCtx     log.Entry
	occurences int
	Fields     log.Fields
	Stacktrace string
}

type SentryErrorHook struct {
	appName string
}

func (h *SentryErrorHook) Levels() []log.Level {
	return []log.Level{
		log.ErrorLevel,
		log.PanicLevel,
		log.FatalLevel,
	}
}

func (h *SentryErrorHook) Fire(logCtx *log.Entry) error {
	WriteToLogErrors(logCtx, h.appName)
	return nil
}

func WriteToLogErrors(logCtx *log.Entry, appName string) {
	services.logErrorsLock.RLock()
	defer services.logErrorsLock.RUnlock()

	if services == nil {
		services = &Services{}
	}

	if services.logErrors == nil {
		services.logErrors = make(map[string]*SentryInfo)
	}

	_, errorExists := services.logErrors[logCtx.Message]
	if !errorExists {
		services.logErrors[logCtx.Message] = &SentryInfo{
			logCtx:     *logCtx,
			occurences: 1,
			Fields:     logCtx.Data,
			Stacktrace: string(debug.Stack()),
		}
	} else {
		// increment the ocurences of the error message if it repeats
		services.logErrors[logCtx.Message].occurences = services.logErrors[logCtx.Message].occurences + 1
	}

	// TODO: Add log fields receive every time to an array.

	logCtx.Data[SENTRY_OCCURRENCE_KEY] = services.logErrors[logCtx.Message].occurences
	services.logErrors[logCtx.Message].Fields = logCtx.Data
}

// ForkLogErrors() - Forks a new log rollup and returns the so far collected.
func ForkLogErrors() map[string]SentryInfo {
	services.logErrorsLock.RLock()
	defer services.logErrorsLock.RUnlock()

	if services == nil {
		services = &Services{}
	}

	if services.logErrors == nil {
		services.logErrors = make(map[string]*SentryInfo)
	}

	// Copy the logs so far.
	forkedErrorLogs := make(map[string]SentryInfo)
	for k, v := range services.logErrors {
		// Not using addresses to avoid any invalid referrences.
		forkedErrorLogs[k] = *v
	}

	// Initialise a new rollup.
	services.logErrors = make(map[string]*SentryInfo)

	return forkedErrorLogs
}

// Ticker is added to send errors to sentry periodically when sentry rollup flag is true
func sendErrorsToSentry(appName string) {
	if !configuration.UseSentryRollup {
		return
	}

	ticker := time.NewTicker(time.Duration(configuration.SentryRollupSyncInSecs) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				errorsMap := ForkLogErrors()
				for message, Info := range errorsMap {
					event := sentry.NewEvent()
					event.Level = sentry.LevelError
					event.Message = message
					event.Environment = configuration.Env
					event.Tags = map[string]string{
						SENTRY_APP_KEY: appName,
					}

					for key, value := range Info.Fields {
						Info.Fields[key] = fmt.Sprintf("%+v", value)
					}
					event.Extra = Info.Fields
					event.Extra["stacktrace"] = Info.Stacktrace

					sentry.CaptureEvent(event)
				}

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func initSentryRollup(sentryDSN, appName string) {
	if !configuration.UseSentryRollup {
		log.Info("Sentry rollup not enabled but called.")
		return
	}

	if services == nil {
		services = &Services{}
	}

	if services.logErrors == nil {
		services.logErrors = make(map[string]*SentryInfo)
	}

	if !configuration.UseSentryRollup {
		return
	}

	// initalizing sentry client
	err := sentry.Init(sentry.ClientOptions{
		Dsn: sentryDSN,
	})
	if err != nil {
		log.Fatal("Sentry rollup initialization failed.")
	}
	// initalizing and adding the sentry hook with specified Levels() and Fire() mechanism
	sentryHook := &SentryErrorHook{appName: appName}
	log.AddHook(sentryHook)
}

func IsInitialized(currentHub *sentry.Hub) bool {
	return currentHub.Client() != nil
}
func InitSentry(sentryDSN string) {
	if IsInitialized(sentry.CurrentHub()) {
		return
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: sentryDSN,
	}); err != nil {
		log.WithError(err).Error("Sentry initialization failed.")
	}
}

// InitSentryLogging Adds sentry hook to capture error logs.
func InitSentryLogging(sentryDSN, appName string) {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	if IsDevelopment() {
		log.SetLevel(log.DebugLevel)
	}
	log.SetReportCaller(true)

	if sentryDSN == "" {
		return
	}

	// Use rollup for sentry capturing.
	if configuration.UseSentryRollup {
		initSentryRollup(sentryDSN, appName)
		sendErrorsToSentry(appName)
		return
	}

	if services == nil {
		services = &Services{}
	}

	sentryHook, err := logrus_sentry.NewAsyncSentryHook(sentryDSN, []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
	})
	if err != nil {
		log.WithError(err).Error("Failed to init sentry webhook")
	} else {
		sentryHook.SetEnvironment(configuration.Env)
		sentryHook.StacktraceConfiguration.Enable = true
		sentryHook.StacktraceConfiguration.SwitchExceptionTypeAndMessage = true
		sentryHook.StacktraceConfiguration.IncludeErrorBreadcrumb = true

		sentryHook.SetTagsContext(map[string]string{
			"AppName":   appName,
			"Datastore": GetPrimaryDatastore(),
		})

		services.SentryHook = sentryHook
		log.AddHook(sentryHook)
		log.Info("Sentry error campturing initialized.")
	}
}

// SafeFlushAllCollectors Safe flush sentry and metrics collectors. Used with `defer` statement.
// Useful while running scripts in development mode where sentry is not initialized.
func SafeFlushAllCollectors() {
	if services != nil {
		if services.SentryHook != nil {
			services.SentryHook.Flush()
		}

		if services.MetricsExporter != nil {
			services.MetricsExporter.StopMetricsExporter()
			services.MetricsExporter.Flush()
		}
	}
}

// WaitAndFlushAllCollectors Waits for given period before flushing and terminating.
// Added as a hack to export metrics before program ends.
func WaitAndFlushAllCollectors(waitPeriod time.Duration) {
	time.Sleep(waitPeriod)
	SafeFlushAllCollectors()
}

func InitMailClient(key, secret, region string) {
	if services == nil {
		services = &Services{}
	}
	if IsDevelopment() {
		services.Mailer = mailer.New()
		return
	}
	services.Mailer = serviceSes.New(key, secret, region)
}

func InitSenderEmail(senderEmail string) {
	if services == nil {
		services = &Services{}
	}
	configuration.EmailSender = senderEmail
}

func initCollectorClient(env, appName, toMail, fromMail string, reportingInterval int) {
	if services == nil {
		services = &Services{}
	}
	dur := time.Second * time.Duration(reportingInterval)
	services.ErrorCollector = error_collector.New(services.Mailer, dur, env, appName, toMail, fromMail)
}

func watchPatternServers(psUpdateChannel clientv3.WatchChan) {
	log.Infoln("Starting to watch on psUpdateChannel")
	for {
		msg := <-psUpdateChannel
		for _, event := range msg.Events {
			log.WithFields(log.Fields{
				"UnitType": event.Type,
				"Key":      string(event.Kv.Key),
				"Value":    string(event.Kv.Value),
			}).Infoln("Event Received on PatternServerUpdateChannel")

			if event.Type == mvccpb.PUT {
				GetServices().addPatternServer(string(event.Kv.Key), string(event.Kv.Value))
			} else if event.Type == mvccpb.DELETE {
				GetServices().removePatternServer(string(event.Kv.Key))
			}
		}
		log.WithField("PatternServers", GetServices().GetPatternServerAddresses()).Info(
			"Updated List of pattern servers")
	}
}

func InitAppServer(config *Configuration) error {
	if !IsConfigInitialized() {
		log.Fatal("Config not initialised on Init.")
	}

	err := initAppServerServices(config)
	if err != nil {
		return err
	}

	return nil
}

func InitMonitoringAPIServices(config *Configuration) {
	if config.MonitoringAPIToken == "" {
		log.Error("Monitoring API Token is not provided. Keeping services disabled.")
		return
	}

	err := InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Error("Failed to initalize queue client.")
	}

	if IsQueueDuplicationEnabled() {
		err := InitDuplicateQueueClient(config.DuplicateQueueRedisHost, config.DuplicateQueueRedisPort)
		if err != nil {
			log.WithError(err).Error("Failed to initialize duplicate queue client.")
		}
	}
}

func InitTestServer(config *Configuration) error {
	if !IsConfigInitialized() {
		log.Fatal("Config not initialised on Init.")
	}

	err := initAppServerServices(config)
	if err != nil {
		return err
	}

	return nil
}

func UseQueueRedis() bool {
	return configuration.UseQueueRedis
}

// UseOpportunityAssociationByProjectID should use salesforce association for opportunity stitching
func UseOpportunityAssociationByProjectID(projectID int64) bool {
	if configuration.UseOpportunityAssociationByProjectID == "" {
		return false
	}

	if configuration.UseOpportunityAssociationByProjectID == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.UseOpportunityAssociationByProjectID, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}

// UseSourcePropertyOverwriteByProjectIDs should use property overwrite by source
func UseSourcePropertyOverwriteByProjectIDs(projectID int64) bool {
	if configuration.UseSourcePropertyOverwriteByProjectIDs == "" {
		return false
	}

	if configuration.UseSourcePropertyOverwriteByProjectIDs == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.UseSourcePropertyOverwriteByProjectIDs, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}

// AllowSupportForUserPropertiesInIdentityCall id used to check if support for user properties
// is allowed for a given (or list of) project
func AllowSupportForUserPropertiesInIdentifyCall(projectID int64) bool {
	if configuration.AllowSupportForUserPropertiesInIdentifyCall == "" {
		return false
	}

	if configuration.AllowSupportForUserPropertiesInIdentifyCall == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.AllowSupportForUserPropertiesInIdentifyCall, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}

// EnableEventLevelEventProperties is used to check if the event level properties
// are to be enabled for a given (or list of) project
func EnableEventLevelEventProperties(projectID int64) bool {
	if configuration.EnableEventLevelEventProperties == "" {
		return false
	}

	if configuration.EnableEventLevelEventProperties == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.EnableEventLevelEventProperties, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}

// EnableOLTPQueriesMemSQLImprovements is used to check if the OLTP queries performance improvements
// for memsql are to be enabled for a given (or list of) project
func EnableOLTPQueriesMemSQLImprovements(projectID int64) bool {
	if configuration.EnableOLTPQueriesMemSQLImprovements == "" {
		return false
	}

	if configuration.EnableOLTPQueriesMemSQLImprovements == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.EnableOLTPQueriesMemSQLImprovements, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}

func InitDataService(config *Configuration) error {
	if !IsConfigInitialized() {
		log.Fatal("Config not initialised on InitDataService.")
	}

	err := InitDB(*config)
	if err != nil {
		return err
	}
	InitRedis(config.RedisHost, config.RedisPort)
	InitSentryLogging(config.SentryDSN, config.AppName)
	InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)

	return nil
}

func InitSDKService(config *Configuration) error {
	if !IsConfigInitialized() {
		log.Fatal("Config not initialised on InitSDKService.")
	}

	// DB dependency for SDK project_settings.
	if err := InitDB(*config); err != nil {
		log.WithError(err).Error("Failed to initialize db on sdk_service.")
	}

	// Cache dependency for requests not using queue.
	InitRedis(config.RedisHost, config.RedisPort)

	InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	InitQueueRedis(config.QueueRedisHost, config.QueueRedisPort)

	initGeoLocationService(config.GeolocationFile)
	initDeviceDetectorPath(config.DeviceDetectorPath)

	err := InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize queue client on init sdk service.")
	}

	if IsQueueDuplicationEnabled() {
		err := InitDuplicateQueueClient(config.DuplicateQueueRedisHost, config.DuplicateQueueRedisPort)
		if err != nil {
			log.WithError(err).Fatal("Failed to initialize duplicate queue client on init sdk service.")
		}
	}

	InitSentryLogging(config.SentryDSN, config.AppName)
	InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)

	return nil
}

func InitQueueWorker(config *Configuration, concurrency int) error {
	if !IsConfigInitialized() {
		log.Fatal("Config not initialised on InitSDKService.")
	}

	err := InitDBWithMaxIdleAndMaxOpenConn(*config, concurrency, concurrency)
	if err != nil {
		return err
	}
	InitRedis(config.RedisHost, config.RedisPort)

	initGeoLocationService(config.GeolocationFile)
	initDeviceDetectorPath(config.DeviceDetectorPath)

	err = InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Fatal("Failed to initalize queue client on init queue worker.")
	}

	if IsQueueDuplicationEnabled() {
		err := InitDuplicateQueueClient(config.DuplicateQueueRedisHost, config.DuplicateQueueRedisPort)
		if err != nil {
			log.WithError(err).Fatal("Failed to initialize duplicate queue client on init queue worker..")
		}
	}

	InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	InitSentryLogging(config.SentryDSN, config.AppName)
	InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)

	return nil
}

func GetConfig() *Configuration {
	return configuration
}

func GetServices() *Services {
	return services
}

func GetCacheRedisConnection() redis.Conn {
	return services.Redis.Get()
}

func GetCacheRedisPersistentConnection() redis.Conn {
	return services.RedisPeristent.Get()
}

func GetCacheQueueRedisConnection() redis.Conn {
	return services.QueueRedis.Get()
}

func IsDevelopment() bool {
	return (strings.Compare(configuration.Env, DEVELOPMENT) == 0)
}

func IsStaging() bool {
	return (strings.Compare(configuration.Env, STAGING) == 0)
}

func IsProduction() bool {
	return (strings.Compare(configuration.Env, PRODUCTION) == 0)
}

func GetFactorsSixSignalAPIKey() string {
	return configuration.FactorsSixSignalAPIKey
}

func GetClearbitProvisionAccountAPIKey() string {
	return configuration.ClearbitProvisionAccountAPIKey
}

func GetMailmodoTriggerCampaignAPIKey() string {
	return configuration.MailModoTriggerCampaignAPIKey
}

func GetAPPDomain() string {
	return configuration.APPDomain
}

func GetAPPOldDomain() string {
	return configuration.APPOldDomain
}

func GetAPIDomain() string {
	return configuration.APIDomain
}

func UseSecureCookie() bool {
	return !IsDevelopment()
}

func UseHTTPOnlyCookie() bool {
	return !IsDevelopment()
}

func GetProtocol() string {
	if IsDevelopment() {
		return "http://"
	}
	return "https://"
}

func GetFacebookAppId() string {
	return configuration.FacebookAppID
}

func GetFacebookAppSecret() string {
	return configuration.FacebookAppSecret
}
func GetLinkedinClientID() string {
	return configuration.LinkedinClientID
}

func GetLinkedinClientSecret() string {
	return configuration.LinkedinClientSecret
}

func GetSalesforceAppId() string {
	return configuration.SalesforceAppID
}

func GetSalesforceAppSecret() string {
	return configuration.SalesforceAppSecret
}

func GetFactorsSenderEmail() string {
	return configuration.EmailSender
}

// IsDryRunCRMSmartEvent checks if dry run flag is set
func IsDryRunCRMSmartEvent() bool {
	return configuration.DryRunCRMSmartEvent
}

// IsDryRunSmartProperties checks if dry run flag is set
func IsDryRunSmartProperties() bool {
	return configuration.DryRunSmartProperties
}

func GetCookieDomian() string {
	domain := GetAPIDomain()
	port := ":" + strconv.Itoa(configuration.Port)
	if strings.Contains(domain, port) {
		return strings.Split(domain, port)[0]
	}
	return domain
}

func GetFactorsCookieName() string {
	return configuration.Cookiename
}

func GetAuth0StateCookieName() string {
	return configuration.Auth0StateName
}
func GetSkipTrackProjectIds() []int64 {
	return configuration.SkipTrackProjectIds
}

func GetLookbackWindowForEventUserCache() int {
	return configuration.LookbackWindowForEventUserCache
}

func GetFactorsTrackedEventsLimit() int {
	return configuration.ActiveFactorsTrackedEventsLimit
}

func GetFactorsTrackedUserPropertiesLimit() int {
	return configuration.ActiveFactorsTrackedUserPropertiesLimit
}

func GetFactorsGoalsLimit() int {
	return configuration.ActiveFactorsGoalsLimit
}

func IsAllowedSmartEventRuleCreation() bool {
	return configuration.AllowSmartEventRuleCreation
}

func ExtractProjectIdDateFromConfig(config string) map[int64]time.Time {
	convertedMap := ParseConfigStringToMap(config)
	projectIdDateMap := make(map[int64]time.Time)
	for projectId, dateString := range convertedMap {
		projId, _ := strconv.Atoi(projectId)
		date, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, dateString)
		projectIdDateMap[int64(projId)] = date
	}
	return projectIdDateMap
}

// ParseConfigStringToMap - Parses config string
// "k1:v1,k2:v2"-> map[string]string{k1: v1, k2: v2}
func ParseConfigStringToMap(configStr string) map[string]string {
	configMap := make(map[string]string, 0)

	if configStr == "" {
		return configMap
	}

	commaSplit := strings.Split(configStr, ",")
	if len(commaSplit) == 0 {
		return configMap
	}

	for _, cs := range commaSplit {
		kv := strings.Split(cs, ":")
		if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
			configMap[kv[0]] = kv[1]
		}
	}

	return configMap
}

func ParseProjectIDToStringMapFromConfig(configValue, configName string) map[int64]string {
	cMap := make(map[int64]string, 0)

	cStringMap := ParseConfigStringToMap(configValue)

	for projectIDString, customerUserID := range cStringMap {
		projectID, err := strconv.ParseInt(projectIDString, 10, 64)
		if err != nil {
			log.WithError(err).WithField("value", configValue).
				Fatal("Invalid project_id on ParseProjectIDToStringMapFromConfig from %s", configName)
		}

		customerUserID = strings.TrimSpace(customerUserID)
		if customerUserID != "" {
			cMap[projectID] = customerUserID
		}
	}

	return cMap
}

func IsSegmentExcludedCustomerUserID(projectID int64, sourceCustomerUserID string) bool {
	customerUserID, projectExists := configuration.SegmentExcludedCustomerIDByProject[projectID]
	return projectExists && customerUserID == sourceCustomerUserID
}

func GetTokensFromStringListAsUint64(stringList string) []int64 {
	uint64Tokens := make([]int64, 0, 0)

	if stringList == "" {
		return uint64Tokens
	}

	tokens := strings.Split(stringList, ",")
	for _, token := range tokens {
		uint64Token, err := strconv.ParseInt(strings.TrimSpace(token), 10, 64)
		if err != nil {
			log.WithError(err).
				Error("Failed to parse project_id on string list config.")
			return uint64Tokens
		}

		uint64Tokens = append(uint64Tokens, uint64Token)
	}

	return uint64Tokens
}

func GetTokensFromStringListAsString(stringList string) []string {
	stringTokens := make([]string, 0, 0)

	if stringList == "" {
		return stringTokens
	}

	tokens := strings.Split(stringList, ",")
	for _, token := range tokens {
		stringTokens = append(stringTokens, strings.TrimSpace(token))
	}

	return stringTokens
}

func GetAttributionDebug() int {
	return configuration.AttributionDebug
}

func IsAllowedAttributionCommonFlow(projectID int64) bool {
	if configuration.AttributionCommonFlow == "" {
		return false
	}

	if configuration.AttributionCommonFlow == "*" {
		return true
	}

	projectIDStr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.AttributionCommonFlow, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDStr {
			return true
		}
	}

	return false
}

func GetAttributionDebugKPI() string {
	return configuration.AttributionDebugKPI
}

func IsAllowedAttributionDBCacheLookup(projectID int64) bool {
	if configuration.AttributionDBCacheLookup == "" {
		return false
	}

	if configuration.AttributionDBCacheLookup == "*" {
		return true
	}

	projectIDStr := fmt.Sprintf("%d", projectID)
	projectIDs := strings.Split(configuration.AttributionDBCacheLookup, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDStr {
			return true
		}
	}

	return false
}

func GetOtpKeyWithQueryCheckEnabled() bool {
	return configuration.OtpKeyWithQueryCheckEnabled
}

func GetOnlyAttributionDashboardCaching() int {
	return configuration.OnlyAttributionDashboardCaching
}

func GetIsRunningForMemsql() int {
	return configuration.IsRunningForMemsql
}

func GetIsHourlyRunEnabled() int {
	return configuration.IsHourlyRunEnabled
}

func GetSkipAttributionDashboardCaching() int {
	return configuration.SkipAttributionDashboardCaching
}

func IsProjectIDSkippedForOtp(projectId int64) bool {

	if configuration.SkippedOtpProjectIDs == "" {
		return false
	}

	if configuration.SkippedOtpProjectIDs == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectId)
	projectIDs := strings.Split(configuration.SkippedOtpProjectIDs, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}

func GetSDKRequestQueueAllowedTokens() []string {
	return configuration.SDKRequestQueueProjectTokens
}

func GetSegmentRequestQueueAllowedTokens() []string {
	return configuration.SegmentRequestQueueProjectTokens
}

func GetFivetranGroupId() string {
	return configuration.FivetranGroupId
}

func GetFivetranLicenseKey() string {
	return configuration.FivetranLicenseKey
}

/*
GetProjectsFromListWithAllProjectSupport -
If project list string is '*':

	Returns all_projects as true and empty allowed projects and disallowed projects.

else:

	  Returns all_projects as false, given projects ids after skipping disallowed
		projects and disallowed projects.

Returns: allProject flag, map of allowed & disallowed projects
*/
func GetProjectsFromListWithAllProjectSupport(projectIdsList,
	disallowedProjectIdsList string) (allProjects bool, allowedMap, disallowedMap map[int64]bool) {
	//allowedProjectIds, skipProjectIds []uint64,
	disallowedProjectIdsList = strings.TrimSpace(disallowedProjectIdsList)
	skipProjectIds := GetTokensFromStringListAsUint64(disallowedProjectIdsList)

	disallowedMap = make(map[int64]bool)
	for i := range skipProjectIds {
		disallowedMap[skipProjectIds[i]] = true
	}

	projectIdsList = strings.TrimSpace(projectIdsList)
	if projectIdsList == "*" {
		return true, map[int64]bool{}, disallowedMap
	}

	projectIds := GetTokensFromStringListAsUint64(projectIdsList)

	allowedProjectIds := make([]int64, 0, len(projectIds))
	for i, cpid := range projectIds {
		//Prioritizing the skip list over project list!
		if _, exists := disallowedMap[cpid]; !exists {
			allowedProjectIds = append(allowedProjectIds, projectIds[i])
		}
	}

	allowedMap = make(map[int64]bool)
	for i := range allowedProjectIds {
		allowedMap[allowedProjectIds[i]] = true
	}

	return false, allowedMap, disallowedMap
}

func GetDashboardUnitIDs(dashboardUnitIDsList string) []int64 {
	dashboardUnitIDsList = strings.TrimSpace(dashboardUnitIDsList)
	if dashboardUnitIDsList == "*" {
		return make([]int64, 0, 0)
	}
	return GetTokensFromStringListAsUint64(dashboardUnitIDsList)
}

func ProjectIdsFromProjectIdBoolMap(mp map[int64]bool) []int64 {

	keys := make([]int64, 0, len(mp))
	for k := range mp {
		keys = append(keys, k)
	}
	return keys
}

// IsBlockedSDKRequestProjectToken - Tells whether to block the sdk request or
// not, based on given token and list of blocked_sdk_requests_project_tokens.
func IsBlockedSDKRequestProjectToken(projectToken string) bool {
	if projectToken == "" {
		return true
	}

	return U.StringValueIn(projectToken, configuration.BlockedSDKRequestProjectTokens)
}

// IsIPBlockingFeatureEnabled - Enables the feature of blocking
// IP for a project, based on given token and list of blocked_ip_project_tokens.
func IsIPBlockingFeatureEnabled(token string) bool {
	if configuration.BlockedIPProjectTokens == "" {
		return false
	}

	if configuration.BlockedIPProjectTokens == "*" {
		return true
	}

	enabledTokens := strings.Split(configuration.BlockedIPProjectTokens, ",")
	for i := range enabledTokens {
		if strings.TrimSpace(enabledTokens[i]) == token {
			return true
		}
	}

	return false
}

// IsDomainEnabled - Checks if $domain is enabled for given project_id in all accounts
func IsDomainEnabled(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AllAccountsProjectId, "")
	if allProjects || projectIDsMap[projectID] {
		return true
	}
	return false
}

// IsTimelinesTablePropsOptEnabled - Checks if optimised query to fetch table properties in Timelines enabled
func IsTimelinesTablePropsOptEnabled(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().TimelinesTablePropsQueryOpt, "")
	if allProjects || projectIDsMap[projectID] {
		return true
	}
	return false
}

// IsMarkerPreviewEnabled - Checks if marker is enabled for all accounts preview
func IsMarkerPreviewEnabled(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().MarkerPreviewAllAccountsProjectId, "")
	if allProjects || projectIDsMap[projectID] {
		return true
	}
	return false
}

func BatchSizePreviewtMarker() int {
	return configuration.BatchSizePreviewDomain
}

func DomainsToProcessForPreview() int {
	return configuration.AccountsToProcessForPreview
}
func RunNumberToProcessDomainsForPreview() int {
	return configuration.NumberOfRunsForPreview
}
func AccountLimitForPreview() int {
	return configuration.AccountLimitPreviewListing
}

// UseSegmentMarker - Checks if segment marker is enabled for given project_id in all accounts listing
func UseSegmentMarker(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().UseMarkerByProjectID, "")
	if allProjects || projectIDsMap[projectID] {
		return true
	}
	return false
}

// AllAccountsRuntMarker - Checks if segment marker is to be run for all $domains for given project
func AllAccountsRuntMarker(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().RunAllAccountsMarkerProjectIDs, "")
	if allProjects || projectIDsMap[projectID] {
		return true
	}
	return false
}

// DisableAccountsRuntMarker - Checks if segment marker is to not be run for all $domains for given project
func DisableAccountsRuntMarker(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().DisableAllAccountsMarkerProjectIDs, "")
	if allProjects || projectIDsMap[projectID] {
		return true
	}
	return false
}

// UseOptimisedEventsQueryProjectIDs - Checks if segment marker should use optimised query for event filters check for given project
func MarkerEnableOptimisedEventsQuery(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().UseOptimisedEventsQueryProjectIDs, "")
	if allProjects || projectIDsMap[projectID] {
		return true
	}
	return false
}

// EnableLatestSegmentsMarkerRun - Checks if segment marker should run all run for recently updated/created segments
func EnableLatestSegmentsMarkerRun(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().LatestSegmentsMarkerRun, "")
	if allProjects || projectIDsMap[projectID] {
		return true
	}
	return false
}

func IsScoringEnabledForAllUsers(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableScoringByProjectID, "")
	if allProjects || projectIDsMap[projectID] {
		return true
	}
	return false
}

// IsAllAccountsEnabled - Checks if $domain is enabled for given project_id in all accounts
func IsAllAccountsEnabled(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableNewAllAccountsByProjectID, "")
	if allProjects || projectIDsMap[projectID] {
		return true
	}
	return false
}

func IsUpdateLastEventAtEnabled(projectID int64) bool {
	return true
}

// PingHealthcheckForSuccess Ping healthchecks.io for cron success.
func PingHealthcheckForSuccess(healthcheckID string, message interface{}) {
	log.Info("Job successful with message ", message)
	if configuration.Env != PRODUCTION {
		return
	}
	var client = &http.Client{
		Timeout: 60 * time.Second,
	}

	payload, _ := json.MarshalIndent(message, "", " ")
	if string(payload) == "{}" {
		payload = []byte(fmt.Sprintf("%#v", message))
	}
	_, err := client.Post("https://hc-ping.com/"+healthcheckID, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.WithError(err).Error("Failed to report to healthchecks.io")
	}
}

// PingHealthcheckForStart Ping healthchecks.io for cron start. Used to show run time for jobs.
func PingHealthcheckForStart(healthcheckID string) {
	if configuration.Env != PRODUCTION {
		return
	}
	var client = &http.Client{
		Timeout: 10 * time.Second,
	}

	_, err := client.Head("https://hc-ping.com/" + healthcheckID + "/start")
	if err != nil {
		log.WithError(err).Error("Failed to report to healthchecks.io")
	}
}

// PingHealthcheckForFailure Ping healthchecks.io for cron failure.
func PingHealthcheckForFailure(healthcheckID string, message interface{}) {
	log.Error("Job failed with message ", message)
	if configuration.Env != PRODUCTION {
		return
	}
	var client = &http.Client{
		Timeout: 10 * time.Second,
	}

	payload, _ := json.MarshalIndent(message, "", " ")
	if string(payload) == "{}" {
		payload = []byte(fmt.Sprintf("%#v", message))
	}
	_, err := client.Post("https://hc-ping.com/"+healthcheckID+"/fail", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.WithError(err).Error("Failed to report to healthchecks.io")
	}
}

// PingHealthcheckForPanic To capture panics in crons and send an alert to healthcheck and SNS.
func PingHealthcheckForPanic(taskID, env, healthcheckID string) {
	if recoveredFrom := recover(); recoveredFrom != nil {
		if configuration == nil {
			// In case panic happens before conf is initialized.
			InitConf(&Configuration{Env: env})
		}
		panicMessage := map[string]interface{}{
			"panic_error": recoveredFrom,
			"stacktrace":  string(debug.Stack()),
		}
		PingHealthcheckForFailure(healthcheckID, panicMessage)

		U.NotifyOnPanicWithErrorLog(taskID, env, recoveredFrom, &log.Fields{})
	}
}

func isProjectOnProjectsList(configProjectIDList string, projectID int64) bool {
	allProjectIDs, allowedProjectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(
		configProjectIDList, "")

	if allProjectIDs {
		return true
	}

	_, exists := allowedProjectIDsMap[projectID]
	return exists
}

func isIDOnIDList(configIDList string, projectID int64) bool {
	allDs, allowedIDsMap, _ := GetProjectsFromListWithAllProjectSupport(
		configIDList, "")

	if allDs {
		return true
	}

	_, exists := allowedIDsMap[projectID]
	return exists
}

func IsFormFillIdentificationAllowedForProject(projectID int64) bool {
	return isProjectOnProjectsList(configuration.FormFillIdentificationAllowedProjects, projectID)
}

func IsEmailUTMParameterAllowed(projectID int64) bool {
	return isProjectOnProjectsList(configuration.EmailUTMParameterAllowedProjects, projectID)
}

func IsChannelGroupingAllowed(projectID int64) bool {
	return isProjectOnProjectsList(configuration.AllowChannelGroupingForProjectIDs, projectID)
}

func IsSDKAndIntegrationRequestQueueDuplicationEnabled() bool {
	return configuration.EnableSDKAndIntegrationRequestQueueDuplication
}

func GetSDKAndIntegrationMetricNameByConfig(metricName string) string {
	if IsSDKAndIntegrationRequestQueueDuplicationEnabled() {
		metricName = "dup_" + metricName
	}

	return metricName
}

func IsSortedSetCachingAllowed() bool {
	return configuration.CacheSortedSet
}

func GetUUIdsFromStringListAsString(stringList string) []string {
	stringTokens := make([]string, 0, 0)

	if stringList == "" {
		return stringTokens
	}

	uuids := strings.Split(stringList, ",")
	for _, uuid := range uuids {
		stringTokens = append(stringTokens, strings.TrimSpace(uuid))
	}

	return stringTokens
}

func IsLastComputedWhitelisted(projectId int64) bool {
	for _, id := range configuration.CustomerEnabledProjectsLastComputed {
		if id == projectId {
			return true
		}
	}

	return false
}

func IsLoggedInUserWhitelistedForProjectAnalytics(loggedInUUID string) bool {
	for _, uuid := range configuration.ProjectAnalyticsWhitelistedUUIds {
		if uuid == loggedInUUID {
			return true
		}
	}
	return false
}

func IsIngestionTimezoneEnabled(projectId int64) bool {
	for _, id := range configuration.IngestionTimezoneEnabledProjectIDs {
		projectIdString := fmt.Sprintf("%v", projectId)
		if id == projectIdString {
			return true
		}
	}
	return false
}

func EnableMQLAPI() bool {
	return configuration.EnableMQLAPI
}

// GetHealthcheckPingID - Choose between default and override ping_id
// based on availability.
func GetHealthcheckPingID(defaultPingID, overridePingID string) string {
	if overridePingID != "" {
		return overridePingID
	}

	return defaultPingID
}

// GetAppName - Choose between default and override app_name
// based on availability.
func GetAppName(defaultAppName, overrideAppName string) string {
	if overrideAppName != "" {
		return overrideAppName
	}

	return defaultAppName
}

func GetCloudManager() filestore.FileManager {
	return configuration.CloudManager
}

func DisableDashboardQueryDBExecution() bool {
	return configuration.DisableDashboardQueryDBExecution
}

func IsDevBox() bool {
	return configuration.DevBox
}

func SetEnableEventLevelEventProperties(projectId int64) {
	configuration.EnableEventLevelEventProperties = fmt.Sprintf("%d", projectId)
}

func IsProfileQuerySourceSupported(projectId int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowSupportForSourceColumnInUsers, "")
	if allProjects || projectIDsMap[projectId] {
		return true
	}
	return false
}

func IsV1AvgKPIEnabled(projectId int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowSupportForV1AvgKPIComputation, "")
	if allProjects || projectIDsMap[projectId] {
		return true
	}
	return false
}

func CheckRestrictReusingUsersByCustomerUserId(projectId int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().RestrictReusingUsersByCustomerUserId, "")
	if allProjects || projectIDsMap[projectId] {
		return true
	}
	return false
}

func IsWebsiteAggregationTestEnabled(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().WebsiteAggregationTestEnabledProjects, "")
	if allProjects || projectIDsMap[projectID] {
		return true
	}
	return false
}

func AllowMergeAmpIDAndSegmentIDWithUserIDByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(configuration.MergeAmpIDAndSegmentIDWithUserIDByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func IsProfileGroupSupportEnabled(projectId int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowProfilesGroupSupport, "")
	if allProjects || projectIDsMap[projectId] {
		return true
	}
	return false
}

func IsEventsFunnelsGroupSupportEnabled(projectId int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowEventsFunnelsGroupSupport, "")
	if allProjects || projectIDsMap[projectId] {
		return true
	}
	return false
}

func GetSessionBatchTransactionBatchSize() int {
	return GetConfig().SessionBatchTransactionBatchSize
}

func DisableCRMUniquenessConstraintsCheckByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().DisableCRMUniquenessConstraintsCheckByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func GetHubspotBatchInsertBatchSize() int {
	return GetConfig().HubspotBatchInsertBatchSize
}

func GetChatDebug() int {
	return GetConfig().ChatDebug
}

func EnableHubspotFormsEventsByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableHubspotFormsEventsByProjectID, "")
	if allProjects {
		return true
	}
	return allowedProjectIDs[projectID]
}
func GetUsageBasedDashboardCaching() int {
	return configuration.UsageBasedDashboardCaching
}

func GetSkipKPICachingCaching() int {
	return configuration.SkipKPICaching
}

func GetOnlyKPICachingCaching() int {
	return configuration.OnlyKPICaching
}

func UseHubspotBatchInsertByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().UseHubspotBatchInsertByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]

}

func GetAuth0Info() Auth0Conf {
	return configuration.Auth0Info
}

func GetSessionStore() string {
	return configuration.SessionStore
}

func GetSessionStoreSecret() string {
	return configuration.SessionStoreSecret
}

func GetSalesforcePropertyLookBackTimeHr() int {
	return GetConfig().SalesforcePropertyLookBackTimeHr
}

func GetSalesforceBatchInsertBatchSize() int {
	return GetConfig().SalesforceBatchInsertBatchSize
}

func AllowHubspotEngagementsByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowHubspotEngagementsByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func DisableHubspotNonMarketingContactsByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().DisableHubspotNonMarketingContactsByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func DisableEventUserPropertiesByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().RemoveDisabledEventUserPropertiesByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func GetHubspotPropertiesLookbackLimit() int {
	return GetConfig().HubspotPropertyLookBackLimit
}

func IsSlowDBQueryLoggingEnabled() bool {
	return configuration.EnableSlowDBQueryLogging
}

func GetSlackClientID() string {
	return configuration.SlackAppClientID
}

func GetSlackClientSecret() string {
	return configuration.SlackAppClientSecret
}

func AllowSalesforcev54APIByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().UseSalesforceV54APIByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func EnableOptimisedFilterOnProfileQuery() bool {
	return configuration.EnableOptimisedFilterOnProfileQuery
}

func GetHubspotAppSecret() string {
	return configuration.HubspotAppSecret
}

func GetHubspotAppID() string {
	return configuration.HubspotAppID
}

func EnableOptimisedFilterOnEventUserQuery() bool {
	return configuration.EnableOptimisedFilterOnEventUserQuery

}

func AllowIdentificationOverwriteUsingSource(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowIdentificationOverwriteUsingSourceByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func SetAllowIdentificationOverwriteUsingSourceByProjectID(value string) {
	GetConfig().AllowIdentificationOverwriteUsingSourceByProjectID = value
}

func EnableEmailDomainBlocking() bool {
	return configuration.EnableEmailBlockingFlag
}

func EnableIPBlocking() bool {
	return configuration.EnableIPBlockingFlag
}

func GetDBMaxAllowedPacket() int64 {
	return configuration.DBMaxAllowedPacket
}

func GetBlockedEmailFromStringListAsString(stringList string) []string {
	emailStringList := make([]string, 0, 0)

	if stringList == "" {
		return emailStringList
	}

	emails := strings.Split(stringList, ",")
	for _, email := range emails {
		emailStringList = append(emailStringList, strings.TrimSpace(email))
	}

	return emailStringList
}

func GetBlockedIPFromStringListAsString(stringList string) []string {
	IPStringList := make([]string, 0, 0)

	if stringList == "" {
		return IPStringList
	}

	ips := strings.Split(stringList, ",")
	for _, ip := range ips {
		IPStringList = append(IPStringList, strings.TrimSpace(ip))
	}

	return IPStringList
}

func GetBlockedEmailDomainFromStringListAsString(stringList string) []string {
	domStringList := make([]string, 0, 0)

	if stringList == "" {
		return domStringList
	}

	domains := strings.Split(stringList, ",")
	for _, dom := range domains {
		domStringList = append(domStringList, strings.TrimSpace(dom))
	}

	return domStringList
}

func PastEventEnrichmentEnabled(projectId int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowHubspotPastEventsEnrichmentByProjectID, "")
	if allProjects {
		return true
	}

	return projectIDsMap[projectId]
}

func ContactListInsertEnabled(projectId int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowHubspotContactListInsertByProjectID, "")
	if allProjects {
		return true
	}

	return projectIDsMap[projectId]
}

func IsAllowedSalesforceActivityTasksByProjectID(projectId int64) bool {
	allProjects, allowedProjects, disabledProjects := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowedSalesforceActivityTasksByProjectIDs, GetConfig().DisallowedSalesforceActivityTasksByProjectIDs)
	if exists := disabledProjects[projectId]; exists {
		return false
	}

	if !allProjects {
		if _, exists := allowedProjects[projectId]; !exists {
			return false
		}
	}

	return true
}

func IsAllowedSalesforceActivityEventsByProjectID(projectId int64) bool {
	allProjects, allowedProjects, disabledProjects := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowedSalesforceActivityEventsByProjectIDs, GetConfig().DisallowedSalesforceActivityEventsByProjectIDs)
	if exists := disabledProjects[projectId]; exists {
		return false
	}

	if !allProjects {
		if _, exists := allowedProjects[projectId]; !exists {
			return false
		}
	}

	return true
}

func EnableUserLevelEventPullForAddSessionByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableUserLevelEventPullForAddSessionByProjectID, "")
	if allProjects {
		return true
	}
	return allowedProjectIDs[projectID]
}

func IsEnabledFeatureGates() bool {
	return configuration.EnableFeatureGates
}

func IsEnabledFeatureGatesV2() bool {
	return configuration.EnableFeatureGatesV2
}

func GetSlackWebhookUrlForInternalAlerts() string {
	return configuration.SlackInternalAlertWebhookUrl
}

func GetChargebeeApiKey() string {
	return configuration.ChargebeeApiKey
}

func GetChargebeeSiteName() string {
	return configuration.ChargebeeSiteName
}

func SkipKpiResultValidation() bool {
	return configuration.SkipKpiResultValidation
}

func EnableSixSignalGroupByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableSixSignalGroupByProjectID, "")
	if allProjects {
		return true
	}
	return allowedProjectIDs[projectID]
}

func IsEnableDebuggingForIP() bool {
	return configuration.EnableDebuggingForIP
}

func GetTeamsClientID() string {
	return configuration.TeamsAppClientID
}

func GetTeamsClientSecret() string {
	return configuration.TeamsAppClientSecret
}

func GetTeamsTenantID() string {
	return configuration.TeamsAppTenantID
}
func GetTeamsApplicationID() string {
	return configuration.TeamsApplicationID
}

func IsAllowedDomainsGroupByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableDomainsGroupByProjectID, "")
	if allProjects {
		return true
	}
	return allowedProjectIDs[projectID]
}

func AllowSyncReferenceFields(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableSyncReferenceFieldsByProjectID, "")
	if allProjects {
		return true
	}
	return allowedProjectIDs[projectID]
}

func GetStartTimestampForWeekMonth() int64 {
	return configuration.StartTimestampForWeekMonth
}

func GetStartTimestampForCustomRun() int64 {
	return configuration.CustomDateStart
}

func GetEndTimestampForCustomRun() int64 {
	return configuration.CustomDateEnd
}

func IsProjectAllowedForLongerExpiry(projectId int64) bool {
	return isProjectOnProjectsList(configuration.CacheForLongerExpiryProjects, projectId)
}

func IsDashboardAllowedForCaching(dashboardID int64) bool {
	return isIDOnIDList(configuration.CacheOnlyDashboards, dashboardID)
}

func IsDashboardUnitAllowedForCaching(dashboardID int64) bool {
	return isIDOnIDList(configuration.CacheOnlyDashboardUnits, dashboardID)
}

func IsSalesforceDocTypeEnabledForSync(docType string) bool {
	allowedSalesforceDocTypesForSync := GetConfig().AllowedSalesforceSyncDocTypes
	if allowedSalesforceDocTypesForSync == "" {
		return false
	}

	if allowedSalesforceDocTypesForSync == "*" {
		return true
	}

	allowedDocTypes := strings.Split(allowedSalesforceDocTypesForSync, ",")
	return U.StringValueIn(docType, allowedDocTypes)
}

func IsFieldsSyncAllowedForProjectID(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableFieldsSyncByProjectID, "")
	if allProjects {
		return true
	}

	return projectIDsMap[projectID]
}

func EnableUserDomainsGroupByProjectID(projectID int64) bool {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableUserDomainsGroupByProjectID, "")
	if allProjects {
		return true
	}

	return projectIDsMap[projectID]
}

func AllowHubspotCompaniesv3APIByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().UseHubspotCompaniesV3APIByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func AllowHubspotEngagementsv3APIByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().UseHubspotEngagementsV3APIByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func AllowEventAnalyticsGroupsByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowEventAnalyticsGroupsByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func AllowEmailDomainsByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AllowEmailDomainsByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func AllowHubspotDealsv3APIByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().UseHubspotDealsV3APIByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func AllowDeviceServiceByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableDeviceServiceByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func DisableOpportunityContactRolesEnrichmentByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().DisableOpportunityContactRolesByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

// IsExcludeBotIPV4AddressByRange exclude ipv4 address by CIDR range
func IsExcludeBotIPV4AddressByRange(ip string) bool {
	if GetConfig().ExcludeBotIPV4AddressByRange == "" {
		return false
	}

	ipRanges := strings.TrimSpace(GetConfig().ExcludeBotIPV4AddressByRange)

	cidrRanges := GetTokensFromStringListAsString(ipRanges)

	for i := range cidrRanges {
		if U.IsIPV4AddressInCIDRRange(cidrRanges[i], ip) {
			return true
		}
	}

	return false
}

func UseHashIDForCRMGroupUserByProject(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().UseHashIDForCRMGroupUserByProject, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func MoveHubspotCompanyAssocationFlowToContactByPojectID(projecID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().MoveHubspotCompanyAssocationFlowToContactByPojectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projecID]
}

func IsUserPropertyUpdateOptProject(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().UserPropertyUpdateOptProjects, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func AssociateDealToDomainByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AssociateDealToDomainByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func IsSyncTriesEnabled() bool {
	return configuration.EnableSyncTriesFlag
}

func SalesforceSkipLeadUpdatesProcessingByProjectID(projectID int64) bool {
	_, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().SalesforceSkipLeadUpdatesProcessingByProjectID, "")

	return allowedProjectIDs[projectID]
}

func SalesforceAllowOpportunityOverrideCreateCreatedEvent(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().SalesforceAllowOpportunityOverrideCreateCreatedEvent, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func IsAggrEventPropertyValuesCacheEnabled(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AggrEventPropertyValuesCacheByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func GetParagonTokenSigningKey() string {
	return configuration.ParagonTokenSigningKey
}

func GetParagonProjectID() string {
	return configuration.ParagonProjectID
}

func AddCRMObjectURLPropertyByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().AddCRMObjectURLPropertyByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func EnableTotalSessionPropertiesV2ByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableTotalSessionPropertiesV2ByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func EnableDomainWebsitePropertiesByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableDomainWebsitePropertiesByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func HubspotEnrichSkipContactUpdatesByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().HubspotEnrichSkipContactUpdatesByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func EnableSalesforceDeletedRecordByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableSalesforceDeletedRecordByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func IsEnrichmentDebugLogsEnabled(projectId int64) bool {

	if configuration.EnableEnrichmentDebugLogsByProjectID == "" {
		return false
	}

	if configuration.EnableEnrichmentDebugLogsByProjectID == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectId)
	projectIDs := strings.Split(configuration.EnableEnrichmentDebugLogsByProjectID, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}

func IsCacheDBWriteEnabled(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableCacheDBWriteProjects, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func IsCacheDBReadEnabled(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableCacheDBReadProjects, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func SkipSalesforceLeadEnrichmentByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().SkipSalesforceLeadEnrichmentByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func IsSalesforceEnabledEnrichObject(docTypeAlias string) bool {
	objects := strings.TrimSpace(GetConfig().SalesforceEnrichOnlyObjects)
	if objects == "" {
		return true
	}

	if strings.Contains(GetConfig().SalesforceEnrichOnlyObjects, docTypeAlias) {
		return true
	}
	return false
}

func IsSalesforceEnabledEnrichObjectSet() bool {
	return strings.TrimSpace(GetConfig().SalesforceEnrichOnlyObjects) != ""
}

func EnableSalesforceRelationshipPullByProjectID(projectID int64) bool {
	allProjects, allowedProjectIDs, _ := GetProjectsFromListWithAllProjectSupport(GetConfig().EnableSalesforceRelationshipPullByProjectID, "")
	if allProjects {
		return true
	}

	return allowedProjectIDs[projectID]
}

func IsSixSignalV3Enabled(projectId int64) bool {

	if configuration.SixSignalV3ProjectIds == "" {
		return false
	}

	if configuration.SixSignalV3ProjectIds == "*" {
		return true
	}

	projectIDstr := fmt.Sprintf("%d", projectId)
	projectIDs := strings.Split(configuration.SixSignalV3ProjectIds, ",")
	for i := range projectIDs {
		if projectIDs[i] == projectIDstr {
			return true
		}
	}

	return false
}
