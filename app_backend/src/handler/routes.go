package handler

import (
	"encoding/json"
	M "model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	r.POST("/events", CreateEventHandler)
	r.GET("/events/:id", GetEventHandler)
}

// Test command.
// curl -H "Content-Type: application/json" -i -X POST http://localhost:8080/events -d '{ "project_id": "1", "user_id": "1", "event_name": "login", "properties": {"ip": "10.0.0.1", "mobile": true}}'
func CreateEventHandler(c *gin.Context) {
	var event M.Event

	r := c.Request
	if r.Body == nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}
	/* Commented out code. Using Json Decoder directly above, since gin.Context.BindJSON is returning error "EOF"
	   despite being able to decode the json. Need to check.
	c.BindJSON(&event)
	if err := c.BindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}*/

	var err_code int
	_, err_code = M.CreateEvent(&event)
	if err_code != M.DB_SUCCESS {
		c.AbortWithStatus(err_code)
	} else {
		c.JSON(http.StatusCreated, event)
	}
}

// Test command.
// curl -i -X GET http://localhost:8080/events/bc7318e8-2b69-49b6-baf3-fdf47bcb1af9
func GetEventHandler(c *gin.Context) {
	id := c.Params.ByName("id")
	event, err_code := M.GetEvent(id)
	if err_code != M.DB_SUCCESS {
		c.AbortWithStatus(err_code)
	} else {
		c.JSON(http.StatusOK, event)
	}
}
