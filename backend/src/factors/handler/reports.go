package handler

import (
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetReportsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get dashboards failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	reports, errCode := M.GetValidReportsListAgentHasAccessTo(projectId, agentUUID)

	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get reports failed."})
		return
	}

	resp := make(map[string]interface{})
	resp["status"] = "success"
	resp["reports"] = reports

	c.JSON(http.StatusOK, resp)
}

func GetReportHandler(c *gin.Context) {

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get dashboards failed. Invalid project."})
		return
	}

	reportId, err := strconv.ParseUint(c.Params.ByName("report_id"), 10, 64)
	if err != nil || reportId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid report id on param."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	report, errCode := M.GetValidReportByID(reportId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to fetch report"})
		return
	}

	hasAccess, _ := M.HasAccessToDashboard(report.ProjectID, agentUUID, report.DashboardID)

	if !hasAccess {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get Report failed. Report cannot be accessed."})
		return
	}

	resp := make(map[string]interface{})
	resp["status"] = "success"
	resp["report"] = report

	c.JSON(http.StatusOK, resp)
}
