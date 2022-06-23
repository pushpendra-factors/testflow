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
		retEmail, _, err := helpers.ParseAndDecryptProtectedFields(key, 0, 0, ad.ProtectedFields, helpers.SecondsInOneMonth)
		assert.Equal(t, email, retEmail)
	})
	t.Run("SuccessWithLogoutInThePast", func(t *testing.T) {
		dur := helpers.SecondsInOneMonth * time.Second
		authStr, err := helpers.GetAuthData(email, agentUUID, key, dur)
		assert.Nil(t, err)
		assert.NotEmpty(t, authStr)
		ad, err := helpers.ParseAuthData(authStr)
		assert.Nil(t, err)
		assert.Equal(t, agentUUID, ad.AgentUUID)
		retEmail, _, err := helpers.ParseAndDecryptProtectedFields(key, time.Now().UTC().Unix()-100, time.Now().UTC().Unix()-100, ad.ProtectedFields, helpers.SecondsInOneMonth)
		assert.Equal(t, email, retEmail)
	})

	var errMsg string
	t.Run("ExpiredData", func(t *testing.T) {
		dur := 2 * time.Second
		autadhStr, err := helpers.GetAuthData(email, agentUUID, key, dur)
		time.Sleep(4 * time.Second)
		ad, err := helpers.ParseAuthData(autadhStr)
		_, errMsg, err = helpers.ParseAndDecryptProtectedFields(key, 0, 0, ad.ProtectedFields, helpers.SecondsInOneMonth)
		assert.Equal(t, errMsg, "ExpiredKey")
		assert.Equal(t, helpers.ErrExpired, err)
	})

	t.Run("DecryptWithTamperedKey", func(t *testing.T) {
		dur := 4 * time.Second
		authStr, err := helpers.GetAuthData(email, agentUUID, key, dur)
		assert.Nil(t, err)
		key = U.RandomString(32)
		ad, err := helpers.ParseAuthData(authStr)
		_, errMsg, err = helpers.ParseAndDecryptProtectedFields(key, 0, 0, ad.ProtectedFields, helpers.SecondsInOneMonth)
		assert.Equal(t, errMsg, "Tampering")
		assert.NotNil(t, err)
	})
	t.Run("LoggedoutUserPastCookieAccess", func(t *testing.T) {
		dur := helpers.SecondsInOneMonth * time.Second
		authStr, err := helpers.GetAuthData(email, agentUUID, key, dur)
		assert.Nil(t, err)
		ad, err := helpers.ParseAuthData(authStr)
		_, errMsg, err = helpers.ParseAndDecryptProtectedFields(key, time.Now().UTC().Add(100*time.Second).Unix(), 0, ad.ProtectedFields, helpers.SecondsInOneMonth)
		assert.Equal(t, errMsg, "CookieInvalid")
		assert.NotNil(t, err)
	})
	t.Run("PasswordChangedUserPastCookieAccesss", func(t *testing.T) {
		dur := helpers.SecondsInOneMonth * time.Second
		authStr, err := helpers.GetAuthData(email, agentUUID, key, dur)
		assert.Nil(t, err)
		ad, err := helpers.ParseAuthData(authStr)
		_, errMsg, err = helpers.ParseAndDecryptProtectedFields(key, 0, time.Now().UTC().Add(100*time.Second).Unix(), ad.ProtectedFields, helpers.SecondsInOneMonth)
		assert.Equal(t, errMsg, "CookieInvalid")
		assert.NotNil(t, err)
	})
}
