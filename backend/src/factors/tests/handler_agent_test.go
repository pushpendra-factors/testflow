package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	V1 "factors/handler/v1"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendSignInRequest(email, password string, r *gin.Engine) *httptest.ResponseRecorder {

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, "/agents/signin").
		WithHeader("Content-UnitType", "application/json").
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

	t.Run("SiginEmailCheck", func(t *testing.T) {
		emailString := U.RandomLowerAphaNumString(6)
		wrongEmail := []string{
			emailString + "@gmail.com",      // Contains blocked domain
			emailString + "@flowminer.com",  // Contains disposable domain and blocked email list
			emailString + "@@@random.local", // Doesn't conform to acceptable email address structure
			emailString + "  @random.local", // Doesn't conform to acceptable email address structure
		}

		idxEmail := U.RandomIntInRange(0, len(wrongEmail))
		email := wrongEmail[idxEmail]
		pass := U.RandomLowerAphaNumString(6)

		w := sendSignInRequest(email, pass, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("SigninSuccess", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := SetupAgentReturnDAO(email, "+93214356")
		assert.Equal(t, http.StatusCreated, errCode)

		plainTextPassword := U.RandomLowerAphaNumString(6)
		errCode = store.GetStore().UpdateAgentPassword(agent.UUID, plainTextPassword, time.Now().UTC())
		assert.Equal(t, http.StatusAccepted, errCode)

		w := sendSignInRequest(email, plainTextPassword, r)
		assert.Equal(t, http.StatusOK, w.Code)

		cookies := w.Result().Cookies()
		assert.True(t, len(cookies) > 0)
	})

	t.Run("FailedLoginAttempt", func(t *testing.T) {
		email := getRandomEmail()
		_, errCode := SetupAgentReturnDAO(email, "+93214356")
		assert.Equal(t, http.StatusCreated, errCode)

		wrongPassword := "wrong@123"

		for i := 1; i < 11; i++ {
			w := sendSignInRequest(email, wrongPassword, r)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		}
		// last failure attempt
		w := sendSignInRequest(email, wrongPassword, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PartialFailedLoginAttempt", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := SetupAgentReturnDAO(email, "+93214356")
		assert.Equal(t, http.StatusCreated, errCode)

		wrongPassword := "wrong@123"

		for i := 1; i < 5; i++ {
			w := sendSignInRequest(email, wrongPassword, r)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		}
		plainTextPassword := U.RandomLowerAphaNumString(6)
		errCode = store.GetStore().UpdateAgentPassword(agent.UUID, plainTextPassword, time.Now().UTC())
		assert.Equal(t, http.StatusAccepted, errCode)

		w := sendSignInRequest(email, plainTextPassword, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

}

func TestAPIAgentSignout(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	w := httptest.NewRecorder()
	_, agent, _ := SetupProjectWithAgentDAO()
	authData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, time.Second*1000)

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, "/agents/signout").
		WithHeader("Content-UnitType", "application/json").
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  authData,
			MaxAge: 1000,
		})

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

func sendAgentInviteRequest(email string, role int64, projectId int64,
	authData string, exp int, r *gin.Engine) *httptest.ResponseRecorder {

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/agents/invite", projectId)).
		WithHeader("Content-UnitType", "application/json").
		WithPostParams(map[string]interface{}{
			"email": email,
			"role":  role,
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
		projectId := U.RandomInt64()
		emptyAuthData := ""
		w := sendAgentInviteRequest(emailToAdd, model.AGENT, projectId, emptyAuthData, 100, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InviteAgentMissingCookieData", func(t *testing.T) {
		emailToAdd := getRandomEmail()
		projectId := U.RandomInt64()
		emptyAuthData := ""
		w := sendAgentInviteRequest(emailToAdd, model.AGENT, projectId, emptyAuthData, 100, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InviteAgentLoggedInAgentDoesNotExist", func(t *testing.T) {
		emailToAdd := getRandomEmail()
		projectId := U.RandomInt64()
		randomAgentUUID := "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
		randomAgentEmail := getRandomEmail()
		key := U.RandomString(model.AgentSaltLength)
		authData, err := helpers.GetAuthData(randomAgentEmail, randomAgentUUID, key, time.Second*1000)
		assert.Nil(t, err)
		w := sendAgentInviteRequest(emailToAdd, model.AGENT, projectId, authData, 100, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InviteAgentLoggedInExpiredAuthData", func(t *testing.T) {
		emailToAdd := getRandomEmail()
		projectId := U.RandomInt64()
		randomAgentUUID := "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
		randomAgentEmail := getRandomEmail()
		key := U.RandomString(model.AgentSaltLength)
		authData, err := helpers.GetAuthData(randomAgentEmail, randomAgentUUID, key, time.Second*1)
		assert.Nil(t, err)
		time.Sleep(time.Second * 2)
		w := sendAgentInviteRequest(emailToAdd, model.AGENT, projectId, authData, 100, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InviteAgentSuccess", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)
		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+3423647568")
		assert.Equal(t, http.StatusCreated, errCode)

		_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: agent.UUID,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		emailToAdd := getRandomEmail()
		authData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, time.Second*1000)
		assert.Nil(t, err)
		w := sendAgentInviteRequest(emailToAdd, model.AGENT, project.ID, authData, 100, r)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("InviteAgentSuccessWithRoleId", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)
		projectAgentMappings, _ := store.GetStore().GetProjectAgentMappingsByProjectId(project.ID)
		billingAccount, _ := store.GetStore().GetBillingAccountByProjectID(project.ID)
		store.GetStore().UpdateBillingAccount(billingAccount.ID, model.StartupPlanID, "1", "2", "3", "4")
		agentUUIDs := make([]string, 0, 0)
		for _, pam := range projectAgentMappings {
			agentUUIDs = append(agentUUIDs, pam.AgentUUID)
		}

		agents, _ := store.GetStore().GetAgentsByUUIDs(agentUUIDs)

		authData, err := helpers.GetAuthData(agents[0].Email, agents[0].UUID, agents[0].Salt, time.Second*1000)
		assert.Nil(t, err)
		emailToAdd1 := getRandomEmail()
		w := sendAgentInviteRequest(emailToAdd1, model.ADMIN, project.ID, authData, 100, r)
		assert.Equal(t, http.StatusCreated, w.Code)
		emailToAdd2 := getRandomEmail()
		w = sendAgentInviteRequest(emailToAdd2, model.AGENT, project.ID, authData, 100, r)
		assert.Equal(t, http.StatusCreated, w.Code)
		emailToAdd3 := getRandomEmail()
		w = sendAgentInviteRequest(emailToAdd3, 3, project.ID, authData, 100, r)
		assert.Equal(t, http.StatusCreated, w.Code)
		projectAgentMappings, _ = store.GetStore().GetProjectAgentMappingsByProjectId(project.ID)
		agentUUIDs = make([]string, 0, 0)
		for _, pam := range projectAgentMappings {
			agentUUIDs = append(agentUUIDs, pam.AgentUUID)
		}
		agents, _ = store.GetStore().GetAgentsByUUIDs(agentUUIDs)
		for _, agent := range agents {
			var agentUUId string
			if agent.Email == emailToAdd1 {
				agentUUId = agent.UUID
				for _, projectMapping := range projectAgentMappings {
					if projectMapping.AgentUUID == agentUUId {
						assert.Equal(t, projectMapping.Role, uint64(2))
					}
				}
			}
			if agent.Email == emailToAdd2 {
				agentUUId = agent.UUID
				for _, projectMapping := range projectAgentMappings {
					if projectMapping.AgentUUID == agentUUId {
						assert.Equal(t, projectMapping.Role, uint64(1))
					}
				}
			}
			if agent.Email == emailToAdd3 {
				agentUUId = agent.UUID
				for _, projectMapping := range projectAgentMappings {
					if projectMapping.AgentUUID == agentUUId {
						assert.Equal(t, projectMapping.Role, uint64(1))
					}
				}
			}
		}
	})

	t.Run("InviteAgentLimitExceeded", func(t *testing.T) {
		td, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)

		for i := 0; i < 2; i++ {
			agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+3423647568")
			assert.Equal(t, http.StatusCreated, errCode)

			_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
				ProjectID: td.Project.ID,
				AgentUUID: agent.UUID,
			})
			assert.Equal(t, http.StatusCreated, errCode)
		}
		emailToAdd := getRandomEmail()
		authData, err := helpers.GetAuthData(td.Agent.Email, td.Agent.UUID, td.Agent.Salt, time.Second*1000)
		assert.Nil(t, err)
		w := sendAgentInviteRequest(emailToAdd, model.AGENT, td.Project.ID, authData, 100, r)
		// Agent limit increased for all plans to 10k temporarily.
		assert.Equal(t, http.StatusCreated, w.Code)
	})

}

func sendProjectAgentRemoveRequest(r *gin.Engine, projectId int64, agentToRemoveUUID string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/agents/remove", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		}).WithPostParams(map[string]string{
		"agent_uuid": agentToRemoveUUID,
	})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating getAgentBillingAccount Req")
	}
	r.ServeHTTP(w, req)
	return w
}

func sendProjectAgentEditRequest(r *gin.Engine, projectId int64, agentToRemoveUUID string, roleId int64, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/agents/update", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		}).WithPostParams(map[string]interface{}{
		"agent_uuid": agentToRemoveUUID,
		"role":       roleId,
	})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating getAgentBillingAccount Req")
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAPIRemoveAgentFromProject(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("FailAdminTryingToRemoveHimself", func(t *testing.T) {
		testData, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)

		project := testData.Project

		w := sendProjectAgentRemoveRequest(r, project.ID, testData.Agent.UUID, testData.Agent)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("SuccessAdminRemovingUser", func(t *testing.T) {
		testData, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)

		project := testData.Project
		agentToRemove, errCode := SetupAgentReturnDAO(getRandomEmail(), "+3423647568")
		assert.Equal(t, http.StatusCreated, errCode)

		_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: agentToRemove.UUID,
			Role:      model.AGENT,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		w := sendProjectAgentRemoveRequest(r, project.ID, agentToRemove.UUID, testData.Agent)
		assert.Equal(t, http.StatusAccepted, w.Code)
	})

	t.Run("SuccessAdminRemovingAdminButNotHimself", func(t *testing.T) {
		testData, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)

		project := testData.Project
		agentToRemove, errCode := SetupAgentReturnDAO(getRandomEmail(), "+3423647568")
		assert.Equal(t, http.StatusCreated, errCode)

		_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: agentToRemove.UUID,
			Role:      model.ADMIN,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		w := sendProjectAgentRemoveRequest(r, project.ID, agentToRemove.UUID, testData.Agent)
		assert.Equal(t, http.StatusAccepted, w.Code)
	})
	t.Run("SuccessUserRemovingHimself", func(t *testing.T) {
		testData, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)

		project := testData.Project
		agentToRemove, errCode := SetupAgentReturnDAO(getRandomEmail(), "+3423647568")
		assert.Equal(t, http.StatusCreated, errCode)

		_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: agentToRemove.UUID,
			Role:      model.AGENT,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		w := sendProjectAgentRemoveRequest(r, project.ID, agentToRemove.UUID, agentToRemove)
		assert.Equal(t, http.StatusAccepted, w.Code)
	})
	t.Run("SuccessUserRemovingOthers", func(t *testing.T) {
		testData, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)

		project := testData.Project
		agentToRemove, errCode := SetupAgentReturnDAO(getRandomEmail(), "+3423647568")
		assert.Equal(t, http.StatusCreated, errCode)

		_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: agentToRemove.UUID,
			Role:      model.AGENT,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		w := sendProjectAgentRemoveRequest(r, project.ID, testData.Agent.UUID, agentToRemove)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("SuccessAddedAdminRemovingBaseAdmin", func(t *testing.T) {
		testData, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)

		project := testData.Project
		agentToRemove, errCode := SetupAgentReturnDAO(getRandomEmail(), "+3423647568")
		assert.Equal(t, http.StatusCreated, errCode)

		_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: agentToRemove.UUID,
			Role:      model.ADMIN,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		w := sendProjectAgentRemoveRequest(r, project.ID, testData.Agent.UUID, agentToRemove)
		assert.Equal(t, http.StatusAccepted, w.Code)
	})
}

func TestAPIUpdateAgentInProject(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("AdminEditingAdmin", func(t *testing.T) {
		testData, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)

		project := testData.Project
		agentToEdit, errCode := SetupAgentReturnDAO(getRandomEmail(), "+3423647568")
		assert.Equal(t, http.StatusCreated, errCode)

		_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: agentToEdit.UUID,
			Role:      model.ADMIN,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		w := sendProjectAgentEditRequest(r, project.ID, agentToEdit.UUID, 1, testData.Agent)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
	t.Run("AdminEditingUser", func(t *testing.T) {
		testData, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)

		project := testData.Project
		agentToEdit, errCode := SetupAgentReturnDAO(getRandomEmail(), "+3423647568")
		assert.Equal(t, http.StatusCreated, errCode)

		_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: agentToEdit.UUID,
			Role:      model.AGENT,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		w := sendProjectAgentEditRequest(r, project.ID, agentToEdit.UUID, 2, testData.Agent)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
	t.Run("UserEditingAdmin", func(t *testing.T) {
		testData, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)

		project := testData.Project
		agentToEdit, errCode := SetupAgentReturnDAO(getRandomEmail(), "+3423647568")
		assert.Equal(t, http.StatusCreated, errCode)

		_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: agentToEdit.UUID,
			Role:      model.AGENT,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		w := sendProjectAgentEditRequest(r, project.ID, testData.Agent.UUID, 1, agentToEdit)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
func sendAgentVerifyRequest(r *gin.Engine, authData, password, firstName, lastName string) *httptest.ResponseRecorder {

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, "/agents/activate").
		WithHeader("Content-UnitType", "application/json").
		WithPostParams(map[string]interface{}{
			"first_name": firstName,
			"last_name":  lastName,
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
		agent, errCode := SetupAgentReturnDAO(email, "+2345634367")
		assert.Equal(t, http.StatusCreated, errCode)

		firstName := U.RandomLowerAphaNumString(8)
		lastName := U.RandomLowerAphaNumString(8)
		password := U.RandomLowerAphaNumString(8)

		authData, err := helpers.GetAuthData(email, agent.UUID, agent.Salt, helpers.SecondsInFifteenDays*time.Second)
		assert.Nil(t, err)

		w := sendAgentVerifyRequest(r, authData, "Test123@!", firstName, lastName)
		assert.Equal(t, http.StatusOK, w.Code)

		// on retrying
		w = sendAgentVerifyRequest(r, authData, password, firstName, lastName)
		//assert.Equal(t, http.StatusIMUsed, w.Code)
	})
	t.Run("Invalid name", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := SetupAgentReturnDAO(email, "+2345634367")
		assert.Equal(t, http.StatusCreated, errCode)

		firstName := U.RandomLowerAphaNumString(8)
		lastName := "testt%%$$"
		password := U.RandomLowerAphaNumString(8)

		authData, err := helpers.GetAuthData(email, agent.UUID, agent.Salt, helpers.SecondsInFifteenDays*time.Second)
		assert.Nil(t, err)

		w := sendAgentVerifyRequest(r, authData, password, firstName, lastName)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("Invalid name", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := SetupAgentReturnDAO(email, "+2345634367")
		assert.Equal(t, http.StatusCreated, errCode)

		firstName := "test !!"
		lastName := U.RandomLowerAphaNumString(8)

		authData, err := helpers.GetAuthData(email, agent.UUID, agent.Salt, helpers.SecondsInFifteenDays*time.Second)
		assert.Nil(t, err)

		w := sendAgentVerifyRequest(r, authData, "Test123@!", firstName, lastName)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func sendAgentResetPasswordEmailReq(r *gin.Engine, email string) *httptest.ResponseRecorder {

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, "/agents/forgotpassword").
		WithHeader("Content-UnitType", "application/json").
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
		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+3423647568")
		assert.Equal(t, http.StatusCreated, errCode)

		w := sendAgentResetPasswordEmailReq(r, agent.Email)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func sendAgentSetPasswordRequest(r *gin.Engine, authData, password string) *httptest.ResponseRecorder {

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, "/agents/setpassword").
		WithHeader("Content-UnitType", "application/json").
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

	t.Run("MissingPassword", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := SetupAgentReturnDAO(email, "+32478243")
		assert.Equal(t, http.StatusCreated, errCode)

		authData, err := helpers.GetAuthData(email, agent.UUID, agent.Salt, helpers.SecondsInFifteenDays*time.Second)
		assert.Nil(t, err)

		w := sendAgentSetPasswordRequest(r, authData, "")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PasswordCheck", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := SetupAgentReturnDAO(email, "+32478243")
		assert.Equal(t, http.StatusCreated, errCode)

		var wrongPassword = []string{
			"Qwert0#",  // Must have 8 characters
			"qwerty0#", // Must have one upper-case character
			"QWERTY0#", // Must have one lower-case character
			"Qwerty##", // Must have one numerical character
			"Qwerty00", // Must have one special character
		}

		idxPassword := U.RandomIntInRange(0, len(wrongPassword))
		password := wrongPassword[idxPassword]

		authData, err := helpers.GetAuthData(email, agent.UUID, agent.Salt, helpers.SecondsInFifteenDays*time.Second)
		assert.Nil(t, err)

		w := sendAgentSetPasswordRequest(r, authData, password)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := SetupAgentReturnDAO(email, "+224443")
		assert.Equal(t, http.StatusCreated, errCode)

		password := U.RandomLowerAphaNumString(6) + "A@"

		authData, err := helpers.GetAuthData(email, agent.UUID, agent.Salt, helpers.SecondsInFifteenDays*time.Second)
		assert.Nil(t, err)

		w := sendAgentSetPasswordRequest(r, authData, password)
		assert.Equal(t, http.StatusOK, w.Code)

		// on retrying should return unauthorised
		w = sendAgentSetPasswordRequest(r, authData, password)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func sendGetProjectAgentsRequest(r *gin.Engine, projectId int64, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/agents", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating getProjectSetting Req")
	}
	r.ServeHTTP(w, req)
	return w
}

func sendGetProjectAgentsV1Request(r *gin.Engine, projectId int64, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/agents", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating getProjectSetting Req")
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAPIGetProjectAgentsHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("Success", func(t *testing.T) {
		td, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)
		agent := td.Agent
		currentTime := time.Now()
		store.GetStore().UpdateAgentLastLoginInfo(agent.UUID, currentTime)
		w := sendGetProjectAgentsRequest(r, td.Project.ID, agent)
		assert.Equal(t, http.StatusOK, w.Code)

		type Resp struct {
			Agents               map[string]*model.AgentInfo `json:"agents"`
			ProjectAgentMappings []model.ProjectAgentMapping `json:"project_agent_mappings"`
		}

		resp := Resp{}
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		json.Unmarshal(jsonResponse, &resp)

		assert.Equal(t, agent.Email, resp.Agents[agent.UUID].Email)
		assert.Equal(t, currentTime.Unix(), resp.Agents[agent.UUID].LastLoggedIn.Unix())
		assert.Equal(t, 1, len(resp.ProjectAgentMappings))
		assert.Equal(t, agent.UUID, resp.ProjectAgentMappings[0].AgentUUID)
	})
}

func sendGetAgentBillingAccountRequest(r *gin.Engine, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, "/agents/billing").
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating getAgentBillingAccount Req")
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAPIGetAgentBillingAccount(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("Success", func(t *testing.T) {
		td, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)
		agent := td.Agent
		expBA := td.BillingAccount
		w := sendGetAgentBillingAccountRequest(r, agent)
		assert.Equal(t, http.StatusOK, w.Code)

		type Resp struct {
			BillingAcc model.BillingAccount `json:"billing_account"`
		}

		resp := Resp{}
		jsonResponse, err := ioutil.ReadAll(w.Body)
		assert.Nil(t, err)
		json.Unmarshal(jsonResponse, &resp)

		assert.NotNil(t, resp.BillingAcc)
		assert.Equal(t, expBA.AgentUUID, resp.BillingAcc.AgentUUID)
		assert.Equal(t, expBA.ID, resp.BillingAcc.ID)
	})
}

func sendUpdateAgentBillingAccountRequest(r *gin.Engine, orgName, billingAddr, pincode, phoneNo, planCode string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, "/agents/billing").
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		}).WithPostParams(map[string]string{
		"organization_name": orgName,
		"pincode":           pincode,
		"phone_no":          phoneNo,
		"billing_address":   billingAddr,
		"plan_code":         planCode,
	})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating getAgentBillingAccount Req")
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAPIUpdateAgentBillingAccount(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("Success", func(t *testing.T) {
		td, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)
		agent := td.Agent
		expBA := td.BillingAccount

		orgName := "test org " + U.RandomString(4)
		addr := "test addr " + U.RandomString(4)
		pincode := "600322"
		phoneNo := "1232452"
		w := sendUpdateAgentBillingAccountRequest(r, orgName, addr, pincode, phoneNo, model.StartupPlanCode, agent)
		assert.Equal(t, http.StatusOK, w.Code)

		type Resp struct {
			BillingAcc model.BillingAccount `json:"billing_account"`
		}

		resp := Resp{}
		jsonResponse, err := ioutil.ReadAll(w.Body)
		assert.Nil(t, err)
		json.Unmarshal(jsonResponse, &resp)

		assert.NotNil(t, resp.BillingAcc)
		assert.Equal(t, expBA.AgentUUID, resp.BillingAcc.AgentUUID)
		assert.Equal(t, expBA.ID, resp.BillingAcc.ID)
		assert.Equal(t, orgName, resp.BillingAcc.OrganizationName)
		assert.Equal(t, pincode, resp.BillingAcc.Pincode)
		assert.Equal(t, addr, resp.BillingAcc.BillingAddress)
		assert.Equal(t, phoneNo, resp.BillingAcc.PhoneNo)
	})
}

func TestAPIGetProjectAgentsV1Handler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("Success", func(t *testing.T) {
		td, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)
		agent := td.Agent
		currentTime := time.Now()
		store.GetStore().UpdateAgentLastLoginInfo(agent.UUID, currentTime)
		w := sendGetProjectAgentsV1Request(r, td.Project.ID, agent)
		assert.Equal(t, http.StatusOK, w.Code)

		resp := make([]V1.AgentInfoWithProjectMapping, 0)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		json.Unmarshal(jsonResponse, &resp)

		assert.Equal(t, agent.Email, resp[0].Email)
		assert.Equal(t, currentTime.Unix(), resp[0].LastLoggedIn.Unix())
		assert.Equal(t, 1, len(resp))
		assert.Equal(t, agent.UUID, resp[0].UUID)
	})
}
