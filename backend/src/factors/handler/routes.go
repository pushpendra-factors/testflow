package handler

import (
	mid "factors/middleware"
	"net/http"

	IH "factors/handler/internal"

	"github.com/gin-gonic/gin"
)

const ROUTE_SDK_ROOT = "/sdk"
const ROUTE_SDK_AMP_ROOT = "/sdk/amp"
const ROUTE_PROJECTS_ROOT = "/projects"
const ROUTE_INTEGRATIONS_ROOT = "/integrations"
const ROUTE_DATA_SERVICE_ROOT = "/data_service"

func InitAppRoutes(r *gin.Engine) {
	r.GET("/status", func(c *gin.Context) {
		resp := map[string]string{
			"status": "success",
		}
		c.JSON(http.StatusOK, resp)
		return
	})

	r.POST("/accounts/signup", SignUp)
	r.POST("/agents/signin", Signin)
	r.GET("/agents/signout", Signout)
	r.POST("/agents/forgotpassword", AgentGenerateResetPasswordLinkEmail)
	r.POST("/agents/setpassword", mid.ValidateAgentSetPasswordRequest(), AgentSetPassword)
	r.PUT("/agents/updatepassword", mid.SetLoggedInAgent(), UpdateAgentPassword)
	r.POST("/agents/activate", mid.ValidateAgentActivationRequest(), AgentActivate)
	r.GET("/agents/billing", mid.SetLoggedInAgent(), GetAgentBillingAccount)
	r.PUT("/agents/billing", mid.SetLoggedInAgent(), UpdateAgentBillingAccount)
	r.GET("/agents/info", mid.SetLoggedInAgent(), AgentInfo)
	r.PUT("/agents/info", mid.SetLoggedInAgent(), UpdateAgentInfo)

	r.POST(ROUTE_PROJECTS_ROOT, mid.SetLoggedInAgent(), CreateProjectHandler)

	r.GET(ROUTE_PROJECTS_ROOT,
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		GetProjectsHandler)

	// Auth route group with authentication an authorization middleware.
	authRouteGroup := r.Group(ROUTE_PROJECTS_ROOT)
	authRouteGroup.Use(mid.SetLoggedInAgent())
	authRouteGroup.Use(mid.SetAuthorizedProjectsByLoggedInAgent())
	authRouteGroup.Use(mid.ValidateLoggedInAgentHasAccessToRequestProject())
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
	authRouteGroup.GET("/:project_id/models", GetProjectModelsHandler)
	authRouteGroup.GET("/:project_id/filters", GetFiltersHandler)
	authRouteGroup.POST("/:project_id/filters", CreateFilterHandler)
	authRouteGroup.PUT("/:project_id/filters/:filter_id", UpdateFilterHandler)
	authRouteGroup.DELETE("/:project_id/filters/:filter_id", DeleteFilterHandler)
	authRouteGroup.GET("/:project_id/event_names/:event_name/properties", GetEventPropertiesHandler)
	authRouteGroup.GET("/:project_id/event_names/:event_name/properties/:property_name/values", GetEventPropertyValuesHandler)
	authRouteGroup.GET("/:project_id/users", GetUsersHandler)
	authRouteGroup.GET("/:project_id/users/:user_id", GetUserHandler)
	authRouteGroup.GET("/:project_id/user_properties", GetUserPropertiesHandler)
	authRouteGroup.GET("/:project_id/user_properties/:property_name/values", GetUserPropertyValuesHandler)
	authRouteGroup.POST("/:project_id/factor", FactorHandler)
	authRouteGroup.POST("/:project_id/query", QueryHandler)
	authRouteGroup.POST("/:project_id/channels/query", ChannelQueryHandler)
	authRouteGroup.GET("/:project_id/channels/filter_values", GetChannelFilterValuesHandler)
	authRouteGroup.GET("/:project_id/reports", GetReportsHandler)
	authRouteGroup.GET("/:project_id/reports/:report_id", GetReportHandler)
	authRouteGroup.POST("/:project_id/attribution/query", AttributionHandler)

	// TODO
	// Scope this with Project Admin
	authRouteGroup.GET("/:project_id/agents", GetProjectAgentsHandler)
	authRouteGroup.POST("/:project_id/agents/invite", AgentInvite)
	authRouteGroup.PUT("/:project_id/agents/remove", RemoveProjectAgent)
	authRouteGroup.GET("/:project_id/settings", GetProjectSettingHandler)
	authRouteGroup.PUT("/:project_id/settings", UpdateProjectSettingsHandler)
}

func InitSDKServiceRoutes(r *gin.Engine) {
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

	ampSdkRouteGroup := r.Group(ROUTE_SDK_AMP_ROOT)
	ampSdkRouteGroup.POST("/event/track", SDKAMPTrackHandler)
	ampSdkRouteGroup.POST("/event/update_properties", SDKAMPUpdateEventPropertiesHandler)

	intRouteGroup := r.Group(ROUTE_INTEGRATIONS_ROOT)
	intRouteGroup.POST("/segment", mid.SetScopeProjectPrivateToken(), IntSegmentHandler)
	intRouteGroup.POST("/segment_platform",
		mid.SetScopeProjectPrivateTokenUsingBasicAuth(), IntSegmentHandler)
}

func InitIntRoutes(r *gin.Engine) {
	intRouteGroup := r.Group(ROUTE_INTEGRATIONS_ROOT)

	intRouteGroup.POST("/shopify",
		mid.SetScopeProjectIdByStoreAndSecret(),
		IntShopifyHandler)

	intRouteGroup.POST("/shopify_sdk",
		mid.SetScopeProjectIdByToken(),
		IntShopifySDKHandler)

	intRouteGroup.POST("/adwords/enable",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		IntEnableAdwordsHandler)

	intRouteGroup.POST("/facebook/add_access_token",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		IntFacebookAddAccessTokenHandler)

	intRouteGroup.POST("/salesforce/enable",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		IntEnableSalesforceHandler)

	intRouteGroup.POST("/salesforce/auth",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		SalesforceAuthRedirectHandler)
	intRouteGroup.GET(SALESFORCE_CALLBACK_URL,
		SalesforceCallbackHandler)
}

func InitDataServiceRoutes(r *gin.Engine) {
	dataServiceRouteGroup := r.Group(ROUTE_DATA_SERVICE_ROOT)

	dataServiceRouteGroup.POST("/adwords/documents/add",
		IH.DataServiceAdwordsAddDocumentHandler)

	dataServiceRouteGroup.POST("/adwords/add_refresh_token",
		IntAdwordsAddRefreshTokenHandler)

	dataServiceRouteGroup.POST("/adwords/get_refresh_token",
		IntAdwordsGetRefreshTokenHandler)

	dataServiceRouteGroup.GET("/adwords/documents/last_sync_info",
		IH.DataServiceAdwordsGetLastSyncInfoHandler)

	dataServiceRouteGroup.POST("/hubspot/documents/add",
		IH.DataServiceHubspotAddDocumentHandler)

	dataServiceRouteGroup.GET("/hubspot/documents/sync_info",
		IH.DataServiceHubspotGetSyncInfoHandler)

	dataServiceRouteGroup.GET("/hubspot/documents/types/form",
		IH.DataServiceGetHubspotFormDocumentsHandler)

	dataServiceRouteGroup.GET("/facebook/project/settings",
		IH.DataServiceFacebookGetProjectSettings)

	dataServiceRouteGroup.POST("/facebook/documents/add",
		IH.DataServiceFacebookAddDocumentHandler)

	dataServiceRouteGroup.GET("/facebook/documents/last_sync_info",
		IH.DataServiceFacebookGetLastSyncInfoHandler)

	dataServiceRouteGroup.POST("/salesforce/documents/add",
		IH.DataServiceSalesforceAddDocumentHandler)
	dataServiceRouteGroup.GET("/salesforce/documents/last_sync_info",
		IH.DataServiceSalesforceGetLastSyncInfoHandler)
}
