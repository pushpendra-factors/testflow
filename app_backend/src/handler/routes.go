package handler

import (
	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	r.POST("/projects/:project_id/users/:user_id/events", CreateEventHandler)
	r.GET("/projects/:project_id/users/:user_id/events/:id", GetEventHandler)
	r.POST("projects/:project_id/users", CreateUserHandler)
	r.GET("/projects/:project_id/users/:user_id", GetUserHandler)
}
