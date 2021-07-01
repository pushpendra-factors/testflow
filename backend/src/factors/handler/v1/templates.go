package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ExecuteTemplateQueryHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	r := c.Request
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	logCtx := log.WithField("reqId", reqID)

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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

	// Run Channel Query
	queryResult, errCode := store.GetStore().RunTemplateQuery(projectId, query, reqID)
	if errCode != http.StatusOK {
		logCtx.Error("Failed to process query from DB")
		if errCode == http.StatusPartialContent {
			return queryResult, errCode, PROCESSING_FAILED, "Failed to process query from DB", true
		}
		return nil, errCode, PROCESSING_FAILED, "Failed to process query from DB", true
	}
	return gin.H{"result": queryResult}, http.StatusOK, "", "", false
}
func GetTemplateConfigHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	logCtx := log.WithField("reqId", reqID)

	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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
