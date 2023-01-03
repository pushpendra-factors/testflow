package tests

import (
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/store"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSignUp(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("CreateAgentSuccess", func(t *testing.T) {
		email := getRandomEmail()
		phone := "+917"
		w := sendSignUpRequest(email, phone, r)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("CreateAgentMissingEmail", func(t *testing.T) {
		w := sendSignUpRequest("", "", r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("CreateAgentMissingPhone", func(t *testing.T) {
		w := sendSignUpRequest(getRandomEmail(), "", r)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("CreateAgentDuplicateEmail", func(t *testing.T) {
		email := getRandomEmail()
		phone := "+912253467"
		w := sendSignUpRequest(email, phone, r)
		assert.Equal(t, http.StatusCreated, w.Code)

		w = sendSignUpRequest(email, phone, r)
		assert.Equal(t, http.StatusFound, w.Code)
	})

	t.Run("CreateAgentWithAdditionalDetails", func(t *testing.T) {
		email := getRandomEmail()
		phone := "+912253467"
		w := sendSignUpRequestWithAdditionalDetails(email, phone, r)
		assert.Equal(t, http.StatusCreated, w.Code)
		agent, code := store.GetStore().GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, code)
		assert.Equal(t, agent.FirstName, "first_name")
		assert.Equal(t, agent.LastName, "last_name")
		assert.Equal(t, agent.CompanyURL, "app.factors.ai")
		assert.Equal(t, agent.SubscribeNewsletter, true)
		authData, err := helpers.GetAuthData(email, agent.UUID, agent.Salt, helpers.SecondsInFifteenDays*time.Second)
		assert.Nil(t, err)

		w = sendAgentVerifyRequest(r, authData, "Test123@!", "", "")
		assert.Equal(t, http.StatusOK, w.Code)
		agent, code = store.GetStore().GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, code)
		assert.Equal(t, agent.FirstName, "first_name")
		assert.Equal(t, agent.LastName, "last_name")
		assert.Equal(t, agent.CompanyURL, "app.factors.ai")
		assert.Equal(t, agent.SubscribeNewsletter, true)
	})

	t.Run("CreateAgentWithAdditionalDetailsAndEditDuringActivate", func(t *testing.T) {
		email := getRandomEmail()
		phone := "+912253467"
		w := sendSignUpRequestWithAdditionalDetails(email, phone, r)
		assert.Equal(t, http.StatusCreated, w.Code)
		agent, code := store.GetStore().GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, code)
		assert.Equal(t, agent.FirstName, "first_name")
		assert.Equal(t, agent.LastName, "last_name")
		assert.Equal(t, agent.CompanyURL, "app.factors.ai")
		assert.Equal(t, agent.SubscribeNewsletter, true)
		authData, err := helpers.GetAuthData(email, agent.UUID, agent.Salt, helpers.SecondsInFifteenDays*time.Second)
		assert.Nil(t, err)

		w = sendAgentVerifyRequest(r, authData, "Test123@!", "first_name_1", "last_name_1")
		assert.Equal(t, http.StatusOK, w.Code)
		agent, code = store.GetStore().GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, code)
		assert.Equal(t, agent.FirstName, "first_name_1")
		assert.Equal(t, agent.LastName, "last_name_1")
		assert.Equal(t, agent.CompanyURL, "app.factors.ai")
		assert.Equal(t, agent.SubscribeNewsletter, true)
	})
}

func sendSignUpRequest(email string, phone string, r *gin.Engine) *httptest.ResponseRecorder {
	params := map[string]string{"email": email, "phone": phone}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, "/accounts/signup").
		WithHeader("Content-UnitType", "application/json").
		WithPostParams(params)

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating Signin Req")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendSignUpRequestWithAdditionalDetails(email string, phone string, r *gin.Engine) *httptest.ResponseRecorder {
	params := map[string]interface{}{"email": email, "phone": phone, "first_name": "first_name", "last_name": "last_name", "company_url": "app.factors.ai", "subscribe_newsletter": true}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, "/accounts/signup").
		WithHeader("Content-UnitType", "application/json").
		WithPostParams(params)

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating Signin Req")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
