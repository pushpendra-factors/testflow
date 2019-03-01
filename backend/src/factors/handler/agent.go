package handler

import (
	C "factors/config"
	"factors/handler/helpers"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

	params, err := getSignInParams(c)
	if err != nil {
		log.WithError(err).Error("Failed to parse SignInParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	email := params.Email
	password := params.Password

	agent, code := M.GetAgentByEmail(email)
	if code == http.StatusInternalServerError {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if code == http.StatusNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if !M.IsPasswordAndHashEqual(password, agent.Password) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ts := time.Now().UTC()
	errCode := M.UpdateAgentLastLoginInfo(email, ts)
	if errCode != http.StatusAccepted {
		log.Error("Failed to update Agent lastLoginInfo")
	}

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, helpers.SecondsInOneMonth*time.Second)

	domain := C.GetCookieDomian()
	c.SetCookie(helpers.FactorsSessionCookieName, cookieData, helpers.SecondsInOneMonth, "/", domain, C.UseSecureCookie(), false)
	resp := map[string]string{
		"status": "success",
	}
	c.JSON(http.StatusOK, resp)
}

// curl -X GET  http://localhost:8080/agents/signout
func Signout(c *gin.Context) {

	domain := C.GetCookieDomian()
	c.SetCookie(helpers.FactorsSessionCookieName, "", helpers.ExpireCookie, "/", domain, C.UseSecureCookie(), false)
	// redirect to login
	resp := map[string]string{
		"status": "success",
	}
	c.JSON(http.StatusOK, resp)
}

type agentInviteParams struct {
	Email string `json:"email" binding:"required"`
}

func getAgentInviteParams(c *gin.Context) (*agentInviteParams, error) {
	params := agentInviteParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// curl -X POST -d '{"email":"value1"}' http://localhost:8080/:project_id/agents/invite -v
func AgentInvite(c *gin.Context) {

	params, err := getAgentInviteParams(c)
	if err != nil {
		log.WithError(err).Error("Failed to parse AgentInviteParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	emailOfAgentToInvite := params.Email

	invitedByAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	// check if project has less than 500 agents

	var invitedAgent *M.Agent
	var errCode int

	invitedAgent, errCode = M.GetAgentByEmail(emailOfAgentToInvite)
	if errCode == http.StatusInternalServerError {
		log.Error("Failed to GetAgentByEmail")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	createNewAgent := errCode == http.StatusNotFound

	if createNewAgent {
		invitedAgent, errCode = M.CreateAgent(&M.Agent{Email: emailOfAgentToInvite, InvitedBy: &invitedByAgentUUID})
		if errCode == http.StatusInternalServerError {
			log.Error("Failed to CreateAgent")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	pam, errCode := M.CreateProjectAgentMapping(&M.ProjectAgentMapping{
		ProjectID: projectId,
		AgentUUID: invitedAgent.UUID,
		InvitedBy: &invitedByAgentUUID,
		Role:      M.AGENT,
	})
	if errCode == http.StatusInternalServerError {
		log.Error("Failed to createProjectAgentMapping")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	sendVerifyProfileLink := createNewAgent

	// Send email
	// You have been added to this project

	if sendVerifyProfileLink {
		authToken, err := helpers.GetAuthData(invitedAgent.Email, invitedAgent.UUID, invitedAgent.Salt, helpers.SecondsInFifteenDays*time.Second)
		if err != nil {
			log.WithError(err).Error("Failed to create auth token for invited agent")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		fe_host := C.GetProtocol() + C.GetAPPDomain()
		link := fmt.Sprintf("%s/#/activate?token=%s", fe_host, authToken)
		log.WithField("link", link).Debugf("Verification LInk")
	}

	c.JSON(http.StatusCreated, pam)
	return
}

type agentVerifyParams struct {
	FirstName string `json:"first_name" binding:"required"`
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
	params, err := getAgentVerifyParams(c)
	if err != nil {
		log.WithError(err).Error("Failed to parse AgentVerifyParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	ts := time.Now().UTC()
	errCode := M.UpdateAgentVerificationDetails(agentUUID, params.Password, params.FirstName, params.LastName, true, ts)
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

	params, err := getResetPasswordEmailParams(c)
	if err != nil {
		log.WithError(err).Error("Failed to parse AgentVerifyParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	email := params.Email

	agent, errCode := M.GetAgentByEmail(email)
	if errCode == http.StatusInternalServerError {
		log.Error("Failed to GetAgentByEmail")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if errCode == http.StatusNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	err = sendAgentResetPasswordEmail(agent)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"status": "success",
	}
	c.JSON(http.StatusOK, resp)
	return
}

func sendAgentResetPasswordEmail(agent *M.Agent) error {
	authToken, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, time.Second*helpers.SecondsInOneDay)
	if err != nil {
		log.WithField("email", agent.Email).Error("Failed To Create Agent Auth Token")
		return err
	}
	fe_host := C.GetProtocol() + C.GetAPPDomain()
	link := fmt.Sprintf("%s/#/setpassword?token=%s", fe_host, authToken)
	log.WithField("link", link).Debugf("Reset Password LInk")

	log.WithField("email", agent.Email).Info("Sending Agent Password Reset Email")

	err = C.GetServices().Mailer.SendMail(agent.Email, C.GetFactorsSenderEmail(), "ResetPassord Factors account", link, link)
	if err != nil {
		log.WithError(err).Error("Sending Agent Password Reset Email")
	}
	return err
}

type setPasswordParams struct {
	Password string `json:"password"`
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
	params, err := getSetPasswordParams(c)
	if err != nil {
		log.WithError(err).Error("Failed to parse getSetPasswordParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	ts := time.Now().UTC()

	errCode := M.UpdateAgentPassword(agentUUID, params.Password, ts)
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if errCode == http.StatusNoContent {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	resp := map[string]string{
		"status": "success",
	}
	c.JSON(http.StatusOK, resp)
}
