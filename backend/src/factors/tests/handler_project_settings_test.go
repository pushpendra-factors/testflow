package tests

import (
	"encoding/json"
	H "factors/handler"
	M "factors/model"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPIGetProjectSettingHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	_, err = SetupProjectDependenciesReturnDAO(project)
	assert.Nil(t, err)

	projectSettingsURL := fmt.Sprintf("/projects/%d/settings", project.ID)

	// Test get project settings.
	w := ServeGetRequest(r, projectSettingsURL)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, jsonResponseMap["id"])
	assert.NotNil(t, jsonResponseMap["auto_track"])

	// Test get project settings with bad id.
	w = ServeGetRequest(r, fmt.Sprintf("/projects/%d/settings", 0))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	var jsonRespMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonRespMap)
	assert.NotNil(t, jsonRespMap["error"])
	assert.Equal(t, 1, len(jsonRespMap))
}

func TestAPIUpdateProjectSettingsHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	_, err = SetupProjectDependenciesReturnDAO(project)
	assert.Nil(t, err)

	projectSettingsURL := fmt.Sprintf("/projects/%d/settings", project.ID)

	// Test update project settings.
	w := ServePutRequest(r, projectSettingsURL, []byte(`{"auto_track": true}`))
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotNil(t, jsonResponseMap["auto_track"])

	// Test updating project id.
	var randomProjectId uint64
	randomProjectId = 999999999
	w = ServePutRequest(r, projectSettingsURL, []byte(fmt.Sprintf(`{"auto_track": true, "project_id":%d}`, randomProjectId)))
	assert.Equal(t, http.StatusOK, w.Code)
	// Tested by getting as M.ProjectSetting omits projectId on json ops.
	projectSetting, errCode := M.GetProjectSetting(project.ID)
	assert.NotNil(t, projectSetting)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	projectSetting, errCode = M.GetProjectSetting(randomProjectId)
	assert.Nil(t, projectSetting)
	assert.NotEqual(t, M.DB_SUCCESS, errCode)

	// Test update project settings with bad project id.
	w = ServePutRequest(r, fmt.Sprintf("/projects/%d/settings", 0), []byte(`{"auto_track": true}`))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	var jsonRespMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonRespMap)
	assert.NotNil(t, jsonRespMap["error"])
	assert.Equal(t, 1, len(jsonRespMap))
}
