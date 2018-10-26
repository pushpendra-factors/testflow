package tests

import (
	"bytes"
	"encoding/json"
	H "factors/handler"
	M "factors/model"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TODO: Use testify.suites to avoid multiple initializations across these tests.
func SetupProject() (uint64, error) {
	var projectId uint64

	// Create random project.
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProject(&M.Project{Name: random_project_name})
	if err_code != M.DB_SUCCESS {
		return projectId, fmt.Errorf("Project Creation failed.")
	}
	projectId = project.ID
	return projectId, nil
}

func TestAPICreateAndGetUser(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	projectId, err := SetupProject()
	assert.Nil(t, err)
	customerUserId := "murthy@autometa"

	// Test CreateUser.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{"c_uid": "%s"}`, customerUserId))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users", projectId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(projectId), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, customerUserId, jsonResponseMap["c_uid"].(string))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.Equal(t, 6, len(jsonResponseMap))

	// Test GetUser on the created id.
	id := jsonResponseMap["id"].(string)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users/%s", projectId, id), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.Equal(t, id, jsonResponseMap["id"].(string))
	assert.Equal(t, float64(projectId), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, customerUserId, jsonResponseMap["c_uid"].(string))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))

	// Test GetUser on random id.
	id = "r4nd0m!234"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/projects/%d/users/%s", projectId, id), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPICreateUserEmptyAndWithAttributes(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	projectId, err := SetupProject()
	assert.Nil(t, err)
	customerUserId := "murthy@autometa"

	// Test CreateUser.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{ "c_uid": "%s", "properties": {"ip": "10.0.0.1", "mobile": true, "code": 1}}`,
		customerUserId))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users", projectId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(projectId), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, customerUserId, jsonResponseMap["c_uid"].(string))
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.NotNil(t, jsonResponseMap["properties"])
	propertiesMap := jsonResponseMap["properties"].(map[string]interface{})
	assert.Equal(t, "10.0.0.1", propertiesMap["ip"].(string))
	assert.Equal(t, true, propertiesMap["mobile"].(bool))
	assert.Equal(t, 1.0, propertiesMap["code"].(float64))
	assert.Equal(t, 6, len(jsonResponseMap))

	// Test CreateUser without customerUserId and data.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(`{}`)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects/%d/users", projectId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(projectId), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, "", jsonResponseMap["c_uid"].(string))
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.Equal(t, 6, len(jsonResponseMap))
}

func TestAPICreateUserWithCreatedTime(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	projectId, err := SetupProject()
	assert.Nil(t, err)
	customerUserId := "murthy@autometa"

	// Test CreateUser.
	w := httptest.NewRecorder()
	timeStr := "2016-06-11T15:40:38.477168Z"

	var reqBodyStr = []byte(fmt.Sprintf(`{ "c_uid": "%s", "created_at": "%s", "updated_at": "%s"}`,
		customerUserId, timeStr, timeStr))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users", projectId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(projectId), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, customerUserId, jsonResponseMap["c_uid"].(string))
	assert.Equal(t, timeStr, jsonResponseMap["created_at"].(string))
	assert.Equal(t, timeStr, jsonResponseMap["updated_at"].(string))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.Equal(t, 6, len(jsonResponseMap))
}

func TestAPICreateUserBadRequest(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	projectId, err := SetupProject()
	assert.Nil(t, err)
	customerUserId := "murthy@autometa"

	// Test CreateUser with id.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{ "id": "a745814b-a820-4f34-a01a-34e623b9c1a2", "c_uid": "%s"}`, customerUserId))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users", projectId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateUser without projectId.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "c_uid": "%s"}`, customerUserId))
	req, _ = http.NewRequest("POST", "/projects//users", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateUser invalid projectId.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{"c_uid": "%s"}`, customerUserId))
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects/%d/users", projectId+10), bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)
}

func TestAPIGetUsers(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	projectId, err := SetupProject()
	assert.Nil(t, err)

	// Create 100 Users.
	var users []map[string]interface{}
	numUsers := 100
	for i := 0; i < numUsers; i++ {
		w := httptest.NewRecorder()
		var reqBodyStr = []byte(`{}`)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users", projectId),
			bytes.NewBuffer(reqBodyStr))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var userMap map[string]interface{}
		json.Unmarshal(jsonResponse, &userMap)
		users = append(users, userMap)
	}

	// Default values of offset and limit. Not sent in params.
	offset := 0
	limit := 10
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users", projectId), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var retUsers []map[string]interface{}
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, limit, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:offset+limit], retUsers)

	offset = 25
	limit = 20
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users?offset=%d&limit=%d", projectId, offset, limit), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, limit, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:offset+limit], retUsers)

	// Overflow
	offset = 95
	limit = 10
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users?offset=%d&limit=%d", projectId, offset, limit), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, numUsers-95, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:numUsers], retUsers)
}

func assertUserMapsWithOffset(t *testing.T, expectedUsers []map[string]interface{}, actualUsers []map[string]interface{}) {
	assert.Equal(t, len(expectedUsers), len(actualUsers))
	for i := 0; i < len(actualUsers); i++ {
		expectedUser := expectedUsers[i]
		actualUser := actualUsers[i]
		assert.Equal(t, expectedUser["id"].(string), actualUser["id"].(string))
		assert.Equal(t, expectedUser["project_id"].(float64), actualUser["project_id"].(float64))
		assert.Equal(t, expectedUser["c_uid"].(string), actualUser["c_uid"].(string))
		assert.Nil(t, actualUser["properties"])
		assert.NotNil(t, actualUser["created_at"].(string))
		assert.NotNil(t, actualUser["updated_at"].(string))
		assert.Equal(t, actualUser["created_at"].(string), actualUser["updated_at"].(string))
	}
}
