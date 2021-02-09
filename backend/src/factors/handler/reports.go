package handler

import (
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetReportsHandler godoc
// @Summary Get reports for given project id.
// @Tags Reports
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "{"status": "success", "report": []ReportDescription}"
// @Router /{project_id}/reports [get]
func GetReportsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get dashboards failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	reports, errCode := store.GetStore().GetValidReportsListAgentHasAccessTo(projectId, agentUUID)

	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get reports failed."})
		return
	}

	resp := make(map[string]interface{})
	resp["status"] = "success"
	resp["reports"] = reports

	c.JSON(http.StatusOK, resp)
}

// GetReportHandler godoc
// @Summary Get report for given project and report id.
// @Tags Reports
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param report_id path integer true "Report ID"
// @Success 200 {string} json "{"status": "success", "report": "report"}"
// @Router /{project_id}/reports/{report_id} [get]
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

	report, errCode := store.GetStore().GetReportByID(reportId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to fetch report"})
		return
	}

	hasAccess, _ := store.GetStore().HasAccessToDashboard(report.ProjectID, agentUUID, report.DashboardID)
	if !hasAccess {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get Report failed. Report cannot be accessed."})
		return
	}

	resp := make(map[string]interface{})
	resp["status"] = "success"
	resp["report"] = report

	c.JSON(http.StatusOK, resp)
}
