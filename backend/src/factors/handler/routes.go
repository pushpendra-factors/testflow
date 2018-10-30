package handler

import (
	C "factors/config"
	Middleware "factors/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func InitAppRoutes(r *gin.Engine) {
	// CORS
	if C.IsDevelopment() {
		log.Info("Running in development.")
		config := cors.DefaultConfig()
		config.AllowOrigins = []string{"http://localhost:8080", "http://localhost:3000"}
		r.Use(cors.New(config))
	}

	r.POST("/projects", CreateProjectHandler)
	r.GET("/projects", GetProjectsHandler)
	r.GET("/projects/:project_id/event_names", GetEventNamesHandler)
	r.POST("/projects/:project_id/users/:user_id/events", CreateEventHandler)
	r.GET("/projects/:project_id/users/:user_id/events/:id", GetEventHandler)
	r.POST("/projects/:project_id/users", CreateUserHandler)
	r.GET("/projects/:project_id/users/:user_id", GetUserHandler)
	r.GET("/projects/:project_id/users", GetUsersHandler)
	r.POST("/projects/:project_id/factor", FactorHandler)
}

func InitSDKRoutes(r *gin.Engine) {
	apiRouteGroup := r.Group("/sdk")
	apiRouteGroup.Use(cors.Default()) // Cors allows all origins.
	apiRouteGroup.Use(Middleware.SetProjectScopeByTokenMiddleware())

	apiRouteGroup.POST("/event/track", SDKTrackHandler)
	apiRouteGroup.POST("/user/identify", SDKIdentifyHandler)
}
