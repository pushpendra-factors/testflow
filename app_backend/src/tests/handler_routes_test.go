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
	var project_id uint64
	var user_id string
	var event_name string

	// Create random project and a corresponding event_name and user.
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProject(&M.Project{Name: random_project_name})
	if err_code != M.DB_SUCCESS {
		return project_id, user_id, event_name, fmt.Errorf("Project Creation failed.")
	}
	user, err_code := M.CreateUser(&M.User{ProjectId: project.ID})
	if err_code != M.DB_SUCCESS {
		return project_id, user_id, event_name, fmt.Errorf("User Creation failed.")
	}
	en, err_code := M.CreateEventName(&M.EventName{ProjectId: project.ID, Name: "login"})
	if err_code != M.DB_SUCCESS {
		return project_id, user_id, event_name, fmt.Errorf("EventName Creation failed.")
	}
	project_id = project.ID
	user_id = user.ID
	event_name = en.Name
	return project_id, user_id, event_name, nil
}

func TestAPICreateAndGetEvent(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitRoutes(r)
	project_id, user_id, event_name, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	// Test CreateEvent.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{ "project_id": %d, "user_id": "%s", "event_name": "%s"}`,
		project_id, user_id, event_name))
	req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(project_id), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, user_id, jsonResponseMap["user_id"].(string))
	assert.Equal(t, event_name, jsonResponseMap["event_name"].(string))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.Equal(t, 7, len(jsonResponseMap))

	// Test GetEvent on the created id.
	id := jsonResponseMap["id"].(string)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/events/"+id, nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.Equal(t, id, jsonResponseMap["id"].(string))
	assert.Equal(t, float64(project_id), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, user_id, jsonResponseMap["user_id"].(string))
	assert.Equal(t, event_name, jsonResponseMap["event_name"].(string))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))

	// Test GetEvent on random id.
	id = "r4nd0m!234"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/events/"+id, nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestAPICreateEventWithAttributes(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitRoutes(r)
	project_id, user_id, event_name, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	// Test CreateEvent.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{ "project_id": %d, "user_id": "%s", "event_name": "%s", "properties": {"ip": "10.0.0.1", "mobile": true, "code": 1}}`,
		project_id, user_id, event_name))
	req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(project_id), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, user_id, jsonResponseMap["user_id"].(string))
	assert.Equal(t, event_name, jsonResponseMap["event_name"].(string))
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

func TestAPICreateEventBadRequest(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitRoutes(r)
	project_id, user_id, event_name, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	// Test CreateEvent with id.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{ "id": "a745814b-a820-4f34-a01a-34e623b9c1a2", "project_id": %d, "user_id": "%s", "event_name": "%s"}`,
		project_id, user_id, event_name))
	req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without project_id.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "user_id": "%s", "event_name": "%s"}`,
		user_id, event_name))
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without user_id.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "project_id": %d, "user_id": "", "event_name": "%s"}`,
		project_id, event_name))
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without event_name.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "project_id": %d, "user_id": "%s"}`,
		project_id, user_id))
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent invalid project_id.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "project_id": 0, "user_id": "%s", "event_name": "%s"}`,
		user_id, event_name))
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent invalid user_id.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "project_id": %d, "user_id": "random1234", "event_name": "%s"}`,
		project_id, event_name))
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent invalid event_name.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "project_id": %d, "user_id": "%s", "event_name": "random1234"}`,
		project_id, user_id))
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)
}
