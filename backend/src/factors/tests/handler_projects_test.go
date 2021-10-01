package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
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

func sendCreateProjectRequest(r *gin.Engine, projectName string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, "/projects").
		WithHeader("Content-Type", "application/json").
		WithPostParams(map[string]string{"name": projectName, "project_uri": "factors.ai", "time_format": "HH:mm:ss", "date_format": "yyyy-MM-dd", "time_zone": "IST"}).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
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

func sendEditProjectRequest(r *gin.Engine, projectId uint64, projectName string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%v", projectId)).
		WithHeader("Content-Type", "application/json").
		WithPostParams(map[string]string{"name": "edit", "project_uri": "factors.ai.edit", "time_format": "HH:mm:ss.edit", "date_format": "yyyy-MM-dd.edit", "time_zone": "IST.edit"}).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
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
		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+254346477")
		assert.Equal(t, http.StatusCreated, errCode)
		w := sendCreateProjectRequest(r, projectName, agent)
		assert.Equal(t, http.StatusCreated, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		fmt.Println(jsonResponseMap)
		assert.NotEqual(t, 0, jsonResponseMap["id"])
		assert.Equal(t, projectName, jsonResponseMap["name"].(string))
		assert.Equal(t, "factors.ai", jsonResponseMap["project_uri"].(string))
		assert.Equal(t, "HH:mm:ss", jsonResponseMap["time_format"].(string))
		assert.Equal(t, "yyyy-MM-dd", jsonResponseMap["date_format"].(string))
		assert.Equal(t, "Asia/Kolkata", jsonResponseMap["time_zone"].(string))
		assert.NotEqual(t, 0, len(jsonResponseMap["token"].(string)))         // Todo: should be removed from response.
		assert.NotEqual(t, 0, len(jsonResponseMap["private_token"].(string))) // Todo: should be removed from response.
		assert.NotNil(t, jsonResponseMap["created_at"].(string))
		assert.NotNil(t, jsonResponseMap["updated_at"].(string))
		assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
		assert.NotNil(t, jsonResponseMap["jobs_metadata"])
		assert.NotNil(t, jsonResponseMap["interaction_settings"])
		assert.NotNil(t, jsonResponseMap["salesforce_touch_points"])
		assert.NotNil(t, jsonResponseMap["hubspot_touch_points"])
		assert.Equal(t, 16, len(jsonResponseMap))
	})
}

func TestAPIEditProject(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("EditProject", func(t *testing.T) {
		projectName := "test_project_name"
		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+254346477")
		assert.Equal(t, http.StatusCreated, errCode)
		w := sendCreateProjectRequest(r, projectName, agent)
		assert.Equal(t, http.StatusCreated, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotEqual(t, 0, jsonResponseMap["id"])
		assert.Equal(t, projectName, jsonResponseMap["name"].(string))
		assert.Equal(t, "factors.ai", jsonResponseMap["project_uri"].(string))
		assert.Equal(t, "HH:mm:ss", jsonResponseMap["time_format"].(string))
		assert.Equal(t, "yyyy-MM-dd", jsonResponseMap["date_format"].(string))
		assert.Equal(t, "Asia/Kolkata", jsonResponseMap["time_zone"].(string))
		assert.NotEqual(t, 0, len(jsonResponseMap["token"].(string)))         // Todo: should be removed from response.
		assert.NotEqual(t, 0, len(jsonResponseMap["private_token"].(string))) // Todo: should be removed from response.
		assert.NotNil(t, jsonResponseMap["created_at"].(string))
		assert.NotNil(t, jsonResponseMap["updated_at"].(string))
		assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
		assert.NotNil(t, jsonResponseMap["jobs_metadata"])
		assert.NotNil(t, jsonResponseMap["interaction_settings"])
		assert.NotNil(t, jsonResponseMap["salesforce_touch_points"])
		assert.NotNil(t, jsonResponseMap["hubspot_touch_points"])
		assert.Equal(t, 16, len(jsonResponseMap))
		w = sendEditProjectRequest(r, uint64(jsonResponseMap["id"].(float64)), projectName, agent)
		assert.Equal(t, http.StatusCreated, w.Code)
		jsonResponse, _ = ioutil.ReadAll(w.Body)
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotEqual(t, 0, jsonResponseMap["id"])
		assert.Equal(t, "edit", jsonResponseMap["name"].(string))
		assert.Equal(t, "factors.ai.edit", jsonResponseMap["project_uri"].(string))
		assert.Equal(t, "HH:mm:ss.edit", jsonResponseMap["time_format"].(string))
		assert.Equal(t, "yyyy-MM-dd.edit", jsonResponseMap["date_format"].(string))
		assert.Equal(t, "IST.edit", jsonResponseMap["time_zone"].(string))
		assert.NotEqual(t, 0, len(jsonResponseMap["token"].(string)))         // Todo: should be removed from response.
		assert.NotEqual(t, 0, len(jsonResponseMap["private_token"].(string))) // Todo: should be removed from response.
		assert.NotNil(t, jsonResponseMap["created_at"].(string))
		assert.NotNil(t, jsonResponseMap["updated_at"].(string))
		assert.NotNil(t, jsonResponseMap["jobs_metadata"])
		assert.NotNil(t, jsonResponseMap["interaction_settings"])
		assert.NotNil(t, jsonResponseMap["salesforce_touch_points"])
		assert.NotNil(t, jsonResponseMap["hubspot_touch_points"])
		assert.Equal(t, 16, len(jsonResponseMap))
	})
}
