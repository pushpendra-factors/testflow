package handler

import (
	"encoding/json"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

//SavedQueryRequestPayload is struct for post request to create saved query
type SavedQueryRequestPayload struct {
	Title string          `json:"title"`
	Query *postgres.Jsonb `json:"query"`
}

// SavedQueryUpdatePayload is struct update
type SavedQueryUpdatePayload struct {
	Title string `json:"title"`
}

// GetSavedQueriesHandler is for getting saved queries
func GetSavedQueriesHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get Queries failed. Invalid project."})
		return
	}
	queries, errCode := M.GetSavedQueriesWithProjectId(projectID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get Saved Queries failed."})
		return
	}

	c.JSON(http.StatusFound, queries)
}
func CreateSavedQueryHandler(c *gin.Context) {
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

	query, errCode, errMsg := M.CreateQuery(projectID,
		&M.Queries{
			Query:     *requestPayload.Query,
			Title:     requestPayload.Title,
			Type:      M.QueryTypeSavedQuery,
			CreatedBy: agentUUID,
		})
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(http.StatusCreated, query)
}
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

	queryID, err := strconv.ParseUint(c.Params.ByName("query_id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid query id."})
		return
	}
	_, errCode := M.UpdateSavedQuery(projectID, queryID,
		&M.Queries{
			Title: requestPayload.Title,
			Type:  M.QueryTypeSavedQuery,
		})
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to update Saved Query."})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully updated."})
}
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

	errCode, errMsg := M.DeleteSavedQuery(projectID, queryID)

	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully deleted."})
}
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

	queries, errCode := M.SearchQueriesWithProjectId(projectID, queryParams)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Search Queries failed. No query found"})
	}

	c.JSON(errCode, queries)
}
