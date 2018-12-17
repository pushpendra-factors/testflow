package handler

import (
	"encoding/json"
	M "factors/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// Custom type to support Name with event.
type EventWithName struct {
	Name       string         `json:"event_name"`
	Properties postgres.Jsonb `json:"properties"`
	ProjectId  uint64         `json:"project_id"`
	UserId     string         `json:"user_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

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

	var event EventWithName
	decoder := json.NewDecoder(r.Body)
	// Rejects requests with unknown properties on decoded struct. E.g id.
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&event)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("CreateEvent Failed. Json Decoding failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}

	eventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{Name: event.Name, ProjectId: projectId})
	if errCode != http.StatusConflict && errCode != http.StatusCreated {
		c.AbortWithStatus(errCode)
		return
	}

	createdEvent, errCode := M.CreateEvent(&M.Event{ProjectId: projectId, UserId: userId, EventNameId: eventName.ID,
		Properties: event.Properties, CreatedAt: event.CreatedAt, UpdatedAt: event.UpdatedAt})
	if errCode != http.StatusCreated {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusCreated, createdEvent)
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
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusOK, event)
	}
}
