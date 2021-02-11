package handler

import (
	"encoding/json"
	H "factors/handler/helpers"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm/dialects/postgres"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type DashboardRequestPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type DashboardIdUnitsPositions struct {
	ID uint64 `json:"id"`
}

// GetDashboardsHandler godoc
// @Summary Fetches all dashboards for the given project id.
// @Tags Dashboard
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 302 {array} model.Dashboard
// @Router /{project_id}/dashboards [get]
func GetDashboardsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get dashboards failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboards, errCode := store.GetStore().GetDashboards(projectId, agentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get dashboards failed."})
		return
	}

	c.JSON(http.StatusFound, dashboards)
}

// CreateDashboardHandler godoc
// @Summary Creates a new dashboard for the given input.
// @Tags Dashboard
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard body handler.DashboardRequestPayload true "Create new dashboard"
// @Success 201 {object} model.Dashboard
// @Router /{project_id}/dashboards [post]
func CreateDashboardHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Create dashboard failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	var requestPayload DashboardRequestPayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Create dashboard failed. Invalid JSON"
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	dashboard, errCode := store.GetStore().CreateDashboard(projectId, agentUUID,
		&model.Dashboard{Name: requestPayload.Name, Description: requestPayload.Description, Type: requestPayload.Type})
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to create dashboard."})
		return
	}

	c.JSON(http.StatusCreated, dashboard)
}

// UpdateDashboardHandler godoc
// @Summary Updates an existing dashboard.
// @Tags Dashboard
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Param unit body model.UpdatableDashboard true "Update dashboard"
// @Success 202
// @Router /{project_id}/dashboards/{dashboard_id} [put]
func UpdateDashboardHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Update dashboard failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboardId, err := strconv.ParseUint(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		log.WithError(err).Error("Update dashboard failed. Invalid dashboard.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}

	var requestPayload model.UpdatableDashboard

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Update dashboard failed. Invalid JSON"
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	errCode := store.GetStore().UpdateDashboard(projectId, agentUUID, dashboardId, &requestPayload)
	if errCode != http.StatusAccepted {
		errMsg := "Update dashboard failed."
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{})
}

// DeleteDashboardHandler godoc
// @Summary To delete an existing dashboard.
// @Tags Dashboard,V1Api
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Success 202 {string} json "{"message": "Successfully deleted."}"
// @Router /{project_id}/v1/dashboards/{dashboard_id} [delete]
func DeleteDashboardHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Delete dashboard unit failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboardId, err := strconv.ParseUint(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}

	errCode := store.GetStore().DeleteDashboard(projectId, agentUUID, dashboardId)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to delete dashboard."})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully deleted."})
}

// GetDashboardUnitsHandler godoc
// @Summary Fetches dashboard units for the given project and dashboard id.
// @Tags DashboardUnit
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Success 302 {array} model.DashboardUnit
// @Router /{project_id}/dashboards/{dashboard_id}/units [get]
func GetDashboardUnitsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get dashboard units failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboardId, err := strconv.ParseUint(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		log.WithError(err).Error("Get dashboard units failed. Invalid dashboard.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}

	dashboardUnits, errCode := store.GetStore().GetDashboardUnits(projectId, agentUUID, dashboardId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get dashboard units failed."})
		return
	}

	c.JSON(http.StatusFound, dashboardUnits)
}

// CreateDashboardUnitHandler godoc
// @Summary Creates a new dashboard unit for the given input.
// @Tags DashboardUnit
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Param unit body model.DashboardUnitRequestPayload true "Create dashboard unit"
// @Success 201 {object} model.DashboardUnit
// @Router /{project_id}/dashboards/{dashboard_id}/units [post]
func CreateDashboardUnitHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Get dashboard units failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboardId, err := strconv.ParseUint(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}

	var requestPayload model.DashboardUnitRequestPayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "dashboard_id": dashboardId})
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Get dashboard units failed. Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	if requestPayload.Query == nil && requestPayload.QueryId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid query. empty query. no queryId."})
		return
	}

	// to support V1 Api with QueryId & settings
	if requestPayload.Query == nil {
		requestPayload.Query = &postgres.Jsonb{RawMessage: json.RawMessage(`{}`)}
	}

	// to support V1 Api with QueryId & settings
	if requestPayload.Settings == nil {
		requestPayload.Settings = &postgres.Jsonb{RawMessage: json.RawMessage(`{}`)}
	}

	dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(projectId, agentUUID,
		&model.DashboardUnit{
			DashboardId:  dashboardId,
			Query:        *requestPayload.Query,
			Title:        requestPayload.Title,
			Presentation: requestPayload.Presentation,
			QueryId:      requestPayload.QueryId,
			Settings:     *requestPayload.Settings,
		}, model.DashboardUnitForNoQueryID)
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(http.StatusCreated, dashboardUnit)
}

// CreateDashboardUnitForMultiDashboardsHandler godoc
// @Summary Creates a new dashboard unit for each of the given dashboard Ids.
// @Tags Dashboard,DashboardUnit,V1Api
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_ids path string true "Dashboard IDs comma separated"
// @Param payload body model.DashboardUnitRequestPayload true "Create dashboard unit"
// @Success 201 {array} model.DashboardUnit
// @Router /{project_id}/v1/dashboards/multi/{dashboard_ids}/units [post]
func CreateDashboardUnitForMultiDashboardsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Get dashboard units failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboardIdsStr := strings.Split(c.Params.ByName("dashboard_ids"), ",")

	var dashboardIds []uint64
	for _, id := range dashboardIdsStr {
		dashboardId, err := strconv.ParseUint(id, 10, 64)
		if err != nil || dashboardId == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id =" + id})
			return
		}
		dashboardIds = append(dashboardIds, dashboardId)
	}

	var requestPayload model.DashboardUnitRequestPayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	logCtx := log.WithFields(log.Fields{"project_id": projectId})
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Get dashboard units failed. Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	// query should have been created before the dashboard unit
	if requestPayload.QueryId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid queryID. empty queryID."})
		return
	}

	dashboardUnits, errCode, errMsg := store.GetStore().CreateDashboardUnitForMultipleDashboards(dashboardIds, projectId, agentUUID, requestPayload)
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}
	c.JSON(http.StatusCreated, dashboardUnits)
}

// CreateDashboardUnitsForMultipleQueriesHandler godoc
// @Summary Creates a new dashboard unit for each of the given queries.
// @Tags Dashboard,DashboardUnit,V1Api
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Param payload body model.DashboardUnitRequestPayload true "Array of DashboardUnitRequestPayload"
// @Success 201 {array} model.DashboardUnit
// @Router /{project_id}/v1/dashboards/queries/{dashboard_id}/units [post]
func CreateDashboardUnitsForMultipleQueriesHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Get dashboard units failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboardId, err := strconv.ParseUint(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}

	var requestPayload []model.DashboardUnitRequestPayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	logCtx := log.WithFields(log.Fields{"project_id": projectId})
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Get dashboard units failed. Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	dashboardUnits, errCode, errMsg := store.GetStore().CreateMultipleDashboardUnits(requestPayload, projectId, agentUUID, dashboardId)
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}
	c.JSON(http.StatusCreated, dashboardUnits)
}

// UpdateDashboardUnitHandler godoc
// @Summary To update an existing dashboard unit.
// @Tags DashboardUnit
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Param unit_id path integer true "Dashboard Unit ID"
// @Param unit body model.DashboardUnitRequestPayload true "Update dashboard unit"
// @Success 202 {string} json "{"message": "Successfully updated."}"
// @Router /{project_id}/dashboards/{dashboard_id}/units/{unit_id} [put]
func UpdateDashboardUnitHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Delete dashboard unit failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	var requestPayload model.DashboardUnitRequestPayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Update dashboard unit failed. Invalid JSON"
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	dashboardId, err := strconv.ParseUint(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}

	unitId, err := strconv.ParseUint(c.Params.ByName("unit_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard unit id."})
		return
	}

	_, errCode := store.GetStore().UpdateDashboardUnit(projectId, agentUUID, dashboardId,
		unitId, &model.DashboardUnit{Title: requestPayload.Title,
			Description:  requestPayload.Description,
			Presentation: requestPayload.Presentation,
			Settings:     *requestPayload.Settings})
	if errCode != http.StatusAccepted && errCode != http.StatusNoContent {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to update dashboard unit."})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully updated."})
}

// DeleteDashboardUnitHandler godoc
// @Summary To delete an existing dashboard unit.
// @Tags DashboardUnit
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Param unit_id path integer true "Dashboard Unit ID"
// @Success 202 {string} json "{"message": "Successfully deleted."}"
// @Router /{project_id}/dashboards/{dashboard_id}/units/{unit_id} [delete]
func DeleteDashboardUnitHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Delete dashboard unit failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboardId, err := strconv.ParseUint(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}

	unitId, err := strconv.ParseUint(c.Params.ByName("unit_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard unit id."})
		return
	}

	errCode := store.GetStore().DeleteDashboardUnit(projectId, agentUUID, dashboardId, unitId)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to delete dashboard unit."})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully deleted."})
}

// DeleteMultiDashboardUnitHandler godoc
// @Summary To delete multiple existing dashboard unit.
// @Tags DashboardUnit
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Param unit_ids path string true "Dashboard Unit IDs"
// @Success 202 {string} json "{"message": "Successfully deleted."}"
// @Router /{project_id}/v1/dashboards/{dashboard_id}/units/multi/{unit_ids} [delete]
func DeleteMultiDashboardUnitHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Delete dashboard unit failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboardID, err := strconv.ParseUint(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}

	dashboardUnitIDsStr := strings.Split(c.Params.ByName("unit_ids"), ",")

	var dashboardUnitIDs []uint64
	for _, ID := range dashboardUnitIDsStr {
		dashboardUnitID, err := strconv.ParseUint(ID, 10, 64)
		if err != nil || dashboardUnitID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id =" + ID})
			return
		}
		dashboardUnitIDs = append(dashboardUnitIDs, dashboardUnitID)
	}

	errCode, errMsg := store.GetStore().DeleteMultipleDashboardUnits(projectID, agentUUID, dashboardID, dashboardUnitIDs)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully deleted."})
}

// DashboardUnitsWebAnalyticsQueryHandler godoc
// @Summary To fetch result for website analytics dashboard queries.
// @Tags DashboardUnit
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Param unit body handler.DashboardUnitsWebAnalyticsQuery true "Web analytics units"
// @Success 200 {string} json "{"result": "result", "cache": "true", "refreshed_at": "timestamp"}"
// @Router /{project_id}/dashboard/{dashboard_id}/units/query/web_analytics [post]
func DashboardUnitsWebAnalyticsQueryHandler(c *gin.Context) {
	logCtx := log.WithFields(log.Fields{"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)})

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Web analytics query failed. Invalid project."})
		return
	}
	logCtx = logCtx.WithField("project_id", projectId)

	var requestPayload model.DashboardUnitsWebAnalyticsQuery
	var queryResult *model.WebAnalyticsQueryResult
	var fromCache, hardRefresh bool
	var lastRefreshedAt int64

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Web analytics query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Web analytics query failed. Json decode failed."})
		return
	}

	dashboardId, err := strconv.ParseUint(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Web analytics query failed. Invalid dashboard id."})
		return
	}

	refreshParam := c.Query("refresh")
	if refreshParam != "" {
		hardRefresh, _ = strconv.ParseBool(refreshParam)
	}

	cacheResult, errCode := model.GetCacheResultForWebAnalyticsDashboard(projectId, dashboardId,
		requestPayload.From, requestPayload.To)
	if errCode == http.StatusFound && !H.IsHardRefreshForToday(requestPayload.From, hardRefresh) {
		queryResult = cacheResult.Result
		fromCache = true
		lastRefreshedAt = cacheResult.RefreshedAt
	} else {
		// build WebAnalyticsQuery based on query names from DashboardUnitsWebAnalyticsQuery
		// response map[query_name]result = Pass it to ExecuteWebAnalyticsQueries.
		// build map[unit_id]result and respond.

		var cacheResult model.WebAnalyticsQueryResult
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &requestPayload, cacheResult, true)
		if shouldReturn {
			c.AbortWithStatusJSON(resCode, resMsg)
			return
		}

		// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
		model.SetQueryCachePlaceholder(projectId, &requestPayload)
		H.SleepIfHeaderSet(c)

		queryNames := make([]string, 0, len(requestPayload.Units))
		for _, unit := range requestPayload.Units {
			queryNames = append(queryNames, unit.QueryName)
		}

		customGroupQueries := make([]model.WebAnalyticsCustomGroupQuery, 0, 0)
		for _, unit := range requestPayload.CustomGroupUnits {
			customGroupQueries = append(customGroupQueries, model.WebAnalyticsCustomGroupQuery{
				UniqueID:          fmt.Sprintf("%d", unit.UnitID),
				Metrics:           unit.Metrics,
				GroupByProperties: unit.GroupByProperties,
			})
		}

		queryResult, errCode = store.GetStore().ExecuteWebAnalyticsQueries(
			projectId,
			&model.WebAnalyticsQueries{
				QueryNames:         queryNames,
				CustomGroupQueries: customGroupQueries,
				From:               requestPayload.From,
				To:                 requestPayload.To,
			},
		)

		if queryResult == nil || errCode != http.StatusOK {
			logCtx.WithField("err_code", errCode).
				Error("Failed to execute web analytics query.")

			c.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{"error": "Web analytics query failed. Execution failed."})
			model.DeleteQueryCacheKey(projectId, &requestPayload)
			return
		}
		model.SetQueryCacheResult(projectId, &requestPayload, queryResult)

		model.SetCacheResultForWebAnalyticsDashboard(queryResult, projectId,
			dashboardId, requestPayload.From, requestPayload.To)
		lastRefreshedAt = U.TimeNowIn(U.TimeZoneStringIST).Unix()
	}

	queryResultsByUnitMap := make(map[uint64]model.GenericQueryResult)

	queryResultsByName := queryResult.QueryResult
	for _, unit := range requestPayload.Units {
		if _, exists := (*queryResultsByName)[unit.QueryName]; exists {
			queryResultsByUnitMap[unit.UnitID] = (*queryResultsByName)[unit.QueryName]
		}
	}

	for _, unit := range requestPayload.CustomGroupUnits {
		uniqueID := fmt.Sprintf("%d", unit.UnitID)
		if _, exists := queryResult.CustomGroupQueryResult[uniqueID]; exists {
			queryResultsByUnitMap[unit.UnitID] = *queryResult.CustomGroupQueryResult[uniqueID]
		}
	}

	c.JSON(http.StatusOK, H.DashboardQueryResponsePayload{
		Result: queryResultsByUnitMap, Cache: fromCache, RefreshedAt: lastRefreshedAt})
}
