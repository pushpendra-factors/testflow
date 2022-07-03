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

func GetTemplateHandler(c *gin.Context){

	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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
		return;
	}

	c.JSON(http.StatusFound, template)
}

func CreateTemplateHandler(c *gin.Context){
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Create template failed. Invalid ProjectID"})
		return
	}

	var requestPayload model.DashboardTemplate

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Get template failed. Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	if requestPayload.ID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid template. empty template."})
		return
	}

	if requestPayload.Title == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid title. empty title."})
		return
	}

	templateRequest := &model.DashboardTemplate{
		Title:			requestPayload.Title,
		Description:	requestPayload.Description,
		Dashboard:		postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		Units: 			postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		IsDeleted:      requestPayload.IsDeleted,	
	}

	template, errCode, errMsg := store.GetStore().CreateTemplate(templateRequest)
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(http.StatusCreated, template)
}

func DeleteTemplateHandler(c *gin.Context){
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

func SearchTemplateHandler(c *gin.Context){
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Search queries failed. Invalid project."})
		return
	}

	templateIDParam, ok := c.GetQuery("id")
	if !ok {
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
	dashboardTemplates, error := store.GetStore().GetAllTemplates();
	c.JSON(error, dashboardTemplates)
}

func GenerateDashboardFromTemplateHandler(c *gin.Context){
	templateIDParam, ok := c.GetQuery("id")
	if !ok || templateIDParam == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid template id."})
		return
	}

	DashboardTemplate, errCode := store.GetStore().SearchTemplateWithTemplateID(templateIDParam)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Search template using templateId failed. Invalid template Id."})
		return
	}

	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Create template failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	var dashboardDetails model.Dashboard
	json.Unmarshal(DashboardTemplate.Dashboard.RawMessage, dashboardDetails)

	var unitsArray []model.UnitInfo
	json.Unmarshal(DashboardTemplate.Units.RawMessage, unitsArray)
	q, _ := json.Marshal(unitsArray)

	var queryArray []model.Query
	for _, unit := range unitsArray {
		var queryValue model.Query
		json.Unmarshal(unit.Query.RawMessage, queryValue)
		queryArray = append(queryArray, queryValue)
	}

	dashboardDetails.UnitsPosition = &postgres.Jsonb{json.RawMessage(q)}
	dashboardDetails.Name = DashboardTemplate.Title
	dashboardDetails.AgentUUID = agentUUID
	dashboardDetails.Description = DashboardTemplate.Description
	dashboardDetails.IsDeleted = false
	dashboardDetails.ProjectId = projectID

	dashboard, err := store.GetStore().CreateDashboard(projectID, agentUUID, &dashboardDetails)
	if err != http.StatusCreated {
		c.AbortWithStatusJSON(err, "error")
	}
	c.JSON(http.StatusCreated, dashboard)
}

func GenerateTemplateFromDashboardHandler(c *gin.Context){
	// extract the project Id and agentUUID from the url similarly to dashboardId. USe it as arguments
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get Dashboard from databse failed."})
		return
	}

	var dashboardTemplate model.DashboardTemplate

	var dashboardValues model.Dashboard

	dashboardValues.Description = dashboardParams.Description
	dashboardValues.Type = dashboardParams.Type
	dashboardValues.Class = dashboardParams.Class
	dashboardValues.UnitsPosition = dashboardParams.UnitsPosition


	UnitsInDashboard, _ := store.GetStore().GetDashboardUnits(projectId, agentUUID, dashboardParams.ID)

	var UnitsInTemplate []model.UnitInfo

	for _, unit := range UnitsInDashboard {
		var unitValues model.UnitInfo

		unitValues.Title = unit.Description
		unitValues.Description = unit.Description
		unitValues.ID = int(unit.ID)
		queryValue, _ := store.GetStore().GetQueryWithQueryId(projectId, unit.QueryId)
		q2, _ := json.Marshal(queryValue)
		unitValues.Query = postgres.Jsonb{json.RawMessage(q2)}
		UnitsInTemplate = append(UnitsInTemplate, unitValues)
	}
	q, _ := json.Marshal(UnitsInTemplate)
	dashboardTemplate.Units = postgres.Jsonb{json.RawMessage(q)}

	dashboardTemplate.Description = dashboardParams.Description
	dashboardTemplate.IsDeleted = false
	dashboardTemplate.Title = dashboardParams.Name

	temp, errCode, errMsg := store.GetStore().CreateTemplate(&dashboardTemplate)
	if errCode != http.StatusCreated || errMsg != ""{
		c.AbortWithStatusJSON(errCode, errMsg)
	}

	c.JSON(http.StatusCreated, temp)
}