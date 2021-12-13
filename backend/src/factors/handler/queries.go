package handler

import (
	"encoding/json"
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

//SavedQueryRequestPayload is struct for post request to create saved query
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
func GetQueriesHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get Queries failed. Invalid project."})
		return
	}
	queries, errCode := store.GetStore().GetALLQueriesWithProjectId(projectID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get Saved Queries failed."})
		return
	}

	c.JSON(http.StatusFound, queries)
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
func CreateQueryHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Create query failed. Invalid project."})
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

	if requestPayload.Query == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid query. empty query."})
		return
	}

	if requestPayload.Title == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid title. empty title"})
		return
	}

	if requestPayload.Type == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid query type. empty type"})
		return
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
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(http.StatusCreated, query)
}

// UpdateSavedQueryHandler godoc
// @Summary To update an existing saved query.
// @Tags SavedQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query_id path integer true "Query ID"
// @Param query body handler.SavedQueryUpdatePayload true "Update saved query"
// @Success 202 {string} string "{"message": "Successfully updated."}"
// @Router /{project_id}/queries/{query_id} [put]
func UpdateSavedQueryHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Update saved query failed. Invalid project."})
		return
	}

	var requestPayload SavedQueryUpdatePayload

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

	queryID, err := strconv.ParseUint(c.Params.ByName("query_id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid query id."})
		return
	}

	query := model.Queries{}
	if requestPayload.Title != "" {
		query.Title = requestPayload.Title
	}

	query.Type = model.QueryTypeSavedQuery
	if requestPayload.Type == model.QueryTypeDashboardQuery {
		query.Type = requestPayload.Type
	}

	if requestPayload.Settings != nil && !U.IsEmptyPostgresJsonb(requestPayload.Settings) {
		query.Settings = *requestPayload.Settings
	}

	_, errCode := store.GetStore().UpdateSavedQuery(projectID, queryID,
		&query)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to update Saved Query."})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully updated."})
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
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Delete query failed. Invalid project."})
		return
	}

	queryID, err := strconv.ParseUint(c.Params.ByName("query_id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid query id."})
		return
	}

	errCode, errMsg := store.GetStore().DeleteQuery(projectID, queryID)

	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, errMsg)
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
func SearchQueriesHandler(c *gin.Context) {

	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Search queries failed. Invalid project."})
		return
	}
	queryParams, ok := c.GetQuery("query")
	if !ok || queryParams == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Search queries failed. Invalid search query."})
		return
	}

	queries, errCode := store.GetStore().SearchQueriesWithProjectId(projectID, queryParams)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Search Queries failed. No query found"})
	}

	c.JSON(errCode, queries)
}
