package handler

import (
	M "model"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	r.POST("/events", CreateEventHandler)
	r.GET("/events/:id", GetEventHandler)
}

// Test command.
// curl -i -X POST http://localhost:8080/events -d '{ "account_id": "1", "user_id": 1, "event_id": 1, "attributes": "{\"ip\": \"10.0.0.0\"}"}'
func CreateEventHandler(c *gin.Context) {
	var event M.Event
	c.BindJSON(&event)

	var err_code int
	_, err_code = M.CreateEvent(&event)
	if err_code != -1 {
		c.AbortWithStatus(err_code)
	} else {
		c.JSON(201, event)
	}
}

// Test command.
// curl -i -X GET http://localhost:8080/events/1
func GetEventHandler(c *gin.Context) {
	id := c.Params.ByName("id")
	event, err_code := M.GetEvent(id)
	if err_code != -1 {
		c.AbortWithStatus(err_code)
	} else {
		c.JSON(200, event)
	}
}
