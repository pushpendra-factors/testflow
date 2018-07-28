package handler

import (
	C "config"
	M "model"

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
	r.POST("/projects/:project_id/patterns/query", QueryPatternsHandler)
	r.POST("/projects/:project_id/patterns/crunch", CrunchPatternsHandler)

	// Static files.
	r.Static("static", C.GetConfig().StaticDir)

	templates := multitemplate.New()
	templates.AddFromFiles("app", C.GetConfig().Templates+"base.tmpl", C.GetConfig().Templates+"app.tmpl")
	r.HTMLRender = templates

	type projectEvents struct {
		Name   string   `json:"name"`
		Events []string `json:"events"`
	}
	projectEventsMap := map[uint64]projectEvents{}
	projects, _ := M.GetProjects()
	for _, project := range projects {
		ens, _ := M.GetEventNames(project.ID)
		eventNames := []string{}
		for _, en := range ens {
			eventNames = append(eventNames, en.Name)
		}
		pe := projectEvents{Name: project.Name, Events: eventNames}
		projectEventsMap[project.ID] = pe
	}
	r.GET("/app", func(c *gin.Context) {
		c.HTML(200, "app", gin.H{
			"projectEventsMap": projectEventsMap,
		})
	})
}
