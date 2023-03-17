package v1

import (
	"factors/delta"
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type WeeklyInsights struct {
	Enabled       bool
	InsightsRange map[int64][]int64
}

type Result struct {
	QueryWiseResult         interface{}
	DashboardUnitWiseResult interface{}
}

func GetWeeklyInsightsMetadata(c *gin.Context) (interface{}, int, string, string, bool) {
	// Get all the dashboard units for the project
	// Get the metadata
	// Check if weekly insights is enabled
	// Merge the results

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	var result Result

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	weeklyInsightsByDashboard := make(map[int64]WeeklyInsights)
	weeklyInsightsByQuery := make(map[int64]WeeklyInsights)
	queryToDashboardUnitMap := make(map[int64][]int64)

	dashboardUnits, errCode := store.GetStore().GetDashboardUnitsForProjectID(projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Getting dashboardunits failed")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "", true
	}

	for _, dashboardUnit := range dashboardUnits {
		if queryToDashboardUnitMap[dashboardUnit.QueryId] == nil {
			queryToDashboardUnitMap[dashboardUnit.QueryId] = make([]int64, 0)
		}
		queryToDashboardUnitMap[dashboardUnit.QueryId] = append(queryToDashboardUnitMap[dashboardUnit.QueryId], dashboardUnit.ID)
	}

	for _, unit := range dashboardUnits {
		_, _, _, enabled, _, _, _ := delta.IsDashboardUnitWIEnabled(unit)
		weeklyInsightsByDashboard[unit.ID] = WeeklyInsights{
			Enabled:       enabled,
			InsightsRange: make(map[int64][]int64),
		}
		weeklyInsightsByQuery[unit.QueryId] = WeeklyInsights{
			Enabled:       enabled,
			InsightsRange: make(map[int64][]int64),
		}
	}

	metadata, errCode, msg := store.GetStore().GetWeeklyInsightsMetadata(projectId)
	if errCode != http.StatusFound {
		logCtx.Error(msg)
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, msg, true
	}
	for _, detail := range metadata {
		mapMetadata, ok := weeklyInsightsByQuery[detail.QueryId]
		if !ok {
			continue
		}
		if mapMetadata.InsightsRange[detail.BaseStartTime] == nil {
			mapMetadata.InsightsRange[detail.BaseStartTime] = make([]int64, 0)
		}
		mapMetadata.InsightsRange[detail.BaseStartTime] = append(mapMetadata.InsightsRange[detail.BaseStartTime], detail.ComparisonStartTime)
		weeklyInsightsByQuery[detail.QueryId] = mapMetadata
		for _, dashboardId := range queryToDashboardUnitMap[detail.QueryId] {
			weeklyInsightsByDashboard[dashboardId] = mapMetadata
		}
	}
	result = Result{
		QueryWiseResult:         weeklyInsightsByQuery,
		DashboardUnitWiseResult: weeklyInsightsByDashboard,
	}
	return result, http.StatusOK, "", "", false
}
