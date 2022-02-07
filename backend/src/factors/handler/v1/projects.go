package v1

import (
	"net/http"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"

	mid "factors/middleware"

	H "factors/handler/helpers"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
)

// Test command.
// curl -i -X GET http://localhost:8080/v1/projects
// Response will be same as list of projects but will be grouped on the calling user's role in that project
// Sample response
/*{
    "2": [
        {
            "id": 1,
            "name": "My Project",
            "token": "",
            "private_token": "",
            "created_at": "2020-08-04T11:04:42.406627+05:30",
            "updated_at": "2020-08-04T11:04:42.406627+05:30",
            "project_uri": "",
            "time_format": "",
            "date_format": "",
            "time_zone": ""
		},...
	]
	"1": [
        {
            "id": 2,
            "name": "My Project",
            "token": "",
            "private_token": "",
            "created_at": "2020-08-04T11:04:42.406627+05:30",
            "updated_at": "2020-08-04T11:04:42.406627+05:30",
            "project_uri": "",
            "time_format": "",
            "date_format": "",
            "time_zone": ""
		},...
	]
*/

func GetProjectsHandler(c *gin.Context) {
	authorizedProjects := U.GetScopeByKey(c, mid.SCOPE_AUTHORIZED_PROJECTS)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projects, errCode := store.GetStore().GetProjectsByIDs(authorizedProjects.([]uint64))
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	} else if errCode == http.StatusNoContent || errCode == http.StatusBadRequest {
		resp := make(map[string]interface{})
		resp["projects"] = []model.Project{}
		if !C.EnableDemoReadAccess() {
			c.JSON(http.StatusNotFound, resp)
			return
		}
	}
	projectRoleMap := make(map[uint64]uint64)
	resp := make(map[uint64][]interface{})
	if len(projects) > 0 {
		projectAgentMappings, errCode := store.GetStore().GetProjectAgentMappingsByProjectIds(authorizedProjects.([]uint64))
		if errCode != http.StatusFound {
			c.AbortWithStatus(errCode)
			return
		}

		for _, projectAgent := range projectAgentMappings {
			if projectAgent.AgentUUID == loggedInAgentUUID {
				projectRoleMap[projectAgent.ProjectID] = projectAgent.Role
			}
		}
		for _, project := range projects {
			project.IsMultipleProjectTimezoneEnabled = C.IsMultipleProjectTimezoneEnabled(project.ID)
			resp[projectRoleMap[project.ID]] = append(resp[projectRoleMap[project.ID]], project)
		}
	}
	if C.EnableDemoReadAccess() {
		trimmedDemoProjects := make([]model.Project, 0)
		demoProjects, _ := store.GetStore().GetProjectsByIDs(C.GetConfig().DemoProjectIds)
		for _, project := range demoProjects {
			project.Token = ""
			project.PrivateToken = ""
			project.InteractionSettings = postgres.Jsonb{}
			project.SalesforceTouchPoints = postgres.Jsonb{}
			project.HubspotTouchPoints = postgres.Jsonb{}
			project.JobsMetadata = nil
			project.ChannelGroupRules = postgres.Jsonb{}
			trimmedDemoProjects = append(trimmedDemoProjects, project)
		}
		for _, project := range trimmedDemoProjects {
			if !H.IsDemoProjectInAuthorizedProjects(authorizedProjects.([]uint64), project.ID) {
				project.IsMultipleProjectTimezoneEnabled = C.IsMultipleProjectTimezoneEnabled(project.ID)
				resp[1] = append(resp[1], project)
			}
		}
	}
	c.JSON(http.StatusOK, resp)
	return
}

func GetDemoProjects(c *gin.Context) {
	demoProjects := make([]uint64, 0)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projects := C.GetConfig().DemoProjectIds

	if C.IsLoggedInUserWhitelistedForProjectAnalytics(loggedInAgentUUID) {
		c.JSON(http.StatusOK, demoProjects)
		return
	} else {
		c.JSON(http.StatusOK, projects)
		return
	}
}
