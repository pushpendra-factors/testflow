package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	H "handler"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPICreateProject(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitRoutes(r)
	projectName := "test_project_name"

	// Test CreateProject.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{"name": "%s"}`, projectName))
	req, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, jsonResponseMap["id"])
	assert.Equal(t, projectName, jsonResponseMap["name"].(string))
	assert.NotEqual(t, 0, len(jsonResponseMap["api_key"].(string)))
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.Equal(t, 5, len(jsonResponseMap))
}
