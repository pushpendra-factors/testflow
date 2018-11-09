package handler

import (
	C "factors/config"
	M "factors/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/event_names
func GetEventNamesHandler(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ens, errCode := M.GetEventNames(projectId)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatus(errCode)
	} else {
		eventNames := []string{}
		for _, en := range ens {
			eventNames = append(eventNames, en.Name)
		}
		c.JSON(http.StatusOK, eventNames)
	}
}

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/event_names/view_100020213/properties
func GetEventPropertiesHandler(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	eventName := c.Params.ByName("event_name")
	if eventName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ps := C.GetServices().PatternService
	if properties, err := ps.GetSeenEventProperties(projectID, eventName); err != nil {
		log.WithFields(log.Fields{
			"error": err, "projectId": projectID, "eventName": eventName}).Error(
			"Get Event Properties failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, properties)
	}
}

// curl -i -X GET http://localhost:8080/projects/1/event_names/view_100020213/properties/offer_id
func GetEventPropertyValuesHandler(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	eventName := c.Params.ByName("event_name")
	if eventName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	propertyName := c.Params.ByName("property_name")
	if propertyName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ps := C.GetServices().PatternService
	if propertyValues, err := ps.GetSeenEventPropertyValues(projectID, eventName, propertyName); err != nil {
		log.WithFields(log.Fields{
			"error": err, "projectId": projectID, "eventName": eventName,
			"propertyName": propertyName}).Error(
			"Get Event Properties failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, propertyValues)
	}
}
