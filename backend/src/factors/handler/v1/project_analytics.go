package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetFactorsAnalyticsHandler godoc
// @Summary To project analytics of all projects
// @Tags V1ApiSmartEvent
// @Accept  json
// @Produce json
// @Success 200 {object} map[string][]map[string]interface{}
// @Router /projectanalytics [get]
func GetFactorsAnalyticsHandler(c *gin.Context) {
	noOfDays := int(7)
	noOfDaysParam := c.Query("days")
	isHtmlRequired := c.Query("html")
	var err error

	if noOfDaysParam != "" {
		noOfDays, err = strconv.Atoi(noOfDaysParam)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	var analytics map[string][]*model.ProjectAnalytics

	analytics, err = store.GetStore().GetEventUserCountsOfAllProjects(noOfDays)

	if err != nil {
		log.WithError(err).Error("GetEventUserCountsOfAllProjects")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if isHtmlRequired == "true" {

		resultJson, err := json.Marshal(analytics)
		if err != nil {
			return
		}

		var resultMap map[string][]map[string]interface{}
		err = json.Unmarshal(resultJson, &resultMap)
		if err != nil {
			log.Error(err)
			return
		}

		U.ReturnReadableHtmlFromMaps(c, resultMap, model.AllProjectAnalyticsColumnsName, model.ProjectAnalyticsColumnsNameToJsonKeys, "")
		return
	}

	c.JSON(http.StatusOK, gin.H{"analytics": analytics})
}

// GetFactorsAnalyticsByprojectIDHandler godoc
// @Summary To project analytics by project id
// @Tags V1ApiSmartEvent
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {object} map[string][]map[string]interface{}
// @Router /{project_id}/projectanalytics [get]
func GetFactorsAnalyticsByprojectIDHandler(c *gin.Context) {
	noOfDays := int(7)
	noOfDaysParam := c.Query("days")
	isHtmlRequired := c.Query("html")
	monthString := c.Query("month")
	var err error

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projIdStr := c.Params.ByName("project_id")
	projectID, _ := strconv.ParseInt(projIdStr, 10, 64)

	if projectID == 0 {
		log.Error("Query failed. Invalid project.")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if noOfDaysParam != "" {
		noOfDays, err = strconv.Atoi(noOfDaysParam)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	var analytics map[string][]*model.ProjectAnalytics

	analytics, err = store.GetStore().GetEventUserCountsByProjectID(projectID, noOfDays)
	if err != nil {
		log.WithError(err).Error("GetEventUserCountsByProjectID")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	data, err := store.GetStore().GetGlobalProjectAnalyticsDataByProjectId(projectID, monthString, agentUUID)
	if err != nil {
		log.WithError(err).Error("GetGlobalProjectAnalyticsDataByProjectId")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var settings *model.ProjectSetting
	var errCode int
	settings, errCode = store.GetStore().GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		log.WithError(err).Error("project settings not found")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	integrationList := store.GetStore().GetIntegrationStatusesCount(*settings, projectID, agentUUID)

	project, _ := store.GetStore().GetProject(projectID)

	globalData := make(map[string][]map[string]interface{})
	integrations := make(map[string][]map[string]interface{})

	globalData["metrics"] = data
	integrations["integrations"] = integrationList

	resultJson, err := json.Marshal(analytics)
	if err != nil {
		log.WithError(err).Error("failed to parse data")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var resultMap map[string][]map[string]interface{}
	err = json.Unmarshal(resultJson, &resultMap)
	if err != nil {
		log.Error(err)
		return
	}

	if isHtmlRequired == "true" {

		U.ReturnReadableHtmlFromMaps(c, globalData, model.GlobalDataProjectAnalyticsColumnsName, model.GlobalDataProjectAnalyticsColumnsNameToJsonKeys, fmt.Sprintf("project : %s (%d)", project.Name, project.ID))
		U.ReturnReadableHtmlFromList(c, integrations, model.GlobalDataIntegrationListColumnsName, model.GlobalDataIntegrationListColumnsNameToJsonKeys, "")
		U.ReturnReadableHtmlFromMaps(c, resultMap, model.ProjectAnalyticsColumnsName, model.ProjectAnalyticsColumnsNameToJsonKeys, "remove")

		return
	}

	resultMap["metrics"] = append(resultMap["metrics"], data...)
	resultMap["metrics"] = append(resultMap["metrics"], integrationList...)
	c.JSON(http.StatusOK, resultMap)

	return

}
