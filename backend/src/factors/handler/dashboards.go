package handler

import (
	"encoding/json"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
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

type DashboardUnitsWebAnalyticsQuery struct {
	// Supports redundant metric keys with different unit_ids.
	Units []DashboardUnitWebAnalyticsQueryName `json:"units"`
	From  int64                                `json:"from"`
	To    int64                                `json:"to"`
}

func DashboardUnitsWebAnalyticsQueryHandler(c *gin.Context) {
	logCtx := log.WithFields(log.Fields{"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)})

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Web analytics query failed. Invalid project."})
		return
	}

	var requestPayload DashboardUnitsWebAnalyticsQuery

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Web analytics query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Web analytics query failed. Json decode failed."})
		return
	}

	// build WebAnalyticsQuery based on query names from DashboardUnitsWebAnalyticsQuery
	// response map[query_name]result = Pass it to ExecuteWebAnalyticsQueries.
	// build map[unit_id]result and respond.

	queryNames := make([]string, 0, len(requestPayload.Units))
	for _, unit := range requestPayload.Units {
		queryNames = append(queryNames, unit.QueryName)
	}

	queryResultsByName, _ := M.ExecuteWebAnalyticsQueries(projectId,
		&M.WebAnalyticsQueries{
			QueryNames: queryNames,
			From:       requestPayload.From,
			To:         requestPayload.To,
		})

	queryResultsByUnitMap := make(map[uint64]M.WebAnalyticsQueryResult)
	for _, unit := range requestPayload.Units {
		if _, exists := (*queryResultsByName)[unit.QueryName]; exists {
			queryResultsByUnitMap[unit.UnitID] = (*queryResultsByName)[unit.QueryName]
		}
	}

	c.JSON(http.StatusOK, queryResultsByUnitMap)
}
