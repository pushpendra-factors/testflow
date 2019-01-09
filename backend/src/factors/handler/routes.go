package handler

import (
	Mid "factors/middleware"

	"github.com/gin-gonic/gin"
)

const ROUTE_SDK_ROOT = "/sdk"
const ROUTE_PROJECTS_ROOT = "/projects"

func InitAppRoutes(r *gin.Engine) {

	// Route not allowed for public access.
	r.POST(ROUTE_PROJECTS_ROOT,
		Mid.DenyPublicAccess(),
		CreateProjectHandler)

	r.GET(ROUTE_PROJECTS_ROOT,
		Mid.SetScopeAuthorizedProjectsBySubdomain(),
		GetProjectsHandler)

	// Auth route group with authentication an authorization middleware.
	authRouteGroup := r.Group(ROUTE_PROJECTS_ROOT)
	authRouteGroup.Use(Mid.SetScopeAuthorizedProjectsBySubdomain())
	authRouteGroup.Use(Mid.IsAuthorized())

	authRouteGroup.GET("/:project_id/settings", GetProjectSettingHandler)
	authRouteGroup.PUT("/:project_id/settings", UpdateProjectSettingsHandler)
	authRouteGroup.GET("/:project_id/event_names", GetEventNamesHandler)
	authRouteGroup.GET("/:project_id/models", GetProjectModelIntervalsHandler)
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
}

func InitSDKRoutes(r *gin.Engine) {
	sdkRouteGroup := r.Group(ROUTE_SDK_ROOT)
	sdkRouteGroup.Use(Mid.SetScopeProjectIdByToken())

	sdkRouteGroup.POST("/event/track", SDKTrackHandler)
	sdkRouteGroup.POST("/user/identify", SDKIdentifyHandler)
	sdkRouteGroup.POST("/user/add_properties", SDKAddUserPropertiesHandler)
	sdkRouteGroup.GET("/project/get_settings", SDKGetProjectSettings)
}
