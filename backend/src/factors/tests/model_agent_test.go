package tests

import (
	M "factors/model"
	U "factors/util"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgentDBCreateAgent(t *testing.T) {
	email := getRandomEmail()
	t.Run("CreateAgent", func(t *testing.T) {
		_, errCode := M.CreateAgent(&M.Agent{Email: email})
		assert.Equal(t, http.StatusCreated, errCode)
	})

	t.Run("CreateAgentDuplicateEmail", func(t *testing.T) {
		_, errCode := M.CreateAgent(&M.Agent{Email: email})
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("CreateAgentDuplicateEmailUppercase", func(t *testing.T) {
		_, errCode := M.CreateAgent(&M.Agent{Email: strings.ToUpper(email)})
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
}

func TestAgentDBGetAgentByEmail(t *testing.T) {
	email := getRandomEmail()
	t.Run("GetAgentByEmailNotFound", func(t *testing.T) {
		_, errCode := M.GetAgentByEmail(email)
		assert.Equal(t, http.StatusNotFound, errCode)
	})

	start := time.Now()

	// Create agent
	_, errCode := M.CreateAgent(&M.Agent{Email: email})
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("GetAgentByEmailFound", func(t *testing.T) {
		resultAgent, errCode := M.GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, errCode)
		// assert.Equal(t, agent, resultAgent)
		assert.Equal(t, email, resultAgent.Email)
		assert.True(t, resultAgent.CreatedAt.After(start))
		assert.True(t, resultAgent.UpdatedAt.After(start))
		assert.True(t, resultAgent.Salt != "")
	})

	t.Run("GetAgentByUpperCaseEmailFound", func(t *testing.T) {
		resultAgent, errCode := M.GetAgentByEmail(strings.ToUpper(email))
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
		agent, errCode := M.CreateAgent(&M.Agent{Email: email})
		assert.Equal(t, http.StatusCreated, errCode)

		uuid := agent.UUID

		retAgent, errCode := M.GetAgentByUUID(uuid)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, agent.UUID, retAgent.UUID)
		assert.Equal(t, agent.Email, retAgent.Email)
	})
}

func TestAgentHashPasswordAndComparePassword(t *testing.T) {
	plainTextPassword := U.RandomString(10)
	hashedPassword, err := M.HashPassword(plainTextPassword)
	assert.Nil(t, err)

	equal := M.IsPasswordAndHashEqual(plainTextPassword, hashedPassword)
	assert.True(t, equal)

	wrongPlainTextPass := plainTextPassword + U.RandomString(4)
	notEqual := M.IsPasswordAndHashEqual(wrongPlainTextPass, hashedPassword)
	assert.False(t, notEqual)
}

func TestAgentDBUpdatePassword(t *testing.T) {
	t.Run("UpdatePasswordAgentNotPresent", func(t *testing.T) {
		uuid := getRandomAgentUUID()
		randPlainTextPassword := U.RandomLowerAphaNumString(8)
		ts := time.Now().UTC()

		errCode := M.UpdateAgentPassword(uuid, randPlainTextPassword, ts)
		assert.Equal(t, http.StatusNoContent, errCode)
	})
	t.Run("UpdatePasswordSuccess", func(t *testing.T) {
		start := time.Now().UTC()
		email := getRandomEmail()
		agent, errCode := M.CreateAgent(&M.Agent{Email: email})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.True(t, agent.Password == "")

		randPlainTextPassword := U.RandomLowerAphaNumString(8)
		ts := time.Now().UTC()

		errCode = M.UpdateAgentPassword(agent.UUID, randPlainTextPassword, ts)
		assert.Equal(t, http.StatusAccepted, errCode)

		retAgent, errCode := M.GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, errCode)

		assert.NotEqual(t, retAgent.Salt, agent.Salt)

		passEqual := M.IsPasswordAndHashEqual(randPlainTextPassword, retAgent.Password)
		assert.True(t, passEqual)
		assert.True(t, (*retAgent.PasswordCreatedAt).After(start))
	})
}

func TestDBAgentUpdateAgentLastLoginInfo(t *testing.T) {
	t.Run("UpdatePasswordAgentLastLoginInfoMissingAgent", func(t *testing.T) {
		email := getRandomEmail()
		ts := time.Now().UTC()
		errCode := M.UpdateAgentLastLoginInfo(email, ts)
		assert.Equal(t, http.StatusNoContent, errCode)
	})
	t.Run("UpdatePasswordAgentLastLoginInfoSuccess", func(t *testing.T) {
		start := time.Now().UTC()
		email := getRandomEmail()
		agent, errCode := M.CreateAgent(&M.Agent{Email: email})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, uint64(0), agent.LoginCount)

		ts := time.Now().UTC()
		errCode = M.UpdateAgentLastLoginInfo(email, ts)
		assert.Equal(t, http.StatusAccepted, errCode)

		retAgent, errCode := M.GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, uint64(1), retAgent.LoginCount)
		assert.True(t, (*retAgent.LastLoggedInAt).After(start))
	})

	t.Run("UpdatePasswordAgentLastLoginInfoSuccessUppercaseEmail", func(t *testing.T) {
		start := time.Now().UTC()
		email := getRandomEmail()
		agent, errCode := M.CreateAgent(&M.Agent{Email: email})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, uint64(0), agent.LoginCount)

		ts := time.Now().UTC()
		errCode = M.UpdateAgentLastLoginInfo(strings.ToUpper(email), ts)
		assert.Equal(t, http.StatusAccepted, errCode)

		retAgent, errCode := M.GetAgentByEmail(email)
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
		errCode := M.UpdateAgentVerificationDetails(agentUUID, password, firstName, lastName, true, ts)
		assert.Equal(t, http.StatusNoContent, errCode)
	})
	t.Run("Success", func(t *testing.T) {
		email := getRandomEmail()
		agent, errCode := M.CreateAgent(&M.Agent{Email: email})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.False(t, agent.IsEmailVerified)
		assert.Empty(t, agent.FirstName)
		assert.Empty(t, agent.LastName)
		assert.Empty(t, agent.Password)
		firstName := U.RandomLowerAphaNumString(8)
		lastName := U.RandomLowerAphaNumString(8)
		password := U.RandomLowerAphaNumString(8)
		ts := time.Now().UTC()
		errCode = M.UpdateAgentVerificationDetails(agent.UUID, password, firstName, lastName, true, ts)
		assert.Equal(t, http.StatusAccepted, errCode)

		updatedAgent, errCode := M.GetAgentByEmail(email)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, true, updatedAgent.IsEmailVerified)
		assert.Equal(t, firstName, updatedAgent.FirstName)
		assert.Equal(t, lastName, updatedAgent.LastName)
		assert.True(t, M.IsPasswordAndHashEqual(password, updatedAgent.Password))

	})
}
