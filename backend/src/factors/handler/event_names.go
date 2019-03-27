package handler

import (
	"encoding/base64"
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
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	}

	eventNames := []string{}
	for _, en := range ens {
		eventNames = append(eventNames, en.Name)
	}
	c.JSON(http.StatusOK, eventNames)
}

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/event_names/view_100020213/properties?model_id=:model_id
func GetEventPropertiesHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)

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

	encodedEName := c.Params.ByName("event_name")
	if encodedEName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	decENameInBytes, err := base64.StdEncoding.DecodeString(encodedEName)
	if err != nil {
		log.WithField("encodedName", encodedEName).Error("Failed decoding event_name.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	eventName := string(decENameInBytes)

	log.WithField("decodedEventName", eventName).Debug("Decoded event name on properties request.")

	properties, err := PC.GetSeenEventProperties(reqId, projectId, modelId, eventName)
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

	reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)

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

	encodedEName := c.Params.ByName("event_name")
	if encodedEName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	decNameInBytes, err := base64.StdEncoding.DecodeString(encodedEName)
	if err != nil {
		log.WithField("encodedName", encodedEName).Error("Failed decoding event_name.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	eventName := string(decNameInBytes)

	log.WithField("decodedEventName", eventName).Debug("Decoded event name on properties value request.")

	propertyName := c.Params.ByName("property_name")
	if propertyName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if propertyValues, err := PC.GetSeenEventPropertyValues(reqId, projectId, modelId, eventName, propertyName); err != nil {
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
