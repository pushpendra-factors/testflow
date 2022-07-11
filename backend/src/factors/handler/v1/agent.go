package v1

import (
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetProjectAgentsHandler godoc
// @Summary Gets agents list for the given project id.
// @Tags V1Api,ProjectAdmin
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} v1.AgentInfoWithProjectMapping
// @Router /{project_id}/v1/agents [get]
func GetProjectAgentsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	projectAgentMappings, errCode := store.GetStore().GetProjectAgentMappingsByProjectId(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}

	agentUUIDs := make([]string, 0, 0)
	for _, pam := range projectAgentMappings {
		agentUUIDs = append(agentUUIDs, pam.AgentUUID)
	}

	agents, errCode := store.GetStore().GetAgentsByUUIDs(agentUUIDs)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}

	agentInfos := model.CreateAgentInfos(agents)
	agentInfoMap := make(map[string]*model.AgentInfo)
	for _, agentInfo := range agentInfos {
		agentInfoMap[agentInfo.UUID] = agentInfo
	}

	agentWithProjectMapping := make([]AgentInfoWithProjectMapping, 0)
	for _, pam := range projectAgentMappings {
		agentWithProjectMapping = append(agentWithProjectMapping, mapAgentInfoWithProjectMapping(pam, agentInfoMap[pam.AgentUUID]))
	}

	c.JSON(http.StatusOK, agentWithProjectMapping)
}

func mapAgentInfoWithProjectMapping(pam model.ProjectAgentMapping, agent *model.AgentInfo) AgentInfoWithProjectMapping {
	agentWithProject := AgentInfoWithProjectMapping{}
	agentWithProject.UUID = agent.UUID
	agentWithProject.Email = agent.Email
	agentWithProject.FirstName = agent.FirstName
	agentWithProject.LastName = agent.LastName
	agentWithProject.IsEmailVerified = agent.IsEmailVerified
	agentWithProject.LastLoggedIn = agent.LastLoggedIn
	agentWithProject.Phone = agent.Phone
	agentWithProject.ProjectID = fmt.Sprintf("%v", pam.ProjectID)
	agentWithProject.Role = pam.Role
	agentWithProject.InvitedBy = pam.InvitedBy
	agentWithProject.CreatedAt = pam.CreatedAt
	agentWithProject.UpdatedAt = pam.UpdatedAt
	return agentWithProject
}

type AgentInfoWithProjectMapping struct {
	UUID            string     `json:"uuid"`
	Email           string     `json:"email"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	IsEmailVerified bool       `json:"is_email_verified"`
	LastLoggedIn    *time.Time `json:"last_logged_in"`
	Phone           string     `json:"phone"`
	ProjectID       string     `json:"project_id"`
	Role            uint64     `json:"role"`
	InvitedBy       *string    `json:"invited_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
