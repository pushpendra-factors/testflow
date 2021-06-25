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

func GetWeeklyInsightsMetadata(c *gin.Context) (interface{}, int, string, string, bool) {
	// Get all the dashboard units for the project
	// Get the metadata
	// Check if weekly insights is enabled
	// Merge the results

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	weeklyInsights := make(map[uint64]WeeklyInsights)
	dashboardUnits, errCode := store.GetStore().GetDashboardUnitsForProjectID(projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Getting dashboardunits failed")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "", true
	}
	for _, unit := range dashboardUnits {
		_, enabled := delta.IsDashboardUnitWIEnabled(unit)
		weeklyInsights[unit.ID] = WeeklyInsights{
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
		mapMetadata, ok := weeklyInsights[detail.DashboardUnitId]
		if !ok {
			continue
		}
		if mapMetadata.InsightsRange[detail.BaseStartTime] == nil {
			mapMetadata.InsightsRange[detail.BaseStartTime] = make([]int64, 0)
		}
		mapMetadata.InsightsRange[detail.BaseStartTime] = append(mapMetadata.InsightsRange[detail.BaseStartTime], detail.ComparisonStartTime)
		weeklyInsights[detail.DashboardUnitId] = mapMetadata
	}
	return weeklyInsights, http.StatusOK, "", "", false
}
