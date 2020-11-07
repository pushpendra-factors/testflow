package v1

import (
	"net/http"

	M "factors/model"
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
	authorizedProjects := U.GetScopeByKey(c, "authorizedProjects")
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projects, errCode := M.GetProjectsByIDs(authorizedProjects.([]uint64))
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	} else if errCode == http.StatusNoContent || errCode == http.StatusBadRequest {
		resp := make(map[string]interface{})
		resp["projects"] = []M.Project{}
		c.JSON(http.StatusNotFound, resp)
		return
	}
	projectAgentMappings, errCode := M.GetProjectAgentMappingsByProjectIds(authorizedProjects.([]uint64))
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}
	projectRoleMap := make(map[uint64]uint64)
	for _, projectAgent := range projectAgentMappings {
		if projectAgent.AgentUUID == loggedInAgentUUID {
			projectRoleMap[projectAgent.ProjectID] = projectAgent.Role
		}
	}
	resp := make(map[uint64][]interface{})
	for _, project := range projects {
		resp[projectRoleMap[project.ID]] = append(resp[projectRoleMap[project.ID]], project)
	}
	c.JSON(http.StatusOK, resp)
	return
}
