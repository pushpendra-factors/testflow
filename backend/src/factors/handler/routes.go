package handler

import (
	C "factors/config"
	IH "factors/handler/internal"
	V1 "factors/handler/v1"
	mid "factors/middleware"
	M "factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"

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
const ROUTE_VERSION_V1_WITHOUT_SLASH = "v1"
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

	r.Use(mid.RestrictHTTPAccess())

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
	r.GET(routePrefix+"/"+ROUTE_PROJECTS_ROOT_V1,
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		V1.GetProjectsHandler)

	// Feature Gates Auth Group
	// authRouteGroup := r.Group(routePrefix + ROUTE_PROJECTS_ROOT)
	// authRouteGroup.Use(mid.SetLoggedInAgent())
	// authRouteGroup.Use(mid.SetAuthorizedProjectsByLoggedInAgent())
	// authRouteGroup.Use(mid.ValidateLoggedInAgentHasAccessToRequestProject())
	// authRouteGroup.Use(mid.FeatureMiddleware())

	// Auth route group with authentication an authorization middleware.
	authRouteGroup := r.Group(routePrefix + ROUTE_PROJECTS_ROOT)
	authRouteGroup.Use(mid.SetLoggedInAgent())
	authRouteGroup.Use(mid.SetAuthorizedProjectsByLoggedInAgent())
	authRouteGroup.Use(mid.ValidateLoggedInAgentHasAccessToRequestProject())

	// Shareable link routes
	shareRouteGroup := r.Group(routePrefix + ROUTE_PROJECTS_ROOT)
	shareRouteGroup.Use(mid.ValidateAccessToSharedEntity(M.ShareableURLEntityTypeQuery))

	shareRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/query", responseWrapper(EventsQueryHandler))
	shareRouteGroup.POST("/:project_id/query", responseWrapper(QueryHandler))
	shareRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/attribution/query", responseWrapper(V1.AttributionHandlerV1))
	shareRouteGroup.POST("/:project_id/attribution/query", responseWrapper(AttributionHandler))
	shareRouteGroup.POST("/:project_id/profiles/query", responseWrapper(ProfilesQueryHandler))
	shareRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/kpi/query", responseWrapper(V1.ExecuteKPIQueryHandler))

	//Six Signal Report
	shareSixSignalRouteGroup := r.Group(routePrefix + ROUTE_PROJECTS_ROOT)
	shareSixSignalRouteGroup.Use(mid.ValidateAccessToSharedEntity(M.ShareableURLEntityTypeSixSignal))
	shareSixSignalRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/sixsignal", responseWrapper(GetSixSignalReportHandler))
	shareSixSignalRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/sixsignal/publicreport", responseWrapper(GetSixSignalPublicReportHandler))
	shareSixSignalRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/sixsignal/report/pageviews", responseWrapper(GetPageViewForSixSignalReport))
	authRouteGroup.POST("/:project_id/sixsignal/share",  mid.FeatureMiddleware([]string{M.FEATURE_SIX_SIGNAL_REPORT}), stringifyWrapper(CreateSixSignalShareableURLHandler))
	authRouteGroup.POST("/:project_id/sixsignal/add_email",  mid.FeatureMiddleware([]string{M.FEATURE_SIX_SIGNAL_REPORT}), stringifyWrapper(AddSixSignalEmailIDHandler))
	authRouteGroup.GET("/:project_id/sixsignal/date_list",  mid.FeatureMiddleware([]string{M.FEATURE_SIX_SIGNAL_REPORT}), stringifyWrapper(FetchListofDatesForSixSignalReport))

	// Dashboard endpoints
	authRouteGroup.GET("/:project_id/dashboards", mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), stringifyWrapper(GetDashboardsHandler))
	authRouteGroup.POST("/:project_id/dashboards",  mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), stringifyWrapper(CreateDashboardHandler))
	authRouteGroup.PUT("/:project_id/dashboards/:dashboard_id",  mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), UpdateDashboardHandler)
	authRouteGroup.GET("/:project_id/dashboards/:dashboard_id/units", mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), stringifyWrapper(GetDashboardUnitsHandler))
	authRouteGroup.POST("/:project_id/dashboards/:dashboard_id/units",  mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), stringifyWrapper(CreateDashboardUnitHandler))
	authRouteGroup.PUT("/:project_id/dashboards/:dashboard_id/units/:unit_id",  mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), UpdateDashboardUnitHandler)
	authRouteGroup.DELETE("/:project_id/dashboards/:dashboard_id/units/:unit_id",  mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), DeleteDashboardUnitHandler)
	authRouteGroup.POST("/:project_id/dashboard/:dashboard_id/units/query/web_analytics", mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}),
		DashboardUnitsWebAnalyticsQueryHandler)

	// Offline Touch Point rules
	authRouteGroup.GET("/:project_id/otp_rules", mid.FeatureMiddleware([]string{M.FEATURE_OFFLINE_TOUCHPOINTS}), responseWrapper(GetOTPRuleHandler))
	authRouteGroup.POST("/:project_id/otp_rules",  mid.FeatureMiddleware([]string{M.FEATURE_OFFLINE_TOUCHPOINTS}), responseWrapper(CreateOTPRuleHandler))
	authRouteGroup.PUT("/:project_id/otp_rules/:rule_id",  mid.FeatureMiddleware([]string{M.FEATURE_OFFLINE_TOUCHPOINTS}), responseWrapper(UpdateOTPRuleHandler))
	authRouteGroup.GET("/:project_id/otp_rules/:rule_id", mid.FeatureMiddleware([]string{M.FEATURE_OFFLINE_TOUCHPOINTS}), responseWrapper(SearchOTPRuleHandler))
	authRouteGroup.DELETE("/:project_id/otp_rules/:rule_id",  mid.FeatureMiddleware([]string{M.FEATURE_OFFLINE_TOUCHPOINTS}), responseWrapper(DeleteOTPRuleHandler))

	// Dashboard templates

	authRouteGroup.POST("/:project_id/dashboard_template/:id/trigger",  mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), GenerateDashboardFromTemplateHandler)
	authRouteGroup.POST("/:project_id/dashboards/:dashboard_id/trigger",  mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), GenerateTemplateFromDashboardHandler)

	authRouteGroup.GET("/:project_id/queries", mid.FeatureMiddleware([]string{M.FEATURE_SAVED_QUERIES}), stringifyWrapper(GetQueriesHandler))
	authRouteGroup.POST("/:project_id/queries",  mid.FeatureMiddleware([]string{M.FEATURE_SAVED_QUERIES}), stringifyWrapper(CreateQueryHandler))
	authRouteGroup.PUT("/:project_id/queries/:query_id",  mid.FeatureMiddleware([]string{M.FEATURE_SAVED_QUERIES}), UpdateSavedQueryHandler)
	authRouteGroup.DELETE("/:project_id/queries/:query_id",  mid.FeatureMiddleware([]string{M.FEATURE_SAVED_QUERIES}), DeleteSavedQueryHandler)
	authRouteGroup.GET("/:project_id/queries/search", mid.FeatureMiddleware([]string{M.FEATURE_SAVED_QUERIES}), stringifyWrapper(SearchQueriesHandler))

	authRouteGroup.GET("/:project_id/models", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), GetProjectModelsHandler)
	authRouteGroup.GET("/:project_id/filters", mid.FeatureMiddleware([]string{M.FEATURE_FILTERS}), GetFiltersHandler)
	authRouteGroup.POST("/:project_id/filters",  mid.FeatureMiddleware([]string{M.FEATURE_FILTERS}), CreateFilterHandler)
	authRouteGroup.PUT("/:project_id/filters/:filter_id",  mid.FeatureMiddleware([]string{M.FEATURE_FILTERS}), UpdateFilterHandler)
	authRouteGroup.DELETE("/:project_id/filters/:filter_id",  mid.FeatureMiddleware([]string{M.FEATURE_FILTERS}), DeleteFilterHandler)
	authRouteGroup.POST("/:project_id/factor", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), FactorHandler)

	// shareable url endpoints
	authRouteGroup.GET("/:project_id/shareable_url", mid.FeatureMiddleware([]string{M.FEATURE_SHAREABLE_URL}), GetShareableURLsHandler)
	authRouteGroup.POST("/:project_id/shareable_url", mid.FeatureMiddleware([]string{M.FEATURE_SHAREABLE_URL}), CreateShareableURLHandler)
	authRouteGroup.DELETE("/:project_id/shareable_url/:share_id",  mid.FeatureMiddleware([]string{M.FEATURE_SHAREABLE_URL}), DeleteShareableURLHandler)
	authRouteGroup.DELETE("/:project_id/shareable_url/revoke/:query_id",  mid.FeatureMiddleware([]string{M.FEATURE_SHAREABLE_URL}), RevokeShareableURLHandler)

	// v1 Dashboard endpoints
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/dashboards/multi/:dashboard_ids/units",  mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), stringifyWrapper(CreateDashboardUnitForMultiDashboardsHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/dashboards/queries/:dashboard_id/units",  mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), stringifyWrapper(CreateDashboardUnitsForMultipleQueriesHandler))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/dashboards/:dashboard_id/units/multi/:unit_ids",  mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), DeleteMultiDashboardUnitHandler)
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/dashboards/:dashboard_id",  mid.FeatureMiddleware([]string{M.FEATURE_DASHBOARD}), DeleteDashboardHandler)

	// attribution V1 endpoints
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/attribution/queries", mid.FeatureMiddleware([]string{M.FEATURE_ATTRIBUTION}), stringifyWrapper(V1.GetAttributionQueriesHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/attribution/queries",  mid.FeatureMiddleware([]string{M.FEATURE_ATTRIBUTION}), stringifyWrapper(V1.CreateAttributionV1QueryAndSaveToDashboardHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/attribution/dashboards", mid.FeatureMiddleware([]string{M.FEATURE_ATTRIBUTION}), stringifyWrapper(V1.GetOrCreateAttributionV1DashboardHandler))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/attribution/dashboards/:dashboard_id/units/:unit_id/query/:query_id",  mid.FeatureMiddleware([]string{M.FEATURE_ATTRIBUTION}), V1.DeleteAttributionDashboardUnitAndQueryHandler)

	// v1 custom metrics - admin/settings side.
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics/config/v1", mid.FeatureMiddleware([]string{M.FEATURE_CUSTOM_METRICS, M.CONF_CUSTOM_KPIS}), V1.GetCustomMetricsConfigV1)
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics",  mid.FeatureMiddleware([]string{M.FEATURE_CUSTOM_METRICS, M.CONF_CUSTOM_KPIS}), responseWrapper(V1.CreateCustomMetric))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics", mid.FeatureMiddleware([]string{M.FEATURE_CUSTOM_METRICS, M.CONF_CUSTOM_KPIS}), responseWrapper(V1.GetCustomMetrics))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics/:id",  mid.FeatureMiddleware([]string{M.FEATURE_CUSTOM_METRICS, M.CONF_CUSTOM_KPIS}), responseWrapper(V1.DeleteCustomMetrics))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/custom_metrics/prebuilt/add_missing", mid.FeatureMiddleware([]string{M.FEATURE_CUSTOM_METRICS, M.CONF_CUSTOM_KPIS}), responseWrapper(V1.CreateMissingPreBuiltCustomKPI))

	// v1 CRM And Smart Event endpoints
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_event", mid.FeatureMiddleware([]string{M.FEATURE_SMART_EVENTS}), GetSmartEventFiltersHandler)
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/smart_event",  mid.FeatureMiddleware([]string{M.FEATURE_SMART_EVENTS}), CreateSmartEventFilterHandler)
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/smart_event",  mid.FeatureMiddleware([]string{M.FEATURE_SMART_EVENTS}), UpdateSmartEventFilterHandler)
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/smart_event",  mid.FeatureMiddleware([]string{M.FEATURE_SMART_EVENTS}), responseWrapper(DeleteSmartEventFilterHandler))

	// template
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/templates/:type/config", mid.FeatureMiddleware([]string{M.FEATURE_TEMPLATES}), responseWrapper(V1.GetTemplateConfigHandler))
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/templates/:type/config",  mid.FeatureMiddleware([]string{M.FEATURE_TEMPLATES}), responseWrapper(V1.UpdateTemplateConfigHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/templates/:type/query", mid.FeatureMiddleware([]string{M.FEATURE_TEMPLATES}), responseWrapper(V1.ExecuteTemplateQueryHandler))

	// smart Properties
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/config/:object_type", mid.FeatureMiddleware([]string{M.FEATURE_SMART_PROPERTIES}), responseWrapper(GetSmartPropertyRulesConfigHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules",  mid.FeatureMiddleware([]string{M.FEATURE_SMART_PROPERTIES}), responseWrapper(CreateSmartPropertyRulesHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules", mid.FeatureMiddleware([]string{M.FEATURE_SMART_PROPERTIES}), responseWrapper(GetSmartPropertyRulesHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules/:rule_id", mid.FeatureMiddleware([]string{M.FEATURE_SMART_PROPERTIES}), responseWrapper(GetSmartPropertyRuleByRuleIDHandler))
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules/:rule_id",  mid.FeatureMiddleware([]string{M.FEATURE_SMART_PROPERTIES}), responseWrapper(UpdateSmartPropertyRulesHandler))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/smart_properties/rules/:rule_id",  mid.FeatureMiddleware([]string{M.FEATURE_SMART_PROPERTIES}), responseWrapper(DeleteSmartPropertyRulesHandler))

	// content groups
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/contentgroup",  mid.FeatureMiddleware([]string{M.FEATURE_CONTENT_GROUPS}), responseWrapper(V1.CreateContentGroupHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/contentgroup", mid.FeatureMiddleware([]string{M.FEATURE_CONTENT_GROUPS}), responseWrapper(V1.GetContentGroupHandler))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/contentgroup/:id", mid.FeatureMiddleware([]string{M.FEATURE_CONTENT_GROUPS}), responseWrapper(V1.GetContentGroupByIDHandler))
	authRouteGroup.PUT("/:project_id"+ROUTE_VERSION_V1+"/contentgroup/:id",  mid.FeatureMiddleware([]string{M.FEATURE_CONTENT_GROUPS}), responseWrapper(V1.UpdateContentGroupHandler))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/contentgroup/:id",  mid.FeatureMiddleware([]string{M.FEATURE_CONTENT_GROUPS}), responseWrapper(V1.DeleteContentGroupHandler))
	// TODO
	// Scope this with Project Admin
	authRouteGroup.GET("/:project_id/clickable_elements",  mid.FeatureMiddleware([]string{M.FEATURE_CLICKABLE_ELEMENTS}), GetClickableElementsHandler)
	authRouteGroup.GET("/:project_id/clickable_elements/:id/toggle",  mid.FeatureMiddleware([]string{M.FEATURE_CLICKABLE_ELEMENTS}), ToggleClickableElementHandler)

	// LeadSquared
	authRouteGroup.PUT("/:project_id/leadsquaredsettings",   mid.FeatureMiddleware([]string{M.FEATURE_LEADSQUARED}), UpdateLeadSquaredConfigHandler)
	authRouteGroup.DELETE("/:project_id/leadsquaredsettings/remove",   mid.FeatureMiddleware([]string{M.FEATURE_LEADSQUARED}), RemoveLeadSquaredConfigHandler)

	// Tracked Events
	authRouteGroup.POST("/:project_id/v1/factors/tracked_event",  mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.CreateFactorsTrackedEventsHandler)
	authRouteGroup.DELETE("/:project_id/v1/factors/tracked_event/remove",  mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.RemoveFactorsTrackedEventsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/tracked_event", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.GetAllFactorsTrackedEventsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/grouped_tracked_event", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.GetAllGroupedFactorsTrackedEventsHandler)

	// Tracked User Property
	authRouteGroup.POST("/:project_id/v1/factors/tracked_user_property",  mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.CreateFactorsTrackedUserPropertyHandler)
	authRouteGroup.DELETE("/:project_id/v1/factors/tracked_user_property/remove",  mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.RemoveFactorsTrackedUserPropertyHandler)
	authRouteGroup.GET("/:project_id/v1/factors/tracked_user_property", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.GetAllFactorsTrackedUserPropertiesHandler)

	// Goals
	authRouteGroup.POST("/:project_id/v1/factors/goals",  mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.CreateFactorsGoalsHandler)
	authRouteGroup.DELETE("/:project_id/v1/factors/goals/remove",  mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.RemoveFactorsGoalsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/goals", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.GetAllFactorsGoalsHandler)
	authRouteGroup.PUT("/:project_id/v1/factors/goals/update",  mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.UpdateFactorsGoalsHandler)
	authRouteGroup.GET("/:project_id/v1/factors/goals/search", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.SearchFactorsGoalHandler)
	authRouteGroup.POST("/:project_id/v1/factor", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.PostFactorsHandler)
	authRouteGroup.POST("/:project_id/v1/factor/compare", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.PostFactorsCompareHandler)
	authRouteGroup.POST("/:project_id/v1/events/displayname",  mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), responseWrapper(V1.CreateDisplayNamesHandler))
	authRouteGroup.GET("/:project_id/v1/events/displayname", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), responseWrapper(V1.GetAllDistinctEventProperties))
	authRouteGroup.GET("/:project_id/v1/factor", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.GetFactorsHandler)
	authRouteGroup.GET("/:project_id/v1/factor/model_metadata", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.GetModelMetaData)

	authRouteGroup.GET("/:project_id/insights", mid.FeatureMiddleware([]string{M.FEATURE_WEEKLY_INSIGHTS}), responseWrapper(V1.GetWeeklyInsightsHandler))
	authRouteGroup.GET("/:project_id/weekly_insights_metadata", mid.FeatureMiddleware([]string{M.FEATURE_WEEKLY_INSIGHTS}), responseWrapper(V1.GetWeeklyInsightsMetadata))
	authRouteGroup.POST("/:project_id/feedback",  mid.FeatureMiddleware([]string{M.FEATURE_WEEKLY_INSIGHTS}), V1.PostFeedbackHandler)

	// bingads integration
	authRouteGroup.POST("/:project_id/v1/bingads",  mid.FeatureMiddleware([]string{M.FEATURE_BING_ADS}), responseWrapper(V1.CreateBingAdsIntegration))
	authRouteGroup.DELETE("/:project_id/v1/bingads/disable",  mid.FeatureMiddleware([]string{M.FEATURE_BING_ADS}), responseWrapper(V1.DisableBingAdsIntegration))
	authRouteGroup.GET("/:project_id/v1/bingads", mid.FeatureMiddleware([]string{M.FEATURE_BING_ADS}), responseWrapper(V1.GetBingAdsIntegration))
	authRouteGroup.PUT("/:project_id/v1/bingads/enable",  mid.FeatureMiddleware([]string{M.FEATURE_BING_ADS}), responseWrapper(V1.EnableBingAdsIntegration))

	// marketo integration
	authRouteGroup.POST("/:project_id/v1/marketo",  mid.FeatureMiddleware([]string{M.FEATURE_MARKETO}), responseWrapper(V1.CreateMarketoIntegration))
	authRouteGroup.DELETE("/:project_id/v1/marketo/disable",  mid.FeatureMiddleware([]string{M.FEATURE_MARKETO}), responseWrapper(V1.DisableMarketoIntegration))
	authRouteGroup.GET("/:project_id/v1/marketo", mid.FeatureMiddleware([]string{M.FEATURE_MARKETO}), responseWrapper(V1.GetMarketoIntegration))
	authRouteGroup.PUT("/:project_id/v1/marketo/enable",  mid.FeatureMiddleware([]string{M.FEATURE_MARKETO}), responseWrapper(V1.EnableMarketoIntegration))

	// alerts
	authRouteGroup.POST("/:project_id/v1/alerts",  mid.FeatureMiddleware([]string{M.FEATURE_KPI_ALERTS}), responseWrapper(V1.CreateAlertHandler))
	authRouteGroup.GET("/:project_id/v1/alerts", mid.FeatureMiddleware([]string{M.FEATURE_KPI_ALERTS}), responseWrapper(V1.GetAlertsHandler))
	authRouteGroup.GET("/:project_id/v1/alerts/:id", mid.FeatureMiddleware([]string{M.FEATURE_KPI_ALERTS}), responseWrapper(V1.GetAlertByIDHandler))
	authRouteGroup.DELETE("/:project_id/v1/alerts/:id",  mid.FeatureMiddleware([]string{M.FEATURE_KPI_ALERTS}), responseWrapper(V1.DeleteAlertHandler))
	authRouteGroup.PUT("/:project_id/v1/alerts/:id",  mid.FeatureMiddleware([]string{M.FEATURE_KPI_ALERTS}), responseWrapper(V1.EditAlertHandler))
	authRouteGroup.GET("/:project_id/v1/all_alerts", mid.FeatureMiddleware([]string{M.FEATURE_EVENT_BASED_ALERTS, M.FEATURE_KPI_ALERTS}), responseWrapper(V1.GetAllAlertsInOneHandler))

	// slack
	authRouteGroup.POST("/:project_id/slack/auth",  mid.FeatureMiddleware([]string{M.FEATURE_SLACK, M.INT_SLACK}), V1.SlackAuthRedirectHandler)
	authRouteGroup.GET("/:project_id/slack/channels",  mid.FeatureMiddleware([]string{M.FEATURE_SLACK, M.INT_SLACK}), V1.GetSlackChannelsListHandler)
	authRouteGroup.DELETE("/:project_id/slack/delete",  mid.FeatureMiddleware([]string{M.FEATURE_SLACK, M.INT_SLACK}), V1.DeleteSlackIntegrationHandler)
	authRouteGroup.POST("/:project_id/v1/alerts/send_now",  mid.FeatureMiddleware([]string{M.FEATURE_SLACK, M.INT_SLACK}), V1.QuerySendNowHandler)

	// Timeline
	authRouteGroup.POST("/:project_id/v1/profiles/users", mid.FeatureMiddleware([]string{M.FEATURE_PEOPLE_PROFILES}), responseWrapper(V1.GetProfileUsersHandler))
	authRouteGroup.GET("/:project_id/v1/profiles/users/:id", mid.FeatureMiddleware([]string{M.FEATURE_PEOPLE_PROFILES}), responseWrapper(V1.GetProfileUserDetailsHandler))
	authRouteGroup.POST("/:project_id/v1/profiles/accounts", mid.FeatureMiddleware([]string{M.FEATURE_ACCOUNT_PROFILES}), responseWrapper(V1.GetProfileAccountsHandler))
	authRouteGroup.GET("/:project_id/v1/profiles/accounts/:group/:id", mid.FeatureMiddleware([]string{M.FEATURE_ACCOUNT_PROFILES}), responseWrapper(V1.GetProfileAccountDetailsHandler))
	authRouteGroup.POST("/:project_id/segments", mid.FeatureMiddleware([]string{M.FEATURE_SEGMENT}), CreateSegmentHandler)
	authRouteGroup.GET("/:project_id/segments", mid.FeatureMiddleware([]string{M.FEATURE_SEGMENT}), responseWrapper(GetSegmentsHandler))
	authRouteGroup.GET("/:project_id/segments/:id", mid.FeatureMiddleware([]string{M.FEATURE_SEGMENT}), responseWrapper(GetSegmentByIdHandler))
	authRouteGroup.PUT("/:project_id/segments/:id", mid.FeatureMiddleware([]string{M.FEATURE_SEGMENT}), UpdateSegmentHandler)
	authRouteGroup.DELETE("/:project_id/segments/:id", mid.FeatureMiddleware([]string{M.FEATURE_SEGMENT}), DeleteSegmentByIdHandler)

	// path analysis
	authRouteGroup.GET("/:project_id/v1/pathanalysis", mid.FeatureMiddleware([]string{M.FEATURE_PATH_ANALYSIS}), responseWrapper(V1.GetPathAnalysisEntityHandler))
	authRouteGroup.POST("/:project_id/v1/pathanalysis", mid.FeatureMiddleware([]string{M.FEATURE_PATH_ANALYSIS}), responseWrapper(V1.CreatePathAnalysisEntityHandler))
	authRouteGroup.DELETE("/:project_id/v1/pathanalysis/:id", mid.FeatureMiddleware([]string{M.FEATURE_PATH_ANALYSIS}), V1.DeleteSavedPathAnalysisEntityHandler)
	authRouteGroup.GET("/:project_id/v1/pathanalysis/:id", mid.FeatureMiddleware([]string{M.FEATURE_PATH_ANALYSIS}), responseWrapper(V1.GetPathAnalysisData))

	//explainV2
	authRouteGroup.GET("/:project_id/v1/explainV2", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.GetFactorsHandlerV2)
	authRouteGroup.GET("/:project_id/v1/explainV2/goals", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), responseWrapper(V1.GetExplainV2EntityHandler))
	authRouteGroup.POST("/:project_id/v1/explainV2", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.PostFactorsHandlerV2)
	authRouteGroup.POST("/:project_id/v1/explainV2/job", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), responseWrapper(V1.CreateExplainV2EntityHandler))
	authRouteGroup.DELETE("/:project_id/v1/explainV2/:id", mid.FeatureMiddleware([]string{M.FEATURE_EXPLAIN}), V1.DeleteSavedExplainV2EntityHandler)

	//acc scoring
	authRouteGroup.PUT("/:project_id/v1/accscore/weights", mid.FeatureMiddleware([]string{M.FEATURE_ACCOUNT_SCORING}), responseWrapper(V1.UpdateAccScoreWeights))
	authRouteGroup.GET("/:project_id/v1/accscore/score/user", mid.FeatureMiddleware([]string{M.FEATURE_ACCOUNT_SCORING}), responseWrapper(V1.GetUserScore))
	authRouteGroup.GET("/:project_id/v1/accscore/score/user/all", mid.FeatureMiddleware([]string{M.FEATURE_ACCOUNT_SCORING}), responseWrapper(V1.GetAllUsersScores))
	authRouteGroup.GET("/:project_id/v1/accscore/score/account", mid.FeatureMiddleware([]string{M.FEATURE_ACCOUNT_SCORING}), responseWrapper(V1.GetAccountScores))
	authRouteGroup.GET("/:project_id/v1/accscore/score/paccount/", mid.FeatureMiddleware([]string{M.FEATURE_ACCOUNT_SCORING}), responseWrapper(V1.GetPerAccountScore))

	// event trigger alert
	authRouteGroup.GET("/:project_id/v1/eventtriggeralert", mid.FeatureMiddleware([]string{M.FEATURE_EVENT_BASED_ALERTS}), responseWrapper(V1.GetEventTriggerAlertsByProjectHandler))
	authRouteGroup.POST("/:project_id/v1/eventtriggeralert", mid.FeatureMiddleware([]string{M.FEATURE_EVENT_BASED_ALERTS}), responseWrapper(V1.CreateEventTriggerAlertHandler))
	authRouteGroup.DELETE("/:project_id/v1/eventtriggeralert/:id", mid.FeatureMiddleware([]string{M.FEATURE_EVENT_BASED_ALERTS}), V1.DeleteEventTriggerAlertHandler)
	authRouteGroup.PUT("/:project_id/v1/eventtriggeralert/:id", mid.FeatureMiddleware([]string{M.FEATURE_EVENT_BASED_ALERTS}), responseWrapper(V1.EditEventTriggerAlertHandler))
	authRouteGroup.PUT("/:project_id/v1/eventtriggeralert/test_wh", mid.FeatureMiddleware([]string{M.FEATURE_EVENT_BASED_ALERTS}), responseWrapper(V1.TestWebhookforEventTriggerAlerts))
	authRouteGroup.GET("/:project_id/v1/eventtriggeralert/:id", mid.FeatureMiddleware([]string{M.FEATURE_EVENT_BASED_ALERTS}), responseWrapper(V1.GetInternalStatusForEventTriggerAlertHandler))
	authRouteGroup.PUT("/:project_id/v1/eventtriggeralert/:id/status", mid.FeatureMiddleware([]string{M.FEATURE_EVENT_BASED_ALERTS}), responseWrapper(V1.UpdateEventTriggerAlertInternalStatusHandler))

	// teams
	authRouteGroup.POST("/:project_id/teams/auth",  mid.FeatureMiddleware([]string{M.FEATURE_TEAMS, M.INT_TEAMS}), V1.TeamsAuthRedirectHandler)
	authRouteGroup.GET("/:project_id/teams/get_teams",  mid.FeatureMiddleware([]string{M.FEATURE_TEAMS}), V1.GetAllTeamsHandler)
	authRouteGroup.GET("/:project_id/teams/channels",  mid.FeatureMiddleware([]string{M.FEATURE_TEAMS}), V1.GetTeamsChannelsHandler)
	authRouteGroup.DELETE("/:project_id/teams/delete",  mid.FeatureMiddleware([]string{M.FEATURE_TEAMS}), V1.DeleteTeamsIntegrationHandler)
	// Upload
	authRouteGroup.POST("/:project_id/uploadlist", mid.FeatureMiddleware([]string{M.FEATURE_EVENT_BASED_ALERTS}), V1.UploadListForFilters)

	authRouteGroup.GET("/:project_id/agents", GetProjectAgentsHandler)
	authRouteGroup.POST("/:project_id/agents/invite",  AgentInvite)
	authRouteGroup.POST("/:project_id/agents/batchinvite",  AgentInviteBatch)
	authRouteGroup.PUT("/:project_id/agents/remove",  RemoveProjectAgent)
	authRouteGroup.PUT("/:project_id/agents/update",  AgentUpdate)
	authRouteGroup.GET("/:project_id/settings",  GetProjectSettingHandler)
	authRouteGroup.GET("/:project_id/v1/settings",  V1.GetProjectSettingHandler)
	authRouteGroup.PUT("/:project_id/settings",  UpdateProjectSettingsHandler)
	authRouteGroup.PUT("/:project_id",  EditProjectHandler)
	authRouteGroup.GET("/:project_id/event_names", GetEventNamesHandler)
	authRouteGroup.GET("/:project_id/user/event_names", GetEventNamesByUserHandler)
	authRouteGroup.GET(":project_id/groups/:group_name/event_names", GetEventNamesByGroupHandler)
	authRouteGroup.GET("/:project_id/event_names/:event_name/properties", GetEventPropertiesHandler)
	authRouteGroup.GET("/:project_id/event_name_category/category/properties", V1.GetPropertiesByEventCategoryType)
	authRouteGroup.GET("/:project_id/event_names/:event_name/properties/:property_name/values", GetEventPropertyValuesHandler)
	authRouteGroup.GET("/:project_id/groups", GetGroupsHandler)
	authRouteGroup.GET("/:project_id/groups/:group_name/properties", GetGroupPropertiesHandler)
	authRouteGroup.GET("/:project_id/groups/:group_name/properties/:property_name/values", GetGroupPropertyValuesHandler)
	authRouteGroup.GET("/:project_id/users", GetUsersHandler)
	authRouteGroup.GET("/:project_id/users/:user_id", GetUserHandler)
	authRouteGroup.GET("/:project_id/user_properties", GetUserPropertiesHandler)
	authRouteGroup.GET("/:project_id/user_properties/:property_name/values", GetUserPropertyValuesHandler)
	authRouteGroup.GET("/:project_id/channel_grouping_properties", GetChannelGroupingPropertiesHandler)
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/crm/:crm_source/:object_type/properties", GetCRMObjectPropertiesHandler)
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/crm/:crm_source/:object_type/properties/:property_name/values", GetCRMObjectValuesByPropertyNameHandler)
	// v1 KPI endpoints
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/kpi/config", responseWrapper(V1.GetKPIConfigHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/kpi/filter_values", responseWrapper(V1.GetKPIFilterValuesHandler))
	authRouteGroup.GET("/:project_id/v1/kpi/:custom_event_kpi/properties", V1.GetPropertiesForCustomKPIEventBased)
	// V1 Routes
	authRouteGroup.GET("/:project_id/v1/event_names", V1.GetEventNamesHandler)
	authRouteGroup.GET("/:project_id/v1/event_names/:type", V1.GetEventNamesByTypeHandler)
	authRouteGroup.GET("/:project_id/v1/agents",  V1.GetProjectAgentsHandler)
	// project analytics
	authRouteGroup.GET("/:project_id/v1/dataobservability/metrics", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.GetAnalyticsMetricsFromStorage))
	authRouteGroup.GET("/:project_id/v1/dataobservability/alerts", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.GetAnalyticsAlertsFromStorage))
	// weekly insights, explain
	authRouteGroup.PUT("/:project_id/v1/weeklyinsights", mid.SetLoggedInAgentInternalOnly(), UpdateWeeklyInsightsHandler)
	authRouteGroup.PUT("/:project_id/v1/explain", mid.SetLoggedInAgentInternalOnly(), UpdateExplainHandler)
	authRouteGroup.PUT("/:project_id/v1/pathanalysis", mid.SetLoggedInAgentInternalOnly(), UpdatePathAnalysisHandler)

	// feature gate
	authRouteGroup.POST("/:project_id/v1/feature_gates", mid.SetLoggedInAgentInternalOnly(), V1.UpdateFeatureStatusHandler)

	authCommonRouteGroup := r.Group(routePrefix + ROUTE_COMMON_ROOT)
	//The dashboard endpoints doesn't have project_id params, hence feature middleware is not added here.
	authCommonRouteGroup.GET("/dashboard_templates/:id/search", SearchTemplateHandler)
	authCommonRouteGroup.GET("/dashboard_templates", GetDashboardTemplatesHandler)
	authCommonRouteGroup.POST("/dashboard_template/create",  CreateTemplateHandler)

	// feature gate v2
	authRouteGroup.GET("/:project_id/v1/features", responseWrapper(V1.GetPlanDetailsForProjectHandler))

	// plans and pricing
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/plan", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.UpdateProjectPlanMappingFieldHandler))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/features/update", mid.SetLoggedInAgentInternalOnly(), responseWrapper(V1.UpdateCustomPlanHandler))
	// property mapping
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/kpi/property_mappings", responseWrapper(V1.CreatePropertyMapping))
	authRouteGroup.GET("/:project_id"+ROUTE_VERSION_V1+"/kpi/property_mappings", responseWrapper(V1.GetPropertyMappings))
	authRouteGroup.DELETE("/:project_id"+ROUTE_VERSION_V1+"/kpi/property_mappings/:id", responseWrapper(V1.DeletePropertyMapping))
	authRouteGroup.POST("/:project_id"+ROUTE_VERSION_V1+"/kpi/property_mappings/commom_properties", responseWrapper(V1.GetCommonPropertyMappings))

	//six signal
	authRouteGroup.POST("/:project_id/sixsignal/email", responseWrapper(SendSixSignalReportViaEmailHandler))
}

func InitSDKServiceRoutes(r *gin.Engine) {
	// Initialize swagger api docs only for development / staging.
	if C.GetConfig().Env != C.PRODUCTION {
		r.GET("/sdk/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	r.Use(mid.AddSecurityResponseHeadersToCustomDomain())

	r.GET("/", SDKStatusHandler) // Default handler for probes.
	r.GET(ROUTE_SDK_ROOT+"/service/status", SDKStatusHandler)
	r.POST(ROUTE_SDK_ROOT+"/service/error", SDKErrorHandler)

	// Robots.txt added to disallow crawling.
	r.GET("/robots.txt", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain", []byte("User-agent: *\nDisallow: *"))
	})

	// Todo(Dinesh): Check integrity of token using encrytion/decryption
	// with secret, on middleware, to avoid spamming queue.

	// Getting project_id is moved to sdk request handler to
	// support queue workers also.
	// sdkRouteGroup.Use(mid.SetScopeProjectIdByToken())

	sdkRouteGroup := r.Group(ROUTE_SDK_ROOT)
	sdkRouteGroup.Use(mid.SetScopeProjectToken())
	sdkRouteGroup.Use(mid.IsBlockedIPByProject())

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
		mid.SetScopeProjectIdByStoreAndSecret(), mid.FeatureMiddleware([]string{M.INT_SHOPFIY}),
		IntShopifyHandler)
	intRouteGroup.POST("/shopify_sdk",
		mid.BlockRequestGracefully(),
		mid.SetScopeProjectIdByToken(), mid.FeatureMiddleware([]string{M.INT_SHOPFIY}),
		IntShopifySDKHandler)

	intRouteGroup.POST("/adwords/enable",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(), mid.FeatureMiddleware([]string{M.INT_ADWORDS}),
		IntEnableAdwordsHandler)

	intRouteGroup.POST("/google_organic/enable",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(), mid.FeatureMiddleware([]string{M.INT_GOOGLE_ORGANIC}),
		IntEnableGoogleOrganicHandler)

	intRouteGroup.POST("/facebook/add_access_token",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(), mid.FeatureMiddleware([]string{M.INT_FACEBOOK}),
		IntFacebookAddAccessTokenHandler)

	intRouteGroup.POST("/linkedin/auth", mid.FeatureMiddleware([]string{M.INT_LINKEDIN}), IntLinkedinAuthHandler)
	intRouteGroup.POST("/linkedin/ad_accounts", mid.FeatureMiddleware([]string{M.INT_LINKEDIN}), IntLinkedinAccountHandler)

	intRouteGroup.POST("/linkedin/add_access_token",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(), mid.FeatureMiddleware([]string{M.INT_LINKEDIN}),
		IntLinkedinAddAccessTokenHandler)

	intRouteGroup.POST("/salesforce/enable",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(), mid.FeatureMiddleware([]string{M.INT_SALESFORCE}),
		IntEnableSalesforceHandler)

	intRouteGroup.POST("/salesforce/auth",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(), mid.FeatureMiddleware([]string{M.INT_SALESFORCE}),
		SalesforceAuthRedirectHandler)
	// salesforce integration.
	intRouteGroup.GET(SalesforceCallbackRoute, mid.FeatureMiddleware([]string{M.INT_SALESFORCE}),
		SalesforceCallbackHandler)

	intRouteGroup.POST("/hubspot/auth",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(), mid.FeatureMiddleware([]string{M.INT_HUBSPOT}),
		HubspotAuthRedirectHandler)

	// hubspot integration.
	intRouteGroup.GET(HubspotCallbackRoute, mid.FeatureMiddleware([]string{M.INT_HUBSPOT}),
		HubspotCallbackHandler)

	intRouteGroup.DELETE("/:project_id/:channel_name",
		mid.SetLoggedInAgent(),
		mid.SetAuthorizedProjectsByLoggedInAgent(),
		 mid.FeatureMiddleware([]string{M.FEATURE_GOOGLE_ADS, M.FEATURE_FACEBOOK, M.FEATURE_LINKEDIN, M.FEATURE_GOOGLE_ORGANIC}),
		IntDeleteHandler)

	intRouteGroup.GET("/slack/callback", mid.FeatureMiddleware([]string{M.INT_SLACK}), V1.SlackCallbackHandler)

	intRouteGroup.GET("/teams/callback", mid.FeatureMiddleware([]string{M.INT_SLACK}), V1.TeamsCallbackHandler)

}

func InitDataServiceRoutes(r *gin.Engine) {
	dataServiceRouteGroup := r.Group(ROUTE_DATA_SERVICE_ROOT)

	//todo @ashhar: merge adwords and google_organic whereever possible
	dataServiceRouteGroup.POST("/adwords/documents/add", mid.FeatureMiddleware([]string{M.DS_ADWORDS}),
		IH.DataServiceAdwordsAddDocumentHandler)
	dataServiceRouteGroup.POST("/adwords/documents/add_multiple", mid.FeatureMiddleware([]string{M.DS_ADWORDS}),
		IH.DataServiceAdwordsAddMultipleDocumentsHandler)

	dataServiceRouteGroup.POST("/adwords/add_refresh_token", mid.FeatureMiddleware([]string{M.DS_ADWORDS}),
		IntAdwordsAddRefreshTokenHandler)

	dataServiceRouteGroup.POST("/adwords/get_refresh_token", mid.FeatureMiddleware([]string{M.DS_ADWORDS}),
		IntAdwordsGetRefreshTokenHandler)

	dataServiceRouteGroup.GET("/adwords/documents/project_last_sync_info", mid.FeatureMiddleware([]string{M.DS_ADWORDS}),
		IH.DataServiceAdwordsGetLastSyncForProjectInfoHandler)

	dataServiceRouteGroup.GET("/adwords/documents/last_sync_info", mid.FeatureMiddleware([]string{M.DS_ADWORDS}),
		IH.DataServiceAdwordsGetLastSyncInfoHandler)

	dataServiceRouteGroup.POST("/google_organic/documents/add", mid.FeatureMiddleware([]string{M.DS_GOOGLE_ORGANIC}),
		IH.DataServiceGoogleOrganicAddDocumentHandler)

	dataServiceRouteGroup.POST("/google_organic/documents/add_multiple", mid.FeatureMiddleware([]string{M.DS_GOOGLE_ORGANIC}),
		IH.DataServiceGoogleOrganicAddMultipleDocumentsHandler)

	dataServiceRouteGroup.POST("/google_organic/add_refresh_token", mid.FeatureMiddleware([]string{M.DS_GOOGLE_ORGANIC}),
		IntGoogleOrganicAddRefreshTokenHandler)

	dataServiceRouteGroup.POST("/google_organic/get_refresh_token", mid.FeatureMiddleware([]string{M.DS_GOOGLE_ORGANIC}),
		IntGoogleOrganicGetRefreshTokenHandler)

	dataServiceRouteGroup.GET("/google_organic/documents/project_last_sync_info", mid.FeatureMiddleware([]string{M.DS_GOOGLE_ORGANIC}),
		IH.DataServiceGoogleOrganicGetLastSyncForProjectInfoHandler)

	dataServiceRouteGroup.GET("/google_organic/documents/last_sync_info", mid.FeatureMiddleware([]string{M.DS_GOOGLE_ORGANIC}),
		IH.DataServiceGoogleOrganicGetLastSyncInfoHandler)

	dataServiceRouteGroup.POST("/hubspot/documents/add", mid.FeatureMiddleware([]string{M.DS_HUBSPOT}),
		IH.DataServiceHubspotAddDocumentHandler)

	dataServiceRouteGroup.POST("/hubspot/documents/add_batch", mid.FeatureMiddleware([]string{M.DS_HUBSPOT}),
		IH.DataServiceHubspotAddBatchDocumentHandler)

	dataServiceRouteGroup.GET("/hubspot/documents/sync_info", mid.FeatureMiddleware([]string{M.DS_HUBSPOT}),
		IH.DataServiceHubspotGetSyncInfoHandler)

	dataServiceRouteGroup.POST("/hubspot/documents/sync_info", mid.FeatureMiddleware([]string{M.DS_HUBSPOT}),
		IH.DataServiceHubspotUpdateSyncInfo)

	dataServiceRouteGroup.GET("/hubspot/documents/types/form", mid.FeatureMiddleware([]string{M.DS_HUBSPOT}),
		IH.DataServiceGetHubspotFormDocumentsHandler)

	dataServiceRouteGroup.GET("/facebook/project/settings", mid.FeatureMiddleware([]string{M.DS_FACEBOOK}),
		IH.DataServiceFacebookGetProjectSettings)

	dataServiceRouteGroup.POST("/facebook/documents/add", mid.FeatureMiddleware([]string{M.DS_FACEBOOK}),
		IH.DataServiceFacebookAddDocumentHandler)

	dataServiceRouteGroup.GET("/facebook/documents/last_sync_info", mid.FeatureMiddleware([]string{M.DS_FACEBOOK}),
		IH.DataServiceFacebookGetLastSyncInfoHandler)

	dataServiceRouteGroup.GET("linkedin/documents/last_sync_info", mid.FeatureMiddleware([]string{M.DS_LINKEDIN}),
		IH.DataServiceLinkedinGetLastSyncInfoHandler)

	dataServiceRouteGroup.POST("/linkedin/documents/add", mid.FeatureMiddleware([]string{M.DS_LINKEDIN}),
		IH.DataServiceLinkedinAddDocumentHandler)

	dataServiceRouteGroup.POST("/linkedin/documents/add_multiple", mid.FeatureMiddleware([]string{M.DS_FACEBOOK}),
		IH.DataServiceLinkedinAddMultipleDocumentsHandler)

	dataServiceRouteGroup.DELETE("/linkedin/documents",
		mid.FeatureMiddleware([]string{M.DS_LINKEDIN}), IH.DataServiceLinkedinDeleteDocumentsHandler)

	dataServiceRouteGroup.PUT("/linkedin/access_token", mid.FeatureMiddleware([]string{}),
		mid.FeatureMiddleware([]string{M.DS_LINKEDIN}), IH.DataServiceLinkedinUpdateAccessToken)

	dataServiceRouteGroup.POST("/metrics", mid.FeatureMiddleware([]string{}),
		mid.FeatureMiddleware([]string{M.DS_METRICS}), IH.DataServiceRecordMetricHandler)

	dataServiceRouteGroup.GET("/linkedin/project/settings", mid.FeatureMiddleware([]string{}),
		mid.FeatureMiddleware([]string{M.DS_LINKEDIN}), IH.DataServiceLinkedinGetProjectSettings)

	dataServiceRouteGroup.GET("/linkedin/project/settings/projects", mid.FeatureMiddleware([]string{}),
		mid.FeatureMiddleware([]string{M.DS_LINKEDIN}), IH.DataServiceLinkedinGetProjectSettingsForProjects)

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
			case []M.Queries:
				queriesResp := make([]M.QueriesString, 0)
				responseObj := data.([]M.Queries)
				for _, query := range responseObj {
					queriesResp = append(queriesResp, ConvertQuery(query))
				}
				c.JSON(statusCode, queriesResp)
				return
			case []*M.Queries:
				queriesResp := make([]M.QueriesString, 0)
				responseObj := data.([]*M.Queries)
				for _, query := range responseObj {
					queriesResp = append(queriesResp, ConvertQuery(*query))
				}
				c.JSON(statusCode, queriesResp)
				return
			case []M.DashboardUnit:
				unitResp := make([]M.DashboardUnitString, 0)
				responseObj := data.([]M.DashboardUnit)
				for _, du := range responseObj {
					unitResp = append(unitResp, ConvertDashboardUnit(du))
				}
				c.JSON(statusCode, unitResp)
				return
			case []*M.DashboardUnit:
				unitResp := make([]M.DashboardUnitString, 0)
				responseObj := data.([]*M.DashboardUnit)
				for _, du := range responseObj {
					unitResp = append(unitResp, ConvertDashboardUnit(*du))
				}
				c.JSON(statusCode, unitResp)
				return
			case []M.Dashboard:
				dashboardResp := make([]M.DashboardString, 0)
				responseObj := data.([]M.Dashboard)
				for _, da := range responseObj {
					dashboardResp = append(dashboardResp, ConvertDashboard(da))
				}
				c.JSON(statusCode, dashboardResp)
				return
			case []*M.Dashboard:
				dashboardResp := make([]M.DashboardString, 0)
				responseObj := data.([]*M.Dashboard)
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
			case M.Queries:
				responseObj := data.(M.Queries)
				c.JSON(statusCode, ConvertQuery(responseObj))
				return
			case *M.Queries:
				responseObj := data.(*M.Queries)
				c.JSON(statusCode, ConvertQuery(*responseObj))
				return
			case M.DashboardUnit:
				responseObj := data.(M.DashboardUnit)
				c.JSON(statusCode, ConvertDashboardUnit(responseObj))
				return
			case *M.DashboardUnit:
				responseObj := data.(*M.DashboardUnit)
				c.JSON(statusCode, ConvertDashboardUnit(*responseObj))
				return
			case M.Dashboard:
				responseObj := data.(M.Dashboard)
				c.JSON(statusCode, ConvertDashboard(responseObj))
				return
			case *M.Dashboard:
				responseObj := data.(*M.Dashboard)
				c.JSON(statusCode, ConvertDashboard(*responseObj))
				return
			default:
				c.JSON(statusCode, data)
				return
			}
		}
	}
}

func ConvertQuery(data M.Queries) M.QueriesString {
	return M.QueriesString{
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

func ConvertDashboardUnit(data M.DashboardUnit) M.DashboardUnitString {
	return M.DashboardUnitString{
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

func ConvertDashboard(data M.Dashboard) M.DashboardString {
	return M.DashboardString{
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
