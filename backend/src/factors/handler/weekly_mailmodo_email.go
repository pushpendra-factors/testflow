package handler

import (
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetWeeklyMailmodoEmailMetricsHandler godoc
// @Summary To get the metrics for the weekly mailmodo mail of the given projectId for internal testing.
// @Tags Weekly Email, Mailmodo
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// Success 302 {object} model.WeeklyMailmodoEmailMetrics
// @Router /{project_id}/internal/weekly_email_metrics [GET]
func GetWeeklyMailmodoEmailMetricsHandler(c *gin.Context) (interface{}, int, string, bool) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		log.Error("Query failed. Invalid project.")
		return nil, http.StatusForbidden, "Invalid project", true
	}

	startTimeStamp, endTimeStamp, err := U.GetQueryRangePresetLastWeekIn(U.TimeZoneStringIST)
	if err != nil {
		log.Error("Failed to fetch start and end timestamp with error ", err)
		return nil, http.StatusInternalServerError, "Failed to fetch start and end timestamp", true

	}

	metrics, err := store.GetStore().GetWeeklyMailmodoEmailsMetrics(projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		log.Error("Failed to fetch metrics with error", err)
		return nil, http.StatusInternalServerError, "Failed to fetch metrics", true
	}

	return metrics, http.StatusFound, "", false

}
