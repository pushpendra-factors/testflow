package handler

import (
	mid "factors/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

const ROUTE_SDK_ROOT = "/sdk"
const ROUTE_PROJECTS_ROOT = "/projects"
const ROUTE_INTEGRATIONS_ROOT = "/integrations"

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
	r.POST("/agents/activate", mid.ValidateAgentActivationRequest(), AgentActivate)
	r.GET("/agents/info", mid.SetLoggedInAgent(), AgentInfo)

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

	authRouteGroup.GET("/:project_id/settings", GetProjectSettingHandler)
	authRouteGroup.PUT("/:project_id/settings", UpdateProjectSettingsHandler)
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
	authRouteGroup.POST("/:project_id/agents/invite", AgentInvite)
}

func InitSDKRoutes(r *gin.Engine) {
	sdkRouteGroup := r.Group(ROUTE_SDK_ROOT)
	sdkRouteGroup.Use(mid.SetScopeProjectIdByToken())

	sdkRouteGroup.POST("/event/track", SDKTrackHandler)
	sdkRouteGroup.POST("/user/identify", SDKIdentifyHandler)
	sdkRouteGroup.POST("/user/add_properties", SDKAddUserPropertiesHandler)
	sdkRouteGroup.GET("/project/get_settings", SDKGetProjectSettingsHandler)
}

func InitIntRoutes(r *gin.Engine) {
	intRouteGroup := r.Group(ROUTE_INTEGRATIONS_ROOT)

	intRouteGroup.POST("/segment",
		mid.SetScopeProjectIdByPrivateToken(),
		IntSegmentHandler)
}
