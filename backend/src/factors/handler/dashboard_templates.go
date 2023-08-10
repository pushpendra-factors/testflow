package handler

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func GetTemplateHandler(c *gin.Context) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get Templates failed. Invalid project."})
		return
	}
	templateIDParam, ok := c.GetQuery("id")
	if !ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Search template failed. Invalid search template."})
		return
	}

	template, errCode := store.GetStore().SearchTemplateWithTemplateID(templateIDParam)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get Templates failed."})
		return
	}

	c.JSON(http.StatusFound, template)
}

func CreateTemplateHandler(c *gin.Context) {

	var requestPayload model.DashboardTemplate

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	logCtx := log.WithFields(log.Fields{"project_id": "None"})
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Get template failed. Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	if requestPayload.Title == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid template title"})
		return
	}

	if requestPayload.Units == nil && U.IsEmptyPostgresJsonb(requestPayload.Units) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid Units"})
		return
	}
	if requestPayload.Dashboard == nil && U.IsEmptyPostgresJsonb(requestPayload.Dashboard) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid Dashboard"})
		return
	}

	template, errCode, errMsg := store.GetStore().CreateTemplate(&requestPayload)
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(http.StatusCreated, template)
}

func DeleteTemplateHandler(c *gin.Context) {
	templateIDParam, ok := c.GetQuery("id")
	logCtx := log.WithFields(log.Fields{
		"id": templateIDParam,
	})
	_, errCode := store.GetStore().SearchTemplateWithTemplateID(templateIDParam)
	if errCode != http.StatusFound || !ok {
		logCtx.Error("Could not find any template with the given template id.")
		return
	}
	store.GetStore().DeleteTemplate(templateIDParam)
}

func SearchTemplateHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Search queries failed. Invalid project."})
		return
	}

	templateIDParam := c.Params.ByName("id")
	if templateIDParam == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Search template failed. Invalid search template."})
		return
	}
	// convert template from string to the structure defined for dashboardTemplate
	template, err := store.GetStore().SearchTemplateWithTemplateID(templateIDParam)
	if err != http.StatusFound {
		c.AbortWithStatusJSON(err, gin.H{"error": "Search template failed. No template found"})
	}

	c.JSON(err, template)
}

func GetDashboardTemplatesHandler(c *gin.Context) {
	dashboardTemplates, error := store.GetStore().GetAllTemplates()
	c.JSON(error, dashboardTemplates)
}

func GenerateDashboardFromTemplateHandler(c *gin.Context) {

	templateIDParam := c.Params.ByName("id")
	if templateIDParam == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid template id."})
		return
	}

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Create template failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	dashboard, errCode, err := store.GetStore().GenerateDashboardFromTemplate(projectID, agentUUID, templateIDParam)
	if dashboard == nil {
		c.AbortWithStatusJSON(errCode, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, dashboard)
}

func GenerateTemplateFromDashboardHandler(c *gin.Context) {
	// extract the project Id and agentUUID from the url similarly to dashboardId. USe it as arguments
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get dashboards failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	dashboardId, err := strconv.ParseInt(c.Params.ByName("dashboard_id"), 10, 64)
	if err != nil || dashboardId == 0 {
		log.WithError(err).Error("Update dashboard failed. Invalid dashboard.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard id."})
		return
	}
	if agentUUID == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Search dashboardId or agentUUID or projectId failed. Invalid dashboard Id."})
		return
	}

	var dashboardParams *model.Dashboard
	dashboardParams, ok4 := store.GetStore().GetDashboard(projectId, agentUUID, dashboardId)
	if ok4 != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get Dashboard from database failed."})
		return
	}

	var dashboardTemplate model.DashboardTemplate

	var dashboardValues model.Dashboard

	dashboardTemplate.Title = dashboardParams.Name
	dashboardTemplate.IsDeleted = false
	dashboardValues.Description = dashboardParams.Description
	dashboardValues.Type = dashboardParams.Type
	dashboardValues.Class = dashboardParams.Class
	dashboardValues.UnitsPosition = dashboardParams.UnitsPosition

	dash, _ := json.Marshal(dashboardValues)
	dashboardTemplate.Dashboard = &postgres.Jsonb{json.RawMessage(dash)}

	UnitsInDashboard, _ := store.GetStore().GetDashboardUnits(projectId, agentUUID, dashboardParams.ID)

	var dashUnitPos map[string]map[int64]int
	err = json.Unmarshal(dashboardTemplate.Units.RawMessage, &dashUnitPos)
	if err != nil {
		log.WithFields(log.Fields{"DashboardUnitData": dashboardTemplate.Units}).Error("Failed json decode for unit")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed json decode for unit"})
		return
	}
	var UnitsInTemplate []model.UnitInfo
	for _, unit := range UnitsInDashboard {
		var unitValues model.UnitInfo

		unitValues.ID = int(unit.ID)
		unitValues.Title = unit.Description
		unitValues.Description = unit.Description
		unitValues.Presentation = unit.Presentation
		//unitValues.Position = dashboardValues.UnitsPosition["position"][int(unit.ID)]
		//unitValues.Size = dashboardValues.UnitsPosition["position"][int(unit.ID)]
		//unitValues.QuerySettings = unit.QuerySettings
		//unitValues.QueryType = unit.QueryType
		queryValue, _ := store.GetStore().GetQueryWithQueryId(projectId, unit.QueryId)
		q2, _ := json.Marshal(queryValue)
		unitValues.Query = postgres.Jsonb{json.RawMessage(q2)}
		UnitsInTemplate = append(UnitsInTemplate, unitValues)
	}
	q, _ := json.Marshal(UnitsInTemplate)
	dashboardTemplate.Units = &postgres.Jsonb{json.RawMessage(q)}

	temp, errCode, errMsg := store.GetStore().CreateTemplate(&dashboardTemplate)
	if errCode != http.StatusCreated || errMsg != "" {
		c.AbortWithStatusJSON(errCode, errMsg)
	}

	c.JSON(http.StatusCreated, *temp)
}
