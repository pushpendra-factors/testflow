package handler

import (
	C "factors/config"
	IH "factors/handler/internal"
	V1 "factors/handler/v1"
	mid "factors/middleware"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"

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
const ROUTE_COMMON_ROOT = "/common"

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
	r.GET(routePrefix+"/agents/signout", mid.SetLoggedInAgent(), Signout)
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
	shareRouteGroup.Use(mid.ValidateAccessToSharedEntity(model.ShareableURLEntityTypeQuery))

	shareRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/query", responseWrapper(EventsQueryHandler))
	shareRouteGroup.POST("/:project_id/query", responseWrapper(QueryHandler))
	shareRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/attribution/query", responseWrapper(V1.AttributionHandlerV1))
	shareRouteGroup.POST("/:project_id/attribution/query", responseWrapper(AttributionHandler))
	shareRouteGroup.POST("/:project_id/profiles/query", responseWrapper(ProfilesQueryHandler))
	shareRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/kpi/query", responseWrapper(V1.ExecuteKPIQueryHandler))

	// Auth route group with authentication an authorization middleware.
	authRouteGroup := r.Group(routePrefix + ROUTE_PROJECTS_ROOT)
	authRouteGroup.Use(mid.SetLoggedInAgent())
	authRouteGroup.Use(mid.SetAuthorizedProjectsByLoggedInAgent())
	authRouteGroup.Use(mid.ValidateLoggedInAgentHasAccessToRequestProject())

	authRouteGroup.PUT("/:project_id", mid.SkipDemoProjectWriteAccess(), EditProjectHandler)

	// Dashboard endpoints
	authRouteGroup.GET("/:project_id/dashboards", stringifyWrapper(GetDashboardsHandler))
	authRouteGroup.POST("/:project_id/dashboards", mid.SkipDemoProjectWriteAccess(), stringifyWrapper(CreateDashboardHandler))
	authRouteGroup.PUT("/:project_id/dashboards/:dashboard_id", mid.SkipDemoProjectWriteAccess(), UpdateDashboardHandler)
	authRouteGroup.GET("/:project_id/dashboards/:dashboard_id/units", stringifyWrapper(GetDashboardUnitsHandler))
	authRouteGroup.POST("/:project_id/dashboards/:dashboard_id/units", mid.SkipDemoProjectWriteAccess(), stringifyWrapper(CreateDashboardUnitHandler))
	authRouteGroup.PUT("/:project_id/dashboards/:dashboard_id/units/:unit_id", mid.SkipDemoProjectWriteAccess(), UpdateDashboardUnitHandler)
	authRouteGroup.DELETE("/:project_id/dashboards/:dashboard_id/units/:unit_id", mid.SkipDemoProjectWriteAccess(), DeleteDashboardUnitHandler)
	authRouteGroup.POST("/:project_id/dashboard/:dashboard_id/units/query/web_analytics",
		DashboardUnitsWebAnalyticsQueryHandler)
	authRouteGroup.GET("/:project_id/event_names", GetEventNamesHandler)
	authRouteGroup.GET("/:project_id/user/event_names", GetEventNamesByUserHandler)
	authRouteGroup.GET(":project_id/groups/:group_name/event_names", GetEventNamesByGroupHandler)

	// Offline Touch Point rules
	authRouteGroup.GET("/:project_id/otp_rules", responseWrapper(GetOTPRuleHandler))
	authRouteGroup.POST("/:project_id/otp_rules", mid.SkipDemoProjectWriteAccess(), responseWrapper(CreateOTPRuleHandler))
	authRouteGroup.PUT("/:project_id/otp_rules/:rule_id", mid.SkipDemoProjectWriteAccess(), responseWrapper(UpdateOTPRuleHandler))
	authRouteGroup.GET("/:project_id/otp_rules/:rule_id", responseWrapper(SearchOTPRuleHandler))
	authRouteGroup.DELETE("/:project_id/otp_rules/:rule_id", mid.SkipDemoProjectWriteAccess(), responseWrapper(DeleteOTPRuleHandler))

	// Dashboard templates
	authCommonRouteGroup := r.Group(routePrefix + ROUTE_COMMON_ROOT)
	authCommonRouteGroup.GET("/dashboard_templates/:id/search", SearchTemplateHandler)
	authCommonRouteGroup.GET("/dashboard_templates", GetDashboardTemplatesHandler)
	authCommonRouteGroup.POST("/dashboard_template/create", mid.SkipDemoProjectWriteAccess(), CreateTemplateHandler)
	authRouteGroup.POST("/:project_id/dashboard_template/:id/trigger", mid.SkipDemoProjectWriteAccess(), GenerateDashboardFromTemplateHandler)
	authRouteGroup.POST("/:project_id/dashboards/:dashboard_id/trigger", mid.SkipDemoProjectWriteAccess(), GenerateTemplateFromDashboardHandler)

	authRouteGroup.GET("/:project_id/queries", stringifyWrapper(GetQueriesHandler))
	authRouteGroup.POST("/:project_id/queries", mid.SkipDemoProjectWriteAccess(), stringifyWrapper(CreateQueryHandler))

	authRouteGroup.PUT("/:project_id/queries/:query_id", mid.SkipDemoProjectWriteAccess(), UpdateSavedQueryHandler)
	authRouteGroup.DELETE("/:project_id/queries/:query_id", mid.SkipDemoProjectWriteAccess(), DeleteSavedQueryHandler)
	authRouteGroup.GET("/:project_id/queries/search", stringifyWrapper(SearchQueriesHandler))
	authRouteGroup.GET("/:project_id/models", GetProjectModelsHandler)
	authRouteGroup.GET("/:project_id/filters", GetFiltersHandler)
	authRouteGroup.POST("/:project_id/filters", mid.SkipDemoProjectWriteAccess(), CreateFilterHandler)
	authRouteGroup.PUT("/:project_id/filters/:filter_id", mid.SkipDemoProjectWriteAccess(), UpdateFilterHandler)
	authRouteGroup.DELETE("/:project_id/filters/:filter_id", mid.SkipDemoProjectWriteAccess(), DeleteFilterHandler)
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

	// shareable url endpoints
	authRouteGroup.GET("/:project_id/shareable_url", GetShareableURLsHandler)
	authRouteGroup.POST("/:project_id/shareable_url", CreateShareableURLHandler)
	authRouteGroup.DELETE("/:project_id/shareable_url/:share_id", mid.SkipDemoProjectWriteAccess(), DeleteShareableURLHandler)
	authRouteGroup.DELETE("/:project_id/shareable_url/revoke/:query_id", mid.SkipDemoProjectWriteAccess(), RevokeShareableURLHandler)

	// v1 Dashboard endpoints
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/dashboards/multi/:dashboard_ids/units", mid.SkipDemoProjectWriteAccess(), stringifyWrapper(CreateDashboardUnitForMultiDashboardsHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/dashboards/queries/:dashboard_id/units", mid.SkipDemoProjectWriteAccess(), stringifyWrapper(CreateDashboardUnitsForMultipleQueriesHandler))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/dashboards/:dashboard_id/units/multi/:unit_ids", mid.SkipDemoProjectWriteAccess(), DeleteMultiDashboardUnitHandler)
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/dashboards/:dashboard_id", mid.SkipDemoProjectWriteAccess(), DeleteDashboardHandler)

	// attribution V1 endpoints
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/attribution/queries", stringifyWrapper(V1.GetAttributionQueriesHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/attribution/queries", mid.SkipDemoProjectWriteAccess(), stringifyWrapper(V1.CreateAttributionV1QueryAndSaveToDashboardHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/attribution/dashboards", stringifyWrapper(V1.GetOrCreateAttributionV1DashboardHandler))

	// v1 KPI endpoints
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/kpi/config", responseWrapper(V1.GetKPIConfigHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/kpi/filter_values", responseWrapper(V1.GetKPIFilterValuesHandler))

	// v1 custom metrics - admin/settings side.
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics/config/v1", V1.GetCustomMetricsConfigV1)
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.CreateCustomMetric))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics", responseWrapper(V1.GetCustomMetrics))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics/:id", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.DeleteCustomMetrics))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics/prebuilt/add_missing", responseWrapper(V1.CreateMissingPreBuiltCustomKPI))

	// v1 CRM And Smart Event endpoints
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_event", GetSmartEventFiltersHandler)
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/smart_event", mid.SkipDemoProjectWriteAccess(), CreateSmartEventFilterHandler)
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/smart_event", mid.SkipDemoProjectWriteAccess(), UpdateSmartEventFilterHandler)
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/smart_event", mid.SkipDemoProjectWriteAccess(), responseWrapper(DeleteSmartEventFilterHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/crm/:crm_source/:object_type/properties", GetCRMObjectPropertiesHandler)
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/crm/:crm_source/:object_type/properties/:property_name/values", GetCRMObjectValuesByPropertyNameHandler)
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/templates/:type/config", responseWrapper(V1.GetTemplateConfigHandler))
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/templates/:type/config", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.UpdateTemplateConfigHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/templates/:type/query", responseWrapper(V1.ExecuteTemplateQueryHandler))

	// smart Properties
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/config/:object_type", responseWrapper(GetSmartPropertyRulesConfigHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules", mid.SkipDemoProjectWriteAccess(), responseWrapper(CreateSmartPropertyRulesHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules", responseWrapper(GetSmartPropertyRulesHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules/:rule_id", responseWrapper(GetSmartPropertyRuleByRuleIDHandler))
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules/:rule_id", mid.SkipDemoProjectWriteAccess(), responseWrapper(UpdateSmartPropertyRulesHandler))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules/:rule_id", mid.SkipDemoProjectWriteAccess(), responseWrapper(DeleteSmartPropertyRulesHandler))

	// content groups
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/contentgroup", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.CreateContentGroupHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/contentgroup", responseWrapper(V1.GetContentGroupHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/contentgroup/:id", responseWrapper(V1.GetContentGroupByIDHandler))
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/contentgroup/:id", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.UpdateContentGroupHandler))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/contentgroup/:id", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.DeleteContentGroupHandler))
	// TODO
	// Scope this with Project Admin
	authRouteGroup.GET("/:project_id/agents", GetProjectAgentsHandler)
	authRouteGroup.POST("/:project_id/agents/invite", mid.SkipDemoProjectWriteAccess(), AgentInvite)
	authRouteGroup.POST("/:project_id/agents/batchinvite", mid.SkipDemoProjectWriteAccess(), AgentInviteBatch)
	authRouteGroup.PUT("/:project_id/agents/remove", mid.SkipDemoProjectWriteAccess(), RemoveProjectAgent)
	authRouteGroup.PUT("/:project_id/agents/update", mid.SkipDemoProjectWriteAccess(), AgentUpdate)
	authRouteGroup.GET("/:project_id/settings", mid.SkipDemoProjectWriteAccess(), GetProjectSettingHandler)
	authRouteGroup.GET("/:project_id/v1/settings", mid.SkipDemoProjectWriteAccess(), V1.GetProjectSettingHandler)
	authRouteGroup.PUT("/:project_id/settings", mid.SkipDemoProjectWriteAccess(), UpdateProjectSettingsHandler)
	authRouteGroup.GET("/:project_id/clickable_elements", GetClickableElementsHandler)
	authRouteGroup.GET("/:project_id/clickable_elements/:id/toggle", ToggleClickableElementHandler)
	authRouteGroup.PUT("/:project_id/leadsquaredsettings", mid.SkipDemoProjectWriteAccess(), UpdateLeadSquaredConfigHandler)
	authRouteGroup.DELETE("/:project_id/leadsquaredsettings/remove", mid.SkipDemoProjectWriteAccess(), RemoveLeadSquaredConfigHandler)

	// V1 Routes
	authRouteGroup.GET("/:project_id/v1/event_names", V1.GetEventNamesHandler)
	authRouteGroup.GET("/:project_id/v1/event_names/:type", V1.GetEventNamesByTypeHandler)
	authRouteGroup.GET("/:project_id/v1/agents", mid.SkipDemoProjectWriteAccess(), V1.GetProjectAgentsHandler)
	r.GET(routePrefix+"/"+ROUTE_PROJECTS_ROOT_V1,
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		V1.GetProjectsHandler)
	r.GET(routePrefix+"/v1/demoprojects",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		V1.GetDemoProjects)

	// Tracked Events
	authRouteGroup.POST("/:project_id/v1/factors/tracked_event", mid.SkipDemoProjectWriteAccess(), V1.CreateFactorsTrackedEventsHandler)
	authRouteGroup.DELETE("/:project_id/v1/factors/tracked_event/remove", mid.SkipDemoProjectWriteAccess(), V1.RemoveFactorsTrackedEventsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/tracked_event", V1.GetAllFactorsTrackedEventsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/grouped_tracked_event", V1.GetAllGroupedFactorsTrackedEventsHandler)

	// Tracked User Property
	authRouteGroup.POST("/:project_id/v1/factors/tracked_user_property", mid.SkipDemoProjectWriteAccess(), V1.CreateFactorsTrackedUserPropertyHandler)
	authRouteGroup.DELETE("/:project_id/v1/factors/tracked_user_property/remove", mid.SkipDemoProjectWriteAccess(), V1.RemoveFactorsTrackedUserPropertyHandler)
	authRouteGroup.GET("/:project_id/v1/factors/tracked_user_property", V1.GetAllFactorsTrackedUserPropertiesHandler)

	// Goals
	authRouteGroup.POST("/:project_id/v1/factors/goals", mid.SkipDemoProjectWriteAccess(), V1.CreateFactorsGoalsHandler)
	authRouteGroup.DELETE("/:project_id/v1/factors/goals/remove", mid.SkipDemoProjectWriteAccess(), V1.RemoveFactorsGoalsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/goals", V1.GetAllFactorsGoalsHandler)
	authRouteGroup.PUT("/:project_id/v1/factors/goals/update", mid.SkipDemoProjectWriteAccess(), V1.UpdateFactorsGoalsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/goals/search", V1.SearchFactorsGoalHandler)
	authRouteGroup.POST("/:project_id/v1/factor", V1.PostFactorsHandler)
	authRouteGroup.POST("/:project_id/v1/factor/compare", V1.PostFactorsCompareHandler)
	authRouteGroup.POST("/:project_id/v1/events/displayname", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.CreateDisplayNamesHandler))
	authRouteGroup.GET("/:project_id/v1/events/displayname", responseWrapper(V1.GetAllDistinctEventProperties))
	authRouteGroup.GET("/:project_id/v1/factor", V1.GetFactorsHandler)
	authRouteGroup.GET("/:project_id/v1/factor/model_metadata", V1.GetModelMetaData)

	authRouteGroup.GET("/:project_id/insights", responseWrapper(V1.GetWeeklyInsightsHandler))
	authRouteGroup.GET("/:project_id/weekly_insights_metadata", responseWrapper(V1.GetWeeklyInsightsMetadata))
	authRouteGroup.POST("/:project_id/feedback", mid.SkipDemoProjectWriteAccess(), V1.PostFeedbackHandler)

	// bingads integration
	authRouteGroup.POST("/:project_id/v1/bingads", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.CreateBingAdsIntegration))
	authRouteGroup.DELETE("/:project_id/v1/bingads/disable", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.DisableBingAdsIntegration))
	authRouteGroup.GET("/:project_id/v1/bingads", responseWrapper(V1.GetBingAdsIntegration))
	authRouteGroup.PUT("/:project_id/v1/bingads/enable", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.EnableBingAdsIntegration))

	// marketo integration
	authRouteGroup.POST("/:project_id/v1/marketo", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.CreateMarketoIntegration))
	authRouteGroup.DELETE("/:project_id/v1/marketo/disable", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.DisableMarketoIntegration))
	authRouteGroup.GET("/:project_id/v1/marketo", responseWrapper(V1.GetMarketoIntegration))
	authRouteGroup.PUT("/:project_id/v1/marketo/enable", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.EnableMarketoIntegration))

	// alerts
	authRouteGroup.POST("/:project_id/v1/alerts", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.CreateAlertHandler))
	authRouteGroup.GET("/:project_id/v1/alerts", responseWrapper(V1.GetAlertsHandler))
	authRouteGroup.GET("/:project_id/v1/alerts/:id", responseWrapper(V1.GetAlertByIDHandler))
	authRouteGroup.DELETE("/:project_id/v1/alerts/:id", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.DeleteAlertHandler))
	authRouteGroup.PUT("/:project_id/v1/alerts/:id", mid.SkipDemoProjectWriteAccess(), responseWrapper(V1.EditAlertHandler))
	authRouteGroup.POST("/:project_id/slack/auth", mid.SkipDemoProjectWriteAccess(), slack.SlackAuthRedirectHandler)
	authRouteGroup.GET("/:project_id/slack/channels", mid.SkipDemoProjectWriteAccess(), slack.GetSlackChannelsListHandler)
	authRouteGroup.DELETE("/:project_id/slack/delete", mid.SkipDemoProjectWriteAccess(), slack.DeleteSlackIntegrationHandler)
	authRouteGroup.POST("/:project_id/v1/alerts/send_now", mid.SkipDemoProjectWriteAccess(), V1.QuerySendNowHandler)
	authRouteGroup.GET("/:project_id/v1/dataobservability/metrics", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.GetAnalyticsMetricsFromStorage))
	authRouteGroup.GET("/:project_id/v1/dataobservability/alerts", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.GetAnalyticsAlertsFromStorage))

	// Timeline
	authRouteGroup.POST("/:project_id/v1/profiles/users", responseWrapper(V1.GetProfileUsersHandler))
	authRouteGroup.GET("/:project_id/v1/profiles/users/:id", responseWrapper(V1.GetProfileUserDetailsHandler))
	authRouteGroup.POST("/:project_id/v1/profiles/accounts", responseWrapper(V1.GetProfileAccountsHandler))
	authRouteGroup.GET("/:project_id/v1/profiles/accounts/:id", responseWrapper(V1.GetProfileAccountDetailsHandler))
	authRouteGroup.POST("/:project_id/segments", CreateSegmentHandler)
	authRouteGroup.GET("/:project_id/segments", responseWrapper(GetSegmentsHandler))
	authRouteGroup.GET("/:project_id/segments/:id", responseWrapper(GetSegmentByIdHandler))
	authRouteGroup.PUT("/:project_id/segments/:id", UpdateSegmentHandler)
	authRouteGroup.DELETE("/:project_id/segments/:id", DeleteSegmentByIdHandler)

	// weekly insights, explain
	authRouteGroup.PUT("/:project_id/v1/weeklyinsights", mid.SetLoggedInAgentInternalOnly(), UpdateWeeklyInsightsHandler)
	authRouteGroup.PUT("/:project_id/v1/explain", mid.SetLoggedInAgentInternalOnly(), UpdateExplainHandler)
	authRouteGroup.PUT("/:project_id/v1/pathanalysis", mid.SetLoggedInAgentInternalOnly(), UpdatePathAnalysisHandler)

	// path analysis
	authRouteGroup.GET("/:project_id/v1/pathanalysis", responseWrapper(V1.GetPathAnalysisEntityHandler))
	authRouteGroup.POST("/:project_id/v1/pathanalysis", responseWrapper(V1.CreatePathAnalysisEntityHandler))
	authRouteGroup.DELETE("/:project_id/v1/pathanalysis/:id", V1.DeleteSavedPathAnalysisEntityHandler)
	authRouteGroup.GET("/:project_id/v1/pathanalysis/:id", responseWrapper(V1.GetPathAnalysisData))

	//explainV2
	authRouteGroup.GET("/:project_id/v1/explainV2", V1.GetFactorsHandlerV2)
	authRouteGroup.GET("/:project_id/v1/explainV2/goals", responseWrapper(V1.GetExplainV2EntityHandler))
	authRouteGroup.POST("/:project_id/v1/explainV2", V1.PostFactorsHandlerV2)
	authRouteGroup.POST("/:project_id/v1/explainV2/job", responseWrapper(V1.CreateExplainV2EntityHandler))
	authRouteGroup.DELETE("/:project_id/v1/explainV2/:id", V1.DeleteSavedExplainV2EntityHandler)

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

	// DEPRECATED: Kept for backward compatibility.
	// Used on only on old npm installations. JS_SDK uses /get_info.
	sdkRouteGroup.GET("/project/get_settings", SDKGetProjectSettingsHandler)

	sdkRouteGroup.POST("/get_info", SDKGetInfoHandler)
	sdkRouteGroup.POST("/event/track", SDKTrackHandler)
	sdkRouteGroup.POST("/event/track/bulk", SDKBulkEventHandler)
	sdkRouteGroup.POST("/event/update_properties", SDKUpdateEventPropertiesHandler)
	sdkRouteGroup.POST("/user/identify", SDKIdentifyHandler)
	sdkRouteGroup.POST("/user/add_properties", SDKAddUserPropertiesHandler)
	sdkRouteGroup.POST("/adwords/documents/add", IH.DataServiceAdwordsAddDocumentHandler)
	sdkRouteGroup.POST("/capture_click", SDKCaptureClickHandler)
	sdkRouteGroup.POST("/form_fill", SDKFormFillHandler)

	ampSdkRouteGroup := r.Group(ROUTE_SDK_AMP_ROOT)
	ampSdkRouteGroup.POST("/event/track", SDKAMPTrackHandler)
	ampSdkRouteGroup.POST("/event/update_properties", SDKAMPUpdateEventPropertiesHandler)
	ampSdkRouteGroup.POST("/user/identify", SDKAMPIdentifyHandler)

	intRouteGroup := r.Group(ROUTE_INTEGRATIONS_ROOT)
	intRouteGroup.POST("/segment_platform",
		mid.SetScopeProjectPrivateTokenUsingBasicAuth(), IntSegmentPlatformHandler)
	intRouteGroup.POST("/rudderstack_platform",
		mid.SetScopeProjectPrivateTokenUsingBasicAuth(), IntSegmentPlatformHandler)

	// Note: /segment is the old segment API Hook which was used directly.
	intRouteGroup.POST("/segment", mid.SetScopeProjectPrivateToken(), IntSegmentHandler)
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
	// salesforce integration.
	intRouteGroup.GET(SalesforceCallbackRoute,
		SalesforceCallbackHandler)

	intRouteGroup.POST("/hubspot/auth",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		HubspotAuthRedirectHandler)

	// hubspot integration.
	intRouteGroup.GET(HubspotCallbackRoute,
		HubspotCallbackHandler)

	intRouteGroup.DELETE("/:project_id/:channel_name",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		mid.SkipDemoProjectWriteAccess(),
		IntDeleteHandler)

	intRouteGroup.GET("/slack/callback", slack.SlackCallbackHandler)

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

func stringifyWrapper(f func(c *gin.Context) (interface{}, int, string, bool)) gin.HandlerFunc {

	return func(c *gin.Context) {
		data, statusCode, errMsg, isErr := f(c)
		if isErr {
			c.AbortWithStatusJSON(statusCode, gin.H{"error": errMsg})
			return
		}
		responseType := reflect.TypeOf(data).Kind()
		if responseType == reflect.Slice {
			switch data.(type) {
			case []model.Queries:
				queriesResp := make([]model.QueriesString, 0)
				responseObj := data.([]model.Queries)
				for _, query := range responseObj {
					queriesResp = append(queriesResp, ConvertQuery(query))
				}
				c.JSON(statusCode, queriesResp)
				return
			case []*model.Queries:
				queriesResp := make([]model.QueriesString, 0)
				responseObj := data.([]*model.Queries)
				for _, query := range responseObj {
					queriesResp = append(queriesResp, ConvertQuery(*query))
				}
				c.JSON(statusCode, queriesResp)
				return
			case []model.DashboardUnit:
				unitResp := make([]model.DashboardUnitString, 0)
				responseObj := data.([]model.DashboardUnit)
				for _, du := range responseObj {
					unitResp = append(unitResp, ConvertDashboardUnit(du))
				}
				c.JSON(statusCode, unitResp)
				return
			case []*model.DashboardUnit:
				unitResp := make([]model.DashboardUnitString, 0)
				responseObj := data.([]*model.DashboardUnit)
				for _, du := range responseObj {
					unitResp = append(unitResp, ConvertDashboardUnit(*du))
				}
				c.JSON(statusCode, unitResp)
				return
			case []model.Dashboard:
				dashboardResp := make([]model.DashboardString, 0)
				responseObj := data.([]model.Dashboard)
				for _, da := range responseObj {
					dashboardResp = append(dashboardResp, ConvertDashboard(da))
				}
				c.JSON(statusCode, dashboardResp)
				return
			case []*model.Dashboard:
				dashboardResp := make([]model.DashboardString, 0)
				responseObj := data.([]*model.Dashboard)
				for _, da := range responseObj {
					dashboardResp = append(dashboardResp, ConvertDashboard(*da))
				}
				c.JSON(statusCode, dashboardResp)
				return
			default:
				c.JSON(statusCode, data)
				return
			}
		} else {
			switch data.(type) {
			case model.Queries:
				responseObj := data.(model.Queries)
				c.JSON(statusCode, ConvertQuery(responseObj))
				return
			case *model.Queries:
				responseObj := data.(*model.Queries)
				c.JSON(statusCode, ConvertQuery(*responseObj))
				return
			case model.DashboardUnit:
				responseObj := data.(model.DashboardUnit)
				c.JSON(statusCode, ConvertDashboardUnit(responseObj))
				return
			case *model.DashboardUnit:
				responseObj := data.(*model.DashboardUnit)
				c.JSON(statusCode, ConvertDashboardUnit(*responseObj))
				return
			case model.Dashboard:
				responseObj := data.(model.Dashboard)
				c.JSON(statusCode, ConvertDashboard(responseObj))
				return
			case *model.Dashboard:
				responseObj := data.(*model.Dashboard)
				c.JSON(statusCode, ConvertDashboard(*responseObj))
				return
			default:
				c.JSON(statusCode, data)
				return
			}
		}
	}
}

func ConvertQuery(data model.Queries) model.QueriesString {
	return model.QueriesString{
		ID: fmt.Sprintf("%d", data.ID),
		// Foreign key queries(project_id) ref projects(id).
		ProjectID:        data.ProjectID,
		Title:            data.Title,
		Query:            data.Query,
		Type:             data.Type,
		IsDeleted:        data.IsDeleted,
		CreatedBy:        data.CreatedBy,
		CreatedByName:    data.CreatedByName,
		CreatedByEmail:   data.CreatedByEmail,
		CreatedAt:        data.CreatedAt,
		UpdatedAt:        data.UpdatedAt,
		Settings:         data.Settings,
		IdText:           data.IdText,
		Converted:        data.Converted,
		IsDashboardQuery: data.IsDashboardQuery,
	}
}

func ConvertDashboardUnit(data model.DashboardUnit) model.DashboardUnitString {
	return model.DashboardUnitString{
		ID: fmt.Sprintf("%d", data.ID),
		// Foreign key dashboard_units(project_id) ref projects(id).
		ProjectID:    data.ProjectID,
		DashboardId:  fmt.Sprintf("%d", data.DashboardId),
		Description:  data.Description,
		Presentation: data.Presentation,
		IsDeleted:    data.IsDeleted,
		CreatedAt:    data.CreatedAt,
		UpdatedAt:    data.UpdatedAt,
		QueryId:      fmt.Sprintf("%d", data.QueryId),
	}
}

func ConvertDashboard(data model.Dashboard) model.DashboardString {
	return model.DashboardString{
		ID: fmt.Sprintf("%d", data.ID),
		// Foreign key dashboards(project_id) ref projects(id).
		ProjectId:     data.ProjectId,
		AgentUUID:     data.AgentUUID,
		Name:          data.Name,
		Description:   data.Description,
		Type:          data.Type,
		Settings:      data.Settings,
		Class:         data.Class,
		UnitsPosition: data.UnitsPosition,
		IsDeleted:     data.IsDeleted,
		CreatedAt:     data.CreatedAt,
		UpdatedAt:     data.UpdatedAt,
	}
}
