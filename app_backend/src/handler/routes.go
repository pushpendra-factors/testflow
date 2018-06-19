package handler

import (
	C "config"

	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	r.POST("/projects", CreateProjectHandler)
	r.POST("/projects/:project_id/users/:user_id/events", CreateEventHandler)
	r.GET("/projects/:project_id/users/:user_id/events/:id", GetEventHandler)
	r.POST("/projects/:project_id/users", CreateUserHandler)
	r.GET("/projects/:project_id/users/:user_id", GetUserHandler)
	r.GET("/projects/:project_id/users", GetUsersHandler)
	r.GET("/projects/:project_id/patterns", QueryPatternsHandler)

	// Static files.
	r.Static("static", C.GetConfig().StaticDir)

	templates := multitemplate.New()
	templates.AddFromFiles("app", C.GetConfig().Templates+"base.tmpl", C.GetConfig().Templates+"app.tmpl")
	r.HTMLRender = templates

	r.GET("/app", func(c *gin.Context) {
		c.HTML(200, "app", gin.H{
			"title": "App",
			"stuff": "Interesting app stuff",
		})
	})
}
