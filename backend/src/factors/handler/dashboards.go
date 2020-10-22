package handler

import (
	"encoding/json"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm/dialects/postgres"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type DashboardRequestPayload struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type DashboardUnitRequestPayload struct {
	Title        string          `json:"title"`
	Presentation string          `json:"presentation"`
	Query        *postgres.Jsonb `json:"query"`
	QueryId      uint64          `json:"query_id"`
}

func GetDashboardsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get dashboards failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboards, errCode := M.GetDashboards(projectId, agentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get dashboards failed."})
		return
	}

	c.JSON(http.StatusFound, dashboards)
}

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

	dashboard, errCode := M.CreateDashboard(projectId, agentUUID,
		&M.Dashboard{Name: requestPayload.Name, Type: requestPayload.Type})
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to create dashboard."})
		return
	}

	c.JSON(http.StatusCreated, dashboard)
}

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

	var requestPayload M.UpdatableDashboard

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Update dashboard failed. Invalid JSON"
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	errCode := M.UpdateDashboard(projectId, agentUUID, dashboardId, &requestPayload)
	if errCode != http.StatusAccepted {
		errMsg := "Update dashboard failed."
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{})
}

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

	dashboardUnits, errCode := M.GetDashboardUnits(projectId, agentUUID, dashboardId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get dashboard units failed."})
		return
	}

	c.JSON(http.StatusFound, dashboardUnits)
}

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

	var requestPayload DashboardUnitRequestPayload

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

	if requestPayload.Query == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid query. empty query."})
		return
	}

	dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(projectId, agentUUID,
		&M.DashboardUnit{
			DashboardId:  dashboardId,
			Query:        *requestPayload.Query,
			Title:        requestPayload.Title,
			Presentation: requestPayload.Presentation,
			QueryId:      requestPayload.QueryId,
		})
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(http.StatusCreated, dashboardUnit)
}

func UpdateDashboardUnitHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Delete dashboard unit failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	var requestPayload DashboardUnitRequestPayload

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

	_, errCode := M.UpdateDashboardUnit(projectId, agentUUID, dashboardId,
		unitId, &M.DashboardUnit{Title: requestPayload.Title})
	if errCode != http.StatusAccepted && errCode != http.StatusNoContent {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to update dashboard unit."})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully updated."})
}

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

	errCode := M.DeleteDashboardUnit(projectId, agentUUID, dashboardId, unitId)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to delete dashboard unit."})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully deleted."})
}

type DashboardUnitWebAnalyticsQueryName struct {
	UnitID    uint64 `json:"unit_id"`
	QueryName string `json:"query_name"`
}

type DashboardUnitWebAnalyticsCustomGroupQuery struct {
	UnitID            uint64   `json:"unit_id"`
	Metrics           []string `json:"metrics"`
	GroupByProperties []string `json:"gbp"`
}

type DashboardUnitsWebAnalyticsQuery struct {
	// Units - Supports redundant metric keys with different unit_ids.
	Units []DashboardUnitWebAnalyticsQueryName `json:"units"`
	// CustomGroupUnits - Customize query with group by properties and metrics.
	CustomGroupUnits []DashboardUnitWebAnalyticsCustomGroupQuery `json:"custom_group_units"`
	From             int64                                       `json:"from"`
	To               int64                                       `json:"to"`
}

func DashboardUnitsWebAnalyticsQueryHandler(c *gin.Context) {
	logCtx := log.WithFields(log.Fields{"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)})

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Web analytics query failed. Invalid project."})
		return
	}
	logCtx = logCtx.WithField("project_id", projectId)

	var requestPayload DashboardUnitsWebAnalyticsQuery
	var queryResult *M.WebAnalyticsQueryResult
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

	cacheResult, errCode := M.GetCacheResultForWebAnalyticsDashboard(projectId, dashboardId,
		requestPayload.From, requestPayload.To)
	if errCode == http.StatusFound && !isHardRefreshForToday(requestPayload.From, hardRefresh) {
		queryResult = cacheResult.Result
		fromCache = true
		lastRefreshedAt = cacheResult.RefreshedAt
	} else {
		// build WebAnalyticsQuery based on query names from DashboardUnitsWebAnalyticsQuery
		// response map[query_name]result = Pass it to ExecuteWebAnalyticsQueries.
		// build map[unit_id]result and respond.

		queryNames := make([]string, 0, len(requestPayload.Units))
		for _, unit := range requestPayload.Units {
			queryNames = append(queryNames, unit.QueryName)
		}

		customGroupQueries := make([]M.WebAnalyticsCustomGroupQuery, 0, 0)
		for _, unit := range requestPayload.CustomGroupUnits {
			customGroupQueries = append(customGroupQueries, M.WebAnalyticsCustomGroupQuery{
				UniqueID:          fmt.Sprintf("%d", unit.UnitID),
				Metrics:           unit.Metrics,
				GroupByProperties: unit.GroupByProperties,
			})
		}

		queryResult, errCode = M.ExecuteWebAnalyticsQueries(
			projectId,
			&M.WebAnalyticsQueries{
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
			return
		}

		M.SetCacheResultForWebAnalyticsDashboard(queryResult, projectId,
			dashboardId, requestPayload.From, requestPayload.To)
		lastRefreshedAt = U.TimeNowIn(U.TimeZoneStringIST).Unix()
	}

	queryResultsByUnitMap := make(map[uint64]M.GenericQueryResult)

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

	c.JSON(http.StatusOK, gin.H{"result": queryResultsByUnitMap,
		"cache": fromCache, "refreshed_at": lastRefreshedAt})
}
