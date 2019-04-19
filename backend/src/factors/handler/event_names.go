package handler

import (
	"encoding/base64"
	mid "factors/middleware"
	M "factors/model"
	PC "factors/pattern_client"
	U "factors/util"
	"net/http"
	"sort"
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

	eventNames, errCode := M.GetEventNames(projectId)
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	}

	reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)

	candidatePatterns := make([][]string, len(eventNames))
	for ei, en := range eventNames {
		candidatePatterns[ei] = []string{en.Name}
	}

	logCtx := log.WithFields(log.Fields{
		"reqId":      reqId,
		"projectId":  projectId,
		"eventNames": eventNames,
	})

	resultPatterns, err := PC.GetPatterns(reqId, projectId, 0, candidatePatterns)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get event names with occurrence count.")
	}

	sort.Slice(resultPatterns, func(i int, j int) bool {
		return resultPatterns[i].Count > resultPatterns[j].Count
	})

	resultEventNames := make([]string, 0, 0)
	// existence look up map.
	eventNameLookupMap := make(map[string]bool, 0)
	for _, p := range resultPatterns {
		name := p.EventNames[0]
		resultEventNames = append(resultEventNames, name)
		eventNameLookupMap[name] = true
	}

	for _, en := range eventNames {
		if _, exists := eventNameLookupMap[en.Name]; !exists {
			resultEventNames = append(resultEventNames, en.Name)
		}
	}

	c.JSON(http.StatusOK, resultEventNames)
}

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/event_names/view_100020213/properties?model_id=:model_id
func GetEventPropertiesHandler(c *gin.Context) {

	reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)

	logCtx := log.WithFields(log.Fields{
		"reqId": reqId,
	})

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

	encodedEName := c.Params.ByName("event_name")
	if encodedEName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	decENameInBytes, err := base64.StdEncoding.DecodeString(encodedEName)
	if err != nil {
		logCtx.WithField("encodedName", encodedEName).Error("Failed decoding event_name.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	eventName := string(decENameInBytes)

	logCtx.WithField("decodedEventName", eventName).Debug("Decoded event name on properties request.")

	properties, err := PC.GetSeenEventProperties(reqId, projectId, modelId, eventName)
	if err != nil {
		logCtx.WithFields(log.Fields{
			log.ErrorKey: err, "projectId": projectId, "eventName": eventName}).Error(
			"Get Event Properties failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var errCode int
	if len(properties) == 0 {
		properties, errCode = M.GetRecentEventPropertyKeys(projectId, eventName)
		if errCode == http.StatusInternalServerError {
			c.AbortWithStatus(errCode)
			return
		}
	}

	c.JSON(http.StatusOK, properties)
}

// curl -i -X GET http://localhost:8080/projects/1/event_names/view_100020213/properties/offer_id?model_id=:model_id
func GetEventPropertyValuesHandler(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

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
		logCtx.WithFields(log.Fields{
			"encodedName": encodedEName,
			log.ErrorKey:  err,
		}).Error("Failed decoding event_name.")
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

	propertyValues, err := PC.GetSeenEventPropertyValues(reqId, projectId, modelId, eventName, propertyName)
	if err != nil {
		logCtx.WithFields(log.Fields{
			log.ErrorKey:   err,
			"projectId":    projectId,
			"modelId":      modelId,
			"eventName":    eventName,
			"propertyName": propertyName}).Error(
			"Get Event Properties failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var errCode int
	if len(propertyValues) == 0 {
		propertyValues, errCode = M.GetRecentEventPropertyValues(projectId, eventName, propertyName)
		if errCode == http.StatusInternalServerError {
			c.AbortWithStatus(errCode)
			return
		}
	}

	c.JSON(http.StatusOK, propertyValues)
}
