package v1

import (
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"

	C "factors/config"

	"bytes"
	"encoding/json"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var FORCED_EVENT_NAMES = map[int64][]string{
	215: []string{
		// Project ExpertRec.
		"cse.expertrec.com/payments/success",
	},
}

// GetEventNamesHandler godoc
// @Summary Get event names for the given project id.
// @Tags V1Api
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "{"event_names": map[string][]string}"
// @Router /{project_id}/v1/event_names [get]
func GetEventNamesHandler(c *gin.Context) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	isDisplayNameEnabled := c.Query("is_display_name_enabled")
	// RedisGet is the only call. In case of Cache crash, job will be manually triggered to repopulate cache
	// No fallback for now.
	eventNames, err := store.GetStore().GetEventNamesOrderedByOccurenceAndRecency(projectId, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get event names ordered by occurence and recency")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Force add specific events.
	if fNames, pExists := FORCED_EVENT_NAMES[projectId]; pExists {
		eventNames[U.FrequentlySeen] = append(eventNames[U.FrequentlySeen], fNames...)
	}
	if isDisplayNameEnabled == "true" {
		eventsWithGroups := make(map[string][]string)
		// Initializing EventGroups to retain order
		eventsWithGroups["Hubspot"] = make([]string, 0)
		eventsWithGroups["Salesforce"] = make([]string, 0)
		eventsWithGroups[U.SmartEvent] = make([]string, 0)
		eventsWithGroups[U.MostRecent] = make([]string, 0)
		eventsWithGroups[U.FrequentlySeen] = make([]string, 0)
		standardGroups := U.STANDARD_EVENTS_GROUP_NAMES
		for groupName, events := range eventNames {
			for _, event := range events {
				group := groupName
				if standardGroups[event] != "" {
					group = standardGroups[event]
				}
				if eventsWithGroups[group] == nil {
					eventsWithGroups[group] = make([]string, 0)
				}
				eventsWithGroups[group] = append(eventsWithGroups[group], event)
			}
		}
		eventsWithGroupsAfterOrdering := make(map[string][]string)
		for groupName, values := range eventsWithGroups {
			if len(values) > 0 {
				eventsWithGroupsAfterOrdering[groupName] = values
			}
		}
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
		for _, values := range eventsWithGroups {
			for _, value := range values {
				displayName := U.CreateVirtualDisplayName(value)
				_, exist := displayNameEvents[value]
				if !exist {
					displayNameEvents[value] = displayName
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"event_names": U.FilterEmptyKeysAndValues(projectId, eventsWithGroupsAfterOrdering), "display_names": U.FilterDisplayNameEmptyKeysAndValues(projectId, displayNameEvents)})
	} else {
		c.JSON(http.StatusOK, gin.H{"event_names": U.FilterEmptyKeysAndValues(projectId, eventNames)})
	}
}

func GetEventNamesByTypeHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	eventType := c.Params.ByName("type")
	if eventType == "" {
		logCtx.WithField("eventType", eventType).Error("null eventType")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	eventNames, err := store.GetStore().GetMostFrequentlyEventNamesByType(projectId, model.FilterValuesOrEventNamesLimit, C.GetLookbackWindowForEventUserCache(), eventType)
	if err != nil {
		logCtx.WithError(err).Error("get event names ordered by occurence and recency")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Force add specific events.
	if fNames, pExists := FORCED_EVENT_NAMES[projectId]; pExists {
		eventNames = append(eventNames, fNames...)
	}

	c.JSON(http.StatusOK, gin.H{"event_names": eventNames})
}

type UploadRequest struct {
	Payload  []byte `json:"payload"`
	FileName string `json:"file_name"`
}

func UploadListForFilters(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	fileReference := U.GetUUID()
	result := make([]string, 0)

	var payload UploadRequest
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		errMsg := "Create File payload failed"
		log.WithFields(log.Fields{"project_id": projectId}).WithError(err).Error(errMsg)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if payload.FileName != "" {
		fileNameSplit := strings.Split(payload.FileName, ".")
		fileName := ""
		for i := 0; i < len(fileNameSplit)-1; i++ {
			fileName = fileName + fileNameSplit[i]
		}
		fileReference = fmt.Sprintf("%s_%s_%s", U.GetUUID(), fileName, fileNameSplit[len(fileNameSplit)-1])
	}
	payloadString := string(payload.Payload)
	if strings.Contains(payloadString, "\r\n") {
		result = strings.Split(payloadString, "\r\n")
	} else {
		result = strings.Split(payloadString, "\n")
	}

	resultTrimmed := make([]string, 0)
	for _, data := range result {
		if data != "" {
			resultTrimmed = append(resultTrimmed, strings.TrimSpace(data))
		}
	}
	if len(resultTrimmed) <= 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "EmptyFile"})
		return
	}

	path, file := C.GetCloudManager().GetListReferenceFileNameAndPathFromCloud(projectId, fileReference)
	resultJson, err := json.Marshal(resultTrimmed)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("failed to unmarshal result Info.")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = C.GetCloudManager().Create(path, file, bytes.NewReader(resultJson))
	if err != nil {
		log.WithError(err).Error("list File Failed to write to cloud")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	store.GetStore().UploadFilterFile(fileReference, projectId)
	c.JSON(http.StatusOK, gin.H{"file_reference": fileReference})
}

func GetPropertiesByEventCategoryType(c *gin.Context) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	eventCategoryType := c.Query("category")
	if eventCategoryType == "" {
		logCtx.WithField("eventCategoryType", eventCategoryType).Error("null eventCategoryType")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	properties := make(map[string][]string)
	if eventCategoryType == "page_views" {
		properties["categorical"] = U.PAGE_VIEWS_STANDARD_PROPERTIES_CATEGORICAL
		properties["numerical"] = U.PAGE_VIEWS_STANDARD_PROPERTIES_NUMERICAL

	} else if eventCategoryType == "button_clicks" {
		properties["categorical"] = U.BUTTON_CLICKS_STANDARD_PROPERTIES_CATEGORICAL
	}
	c.JSON(http.StatusOK, gin.H{"properties": U.FilterEmptyKeysAndValues(projectId, properties)})
}
