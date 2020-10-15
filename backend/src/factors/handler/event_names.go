package handler

import (
	"encoding/base64"
	"factors/handler/helpers"
	mid "factors/middleware"
	M "factors/model"
	PC "factors/pattern_client"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"

	C "factors/config"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/event_names
// TODO(aravind): Hack below to force some important but not frequent events to show up on production.

// All handlers here have a back up DB call. Will remove this after the cache is functional/updated for all the projects
var FORCED_EVENT_NAMES = map[uint64][]string{
	215: []string{
		// Project ExpertRec.
		"cse.expertrec.com/payments/success",
	},
}

func GetEventNamesHandler(c *gin.Context) {

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	if helpers.IsProjectWhitelistedForEventUserCache(projectId) {
		// RedisGet is the only call. In case of Cache crash, job will be manually triggered to repopulate cache
		// No fallback for now.
		eventNames, err := M.GetEventNamesOrderedByOccurenceAndRecency(projectId, 2500, C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.WithError(err).Error("get event names ordered by occurence and recency")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if len(eventNames) == 0 {

			logCtx.WithError(err).Error(fmt.Sprintf("No Events Returned - ProjectID - %s", projectId))
		}

		eventNameStrings := make([]string, 0)
		if len(eventNames[U.MostRecent]) > 0 {
			eventNameStrings = append(eventNameStrings, eventNames[U.MostRecent]...)
		}
		if len(eventNames[U.FrequentlySeen]) > 0 {
			eventNameStrings = append(eventNameStrings, eventNames[U.FrequentlySeen]...)
		}
		// Force add specific events.
		if fNames, pExists := FORCED_EVENT_NAMES[projectId]; pExists {
			eventNameStrings = append(eventNameStrings, fNames...)
		}

		// TODO: Janani Removing the IsExact property from output since its anyway backward compat with UI
		// Will remove exact/approx logic in UI as well
		c.JSON(http.StatusOK, gin.H{"event_names": eventNameStrings})
	} else {
		requestType := c.Query("type")
		if requestType != M.EVENT_NAME_REQUEST_TYPE_APPROX && requestType != M.EVENT_NAME_REQUEST_TYPE_EXACT {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		eventNames, isExact, errCode := M.GetEventNamesOrderedByOccurrence(projectId, requestType)
		if errCode != http.StatusFound {
			c.AbortWithStatus(errCode)
		}
		names := make([]string, 0, 0)
		for _, eventName := range eventNames {
			names = append(names, eventName.Name)
		}
		if fNames, pExists := FORCED_EVENT_NAMES[projectId]; pExists {
			names = append(names, fNames...)
		}

		c.JSON(http.StatusOK, gin.H{"event_names": names, "exact": isExact})
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

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	encodedEName := c.Params.ByName("event_name")
	if encodedEName == "" {
		logCtx.WithField("event_name", encodedEName).Error("null event_name")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var err error
	var properties map[string][]string
	var decENameInBytes []byte
	decENameInBytes, err = base64.StdEncoding.DecodeString(encodedEName)
	if err != nil {
		logCtx.WithField("encodedName", encodedEName).Error("Failed decoding event_name.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	eventName := string(decENameInBytes)

	logCtx.WithField("decodedEventName", eventName).Debug("Decoded event name on properties request.")

	if helpers.IsProjectWhitelistedForEventUserCache(projectId) {
		properties, err = M.GetPropertiesByEvent(projectId, eventName, 2500, C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.WithError(err).Error("get properties by event")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if len(properties) == 0 {
			logCtx.WithError(err).Error(fmt.Sprintf("No event properties Returned - ProjectID - %s, EventName - %s", projectId, eventName))
		}
	} else {
		reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
		modelId := uint64(0)

		modelIdParam := c.Query("model_id")
		if modelIdParam != "" {
			modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			properties, err = PC.GetSeenEventProperties(reqId, projectId, modelId, eventName)
			if err != nil {
				logCtx.WithFields(log.Fields{
					log.ErrorKey: err, "projectId": projectId, "eventName": eventName}).Error(
					"Get Event Properties failed.")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}

		var errCode int
		if len(properties) == 0 {
			properties, errCode = M.GetRecentEventPropertyKeys(projectId, eventName)
			if errCode == http.StatusInternalServerError {
				c.AbortWithStatus(errCode)
				return
			}
		}
	}

	U.FilterDisabledCoreEventProperties(&properties)

	c.JSON(http.StatusOK, properties)
}

// curl -i -X GET http://localhost:8080/projects/1/event_names/view_100020213/properties/offer_id?model_id=:model_id
func GetEventPropertyValuesHandler(c *gin.Context) {

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	encodedEName := c.Params.ByName("event_name")
	if encodedEName == "" {
		logCtx.WithField("event_name", encodedEName).Error("null event_name")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	propertyName := c.Params.ByName("property_name")
	if propertyName == "" {
		logCtx.WithField("propertyName", propertyName).Error("null property name")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var err error
	var propertyValues []string
	var decNameInBytes []byte
	decNameInBytes, err = base64.StdEncoding.DecodeString(encodedEName)
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

	if helpers.IsProjectWhitelistedForEventUserCache(projectId) {
		propertyValues, err = M.GetPropertyValuesByEventProperty(projectId, eventName, propertyName, 2500, C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.WithError(err).Error("get properties values by event property")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if len(propertyValues) == 0 {
			logCtx.WithError(err).Error(fmt.Sprintf("No event values Returned - ProjectID - %s, EventName - %s, propertyName", projectId, eventName, propertyName))
		}
	} else {
		reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
		modelId := uint64(0)

		modelIdParam := c.Query("model_id")
		if modelIdParam != "" {
			modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			propertyValues, err = PC.GetSeenEventPropertyValues(reqId, projectId, modelId, eventName, propertyName)
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
		}

		var errCode int
		if len(propertyValues) == 0 {
			propertyValues, errCode = M.GetRecentEventPropertyValues(projectId, eventName, propertyName)
			if errCode == http.StatusInternalServerError {
				c.AbortWithStatus(errCode)
				return
			}
		}
	}
	c.JSON(http.StatusOK, propertyValues)
}
