package tests

import (
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendSignInRequest(email, password string, r *gin.Engine) *httptest.ResponseRecorder {

	rb := U.NewRequestBuilder(http.MethodPost, "/agents/signin").
		WithHeader("Content-Type", "application/json").
		WithPostParams(map[string]string{"email": email, "password": password})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating Signin Req")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func TestAPIAgentSignin(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("SigninMissingParams", func(t *testing.T) {
		email := getRandomEmail()
		pass := ""
		w := sendSignInRequest(email, pass, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("SigninMissingAgent", func(t *testing.T) {
		email := getRandomEmail()
		pass := U.RandomLowerAphaNumString(6)
		w := sendSignInRequest(email, pass, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("SigninSuccess", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := M.CreateAgent(&M.Agent{Email: email})
		assert.Equal(t, http.StatusCreated, errCode)

		plainTextPassword := U.RandomLowerAphaNumString(6)
		errCode = M.UpdateAgentPassword(agent.UUID, plainTextPassword, time.Now().UTC())
		assert.Equal(t, http.StatusAccepted, errCode)

		w := sendSignInRequest(email, plainTextPassword, r)
		assert.Equal(t, http.StatusOK, w.Code)

		cookies := w.Result().Cookies()
		assert.True(t, len(cookies) > 0)
	})
}

func TestAPIAgentSignout(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	w := httptest.NewRecorder()

	rb := U.NewRequestBuilder(http.MethodGet, "/agents/signout").
		WithHeader("Content-Type", "application/json")

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating Signout Req")
	}

	r.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	assert.Equal(t, 1, len(cookies))
	cookie := cookies[0]
	assert.Equal(t, C.GetFactorsCookieName(), cookie.Name)
	assert.Equal(t, helpers.ExpireCookie, cookie.MaxAge)
}

func sendAgentInviteRequest(email string, projectId uint64, authData string, exp int, r *gin.Engine) *httptest.ResponseRecorder {

	rb := U.NewRequestBuilder(http.MethodPost, fmt.Sprintf("/projects/%d/agents/invite", projectId)).
		WithHeader("Content-Type", "application/json").
		WithPostParams(map[string]interface{}{
			"email": email,
		}).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  authData,
			MaxAge: exp,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Building Request")
	}
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	return w
}

func TestAPIAgentInvite(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("InviteAgentNotLoggedIn", func(t *testing.T) {
		emailToAdd := getRandomEmail()
		projectId := U.RandomUint64()
		emptyAuthData := ""
		w := sendAgentInviteRequest(emailToAdd, projectId, emptyAuthData, 100, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InviteAgentMissingCookieData", func(t *testing.T) {
		emailToAdd := getRandomEmail()
		projectId := U.RandomUint64()
		emptyAuthData := ""
		w := sendAgentInviteRequest(emailToAdd, projectId, emptyAuthData, 100, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InviteAgentLoggedInAgentDoesNotExist", func(t *testing.T) {
		emailToAdd := getRandomEmail()
		projectId := U.RandomUint64()
		randomAgentUUID := "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
		randomAgentEmail := getRandomEmail()
		key := U.RandomString(M.SALT_LEN)
		authData, err := helpers.GetAuthData(randomAgentEmail, randomAgentUUID, key, time.Second*1000)
		assert.Nil(t, err)
		w := sendAgentInviteRequest(emailToAdd, projectId, authData, 100, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InviteAgentLoggedInExpiredAuthData", func(t *testing.T) {
		emailToAdd := getRandomEmail()
		projectId := U.RandomUint64()
		randomAgentUUID := "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
		randomAgentEmail := getRandomEmail()
		key := U.RandomString(M.SALT_LEN)
		authData, err := helpers.GetAuthData(randomAgentEmail, randomAgentUUID, key, time.Second*1)
		assert.Nil(t, err)
		time.Sleep(time.Second * 2)
		w := sendAgentInviteRequest(emailToAdd, projectId, authData, 100, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InviteAgentSuccess", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)
		agent, errCode := M.CreateAgent(&M.Agent{Email: getRandomEmail()})
		assert.Equal(t, http.StatusCreated, errCode)

		_, errCode = M.CreateProjectAgentMapping(&M.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: agent.UUID,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		emailToAdd := getRandomEmail()
		authData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, time.Second*1000)
		assert.Nil(t, err)
		w := sendAgentInviteRequest(emailToAdd, project.ID, authData, 100, r)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

}

func sendAgentVerifyRequest(r *gin.Engine, authData, password, firstName, lastName string) *httptest.ResponseRecorder {

	rb := U.NewRequestBuilder(http.MethodPost, "/agents/activate").
		WithHeader("Content-Type", "application/json").
		WithPostParams(map[string]interface{}{
			"first_name": firstName,
			"last_name":  firstName,
			"password":   password,
		}).WithQueryParams(map[string]string{
		"token": authData,
	})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Building agent verfication Request")
	}
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	return w
}
func TestAPIAgentVerify(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	t.Run("MissingToken", func(t *testing.T) {
		emptyAuthData := ""
		firstName := U.RandomLowerAphaNumString(8)
		lastName := U.RandomLowerAphaNumString(8)
		password := U.RandomLowerAphaNumString(8)
		w := sendAgentVerifyRequest(r, emptyAuthData, password, firstName, lastName)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("MalformedToken", func(t *testing.T) {
		emptyAuthData := U.RandomLowerAphaNumString(20)
		firstName := U.RandomLowerAphaNumString(8)
		lastName := U.RandomLowerAphaNumString(8)
		password := U.RandomLowerAphaNumString(8)

		w := sendAgentVerifyRequest(r, emptyAuthData, password, firstName, lastName)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := M.CreateAgent(&M.Agent{Email: email})
		assert.Equal(t, http.StatusCreated, errCode)

		firstName := U.RandomLowerAphaNumString(8)
		lastName := U.RandomLowerAphaNumString(8)
		password := U.RandomLowerAphaNumString(8)

		authData, err := helpers.GetAuthData(email, agent.UUID, agent.Salt, helpers.SecondsInFifteenDays*time.Second)
		assert.Nil(t, err)

		w := sendAgentVerifyRequest(r, authData, password, firstName, lastName)
		assert.Equal(t, http.StatusOK, w.Code)

		// on retrying
		w = sendAgentVerifyRequest(r, authData, password, firstName, lastName)
		assert.Equal(t, http.StatusIMUsed, w.Code)
	})
}

func sendAgentResetPasswordEmailReq(r *gin.Engine, email string) *httptest.ResponseRecorder {

	rb := U.NewRequestBuilder(http.MethodPost, "/agents/forgotpassword").
		WithHeader("Content-Type", "application/json").
		WithPostParams(map[string]interface{}{
			"email": email,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Building agent verfication Request")
	}
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	return w
}

func TestAPIAgentGenerateResetPasswordEmail(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	t.Run("MissingParams", func(t *testing.T) {
		w := sendAgentResetPasswordEmailReq(r, "")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("AgentMissing", func(t *testing.T) {
		email := getRandomEmail()
		w := sendAgentResetPasswordEmailReq(r, email)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("AgentExists", func(t *testing.T) {
		agent, err := SetupAgentReturnDAO()
		assert.Nil(t, err)

		w := sendAgentResetPasswordEmailReq(r, agent.Email)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func sendAgentSetPasswordRequest(r *gin.Engine, authData, password string) *httptest.ResponseRecorder {

	rb := U.NewRequestBuilder(http.MethodPost, "/agents/setpassword").
		WithHeader("Content-Type", "application/json").
		WithPostParams(map[string]interface{}{
			"password": password,
		}).WithQueryParams(map[string]string{
		"token": authData,
	})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Building agent set password Request")
	}
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	return w
}
func TestAPIAgentSetPassword(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	t.Run("MissingToken", func(t *testing.T) {
		emptyAuthData := ""
		password := U.RandomLowerAphaNumString(8)
		w := sendAgentSetPasswordRequest(r, emptyAuthData, password)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("MalformedToken", func(t *testing.T) {
		emptyAuthData := U.RandomLowerAphaNumString(20)
		password := U.RandomLowerAphaNumString(8)
		w := sendAgentSetPasswordRequest(r, emptyAuthData, password)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := M.CreateAgent(&M.Agent{Email: email})
		assert.Equal(t, http.StatusCreated, errCode)

		password := U.RandomLowerAphaNumString(8)

		authData, err := helpers.GetAuthData(email, agent.UUID, agent.Salt, helpers.SecondsInFifteenDays*time.Second)
		assert.Nil(t, err)

		w := sendAgentSetPasswordRequest(r, authData, password)
		assert.Equal(t, http.StatusOK, w.Code)

		// on retrying should return unauthorised
		w = sendAgentSetPasswordRequest(r, authData, password)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
