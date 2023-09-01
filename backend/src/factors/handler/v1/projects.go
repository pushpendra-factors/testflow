package v1

import (
	"fmt"
	"net/http"

	"factors/model/model"
	"factors/model/store"
	U "factors/util"

	mid "factors/middleware"

	"github.com/gin-gonic/gin"
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
	projects, errCode := store.GetStore().GetProjectsByIDs(authorizedProjects.([]int64))
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	} else if errCode == http.StatusNoContent || errCode == http.StatusBadRequest {
		resp := make(map[string]interface{})
		resp["projects"] = []model.Project{}
	}
	projectRoleMap := make(map[int64]uint64)
	resp := make(map[uint64][]interface{})
	if len(projects) > 0 {
		projectAgentMappings, errCode := store.GetStore().GetProjectAgentMappingsByProjectIds(authorizedProjects.([]int64))
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
			resp[projectRoleMap[project.ID]] = append(resp[projectRoleMap[project.ID]], MapProjectToString(project))
		}
	}
	c.JSON(http.StatusOK, resp)
	return
}

func MapProjectToString(project model.Project) model.ProjectString {
	return model.ProjectString{
		ID:                    fmt.Sprintf("%v", project.ID),
		Name:                  project.Name,
		ProfilePicture:        project.ProfilePicture,
		Token:                 project.Token,
		PrivateToken:          project.PrivateToken,
		CreatedAt:             project.CreatedAt,
		UpdatedAt:             project.UpdatedAt,
		ProjectURI:            project.ProjectURI,
		TimeFormat:            project.TimeFormat,
		DateFormat:            project.DateFormat,
		TimeZone:              project.TimeZone,
		InteractionSettings:   project.InteractionSettings,
		SalesforceTouchPoints: project.SalesforceTouchPoints,
		HubspotTouchPoints:    project.HubspotTouchPoints,
		JobsMetadata:          project.JobsMetadata,
		ChannelGroupRules:     project.ChannelGroupRules,
	}
}
