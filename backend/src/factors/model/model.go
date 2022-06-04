package model

import (
	"database/sql"
	"factors/filestore"
	"factors/model/model"
	U "factors/util"

	"sync"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// Model - Interface of all methods to be implemented by the stores.
type Model interface {
	// adwords_document
	CreateAdwordsDocument(adwordsDoc *model.AdwordsDocument) int
	CreateMultipleAdwordsDocument(adwordsDoc []model.AdwordsDocument) int
	GetAdwordsLastSyncInfoForProject(projectID uint64) ([]model.AdwordsLastSyncInfo, int)
	GetAllAdwordsLastSyncInfoForAllProjects() ([]model.AdwordsLastSyncInfo, int)
	PullGCLIDReport(projectID uint64, from, to int64, adwordsAccountIDs string, campaignIDReport, adgroupIDReport, keywordIDReport map[string]model.MarketingData, timeZone string) (map[string]model.MarketingData, error)
	GetAdwordsFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int)
	GetAdwordsSQLQueryAndParametersForFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int)
	ExecuteAdwordsChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int)
	ExecuteAdwordsChannelQuery(projectID uint64, query *model.ChannelQuery) (*model.ChannelQueryResult, int)
	GetAdwordsFilterValuesByType(projectID uint64, docType int) ([]string, int)
	GetSQLQueryAndParametersForAdwordsQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, int)
	GetAdwordsChannelResultMeta(projectID uint64, customerAccountID string, query *model.ChannelQuery) (*model.ChannelQueryResultMeta, error)

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
	GetPrimaryAgentOfProject(projectId uint64) (uuid string, errCode int)
	UpdateAgentSalesforceInstanceURL(agentUUID string, instanceURL string) int
	IsSlackIntegratedForProject(projectID uint64, agentUUID string) (bool, int)
	UpdateLastLoggedOut(agentUUID string, timestamp int64) int

	// analytics
	ExecQuery(stmnt string, params []interface{}) (*model.QueryResult, error, string)
	ExecQueryWithContext(stmnt string, params []interface{}) (*sql.Rows, *sql.Tx, error, string)
	Analyze(projectID uint64, query model.Query) (*model.QueryResult, int, string)

	// archival
	GetNextArchivalBatches(projectID uint64, startTime int64, maxLookbackDays int, hardStartTime, hardEndTime time.Time) ([]model.EventsArchivalBatch, error)

	// attribution
	ExecuteAttributionQuery(projectID uint64, query *model.AttributionQuery, debugQueryKey string) (*model.QueryResult, error)
	GetCoalesceIDFromUserIDs(userIDs []string, projectID uint64, logCtx log.Entry) (map[string]model.UserInfo, error)
	GetLinkedFunnelEventUsersFilter(projectID uint64, queryFrom, queryTo int64,
		linkedEvents []model.QueryEventWithProperties, eventNameToId map[string][]interface{},
		userIDInfo map[string]model.UserInfo, logCtx log.Entry) (error, []model.UserEventInfo)
	GetAdwordsCurrency(projectId uint64, customerAccountId string, from, to int64) (string, error)
	GetConvertedUsersWithFilter(projectID uint64, goalEventName string,
		goalEventProperties []model.QueryProperty, conversionFrom, conversionTo int64,
		eventNameToIdList map[string][]interface{}, logCtx log.Entry) (map[string]model.UserInfo,
		map[string][]model.UserIDPropID, map[string]int64, error)

	// bigquery_setting
	CreateBigquerySetting(setting *model.BigquerySetting) (*model.BigquerySetting, int)
	UpdateBigquerySettingLastRunAt(settingID string, lastRunAt int64) (int64, int)
	GetBigquerySettingByProjectID(projectID uint64) (*model.BigquerySetting, int)

	// billing_account
	GetBillingAccountByProjectID(projectID uint64) (*model.BillingAccount, int)
	GetBillingAccountByAgentUUID(AgentUUID string) (*model.BillingAccount, int)
	UpdateBillingAccount(id string, planId uint64, orgName, billingAddr, pinCode, phoneNo string) int
	GetProjectsUnderBillingAccountID(ID string) ([]model.Project, int)
	GetAgentsByProjectIDs(projectIDs []uint64) ([]*model.Agent, int)
	GetAgentsUnderBillingAccountID(ID string) ([]*model.Agent, int)
	IsNewProjectAgentMappingCreationAllowed(projectID uint64, emailOfAgentToAdd string) (bool, int)

	// channel_analytics
	GetChannelFilterValuesV1(projectID uint64, channel, filterObject, filterProperty string, reqID string) (model.ChannelFilterValues, int)
	GetAllChannelFilterValues(projectID uint64, filterObject, filterProperty string, reqID string) ([]interface{}, int)
	RunChannelGroupQuery(projectID uint64, queries []model.ChannelQueryV1, reqID string) (model.ChannelResultGroupV1, int)
	ExecuteChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) (*model.ChannelQueryResultV1, int)
	GetChannelFilterValues(projectID uint64, channel, filter string) ([]string, int)
	ExecuteChannelQuery(projectID uint64, query *model.ChannelQuery) (*model.ChannelQueryResult, int)
	ExecuteSQL(sqlStatement string, params []interface{}, logCtx *log.Entry) ([]string, [][]interface{}, error)
	GetChannelConfig(projectID uint64, channel string, reqID string) (*model.ChannelConfigResult, int)

	// KPI Related but in different modules
	GetKPIConfigsForWebsiteSessions(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForPageViews(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForFormSubmissions(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForHubspotContacts(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForHubspotCompanies(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForHubspotDeals(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForSalesforceUsers(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForSalesforceAccounts(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForSalesforceOpportunities(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForAdwords(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForBingAds(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForGoogleOrganic(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForFacebook(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForLinkedin(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForAllChannels(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForMarketoLeads(projectID uint64, reqID string) (map[string]interface{}, int)
	GetKPIConfigsForMarketo(projectID uint64, reqID string, displayCategory string) (map[string]interface{}, int)

	// ExecuteKPIQueryGroup(kpiQueryGroup model.KPIQueryGroup)
	ExecuteKPIQueryGroup(projectID uint64, reqID string, kpiQueryGroup model.KPIQueryGroup) ([]model.QueryResult, int)
	ExecuteKPIQueryForEvents(projectID uint64, reqID string, kpiQuery model.KPIQuery) ([]model.QueryResult, int)
	ExecuteKPIQueryForChannels(projectID uint64, reqID string, kpiQuery model.KPIQuery) ([]model.QueryResult, int)

	// Custom Metrics
	CreateCustomMetric(customMetric model.CustomMetric) (*model.CustomMetric, string, int)
	GetCustomMetricsByProjectId(projectID uint64) ([]model.CustomMetric, string, int)
	GetCustomMetricByProjectIdAndObjectType(projectID uint64, queryType int, objectType string) ([]model.CustomMetric, string, int)
	GetCustomMetricsByName(projectID uint64, name string) (model.CustomMetric, string, int)
	GetCustomMetricsByID(projectID uint64, id string) (model.CustomMetric, string, int)
	DeleteCustomMetricByID(projectID uint64, id string) int

	//templates
	RunTemplateQuery(projectID uint64, query model.TemplateQuery, reqID string) (model.TemplateResponse, int)
	GetTemplateConfig(projectID uint64, templateType int) (model.TemplateConfig, int)
	UpdateTemplateConfig(projectID uint64, templateType int, thresholds []model.TemplateThreshold) ([]model.TemplateThreshold, string)

	// dashboard_unit
	CreateDashboardUnit(projectID uint64, agentUUID string, dashboardUnit *model.DashboardUnit) (*model.DashboardUnit, int, string)
	CreateDashboardUnitForDashboardClass(projectID uint64, agentUUID string, dashboardUnit *model.DashboardUnit, dashboardClass string) (*model.DashboardUnit, int, string)
	CreateDashboardUnitForMultipleDashboards(dashboardIds []uint64, projectId uint64, agentUUID string, unitPayload model.DashboardUnitRequestPayload) ([]*model.DashboardUnit, int, string)
	CreateMultipleDashboardUnits(requestPayload []model.DashboardUnitRequestPayload, projectId uint64, agentUUID string, dashboardId uint64) ([]*model.DashboardUnit, int, string)
	GetDashboardUnitsForProjectID(projectID uint64) ([]model.DashboardUnit, int)
	GetDashboardUnits(projectID uint64, agentUUID string, dashboardId uint64) ([]model.DashboardUnit, int)
	GetDashboardUnitByUnitID(projectID, unitID uint64) (*model.DashboardUnit, int)
	GetDashboardUnitsByProjectIDAndDashboardIDAndTypes(projectID, dashboardID uint64, types []string) ([]model.DashboardUnit, int)
	DeleteDashboardUnit(projectID uint64, agentUUID string, dashboardId uint64, id uint64) int
	DeleteMultipleDashboardUnits(projectID uint64, agentUUID string, dashboardID uint64, dashboardUnitIDs []uint64) (int, string)
	UpdateDashboardUnit(projectId uint64, agentUUID string, dashboardId uint64, id uint64, unit *model.DashboardUnit) (*model.DashboardUnit, int)
	CacheDashboardUnitsForProjects(stringProjectsIDs, excludeProjectIDs string, numRoutines int, reportCollector *sync.Map)
	CacheDashboardUnitsForProjectID(projectID uint64, dashboardUnits []model.DashboardUnit, queryClasses []string, numRoutines int, reportCollector *sync.Map) int
	CacheDashboardUnit(dashboardUnit model.DashboardUnit, waitGroup *sync.WaitGroup, reportCollector *sync.Map, queryClass string)
	GetQueryAndClassFromDashboardUnit(dashboardUnit *model.DashboardUnit) (queryClass string, queryInfo *model.Queries, errMsg string)
	GetQueryClassFromQueries(query model.Queries) (queryClass, errMsg string)
	GetQueryAndClassFromQueryIdString(queryIdString string, projectId uint64) (queryClass string, queryInfo *model.Queries, errMsg string)
	GetQueryWithQueryIdString(projectID uint64, queryIDString string) (*model.Queries, int)
	CacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload) (int, string, model.CachingUnitReport)
	CacheDashboardsForMonthlyRange(projectIDs, excludeProjectIDs string, numMonths, numRoutines int, reportCollector *sync.Map)
	GetDashboardUnitNamesByProjectIdTypeAndName(projectID uint64, reqID string, typeOfQuery string, nameOfQuery string) ([]string, int)

	// dashboard
	CreateDashboard(projectID uint64, agentUUID string, dashboard *model.Dashboard) (*model.Dashboard, int)
	CreateAgentPersonalDashboardForProject(projectID uint64, agentUUID string) (*model.Dashboard, int)
	GetDashboards(projectID uint64, agentUUID string) ([]model.Dashboard, int)
	GetDashboard(projectID uint64, agentUUID string, id uint64) (*model.Dashboard, int)
	HasAccessToDashboard(projectID uint64, agentUUID string, id uint64) (bool, *model.Dashboard)
	UpdateDashboard(projectID uint64, agentUUID string, id uint64, dashboard *model.UpdatableDashboard) int
	DeleteDashboard(projectID uint64, agentUUID string, dashboardID uint64) int

	// event_analytics
	RunEventsGroupQuery(queriesOriginal []model.Query, projectId uint64) (model.ResultGroup, int)
	ExecuteEventsQuery(projectID uint64, query model.Query) (*model.QueryResult, int, string)
	RunInsightsQuery(projectID uint64, query model.Query) (*model.QueryResult, int, string)
	BuildInsightsQuery(projectID uint64, query model.Query) (string, []interface{}, error)

	// Profile
	RunProfilesGroupQuery(queriesOriginal []model.ProfileQuery, projectID uint64) (model.ResultGroup, int)
	ExecuteProfilesQuery(projectID uint64, query model.ProfileQuery) (*model.QueryResult, int, string)

	// event_name
	CreateOrGetEventName(eventName *model.EventName) (*model.EventName, int)
	CreateOrGetUserCreatedEventName(eventName *model.EventName) (*model.EventName, int)
	CreateOrGetAutoTrackedEventName(eventName *model.EventName) (*model.EventName, int)
	CreateOrGetFilterEventName(eventName *model.EventName) (*model.EventName, int)
	CreateOrGetCRMSmartEventFilterEventName(projectID uint64, eventName *model.EventName, filterExpr *model.SmartCRMEventFilter) (*model.EventName, int)
	GetSmartEventEventName(eventName *model.EventName) (*model.EventName, int)
	GetSmartEventEventNameByNameANDType(projectID uint64, name, typ string) (*model.EventName, int)
	CreateOrGetSessionEventName(projectID uint64) (*model.EventName, int)
	CreateOrGetOfflineTouchPointEventName(projectID uint64) (*model.EventName, int)
	GetSessionEventName(projectID uint64) (*model.EventName, int)
	GetEventName(name string, projectID uint64) (*model.EventName, int)
	GetEventNames(projectID uint64) ([]model.EventName, int)
	GetOrderedEventNamesFromDb(projectID uint64, startTimestamp int64, endTimestamp int64, limit int) ([]model.EventNameWithAggregation, error)
	GetFilterEventNames(projectID uint64) ([]model.EventName, int)
	GetSmartEventFilterEventNames(projectID uint64, includeDeleted bool) ([]model.EventName, int)
	GetSmartEventFilterEventNameByID(projectID uint64, id string, isDeleted bool) (*model.EventName, int)
	GetEventNamesByNames(projectID uint64, names []string) ([]model.EventName, int)
	DeleteSmartEventFilter(projectID uint64, id string) (*model.EventName, int)
	GetFilterEventNamesByExprPrefix(projectID uint64, prefix string) ([]model.EventName, int)
	UpdateEventName(projectId uint64, id string, nameType string, eventName *model.EventName) (*model.EventName, int)
	UpdateCRMSmartEventFilter(projectID uint64, id string, eventName *model.EventName, filterExpr *model.SmartCRMEventFilter) (*model.EventName, int)
	UpdateFilterEventName(projectID uint64, id string, eventName *model.EventName) (*model.EventName, int)
	DeleteFilterEventName(projectID uint64, id string) int
	FilterEventNameByEventURL(projectID uint64, eventURL string) (*model.EventName, int)
	GetEventNameFromEventNameId(eventNameId string, projectID uint64) (*model.EventName, error)
	GetEventTypeFromDb(projectID uint64, eventNames []string, limit int64) (map[string]string, error)
	GetMostFrequentlyEventNamesByType(projectID uint64, limit int, lastNDays int, typeOfEvent string) ([]string, error)
	GetEventNamesOrderedByOccurenceAndRecency(projectID uint64, limit int, lastNDays int) (map[string][]string, error)
	GetPropertiesByEvent(projectID uint64, eventName string, limit int, lastNDays int) (map[string][]string, error)
	GetPropertyValuesByEventProperty(projectID uint64, eventName string, propertyName string, limit int, lastNDays int) ([]string, error)
	GetPropertiesForHubspot(projectID uint64, reqID string) []map[string]string
	GetPropertiesForSalesforce(projectID uint64, reqID string) []map[string]string
	GetPropertiesForMarketo(projectID uint64, reqID string) []map[string]string

	// events
	GetEventCountOfUserByEventName(projectID uint64, userId string, eventNameId string) (uint64, int)
	GetEventCountOfUsersByEventName(projectID uint64, userIDs []string, eventNameID string) (uint64, int)
	CreateEvent(event *model.Event) (*model.Event, int)
	GetEvent(projectID uint64, userId string, id string) (*model.Event, int)
	GetEventById(projectID uint64, id, userID string) (*model.Event, int)
	GetLatestEventOfUserByEventNameId(projectId uint64, userId string, eventNameId string, startTimestamp int64, endTimestamp int64) (*model.Event, int)
	GetRecentEventPropertyKeysWithLimits(projectID uint64, eventName string, starttime int64, endtime int64, eventsLimit int) ([]U.Property, error)
	GetRecentEventPropertyValuesWithLimits(projectID uint64, eventName string, property string, valuesLimit int, rowsLimit int, starttime int64, endtime int64) ([]U.PropertyValue, string, error)
	UpdateEventProperties(projectId uint64, id, userID string, properties *U.PropertiesMap, updateTimestamp int64, optionalEventUserProperties *postgres.Jsonb) int
	GetUserEventsByEventNameId(projectID uint64, userId string, eventNameId string) ([]model.Event, int)
	OverwriteEventProperties(projectId uint64, userId string, eventId string, newEventProperties *postgres.Jsonb) int
	OverwriteEventPropertiesByID(projectId uint64, id string, newEventProperties *postgres.Jsonb) int
	AddSessionForUser(projectId uint64, userId string, userEvents []model.Event, bufferTimeBeforeSessionCreateInSecs int64, sessionEventNameId string) (int, int, bool, int, int)
	GetDatesForNextEventsArchivalBatch(projectID uint64, startTime int64) (map[string]int64, int)
	GetAllEventsForSessionCreationAsUserEventsMap(projectId uint64, sessionEventNameId string, startTimestamp, endTimestamp int64) (*map[string][]model.Event, int, int)
	GetEventsWithoutPropertiesAndWithPropertiesByNameForYourStory(projectID uint64, from, to int64, mandatoryProperties []string) ([]model.EventWithProperties, *map[string]U.PropertiesMap, int)
	OverwriteEventUserPropertiesByID(projectID uint64, userID, id string, properties *postgres.Jsonb) int
	PullEventRowsForBuildSequenceJob(projectID uint64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error)
	PullEventRowsForArchivalJob(projectID uint64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error)
	GetUnusedSessionIDsForJob(projectID uint64, startTimestamp, endTimestamp int64) ([]string, int)
	DeleteEventsByIDsInBatchForJob(projectID uint64, eventNameID string, ids []string, batchSize int) int
	DeleteEventByIDs(projectID uint64, eventNameID string, ids []string) int
	AssociateSessionByEventIds(projectId uint64, userID string, events []*model.Event, sessionId string, sessionEventNameId string) int
	GetHubspotFormEvents(projectID uint64, userId string, timestamps []interface{}) ([]model.Event, int)

	// facebook_document
	CreateFacebookDocument(projectID uint64, document *model.FacebookDocument) int
	GetFacebookSQLQueryAndParametersForFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int)
	ExecuteFacebookChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int)
	GetFacebookLastSyncInfo(projectID uint64, CustomerAdAccountID string) ([]model.FacebookLastSyncInfo, int)
	ExecuteFacebookChannelQuery(projectID uint64, query *model.ChannelQuery) (*model.ChannelQueryResult, int)
	GetSQLQueryAndParametersForFacebookQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, int)
	GetFacebookMetricBreakdown(projectID uint64, customerAccountID string,
		query *model.ChannelQuery) (*model.ChannelBreakdownResult, error)
	GetFacebookChannelResult(projectID uint64, customerAccountID string,
		query *model.ChannelQuery) (*model.ChannelQueryResult, error)

	// linkedin document
	CreateLinkedinDocument(projectID uint64, document *model.LinkedinDocument) int
	GetLinkedinSQLQueryAndParametersForFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int)
	ExecuteLinkedinChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int)
	GetLinkedinLastSyncInfo(projectID uint64, CustomerAdAccountID string) ([]model.LinkedinLastSyncInfo, int)
	ExecuteLinkedinChannelQuery(projectID uint64, query *model.ChannelQuery) (*model.ChannelQueryResult, int)
	GetSQLQueryAndParametersForLinkedinQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool,
		limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, int)

	//bingads document
	GetBingadsFilterValuesSQLAndParams(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int)

	// funnel_analytics
	RunFunnelQuery(projectID uint64, query model.Query) (*model.QueryResult, int, string)

	// goals
	GetAllFactorsGoals(ProjectID uint64) ([]model.FactorsGoal, int)
	GetAllActiveFactorsGoals(ProjectID uint64) ([]model.FactorsGoal, int)
	CreateFactorsGoal(ProjectID uint64, Name string, Rule model.FactorsGoalRule, agentUUID string) (int64, int, string)
	DeactivateFactorsGoal(ID int64, ProjectID uint64) (int64, int)
	ActivateFactorsGoal(Name string, ProjectID uint64) (int64, int)
	UpdateFactorsGoal(ID int64, Name string, Rule model.FactorsGoalRule, ProjectID uint64) (int64, int)
	GetFactorsGoal(Name string, ProjectID uint64) (*model.FactorsGoal, error)
	GetFactorsGoalByID(ID int64, ProjectID uint64) (*model.FactorsGoal, error)
	GetAllFactorsGoalsWithNamePattern(ProjectID uint64, NamePattern string) ([]model.FactorsGoal, int)

	// hubspot_document
	GetHubspotDocumentByTypeAndActions(projectId uint64, ids []string, docType int, actions []int) ([]model.HubspotDocument, int)
	GetSyncedHubspotDocumentByFilter(projectId uint64, id string, docType, action int) (*model.HubspotDocument, int)
	CreateHubspotDocument(projectID uint64, document *model.HubspotDocument) int
	CreateHubspotDocumentInBatch(projectID uint64, docType int, documents []*model.HubspotDocument, batchSize int) int
	GetHubspotSyncInfo() (*model.HubspotSyncInfo, int)
	GetHubspotFirstSyncProjectsInfo() (*model.HubspotSyncInfo, int)
	UpdateHubspotProjectSettingsBySyncStatus(success []model.HubspotProjectSyncStatus, failure []model.HubspotProjectSyncStatus, syncAll bool) int
	GetHubspotDocumentBeginingTimestampByDocumentTypeForSync(projectID uint64, docTypes []int) (int64, int)
	GetHubspotFormDocuments(projectID uint64) ([]model.HubspotDocument, int)
	GetHubspotDocumentsByTypeForSync(projectID uint64, typ int, maxCreatedAtSec int64) ([]model.HubspotDocument, int)
	GetHubspotContactCreatedSyncIDAndUserID(projectID uint64, docID string) ([]model.HubspotDocument, int)
	GetHubspotDocumentsByTypeANDRangeForSync(projectID uint64, docType int, from, to, maxCreatedAtSec int64) ([]model.HubspotDocument, int)
	GetSyncedHubspotDealDocumentByIdAndStage(projectId uint64, id string, stage string) (*model.HubspotDocument, int)
	GetHubspotObjectPropertiesName(ProjectID uint64, objectType string) ([]string, []string)
	UpdateHubspotDocumentAsSynced(projectID uint64, id string, docType int, syncId string, timestamp int64, action int, userID, groupUserID string) int
	GetLastSyncedHubspotDocumentByID(projectID uint64, docID string, docType int) (*model.HubspotDocument, int)
	GetLastSyncedHubspotUpdateDocumentByID(projectID uint64, docID string, docType int) (*model.HubspotDocument, int)
	GetAllHubspotObjectValuesByPropertyName(ProjectID uint64, objectType, propertyName string) []interface{}
	GetHubspotDocumentCountForSync(projectIDs []uint64, docTypes []int, maxCreatedAtSec int64) ([]model.HubspotDocumentCount, int)

	// plan
	GetPlanByID(planID uint64) (*model.Plan, int)
	GetPlanByCode(Code string) (*model.Plan, int)

	// project_agent_mapping
	CreateProjectAgentMappingWithDependencies(pam *model.ProjectAgentMapping) (*model.ProjectAgentMapping, int)
	CreateProjectAgentMappingWithDependenciesWithoutDashboard(pam *model.ProjectAgentMapping) (*model.ProjectAgentMapping, int)
	GetProjectAgentMapping(projectID uint64, agentUUID string) (*model.ProjectAgentMapping, int)
	GetProjectAgentMappingsByProjectId(projectID uint64) ([]model.ProjectAgentMapping, int)
	GetProjectAgentMappingsByProjectIds(projectIds []uint64) ([]model.ProjectAgentMapping, int)
	GetProjectAgentMappingsByAgentUUID(agentUUID string) ([]model.ProjectAgentMapping, int)
	DoesAgentHaveProject(agentUUID string) int
	DeleteProjectAgentMapping(projectID uint64, agentUUIDToRemove string) int
	EditProjectAgentMapping(projectID uint64, agentUUIDToEdit string, role int64) int

	// project_billing_account
	GetProjectBillingAccountMappings(billingAccountID string) ([]model.ProjectBillingAccountMapping, int)
	GetProjectBillingAccountMapping(projectID uint64) (*model.ProjectBillingAccountMapping, int)

	// project_setting
	GetProjectSetting(projectID uint64) (*model.ProjectSetting, int)
	GetClearbitKeyFromProjectSetting(projectId uint64) (string, int)
	GetProjectSettingByKeyWithTimeout(key, value string, timeout time.Duration) (*model.ProjectSetting, int)
	GetProjectSettingByTokenWithCacheAndDefault(token string) (*model.ProjectSetting, int)
	GetProjectSettingByPrivateTokenWithCacheAndDefault(privateToken string) (*model.ProjectSetting, int)
	GetProjectIDByToken(token string) (uint64, int)
	UpdateProjectSettings(projectID uint64, settings *model.ProjectSetting) (*model.ProjectSetting, int)
	GetIntAdwordsRefreshTokenForProject(projectID uint64) (string, int)
	GetIntGoogleOrganicRefreshTokenForProject(projectID uint64) (string, int)
	GetIntAdwordsProjectSettingsForProjectID(projectID uint64) ([]model.AdwordsProjectSettings, int)
	GetAllIntAdwordsProjectSettings() ([]model.AdwordsProjectSettings, int)
	GetAllHubspotProjectSettings() ([]model.HubspotProjectSettings, int)
	GetFacebookEnabledIDsAndProjectSettings() ([]uint64, []model.FacebookProjectSettings, int)
	GetFacebookEnabledIDsAndProjectSettingsForProject(projectIDs []uint64) ([]uint64, []model.FacebookProjectSettings, int)
	GetLinkedinEnabledProjectSettings() ([]model.LinkedinProjectSettings, int)
	GetLinkedinEnabledProjectSettingsForProjects(projectIDs []string) ([]model.LinkedinProjectSettings, int)
	GetArchiveEnabledProjectIDs() ([]uint64, int)
	GetBigqueryEnabledProjectIDs() ([]uint64, int)
	GetAllSalesforceProjectSettings() ([]model.SalesforceProjectSettings, int)
	IsPSettingsIntShopifyEnabled(projectId uint64) bool
	GetProjectDetailsByShopifyDomain(shopifyDomain string) (uint64, string, bool, int)
	EnableBigqueryArchivalForProject(projectID uint64) int
	IsBingIntegrationAvailable(projectID uint64) bool

	// project
	UpdateProject(projectID uint64, project *model.Project) int
	CreateProjectWithDependencies(project *model.Project, agentUUID string, agentRole uint64, billingAccountID string, createDashboard bool) (*model.Project, int)
	CreateDefaultProjectForAgent(agentUUID string) (*model.Project, int)
	GetProject(id uint64) (*model.Project, int)
	GetProjectByToken(token string) (*model.Project, int)
	GetProjectByPrivateToken(privateToken string) (*model.Project, int)
	GetProjects() ([]model.Project, int)
	GetProjectsByIDs(ids []uint64) ([]model.Project, int)
	GetAllProjectIDs() ([]uint64, int)
	GetNextSessionStartTimestampForProject(projectID uint64) (int64, int)
	UpdateNextSessionStartTimestampForProject(projectID uint64, timestamp int64) int
	GetProjectsToRunForIncludeExcludeString(projectIDs, excludeProjectIDs string) []uint64
	GetProjectsWithoutWebAnalyticsDashboard(onlyProjectsMap map[uint64]bool) (projectIds []uint64, errCode int)
	GetTimezoneForProject(projectID uint64) (U.TimeZoneString, int)

	// queries
	CreateQuery(projectID uint64, query *model.Queries) (*model.Queries, int, string)
	GetALLQueriesWithProjectId(projectID uint64) ([]model.Queries, int)
	GetDashboardQueryWithQueryId(projectID uint64, queryID uint64) (*model.Queries, int)
	GetSavedQueryWithQueryId(projectID uint64, queryID uint64) (*model.Queries, int)
	GetQueryWithQueryId(projectID uint64, queryID uint64) (*model.Queries, int)
	DeleteQuery(projectID uint64, queryID uint64) (int, string)
	DeleteSavedQuery(projectID uint64, queryID uint64) (int, string)
	DeleteDashboardQuery(projectID uint64, queryID uint64) (int, string)
	UpdateSavedQuery(projectID uint64, queryID uint64, query *model.Queries) (*model.Queries, int)
	UpdateQueryIDsWithNewIDs(projectID uint64, shareableURLs []string) int
	SearchQueriesWithProjectId(projectID uint64, searchString string) ([]model.Queries, int)
	GetAllNonConvertedQueries(projectID uint64) ([]model.Queries, int)

	// salesforce_document
	GetSalesforceSyncInfo() (model.SalesforceSyncInfo, int)
	GetSalesforceObjectPropertiesName(ProjectID uint64, objectType string) ([]string, []string)
	GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(projectID uint64, customerUserID, userID string, docType int) (*model.SalesforceDocument, int)
	UpdateSalesforceDocumentBySyncStatus(projectID uint64, document *model.SalesforceDocument, syncID, userID, groupUserID string, synced bool) int
	BuildAndUpsertDocument(projectID uint64, objectName string, value model.SalesforceRecord) error
	BuildAndUpsertDocumentInBatch(projectID uint64, objectName string, values []model.SalesforceRecord) error
	CreateSalesforceDocument(projectID uint64, document *model.SalesforceDocument) int
	CreateSalesforceDocumentByAction(projectID uint64, document *model.SalesforceDocument, action model.SalesforceAction) int
	GetSyncedSalesforceDocumentByType(projectID uint64, ids []string, docType int, includeUnSynced bool) ([]model.SalesforceDocument, int)
	GetSalesforceObjectValuesByPropertyName(ProjectID uint64, objectType string, propertyName string) []interface{}
	GetSalesforceDocumentsByTypeForSync(projectID uint64, typ int, from, to int64) ([]model.SalesforceDocument, int)
	GetLatestSalesforceDocumentByID(projectID uint64, documentIDs []string, docType int, maxTimestamp int64) ([]model.SalesforceDocument, int)
	GetSalesforceDocumentBeginingTimestampByDocumentTypeForSync(projectID uint64) (map[int]int64, int64, int)

	// scheduled_task
	CreateScheduledTask(task *model.ScheduledTask) int
	UpdateScheduledTask(taskID string, taskDetails *postgres.Jsonb, endTime int64, status model.ScheduledTaskStatus) (int64, int)
	GetScheduledTaskByID(taskID string) (*model.ScheduledTask, int)
	GetScheduledTaskInProgressCount(projectID uint64, taskType model.ScheduledTaskType) (int64, int)
	GetScheduledTaskLastRunTimestamp(projectID uint64, taskType model.ScheduledTaskType) (int64, int)
	GetNewArchivalFileNamesAndEndTimeForProject(projectID uint64,
		lastRunAt int64, hardStartTime, hardEndTime time.Time) (map[int64]map[string]interface{}, int)
	GetArchivalFileNamesForProject(projectID uint64, startTime, endTime time.Time) ([]string, []string, int)
	FailScheduleTask(taskID string)
	GetCompletedArchivalBatches(projectID uint64, startTime, endTime time.Time) (map[int64]int64, int)

	// tracked_events
	CreateFactorsTrackedEvent(ProjectID uint64, EventName string, agentUUID string) (int64, int)
	DeactivateFactorsTrackedEvent(ID int64, ProjectID uint64) (int64, int)
	GetAllFactorsTrackedEventsByProject(ProjectID uint64) ([]model.FactorsTrackedEventInfo, int)
	GetAllActiveFactorsTrackedEventsByProject(ProjectID uint64) ([]model.FactorsTrackedEventInfo, int)
	GetFactorsTrackedEvent(EventNameID string, ProjectID uint64) (*model.FactorsTrackedEvent, error)
	GetFactorsTrackedEventByID(ID int64, ProjectID uint64) (*model.FactorsTrackedEvent, error)

	// tracked_user_properties
	CreateFactorsTrackedUserProperty(ProjectID uint64, UserPropertyName string, agentUUID string) (int64, int)
	RemoveFactorsTrackedUserProperty(ID int64, ProjectID uint64) (int64, int)
	GetAllFactorsTrackedUserPropertiesByProject(ProjectID uint64) ([]model.FactorsTrackedUserProperty, int)
	GetAllActiveFactorsTrackedUserPropertiesByProject(ProjectID uint64) ([]model.FactorsTrackedUserProperty, int)
	GetFactorsTrackedUserProperty(UserPropertyName string, ProjectID uint64) (*model.FactorsTrackedUserProperty, error)
	GetFactorsTrackedUserPropertyByID(ID int64, ProjectID uint64) (*model.FactorsTrackedUserProperty, error)
	IsUserPropertyValid(ProjectID uint64, UserProperty string) bool

	// user
	CreateUser(user *model.User) (string, int)
	CreateOrGetAMPUser(projectID uint64, ampUserId string, timestamp int64, requestSource int) (string, int)
	CreateOrGetSegmentUser(projectID uint64, segAnonId, custUserId string, requestTimestamp int64, requestSource int) (*model.User, int)
	GetUserPropertiesByUserID(projectID uint64, id string) (*postgres.Jsonb, int)
	GetUser(projectID uint64, id string) (*model.User, int)
	GetUserIDByAMPUserID(projectId uint64, ampUserId string) (string, int)
	IsUserExistByID(projectID uint64, id string) int
	GetUsers(projectID uint64, offset uint64, limit uint64) ([]model.User, int)
	GetUsersByCustomerUserID(projectID uint64, customerUserID string) ([]model.User, int)
	GetUserLatestByCustomerUserId(projectID uint64, customerUserId string, requestSource int) (*model.User, int)
	GetExistingCustomerUserID(projectID uint64, arrayCustomerUserID []string) (map[string]string, int)
	GetUserBySegmentAnonymousId(projectID uint64, segAnonId string) (*model.User, int)
	GetAllUserIDByCustomerUserID(projectID uint64, customerUserID string) ([]string, int)
	GetRecentUserPropertyKeysWithLimits(projectID uint64, usersLimit int, propertyLimit int, seedDate time.Time) ([]U.Property, error)
	GetRecentUserPropertyValuesWithLimits(projectID uint64, propertyKey string, usersLimit, valuesLimit int, seedDate time.Time) ([]U.PropertyValue, string, error)
	GetUserPropertiesByProject(projectID uint64, limit int, lastNDays int) (map[string][]string, error)
	GetPropertyValuesByUserProperty(projectID uint64, propertyName string, limit int, lastNDays int) ([]string, error)
	GetLatestUserPropertiesOfUserAsMap(projectID uint64, id string) (*map[string]interface{}, int)
	GetDistinctCustomerUserIDSForProject(projectID uint64) ([]string, int)
	GetUserIdentificationPhoneNumber(projectID uint64, phoneNo string) (string, string)
	UpdateUser(projectID uint64, id string, user *model.User, updateTimestamp int64) (*model.User, int)
	UpdateUserProperties(projectId uint64, id string, properties *postgres.Jsonb, updateTimestamp int64) (*postgres.Jsonb, int)
	UpdateUserPropertiesV2(projectID uint64, id string, newProperties *postgres.Jsonb, newUpdateTimestamp int64, sourceValue string, objectType string) (*postgres.Jsonb, int)
	OverwriteUserPropertiesByID(projectID uint64, id string, properties *postgres.Jsonb, withUpdateTimestamp bool, updateTimestamp int64, source string) int
	OverwriteUserPropertiesByCustomerUserID(projectID uint64, customerUserID string, properties *postgres.Jsonb, updateTimestamp int64) int
	GetUserByPropertyKey(projectID uint64, key string, value interface{}) (*model.User, int)
	UpdateUserPropertiesForSession(projectID uint64, sessionUserPropertiesRecordMap *map[string]model.SessionUserProperties) int
	GetCustomerUserIDAndUserPropertiesFromFormSubmit(projectID uint64, userID string, formSubmitProperties *U.PropertiesMap) (string, *U.PropertiesMap, int)
	UpdateIdentifyOverwriteUserPropertiesMeta(projectID uint64, customerUserID, userID, pageURL, source string, userProperties *postgres.Jsonb, timestamp int64, isNewUser bool) error
	GetSelectedUsersByCustomerUserID(projectID uint64, customerUserID string, limit uint64, numUsers uint64) ([]model.User, int)
	CreateGroupUser(user *model.User, groupName, groupID string) (string, int)
	UpdateUserGroup(projectID uint64, userID, groupName, groupID, groupUserID string) (*model.User, int)
	UpdateUserGroupProperties(projectID uint64, userID string, newProperties *postgres.Jsonb, updateTimestamp int64) (*postgres.Jsonb, int)
	GetPropertiesUpdatedTimestampOfUser(projectId uint64, id string) (int64, int)

	// web_analytics
	GetWebAnalyticsQueriesFromDashboardUnits(projectID uint64) (uint64, *model.WebAnalyticsQueries, int)
	CreateWebAnalyticsDefaultDashboardWithUnits(projectID uint64, agentUUID string) int
	ExecuteWebAnalyticsQueries(projectID uint64, queries *model.WebAnalyticsQueries) (queryResult *model.WebAnalyticsQueryResult, errCode int)
	CacheWebsiteAnalyticsForProjects(stringProjectsIDs, excludeProjectIDs string, numRoutines int, reportCollector *sync.Map)
	GetWebAnalyticsEnabledProjectIDsFromList(stringProjectIDs, excludeProjectIDs string) []uint64
	GetWebAnalyticsCachePayloadsForProject(projectID uint64) ([]model.WebAnalyticsCachePayload, int, string)
	CacheWebsiteAnalyticsForDateRange(cachePayload model.WebAnalyticsCachePayload) (int, model.CachingUnitReport)
	CacheWebsiteAnalyticsForMonthlyRange(projectIDs, excludeProjectIDs string, numMonths, numRoutines int, reportCollector *sync.Map)

	// journey_mining
	GetWeightedJourneyMatrix(projectID uint64, journeyEvents []model.QueryEventWithProperties,
		goalEvents []model.QueryEventWithProperties, startTime, endTime, lookbackDays int64, eventFiles,
		userFiles string, includeSession bool, sessionProperty string, cloudManager filestore.FileManager)

	// smart_properties
	GetSmartPropertyRulesConfig(projectID uint64, objectType string) (model.SmartPropertyRulesConfig, int)
	CreateSmartPropertyRules(projectID uint64, smartProperty *model.SmartPropertyRules) (*model.SmartPropertyRules, string, int)
	GetSmartPropertyRules(projectID uint64) ([]model.SmartPropertyRules, int)
	GetAllChangedSmartPropertyRulesForProject(projectID uint64) ([]model.SmartPropertyRules, int)
	GetSmartPropertyRule(projectID uint64, ruleID string) (model.SmartPropertyRules, int)
	DeleteSmartPropertyRules(projectID uint64, ruleID string) int
	UpdateSmartPropertyRules(projectID uint64, ruleID string, smartPropertyDoc model.SmartPropertyRules) (model.SmartPropertyRules, string, int)
	GetProjectIDsHavingSmartPropertyRules() ([]uint64, int)
	GetLatestMetaForAdwordsForGivenDays(projectID uint64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields)
	GetLatestMetaForFacebookForGivenDays(projectID uint64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields)
	GetLatestMetaForLinkedinForGivenDays(projectID uint64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields)
	GetLatestMetaForBingAdsForGivenDays(projectID uint64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields)
	BuildAndCreateSmartPropertyFromChannelDocumentAndRule(smartPropertyRule *model.SmartPropertyRules, rule model.Rule,
		channelDocument model.ChannelDocumentsWithFields, source string) int
	DeleteSmartPropertyByRuleID(projectID uint64, ruleID string) (int, int, int)
	GetSmartPropertyByProjectIDAndObjectIDAndObjectType(projectID uint64, objectID string, objectType int) (model.SmartProperties, int)
	GetSmartPropertyByProjectIDAndSourceAndObjectType(projectID uint64, source string, objectType int) ([]model.SmartProperties, int)
	DeleteSmartPropertyByProjectIDAndSourceAndObjectID(projectID uint64, source string, objectID string) int

	//properties_type
	GetPropertyTypeByKeyValue(projectID uint64, eventName string, propertyKey string, propertyValue interface{}, isUserProperty bool) string
	GetPropertyTypeFromDB(projectID uint64, eventName, propertyKey string, isUserProperty bool) (int, *model.PropertyDetail)

	// project_analytics
	GetEventUserCountsOfAllProjects(lastNDays int) (map[string][]*model.ProjectAnalytics, error)

	// Property details
	CreatePropertyDetails(projectID uint64, eventName, propertyKey, propertyType string, isUserProperty bool, allowOverWrite bool) int
	CreateOrDeletePropertyDetails(projectID uint64, eventName, enKey, pType string, isUserProperty, allowOverWrite bool) error
	GetAllPropertyDetailsByProjectID(projectID uint64, eventName string, isUserProperty bool) (*map[string]string, int)

	// display names
	CreateOrUpdateDisplayNameByObjectType(projectID uint64, propertyName, objectType, displayName, group string) int
	GetDisplayNamesForAllEvents(projectID uint64) (int, map[string]string)
	GetDisplayNamesForAllEventProperties(projectID uint64, eventName string) (int, map[string]string)
	GetDisplayNamesForAllUserProperties(projectID uint64) (int, map[string]string)
	GetDisplayNamesForObjectEntities(projectID uint64) (int, map[string]string)
	CreateOrUpdateDisplayName(projectID uint64, eventName, propertyName, displayName, tag string) int

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
	GetAllProcessedIntervals(taskID uint64, projectId uint64, lookbackInDays int, endDate *time.Time) ([]uint64, int, string)
	InsertTaskBeginRecord(taskId uint64, projectId uint64, delta uint64) (int, string)
	InsertTaskEndRecord(taskId uint64, projectId uint64, delta uint64) (int, string)
	GetAllToBeExecutedDeltas(taskId uint64, projectId uint64, lookbackInDays int, endDate *time.Time) ([]uint64, int, string)
	GetAllInProgressIntervals(taskID uint64, projectId uint64, lookbackInDays int, endDate *time.Time) ([]uint64, int, string)
	IsDependentTaskDone(taskId uint64, projectId uint64, delta uint64) bool
	DeleteTaskEndRecord(taskId uint64, projectId uint64, delta uint64) (int, string)
	GetAllProcessedIntervalsFromStartDate(taskID uint64, projectId uint64, startDate *time.Time) ([]uint64, int, string)

	// project model metadata
	CreateProjectModelMetadata(pmm *model.ProjectModelMetadata) (int, string)
	GetProjectModelMetadata(projectId uint64) ([]model.ProjectModelMetadata, int, string)
	GetAllProjectModelMetadata() ([]model.ProjectModelMetadata, int, string)

	// search console
	GetGoogleOrganicLastSyncInfoForProject(projectID uint64) ([]model.GoogleOrganicLastSyncInfo, int)
	GetAllGoogleOrganicLastSyncInfoForAllProjects() ([]model.GoogleOrganicLastSyncInfo, int)
	CreateGoogleOrganicDocument(gscDoc *model.GoogleOrganicDocument) int
	CreateMultipleGoogleOrganicDocument(gscDocuments []model.GoogleOrganicDocument) int

	// monitoring
	MonitorSlowQueries() ([]interface{}, []interface{}, error)
	CollectTableSizes() map[string]string

	// weekly insights
	CreateWeeklyInsightsMetadata(pmm *model.WeeklyInsightsMetadata) (int, string)
	GetWeeklyInsightsMetadata(projectId uint64) ([]model.WeeklyInsightsMetadata, int, string)

	// feedback
	PostFeedback(ProjectID uint64, agentUUID string, Feature string, Property *postgres.Jsonb, VoteType int) (int, string)
	GetRecordsFromFeedback(projectID uint64, agentUUID string) ([]model.Feedback, error)

	//Group
	CreateGroup(projectID uint64, groupName string, allowedGroupNames map[string]bool) (*model.Group, int)
	GetGroup(projectID uint64, groupName string) (*model.Group, int)
	CreateOrUpdateGroupPropertiesBySource(projectID uint64, groupName string, groupID, groupUserID string,
		enProperties *map[string]interface{}, createdTimestamp, updatedTimestamp int64, source string) (string, error)
	GetGroups(projectID uint64) ([]model.Group, int)
	GetPropertiesByGroup(projectID uint64, groupName string, limit int, lastNDays int) (map[string][]string, int)
	GetPropertyValuesByGroupProperty(projectID uint64, groupName string, propertyName string, limit int, lastNDays int) ([]string, error)

	// Delete channel Integrations
	DeleteChannelIntegration(projectID uint64, channelName string) (int, error)

	//group_relationship
	CreateGroupRelationship(projectID uint64, leftGroupName, leftGroupUserID, rightGroupName, rightGroupUserID string) (*model.GroupRelationship, int)
	GetGroupRelationshipByUserID(projectID uint64, leftGroupUserID string) ([]model.GroupRelationship, int)

	//Content-groups
	GetAllContentGroups(projectID uint64) ([]model.ContentGroup, int)
	GetContentGroupById(id string, projectID uint64) (model.ContentGroup, int)
	CreateContentGroup(projectID uint64, contentGroup model.ContentGroup) (model.ContentGroup, int, string)
	DeleteContentGroup(id string, projectID uint64) (int, string)
	UpdateContentGroup(id string, projectID uint64, contentGroup model.ContentGroup) (model.ContentGroup, int, string)
	CheckURLContentGroupValue(pageUrl string, projectID uint64) map[string]string

	// fivetran mappings
	DisableFiveTranMapping(ProjectID uint64, Integration string, ConnectorId string) error
	EnableFiveTranMapping(ProjectID uint64, Integration string, ConnectorId string, Accounts string) error
	GetFiveTranMapping(ProjectID uint64, Integration string) (string, error)
	GetActiveFiveTranMapping(ProjectID uint64, Integration string) (model.FivetranMappings, error)
	GetAllActiveFiveTranMapping(ProjectID uint64, Integration string) ([]string, error)
	GetLatestFiveTranMapping(ProjectID uint64, Integration string) (string, string, error)
	PostFiveTranMapping(ProjectID uint64, Integration string, ConnectorId string, SchemaId string, Accounts string) error
	GetAllActiveFiveTranMappingByIntegration(Integration string) ([]model.FivetranMappings, error)
	UpdateFiveTranMappingAccount(ProjectID uint64, Integration string, ConnectorId string, Accounts string) error

	//leadgen
	GetLeadgenSettingsForProject(projectID uint64) ([]model.LeadgenSettings, error)
	GetLeadgenSettings() ([]model.LeadgenSettings, error)
	UpdateRowRead(projectID uint64, source int, rowRead int64) (int, error)

	// integration document
	InsertIntegrationDocument(doc model.IntegrationDocument) error
	UpsertIntegrationDocument(doc model.IntegrationDocument) error

	// alerts
	SetAuthTokenforSlackIntegration(projectID uint64, agentUUID string, authTokens model.SlackAccessTokens) error
	GetSlackAuthToken(projectID uint64, agentUUID string) (model.SlackAccessTokens, error)
	DeleteSlackIntegration(projectID uint64, agentUUID string) error
	GetAlertById(id string, projectID uint64) (model.Alert, int)
	GetAllAlerts(projectID uint64) ([]model.Alert, int)
	DeleteAlert(id string, projectID uint64) (int, string)
	CreateAlert(projectID uint64, alert model.Alert) (model.Alert, int, string)
	GetAlertNamesByProjectIdTypeAndName(projectID uint64, nameOfQuery string) ([]string, int)
	UpdateAlert(lastAlertSent bool) (int, string)

	// sharable url
	CreateShareableURL(sharableURLParams *model.ShareableURL) (*model.ShareableURL, int)
	GetAllShareableURLsWithProjectIDAndAgentID(projectID uint64, agentUUID string) ([]*model.ShareableURL, int)
	GetShareableURLWithShareStringAndAgentID(projectID uint64, shareId, agentUUID string) (*model.ShareableURL, int)
	GetShareableURLWithShareStringWithLargestScope(projectID uint64, shareId string, entityType int) (*model.ShareableURL, int)
	// GetShareableURLWithID(projectID uint64, shareId string) (*model.ShareableURL, int)
	// UpdateShareableURLShareTypeWithShareIDandCreatedBy(projectID uint64, shareId, createdBy string, shareType int, allowedUsers string) int
	DeleteShareableURLWithShareIDandAgentID(projectID uint64, shareId, createdBy string) int
	DeleteShareableURLWithEntityIDandType(projectID, entityID uint64, entityType int) int
	RevokeShareableURLsWithShareString(projectId uint64, shareString string) (int, string)
	RevokeShareableURLsWithProjectID(projectId uint64) (int, string)

	CreateSharableURLAudit(sharableURL *model.ShareableURL, agentId string) int

	//crm
	CreateCRMUser(crmUser *model.CRMUser) (int, error)
	CreateCRMGroup(crmGroup *model.CRMGroup) (int, error)
	CreateCRMActivity(crmActivity *model.CRMActivity) (int, error)
	CreateCRMRelationship(crmRelationship *model.CRMRelationship) (int, error)
	CreateCRMProperties(crmProperty *model.CRMProperty) (int, error)
	GetCRMUserByTypeAndAction(projectID uint64, source U.CRMSource, id string, userType int, action model.CRMAction) (*model.CRMUser, int)
	UpdateCRMUserAsSynced(projectID uint64, source U.CRMSource, crmUser *model.CRMUser, userID, syncID string) (*model.CRMUser, int)
	GetCRMUsersInOrderForSync(projectID uint64, source U.CRMSource, startTimestamp, endTimestamp int64) ([]model.CRMUser, int)
	GetCRMActivityInOrderForSync(projectID uint64, source U.CRMSource, startTimestamp, endTimestamp int64) ([]model.CRMActivity, int)
	GetCRMActivityMinimumTimestampForSync(projectID uint64, source U.CRMSource) (int64, int)
	GetCRMUsersMinimumTimestampForSync(projectID uint64, source U.CRMSource) (int64, int)
	GetCRMPropertiesForSync(projectID uint64) ([]model.CRMProperty, int)
	GetActivitiesDistinctEventNamesByType(projectID uint64, objectTypes []int) (map[int][]string, int)
	UpdateCRMProperyAsSynced(projectID uint64, source U.CRMSource, crmProperty *model.CRMProperty) (*model.CRMProperty, int)
	UpdateCRMActivityAsSynced(projectID uint64, source U.CRMSource, crmActivity *model.CRMActivity, syncID, userID string) (*model.CRMActivity, int)

	GetCRMSetting(projectID uint64) (*model.CRMSetting, int)
	GetAllCRMSetting() ([]model.CRMSetting, int)
	UpdateCRMSetting(projectID uint64, option model.CRMSettingOption) int
	CreateOrUpdateCRMSetting(projectID uint64, crmSetting *model.CRMSetting) int

	// Timeline
	GetProfileUsersListByProjectId(projectID uint64) ([]model.Contact, int)
	GetProfileUserDetailsByID(projectID uint64, identity string, isAnonymous string) (*model.ContactDetails, int)
}
