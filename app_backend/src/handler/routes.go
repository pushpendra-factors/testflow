package handler

import (
	M "model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	r.POST("/events", CreateEventHandler)
	r.GET("/events/:id", GetEventHandler)
}

// Test command.
// curl -H "Content-Type: application/json" -i -X POST http://localhost:8080/events -d '{ "account_id": "1", "user_id": "1", "event_name": "login", "attributes": "{\"ip\": \"10.0.0.1\"}"}'
func CreateEventHandler(c *gin.Context) {
	var event M.Event
	c.BindJSON(&event)

	if err := c.BindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}

	var err_code int
	_, err_code = M.CreateEvent(&event)
	if err_code != -1 {
		c.AbortWithStatus(err_code)
	} else {
		c.JSON(201, event)
	}
}

// Test command.
// curl -i -X GET http://localhost:8080/events/bc7318e8-2b69-49b6-baf3-fdf47bcb1af9
func GetEventHandler(c *gin.Context) {
	id := c.Params.ByName("id")
	event, err_code := M.GetEvent(id)
	if err_code != -1 {
		c.AbortWithStatus(err_code)
	} else {
		c.JSON(200, event)
	}
}
