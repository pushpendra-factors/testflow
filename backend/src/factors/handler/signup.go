package handler

import (
	"factors/handler/helpers"
	M "factors/model"
	"fmt"
	"net/http"
	"time"

	C "factors/config"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// curl -X POST --data "email=value1" http://localhost:8080/accounts/signup
func SignUp(c *gin.Context) {

	type signupParams struct {
		Email string `json:"email" binding:"required"`
	}
	params := signupParams{}
	err := c.BindJSON(&params)
	if err != nil {
		log.WithError(err).Error("Failed to parse SignUpParams")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	email := params.Email

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

	agent, code := M.CreateAgent(&M.Agent{Email: email})
	if code == http.StatusInternalServerError {
		log.WithField("email", email).Error("Failed To Create Agent")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
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
	log.WithField("link", link).Debugf("Activation LInk")

	// Create link & Send Agent Activation Email
	log.WithField("email", agent.Email).Info("Sending Agent Activation Email")

	err = C.GetServices().Mailer.SendMail(agent.Email, C.GetFactorsSenderEmail(), "Activate Factors account", link, link)
	if err != nil {
		log.WithError(err).Error("Failed to send activation email")
	}
	return err
}
