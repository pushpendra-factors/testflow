package handler

import (
	"encoding/json"
	"factors/delta"
	V1 "factors/handler/v1"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

//GetSixSignalReportHandler fetches the sixsignal report from cloud storage for app-server
func GetSixSignalReportHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	r := c.Request

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		log.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, "Query failed. Invalid project.", true
	}

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
	})

	var requestPayload model.SixSignalQueryGroup

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
	}

	var commonQueryFrom int64
	var commonQueryTo int64
	var timezoneString U.TimeZoneString
	if len(requestPayload.Queries) == 0 {
		logCtx.Error("Query failed. Empty query group.")
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Empty query group.", true
	} else {
		// all group queries are run for same time duration, used in dashboard unit caching
		commonQueryFrom = requestPayload.Queries[0].From
		commonQueryTo = requestPayload.Queries[0].To
		timezoneString = requestPayload.Queries[0].Timezone
	}

	fromDate := U.GetDateOnlyFormatFromTimestampAndTimezone(commonQueryFrom, timezoneString)
	toDate := U.GetDateOnlyFormatFromTimestampAndTimezone(commonQueryTo, timezoneString)
	folderName := fmt.Sprintf("%v-%v", fromDate, toDate)
	logCtx.WithFields(log.Fields{"folder name": folderName}).Info("Folder name for reading the result")

	result := delta.GetSixSignalAnalysisData(projectId, folderName)
	return result, http.StatusOK, "", "", false
}

//GetSixSignalPublicReportHandler fetches the sixsignal report from cloud storage for public URLs
func GetSixSignalPublicReportHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		log.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, "Query failed. Invalid project.", true
	}

	queryID := c.Query("query_id")

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
		"queryID":    queryID,
	})

	share, errCode := store.GetStore().GetShareableURLWithShareStringWithLargestScope(projectId, queryID, model.ShareableURLEntityTypeSixSignal)
	if errCode != http.StatusFound {
		logCtx.Error("Failed fetching Shareable URLs in GetSixSignalPublicReportHandler with errCode: ", errCode)
		return nil, http.StatusNotFound, "", "No Shareable URLs found", true
	}

	query, err := store.GetStore().GetSixSignalQueryWithQueryId(projectId, share.EntityID)
	if err != http.StatusFound {
		logCtx.Error("Failed fetching queries in GetSixSignalPublicReportHandler with errCode: ", errCode)
		return nil, http.StatusNotFound, "", "No Query found", true
	}

	var sixSignalQuery model.SixSignalQuery
	err1 := json.Unmarshal(query.Query.RawMessage, &sixSignalQuery)
	if err1 != nil {
		logCtx.WithError(err1).Error("Failed to unmarshal query in GetSixSignalPublicReportHandler with error: ", err1)
		return nil, http.StatusNotFound, "", "Failed to unmarshal query", true
	}

	commonQueryFrom := sixSignalQuery.From
	commonQueryTo := sixSignalQuery.To
	timezoneString := sixSignalQuery.Timezone

	fromDate := U.GetDateOnlyFormatFromTimestampAndTimezone(commonQueryFrom, timezoneString)
	toDate := U.GetDateOnlyFormatFromTimestampAndTimezone(commonQueryTo, timezoneString)
	folderName := fmt.Sprintf("%v-%v", fromDate, toDate)
	logCtx.WithFields(log.Fields{"folder name": folderName}).Info("Folder name for reading the result")

	result := delta.GetSixSignalAnalysisData(projectId, folderName)
	return result, http.StatusOK, "", "", false

}

//CreateSixSignalShareableURLHandler saves the query to the queries table and generate a queryID for shareable URL
func CreateSixSignalShareableURLHandler(c *gin.Context) (interface{}, int, string, bool) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Six Signal Shareable query creation failed. Invalid project."})
		return nil, http.StatusForbidden, "Create SixSignal Shareable URL Failed. Invalid project", true
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	logCtx := log.WithFields(log.Fields{
		"reqId":         U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"loggedInAgent": agentUUID,
		"project_id":    projectID,
	})

	logCtx.Info("Six Signal report access is being changed to public by agent: ", agentUUID)

	params := model.SixSignalShareableURLParams{}
	err := c.BindJSON(&params)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse sixsignal shareable url request body")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Six Signal Shareable query creation failed. Invalid params."})
		return nil, http.StatusBadRequest, "Six Signal Shareable query creation failed. Invalid params.", true
	}

	//Getting sixSignalQuery struct to create data for sha256 encryption
	var sixSignalQuery model.SixSignalQuery
	err1 := json.Unmarshal(params.Query.RawMessage, &sixSignalQuery)
	if err1 != nil {
		logCtx.Error("Cannot Unmarshal SixSignalQueryGroup json in CreateSixSignalShareableURLHandler with error: ", err1)
		return nil, http.StatusBadRequest, "Cannot Unmarshal SixSignalQueryGroup json in CreateSixSignalShareableURLHandler", true
	}
	data := fmt.Sprintf("%d%s%d%d", projectID, sixSignalQuery.Timezone, sixSignalQuery.From, sixSignalQuery.To)

	queryRequest := &model.Queries{
		Query:     *params.Query,
		Title:     "Six Signal Report",
		CreatedBy: agentUUID,
		Settings:  postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		IdText:    U.HashKeyUsingSha256Checksum(data),
		Type:      model.QueryTypeSixSignalQuery,
	}

	queries, errCode, errMsg := store.GetStore().CreateQuery(projectID, queryRequest)
	if errCode != http.StatusCreated {
		return nil, errCode, errMsg, true
	}

	var response model.SixSignalPublicURLResponse
	isShared, _ := isReportShared(projectID, queries.IdText)
	if isShared {
		response = model.SixSignalPublicURLResponse{
			ProjectID:    projectID,
			QueryID:      queries.IdText,
			RouteVersion: ROUTE_VERSION_V1,
		}
		logCtx.Info("Response structure if shared already: ", response)
		return response, http.StatusCreated, "Shareable Query already shared", false
	}

	shareableUrlRequest := &model.ShareableURL{
		EntityType: params.EntityType,
		EntityID:   queries.ID,
		ShareType:  params.ShareType,
		ProjectID:  projectID,
		CreatedBy:  agentUUID,
	}

	if params.IsExpirationSet && params.ExpirationTime > time.Now().Unix() {
		shareableUrlRequest.ExpiresAt = params.ExpirationTime
	} else {
		shareableUrlRequest.ExpiresAt = time.Now().AddDate(0, 3, 0).Unix()
	}

	valid, errMsg := validateCreateShareableURLRequest(shareableUrlRequest, projectID, agentUUID)
	if !valid {
		logCtx.Error(errMsg)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return nil, http.StatusBadRequest, errMsg, true
	}

	logCtx.Info("Shareable urls after validation: ", shareableUrlRequest)
	shareableUrlRequest.QueryID = queries.IdText
	share, errCode := store.GetStore().CreateShareableURL(shareableUrlRequest)
	if errCode != http.StatusCreated {
		logCtx.WithError(err).Error("Failed to create shareable query")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Shareable query creation failed."})
		return nil, http.StatusInternalServerError, "Shareable query creation failed.", true
	}

	response = model.SixSignalPublicURLResponse{
		ProjectID:    projectID,
		RouteVersion: ROUTE_VERSION_V1,
		QueryID:      share.QueryID,
	}
	logCtx.Info("Response structure for sharing: ", response)

	return response, http.StatusCreated, "Shareable Query creation successful", false
}

//isReportShared checks if the report has been already made public
func isReportShared(projectID int64, idText string) (bool, string) {

	share, err := store.GetStore().GetShareableURLWithShareStringWithLargestScope(projectID, idText, model.ShareableURLEntityTypeSixSignal)
	if err == http.StatusBadRequest || err == http.StatusInternalServerError {
		return false, "Shareable query fetch failed. DB error."
	} else if err == http.StatusFound {
		if share.ShareType == model.ShareableURLShareTypePublic {
			return true, "Shareable url already exists."
		}
	}
	return false, "Shareable url doesn't exist"

}
