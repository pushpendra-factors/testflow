package handler

import (
	Middleware "factors/middleware"

	"github.com/gin-gonic/gin"
)

const ROUTE_GROUP_PREFIX_SDK = "/sdk"

func InitAppRoutes(r *gin.Engine) {
	r.POST("/projects", CreateProjectHandler)
	r.GET("/projects", GetProjectsHandler)
	r.GET("/projects/:project_id/event_names", GetEventNamesHandler)
	r.GET("/projects/:project_id/event_names/:event_name/properties", GetEventPropertiesHandler)
	r.GET("/projects/:project_id/event_names/:event_name/properties/:property_name/values", GetEventPropertyValuesHandler)
	r.POST("/projects/:project_id/users/:user_id/events", CreateEventHandler)
	r.GET("/projects/:project_id/users/:user_id/events/:id", GetEventHandler)
	r.POST("/projects/:project_id/users", CreateUserHandler)
	r.GET("/projects/:project_id/users/:user_id", GetUserHandler)
	r.GET("/projects/:project_id/users", GetUsersHandler)
	r.POST("/projects/:project_id/factor", FactorHandler)
}

func InitSDKRoutes(r *gin.Engine) {
	sdkRG := r.Group(ROUTE_GROUP_PREFIX_SDK)
	sdkRG.Use(Middleware.SetProjectScopeByTokenMiddleware())

	sdkRG.POST("/event/track", SDKTrackHandler)
	sdkRG.POST("/user/identify", SDKIdentifyHandler)
	sdkRG.POST("/user/add_properties", SDKAddUserPropertiesHandler)
	sdkRG.GET("/project/get_settings", SDKGetProjectSettings)
}
