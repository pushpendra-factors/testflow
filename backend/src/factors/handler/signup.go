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

	_, code := M.GetAgentByEmail(email)
	if code == http.StatusInternalServerError {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if code == http.StatusFound {
		c.AbortWithStatus(http.StatusFound)
		return
	}

	agent := &M.Agent{
		Email: email,
	}

	agent, code = M.CreateAgent(agent)
	if code == http.StatusInternalServerError {
		log.WithField("email", email).Error("Failed To Create Agent")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	authToken, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, time.Second*helpers.SecondsInFifteenDays)
	if err != nil {
		log.WithField("email", email).Error("Failed To Create Agent Auth Token")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	fe_host := C.GetProtocol() + C.GetAPPDomain()
	link := fmt.Sprintf("%s/#/verify?token=%s", fe_host, authToken)
	log.WithField("link", link).Debugf("Verification LInk")

	// Create link & Send Agent Activation Email
	log.WithField("email", email).Info("Sending Agent Activation Email")

	err = C.GetServices().Mailer.SendMail(email, "Factors", "Activate Factors account", link, link)
	if err != nil {
		log.WithError(err).Error("Failed to send activation email")
	}

	// Set Cookie with exp 1 day. After that the agent will be forced to set password
	// And probably redirect to default project view
	c.Status(http.StatusCreated)
}
