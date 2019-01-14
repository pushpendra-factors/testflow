package handler

import (
	mid "factors/middleware"
	M "factors/model"
	PC "factors/pattern_client"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/event_names
func GetEventNamesHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ens, errCode := M.GetEventNames(projectId)
	if errCode != http.StatusFound {
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
// curl -i -X GET http://localhost:8080/projects/1/event_names/view_100020213/properties?model_id=:model_id
func GetEventPropertiesHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var err error
	modelId := uint64(0)

	modelIdParam := c.Query("model_id")
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	eventName := c.Params.ByName("event_name")
	if eventName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	properties, err := PC.GetSeenEventProperties(projectId, modelId, eventName)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "projectId": projectId, "eventName": eventName}).Error(
			"Get Event Properties failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusOK, properties)

}

// curl -i -X GET http://localhost:8080/projects/1/event_names/view_100020213/properties/offer_id?model_id=:model_id
func GetEventPropertyValuesHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var err error
	modelId := uint64(0)

	modelIdParam := c.Query("model_id")
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
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

	if propertyValues, err := PC.GetSeenEventPropertyValues(projectId, modelId, eventName, propertyName); err != nil {
		log.WithFields(log.Fields{
			"error":        err,
			"projectId":    projectId,
			"modelId":      modelId,
			"eventName":    eventName,
			"propertyName": propertyName}).Error(
			"Get Event Properties failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, propertyValues)
	}
}
