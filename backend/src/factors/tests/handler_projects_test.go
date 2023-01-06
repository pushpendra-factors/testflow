package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
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
		WithHeader("Content-UnitType", "application/json").
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

func sendGetProjectsRequest(r *gin.Engine, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, "/v1/projects").
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Getting projects")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w

}

func sendEditProjectRequest(r *gin.Engine, projectId int64, projectName string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%v", projectId)).
		WithHeader("Content-UnitType", "application/json").
		WithPostParams(map[string]string{"name": projectName, "project_uri": "factors.ai.edit", "time_format": "HH:mm:ss.edit", "date_format": "yyyy-MM-dd.edit", "time_zone": "IST.edit"}).
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
		projectName := "Test_Project_Name"
		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+254346477")
		assert.Equal(t, http.StatusCreated, errCode)
		w := sendCreateProjectRequest(r, projectName, agent)
		assert.Equal(t, http.StatusCreated, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotEqual(t, "", jsonResponseMap["id"])
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
		assert.Nil(t, jsonResponseMap["salesforce_touch_points"])
		assert.Nil(t, jsonResponseMap["hubspot_touch_points"])
		assert.Equal(t, 16, len(jsonResponseMap))

		id, _ := strconv.Atoi(jsonResponseMap["id"].(string))
		project, status := store.GetStore().GetProject(int64(id))
		assert.Equal(t, http.StatusFound, status)
		// The project name should match exactly by case.
		assert.Equal(t, projectName, project.Name)

		projectName = "Test_Project_Name!!!"
		agent, errCode = SetupAgentReturnDAO(getRandomEmail(), "+254346477")
		assert.Equal(t, http.StatusCreated, errCode)
		w = sendCreateProjectRequest(r, projectName, agent)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestBlockMaliciousPayloadMiddleware(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	t.Run("CreateProjectPayloadWithScript", func(t *testing.T) {
		projectName := "random tex||<script>alert(0)</script>||random text."
		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+254346477")
		assert.Equal(t, http.StatusCreated, errCode)
		w := sendCreateProjectRequest(r, projectName, agent)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateProjectPayloadWithSQL", func(t *testing.T) {
		projectName := "random text||SELECT * FROM projects||random text."
		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+254346477")
		assert.Equal(t, http.StatusCreated, errCode)
		w := sendCreateProjectRequest(r, projectName, agent)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHasMaliciousContent(t *testing.T) {
	y, _ := U.HasMaliciousContent("text text text <script>|| text.")
	assert.True(t, y)

	y, _ = U.HasMaliciousContent("text text text SELECT * FROM x; text.")
	assert.True(t, y)

	y, _ = U.HasMaliciousContent("text text text INSERT INTO x VALUES(y);|| text.")
	assert.True(t, y)

	y, _ = U.HasMaliciousContent("text text text UPDATE x SET y=10 text.")
	assert.True(t, y)

	y, _ = U.HasMaliciousContent("text text text DELETE FROM text.")
	assert.True(t, y)

	y, _ = U.HasMaliciousContent("text text text ALTER TABLE x ADD COLUMN y text.")
	assert.True(t, y)

	// Usage of SQL keywords without valid query, should not be considered malicious.
	y, _ = U.HasMaliciousContent("text text text|| SELECT || DELETE || UPDATE || INSERT || ALTER || text.")
	assert.False(t, y)
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
		assert.NotEqual(t, "", jsonResponseMap["id"])
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
		assert.Nil(t, jsonResponseMap["salesforce_touch_points"])
		assert.Nil(t, jsonResponseMap["hubspot_touch_points"])
		assert.Equal(t, 16, len(jsonResponseMap))
		id, _ := strconv.Atoi(jsonResponseMap["id"].(string))
		w = sendEditProjectRequest(r, int64(id), "edit", agent)
		assert.Equal(t, http.StatusCreated, w.Code)
		jsonResponse, _ = ioutil.ReadAll(w.Body)
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotEqual(t, "", jsonResponseMap["id"])
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
		assert.Nil(t, jsonResponseMap["salesforce_touch_points"])
		assert.Nil(t, jsonResponseMap["hubspot_touch_points"])
		assert.Equal(t, 16, len(jsonResponseMap))
		id, _ = strconv.Atoi(jsonResponseMap["id"].(string))
		w = sendEditProjectRequest(r, int64(id), "edit@@", agent)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAccessControl(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	// create 2 projects belonging to 2 different agents
	// proj1  created by agent1 is set as demo project
	// on enabling the flag,
	// try to access all projects from agent 2, demo will also be listed but project settings API will fail with 405
	// on disabling the flag it lists only agent2's project and project setting will throw 403
	t.Run("CreateProject", func(t *testing.T) {
		projectName := "demoProject"
		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+254346477")
		assert.Equal(t, http.StatusCreated, errCode)
		w := sendCreateProjectRequest(r, projectName, agent)
		assert.Equal(t, http.StatusCreated, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		id, _ := strconv.Atoi(jsonResponseMap["id"].(string))
		demoProjectId := int64(id)

		projectName1 := "Test Project"
		agent1, errCode1 := SetupAgentReturnDAO(getRandomEmail(), "+12345678")
		assert.Equal(t, http.StatusCreated, errCode1)
		w1 := sendCreateProjectRequest(r, projectName1, agent1)
		assert.Equal(t, http.StatusCreated, w1.Code)
		jsonResponse1, _ := ioutil.ReadAll(w1.Body)
		var jsonResponseMap1 map[string]interface{}
		json.Unmarshal(jsonResponse1, &jsonResponseMap1)
		id, _ = strconv.Atoi(jsonResponseMap1["id"].(string))
		agentProjectId := int64(id)

		var trueFlag = true
		C.GetConfig().EnableDemoReadAccess = &trueFlag
		C.GetConfig().DemoProjectIds = []string{fmt.Sprintf("%v", demoProjectId)}
		w = sendGetProjectsRequest(r, agent1)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ = ioutil.ReadAll(w.Body)
		projects := make(map[int64][]model.Project)
		json.Unmarshal(jsonResponse, &projects)
		assert.Equal(t, 2, len(projects))
		assert.Equal(t, projects[2][0].Name, "Test Project")
		assert.Equal(t, projects[1][0].Name, "demoProject")
		assert.NotEqual(t, projects[2][0].PrivateToken, "")
		assert.Equal(t, projects[1][0].PrivateToken, "")

		w = sendGetProjectSettingsReq(r, agentProjectId, agent1)
		assert.Equal(t, http.StatusOK, w.Code)
		w = sendGetProjectSettingsReq(r, demoProjectId, agent1)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

		C.GetConfig().DemoProjectIds = []string{}
		w = sendGetProjectSettingsReq(r, demoProjectId, agent1)
		assert.Equal(t, http.StatusForbidden, w.Code)
		w = sendGetProjectSettingsReq(r, demoProjectId, agent1)
		assert.Equal(t, http.StatusForbidden, w.Code)
		w = sendGetProjectsRequest(r, agent1)
		jsonResponse, _ = ioutil.ReadAll(w.Body)
		projects = make(map[int64][]model.Project)
		json.Unmarshal(jsonResponse, &projects)
		assert.Equal(t, 1, len(projects))

		var falseFlag = false
		C.GetConfig().EnableDemoReadAccess = &falseFlag
		C.GetConfig().DemoProjectIds = []string{fmt.Sprintf("%v", demoProjectId)}
		w = sendGetProjectSettingsReq(r, demoProjectId, agent1)
		assert.Equal(t, http.StatusForbidden, w.Code)
		w = sendGetProjectsRequest(r, agent1)
		jsonResponse, _ = ioutil.ReadAll(w.Body)
		projects = make(map[int64][]model.Project)
		json.Unmarshal(jsonResponse, &projects)
		assert.Equal(t, 1, len(projects))

		C.GetConfig().EnableDemoReadAccess = &trueFlag
		w = sendGetProjectSettingsReq(r, demoProjectId, agent)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		C.GetConfig().ProjectAnalyticsWhitelistedUUIds = []string{agent.UUID}
		w = sendGetProjectSettingsReq(r, demoProjectId, agent)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
