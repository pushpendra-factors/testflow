package handler

import (
	"encoding/json"
	H "factors/handler/helpers"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// SavedQueryRequestPayload is struct for post request to create saved query
type SavedQueryRequestPayload struct {
	Title    string          `json:"title"`
	Type     int             `json:"type"`
	Query    *postgres.Jsonb `json:"query"`
	Settings *postgres.Jsonb `json:"settings"`
}

// SavedQueryUpdatePayload is struct update
type SavedQueryUpdatePayload struct {
	Title    string          `json:"title"`
	Settings *postgres.Jsonb `json:"settings"`
	Type     int             `json:"type"`
}

// GetQueriesHandler godoc
// @Summary To get list of all saved queries in project.
// @Tags SavedQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 302 {array} model.Queries
// @Router /{project_id}/queries [get]
func GetQueriesHandler(c *gin.Context) (interface{}, int, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusForbidden, "Get Queries failed. Invalid project.", true
	}
	queries, errCode := store.GetStore().GetALLQueriesWithProjectId(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, "Get Saved Queries failed.", true
	}
	queries = model.RemoveAttributionV1Queries(queries)
	queries = model.RemoveSixSignalQueries(queries)
	return queries, http.StatusOK, "", false
}

// CreateQueryHandler godoc
// @Summary To create a new saved query for given query.
// @Tags SavedQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query body handler.SavedQueryRequestPayload true "Create saved query"
// @Success 201 {array} model.Queries
// @Router /{project_id}/queries [post]
func CreateQueryHandler(c *gin.Context) (interface{}, int, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusForbidden, "Create query failed. Invalid project", true
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	var requestPayload SavedQueryRequestPayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Get queries failed. Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, true
	}

	if requestPayload.Query == nil {
		return nil, http.StatusBadRequest, "invalid query. empty query.", true
	}

	if requestPayload.Title == "" {
		return nil, http.StatusBadRequest, "invalid title. empty title", true
	}

	if requestPayload.Type == 0 {
		return nil, http.StatusBadRequest, "invalid query type. empty type", true
	}

	queryRequest := &model.Queries{
		Query:     *requestPayload.Query,
		Title:     requestPayload.Title,
		Type:      requestPayload.Type,
		CreatedBy: agentUUID,
		// To support empty settings value.
		Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		IdText:   U.RandomStringForSharableQuery(50),
	}

	if requestPayload.Settings != nil && !U.IsEmptyPostgresJsonb(requestPayload.Settings) {
		queryRequest.Settings = *requestPayload.Settings
	}

	query, errCode, errMsg := store.GetStore().CreateQuery(projectID, queryRequest)
	if errCode != http.StatusCreated {
		return nil, errCode, errMsg, true
	}

	return query, http.StatusCreated, "", false
}

// UpdateSavedQueryForReportHandler godoc
// @Summary To update an existing saved query.
// @Tags SavedQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query_id path integer true "Query ID"
// @Param query body handler.SavedQueryRequestPayload true "Update saved query"
// @Success 202 {string} string "{"message": "Successfully updated."}"
// @Router /{project_id}/queries/{query_id} [put]
func UpdateSavedQueryHandler(c *gin.Context) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Update saved query failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	var requestPayload SavedQueryRequestPayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Get queries failed. Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	if requestPayload.Title == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid title. empty title"})
		return
	}

	if requestPayload.Type != 0 && requestPayload.Type != model.QueryTypeDashboardQuery && requestPayload.Type != model.QueryTypeSavedQuery {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}

	queryID, err := strconv.ParseInt(c.Params.ByName("query_id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid query id."})
		return
	}
	requestPayloadContainsQueryUpdate := (requestPayload.Query != nil && !U.IsEmptyPostgresJsonb(requestPayload.Query))

	queryRequest := &model.Queries{
		Query:     postgres.Jsonb{},
		Title:     requestPayload.Title,
		Type:      requestPayload.Type,
		CreatedBy: agentUUID,
		// To support empty settings value.
		Settings:                   postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		IdText:                     U.RandomStringForSharableQuery(50),
		LockedForCacheInvalidation: true,
	}
	if requestPayloadContainsQueryUpdate {
		queryRequest.Query = *requestPayload.Query
	}

	query, status := store.GetStore().GetQueryWithQueryId(projectID, queryID)
	if status != http.StatusFound {
		log.Error("query not found")
		return
	}
	queryRequest.IdText = query.IdText

	if query.LockedForCacheInvalidation {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Saved Query. There is already an update"})
		return
	}

	if requestPayload.Settings != nil && !U.IsEmptyPostgresJsonb(requestPayload.Settings) {
		queryRequest.Settings = *requestPayload.Settings
	}

	_, errCode := store.GetStore().UpdateSavedQuery(projectID, queryID,
		queryRequest)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to update Saved Query."})
		return
	}
	if requestPayloadContainsQueryUpdate {
		go invalidateSavedQueryCache(projectID, query)
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully updated."})
}

func invalidateSavedQueryCache(projectID int64, query *model.Queries) {
	statusCode := H.InValidateSavedQueryCache(query)
	if statusCode != http.StatusOK {
		log.WithField("query_id", query.ID).Error("Failed in invalidating saved query cache.")
	}
	query.LockedForCacheInvalidation = false
	_, errCode := store.GetStore().UpdateSavedQuery(projectID, query.ID, query)
	if errCode != http.StatusAccepted {
		log.WithField("query_id", query.ID).Error("Failed to unset lock for cache.")
	}
}

// DeleteSavedQueryHandler godoc
// @Summary To delete an existing saved query.
// @Tags SavedQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query_id path integer true "Query ID"
// @Success 202 {string} string "{"message": "Successfully deleted."}"
// @Router /{project_id}/queries/{query_id} [delete]
func DeleteSavedQueryHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Delete query failed. Invalid project."})
		return
	}

	queryID, err := strconv.ParseInt(c.Params.ByName("query_id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid query id."})
		return
	}

	errCode, errMsg := store.GetStore().DeleteQuery(projectID, queryID)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	errCode_shareableurl := store.GetStore().DeleteShareableURLWithEntityIDandType(projectID, queryID, model.ShareableURLEntityTypeQuery)
	if errCode_shareableurl != http.StatusNotFound && errCode_shareableurl != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode_shareableurl, "Failed to delete shareable urls.")
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully deleted."})
}

// SearchQueriesHandler godoc
// @Summary To search on existing saved queries.
// @Tags SavedQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 302 {array} model.Queries
// @Router /{project_id}/queries/search [get]
func SearchQueriesHandler(c *gin.Context) (interface{}, int, string, bool) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusForbidden, "Search queries failed. Invalid project.", true
	}
	queryParams, ok := c.GetQuery("query")
	if !ok || queryParams == "" {
		return nil, http.StatusForbidden, "Search queries failed. Invalid search query.", true
	}

	queries, errCode := store.GetStore().SearchQueriesWithProjectId(projectID, queryParams)
	if errCode != http.StatusFound {
		return nil, errCode, "Search Queries failed. No query found", true
	}

	return queries, errCode, "", false
}
