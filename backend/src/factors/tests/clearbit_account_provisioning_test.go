package tests

import (
	"encoding/json"
	"factors/integration/clear_bit"
	"factors/model/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvisionClearbitAccounts(t *testing.T) {

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "testID",
			"domain": "example.com",
			"email": "test@example.com",
			"plans": {
				"factors-pbc-customer": [
					{
						"name": "Reveal",
						"count": 0,
						"limit": 10000
					}
				]
			},
			"keys": {
				"public": "pk_test123",
				"secret": "sk_test123"
			}
		}
		`))
	}))
	apiUrl := testServer.URL
	testEmail := "test@example.com"
	testDomain := "example.com"
	res, err := clear_bit.GetClearbitProvisionAccountResponse(apiUrl, testEmail, testDomain, "")
	assert.Nil(t, err)
	assert.NotNil(t, res)

	var response model.ClearbitProvisionAPIResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Error(err)
	}

	assert.NotNil(t, response.Id)
	assert.NotNil(t, response.Domain)
	assert.NotNil(t, response.Email)
	assert.Equal(t, response.Id, "testID")
	assert.Equal(t, response.Email, "test@example.com")
	assert.Equal(t, response.Domain, "example.com")
	assert.Equal(t, response.Keys.Secret, "sk_test123")
	assert.Equal(t, response.Keys.Public, "pk_test123")

}
