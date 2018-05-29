package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	H "handler"
	"io/ioutil"
	M "model"
	"net/http"
	"net/http/httptest"
	"testing"
	U "util"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TODO: Use testify.suites to avoid multiple initializations across these tests.
func SetupProjectUserEventName() (uint64, string, string, error) {
	var projectId uint64
	var userId string
	var eventName string

	// Create random project and a corresponding eventName and user.
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProject(&M.Project{Name: random_project_name})
	if err_code != M.DB_SUCCESS {
		return projectId, userId, eventName, fmt.Errorf("Project Creation failed.")
	}
	user, err_code := M.CreateUser(&M.User{ProjectId: project.ID})
	if err_code != M.DB_SUCCESS {
		return projectId, userId, eventName, fmt.Errorf("User Creation failed.")
	}
	en, err_code := M.CreateOrGetEventName(&M.EventName{ProjectId: project.ID, Name: "login"})
	if err_code != M.DB_SUCCESS {
		return projectId, userId, eventName, fmt.Errorf("EventName Creation failed.")
	}
	projectId = project.ID
	userId = user.ID
	eventName = en.Name
	return projectId, userId, eventName, nil
}

func TestAPICreateAndGetEvent(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitRoutes(r)
	projectId, userId, eventName, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	// Test CreateEvent.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{"event_name": "%s"}`, eventName))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/%s/events", projectId, userId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(projectId), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, userId, jsonResponseMap["user_id"].(string))
	assert.Equal(t, eventName, jsonResponseMap["event_name"].(string))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.Equal(t, 7, len(jsonResponseMap))

	// Test GetEvent on the created id.
	id := jsonResponseMap["id"].(string)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users/%s/events/%s", projectId, userId, id), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.Equal(t, id, jsonResponseMap["id"].(string))
	assert.Equal(t, float64(projectId), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, userId, jsonResponseMap["user_id"].(string))
	assert.Equal(t, eventName, jsonResponseMap["event_name"].(string))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))

	// Test GetEvent on random id.
	id = "r4nd0m!234"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users/%s/events/%s", projectId, userId, id), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test GetEvent with no id.
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users/%s/events/", projectId, userId), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPICreateEventWithAttributes(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitRoutes(r)
	projectId, userId, eventName, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	// Test CreateEvent.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{ "event_name": "%s", "properties": {"ip": "10.0.0.1", "mobile": true, "code": 1}}`,
		eventName))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/%s/events", projectId, userId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(projectId), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, userId, jsonResponseMap["user_id"].(string))
	assert.Equal(t, eventName, jsonResponseMap["event_name"].(string))
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.NotNil(t, jsonResponseMap["properties"])
	propertiesMap := jsonResponseMap["properties"].(map[string]interface{})
	assert.Equal(t, "10.0.0.1", propertiesMap["ip"].(string))
	assert.Equal(t, true, propertiesMap["mobile"].(bool))
	assert.Equal(t, 1.0, propertiesMap["code"].(float64))
	assert.Equal(t, 7, len(jsonResponseMap))
}

func TestAPICreateEventNonExistentEventName(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitRoutes(r)
	projectId, userId, _, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	// Test CreateEvent nonexistent eventName.
	w := httptest.NewRecorder()
	randomEventName := "random1234"
	reqBodyStr := []byte(fmt.Sprintf(`{"event_name": "%s"}`, randomEventName))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/%s/events", projectId, userId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	// Creates a corresponding eventName on the fly and returns created.
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(projectId), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, userId, jsonResponseMap["user_id"].(string))
	assert.Equal(t, randomEventName, jsonResponseMap["event_name"].(string))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.Equal(t, 7, len(jsonResponseMap))
}

func TestAPICreateEventBadRequest(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitRoutes(r)
	projectId, userId, eventName, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	// Test CreateEvent with id.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{ "id": "a745814b-a820-4f34-a01a-34e623b9c1a2", "event_name": "%s"}`, eventName))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/%s/events", projectId, userId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without projectId.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "event_name": "%s"}`, eventName))
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects//users/%s/events", userId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without userId.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "event_name": "%s"}`, eventName))
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects/%d/users//events", projectId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without eventName.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(`{}`)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/%s/events", projectId, userId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent invalid projectId.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{"event_name": "%s"}`, eventName))
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects/0/users/%s/events", userId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent invalid userId.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "event_name": "%s"}`, eventName))
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/random123/events", projectId),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)
}
