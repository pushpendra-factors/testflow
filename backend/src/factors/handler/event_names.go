package handler

import (
	"encoding/base64"
	C "factors/config"
	mid "factors/middleware"
	"factors/model/store"
	PW "factors/pattern_service_wrapper"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"factors/model/model"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/event_names
// TODO(aravind): Hack below to force some important but not frequent events to show up on production.

//  FORCED_EVENT_NAMES All handlers here have a back up DB call. Will remove this after the cache is functional/updated for all the projects
var FORCED_EVENT_NAMES = map[int64][]string{
	215: []string{
		// Project ExpertRec.
		"cse.expertrec.com/payments/success",
	},
}

func GetDisplayEventNamesHandler(displayNames map[string]string) map[string]string {
	displayNameEvents := make(map[string]string)
	standardEvents := U.STANDARD_EVENTS_DISPLAY_NAMES
	for event, displayName := range standardEvents {
		displayNameEvents[event] = strings.Title(displayName)
	}
	for event, displayName := range displayNames {
		displayNameEvents[event] = strings.Title(displayName)
	}
	return displayNameEvents
}

func RemoveGroupEventNamesOnUserEventNames(categoryToEventNames map[string][]string) map[string][]string {

	for category, eventNames := range categoryToEventNames {
		nonGroupEventNames := make([]string, 0)
		for _, eventName := range eventNames {
			_, isPresent := U.GROUP_EVENT_NAME_TO_GROUP_NAME_MAPPING[eventName]
			if !isPresent {
				nonGroupEventNames = append(nonGroupEventNames, eventName)
			}
		}
		categoryToEventNames[category] = nonGroupEventNames
	}
	return categoryToEventNames
}

func MoveCustomEventNamesOnUserEventNames(categoryToEventNames map[string][]string, customEventNameGroups map[string]string) map[string][]string {
	nonCustomEventNames := make(map[string][]string)
	for group := range categoryToEventNames {
		if group != U.SmartEvent {
			nonCustomEventNames[group] = categoryToEventNames[group]
		}
	}

	customEventNames := categoryToEventNames[U.SmartEvent]
	for i := range customEventNames {
		name := customEventNameGroups[customEventNames[i]]
		if groupDisplayName, exist := U.STANDARD_GROUP_DISPLAY_NAMES[name]; exist {
			nonCustomEventNames[groupDisplayName] = append(nonCustomEventNames[groupDisplayName], customEventNames[i])
			continue
		}

		if groupDisplayName, exist := U.CRM_USER_EVENT_NAME_LABELS[name]; exist {
			nonCustomEventNames[groupDisplayName] = append(nonCustomEventNames[groupDisplayName], customEventNames[i])
			continue
		}

		nonCustomEventNames[U.SmartEvent] = append(nonCustomEventNames[U.SmartEvent], customEventNames[i])
	}

	return nonCustomEventNames
}

func GetCustomEventsByGroupNameAndEventName(projectID int64, eventNames []model.EventName) map[string]string {
	groupEventNames := make(map[string]string)
	for i := range eventNames {
		groupName, exist := model.IsGroupSmartEventName(projectID, &eventNames[i])
		if exist {
			groupEventNames[eventNames[i].Name] = groupName
			continue
		}

		eventName, exist := model.IsUserSmartEventName(projectID, &eventNames[i])
		if exist {
			groupEventNames[eventNames[i].Name] = eventName
			continue
		}
		log.WithFields(log.Fields{"custom_event": eventNames[i]}).Error("Failed to move custom event to display group.")
	}

	return groupEventNames
}

func RemoveLabeledEventNamesFromOtherUserEventNames(categoryToEventNames map[string][]string) map[string][]string {
	for category, eventNames := range categoryToEventNames {
		flag := false
		for _, tempCategory := range U.CRM_USER_EVENT_NAME_LABELS {
			if tempCategory == category {
				flag = true
				break
			}
		}

		if !flag {
			tempString := make([]string, 0)
			for _, eventName := range eventNames {
				_, isPresent := U.CRM_USER_EVENT_NAME_LABELS[eventName]
				if !isPresent {
					tempString = append(tempString, eventName)
				}
			}
			categoryToEventNames[category] = tempString
		}
	}
	return categoryToEventNames
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

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	// RedisGet is the only call. In case of Cache crash, job will be manually triggered to repopulate cache
	// No fallback for now.
	eventNames, err := store.GetStore().GetEventNamesOrderedByOccurenceAndRecency(projectId, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get event names ordered by occurence and recency")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(eventNames) == 0 {

		logCtx.WithError(err).Error(fmt.Sprintf("No Events Returned - ProjectID - %v", projectId))
	}

	eventNameStrings := make([]string, 0)

	if len(eventNames[U.SmartEvent]) > 0 {
		eventNameStrings = append(eventNameStrings, eventNames[U.SmartEvent]...)
	}
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

type EventNamesByUserResponsePayload struct {
	EventNames               map[string][]string `json:"event_names"`
	DisplayNames             map[string]string   `json:"display_names"`
	AllowedDisplayNameGroups map[string]string   `json:"allowed_display_name_groups"`
}

func GetEventNamesByUserHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	// RedisGet is the only call. In case of Cache crash, job will be manually triggered to repopulate cache
	// No fallback for now.
	eventNames, err := store.GetStore().GetEventNamesOrderedByOccurenceAndRecency(projectId, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get event names ordered by occurence and recency")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(eventNames) == 0 {

		logCtx.WithError(err).Error(fmt.Sprintf("No Events Returned - ProjectID - %v", projectId))
	}

	eventNames = RemoveGroupEventNamesOnUserEventNames(eventNames)

	// labeled the non group user event names

	tempEventNames := make(map[string][]string)
	for category, userEventNames := range eventNames {
		tempEventNames[category] = userEventNames
		for _, eventName := range userEventNames {
			if _, ok := U.CRM_USER_EVENT_NAME_LABELS[eventName]; ok {
				category := U.CRM_USER_EVENT_NAME_LABELS[eventName]
				tempEventNames[category] = append(tempEventNames[category], eventName)
			}
		}
	}
	eventNames = tempEventNames
	eventNames = RemoveLabeledEventNamesFromOtherUserEventNames(eventNames)

	_, displayNames := store.GetStore().GetDisplayNamesForAllEvents(projectId)

	displayNameEvents := make(map[string]string)
	standardEvents := U.STANDARD_EVENTS_DISPLAY_NAMES
	for event, displayName := range standardEvents {
		displayNameEvents[event] = strings.Title(displayName)
	}
	for event, displayName := range displayNames {
		displayNameEvents[event] = strings.Title(displayName)
	}
	// TODO: Janani Removing the IsExact property from output since its anyway backward compat with UI
	// Will remove exact/approx logic in UI as well
	for _, values := range eventNames {
		for _, value := range values {
			displayName := U.CreateVirtualDisplayName(value)
			_, exist := displayNameEvents[value]
			if !exist {
				displayNameEvents[value] = displayName
			}
		}
	}
	groups, errCode := store.GetStore().GetGroups(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get groups failed."})
		return
	}

	// all groups event names added to the api response
	for _, group := range groups {
		for key := range U.GROUP_EVENT_NAME_TO_GROUP_NAME_MAPPING {
			groupName := U.GROUP_EVENT_NAME_TO_GROUP_NAME_MAPPING[key]
			groupDisplayName := U.STANDARD_GROUP_DISPLAY_NAMES[groupName]
			_, isPresent := eventNames[groupDisplayName]
			if groupName == group.Name {
				if !isPresent {
					eventNames[groupDisplayName] = make([]string, 0)
				}
				eventNames[groupDisplayName] = append(eventNames[groupDisplayName], key)
			}
		}
	}

	// Force add specific events.
	if fNames, pExists := FORCED_EVENT_NAMES[projectId]; pExists {
		eventNames[U.FrequentlySeen] = append(eventNames[U.FrequentlySeen], fNames...)
	}

	customEventNames, status := store.GetStore().GetSmartEventFilterEventNames(projectId, true)
	if errCode != http.StatusFound {
		if status != http.StatusNotFound {
			logCtx.Error("Failed to get smart event names for event names dropdown.")
		}
	}

	if status == http.StatusFound {
		customEventNameGroups := GetCustomEventsByGroupNameAndEventName(projectId, customEventNames)
		eventNames = MoveCustomEventNamesOnUserEventNames(eventNames, customEventNameGroups)
	}

	c.JSON(http.StatusOK, EventNamesByUserResponsePayload{EventNames: eventNames, DisplayNames: displayNameEvents, AllowedDisplayNameGroups: U.GetStandardDisplayNameGroups()})

}

func GetEventNamesByGroupHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	eventNames, err := store.GetStore().GetEventNamesOrderedByOccurenceAndRecency(projectId, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get event names ordered by occurence and recency")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(eventNames) == 0 {

		logCtx.WithError(err).Error(fmt.Sprintf("No Events Returned - ProjectID - %v", projectId))
	}

	eventNames = RemoveGroupEventNamesOnUserEventNames(eventNames)

	_, displayNames := store.GetStore().GetDisplayNamesForAllEvents(projectId)
	displayNameEvents := GetDisplayEventNamesHandler(displayNames)

	groupName := c.Params.ByName("group_name")
	groupDisplayName := U.STANDARD_GROUP_DISPLAY_NAMES[groupName]
	eventNames[groupDisplayName] = make([]string, 0)

	for key := range U.GROUP_EVENT_NAME_TO_GROUP_NAME_MAPPING {
		if U.GROUP_EVENT_NAME_TO_GROUP_NAME_MAPPING[key] == groupName {
			eventNames[groupDisplayName] = append(eventNames[groupDisplayName], key)
		}
	}
	// Force add specific events.
	if fNames, pExists := FORCED_EVENT_NAMES[projectId]; pExists {
		eventNames[U.FrequentlySeen] = append(eventNames[U.FrequentlySeen], fNames...)
	}

	c.JSON(http.StatusOK, gin.H{"event_names": eventNames, "display_names": displayNameEvents})

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

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	isExplain := c.Query("is_explain")
	isDisplayNameEnabled := c.Query("is_display_name_enabled")
	modelId := uint64(0)
	modelIdParam := c.Query("model_id")
	var err error
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	encodedEName := c.Params.ByName("event_name")
	if encodedEName == "" {
		logCtx.WithField("event_name", encodedEName).Error("null event_name")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	properties := make(map[string][]string)
	var decNameInBytes_1 []byte
	var decNameInBytes_2 []byte
	decNameInBytes_1, err = base64.StdEncoding.DecodeString(encodedEName)
	if err != nil {
		logCtx.WithFields(log.Fields{
			"encodedName": encodedEName,
			log.ErrorKey:  err,
		}).Error("Failed decoding event_name_1.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	decodedEventName := string(decNameInBytes_1)

	decNameInBytes_2, err = base64.StdEncoding.DecodeString(decodedEventName)
	if err != nil {
		logCtx.WithFields(log.Fields{
			"encodedName": decodedEventName,
			log.ErrorKey:  err,
		}).Error("Failed decoding event_name_2.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	eventName := string(decNameInBytes_2)

	logCtx.WithField("decodedEventName", eventName).Debug("Decoded event name on properties request.")

	if isExplain != "true" {
		var statusCode int
		properties, statusCode = store.GetStore().GetEventPropertiesAndModifyResultsForNonExplain(projectId, eventName)
		if statusCode != http.StatusOK {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	} else {
		var status int
		var errMsg string
		properties, status, errMsg = getEventPropertiesFromPatternServer(projectId, modelId, eventName)
		if status != 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  errMsg,
				"status": status,
			})
			return
		}
	}

	_, overrides := store.GetStore().GetPropertyOverridesByType(projectId, U.PROPERTY_OVERRIDE_BLACKLIST, model.GetEntity(false))
	U.FilterDisabledCoreEventProperties(overrides, &properties)

	if isDisplayNameEnabled == "true" {
		displayNamesOp := store.GetStore().GetDisplayNamesForEventProperties(projectId, properties, eventName)
		c.JSON(http.StatusOK, gin.H{"properties": properties, "display_names": displayNamesOp})
		return
	}
	c.JSON(http.StatusOK, properties)
}

func GetChannelGroupingPropertiesHandler(c *gin.Context) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{"display_names": U.CHANNEL_PROPERTIES_DISPLAY_NAMES})
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

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	isExplain := c.Query("is_explain")
	modelId := uint64(0)
	modelIdParam := c.Query("model_id")
	var err error
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

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

	var propertyValues []string
	var decNameInBytes_1 []byte
	var decNameInBytes_2 []byte
	decNameInBytes_1, err = base64.StdEncoding.DecodeString(encodedEName)
	if err != nil {
		logCtx.WithFields(log.Fields{
			"encodedName": encodedEName,
			log.ErrorKey:  err,
		}).Error("Failed decoding event_name_1.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	decodedEventName := string(decNameInBytes_1)

	decNameInBytes_2, err = base64.StdEncoding.DecodeString(decodedEventName)
	if err != nil {
		logCtx.WithFields(log.Fields{
			"encodedName": decodedEventName,
			log.ErrorKey:  err,
		}).Error("Failed decoding event_name_2.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	eventName := string(decNameInBytes_2)

	log.WithField("decodedEventName", eventName).Debug("Decoded event name on properties value request.")

	if isExplain != "true" {
		propertyValues, err = store.GetStore().GetPropertyValuesByEventProperty(projectId, eventName,
			propertyName, 2500, C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.WithError(err).Error("get properties values by event property")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if len(propertyValues) == 0 {
			logCtx.WithError(err).Error(fmt.Sprintf("No event values Returned - ProjectID - %v, EventName - %s, propertyName -%s", projectId, eventName, propertyName))
		}
	} else {
		var status int
		var errMsg string
		propertyValues, status, errMsg = getEventPropertyValuesFromPatternServer(projectId, modelId, eventName, propertyName)
		if status != 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  errMsg,
				"status": status,
			})
			return
		}
	}

	label := c.Query("label")
	if label == "true" {
		propertyValueLabel, err, isSourceEmpty := getPropertyValueLabel(projectId, propertyName, propertyValues)
		if err != nil {
			logCtx.WithError(err).Error("get event properties labels and values by property name")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if isSourceEmpty {
			logCtx.WithField("property_name", propertyName).Warning("source is empty")
		}

		if len(propertyValueLabel) == 0 {
			logCtx.WithField("property_name", propertyName).Error("No event property value labels Returned")
		}

		c.JSON(http.StatusOK, propertyValueLabel)
		return
	}

	c.JSON(http.StatusOK, propertyValues)
}

func getEventPropertyValuesFromPatternServer(projectId int64, modelId uint64, eventName, propertyName string) ([]string, int, string) {
	propertyValues := make([]string, 0)
	ps, err := PW.NewPatternServiceWrapper("", projectId, modelId)
	if err != nil {
		return propertyValues, http.StatusBadRequest, err.Error()
	}
	userInfo := ps.GetUserAndEventsInfo()
	if userInfo.EventPropertiesInfoMap != nil && (*userInfo.EventPropertiesInfoMap)[eventName] != nil {
		for property, values := range (*userInfo.EventPropertiesInfoMap)[eventName].CategoricalPropertyKeyValues {
			if property == propertyName {
				for value, _ := range values {
					propertyValues = append(propertyValues, value)
				}
			}
		}
	}
	return propertyValues, 0, ""
}

func getEventPropertiesFromPatternServer(projectId int64, modelId uint64, eventName string) (map[string][]string, int, string) {
	properties := make(map[string][]string)
	ps, err := PW.NewPatternServiceWrapper("", projectId, modelId)
	if err != nil {
		return properties, http.StatusBadRequest, err.Error()
	}
	userInfo := ps.GetUserAndEventsInfo()

	properties[U.PropertyTypeNumerical] = make([]string, 0)
	properties[U.PropertyTypeCategorical] = make([]string, 0)
	if userInfo.EventPropertiesInfoMap != nil && (*userInfo.EventPropertiesInfoMap)[eventName] != nil {
		for property := range (*userInfo.EventPropertiesInfoMap)[eventName].NumericPropertyKeys {
			properties[U.PropertyTypeNumerical] = append(properties[U.PropertyTypeNumerical], property)
		}
		for property := range (*userInfo.EventPropertiesInfoMap)[eventName].CategoricalPropertyKeyValues {
			properties[U.PropertyTypeCategorical] = append(properties[U.PropertyTypeCategorical], property)
		}
	}
	return properties, 0, ""
}

func getPropertyValueLabel(projectID int64, propertyName string, propertyValues []string) (map[string]string, error, bool) {
	var source string
	if U.IsAllowedCRMPropertyPrefix(propertyName) {
		source = strings.Split(propertyName, "_")[0]
		source = strings.TrimPrefix(source, "$")
	}

	propertyValueLabelMap := make(map[string]string, 0)
	var err error

	if source != "" {
		propertyValueLabelMap, err = store.GetStore().GetPropertyLabelAndValuesByProjectIdAndPropertyKey(projectID, source, propertyName)
		if err != nil {
			return nil, err, false
		}
	}

	for _, value := range propertyValues {
		if _, exists := propertyValueLabelMap[value]; !exists {
			propertyValueLabelMap[value] = value
		}
	}

	return propertyValueLabelMap, nil, source == ""
}
