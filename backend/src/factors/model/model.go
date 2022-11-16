package model

import (
	"database/sql"
	"factors/filestore"
	U "factors/util"
	"sync"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"factors/model/model"
)

// Model - Interface of all methods to be implemented by the stores.
type Model interface {
	// adwords_document
	CreateAdwordsDocument(adwordsDoc *model.AdwordsDocument) int
	CreateMultipleAdwordsDocument(adwordsDoc []model.AdwordsDocument) int
	GetAdwordsLastSyncInfoForProject(projectID int64) ([]model.AdwordsLastSyncInfo, int)
	GetAllAdwordsLastSyncInfoForAllProjects() ([]model.AdwordsLastSyncInfo, int)
	PullGCLIDReport(projectID int64, from, to int64, adwordsAccountIDs string, campaignIDReport, adgroupIDReport, keywordIDReport map[string]model.MarketingData, timeZone string) (map[string]model.MarketingData, error)
	GetAdwordsFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int)
	GetAdwordsSQLQueryAndParametersForFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int)
	ExecuteAdwordsChannelQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int)
	GetSQLQueryAndParametersForAdwordsQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string, int)

	// agent
	CreateAgentWithDependencies(params *model.CreateAgentParams) (*model.CreateAgentResponse, int)
	GetAgentByEmail(email string) (*model.Agent, int)
	GetAgentByUUID(uuid string) (*model.Agent, int)
	GetAgentsByUUIDs(uuids []string) ([]*model.Agent, int)
	GetAgentInfo(uuid string) (*model.AgentInfo, int)
	UpdateAgentIntAdwordsRefreshToken(uuid, refreshToken string) int
	UpdateAgentIntGoogleOrganicRefreshToken(uuid, refreshToken string) int
	UpdateAgentIntSalesforce(uuid, refreshToken string, instanceURL string) int
	UpdateAgentPassword(uuid, plainTextPassword string, passUpdatedAt time.Time) int
	UpdateAgentLastLoginInfo(agentUUID string, ts time.Time) int
	UpdateAgentInformation(agentUUID, firstName, lastName, phone string, isOnboardingFlowSeen *bool) int
	UpdateAgentVerificationDetails(agentUUID, password, firstName, lastName string, verified bool, passUpdatedAt time.Time) int
	UpdateAgentVerificationDetailsFromAuth0(agentUUID, firstName, lastName string, verified bool, value *postgres.Jsonb) int
	GetPrimaryAgentOfProject(projectId int64) (uuid string, errCode int)
	UpdateAgentSalesforceInstanceURL(agentUUID string, instanceURL string) int
	IsSlackIntegratedForProject(projectID int64, agentUUID string) (bool, int)
	UpdateLastLoggedOut(agentUUID string, timestamp int64) int

	// analytics
	ExecQuery(stmnt string, params []interface{}) (*model.QueryResult, error, string)
	ExecQueryWithContext(stmnt string, params []interface{}) (*sql.Rows, *sql.Tx, error, string)
	Analyze(projectID int64, query model.Query, enableFilterOpt bool) (*model.QueryResult, int, string)
	IsGroupEventNameByQueryEventWithProperties(projectID int64, ewp model.QueryEventWithProperties) (string, int)

	// archival
	GetNextArchivalBatches(projectID int64, startTime int64, maxLookbackDays int, hardStartTime, hardEndTime time.Time) ([]model.EventsArchivalBatch, error)

	// attribution
	ExecuteAttributionQuery(projectID int64, query *model.AttributionQuery, debugQueryKey string,
		enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) (*model.QueryResult, error)
	ExecuteAttributionQueryV0(projectID int64, query *model.AttributionQuery, debugQueryKey string,
		enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) (*model.QueryResult, error)
	ExecuteAttributionQueryV1(projectID int64, query *model.AttributionQuery, debugQueryKey string,
		enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) (*model.QueryResult, error)
	GetCoalesceIDFromUserIDs(userIDs []string, projectID int64, logCtx log.Entry) (map[string]model.UserInfo, []string, error)
	PullAllUsersByCustomerUserID(projectID int64, kpiData *map[string]model.KPIInfo, logCtx log.Entry) error
	FetchAllUsersAndCustomerUserData(projectID int64, customerUserIdList []string, logCtx log.Entry) (map[string]string, map[string][]string, error)
	FetchAllUsersAndCustomerUserDataInBatches(projectID int64, customerUserIdList []string, logCtx log.Entry) (map[string]string, map[string][]string, error)
	GetConvertedUsers(projectID,
		conversionFrom, conversionTo int64, goalEvent model.QueryEventWithProperties,
		query *model.AttributionQuery, eventNameToIDList map[string][]interface{},
		logCtx log.Entry) (map[string]model.UserInfo, []model.UserEventInfo, map[string]int64, error)
	PullConvertedUsers(projectID int64, query *model.AttributionQuery, conversionFrom int64, conversionTo int64,
		eventNameToIDList map[string][]interface{}, kpiData map[string]model.KPIInfo,
		debugQueryKey string, enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool,
		logCtx *log.Entry) (map[string]int64, []model.UserEventInfo, map[string]model.KPIInfo, []string, error)
	GetAttributionData(projectID int64, query *model.AttributionQuery, sessions map[string]map[string]model.UserSessionData,
		usersToBeAttributed []model.UserEventInfo, coalUserIdConversionTimestamp map[string]int64, marketingReports *model.MarketingReports,
		kpiData map[string]model.KPIInfo, logCtx *log.Entry) (*map[string]*model.AttributionData, bool, error)
	PullSessionsOfConvertedUsers(projectID int64, query *model.AttributionQuery, sessionEventNameID string, usersToBeAttributed []string,
		marketingReports *model.MarketingReports, contentGroupNamesList []string, logCtx *log.Entry) (map[string]map[string]model.UserSessionData, error)
	ExecuteKPIForAttribution(projectID int64, query *model.AttributionQuery, debugQueryKey string,
		logCtx log.Entry, enableOptimisedFilterOnProfileQuery bool,
		enableOptimisedFilterOnEventUserQuery bool) (map[string]model.KPIInfo, error)

	GetLinkedFunnelEventUsersFilter(projectID int64, queryFrom, queryTo int64,
		linkedEvents []model.QueryEventWithProperties, eventNameToId map[string][]interface{},
		userIDInfo map[string]model.UserInfo, logCtx log.Entry) (error, []model.UserEventInfo)
	GetAdwordsCurrency(projectId int64, customerAccountId string, from, to int64, logCtx log.Entry) (string, error)
	GetConvertedUsersWithFilter(projectID int64, goalEventName string,
		goalEventProperties []model.QueryProperty, conversionFrom, conversionTo int64,
		eventNameToIdList map[string][]interface{}, logCtx log.Entry) (map[string]model.UserInfo,
		map[string][]model.UserIDPropID, map[string]int64, error)

	// bigquery_setting
	CreateBigquerySetting(setting *model.BigquerySetting) (*model.BigquerySetting, int)
	UpdateBigquerySettingLastRunAt(settingID string, lastRunAt int64) (int64, int)
	GetBigquerySettingByProjectID(projectID int64) (*model.BigquerySetting, int)

	// billing_account
	GetBillingAccountByProjectID(projectID int64) (*model.BillingAccount, int)
	GetBillingAccountByAgentUUID(AgentUUID string) (*model.BillingAccount, int)
	UpdateBillingAccount(id string, planId uint64, orgName, billingAddr, pinCode, phoneNo string) int
	GetProjectsUnderBillingAccountID(ID string) ([]model.Project, int)
	GetAgentsByProjectIDs(projectIDs []int64) ([]*model.Agent, int)
	GetAgentsUnderBillingAccountID(ID string) ([]*model.Agent, int)
	IsNewProjectAgentMappingCreationAllowed(projectID int64, emailOfAgentToAdd string) (bool, int)

	// channel_analytics
	GetChannelFilterValuesV1(projectID int64, channel, filterObject, filterProperty string, reqID string) (model.ChannelFilterValues, int)
	GetAllChannelFilterValues(projectID int64, filterObject, filterProperty string, source string, reqID string) ([]interface{}, int)
	RunChannelGroupQuery(projectID int64, queries []model.ChannelQueryV1, reqID string) (model.ChannelResultGroupV1, int)
	ExecuteChannelQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string) (*model.ChannelQueryResultV1, int)
	ExecuteSQL(sqlStatement string, params []interface{}, logCtx *log.Entry) ([]string, [][]interface{}, error)
	GetChannelConfig(projectID int64, channel string, reqID string) (*model.ChannelConfigResult, int)

	// KPI Related but in different modules
	GetKPIConfigsForWebsiteSessions(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForPageViews(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForFormSubmissions(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForHubspotContacts(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForHubspotCompanies(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForHubspotDeals(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForSalesforceUsers(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForSalesforceAccounts(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForSalesforceOpportunities(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForAdwords(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForBingAds(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForGoogleOrganic(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForFacebook(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForLinkedin(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForAllChannels(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForMarketoLeads(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForMarketo(projectID int64, reqID string, displayCategory string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetPropertiesForLeadSquared(projectID int64, reqID string) []map[string]string
	GetKPIConfigsForLeadSquaredLeads(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForLeadSquared(projectID int64, reqID string, displayCategory string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForOthers(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)
	GetKPIConfigsForCustomEvents(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int)		

	ExecuteKPIQueryGroup(projectID int64, reqID string, kpiQueryGroup model.KPIQueryGroup,
		enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) ([]model.QueryResult, int)
	ExecuteKPIQueryForEvents(projectID int64, reqID string, kpiQuery model.KPIQuery, enableFilterOpt bool) ([]model.QueryResult, int)
	ExecuteKPIQueryForChannels(projectID int64, reqID string, kpiQuery model.KPIQuery) ([]model.QueryResult, int)

	// Custom Metrics
	CreateCustomMetric(customMetric model.CustomMetric) (*model.CustomMetric, string, int)
	GetCustomMetricsByProjectId(projectID int64) ([]model.CustomMetric, string, int)
	GetCustomMetricByProjectIdQueryTypeAndObjectType(projectID int64, queryType int, objectType string) ([]model.CustomMetric, string, int)
	GetCustomKPIMetricsByProjectIdAndDisplayCategory(projectID int64, displayCategory string) []map[string]string
	GetKpiRelatedCustomMetricsByName(projectID int64, name string) (model.CustomMetric, string, int)
	GetProfileCustomMetricByProjectIdName(projectID int64, name string) (model.CustomMetric, string, int)
	GetDerivedCustomMetricByProjectIdName(projectID int64, name string) (model.CustomMetric, string, int)
	GetEventBasedCustomMetricByProjectIdName(projectID int64, name string) (model.CustomMetric, string, int)
	GetCustomMetricsByID(projectID int64, id string) (model.CustomMetric, string, int)
	DeleteCustomMetricByID(projectID int64, id string) int
	GetDerivedKPIsHavingNameInInternalQueries(projectID int64, customMetricName string) []string
	GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID int64, displayCategory string, includeDerivedKPIs bool) []map[string]string
	GetCustomMetricAndDerivedMetricByProjectIdAndDisplayCategory(projectID int64, displayCategory string, includeDerivedKPIs bool) []map[string]string
	GetCustomEventKPIMetricsByProjectIdAndDisplayCategory(projectID int64, displayCategory string, includeDerivedKPIs bool) []map[string]string

	//templates
	RunTemplateQuery(projectID int64, query model.TemplateQuery, reqID string) (model.TemplateResponse, int)
	GetTemplateConfig(projectID int64, templateType int) (model.TemplateConfig, int)
	UpdateTemplateConfig(projectID int64, templateType int, thresholds []model.TemplateThreshold) ([]model.TemplateThreshold, string)

	// dashboard_unit
	CreateDashboardUnit(projectID int64, agentUUID string, dashboardUnit *model.DashboardUnit) (*model.DashboardUnit, int, string)
	CreateDashboardUnitForDashboardClass(projectID int64, agentUUID string, dashboardUnit *model.DashboardUnit, dashboardClass string) (*model.DashboardUnit, int, string)
	CreateDashboardUnitForMultipleDashboards(dashboardIds []int64, projectId int64, agentUUID string, unitPayload model.DashboardUnitRequestPayload) ([]*model.DashboardUnit, int, string)
	CreateMultipleDashboardUnits(requestPayload []model.DashboardUnitRequestPayload, projectId int64, agentUUID string, dashboardId int64) ([]*model.DashboardUnit, int, string)
	GetDashboardUnitsForProjectID(projectID int64) ([]model.DashboardUnit, int)
	GetDashboardUnits(projectID int64, agentUUID string, dashboardId int64) ([]model.DashboardUnit, int)
	GetDashboardUnitByUnitID(projectID int64, unitID int64) (*model.DashboardUnit, int)
	GetDashboardUnitsByProjectIDAndDashboardIDAndTypes(projectID int64, dashboardID int64, types []string) ([]model.DashboardUnit, int)
	DeleteDashboardUnit(projectID int64, agentUUID string, dashboardId int64, id int64) int
	DeleteMultipleDashboardUnits(projectID int64, agentUUID string, dashboardID int64, dashboardUnitIDs []int64) (int, string)
	UpdateDashboardUnit(projectId int64, agentUUID string, dashboardId int64, id int64, unit *model.DashboardUnit) (*model.DashboardUnit, int)
	CacheDashboardUnitsForProjects(stringProjectsIDs, excludeProjectIDs string, numRoutines int, reportCollector *sync.Map, enableFilterOpt bool)
	CacheDashboardUnitsForProjectID(projectID int64, dashboardUnits []model.DashboardUnit, queryClasses []string, numRoutines int, reportCollector *sync.Map, enableFilterOpt bool) int
	CacheDashboardUnit(dashboardUnit model.DashboardUnit, waitGroup *sync.WaitGroup, reportCollector *sync.Map, queryClass string, enableFilterOpt bool)
	GetQueryAndClassFromDashboardUnit(dashboardUnit *model.DashboardUnit) (queryClass string, queryInfo *model.Queries, errMsg string)
	GetQueryClassFromQueries(query model.Queries) (queryClass, errMsg string)
	GetQueryAndClassFromQueryIdString(queryIdString string, projectId int64) (queryClass string, queryInfo *model.Queries, errMsg string)
	GetQueryWithQueryIdString(projectID int64, queryIDString string) (*model.Queries, int)
	CacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload, enableFilterOpt bool) (int, string, model.CachingUnitReport)
	CacheDashboardsForMonthlyRange(projectIDs, excludeProjectIDs string, numMonths, numRoutines int, reportCollector *sync.Map, enableFilterOpt bool)
	GetDashboardUnitNamesByProjectIdTypeAndName(projectID int64, reqID string, typeOfQuery string, nameOfQuery string) ([]string, int)

	// dashboard
	CreateDashboard(projectID int64, agentUUID string, dashboard *model.Dashboard) (*model.Dashboard, int)
	CreateAgentPersonalDashboardForProject(projectID int64, agentUUID string) (*model.Dashboard, int)
	GetDashboards(projectID int64, agentUUID string) ([]model.Dashboard, int)
	GetDashboard(projectID int64, agentUUID string, id int64) (*model.Dashboard, int)
	HasAccessToDashboard(projectID int64, agentUUID string, id int64) (bool, *model.Dashboard)
	UpdateDashboard(projectID int64, agentUUID string, id int64, dashboard *model.UpdatableDashboard) int
	DeleteDashboard(projectID int64, agentUUID string, dashboardID int64) int

	// event_analytics
	RunEventsGroupQuery(queriesOriginal []model.Query, projectId int64, enableFilterOpt bool) (model.ResultGroup, int)
	ExecuteEventsQuery(projectID int64, query model.Query, enableFilterOpt bool) (*model.QueryResult, int, string)
	RunInsightsQuery(projectID int64, query model.Query, enableFilterOpt bool) (*model.QueryResult, int, string)
	BuildInsightsQuery(projectID int64, query model.Query, enableFilterOpt bool) (string, []interface{}, error)

	// Profile
	RunProfilesGroupQuery(queriesOriginal []model.ProfileQuery, projectID int64, enableOptimisedFilter bool) (model.ResultGroup, int)
	ExecuteProfilesQuery(projectID int64, query model.ProfileQuery, enableOptimisedFilter bool) (*model.QueryResult, int, string)

	// event_name
	CreateOrGetEventName(eventName *model.EventName) (*model.EventName, int)
	CreateOrGetUserCreatedEventName(eventName *model.EventName) (*model.EventName, int)
	CreateOrGetAutoTrackedEventName(eventName *model.EventName) (*model.EventName, int)
	CreateOrGetFilterEventName(eventName *model.EventName) (*model.EventName, int)
	CreateOrGetCRMSmartEventFilterEventName(projectID int64, eventName *model.EventName, filterExpr *model.SmartCRMEventFilter) (*model.EventName, int)
	GetSmartEventEventName(eventName *model.EventName) (*model.EventName, int)
	GetSmartEventEventNameByNameANDType(projectID int64, name, typ string) (*model.EventName, int)
	CreateOrGetSessionEventName(projectID int64) (*model.EventName, int)
	CreateOrGetOfflineTouchPointEventName(projectID int64) (*model.EventName, int)
	GetSessionEventName(projectID int64) (*model.EventName, int)
	GetEventName(name string, projectID int64) (*model.EventName, int)
	GetEventNames(projectID int64) ([]model.EventName, int)
	GetOrderedEventNamesFromDb(projectID int64, startTimestamp int64, endTimestamp int64, limit int) ([]model.EventNameWithAggregation, error)
	GetFilterEventNames(projectID int64) ([]model.EventName, int)
	GetSmartEventFilterEventNames(projectID int64, includeDeleted bool) ([]model.EventName, int)
	GetSmartEventFilterEventNameByID(projectID int64, id string, isDeleted bool) (*model.EventName, int)
	GetEventNamesByNames(projectID int64, names []string) ([]model.EventName, int)
	DeleteSmartEventFilter(projectID int64, id string) (*model.EventName, int)
	GetFilterEventNamesByExprPrefix(projectID int64, prefix string) ([]model.EventName, int)
	UpdateEventName(projectId int64, id string, nameType string, eventName *model.EventName) (*model.EventName, int)
	UpdateCRMSmartEventFilter(projectID int64, id string, eventName *model.EventName, filterExpr *model.SmartCRMEventFilter) (*model.EventName, int)
	UpdateFilterEventName(projectID int64, id string, eventName *model.EventName) (*model.EventName, int)
	DeleteFilterEventName(projectID int64, id string) int
	FilterEventNameByEventURL(projectID int64, eventURL string) (*model.EventName, int)
	GetEventNameFromEventNameId(eventNameId string, projectID int64) (*model.EventName, error)
	GetEventNameIDFromEventName(eventName string, projectId int64) (*model.EventName, error)
	GetEventTypeFromDb(projectID int64, eventNames []string, limit int64) (map[string]string, error)
	GetMostFrequentlyEventNamesByType(projectID int64, limit int, lastNDays int, typeOfEvent string) ([]string, error)
	GetEventNamesOrderedByOccurenceAndRecency(projectID int64, limit int, lastNDays int) (map[string][]string, error)
	GetPropertiesByEvent(projectID int64, eventName string, limit int, lastNDays int) (map[string][]string, error)
	GetPropertyValuesByEventProperty(projectID int64, eventName string, propertyName string, limit int, lastNDays int) ([]string, error)
	GetPropertiesForHubspotContacts(projectID int64, reqID string) []map[string]string
	GetPropertiesForHubspotCompanies(projectID int64, reqID string) []map[string]string
	GetPropertiesForHubspotDeals(projectID int64, reqID string) []map[string]string
	GetPropertiesForSalesforceAccounts(projectID int64, reqID string) []map[string]string
	GetPropertiesForSalesforceOpportunities(projectID int64, reqID string) []map[string]string
	GetPropertiesForSalesforceUsers(projectID int64, reqID string) []map[string]string
	GetPropertiesForMarketo(projectID int64, reqID string) []map[string]string

	// form_fill
	CreateFormFillEventById(projectId int64, formFill *model.SDKFormFillPayload) (int, error)
	GetFormFillEventById(projectId int64, userId string, formId string, fieldId string) (*model.FormFill, int)
	GetFormFillEventsUpdatedBeforeTenMinutes(projectIds []int64) ([]model.FormFill, error)
	DeleteFormFillProcessedRecords(projectId int64, userId string, formId string, fieldId string) (int, error)

	// events
	GetEventCountOfUserByEventName(projectID int64, userId string, eventNameId string) (uint64, int)
	GetEventCountOfUsersByEventName(projectID int64, userIDs []string, eventNameID string) (uint64, int)
	CreateEvent(event *model.Event) (*model.Event, int)
	GetEvent(projectID int64, userId string, id string) (*model.Event, int)
	GetEventById(projectID int64, id, userID string) (*model.Event, int)
	GetLatestEventOfUserByEventNameId(projectID int64, userId string, eventNameId string, startTimestamp int64, endTimestamp int64) (*model.Event, int)
	GetEventsByEventNameId(projectID int64, eventNameId string, startTimestamp int64, endTimestamp int64) ([]model.Event, int)
	GetRecentEventPropertyKeysWithLimits(projectID int64, eventName string, starttime int64, endtime int64, eventsLimit int) ([]U.Property, error)
	GetRecentEventPropertyValuesWithLimits(projectID int64, eventName string, property string, valuesLimit int, rowsLimit int, starttime int64, endtime int64) ([]U.PropertyValue, string, error)
	UpdateEventProperties(projectID int64, id, userID string, properties *U.PropertiesMap, updateTimestamp int64, optionalEventUserProperties *postgres.Jsonb) int
	GetUserEventsByEventNameId(projectID int64, userId string, eventNameId string) ([]model.Event, int)
	OverwriteEventProperties(projectID int64, userId string, eventId string, newEventProperties *postgres.Jsonb) int
	OverwriteEventPropertiesByID(projectID int64, id string, newEventProperties *postgres.Jsonb) int
	AddSessionForUser(projectID int64, userId string, userEvents []model.Event, bufferTimeBeforeSessionCreateInSecs int64, sessionEventNameId string) (int, int, bool, int, int)
	GetDatesForNextEventsArchivalBatch(projectID int64, startTime int64) (map[string]int64, int)
	GetAllEventsForSessionCreationAsUserEventsMap(projectID int64, sessionEventNameId string, startTimestamp, endTimestamp int64) (*map[string][]model.Event, int, int)
	GetEventsWithoutPropertiesAndWithPropertiesByNameForYourStory(projectID int64, from, to int64, mandatoryProperties []string) ([]model.EventWithProperties, *map[string]U.PropertiesMap, int)
	OverwriteEventUserPropertiesByID(projectID int64, userID, id string, properties *postgres.Jsonb) int
	PullEventRows(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error)
	PullAdwordsRows(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error)
	PullFacebookRows(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error)
	PullBingAdsRows(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error)
	PullLinkedInRows(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error)
	PullGoogleOrganicRows(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error)
	PullUsersRowsForWI(projectID int64, startTime, endTime int64, dateField string, source int, group int) (*sql.Rows, *sql.Tx, error)
	PullEventRowsForArchivalJob(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error)
	GetUnusedSessionIDsForJob(projectID int64, startTimestamp, endTimestamp int64) ([]string, int)
	DeleteEventsByIDsInBatchForJob(projectID int64, eventNameID string, ids []string, batchSize int) int
	DeleteEventByIDs(projectID int64, eventNameID string, ids []string) int
	AssociateSessionByEventIds(projectID int64, userID string, events []*model.Event, sessionId string, sessionEventNameId string) int
	GetHubspotFormEvents(projectID int64, userId string, timestamps []interface{}) ([]model.Event, int)
	IsSmartEventAlreadyExist(projectID int64, userID, eventNameID, referenceEventID string, eventTimestamp int64) (bool, error)

	// clickable_elements
	UpsertCountAndCheckEnabledClickableElement(projectID int64, payload *model.CaptureClickPayload) (isEnabled bool, status int, err error)
	CreateClickableElement(projectId int64, payload *model.CaptureClickPayload) (int, error)
	GetClickableElement(projectID int64, displayName string, elementType string) (*model.ClickableElements, int)
	ToggleEnabledClickableElement(projectId int64, id string) int
	GetAllClickableElements(projectId int64) ([]model.ClickableElements, int)

	// facebook_document
	CreateFacebookDocument(projectID int64, document *model.FacebookDocument) int
	GetFacebookSQLQueryAndParametersForFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int)
	ExecuteFacebookChannelQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int)
	GetFacebookLastSyncInfo(projectID int64, CustomerAdAccountID string) ([]model.FacebookLastSyncInfo, int)
	GetSQLQueryAndParametersForFacebookQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string, int)

	// linkedin document
	CreateLinkedinDocument(projectID int64, document *model.LinkedinDocument) int
	GetLinkedinSQLQueryAndParametersForFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int)
	ExecuteLinkedinChannelQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int)
	GetLinkedinLastSyncInfo(projectID int64, CustomerAdAccountID string) ([]model.LinkedinLastSyncInfo, int)
	GetSQLQueryAndParametersForLinkedinQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string, fetchSource bool,
		limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string, int)

	//bingads document
	GetBingadsFilterValuesSQLAndParams(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int)

	// funnel_analytics
	RunFunnelQuery(projectID int64, query model.Query, enableFilterOpt bool) (*model.QueryResult, int, string)

	// goals
	GetAllFactorsGoals(ProjectID int64) ([]model.FactorsGoal, int)
	GetAllActiveFactorsGoals(ProjectID int64) ([]model.FactorsGoal, int)
	CreateFactorsGoal(ProjectID int64, Name string, Rule model.FactorsGoalRule, agentUUID string) (int64, int, string)
	DeactivateFactorsGoal(ID int64, ProjectID int64) (int64, int)
	ActivateFactorsGoal(Name string, ProjectID int64) (int64, int)
	UpdateFactorsGoal(ID int64, Name string, Rule model.FactorsGoalRule, ProjectID int64) (int64, int)
	GetFactorsGoal(Name string, ProjectID int64) (*model.FactorsGoal, error)
	GetFactorsGoalByID(ID int64, ProjectID int64) (*model.FactorsGoal, error)
	GetAllFactorsGoalsWithNamePattern(ProjectID int64, NamePattern string) ([]model.FactorsGoal, int)

	// hubspot_document
	GetHubspotDocumentByTypeAndActions(projectId int64, ids []string, docType int, actions []int) ([]model.HubspotDocument, int)
	GetSyncedHubspotDocumentByFilter(projectId int64, id string, docType, action int) (*model.HubspotDocument, int)
	CreateHubspotDocument(projectID int64, document *model.HubspotDocument) int
	CreateHubspotDocumentInBatch(projectID int64, docType int, documents []*model.HubspotDocument, batchSize int) int
	GetHubspotSyncInfo() (*model.HubspotSyncInfo, int)
	GetHubspotFirstSyncProjectsInfo() (*model.HubspotSyncInfo, int)
	UpdateHubspotProjectSettingsBySyncStatus(success []model.HubspotProjectSyncStatus, failure []model.HubspotProjectSyncStatus, syncAll bool) int
	GetHubspotDocumentBeginingTimestampByDocumentTypeForSync(projectID int64, docTypes []int) (int64, int)
	GetMinTimestampByFirstSync(projectID int64, docType int) (int64, int)
	GetHubspotFormDocuments(projectID int64) ([]model.HubspotDocument, int)
	GetHubspotDocumentsByTypeForSync(projectID int64, typ int, maxCreatedAtSec int64) ([]model.HubspotDocument, int)
	GetHubspotContactCreatedSyncIDAndUserID(projectID int64, docID string) ([]model.HubspotDocument, int)
	GetHubspotDocumentsByTypeANDRangeForSync(projectID int64, docType int, from, to, maxCreatedAtSec int64) ([]model.HubspotDocument, int)
	GetSyncedHubspotDealDocumentByIdAndStage(projectId int64, id string, stage string) (*model.HubspotDocument, int)
	GetHubspotObjectPropertiesName(ProjectID int64, objectType string) ([]string, []string)
	UpdateHubspotDocumentAsSynced(projectID int64, id string, docType int, syncId string, timestamp int64, action int, userID, groupUserID string) int
	GetLastSyncedHubspotDocumentByID(projectID int64, docID string, docType int) (*model.HubspotDocument, int)
	GetLastSyncedHubspotUpdateDocumentByID(projectID int64, docID string, docType int) (*model.HubspotDocument, int)
	GetAllHubspotObjectValuesByPropertyName(ProjectID int64, objectType, propertyName string) []interface{}
	GetHubspotDocumentCountForSync(projectIDs []int64, docTypes []int, maxCreatedAtSec int64) ([]model.HubspotDocumentCount, int)

	// plan
	GetPlanByID(planID uint64) (*model.Plan, int)
	GetPlanByCode(Code string) (*model.Plan, int)

	// project_agent_mapping
	CreateProjectAgentMappingWithDependencies(pam *model.ProjectAgentMapping) (*model.ProjectAgentMapping, int)
	CreateProjectAgentMappingWithDependenciesWithoutDashboard(pam *model.ProjectAgentMapping) (*model.ProjectAgentMapping, int)
	GetProjectAgentMapping(projectID int64, agentUUID string) (*model.ProjectAgentMapping, int)
	GetProjectAgentMappingsByProjectId(projectID int64) ([]model.ProjectAgentMapping, int)
	GetProjectAgentMappingsByProjectIds(projectIds []int64) ([]model.ProjectAgentMapping, int)
	GetProjectAgentMappingsByAgentUUID(agentUUID string) ([]model.ProjectAgentMapping, int)
	DoesAgentHaveProject(agentUUID string) int
	DeleteProjectAgentMapping(projectID int64, agentUUIDToRemove string) int
	EditProjectAgentMapping(projectID int64, agentUUIDToEdit string, role int64) int

	// project_billing_account
	GetProjectBillingAccountMappings(billingAccountID string) ([]model.ProjectBillingAccountMapping, int)
	GetProjectBillingAccountMapping(projectID int64) (*model.ProjectBillingAccountMapping, int)

	// project_setting
	GetProjectSetting(projectID int64) (*model.ProjectSetting, int)
	GetClearbitKeyFromProjectSetting(projectId int64) (string, int)
	GetClient6SignalKeyFromProjectSetting(projectId int64) (string, int)
	GetFactors6SignalKeyFromProjectSetting(projectId int64) (string, int)
	GetProjectSettingByKeyWithTimeout(key, value string, timeout time.Duration) (*model.ProjectSetting, int)
	GetProjectSettingByTokenWithCacheAndDefault(token string) (*model.ProjectSetting, int)
	GetProjectSettingByPrivateTokenWithCacheAndDefault(privateToken string) (*model.ProjectSetting, int)
	GetProjectIDByToken(token string) (int64, int)
	UpdateProjectSettings(projectID int64, settings *model.ProjectSetting) (*model.ProjectSetting, int)
	GetIntAdwordsRefreshTokenForProject(projectID int64) (string, int)
	GetIntGoogleOrganicRefreshTokenForProject(projectID int64) (string, int)
	GetIntAdwordsProjectSettingsForProjectID(projectID int64) ([]model.AdwordsProjectSettings, int)
	GetAllIntAdwordsProjectSettings() ([]model.AdwordsProjectSettings, int)
	GetAllHubspotProjectSettings() ([]model.HubspotProjectSettings, int)
	GetFacebookEnabledIDsAndProjectSettings() ([]int64, []model.FacebookProjectSettings, int)
	GetFacebookEnabledIDsAndProjectSettingsForProject(projectIDs []int64) ([]int64, []model.FacebookProjectSettings, int)
	GetLinkedinEnabledProjectSettings() ([]model.LinkedinProjectSettings, int)
	GetLinkedinEnabledProjectSettingsForProjects(projectIDs []string) ([]model.LinkedinProjectSettings, int)
	GetArchiveEnabledProjectIDs() ([]int64, int)
	GetBigqueryEnabledProjectIDs() ([]int64, int)
	GetAllSalesforceProjectSettings() ([]model.SalesforceProjectSettings, int)
	IsPSettingsIntShopifyEnabled(projectId int64) bool
	GetProjectDetailsByShopifyDomain(shopifyDomain string) (int64, string, bool, int)
	EnableBigqueryArchivalForProject(projectID int64) int
	IsBingIntegrationAvailable(projectID int64) bool
	GetAllLeadSquaredEnabledProjects() (map[int64]model.LeadSquaredConfig, error)
	UpdateLeadSquaredFirstTimeSyncStatus(projectId int64) int
	UpdateLeadSquaredConfig(projectId int64, accessKey string, host string, secretkey string) int
	EnableWeeklyInsights(projectId int64) int
	EnableExplain(projectId int64) int
	DisableWeeklyInsights(projectId int64) int
	DisableExplain(projectId int64) int
	GetAllWeeklyInsightsEnabledProjects() ([]int64, error)
	GetAllExplainEnabledProjects() ([]int64, error)
	GetFormFillEnabledProjectIDs() ([]int64, error)

	// project
	UpdateProject(projectID int64, project *model.Project) int
	CreateProjectWithDependencies(project *model.Project, agentUUID string, agentRole uint64, billingAccountID string, createDashboard bool) (*model.Project, int)
	CreateDefaultProjectForAgent(agentUUID string) (*model.Project, int)
	GetProject(id int64) (*model.Project, int)
	GetProjectByToken(token string) (*model.Project, int)
	GetProjectByPrivateToken(privateToken string) (*model.Project, int)
	GetProjects() ([]model.Project, int)
	GetProjectsByIDs(ids []int64) ([]model.Project, int)
	GetAllProjectIDs() ([]int64, int)
	GetNextSessionStartTimestampForProject(projectID int64) (int64, int)
	UpdateNextSessionStartTimestampForProject(projectID int64, timestamp int64) int
	GetProjectsToRunForIncludeExcludeString(projectIDs, excludeProjectIDs string) []int64
	GetProjectsWithoutWebAnalyticsDashboard(onlyProjectsMap map[int64]bool) (projectIds []int64, errCode int)
	GetTimezoneForProject(projectID int64) (U.TimeZoneString, int)

	// queries
	CreateQuery(projectID int64, query *model.Queries) (*model.Queries, int, string)
	GetALLQueriesWithProjectId(projectID int64) ([]model.Queries, int)
	GetDashboardQueryWithQueryId(projectID int64, queryID int64) (*model.Queries, int)
	GetSavedQueryWithQueryId(projectID int64, queryID int64) (*model.Queries, int)
	GetQueryWithQueryId(projectID int64, queryID int64) (*model.Queries, int)
	DeleteQuery(projectID int64, queryID int64) (int, string)
	DeleteSavedQuery(projectID int64, queryID int64) (int, string)
	DeleteDashboardQuery(projectID int64, queryID int64) (int, string)
	UpdateSavedQuery(projectID int64, queryID int64, query *model.Queries) (*model.Queries, int)
	UpdateQueryIDsWithNewIDs(projectID int64, shareableURLs []string) int
	SearchQueriesWithProjectId(projectID int64, searchString string) ([]model.Queries, int)
	GetAllNonConvertedQueries(projectID int64) ([]model.Queries, int)

	// dashboard_templates
	CreateTemplate(template *model.DashboardTemplate) (*model.DashboardTemplate, int, string)
	DeleteTemplate(templateId string) int
	SearchTemplateWithTemplateID(templateId string) (model.DashboardTemplate, int)
	GetAllTemplates() ([]model.DashboardTemplate, int)

	// offline touchpoints
	CreateOTPRule(projectId int64, rule *model.OTPRule) (*model.OTPRule, int, string)
	GetALLOTPRuleWithProjectId(projectID int64) ([]model.OTPRule, int)
	GetAllRulesDeletedNotDeleted(projectID int64) ([]model.OTPRule, int)
	GetOTPRuleWithRuleId(projectID int64, ruleID string) (*model.OTPRule, int)
	GetAnyOTPRuleWithRuleId(projectID int64, ruleID string) (*model.OTPRule, int)
	DeleteOTPRule(projectID int64, ruleID string) (int, string)
	UpdateOTPRule(projectID int64, ruleID string, rule *model.OTPRule) (*model.OTPRule, int)

	// salesforce_document
	GetSalesforceSyncInfo() (model.SalesforceSyncInfo, int)
	GetSalesforceObjectPropertiesName(ProjectID int64, objectType string) ([]string, []string)
	GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(projectID int64, customerUserID, userID string, docType int) (*model.SalesforceDocument, int)
	UpdateSalesforceDocumentBySyncStatus(projectID int64, document *model.SalesforceDocument, syncID, userID, groupUserID string, synced bool) int
	BuildAndUpsertDocument(projectID int64, objectName string, value model.SalesforceRecord) error
	BuildAndUpsertDocumentInBatch(projectID int64, objectName string, values []model.SalesforceRecord) error
	CreateSalesforceDocument(projectID int64, document *model.SalesforceDocument) int
	CreateSalesforceDocumentByAction(projectID int64, document *model.SalesforceDocument, action model.SalesforceAction) int
	GetSalesforceDocumentsByTypeAndAction(projectID int64, docType int, action model.SalesforceAction, from int64, to int64) ([]model.SalesforceDocument, int)
	GetSyncedSalesforceDocumentByType(projectID int64, ids []string, docType int, includeUnSynced bool) ([]model.SalesforceDocument, int)
	GetSalesforceObjectValuesByPropertyName(ProjectID int64, objectType string, propertyName string) []interface{}
	GetSalesforceDocumentsByTypeForSync(projectID int64, typ int, from, to int64) ([]model.SalesforceDocument, int)
	GetLatestSalesforceDocumentByID(projectID int64, documentIDs []string, docType int, maxTimestamp int64) ([]model.SalesforceDocument, int)
	GetSalesforceDocumentBeginingTimestampByDocumentTypeForSync(projectID int64) (map[int]int64, int64, int)
	GetSalesforceDocumentByType(projectID int64, docType int, from, to int64) ([]model.SalesforceDocument, int)

	// scheduled_task
	CreateScheduledTask(task *model.ScheduledTask) int
	UpdateScheduledTask(taskID string, taskDetails *postgres.Jsonb, endTime int64, status model.ScheduledTaskStatus) (int64, int)
	GetScheduledTaskByID(taskID string) (*model.ScheduledTask, int)
	GetScheduledTaskInProgressCount(projectID int64, taskType model.ScheduledTaskType) (int64, int)
	GetScheduledTaskLastRunTimestamp(projectID int64, taskType model.ScheduledTaskType) (int64, int)
	GetNewArchivalFileNamesAndEndTimeForProject(projectID int64,
		lastRunAt int64, hardStartTime, hardEndTime time.Time) (map[int64]map[string]interface{}, int)
	GetArchivalFileNamesForProject(projectID int64, startTime, endTime time.Time) ([]string, []string, int)
	FailScheduleTask(taskID string)
	GetCompletedArchivalBatches(projectID int64, startTime, endTime time.Time) (map[int64]int64, int)

	// tracked_events
	CreateFactorsTrackedEvent(ProjectID int64, EventName string, agentUUID string) (int64, int)
	DeactivateFactorsTrackedEvent(ID int64, ProjectID int64) (int64, int)
	GetAllFactorsTrackedEventsByProject(ProjectID int64) ([]model.FactorsTrackedEventInfo, int)
	GetAllActiveFactorsTrackedEventsByProject(ProjectID int64) ([]model.FactorsTrackedEventInfo, int)
	GetFactorsTrackedEvent(EventNameID string, ProjectID int64) (*model.FactorsTrackedEvent, error)
	GetFactorsTrackedEventByID(ID int64, ProjectID int64) (*model.FactorsTrackedEvent, error)

	// tracked_user_properties
	CreateFactorsTrackedUserProperty(ProjectID int64, UserPropertyName string, agentUUID string) (int64, int)
	RemoveFactorsTrackedUserProperty(ID int64, ProjectID int64) (int64, int)
	GetAllFactorsTrackedUserPropertiesByProject(ProjectID int64) ([]model.FactorsTrackedUserProperty, int)
	GetAllActiveFactorsTrackedUserPropertiesByProject(ProjectID int64) ([]model.FactorsTrackedUserProperty, int)
	GetFactorsTrackedUserProperty(UserPropertyName string, ProjectID int64) (*model.FactorsTrackedUserProperty, error)
	GetFactorsTrackedUserPropertyByID(ID int64, ProjectID int64) (*model.FactorsTrackedUserProperty, error)
	IsUserPropertyValid(ProjectID int64, UserProperty string) bool

	// user
	CreateUser(user *model.User) (string, int)
	CreateOrGetAMPUser(projectID int64, ampUserId string, timestamp int64, requestSource int) (string, int)
	CreateOrGetSegmentUser(projectID int64, segAnonId, custUserId string, requestTimestamp int64, requestSource int) (*model.User, int)
	GetUserPropertiesByUserID(projectID int64, id string) (*postgres.Jsonb, int)
	GetUser(projectID int64, id string) (*model.User, int)
	GetUserIDByAMPUserID(projectId int64, ampUserId string) (string, int)
	IsUserExistByID(projectID int64, id string) int
	GetUsers(projectID int64, offset uint64, limit uint64) ([]model.User, int)
	GetUsersByCustomerUserID(projectID int64, customerUserID string) ([]model.User, int)
	GetUserLatestByCustomerUserId(projectID int64, customerUserId string, requestSource int) (*model.User, int)
	GetExistingUserByCustomerUserID(projectID int64, arrayCustomerUserID []string, source ...int) (map[string]string, int)
	GetUserWithoutProperties(projectID int64, id string) (*model.User, int)
	GetUserBySegmentAnonymousId(projectID int64, segAnonId string) (*model.User, int)
	GetAllUserIDByCustomerUserID(projectID int64, customerUserID string) ([]string, int)
	GetRecentUserPropertyKeysWithLimits(projectID int64, usersLimit int, propertyLimit int, seedDate time.Time) ([]U.Property, error)
	GetRecentUserPropertyValuesWithLimits(projectID int64, propertyKey string, usersLimit, valuesLimit int, seedDate time.Time) ([]U.PropertyValue, string, error)
	GetUserPropertiesByProject(projectID int64, limit int, lastNDays int) (map[string][]string, error)
	GetPropertyValuesByUserProperty(projectID int64, propertyName string, limit int, lastNDays int) ([]string, error)
	GetLatestUserPropertiesOfUserAsMap(projectID int64, id string) (*map[string]interface{}, int)
	GetDistinctCustomerUserIDSForProject(projectID int64) ([]string, int)
	GetUserIdentificationPhoneNumber(projectID int64, phoneNo string) (string, string)
	UpdateUser(projectID int64, id string, user *model.User, updateTimestamp int64) (*model.User, int)
	UpdateUserProperties(projectId int64, id string, properties *postgres.Jsonb, updateTimestamp int64) (*postgres.Jsonb, int)
	UpdateUserPropertiesV2(projectID int64, id string, newProperties *postgres.Jsonb, newUpdateTimestamp int64, sourceValue string, objectType string) (*postgres.Jsonb, int)
	OverwriteUserPropertiesByID(projectID int64, id string, properties *postgres.Jsonb, withUpdateTimestamp bool, updateTimestamp int64, source string) int
	OverwriteUserPropertiesByCustomerUserID(projectID int64, customerUserID string, properties *postgres.Jsonb, updateTimestamp int64) int
	GetUserByPropertyKey(projectID int64, key string, value interface{}) (*model.User, int)
	UpdateUserPropertiesForSession(projectID int64, sessionUserPropertiesRecordMap *map[string]model.SessionUserProperties) int
	GetCustomerUserIDAndUserPropertiesFromFormSubmit(projectID int64, userID string, formSubmitProperties *U.PropertiesMap) (string, *U.PropertiesMap, int)
	UpdateIdentifyOverwriteUserPropertiesMeta(projectID int64, customerUserID, userID, pageURL, source string, userProperties *postgres.Jsonb, timestamp int64, isNewUser bool) error
	GetSelectedUsersByCustomerUserID(projectID int64, customerUserID string, limit uint64, numUsers uint64) ([]model.User, int)
	CreateGroupUser(user *model.User, groupName, groupID string) (string, int)
	UpdateUserGroup(projectID int64, userID, groupName, groupID, groupUserID string) (*model.User, int)
	UpdateUserGroupProperties(projectID int64, userID string, newProperties *postgres.Jsonb, updateTimestamp int64) (*postgres.Jsonb, int)
	GetPropertiesUpdatedTimestampOfUser(projectId int64, id string) (int64, int)

	// web_analytics
	GetWebAnalyticsQueriesFromDashboardUnits(projectID int64) (int64, *model.WebAnalyticsQueries, int)
	CreateWebAnalyticsDefaultDashboardWithUnits(projectID int64, agentUUID string) int
	ExecuteWebAnalyticsQueries(projectID int64, queries *model.WebAnalyticsQueries) (queryResult *model.WebAnalyticsQueryResult, errCode int)
	CacheWebsiteAnalyticsForProjects(stringProjectsIDs, excludeProjectIDs string, numRoutines int, reportCollector *sync.Map)
	GetWebAnalyticsEnabledProjectIDsFromList(stringProjectIDs, excludeProjectIDs string) []int64
	GetWebAnalyticsCachePayloadsForProject(projectID int64) ([]model.WebAnalyticsCachePayload, int, string)
	CacheWebsiteAnalyticsForDateRange(cachePayload model.WebAnalyticsCachePayload) (int, model.CachingUnitReport)
	CacheWebsiteAnalyticsForMonthlyRange(projectIDs, excludeProjectIDs string, numMonths, numRoutines int, reportCollector *sync.Map)

	// journey_mining
	GetWeightedJourneyMatrix(projectID int64, journeyEvents []model.QueryEventWithProperties,
		goalEvents []model.QueryEventWithProperties, startTime, endTime, lookbackDays int64, eventFiles,
		userFiles string, includeSession bool, sessionProperty string, cloudManager filestore.FileManager)

	// smart_properties
	GetSmartPropertyRulesConfig(projectID int64, objectType string) (model.SmartPropertyRulesConfig, int)
	CreateSmartPropertyRules(projectID int64, smartProperty *model.SmartPropertyRules) (*model.SmartPropertyRules, string, int)
	GetSmartPropertyRules(projectID int64) ([]model.SmartPropertyRules, int)
	GetAllChangedSmartPropertyRulesForProject(projectID int64) ([]model.SmartPropertyRules, int)
	GetSmartPropertyRule(projectID int64, ruleID string) (model.SmartPropertyRules, int)
	DeleteSmartPropertyRules(projectID int64, ruleID string) int
	UpdateSmartPropertyRules(projectID int64, ruleID string, smartPropertyDoc model.SmartPropertyRules) (model.SmartPropertyRules, string, int)
	GetProjectIDsHavingSmartPropertyRules() ([]int64, int)
	GetLatestMetaForAdwordsForGivenDays(projectID int64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields)
	GetLatestMetaForFacebookForGivenDays(projectID int64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields)
	GetLatestMetaForLinkedinForGivenDays(projectID int64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields)
	GetLatestMetaForBingAdsForGivenDays(projectID int64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields)
	GetLatestMetaForCustomAdsForGivenDays(projectID int64, source string, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields)
	BuildAndCreateSmartPropertyFromChannelDocumentAndRule(smartPropertyRule *model.SmartPropertyRules, rule model.Rule,
		channelDocument model.ChannelDocumentsWithFields, source string) int
	DeleteSmartPropertyByRuleID(projectID int64, ruleID string) (int, int, int)
	GetSmartPropertyByProjectIDAndObjectIDAndObjectType(projectID int64, objectID string, objectType int) (model.SmartProperties, int)
	GetSmartPropertyByProjectIDAndSourceAndObjectType(projectID int64, source string, objectType int) ([]model.SmartProperties, int)
	DeleteSmartPropertyByProjectIDAndSourceAndObjectID(projectID int64, source string, objectID string) int

	//properties_type
	GetPropertyTypeByKeyValue(projectID int64, eventName string, propertyKey string, propertyValue interface{}, isUserProperty bool) string
	GetPropertyTypeFromDB(projectID int64, eventName, propertyKey string, isUserProperty bool) (int, *model.PropertyDetail)

	// project_analytics
	GetEventUserCountsOfAllProjects(lastNDays int) (map[string][]*model.ProjectAnalytics, error)
	GetEventUserCountsMerged(projectIdsList []int64, lastNDays int, currentDate time.Time) (map[int64]*model.ProjectAnalytics, error)

	// Property details
	CreatePropertyDetails(projectID int64, eventName, propertyKey, propertyType string, isUserProperty bool, allowOverWrite bool) int
	CreateOrDeletePropertyDetails(projectID int64, eventName, enKey, pType string, isUserProperty, allowOverWrite bool) error
	GetAllPropertyDetailsByProjectID(projectID int64, eventName string, isUserProperty bool) (*map[string]string, int)

	// display names
	CreateOrUpdateDisplayNameByObjectType(projectID int64, propertyName, objectType, displayName, group string) int
	GetDisplayNamesForAllEvents(projectID int64) (int, map[string]string)
	GetDisplayNamesForAllEventProperties(projectID int64, eventName string) (int, map[string]string)
	GetDistinctDisplayNamesForAllEventProperties(projectID int64) (int, map[string]string)
	GetDisplayNamesForAllUserProperties(projectID int64) (int, map[string]string)
	GetDisplayNamesForObjectEntities(projectID int64) (int, map[string]string)
	CreateOrUpdateDisplayName(projectID int64, eventName, propertyName, displayName, tag string) int

	// task and task-execution
	RegisterTaskWithDefaultConfiguration(taskName string, source string, frequency int, isProjectEnabled bool) (uint64, int, string)
	RegisterTask(taskName string, source string, frequency int, isProjectEnabled bool, frequencyInterval int, skipStartIndex int, skipEndIndex int, recurrence bool, offsetStartMinutes int) (uint64, int, string)
	GetTaskDetailsByName(taskName string) (model.TaskDetails, int, string)
	GetTaskDetailsById(taskID uint64) (model.TaskDetails, int, string)

	DeregisterTaskDependency(taskId uint64, dependentTaskId uint64) (int, string)
	RegisterTaskDependency(taskId uint64, dependentTaskId uint64, offset int) (int, string)
	GetAllDependentTasks(taskID uint64) ([]model.TaskExecutionDependencyDetails, int, string)
	IsDependencyCircular(taskId, dependentTaskId uint64) bool

	GetAllDeltasByConfiguration(taskID uint64, lookbackInDays int, endDate *time.Time) ([]uint64, int, string)
	GetAllProcessedIntervals(taskID uint64, projectId int64, lookbackInDays int, endDate *time.Time) ([]uint64, int, string)
	InsertTaskBeginRecord(taskId uint64, projectId int64, delta uint64) (int, string)
	InsertTaskEndRecord(taskId uint64, projectId int64, delta uint64) (int, string)
	GetAllToBeExecutedDeltas(taskId uint64, projectId int64, lookbackInDays int, endDate *time.Time) ([]uint64, int, string)
	GetAllInProgressIntervals(taskID uint64, projectId int64, lookbackInDays int, endDate *time.Time) ([]uint64, int, string)
	IsDependentTaskDone(taskId uint64, projectId int64, delta uint64) bool
	DeleteTaskEndRecord(taskId uint64, projectId int64, delta uint64) (int, string)
	GetAllProcessedIntervalsFromStartDate(taskID uint64, projectId int64, startDate *time.Time) ([]uint64, int, string)

	// project model metadata
	CreateProjectModelMetadata(pmm *model.ProjectModelMetadata) (int, string)
	GetProjectModelMetadata(projectId int64) ([]model.ProjectModelMetadata, int, string)
	GetAllProjectModelMetadata() ([]model.ProjectModelMetadata, int, string)

	// search console
	GetGoogleOrganicLastSyncInfoForProject(projectID int64) ([]model.GoogleOrganicLastSyncInfo, int)
	GetAllGoogleOrganicLastSyncInfoForAllProjects() ([]model.GoogleOrganicLastSyncInfo, int)
	CreateGoogleOrganicDocument(gscDoc *model.GoogleOrganicDocument) int
	CreateMultipleGoogleOrganicDocument(gscDocuments []model.GoogleOrganicDocument) int

	// monitoring
	GetProjectIdFromInfo(string) int
	MonitorSlowQueries() ([]interface{}, []interface{}, error)
	CollectTableSizes() map[string]string

	// weekly insights
	CreateWeeklyInsightsMetadata(pmm *model.WeeklyInsightsMetadata) (int, string)
	GetWeeklyInsightsMetadata(projectId int64) ([]model.WeeklyInsightsMetadata, int, string)

	// feedback
	PostFeedback(ProjectID int64, agentUUID string, Feature string, Property *postgres.Jsonb, VoteType int) (int, string)
	GetRecordsFromFeedback(projectID int64, agentUUID string) ([]model.Feedback, error)

	//Group
	CreateGroup(projectID int64, groupName string, allowedGroupNames map[string]bool) (*model.Group, int)
	GetGroup(projectID int64, groupName string) (*model.Group, int)
	CreateOrUpdateGroupPropertiesBySource(projectID int64, groupName string, groupID, groupUserID string,
		enProperties *map[string]interface{}, createdTimestamp, updatedTimestamp int64, source string) (string, error)
	GetGroups(projectID int64) ([]model.Group, int)
	GetPropertiesByGroup(projectID int64, groupName string, limit int, lastNDays int) (map[string][]string, int)
	GetPropertyValuesByGroupProperty(projectID int64, groupName string, propertyName string, limit int, lastNDays int) ([]string, error)

	// Delete channel Integrations
	DeleteChannelIntegration(projectID int64, channelName string) (int, error)

	//group_relationship
	CreateGroupRelationship(projectID int64, leftGroupName, leftGroupUserID, rightGroupName, rightGroupUserID string) (*model.GroupRelationship, int)
	GetGroupRelationshipByUserID(projectID int64, leftGroupUserID string) ([]model.GroupRelationship, int)

	//Content-groups
	GetAllContentGroups(projectID int64) ([]model.ContentGroup, int)
	GetContentGroupById(id string, projectID int64) (model.ContentGroup, int)
	CreateContentGroup(projectID int64, contentGroup model.ContentGroup) (model.ContentGroup, int, string)
	DeleteContentGroup(id string, projectID int64) (int, string)
	UpdateContentGroup(id string, projectID int64, contentGroup model.ContentGroup) (model.ContentGroup, int, string)
	CheckURLContentGroupValue(pageUrl string, projectID int64) map[string]string

	// fivetran mappings
	DisableFiveTranMapping(ProjectID int64, Integration string, ConnectorId string) error
	EnableFiveTranMapping(ProjectID int64, Integration string, ConnectorId string, Accounts string) error
	GetFiveTranMapping(ProjectID int64, Integration string) (string, error)
	GetActiveFiveTranMapping(ProjectID int64, Integration string) (model.FivetranMappings, error)
	GetAllActiveFiveTranMapping(ProjectID int64, Integration string) ([]string, error)
	GetLatestFiveTranMapping(ProjectID int64, Integration string) (string, string, error)
	PostFiveTranMapping(ProjectID int64, Integration string, ConnectorId string, SchemaId string, Accounts string) error
	GetAllActiveFiveTranMappingByIntegration(Integration string) ([]model.FivetranMappings, error)
	UpdateFiveTranMappingAccount(ProjectID int64, Integration string, ConnectorId string, Accounts string) error

	//leadgen
	GetLeadgenSettingsForProject(projectID int64) ([]model.LeadgenSettings, error)
	GetLeadgenSettings() ([]model.LeadgenSettings, error)
	UpdateRowRead(projectID int64, source int, rowRead int64) (int, error)

	// integration document
	InsertIntegrationDocument(doc model.IntegrationDocument) error
	UpsertIntegrationDocument(doc model.IntegrationDocument) error

	// alerts
	SetAuthTokenforSlackIntegration(projectID int64, agentUUID string, authTokens model.SlackAccessTokens) error
	GetSlackAuthToken(projectID int64, agentUUID string) (model.SlackAccessTokens, error)
	DeleteSlackIntegration(projectID int64, agentUUID string) error
	GetAlertById(id string, projectID int64) (model.Alert, int)
	GetAllAlerts(projectID int64, excludeSavedQueries bool) ([]model.Alert, int)
	DeleteAlert(id string, projectID int64) (int, string)
	CreateAlert(projectID int64, alert model.Alert) (model.Alert, int, string)
	GetAlertNamesByProjectIdTypeAndName(projectID int64, nameOfQuery string) ([]string, int)
	UpdateAlert(lastAlertSent bool) (int, string)

	// sharable url
	CreateShareableURL(sharableURLParams *model.ShareableURL) (*model.ShareableURL, int)
	GetAllShareableURLsWithProjectIDAndAgentID(projectID int64, agentUUID string) ([]*model.ShareableURL, int)
	GetShareableURLWithShareStringAndAgentID(projectID int64, shareId, agentUUID string) (*model.ShareableURL, int)
	GetShareableURLWithShareStringWithLargestScope(projectID int64, shareId string, entityType int) (*model.ShareableURL, int)
	// GetShareableURLWithID(projectID int64, shareId string) (*model.ShareableURL, int)
	// UpdateShareableURLShareTypeWithShareIDandCreatedBy(projectID int64, shareId, createdBy string, shareType int, allowedUsers string) int
	DeleteShareableURLWithShareIDandAgentID(projectID int64, shareId, createdBy string) int
	DeleteShareableURLWithEntityIDandType(projectID int64, entityID int64, entityType int) int
	RevokeShareableURLsWithShareString(projectId int64, shareString string) (int, string)
	RevokeShareableURLsWithProjectID(projectId int64) (int, string)

	CreateSharableURLAudit(sharableURL *model.ShareableURL, agentId string) int

	//crm
	CreateCRMUser(crmUser *model.CRMUser) (int, error)
	CreateCRMGroup(crmGroup *model.CRMGroup) (int, error)
	CreateCRMActivity(crmActivity *model.CRMActivity) (int, error)
	CreateCRMRelationship(crmRelationship *model.CRMRelationship) (int, error)
	CreateCRMProperties(crmProperty *model.CRMProperty) (int, error)
	GetCRMUserByTypeAndAction(projectID int64, source U.CRMSource, id string, userType int, action model.CRMAction) (*model.CRMUser, int)
	UpdateCRMUserAsSynced(projectID int64, source U.CRMSource, crmUser *model.CRMUser, userID, syncID string) (*model.CRMUser, int)
	GetCRMUsersInOrderForSync(projectID int64, source U.CRMSource, startTimestamp, endTimestamp int64) ([]model.CRMUser, int)
	GetCRMActivityInOrderForSync(projectID int64, source U.CRMSource, startTimestamp, endTimestamp int64) ([]model.CRMActivity, int)
	GetCRMActivityMinimumTimestampForSync(projectID int64, source U.CRMSource) (int64, int)
	GetCRMUsersMinimumTimestampForSync(projectID int64, source U.CRMSource) (int64, int)
	GetCRMPropertiesForSync(projectID int64) ([]model.CRMProperty, int)
	GetActivitiesDistinctEventNamesByType(projectID int64, objectTypes []int) (map[int][]string, int)
	UpdateCRMProperyAsSynced(projectID int64, source U.CRMSource, crmProperty *model.CRMProperty) (*model.CRMProperty, int)
	UpdateCRMActivityAsSynced(projectID int64, source U.CRMSource, crmActivity *model.CRMActivity, syncID, userID string) (*model.CRMActivity, int)

	GetCRMSetting(projectID int64) (*model.CRMSetting, int)
	GetAllCRMSetting() ([]model.CRMSetting, int)
	UpdateCRMSetting(projectID int64, option model.CRMSettingOption) int
	CreateOrUpdateCRMSettingHubspotEnrich(projectID int64, isHeavy bool, maxCreatedAtSec *int64) int

	// data availability checks
	GetLatestDataStatus(integrations []string, project_id int64, hardRefresh bool) (map[string]model.DataAvailabilityStatus, error)
	IsIntegrationAvailable(projectID int64) map[string]bool
	GetIntegrationStatusByProjectId(project_id int64) (map[string]model.DataAvailability, int)
	IsDataAvailable(project_id int64, integration string, timestamp uint64) bool
	FindLatestProcessedStatus(integration string, projectID int64) uint64
	IsAdwordsIntegrationAvailable(projectID int64) bool
	IsFacebookIntegrationAvailable(projectID int64) bool
	IsGoogleOrganicIntegrationAvailable(projectID int64) bool
	IsLinkedInIntegrationAvailable(projectID int64) bool
	IsHubspotIntegrationAvailable(projectID int64) bool
	IsSalesforceIntegrationAvailable(projectID int64) bool
	IsMarketoIntegrationAvailable(projectID int64) bool

	// Timeline
	GetProfilesListByProjectId(projectID int64, payload model.TimelinePayload, profileType string) ([]model.Profile, int)
	GetProfileUserDetailsByID(projectID int64, identity string, isAnonymous string) (*model.ContactDetails, int)
	GetGroupsForUserTimeline(projectID int64, userDetails model.ContactDetails) []model.GroupsInfo
	GetUserActivitiesAndSessionCount(projectID int64, identity string, userId string) ([]model.UserActivity, uint64)
	GetProfileAccountDetailsByID(projectID int64, id string) (*model.AccountDetails, int)

	// Ads import
	GetAllAdsImportEnabledProjects() (map[int64]map[string]model.LastProcessedAdsImport, error)
	UpdateLastProcessedAdsData(updatedFields map[string]model.LastProcessedAdsImport, projectId int64) int
	GetCustomAdsSourcesByProject(projectID int64) ([]string, int)
	IsCustomAdsAvailable(projectID int64) bool

	// custom ads
	GetKPIConfigsForCustomAds(projectID int64, reqID string, includeDerivedKPIs bool) ([]map[string]interface{}, int)
	GetKPIConfigsForCustomAdsFromDB(projectID int64, includeDerivedKPIs bool) []map[string]interface{}
	GetCustomChannelFilterValuesV1(projectID int64, source, channel, filterObject, filterProperty string, reqID string) (model.ChannelFilterValues, int)

	// Predict Job
	GetGroupsOnEvent(projectID int64, event_name string) (*sql.Rows, *sql.Tx, error)
	PullUserCohortDataOnEvent(projectID int64, startTime, endTime int64, event_id string, filter_property string) (*sql.Rows, error)
	// PullUsersEventRowsForPredictJob(project_id int64, event_id string, start_time int64, end_time int64) (*sql.Rows, *sql.Tx, error)
	GetAllEventsWithUsers(projectID int64, event_name string, start_time int64, end_time int64) (*sql.Rows, error)
	GetAllEventsOnUsersBetweenTime(projectID int64, users []string, start_time int64, end_time int64) (*sql.Rows, *sql.Tx, error)
	GetUsersTimestampOnFirstEvent(projectID int64, event_name string, start_time int64, end_time int64) (map[string]int64, error)
	GetAllEventsOnUsers(projectId int64, arrayCustomerUserID []string) (*sql.Rows, error)
	GetCountOfGroupIDS(projectId int64, arrayGroupID []string) (*sql.Rows, error)
	PullEventRowsOnUsers(projectID int64, users []string, start_time int64, end_time int64) (*sql.Rows, error)
	GetUsersEventTimeStampFromHistory(projectID int64, event_name_id string, userIdFiltered map[string]int64) (map[string]int64, error)
	GetBaseEventsOnUsers(projectID int64, event_name_id string, start_time int64, end_time int64, users []string) (*sql.Rows, error)

	// property overides
	GetPropertyOverridesByType(projectID int64, typeConstant int, entity int) (int, []string)

	UpdatePathAnalysisEntity(projectID int64, id string, status string) (int, string)
	GetAllSavedPathAnalysisEntityByProject(projectID int64) ([]model.PathAnalysis, int)
	//path analysis
	GetAllPathAnalysisEntityByProject(projectID int64) ([]model.PathAnalysisEntityInfo, int)
	GetPathAnalysisEntity(projectID int64, id string) (model.PathAnalysis, int)
	CreatePathAnalysisEntity(userID string, projectId int64, entity *model.PathAnalysisQuery) (*model.PathAnalysis, int, string)
	DeletePathAnalysisEntity(projectID int64, id string) (int, string)
	GetProjectCountWithStatus(projectID int64, status []string) (int, int, string)

	// leadsquaredmarker
	CreateLeadSquaredMarker(marker model.LeadsquaredMarker) int
	GetLeadSquaredMarker(ProjectID int64, Delta int64, Document string, Tag string) (int, int)
}
