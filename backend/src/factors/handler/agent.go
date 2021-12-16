package handler

import (
	C "factors/config"
	"factors/handler/helpers"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type signInParams struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func getSignInParams(c *gin.Context) (*signInParams, error) {
	params := signInParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// curl -X POST -d '{"email":"value1", "password":"value1"}' http://localhost:8080/agents/signin -v
func Signin(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	params, err := getSignInParams(c)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse SignInParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	email := params.Email
	password := params.Password

	agent, code := store.GetStore().GetAgentByEmail(email)
	if code == http.StatusInternalServerError {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if code == http.StatusNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if !model.IsPasswordAndHashEqual(password, agent.Password) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ts := time.Now().UTC()
	errCode := store.GetStore().UpdateAgentLastLoginInfo(agent.UUID, ts)
	if errCode != http.StatusAccepted {
		logCtx.WithField("email", email).Error("Failed to update Agent lastLoginInfo")
	}

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, helpers.SecondsInOneMonth*time.Second)

	domain := C.GetCookieDomian()

	cookie := C.UseSecureCookie()
	httpOnly := C.UseHTTPOnlyCookie()
	if C.IsDevBox() {
		cookie = true
		httpOnly = true
		c.SetSameSite(http.SameSiteNoneMode)
	}
	c.SetCookie(C.GetFactorsCookieName(), cookieData, helpers.SecondsInOneMonth, "/", domain, cookie, httpOnly)
	resp := map[string]string{
		"status": "success",
	}
	c.JSON(http.StatusOK, resp)
}

// curl -X GET  http://localhost:8080/agents/signout
func Signout(c *gin.Context) {

	domain := C.GetCookieDomian()
	c.SetCookie(C.GetFactorsCookieName(), "", helpers.ExpireCookie, "/", domain, C.UseSecureCookie(), C.UseHTTPOnlyCookie())
	// redirect to login
	resp := map[string]string{
		"status": "success",
	}
	c.JSON(http.StatusOK, resp)
}
type agentInviteParams struct {
	Email string `json:"email" binding:"required"`
	Role  int64  `json:"role"`
}

func getAgentInviteParams(c *gin.Context) (*agentInviteParams, error) {
	params := agentInviteParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}
func getAgentBatchInviteParams(c *gin.Context) (*[]agentInviteParams, error) {
	params := []agentInviteParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// curl -X POST -d '{"email":"value1"}' http://localhost:8080/:project_id/agents/invite -v
// AgentInvite godoc
// @Summary To invite an agent to the given project id.
// @Tags ProjectAdmin
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param invite body handler.agentInviteParams true "Invite"
// @Success 201 {string} json "{"status": "success", "agents": agentInfoMap, "project_agent_mappings": projectAgentMappings}"
// @Router /{project_id}/agents/invite [post]
func AgentInvite(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	params, err := getAgentInviteParams(c)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse AgentInviteParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	emailOfAgentToInvite := params.Email
	roleOfAgent := params.Role

	invitedByAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	createProjectAgentMapping, errCode := store.GetStore().IsNewProjectAgentMappingCreationAllowed(projectId, emailOfAgentToInvite)
	if errCode != http.StatusOK {
		c.AbortWithStatus(errCode)
		return
	}

	if !createProjectAgentMapping {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	project, errCode := store.GetStore().GetProject(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	invitedAgent, errCode := store.GetStore().GetAgentByEmail(emailOfAgentToInvite)
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Failed to GetAgentByEmail")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	createNewAgent := errCode == http.StatusNotFound

	if createNewAgent {
		createAgentParams := model.CreateAgentParams{
			Agent:    &model.Agent{Email: emailOfAgentToInvite, InvitedBy: &invitedByAgentUUID},
			PlanCode: model.FreePlanCode,
		}
		resp, errCode := store.GetStore().CreateAgentWithDependencies(&createAgentParams)
		if errCode == http.StatusInternalServerError {
			logCtx.Error("Failed to CreateAgent")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		invitedAgent = resp.Agent
	}

	newProjectAgentRole := uint64(model.AGENT)
	if roleOfAgent == model.ADMIN {
		newProjectAgentRole = uint64(model.ADMIN)
	}
	pam, errCode := store.GetStore().CreateProjectAgentMappingWithDependencies(
		&model.ProjectAgentMapping{
			ProjectID: projectId,
			AgentUUID: invitedAgent.UUID,
			InvitedBy: &invitedByAgentUUID,
			Role:      newProjectAgentRole,
		})
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Failed to createProjectAgentMapping")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if errCode == http.StatusFound {
		c.AbortWithStatusJSON(http.StatusFound, gin.H{"error": "User is already mapped to project"})
		return
	}

	sendVerifyProfileLink := createNewAgent

	// Send email
	// You have been added to this project
	link := ""
	if sendVerifyProfileLink {
		authToken, err := helpers.GetAuthData(invitedAgent.Email, invitedAgent.UUID, invitedAgent.Salt, helpers.SecondsInFifteenDays*time.Second)
		if err != nil {
			wrapErr := errors.Wrap(err, "Failed to create auth token for invited agent")
			logCtx.WithError(wrapErr).Error("Failed to create auth token for invited agent")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		fe_host := C.GetProtocol() + C.GetAPPDomain()
		link = fmt.Sprintf("%s/activate?token=%s", fe_host, authToken)
		logCtx.WithField("link", link).Debugf("Verification LInk")
	}

	invitedAgentInfo := model.CreateAgentInfo(invitedAgent)
	agentInfoMap := make(map[string]*model.AgentInfo)

	agentInfoMap[invitedAgentInfo.UUID] = invitedAgentInfo

	sub, text, html := U.CreateAgentInviteTemplate(project.Name, link)
	err = C.GetServices().Mailer.SendMail(invitedAgent.Email, C.GetFactorsSenderEmail(), sub, html, text)
	if err != nil {
		logCtx.WithError(err).Error("Failed to send activation email")
		c.AbortWithStatusJSON(http.StatusFound, gin.H{"error": "Failed to send invitation email"})
		return
	}

	resp := make(map[string]interface{})
	resp["status"] = "success"
	resp["agents"] = agentInfoMap
	resp["project_agent_mappings"] = []model.ProjectAgentMapping{*pam}
	
	c.JSON(http.StatusCreated, resp)
	return
}
func AgentInviteBatch(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	params, err := getAgentBatchInviteParams(c)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse AgentInviteParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	invitedByAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	project, errCode := store.GetStore().GetProject(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	agentInfoMap := make(map[string]*model.AgentInfo)
	pam := []model.ProjectAgentMapping{}
	failedToInviteAgentIndexes:= make(map[int]bool)
	for idx, agentDetail := range *params {
		emailOfAgentToInvite := agentDetail.Email
		roleOfAgent := agentDetail.Role

		createProjectAgentMapping, errCode := store.GetStore().IsNewProjectAgentMappingCreationAllowed(projectId, emailOfAgentToInvite)
		if errCode != http.StatusOK {
			failedToInviteAgentIndexes[idx] = true
			continue
		}

		if !createProjectAgentMapping {
			failedToInviteAgentIndexes[idx] = true
			continue
		}

		invitedAgent, errCode := store.GetStore().GetAgentByEmail(emailOfAgentToInvite)
		if errCode == http.StatusInternalServerError {
			logCtx.Error("Failed to GetAgentByEmail")
			failedToInviteAgentIndexes[idx] = true
			continue
		}

		createNewAgent := errCode == http.StatusNotFound

		if createNewAgent {
			createAgentParams := model.CreateAgentParams{
				Agent:    &model.Agent{Email: emailOfAgentToInvite, InvitedBy: &invitedByAgentUUID},
				PlanCode: model.FreePlanCode,
			}
			resp, errCode := store.GetStore().CreateAgentWithDependencies(&createAgentParams)
			if errCode == http.StatusInternalServerError {
				logCtx.Error("Failed to CreateAgent")
				failedToInviteAgentIndexes[idx] = true
				continue
			}
			invitedAgent = resp.Agent
		}
		newProjectAgentRole := uint64(model.AGENT)

		if roleOfAgent == model.ADMIN {
			newProjectAgentRole = uint64(model.ADMIN)
		}
		projectAgentMapping, errCode := store.GetStore().CreateProjectAgentMappingWithDependencies(
			&model.ProjectAgentMapping{
				ProjectID: projectId,
				AgentUUID: invitedAgent.UUID,
				InvitedBy: &invitedByAgentUUID,
				Role:      newProjectAgentRole,
			})
		if errCode == http.StatusInternalServerError {
			logCtx.Error("Failed to createProjectAgentMapping")
			failedToInviteAgentIndexes[idx] = true
			continue
		} else if errCode == http.StatusFound {
			//c.AbortWithStatusJSON(http.StatusFound, gin.H{"error": "User is already mapped to project"})
			failedToInviteAgentIndexes[idx] = true
			continue
		}

		sendVerifyProfileLink := createNewAgent

		// Send email
		// You have been added to this project
		link := ""
		if sendVerifyProfileLink {
			authToken, err := helpers.GetAuthData(invitedAgent.Email, invitedAgent.UUID, invitedAgent.Salt, helpers.SecondsInFifteenDays*time.Second)
			if err != nil {
				wrapErr := errors.Wrap(err, "Failed to create auth token for invited agent")
				logCtx.WithError(wrapErr).Error("Failed to create auth token for invited agent")
				failedToInviteAgentIndexes[idx] = true
				continue
			}
			fe_host := C.GetProtocol() + C.GetAPPDomain()
			link = fmt.Sprintf("%s/activate?token=%s", fe_host, authToken)
			logCtx.WithField("link", link).Debugf("Verification LInk")
		}

		invitedAgentInfo := model.CreateAgentInfo(invitedAgent)
		

		agentInfoMap[invitedAgentInfo.UUID] = invitedAgentInfo

		sub, text, html := U.CreateAgentInviteTemplate(project.Name, link)
		err = C.GetServices().Mailer.SendMail(invitedAgent.Email, C.GetFactorsSenderEmail(), sub, html, text)
		if err != nil {
			logCtx.WithError(err).Error("Failed to send activation email")
			//c.AbortWithStatusJSON(http.StatusFound, gin.H{"error": "Failed to send invitation email"})
			failedToInviteAgentIndexes[idx] = true
			continue
		}
		pam = append(pam, *projectAgentMapping)
	}
	resp := make(map[string]interface{})
	resp["status"] = "success"
	resp["agents"] = agentInfoMap
	resp["project_agent_mappings"] = pam
	resp["failed_to_invite_agent_idx"] = failedToInviteAgentIndexes
	c.JSON(http.StatusCreated, resp)
	return
}

// curl -X PUT -d '{"email":"value1"}' http://localhost:8080/:project_id/agents/update -v
// AgentUpdate godoc
// @Summary To update an agent from the given project id.
// @Tags ProjectAdmin
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param update body handler.updateProjectAgentParams true "Update"
// @Success 201 {string} json "{"status": "success"}"
// @Router /{project_id}/agents/update [put]
func AgentUpdate(c *gin.Context) {
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"reqId":         U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"loggedInAgent": loggedInAgentUUID,
		"projectId":     projectId,
	})

	params, err := getUpdateProjectAgentParams(c)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse removeProjectAgentParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	agentUUIDToEdit := params.AgentUUID
	roleIDToUpdate := params.Role
	loggedInAgentPAM, errCode := store.GetStore().GetProjectAgentMapping(projectId, loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		logCtx.Errorln("Failed to fetch loggedInAgentPAM")
		return
	}

	if !isAdmin(loggedInAgentPAM.Role) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Agent Edit is allowed only fro admins"})
		return
	}

	if !(roleIDToUpdate == 1 || roleIDToUpdate == 2) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid RoleID"})
		return
	}

	errCode = store.GetStore().EditProjectAgentMapping(projectId, agentUUIDToEdit, roleIDToUpdate)
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	}
	agentMappingDetails, errCode := store.GetStore().GetProjectAgentMapping(projectId, agentUUIDToEdit)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		logCtx.Errorln("Failed to fetch agentMappingDetails")
		return
	}
	c.JSON(http.StatusCreated, agentMappingDetails)
	return
}

type removeProjectAgentParams struct {
	AgentUUID string `json:"agent_uuid" binding:"required"`
}

type updateProjectAgentParams struct {
	AgentUUID string `json:"agent_uuid" binding:"required"`
	Role      int64  `json:"role"`
}

func getRemoveProjectAgentParams(c *gin.Context) (*removeProjectAgentParams, error) {
	params := removeProjectAgentParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func getUpdateProjectAgentParams(c *gin.Context) (*updateProjectAgentParams, error) {
	params := updateProjectAgentParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// curl -X POST -d '{"agent_uuid":"value1"}' http://localhost:8080/:project_id/agents/remove -v
// RemoveProjectAgent godoc
// @Summary To remove an agent from the given project id.
// @Tags ProjectAdmin
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param remove body handler.removeProjectAgentParams true "Remove"
// @Success 202 {string} json "{"project_id": uint64, "agent_uuid": string}"
// @Router /{project_id}/agents/remove [put]
func RemoveProjectAgent(c *gin.Context) {
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"reqId":         U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"loggedInAgent": loggedInAgentUUID,
		"projectId":     projectId,
	})

	params, err := getRemoveProjectAgentParams(c)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse removeProjectAgentParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	agentUUIDToRemove := params.AgentUUID

	loggedInAgentPAM, errCode := store.GetStore().GetProjectAgentMapping(projectId, loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		logCtx.Errorln("Failed to fetch loggedInAgentPAM")
		return
	}

	if isAdmin(loggedInAgentPAM.Role) {
		if loggedInAgentUUID == agentUUIDToRemove {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Agent Admin cannot remove himself"})
			return
		}
	}

	if !isAdmin(loggedInAgentPAM.Role) {
		if loggedInAgentUUID != agentUUIDToRemove {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Agent User cannot remove others"})
			return
		}
	}

	errCode = store.GetStore().DeleteProjectAgentMapping(projectId, agentUUIDToRemove)
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	}

	resp := map[string]interface{}{
		"project_id": projectId,
		"agent_uuid": agentUUIDToRemove,
	}
	c.JSON(http.StatusAccepted, resp)

}

type agentVerifyParams struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password" binding:"required"`
}

func getAgentVerifyParams(c *gin.Context) (*agentVerifyParams, error) {
	params := agentVerifyParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// curl -X POST -d '{"first_name":"value1", "last_name":"value1", "password":"value"}' http://localhost:8080/agents/activate?token=value -v
func AgentActivate(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	params, err := getAgentVerifyParams(c)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse AgentVerifyParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	ts := time.Now().UTC()
	errCode := store.GetStore().UpdateAgentVerificationDetails(agentUUID, params.Password, params.FirstName, params.LastName, true, ts)
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if errCode == http.StatusNoContent {
		c.AbortWithStatus(http.StatusNoContent)
		return
	} else if errCode == http.StatusBadRequest {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	_, errCode = store.GetStore().CreateDefaultProjectForAgent(agentUUID)
	if errCode != http.StatusConflict && errCode != http.StatusCreated {
		logCtx.WithField("agent_uuid", agentUUID).Error("Failed to create default project for agent.")
	}

	resp := map[string]string{
		"status": "success",
	}
	c.JSON(http.StatusOK, resp)
	return
}

type resetPasswordEmailParams struct {
	Email string `json:"email" binding:"required"`
}

func getResetPasswordEmailParams(c *gin.Context) (*resetPasswordEmailParams, error) {
	params := resetPasswordEmailParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// curl -X POST -d '{"email":"value1"}' http://localhost:8080/:project_id/agents/forgotpassword -v
func AgentGenerateResetPasswordLinkEmail(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	params, err := getResetPasswordEmailParams(c)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse AgentVerifyParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	email := params.Email

	agent, errCode := store.GetStore().GetAgentByEmail(email)
	if errCode == http.StatusInternalServerError {
		logCtx.WithField("email", email).Error("Failed to GetAgentByEmail")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if errCode == http.StatusNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	err = sendAgentResetPasswordEmail(agent)
	if err != nil {
		logCtx.WithField("email", email).Error("Failed to sendAgentResetPasswordEmail")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"status": "success",
	}
	c.JSON(http.StatusOK, resp)
	return
}

func sendAgentResetPasswordEmail(agent *model.Agent) error {
	authToken, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, time.Second*helpers.SecondsInOneDay)
	if err != nil {
		return err
	}
	fe_host := C.GetProtocol() + C.GetAPPDomain()
	link := fmt.Sprintf("%s/setpassword?token=%s", fe_host, authToken)
	log.WithField("link", link).Debug("Reset Password LInk")

	log.WithField("email", agent.Email).Debug("Sending Agent Password Reset Email")

	sub, text, html := U.CreateForgotPasswordTemplate(agent.Email, link)

	err = C.GetServices().Mailer.SendMail(agent.Email, C.GetFactorsSenderEmail(), sub, html, text)
	return err
}

type setPasswordParams struct {
	Password string `json:"password" binding:"required"`
}

func getSetPasswordParams(c *gin.Context) (*setPasswordParams, error) {
	params := setPasswordParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func AgentSetPassword(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	params, err := getSetPasswordParams(c)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse getSetPasswordParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	ts := time.Now().UTC()

	errCode := store.GetStore().UpdateAgentPassword(agentUUID, params.Password, ts)
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if errCode == http.StatusNoContent {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if errCode == http.StatusBadRequest {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	resp := map[string]string{
		"status": "success",
	}
	c.JSON(http.StatusOK, resp)
}

func AgentInfo(c *gin.Context) {
	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	agentInfo, errCode := store.GetStore().GetAgentInfo(currentAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}

	c.JSON(http.StatusOK, agentInfo)
}

// GetProjectAgentsHandler godoc
// @Summary Gets agents list for the given project id.
// @Tags ProjectAdmin
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "{"agents": agentInfoMap, "project_agent_mappings": projectAgentMappings}"
// @Router /{project_id}/agents [get]
func GetProjectAgentsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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

	resp := make(map[string]interface{})
	resp["agents"] = agentInfoMap
	resp["project_agent_mappings"] = projectAgentMappings

	c.JSON(http.StatusOK, resp)
}

func GetAgentBillingAccount(c *gin.Context) {
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	bA, errCode := store.GetStore().GetBillingAccountByAgentUUID(loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}

	projects, errCode := store.GetStore().GetProjectsUnderBillingAccountID(bA.ID)

	projectIDs := make([]uint64, len(projects), len(projects))
	for i := range projects {
		projectIDs[i] = projects[i].ID
	}

	plan, errCode := store.GetStore().GetPlanByID(bA.PlanID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}

	agents, errCode := store.GetStore().GetAgentsByProjectIDs(projectIDs)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}

	agentsInfo := model.CreateAgentInfos(agents)

	resp := make(map[string]interface{})
	resp["billing_account"] = bA
	resp["projects"] = projects
	resp["agents"] = agentsInfo
	resp["plan"] = plan
	resp["available_plans"] = map[string]string{
		model.FreePlanCode:    "Free",
		model.StartupPlanCode: "Startup",
	}
	c.JSON(http.StatusOK, resp)
}

type updateAgentBillingAccParams struct {
	OrganizationName string `json:"organization_name"`
	BillingAddress   string `json:"billing_address"`
	Pincode          string `json:"pincode"`
	PhoneNo          string `json:"phone_no"`
	PlanCode         string `json:"plan_code"`
}

func getUpdateAgentBillingAccountParams(c *gin.Context) (*updateAgentBillingAccParams, error) {
	params := updateAgentBillingAccParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func UpdateAgentBillingAccount(c *gin.Context) {
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	params, err := getUpdateAgentBillingAccountParams(c)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse getUpdateAgentBillingAccountParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	bA, errCode := store.GetStore().GetBillingAccountByAgentUUID(loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}

	currPlan, errCode := store.GetStore().GetPlanByID(bA.PlanID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}

	newPlan, errCode := store.GetStore().GetPlanByCode(params.PlanCode)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}

	planToSet := currPlan
	if newPlan.ID != currPlan.ID {
		planToSet = newPlan
	}

	errCode = store.GetStore().UpdateBillingAccount(bA.ID, planToSet.ID, params.OrganizationName, params.BillingAddress, params.Pincode, params.PhoneNo)
	if errCode != http.StatusAccepted {
		c.AbortWithStatus(errCode)
		return
	}

	// Fetch the updated billing_account and return
	bA, errCode = store.GetStore().GetBillingAccountByAgentUUID(loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}

	resp := make(map[string]interface{})
	resp["billing_account"] = bA
	resp["plan"] = planToSet
	c.JSON(http.StatusOK, resp)
}

type updateAgentParams struct {
	FirstName             string `json:"first_name"`
	LastName              string `json:"last_name"`
	Phone                 string `json:"phone"`
	IsOnboardingFlowSeen bool   `json:"is_onboarding_flow_seen"`
}

func getUpdateAgentParams(c *gin.Context) (*updateAgentParams, error) {
	params := updateAgentParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func UpdateAgentInfo(c *gin.Context) {
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})
	params, err := getUpdateAgentParams(c)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse UpdateAgent params")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	errCode := store.GetStore().UpdateAgentInformation(loggedInAgentUUID, params.FirstName, params.LastName, params.Phone, params.IsOnboardingFlowSeen)
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	}

	agent, errCode := store.GetStore().GetAgentInfo(loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	resp := make(map[string]interface{})
	resp["status"] = "success"
	resp["agent"] = agent
	c.JSON(http.StatusOK, agent)
}

type updateAgentPasswordParams struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

func getUpdateAgentPasswordParams(c *gin.Context) (*updateAgentPasswordParams, error) {
	params := updateAgentPasswordParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func UpdateAgentPassword(c *gin.Context) {
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})
	params, err := getUpdateAgentPasswordParams(c)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse UpdateAgentPassword params")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	agent, errCode := store.GetStore().GetAgentByUUID(loggedInAgentUUID)
	if errCode == http.StatusInternalServerError {
		logCtx.WithField("uuid", loggedInAgentUUID).Error("Failed to GetAgentByUUID")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if errCode == http.StatusNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if !model.IsPasswordAndHashEqual(params.CurrentPassword, agent.Password) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Incorrect Current Password"})
		return
	}

	errCode = store.GetStore().UpdateAgentPassword(loggedInAgentUUID, params.NewPassword, time.Now().UTC())
	c.Status(errCode)
}

func isAdmin(role uint64) bool {
	if role == model.ADMIN {
		return true
	}
	return false
}
