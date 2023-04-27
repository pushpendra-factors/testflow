package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type CreateQueryAndSaveToDashboardPayload struct {
	Title                     string          `json:"title"`
	Type                      int             `json:"type"`
	Query                     *postgres.Jsonb `json:"query"`
	Settings                  *postgres.Jsonb `json:"settings"`
	DashboardUnitDescription  string          `json:"description"`
	DashboardUnitPresentation string          `json:"presentation"`
}

// GetOrCreateAttributionV1DashboardHandler godoc
// @Summary Fetches attribution V1 dashboard for the given project id.
// @Tags Dashboard
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 302 {object} model.Dashboard
// @Router /{project_id}/v1/attribution/dashboards [get]
func GetOrCreateAttributionV1DashboardHandler(c *gin.Context) (interface{}, int, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusForbidden, "Get dashboards failed. Invalid project.", true
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboard, errCode := store.GetStore().GetOrCreateAttributionV1Dashboard(projectId, agentUUID)
	if errCode != http.StatusFound {
		return nil, errCode, "Get dashboards failed.", true
	}

	return dashboard, http.StatusFound, "", false
}

// GetAttributionQueriesHandler godoc
// @Summary To get list of all Attribution queries in project.
// @Tags AttributionQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 302 {array} model.Queries
// @Router /{project_id}/v1/attribution/queries [get]
func GetAttributionQueriesHandler(c *gin.Context) (interface{}, int, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusForbidden, "Get Queries failed. Invalid project.", true
	}
	queries, errCode := store.GetStore().GetALLQueriesWithProjectId(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, "Get Saved Queries failed.", true
	}
	queries = model.SelectAttributionV1Queries(queries)
	return queries, http.StatusOK, "", false
}

// CreateAttributionV1QueryAndSaveToDashboardHandler godoc
// @Summary To create a new query and save it to attribution v1 dashboard for given query.
// @Tags SavedQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query body handler.CreateQueryAndSaveToDashboardPayload
// @Success 201 {object} model.QueryAndDashboardUnit
// @Router /{project_id}/v1/attribution/queries [post]
func CreateAttributionV1QueryAndSaveToDashboardHandler(c *gin.Context) (interface{}, int, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusForbidden, "Create query failed. Invalid project", true
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	var requestPayload CreateQueryAndSaveToDashboardPayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Create query failed. Invalid JSON."
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

	queryRequest := &model.CreateQueryAndSaveToDashboardInfo{
		Query:                     requestPayload.Query,
		Title:                     requestPayload.Title,
		Type:                      requestPayload.Type,
		Settings:                  &postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		DashboardUnitDescription:  requestPayload.DashboardUnitDescription,
		DashboardUnitPresentation: requestPayload.DashboardUnitPresentation,
		CreatedBy:                 agentUUID,
	}

	if requestPayload.Settings != nil && !U.IsEmptyPostgresJsonb(requestPayload.Settings) {
		queryRequest.Settings = requestPayload.Settings
	}

	queryAndDashboardUnit, errCode, errMsg := store.GetStore().CreateQueryAndSaveToDashboard(projectID, queryRequest)
	if errCode != http.StatusCreated {
		return nil, errCode, errMsg, true
	}

	return queryAndDashboardUnit, http.StatusCreated, "", false

}

// DeleteAttributionDashboardUnitAndQueryHandler godoc
// @Summary To delete an existing dashboard unit and corresponding query for Attribution V1.
// @Tags DashboardUnit
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Param unit_id path integer true "Dashboard Unit ID"
// @Param query_id path integer true "Query ID"
// @Success 202 {string} json "{"message": "Successfully deleted."}"
// @Router /{project_id}/dashboards/{dashboard_id}/units/{unit_id}/query/{query_id} [delete]
func DeleteAttributionDashboardUnitAndQueryHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Delete dashboard unit failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboardId, err := strconv.ParseInt(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}

	unitId, err := strconv.ParseInt(c.Params.ByName("unit_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard unit id."})
		return
	}

	queryID, err := strconv.ParseInt(c.Params.ByName("query_id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid query id."})
		return
	}

	errCode, errMsg := store.GetStore().DeleteAttributionDashboardUnitAndQuery(projectId, queryID, agentUUID, dashboardId, unitId)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully deleted."})
}
