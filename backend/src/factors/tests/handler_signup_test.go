package tests

import (
	"bytes"
	"encoding/json"
	H "factors/handler"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSignUp(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("CreateAgentSuccess", func(t *testing.T) {
		email := getRandomEmail()
		w := sendSignUpRequest(email, r)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("CreateAgentMissingEmail", func(t *testing.T) {
		w := sendSignUpRequest("", r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateAgentDuplicateEmail", func(t *testing.T) {
		email := getRandomEmail()
		w := sendSignUpRequest(email, r)
		assert.Equal(t, http.StatusCreated, w.Code)

		w = sendSignUpRequest(email, r)
		assert.Equal(t, http.StatusFound, w.Code)
	})

}

func sendSignUpRequest(email string, r *gin.Engine) *httptest.ResponseRecorder {
	params := map[string]string{"email": email}
	jsonValue, err := json.Marshal(params)
	if err != nil {
		log.WithError(err).Error("Error Creating json params")
	}
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/accounts/signup", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.WithError(err).Error("Error Creating Signup Req")
	}
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}
