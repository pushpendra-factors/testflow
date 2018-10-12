package handler

import (
	"encoding/json"
	M "model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -H "Content-Type: application/json" -i -X POST http://localhost:8080/projects/1/users/1/events -d '{ "event_name": "login", "properties": {"ip": "10.0.0.1", "mobile": true}}'
func CreateEventHandler(c *gin.Context) {
	r := c.Request

	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("CreateEvent Failed. ProjectId parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	userId := c.Params.ByName("user_id")
	if userId == "" || r.Body == nil {
		log.Error("CreateEvent Failed. Missing UserId or Body.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var event M.Event
	err = json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("CreateEvent Failed. Json Decoding failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}
	event.ProjectId = projectId
	event.UserId = userId

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

	// Create corresponding event_name if not exists.
	_, errCode := M.CreateOrGetEventName(&M.EventName{Name: event.EventName, ProjectId: event.ProjectId})
	if errCode != http.StatusConflict && errCode != M.DB_SUCCESS {
		c.AbortWithStatus(errCode)
		return
	}
	_, errCode = M.CreateEvent(&event)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusCreated, event)
	}
}

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/users/1/events/bc7318e8-2b69-49b6-baf3-fdf47bcb1af9
func GetEventHandler(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("GetEvent Failed. ProjectId parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	userId := c.Params.ByName("user_id")
	id := c.Params.ByName("id")
	if userId == "" || id == "" {
		log.Error("GetEvent Failed. Missing UserId or Id.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	event, errCode := M.GetEvent(projectId, userId, id)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusOK, event)
	}
}
