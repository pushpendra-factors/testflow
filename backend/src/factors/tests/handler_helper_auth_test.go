package tests

import (
	"factors/handler/helpers"
	U "factors/util"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetAuthData(t *testing.T) {
	email := getRandomEmail()
	agentUUID := U.RandomString(10)
	key := U.RandomString(32)

	t.Run("Success", func(t *testing.T) {
		dur := 4 * time.Second
		authStr, err := helpers.GetAuthData(email, agentUUID, key, dur)
		assert.Nil(t, err)
		assert.NotEmpty(t, authStr)
		ad, err := helpers.ParseAuthData(authStr)
		assert.Nil(t, err)
		assert.Equal(t, agentUUID, ad.AgentUUID)
		retEmail, err := helpers.ParseAndDecryptProtectedFields(key, ad.ProtectedFields)
		assert.Equal(t, email, retEmail)
	})

	t.Run("ExpiredData", func(t *testing.T) {
		dur := 2 * time.Second
		autadhStr, err := helpers.GetAuthData(email, agentUUID, key, dur)
		time.Sleep(4 * time.Second)
		ad, err := helpers.ParseAuthData(autadhStr)
		_, err = helpers.ParseAndDecryptProtectedFields(key, ad.ProtectedFields)
		assert.Equal(t, helpers.ErrExpired, err)
	})

	t.Run("DecryptWithTamperedKey", func(t *testing.T) {
		dur := 4 * time.Second
		authStr, err := helpers.GetAuthData(email, agentUUID, key, dur)
		assert.Nil(t, err)
		key = U.RandomString(32)
		ad, err := helpers.ParseAuthData(authStr)
		_, err = helpers.ParseAndDecryptProtectedFields(key, ad.ProtectedFields)
		assert.NotNil(t, err)
	})
}
