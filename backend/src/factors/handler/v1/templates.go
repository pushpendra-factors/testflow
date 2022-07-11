package v1

import (
	"encoding/json"
	H "factors/handler/helpers"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ExecuteTemplateQueryHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	r := c.Request
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	logCtx := log.WithField("reqId", reqID)

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		logCtx.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, "Query failed. Invalid project.", true
	}
	logCtx = logCtx.WithField("project_id", projectId).WithField("reqId", reqID)

	templateType, err := strconv.ParseInt(c.Params.ByName("type"), 10, 64)
	if templateType != 1 || err != nil {
		logCtx.Error("Query failed. Invalid type.")
		return nil, http.StatusUnauthorized, INVALID_INPUT, "Query failed. Invalid type.", true
	}

	var query model.TemplateQuery
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&query); err != nil {
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
	}
	query.Type = int(templateType)
	emptyThresholds := model.RequestThresholds{}
	if query.Thresholds == emptyThresholds {
		query.Thresholds = model.DefaultThresholds
	}
	statusCode, timezoneString := getTimezoneForTemplates(projectId, reqID, query.Timezone)
	if statusCode != http.StatusOK {
		logCtx.Error("Query failed. Failed to get Timezone.")
		return nil, statusCode, INVALID_INPUT, "Query failed. Failed to get Timezone.", true
	}
	query.From, query.To, query.PrevFrom, query.PrevTo, err = model.GetInputOrDefaultTimestampsForTemplateQueryWithDays(query, timezoneString, 7)
	if err != nil {
		logCtx.WithError(err).Error("Query failed. Getting date ranges failed.")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error Processing/Fetching date ranges from timestamps", true
	}

	var cacheResult model.TemplateResponse
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &query, cacheResult, false, reqID)
	if shouldReturn {
		if resCode == http.StatusOK {
			return gin.H{"result": resMsg}, resCode, "", "", false
		}
		logCtx.Error("Query failed. Error Processing/Fetching data from Query cache")
		return nil, resCode, PROCESSING_FAILED, "Error Processing/Fetching data from Query cache", true
	}

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	model.SetQueryCachePlaceholder(projectId, &query)
	H.SleepIfHeaderSet(c)
	// Run Channel Query
	queryResult, errCode := store.GetStore().RunTemplateQuery(projectId, query, reqID)
	if errCode != http.StatusOK {
		model.DeleteQueryCacheKey(projectId, &query)
		logCtx.Error("Failed to process query from DB")
		if errCode == http.StatusPartialContent {
			return queryResult, errCode, PROCESSING_FAILED, "Failed to process query from DB", true
		}
		return nil, errCode, PROCESSING_FAILED, "Failed to process query from DB", true
	}
	model.SetQueryCacheResult(projectId, &query, queryResult)
	return gin.H{"result": queryResult}, http.StatusOK, "", "", false
}

// TODO later. Move all common handler timezone methods to this.
func getTimezoneForTemplates(projectID int64, reqID string, inputTimezoneString string) (int, U.TimeZoneString) {
	var timezoneString U.TimeZoneString
	var statusCode int
	logCtx := log.WithField("project_id", projectID).WithField("reqId", reqID).WithField("inputTimezone", inputTimezoneString)
	if inputTimezoneString != "" {
		_, errCode := time.LoadLocation(inputTimezoneString)
		if errCode != nil {
			logCtx.Error("Query failed to load the location input")
			return http.StatusBadRequest, ""
		}
		timezoneString = U.TimeZoneString(inputTimezoneString)
	} else {
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectID)
		if statusCode != http.StatusFound {
			logCtx.Error("Query failed. Failed to get Timezone.")
			return statusCode, ""
		}
	}
	return http.StatusOK, timezoneString
}
func GetTemplateConfigHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	logCtx := log.WithField("reqId", reqID)

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	logCtx = logCtx.WithField("project_id", projectID).WithField("reqId", reqID)

	templateType, err := strconv.ParseInt(c.Params.ByName("type"), 10, 64)
	if templateType != 1 || err != nil {
		return nil, http.StatusUnauthorized, INVALID_INPUT, "Invalid template type", true
	}
	templateConfig, errCode := store.GetStore().GetTemplateConfig(projectID, int(templateType))
	if errCode != http.StatusOK {
		return nil, errCode, PROCESSING_FAILED, ErrorMessages[PROCESSING_FAILED], true
	}
	return templateConfig, http.StatusOK, "", "", false
}

type updateThresholdsPayload struct {
	Thresholds []model.TemplateThreshold `json:"thresholds"`
}

func UpdateTemplateConfigHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	r := c.Request
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	logCtx := log.WithField("reqId", reqID)

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		logCtx.Error("Update thresholds failed. Invalid project.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, "Update thresholds failed. Invalid project.", true
	}
	logCtx = logCtx.WithField("project_id", projectId).WithField("reqId", reqID)

	templateType, err := strconv.ParseInt(c.Params.ByName("type"), 10, 64)
	if templateType != 1 || err != nil {
		logCtx.Error("Update thresholds failed. Invalid type.")
		return nil, http.StatusUnauthorized, INVALID_INPUT, "Update thresholds failed. Invalid type.", true
	}

	var thresholdsPayload updateThresholdsPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&thresholdsPayload); err != nil {
		logCtx.WithError(err).Error("Update thresholds failed. Json decode failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Update thresholds failed. Json decode failed.", true
	}

	// Run Channel Query
	thresholds, errString := store.GetStore().UpdateTemplateConfig(projectId, int(templateType), thresholdsPayload.Thresholds)
	if errString != "" {
		logCtx.Error(errString)
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, errString, true
	}
	return gin.H{"result": thresholds}, http.StatusOK, "", "", false
}
