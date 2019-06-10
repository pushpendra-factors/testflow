package handler

import (
	"factors/handler/helpers"
	mid "factors/middleware"
	M "factors/model"
	"fmt"
	"net/http"
	"time"

	C "factors/config"
	U "factors/util"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// curl -X POST --data "email=value1" http://localhost:8080/accounts/signup
func SignUp(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	type signupParams struct {
		Email    string `json:"email" binding:"required"`
		PlanCode string `json:"plan_code"`
	}
	params := signupParams{}
	err := c.BindJSON(&params)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse SignUpParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	email := params.Email
	planCode := params.PlanCode
	if planCode == "" {
		planCode = M.FreePlanCode
	}

	if existingAgent, code := M.GetAgentByEmail(email); code == http.StatusInternalServerError {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if code == http.StatusFound {
		if !existingAgent.IsEmailVerified {
			err = sendSignUpEmail(existingAgent)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}
		c.AbortWithStatus(http.StatusFound)
		return
	}
	createAgentParams := M.CreateAgentParams{
		Agent:    &M.Agent{Email: email},
		PlanCode: planCode,
	}
	createAgentResp, code := M.CreateAgentWithDependencies(&createAgentParams)
	if code == http.StatusInternalServerError {
		log.WithField("email", email).Error("Failed To Create Agent")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	agent := createAgentResp.Agent
	err = sendSignUpEmail(agent)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	// Set Cookie with exp 1 day. After that the agent will be forced to set password
	// And probably redirect to default project view
	resp := map[string]string{
		"status": "success",
	}
	c.JSON(http.StatusCreated, resp)
}

func sendSignUpEmail(agent *M.Agent) error {
	authToken, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, time.Second*helpers.SecondsInFifteenDays)
	if err != nil {
		log.WithField("email", agent.Email).Error("Failed To Create Agent Auth Token")
		return err
	}

	fe_host := C.GetProtocol() + C.GetAPPDomain()
	link := fmt.Sprintf("%s/#/activate?token=%s", fe_host, authToken)

	log.WithField("link", link).Debug("Activation LInk")

	sub, text, html := U.CreateActivationTemplate(link)

	err = C.GetServices().Mailer.SendMail(agent.Email, C.GetFactorsSenderEmail(), sub, html, text)
	if err != nil {
		log.WithError(err).Error("Failed to send activation email")
	}
	return err
}
