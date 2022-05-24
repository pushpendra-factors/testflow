package handler

import (
	C "factors/config"
	IH "factors/handler/internal"
	V1 "factors/handler/v1"
	mid "factors/middleware"
	"factors/model/model"
	U "factors/util"
	"net/http"

	slack "factors/slack_bot/handler"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const ROUTE_SDK_ROOT = "/sdk"
const ROUTE_SDK_AMP_ROOT = "/sdk/amp"
const ROUTE_PROJECTS_ROOT = "/projects"
const ROUTE_PROJECTS_ROOT_V1 = "v1/projects"
const ROUTE_INTEGRATIONS_ROOT = "/integrations"
const ROUTE_DATA_SERVICE_ROOT = "/data_service"
const ROUTE_SDK_ADWORDS_ROOT = "/adwords_sdk_service"
const ROUTE_VERSION_V1 = "/v1"

func InitExternalAuth(r *gin.Engine, auth *Authenticator) {
	routePrefix := C.GetRoutesURLPrefix() + "/oauth"
	r.Use(mid.BlockMaliciousPayload())
	r.GET(routePrefix+"/signup", ExternalAuthentication(auth, SIGNUP_FLOW))
	r.GET(routePrefix+"/login", ExternalAuthentication(auth, SIGNIN_FLOW))
	r.GET(routePrefix+"/activate", ExternalAuthentication(auth, ACTIVATE_FLOW))
	r.GET(routePrefix+"/callback", CallbackHandler(auth))
}

func InitAppRoutes(r *gin.Engine) {
	routePrefix := C.GetRoutesURLPrefix()

	r.GET(routePrefix+"/status", func(c *gin.Context) {
		resp := map[string]string{
			"status": "success",
		}
		c.JSON(http.StatusOK, resp)
		return
	})

	// Initialize swagger api docs only for development / staging.
	if C.GetConfig().Env != C.PRODUCTION {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// NOTE: Always keep BlockMaliciousPayload middlware on top of the chain.
	r.Use(mid.BlockMaliciousPayload())

	r.Use(mid.SkipAPIWritesIfDisabled())
	r.GET(routePrefix+"/health", mid.MonitoringAPIMiddleware(), Monitoring)
	r.POST(routePrefix+"/accounts/signup", SignUp)
	r.POST(routePrefix+"/agents/signin", Signin)
	r.GET(routePrefix+"/agents/signout", Signout)
	r.POST(routePrefix+"/agents/forgotpassword", AgentGenerateResetPasswordLinkEmail)
	r.POST(routePrefix+"/agents/setpassword", mid.ValidateAgentSetPasswordRequest(), AgentSetPassword)
	r.PUT(routePrefix+"/agents/updatepassword", mid.SetLoggedInAgent(), UpdateAgentPassword)
	r.POST(routePrefix+"/agents/activate", mid.ValidateAgentActivationRequest(), AgentActivate)
	r.GET(routePrefix+"/agents/billing", mid.SetLoggedInAgent(), GetAgentBillingAccount)
	r.PUT(routePrefix+"/agents/billing", mid.SetLoggedInAgent(), UpdateAgentBillingAccount)
	r.GET(routePrefix+"/agents/info", mid.SetLoggedInAgent(), AgentInfo)
	r.PUT(routePrefix+"/agents/info", mid.SetLoggedInAgent(), UpdateAgentInfo)
	r.GET(routePrefix+"/projectanalytics", mid.SetLoggedInAgentInternalOnly(), V1.GetFactorsAnalyticsHandler)
	r.POST(routePrefix+"/registertask", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.RegisterTaskHandler))
	r.POST(routePrefix+"/registertaskdependency", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.RegisterTaskDependencyHandler))
	r.GET(routePrefix+"/GetAllProcessedIntervals", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.GetAllProcessedIntervalsHandler))
	r.DELETE(routePrefix+"/DeleteTaskEndRecord", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.DeleteTaskEndRecordHandler))
	r.POST(routePrefix+ROUTE_PROJECTS_ROOT, mid.SetLoggedInAgent(), CreateProjectHandler)
	r.GET(routePrefix+ROUTE_PROJECTS_ROOT,
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		GetProjectsHandler)
	r.GET(routePrefix+"/GetTaskDetailsByName", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.GetTaskDetailsByNameHandler))
	r.GET(routePrefix+"/GetAllToBeExecutedDeltas", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.GetAllToBeExecutedDeltasHandler))
	r.GET(routePrefix+"/IsDependentTaskDone", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.IsDependentTaskDoneHandler))
	r.POST(routePrefix+"/InsertTaskBeginRecord", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.InsertTaskBeginRecordHandler))
	r.POST(routePrefix+"/InsertTaskEndRecord", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.InsertTaskEndRecordHandler))
	r.GET("/hubspot/getcontact", V1.GetHubspotContactByEmail)

	// Shareable link routes
	shareRouteGroup := r.Group(routePrefix + ROUTE_PROJECTS_ROOT)
	shareRouteGroup.Use(mid.SkipDemoProjectWriteAccess())
	shareRouteGroup.Use(mid.ValidateAccessToSharedEntity(model.ShareableURLEntityTypeQuery))

	shareRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/query", responseWrapper(EventsQueryHandler))
	shareRouteGroup.POST("/:project_id/query", responseWrapper(QueryHandler))
	shareRouteGroup.POST("/:project_id/attribution/query", responseWrapper(AttributionHandler))
	shareRouteGroup.POST("/:project_id/profiles/query", responseWrapper(ProfilesQueryHandler))
	shareRouteGroup.POST("/:project_id/channels/query", ChannelQueryHandler)
	shareRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/channels/query", responseWrapper(V1.ExecuteChannelQueryHandler))
	shareRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/kpi/query", responseWrapper(V1.ExecuteKPIQueryHandler))

	// Auth route group with authentication an authorization middleware.
	authRouteGroup := r.Group(routePrefix + ROUTE_PROJECTS_ROOT)
	authRouteGroup.Use(mid.SetLoggedInAgent())
	authRouteGroup.Use(mid.SetAuthorizedProjectsByLoggedInAgent())
	authRouteGroup.Use(mid.ValidateLoggedInAgentHasAccessToRequestProject())
	authRouteGroup.Use(mid.SkipDemoProjectWriteAccess())

	authRouteGroup.PUT("/:project_id", EditProjectHandler)

	// Dashboard endpoints
	authRouteGroup.GET("/:project_id/dashboards", GetDashboardsHandler)
	authRouteGroup.POST("/:project_id/dashboards", CreateDashboardHandler)
	authRouteGroup.PUT("/:project_id/dashboards/:dashboard_id", UpdateDashboardHandler)
	authRouteGroup.GET("/:project_id/dashboards/:dashboard_id/units", GetDashboardUnitsHandler)
	authRouteGroup.POST("/:project_id/dashboards/:dashboard_id/units", CreateDashboardUnitHandler)
	authRouteGroup.PUT("/:project_id/dashboards/:dashboard_id/units/:unit_id", UpdateDashboardUnitHandler)
	authRouteGroup.DELETE("/:project_id/dashboards/:dashboard_id/units/:unit_id", DeleteDashboardUnitHandler)
	authRouteGroup.POST("/:project_id/dashboard/:dashboard_id/units/query/web_analytics",
		DashboardUnitsWebAnalyticsQueryHandler)

	authRouteGroup.GET("/:project_id/event_names", GetEventNamesHandler)
	authRouteGroup.GET("/:project_id/user/event_names", GetEventNamesByUserHandler)
	authRouteGroup.GET(":project_id/groups/:group_name/event_names", GetEventNamesByGroupHandler)
	authRouteGroup.GET("/:project_id/queries", GetQueriesHandler)
	authRouteGroup.POST("/:project_id/queries", CreateQueryHandler)
	authRouteGroup.PUT("/:project_id/queries/:query_id", UpdateSavedQueryHandler)
	authRouteGroup.DELETE("/:project_id/queries/:query_id", DeleteSavedQueryHandler)
	authRouteGroup.GET("/:project_id/queries/search", SearchQueriesHandler)
	authRouteGroup.GET("/:project_id/models", GetProjectModelsHandler)
	authRouteGroup.GET("/:project_id/filters", GetFiltersHandler)
	authRouteGroup.POST("/:project_id/filters", CreateFilterHandler)
	authRouteGroup.PUT("/:project_id/filters/:filter_id", UpdateFilterHandler)
	authRouteGroup.DELETE("/:project_id/filters/:filter_id", DeleteFilterHandler)
	authRouteGroup.GET("/:project_id/event_names/:event_name/properties", GetEventPropertiesHandler)
	authRouteGroup.GET("/:project_id/event_names/:event_name/properties/:property_name/values", GetEventPropertyValuesHandler)
	authRouteGroup.GET("/:project_id/groups", GetGroupsHandler)
	authRouteGroup.GET("/:project_id/groups/:group_name/properties", GetGroupPropertiesHandler)
	authRouteGroup.GET("/:project_id/groups/:group_name/properties/:property_name/values", GetGroupPropertyValuesHandler)
	authRouteGroup.GET("/:project_id/users", GetUsersHandler)
	authRouteGroup.GET("/:project_id/users/:user_id", GetUserHandler)
	authRouteGroup.GET("/:project_id/user_properties", GetUserPropertiesHandler)
	authRouteGroup.GET("/:project_id/user_properties/:property_name/values", GetUserPropertyValuesHandler)
	authRouteGroup.GET("/:project_id/channel_grouping_properties", GetChannelGroupingPropertiesHandler)
	authRouteGroup.POST("/:project_id/factor", FactorHandler)

	// Moved to shareable routes
	// authRouteGroup.POST("/:project_id/query", responseWrapper(QueryHandler))
	// authRouteGroup.POST("/:project_id/profiles/query", responseWrapper(ProfilesQueryHandler))
	// authRouteGroup.POST("/:project_id/channels/query", ChannelQueryHandler)
	// authRouteGroup.POST("/:project_id/attribution/query", responseWrapper(AttributionHandler))
	// v1 API endpoints
	// authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/query", responseWrapper(EventsQueryHandler))
	// authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/channels/query", responseWrapper(V1.ExecuteChannelQueryHandler))
	// authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/kpi/query", responseWrapper(V1.ExecuteKPIQueryHandler))

	authRouteGroup.GET("/:project_id/channels/filter_values", GetChannelFilterValuesHandler)

	// shareable url endpoints
	authRouteGroup.GET("/:project_id/shareable_url", GetShareableURLsHandler)
	authRouteGroup.POST("/:project_id/shareable_url", CreateShareableURLHandler)
	authRouteGroup.DELETE("/:project_id/shareable_url/:share_id", DeleteShareableURLHandler)
	authRouteGroup.DELETE("/:project_id/shareable_url/revoke/:query_id", RevokeShareableURLHandler)

	// v1 Dashboard endpoints
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/dashboards/multi/:dashboard_ids/units", CreateDashboardUnitForMultiDashboardsHandler)
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/dashboards/queries/:dashboard_id/units", CreateDashboardUnitsForMultipleQueriesHandler)
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/dashboards/:dashboard_id/units/multi/:unit_ids", DeleteMultiDashboardUnitHandler)
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/dashboards/:dashboard_id", DeleteDashboardHandler)

	// v1 Channel endpoints
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/channels/config", V1.GetChannelConfigHandler)
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/channels/filter_values", V1.GetChannelFilterValuesHandler)

	// v1 KPI endpoints
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/kpi/config", responseWrapper(V1.GetKPIConfigHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/kpi/filter_values", responseWrapper(V1.GetKPIFilterValuesHandler))

	// v1 custom metrics - admin/settings side.
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics/config", V1.GetCustomMetricsConfig)
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics", responseWrapper(V1.CreateCustomMetric))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics", responseWrapper(V1.GetCustomMetrics))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics/:id", responseWrapper(V1.DeleteCustomMetrics))

	// v1 CRM And Smart Event endpoints
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_event", GetSmartEventFiltersHandler)
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/smart_event", CreateSmartEventFilterHandler)
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/smart_event", UpdateSmartEventFilterHandler)
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/smart_event", responseWrapper(DeleteSmartEventFilterHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/crm/:crm_source/:object_type/properties", GetCRMObjectPropertiesHandler)
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/crm/:crm_source/:object_type/properties/:property_name/values", GetCRMObjectValuesByPropertyNameHandler)
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/templates/:type/config", responseWrapper(V1.GetTemplateConfigHandler))
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/templates/:type/config", responseWrapper(V1.UpdateTemplateConfigHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/templates/:type/query", responseWrapper(V1.ExecuteTemplateQueryHandler))

	// smart Properties
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/config/:object_type", responseWrapper(GetSmartPropertyRulesConfigHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules", responseWrapper(CreateSmartPropertyRulesHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules", responseWrapper(GetSmartPropertyRulesHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules/:rule_id", responseWrapper(GetSmartPropertyRuleByRuleIDHandler))
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules/:rule_id", responseWrapper(UpdateSmartPropertyRulesHandler))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules/:rule_id", responseWrapper(DeleteSmartPropertyRulesHandler))

	// content groups
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/contentgroup", responseWrapper(V1.CreateContentGroupHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/contentgroup", responseWrapper(V1.GetContentGroupHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/contentgroup/:id", responseWrapper(V1.GetContentGroupByIDHandler))
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/contentgroup/:id", responseWrapper(V1.UpdateContentGroupHandler))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/contentgroup/:id", responseWrapper(V1.DeleteContentGroupHandler))
	// TODO
	// Scope this with Project Admin
	authRouteGroup.GET("/:project_id/agents", GetProjectAgentsHandler)
	authRouteGroup.POST("/:project_id/agents/invite", AgentInvite)
	authRouteGroup.POST("/:project_id/agents/batchinvite", AgentInviteBatch)
	authRouteGroup.PUT("/:project_id/agents/remove", RemoveProjectAgent)
	authRouteGroup.PUT("/:project_id/agents/update", AgentUpdate)
	authRouteGroup.GET("/:project_id/settings", GetProjectSettingHandler)
	authRouteGroup.GET("/:project_id/v1/settings", V1.GetProjectSettingHandler)
	authRouteGroup.PUT("/:project_id/settings", UpdateProjectSettingsHandler)

	// V1 Routes
	authRouteGroup.GET("/:project_id/v1/event_names", V1.GetEventNamesHandler)
	authRouteGroup.GET("/:project_id/v1/event_names/:type", V1.GetEventNamesByTypeHandler)
	authRouteGroup.GET("/:project_id/v1/agents", V1.GetProjectAgentsHandler)
	r.GET(routePrefix+"/"+ROUTE_PROJECTS_ROOT_V1,
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		V1.GetProjectsHandler)
	r.GET(routePrefix+"/v1/demoprojects",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		V1.GetDemoProjects)

	// Tracked Events
	authRouteGroup.POST("/:project_id/v1/factors/tracked_event", V1.CreateFactorsTrackedEventsHandler)
	authRouteGroup.DELETE("/:project_id/v1/factors/tracked_event/remove", V1.RemoveFactorsTrackedEventsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/tracked_event", V1.GetAllFactorsTrackedEventsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/grouped_tracked_event", V1.GetAllGroupedFactorsTrackedEventsHandler)

	// Tracked User Property
	authRouteGroup.POST("/:project_id/v1/factors/tracked_user_property", V1.CreateFactorsTrackedUserPropertyHandler)
	authRouteGroup.DELETE("/:project_id/v1/factors/tracked_user_property/remove", V1.RemoveFactorsTrackedUserPropertyHandler)
	authRouteGroup.GET("/:project_id/v1/factors/tracked_user_property", V1.GetAllFactorsTrackedUserPropertiesHandler)

	// Goals
	authRouteGroup.POST("/:project_id/v1/factors/goals", V1.CreateFactorsGoalsHandler)
	authRouteGroup.DELETE("/:project_id/v1/factors/goals/remove", V1.RemoveFactorsGoalsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/goals", V1.GetAllFactorsGoalsHandler)
	authRouteGroup.PUT("/:project_id/v1/factors/goals/update", V1.UpdateFactorsGoalsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/goals/search", V1.SearchFactorsGoalHandler)
	authRouteGroup.POST("/:project_id/v1/factor", V1.PostFactorsHandler)
	authRouteGroup.POST("/:project_id/v1/factor/compare", V1.PostFactorsCompareHandler)
	authRouteGroup.POST("/:project_id/v1/events/displayname", responseWrapper(V1.CreateDisplayNamesHandler))
	authRouteGroup.GET("/:project_id/v1/factor", V1.GetFactorsHandler)
	authRouteGroup.GET("/:project_id/v1/factor/model_metadata", V1.GetModelMetaData)

	authRouteGroup.GET("/:project_id/insights", responseWrapper(V1.GetWeeklyInsightsHandler))
	authRouteGroup.GET("/:project_id/weekly_insights_metadata", responseWrapper(V1.GetWeeklyInsightsMetadata))
	authRouteGroup.POST("/:project_id/feedback", V1.PostFeedbackHandler)

	// bingads integration
	authRouteGroup.POST("/:project_id/v1/bingads", responseWrapper(V1.CreateBingAdsIntegration))
	authRouteGroup.DELETE("/:project_id/v1/bingads/disable", responseWrapper(V1.DisableBingAdsIntegration))
	authRouteGroup.GET("/:project_id/v1/bingads", responseWrapper(V1.GetBingAdsIntegration))
	authRouteGroup.PUT("/:project_id/v1/bingads/enable", responseWrapper(V1.EnableBingAdsIntegration))

	// marketo integration
	authRouteGroup.POST("/:project_id/v1/marketo", responseWrapper(V1.CreateMarketoIntegration))
	authRouteGroup.DELETE("/:project_id/v1/marketo/disable", responseWrapper(V1.DisableMarketoIntegration))
	authRouteGroup.GET("/:project_id/v1/marketo", responseWrapper(V1.GetMarketoIntegration))
	authRouteGroup.PUT("/:project_id/v1/marketo/enable", responseWrapper(V1.EnableMarketoIntegration))

	// alerts
	authRouteGroup.POST("/:project_id/v1/alerts", responseWrapper(V1.CreateAlertHandler))
	authRouteGroup.GET("/:project_id/v1/alerts", responseWrapper(V1.GetAlertsHandler))
	authRouteGroup.GET("/:project_id/v1/alerts/:id", responseWrapper(V1.GetAlertByIDHandler))
	authRouteGroup.DELETE("/:project_id/v1/alerts/:id", responseWrapper(V1.DeleteAlertHandler))
	authRouteGroup.POST("/:project_id/slack/auth", slack.SlackAuthRedirectHandler)
	authRouteGroup.GET("/:project_id/slack/channels", slack.GetSlackChannelsListHandler)
	authRouteGroup.DELETE("/:project_id/slack/delete", slack.DeleteSlackIntegrationHandler)

}

func InitSDKServiceRoutes(r *gin.Engine) {
	// Initialize swagger api docs only for development / staging.
	if C.GetConfig().Env != C.PRODUCTION {
		r.GET("/sdk/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	r.GET(ROUTE_SDK_ROOT+"/service/status", SDKStatusHandler)
	r.POST(ROUTE_SDK_ROOT+"/service/error", SDKErrorHandler)

	// Todo(Dinesh): Check integrity of token using encrytion/decryption
	// with secret, on middleware, to avoid spamming queue.

	// Getting project_id is moved to sdk request handler to
	// support queue workers also.
	// sdkRouteGroup.Use(mid.SetScopeProjectIdByToken())

	sdkRouteGroup := r.Group(ROUTE_SDK_ROOT)
	sdkRouteGroup.Use(mid.SetScopeProjectToken())
	sdkRouteGroup.GET("/project/get_settings", SDKGetProjectSettingsHandler)
	sdkRouteGroup.POST("/event/track", SDKTrackHandler)
	sdkRouteGroup.POST("/event/track/bulk", SDKBulkEventHandler)
	sdkRouteGroup.POST("/event/update_properties", SDKUpdateEventPropertiesHandler)
	sdkRouteGroup.POST("/user/identify", SDKIdentifyHandler)
	sdkRouteGroup.POST("/user/add_properties", SDKAddUserPropertiesHandler)
	sdkRouteGroup.POST("/adwords/documents/add", IH.DataServiceAdwordsAddDocumentHandler)

	ampSdkRouteGroup := r.Group(ROUTE_SDK_AMP_ROOT)
	ampSdkRouteGroup.POST("/event/track", SDKAMPTrackHandler)
	ampSdkRouteGroup.POST("/event/update_properties", SDKAMPUpdateEventPropertiesHandler)
	ampSdkRouteGroup.POST("/user/identify", SDKAMPIdentifyHandler)

	intRouteGroup := r.Group(ROUTE_INTEGRATIONS_ROOT)
	intRouteGroup.POST("/segment", mid.SetScopeProjectPrivateToken(), IntSegmentHandler)
	intRouteGroup.POST("/segment_platform",
		mid.SetScopeProjectPrivateTokenUsingBasicAuth(), IntSegmentPlatformHandler)
}

func InitIntRoutes(r *gin.Engine) {
	intRouteGroup := r.Group(ROUTE_INTEGRATIONS_ROOT)

	// Deprecated: /shopify routes are deprecated.
	// blocked gracefully for existing projects.
	intRouteGroup.POST("/shopify",
		mid.BlockRequestGracefully(),
		mid.SetScopeProjectIdByStoreAndSecret(),
		IntShopifyHandler)
	intRouteGroup.POST("/shopify_sdk",
		mid.BlockRequestGracefully(),
		mid.SetScopeProjectIdByToken(),
		IntShopifySDKHandler)

	intRouteGroup.POST("/adwords/enable",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		IntEnableAdwordsHandler)

	intRouteGroup.POST("/google_organic/enable",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		IntEnableGoogleOrganicHandler)

	intRouteGroup.POST("/facebook/add_access_token",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		IntFacebookAddAccessTokenHandler)

	intRouteGroup.POST("/linkedin/auth", IntLinkedinAuthHandler)
	intRouteGroup.POST("/linkedin/ad_accounts", IntLinkedinAccountHandler)

	intRouteGroup.POST("/linkedin/add_access_token",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		IntLinkedinAddAccessTokenHandler)

	intRouteGroup.POST("/salesforce/enable",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		IntEnableSalesforceHandler)

	intRouteGroup.POST("/salesforce/auth",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		SalesforceAuthRedirectHandler)
	intRouteGroup.GET(SalesforceCallbackRoute,
		SalesforceCallbackHandler)

	intRouteGroup.DELETE("/:project_id/:channel_name",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		IntDeleteHandler)

	intRouteGroup.GET("/slack/callback",slack.SlackCallbackHandler)

}

func InitDataServiceRoutes(r *gin.Engine) {
	dataServiceRouteGroup := r.Group(ROUTE_DATA_SERVICE_ROOT)

	//todo @ashhar: merge adwords and google_organic whereever possible
	dataServiceRouteGroup.POST("/adwords/documents/add",
		IH.DataServiceAdwordsAddDocumentHandler)
	dataServiceRouteGroup.POST("/adwords/documents/add_multiple",
		IH.DataServiceAdwordsAddMultipleDocumentsHandler)

	dataServiceRouteGroup.POST("/adwords/add_refresh_token",
		IntAdwordsAddRefreshTokenHandler)

	dataServiceRouteGroup.POST("/adwords/get_refresh_token",
		IntAdwordsGetRefreshTokenHandler)

	dataServiceRouteGroup.GET("/adwords/documents/project_last_sync_info",
		IH.DataServiceAdwordsGetLastSyncForProjectInfoHandler)

	dataServiceRouteGroup.GET("/adwords/documents/last_sync_info",
		IH.DataServiceAdwordsGetLastSyncInfoHandler)

	dataServiceRouteGroup.POST("/google_organic/documents/add",
		IH.DataServiceGoogleOrganicAddDocumentHandler)

	dataServiceRouteGroup.POST("/google_organic/documents/add_multiple",
		IH.DataServiceGoogleOrganicAddMultipleDocumentsHandler)

	dataServiceRouteGroup.POST("/google_organic/add_refresh_token",
		IntGoogleOrganicAddRefreshTokenHandler)

	dataServiceRouteGroup.POST("/google_organic/get_refresh_token",
		IntGoogleOrganicGetRefreshTokenHandler)

	dataServiceRouteGroup.GET("/google_organic/documents/project_last_sync_info",
		IH.DataServiceGoogleOrganicGetLastSyncForProjectInfoHandler)

	dataServiceRouteGroup.GET("/google_organic/documents/last_sync_info",
		IH.DataServiceGoogleOrganicGetLastSyncInfoHandler)

	dataServiceRouteGroup.POST("/hubspot/documents/add",
		IH.DataServiceHubspotAddDocumentHandler)

	dataServiceRouteGroup.POST("/hubspot/documents/add_batch",
		IH.DataServiceHubspotAddBatchDocumentHandler)

	dataServiceRouteGroup.GET("/hubspot/documents/sync_info",
		IH.DataServiceHubspotGetSyncInfoHandler)

	dataServiceRouteGroup.POST("/hubspot/documents/sync_info",
		IH.DataServiceHubspotUpdateSyncInfo)

	dataServiceRouteGroup.GET("/hubspot/documents/types/form",
		IH.DataServiceGetHubspotFormDocumentsHandler)

	dataServiceRouteGroup.GET("/facebook/project/settings",
		IH.DataServiceFacebookGetProjectSettings)

	dataServiceRouteGroup.POST("/facebook/documents/add",
		IH.DataServiceFacebookAddDocumentHandler)

	dataServiceRouteGroup.GET("/facebook/documents/last_sync_info",
		IH.DataServiceFacebookGetLastSyncInfoHandler)

	dataServiceRouteGroup.GET("linkedin/documents/last_sync_info",
		IH.DataServiceLinkedinGetLastSyncInfoHandler)

	dataServiceRouteGroup.POST("/linkedin/documents/add",
		IH.DataServiceLinkedinAddDocumentHandler)

	dataServiceRouteGroup.PUT("/linkedin/access_token",
		IH.DataServiceLinkedinUpdateAccessToken)

	dataServiceRouteGroup.POST("/metrics",
		IH.DataServiceRecordMetricHandler)

	dataServiceRouteGroup.GET("/linkedin/project/settings",
		IH.DataServiceLinkedinGetProjectSettings)

	dataServiceRouteGroup.GET("/linkedin/project/settings/projects",
		IH.DataServiceLinkedinGetProjectSettingsForProjects)

	dataServiceRouteGroup.GET("/task/details", responseWrapper(V1.GetTaskDetailsByNameHandler))
	dataServiceRouteGroup.GET("/task/deltas", responseWrapper(V1.GetAllToBeExecutedDeltasHandler))
	dataServiceRouteGroup.GET("/task/delta_timestamp", responseWrapper(V1.GetTaskDeltaAsTimeHandler))
	dataServiceRouteGroup.GET("/task/delta_end_timestamp", responseWrapper(V1.GetTaskEndTimeHandler))
	dataServiceRouteGroup.POST("/task/begin", responseWrapper(V1.InsertTaskBeginRecordHandler))
	dataServiceRouteGroup.POST("/task/end", responseWrapper(V1.InsertTaskEndRecordHandler))
	dataServiceRouteGroup.DELETE("/task/end", responseWrapper(V1.DeleteTaskEndRecordHandler))
	dataServiceRouteGroup.GET("/task/dependent_task_done", responseWrapper(V1.IsDependentTaskDoneHandler))

}

type Error struct {
	Code           string `json:"code"`
	DisplayMessage string `json:"display_message"`
	Details        string `json:"details"`
	TrackingId     string `json:"tracking_id"`
}

func responseWrapper(f func(c *gin.Context) (interface{}, int, string, string, bool)) gin.HandlerFunc {

	return func(c *gin.Context) {
		data, statusCode, errorCode, errMsg, isErr := f(c)
		if isErr {
			err := Error{
				Code:           errorCode,
				DisplayMessage: V1.ErrorMessages[errorCode],
				Details:        "",
				TrackingId:     U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
			}
			if statusCode == http.StatusPartialContent {
				c.JSON(statusCode, gin.H{"data": data, "error": errMsg, "err": err})
				return
			}
			c.JSON(statusCode, gin.H{"error": errMsg, "err": err})
			return
		}
		c.JSON(statusCode, data)
	}
}
