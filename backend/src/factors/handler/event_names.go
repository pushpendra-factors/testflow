package handler

import (
	"encoding/base64"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"

	C "factors/config"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/event_names
// TODO(aravind): Hack below to force some important but not frequent events to show up on production.

//  FORCED_EVENT_NAMES All handlers here have a back up DB call. Will remove this after the cache is functional/updated for all the projects
var FORCED_EVENT_NAMES = map[uint64][]string{
	215: []string{
		// Project ExpertRec.
		"cse.expertrec.com/payments/success",
	},
}

// GetEventNamesHandler godoc
// @Summary Te fetch event names for a given project id.
// @Tags Events
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "{"event_names": []string}"
// @Router /{project_id}/event_names [get]
func GetEventNamesHandler(c *gin.Context) {

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	// RedisGet is the only call. In case of Cache crash, job will be manually triggered to repopulate cache
	// No fallback for now.
	eventNames, err := M.GetEventNamesOrderedByOccurenceAndRecency(projectId, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get event names ordered by occurence and recency")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(eventNames) == 0 {

		logCtx.WithError(err).Error(fmt.Sprintf("No Events Returned - ProjectID - %v", projectId))
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

}

// GetEventPropertiesHandler Test command.
// curl -i -X GET http://localhost:8080/projects/1/event_names/view_100020213/properties?model_id=:model_id
// GetEventPropertiesHandler godoc
// @Summary To get properties for a given event name.
// @Tags Events
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param event_name path string true "Event Name"
// @Success 200 {string} json "map[string]string"
// @Router /{project_id}/event_names/{event_name}/properties [get]
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

	properties, err = M.GetPropertiesByEvent(projectId, eventName, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get properties by event")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if len(properties) == 0 {
		logCtx.WithError(err).Error(fmt.Sprintf("No event properties Returned - ProjectID - %v, EventName - %s", projectId, eventName))
	}

	U.FilterDisabledCoreEventProperties(&properties)

	c.JSON(http.StatusOK, properties)
}

// GetEventPropertyValuesHandler curl -i -X GET http://localhost:8080/projects/1/event_names/view_100020213/properties/offer_id?model_id=:model_id
// GetEventPropertyValuesHandler godoc
// @Summary Creates a new dashboard unit for the given input.
// @Tags Events
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param event_name path integer true "Event Name"
// @Param property_name path integer true "Property Name"
// @Success 200 {string} json "[]string"
// @Router /{project_id}/event_names/{event_name}/properties/{property_name}/values [get]
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

	propertyValues, err = M.GetPropertyValuesByEventProperty(projectId, eventName, propertyName, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get properties values by event property")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if len(propertyValues) == 0 {
		logCtx.WithError(err).Error(fmt.Sprintf("No event values Returned - ProjectID - %v, EventName - %s, propertyName -%s", projectId, eventName, propertyName))
	}
	c.JSON(http.StatusOK, propertyValues)
}
