package tests

import (
	"encoding/json"
	H "factors/handler"
	"factors/handler/helpers"
	M "factors/model"
	U "factors/util"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendCreateProjectRequest(r *gin.Engine, projectName string, agent *M.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := U.NewRequestBuilder(http.MethodPost, "/projects").
		WithHeader("Content-Type", "application/json").
		WithPostParams(map[string]string{"name": projectName}).
		WithCookie(&http.Cookie{
			Name:   helpers.FactorsSessionCookieName,
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating Signin Req")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w

}

func TestAPICreateProject(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("CreateProject", func(t *testing.T) {
		projectName := "test_project_name"
		agent, err := SetupAgentReturnDAO()
		assert.Nil(t, err)
		w := sendCreateProjectRequest(r, projectName, agent)
		assert.Equal(t, http.StatusCreated, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotEqual(t, 0, jsonResponseMap["id"])
		assert.Equal(t, projectName, jsonResponseMap["name"].(string))
		assert.NotEqual(t, 0, len(jsonResponseMap["token"].(string)))         // Todo: should be removed from response.
		assert.NotEqual(t, 0, len(jsonResponseMap["private_token"].(string))) // Todo: should be removed from response.
		assert.NotNil(t, jsonResponseMap["created_at"].(string))
		assert.NotNil(t, jsonResponseMap["updated_at"].(string))
		assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
		assert.Equal(t, 6, len(jsonResponseMap))
	})
}
