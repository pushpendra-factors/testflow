package tests

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgentDBCreateAgent(t *testing.T) {
	email := getRandomEmail()
	t.Run("CreateAgent", func(t *testing.T) {
		_, errCode := SetupAgentReturnDAO(email, "+13425354765")
		assert.Equal(t, http.StatusCreated, errCode)
	})

	t.Run("CreateAgentDuplicateEmail", func(t *testing.T) {
		_, errCode := SetupAgentReturnDAO(email, "+13425354765")
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("CreateAgentDuplicateEmailUppercase", func(t *testing.T) {
		_, errCode := SetupAgentReturnDAO(strings.ToUpper(email), "+1356576")
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
}

func TestAgentDBGetAgentByEmail(t *testing.T) {
	email := getRandomEmail()
	t.Run("GetAgentByEmailNotFound", func(t *testing.T) {
		_, errCode := store.GetStore().GetAgentByEmail(email)
		assert.Equal(t, http.StatusNotFound, errCode)
	})
	phone := "+12322365"
	start := time.Now()

	// Create agent
	_, errCode := SetupAgentReturnDAO(email, phone)
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("GetAgentByEmailFound", func(t *testing.T) {
		resultAgent, errCode := store.GetStore().GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, errCode)
		// assert.Equal(t, agent, resultAgent)
		assert.Equal(t, email, resultAgent.Email)
		assert.True(t, resultAgent.CreatedAt.After(start))
		assert.True(t, resultAgent.UpdatedAt.After(start))
		assert.True(t, resultAgent.Salt != "")
	})

	t.Run("GetAgentByUpperCaseEmailFound", func(t *testing.T) {
		resultAgent, errCode := store.GetStore().GetAgentByEmail(strings.ToUpper(email))
		assert.Equal(t, http.StatusFound, errCode)
		// assert.Equal(t, agent, resultAgent)
		assert.Equal(t, email, resultAgent.Email)
		assert.True(t, resultAgent.CreatedAt.After(start))
		assert.True(t, resultAgent.UpdatedAt.After(start))
		assert.True(t, resultAgent.Salt != "")
	})

}

func TestAgentDBGetAgentByUUID(t *testing.T) {

	// t.Run("GetAgentByUUIDAgentNotFound", func(t *testing.T) {
	// TODO: Create a method to generate random uuid of v4 format
	// possibly use library
	// })

	t.Run("GetAgentByUUIDAgentFound", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := SetupAgentReturnDAO(email, "+13425354765")
		assert.Equal(t, http.StatusCreated, errCode)

		uuid := agent.UUID

		retAgent, errCode := store.GetStore().GetAgentByUUID(uuid)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, agent.UUID, retAgent.UUID)
		assert.Equal(t, agent.Email, retAgent.Email)
	})
}

func TestAgentDBGetAgentsByUUIDs(t *testing.T) {
	t.Run("GetAgentsByUUIDsAgentsFound", func(t *testing.T) {
		noOfAgentsToCreate := 3
		expEmails := make([]string, 0, 0)
		expUUIDs := make([]string, 0, 0)
		for i := 0; i < noOfAgentsToCreate; i++ {
			email := getRandomEmail()
			expEmails = append(expEmails, email)
			agent, errCode := SetupAgentReturnDAO(email, "+13425354765")
			assert.Equal(t, http.StatusCreated, errCode)
			expUUIDs = append(expUUIDs, agent.UUID)
		}

		agents, errCode := store.GetStore().GetAgentsByUUIDs(expUUIDs)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, noOfAgentsToCreate, len(agents))

		resultEmails := make([]string, 0, 0)
		for _, agent := range agents {
			resultEmails = append(resultEmails, agent.Email)
		}

		sort.Strings(expEmails)
		sort.Strings(resultEmails)

		assert.Equal(t, expEmails, resultEmails)
	})
}

func TestAgentHashPasswordAndComparePassword(t *testing.T) {
	plainTextPassword := U.RandomString(10)
	hashedPassword, err := model.HashPassword(plainTextPassword)
	assert.Nil(t, err)

	equal := model.IsPasswordAndHashEqual(plainTextPassword, hashedPassword)
	assert.True(t, equal)

	wrongPlainTextPass := plainTextPassword + U.RandomString(4)
	notEqual := model.IsPasswordAndHashEqual(wrongPlainTextPass, hashedPassword)
	assert.False(t, notEqual)
}

func TestAgentDBUpdatePassword(t *testing.T) {
	t.Run("UpdatePasswordAgentNotPresent", func(t *testing.T) {
		uuid := getRandomAgentUUID()
		randPlainTextPassword := U.RandomLowerAphaNumString(8)
		ts := time.Now().UTC()

		errCode := store.GetStore().UpdateAgentPassword(uuid, randPlainTextPassword, ts)
		assert.Equal(t, http.StatusNoContent, errCode)
	})
	t.Run("UpdatePasswordSuccess", func(t *testing.T) {
		start := time.Now().UTC()
		email := getRandomEmail()
		agent, errCode := SetupAgentReturnDAO(email, "+13425354765")
		assert.Equal(t, http.StatusCreated, errCode)
		assert.True(t, agent.Password == "")

		randPlainTextPassword := U.RandomLowerAphaNumString(8)
		ts := time.Now().UTC()

		errCode = store.GetStore().UpdateAgentPassword(agent.UUID, randPlainTextPassword, ts)
		assert.Equal(t, http.StatusAccepted, errCode)

		retAgent, errCode := store.GetStore().GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, errCode)

		assert.NotEqual(t, retAgent.Salt, agent.Salt)

		passEqual := model.IsPasswordAndHashEqual(randPlainTextPassword, retAgent.Password)
		assert.True(t, passEqual)
		assert.True(t, (*retAgent.PasswordCreatedAt).After(start))
	})
}

func TestDBAgentUpdateAgentLastLoginInfo(t *testing.T) {
	t.Run("UpdatePasswordAgentLastLoginInfoMissingAgent", func(t *testing.T) {
		ts := time.Now().UTC()
		errCode := store.GetStore().UpdateAgentLastLoginInfo(getRandomAgentUUID(), ts)
		assert.Equal(t, http.StatusNoContent, errCode)
	})
	t.Run("UpdatePasswordAgentLastLoginInfoSuccess", func(t *testing.T) {
		start := time.Now().UTC()
		email := getRandomEmail()
		agent, errCode := SetupAgentReturnDAO(email, "+13425354765")
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, uint64(0), agent.LoginCount)

		ts := time.Now().UTC()
		errCode = store.GetStore().UpdateAgentLastLoginInfo(agent.UUID, ts)
		assert.Equal(t, http.StatusAccepted, errCode)

		retAgent, errCode := store.GetStore().GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, uint64(1), retAgent.LoginCount)
		assert.True(t, (*retAgent.LastLoggedInAt).After(start))
	})
}

func TestDBAgentUpdateAgentVerificationDetails(t *testing.T) {
	t.Run("MissingAgent", func(t *testing.T) {
		agentUUID := getRandomAgentUUID()
		firstName := U.RandomLowerAphaNumString(8)
		lastName := U.RandomLowerAphaNumString(8)
		password := U.RandomLowerAphaNumString(8)
		ts := time.Now().UTC()
		errCode := store.GetStore().UpdateAgentVerificationDetails(agentUUID, password, firstName, lastName, true, ts)
		assert.Equal(t, http.StatusNoContent, errCode)
	})
	t.Run("Success", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := SetupAgentReturnDAO(email, "+13425354765")
		assert.Equal(t, http.StatusCreated, errCode)
		assert.False(t, agent.IsEmailVerified)
		assert.NotEmpty(t, agent.FirstName)
		assert.NotEmpty(t, agent.LastName)
		assert.Empty(t, agent.Password)
		firstName := U.RandomLowerAphaNumString(8)
		lastName := U.RandomLowerAphaNumString(8)
		password := U.RandomLowerAphaNumString(8)
		ts := time.Now().UTC()
		errCode = store.GetStore().UpdateAgentVerificationDetails(agent.UUID, password, firstName, lastName, true, ts)
		assert.Equal(t, http.StatusAccepted, errCode)

		updatedAgent, errCode := store.GetStore().GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, true, updatedAgent.IsEmailVerified)
		assert.Equal(t, firstName, updatedAgent.FirstName)
		assert.Equal(t, lastName, updatedAgent.LastName)
		assert.True(t, model.IsPasswordAndHashEqual(password, updatedAgent.Password))
		assert.Empty(t, updatedAgent.IntAdwordsRefreshToken)
	})
}

func TestUpdateAgentIntAdwordsRefreshToken(t *testing.T) {
	email := getRandomEmail()
	agent, errCode := SetupAgentReturnDAO(email, "+13425354765")
	assert.Equal(t, http.StatusCreated, errCode)

	token := U.RandomLowerAphaNumString(10)
	errCode = store.GetStore().UpdateAgentIntAdwordsRefreshToken(agent.UUID, token)
	assert.Equal(t, http.StatusAccepted, errCode)
	updatedAgent, errCode := store.GetStore().GetAgentByEmail(email)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, false, updatedAgent.IsEmailVerified)

	// updating other cols should not affect token.
	ts := time.Now().UTC()
	errCode = store.GetStore().UpdateAgentLastLoginInfo(agent.UUID, ts)
	assert.Equal(t, http.StatusAccepted, errCode)
	updatedAgent, errCode = store.GetStore().GetAgentByEmail(email)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, token, updatedAgent.IntAdwordsRefreshToken)
}

func TestUpdateAgentInformation(t *testing.T) {
	FalseFlag := false
	email := getRandomEmail()
	agent, errCode := SetupAgentReturnDAO(email, "+13425354765")
	assert.Equal(t, http.StatusCreated, errCode)

	store.GetStore().UpdateAgentInformation(agent.UUID, "A", "B", "",&FalseFlag)
	updatedAgent, errCode := store.GetStore().GetAgentByEmail(email)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, "A", updatedAgent.FirstName)
	assert.Equal(t, "B", updatedAgent.LastName)
	assert.Equal(t, "+13425354765", updatedAgent.Phone)

	store.GetStore().UpdateAgentInformation(agent.UUID, "", "", "+13425354567",&FalseFlag)
	updatedAgent, errCode = store.GetStore().GetAgentByEmail(email)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, "A", updatedAgent.FirstName)
	assert.Equal(t, "B", updatedAgent.LastName)
	assert.Equal(t, "+13425354567", updatedAgent.Phone)
}
