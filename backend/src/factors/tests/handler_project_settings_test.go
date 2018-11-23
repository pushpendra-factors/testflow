package tests

import (
	"encoding/json"
	H "factors/handler"
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
	// project_id becomes unknown field as omitted on json.
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	var jsonRespMap1 map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonRespMap1)
	assert.NotNil(t, jsonRespMap1["error"])

	// Test update project settings with bad project id.
	w = ServePutRequest(r, fmt.Sprintf("/projects/%d/settings", 0), []byte(`{"auto_track": true}`))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	var jsonRespMap2 map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonRespMap2)
	assert.NotNil(t, jsonRespMap2["error"])
	assert.Equal(t, 1, len(jsonRespMap2))
}
