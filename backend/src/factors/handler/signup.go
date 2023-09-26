package handler

import (
	"factors/handler/helpers"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"net/http"
	"strings"
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
		Email      string `json:"email" binding:"required"`
		Phone      string `json:"phone"`
		CompanyURL string `json:"company_url"`
		PlanCode   string `json:"plan_code"`
	}
	params := signupParams{}
	err := c.BindJSON(&params)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse SignUpParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Basic email sanity check.
	if !U.IsEmail(strings.TrimSpace(params.Email)) {
		logCtx.WithError(err).Error("Invalid email provided.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if U.IsPersonalEmail(strings.TrimSpace(params.Email)) {
		logCtx.WithError(err).Error("Personal email is not allowed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	email := params.Email
	phone := params.Phone
	planCode := params.PlanCode
	companyUrl := params.CompanyURL
	subscribeNewsletter := true
	if planCode == "" {
		planCode = model.FreePlanCode
	}

	if existingAgent, code := store.GetStore().GetAgentByEmail(email); code == http.StatusInternalServerError {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if code == http.StatusFound {

		if existingAgent.IsAuth0User {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already signed up with OAuth flow with this email."})
			return
		} else if !existingAgent.IsEmailVerified {
			err = sendSignUpEmail(existingAgent)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}
		c.AbortWithStatus(http.StatusFound)
		return
	}

	createAgentParams := model.CreateAgentParams{
		Agent:    &model.Agent{Email: email, Phone: phone, CompanyURL: companyUrl, SubscribeNewsletter: subscribeNewsletter},
		PlanCode: planCode,
	}
	createAgentResp, code := store.GetStore().CreateAgentWithDependencies(&createAgentParams)
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

	code = onboardingMailModoAPICall(agent)
	if code != http.StatusOK {
		log.WithField("status_code", code).
			WithField("email", email).
			Error("Failed To Send Onboarding Mail")
	}

	code = onboardingSlackAPICall(agent)
	if code != http.StatusOK {
		log.WithField("email", email).
			WithField("status_code", code).
			Error("Failed To Send Onboarding Slack")
	}

	// Set Cookie with exp 1 day. After that the agent will be forced to set password
	// And probably redirect to default project view
	resp := map[string]string{
		"status": "success",
	}
	c.JSON(http.StatusCreated, resp)
}

func sendSignUpEmail(agent *model.Agent) error {
	authToken, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, time.Second*helpers.SecondsInFifteenDays)
	if err != nil {
		log.WithField("email", agent.Email).Error("Failed To Create Agent Auth Token")
		return err
	}

	fe_host := C.GetProtocol() + C.GetAPPDomain()
	link := fmt.Sprintf("%s/activate?token=%s", fe_host, authToken)

	log.WithField("link", link).Debug("Activation Link")

	sub, text, html := U.CreateActivationTemplate(link)
	err = C.GetServices().Mailer.SendMail(agent.Email, C.GetFactorsSenderEmail(), sub, html, text)

	if err != nil {
		log.WithError(err).Error("Failed to send activation email")
	}
	return err
}
