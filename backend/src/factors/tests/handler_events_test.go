package tests

import (
	"bytes"
	"encoding/json"
	H "factors/handler"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPICreateAndGetEvent(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
	assert.Nil(t, err)

	// Test CreateEvent.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{"event_name": "%s"}`, eventName.Name))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/%s/events", project.ID, user.ID),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(project.ID), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, user.ID, jsonResponseMap["user_id"].(string))
	assert.EqualValues(t, eventName.ID, jsonResponseMap["event_name_id"].(float64))
	assert.Equal(t, 1.0, jsonResponseMap["count"].(float64))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.Equal(t, 8, len(jsonResponseMap))

	// Test GetEvent on the created id.
	id := jsonResponseMap["id"].(string)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users/%s/events/%s", project.ID, user.ID, id), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.Equal(t, id, jsonResponseMap["id"].(string))
	assert.Equal(t, float64(project.ID), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, user.ID, jsonResponseMap["user_id"].(string))
	assert.EqualValues(t, eventName.ID, jsonResponseMap["event_name_id"].(float64))
	assert.Equal(t, 1.0, jsonResponseMap["count"].(float64))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.Equal(t, 8, len(jsonResponseMap))

	// Test CreateEvent with increment
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{"event_name": "%s"}`, eventName.Name))
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/%s/events", project.ID, user.ID),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(project.ID), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, user.ID, jsonResponseMap["user_id"].(string))
	assert.EqualValues(t, eventName.ID, jsonResponseMap["event_name_id"].(float64))
	assert.Equal(t, 2.0, jsonResponseMap["count"].(float64))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.Equal(t, 8, len(jsonResponseMap))

	// Test GetEvent on random id.
	id = "r4nd0m!234"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users/%s/events/%s", project.ID, user.ID, id), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test GetEvent with no id.
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users/%s/events/", project.ID, user.ID), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPICreateEventWithAttributes(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
	assert.Nil(t, err)

	// Test CreateEvent.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{ "event_name": "%s", "properties": {"ip": "10.0.0.1", "mobile": true, "code": 1}}`,
		eventName.Name))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/%s/events", project.ID, user.ID),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(project.ID), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, user.ID, jsonResponseMap["user_id"].(string))
	assert.EqualValues(t, eventName.ID, jsonResponseMap["event_name_id"].(float64))
	assert.Equal(t, 1.0, jsonResponseMap["count"].(float64))
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.NotNil(t, jsonResponseMap["properties"])
	propertiesMap := jsonResponseMap["properties"].(map[string]interface{})
	assert.Equal(t, "10.0.0.1", propertiesMap["ip"].(string))
	assert.Equal(t, true, propertiesMap["mobile"].(bool))
	assert.Equal(t, 1.0, propertiesMap["code"].(float64))
	assert.Equal(t, 8, len(jsonResponseMap))
}

func TestAPICreateEventNonExistentEventName(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	project, user, _, err := SetupProjectUserEventNameReturnDAO()
	assert.Nil(t, err)

	// Test CreateEvent nonexistent eventName.
	w := httptest.NewRecorder()
	randomEventName := "random1234"
	reqBodyStr := []byte(fmt.Sprintf(`{"event_name": "%s"}`, randomEventName))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/%s/events", project.ID, user.ID),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	// Creates a corresponding eventName on the fly and returns created.
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(project.ID), jsonResponseMap["project_id"].(float64))
	assert.Equal(t, user.ID, jsonResponseMap["user_id"].(string))
	assert.NotEqual(t, 0, jsonResponseMap["event_name_id"].(float64))
	assert.Equal(t, 1.0, jsonResponseMap["count"].(float64))
	assert.Nil(t, jsonResponseMap["properties"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.Equal(t, 8, len(jsonResponseMap))
}

func TestAPICreateEventBadRequest(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
	assert.Nil(t, err)

	// Test CreateEvent with id.
	w := httptest.NewRecorder()
	randomEventId := "a745814b-a820-4f34-a01a-34e623b9c1a2"
	var reqBodyStr = []byte(fmt.Sprintf(`{ "id": "%s" , "event_name": "%s"}`, randomEventId, eventName.Name))
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/%s/events", project.ID, user.ID),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotNil(t, len(jsonResponseMap["error"].(string)))

	// Test CreateEvent without project.ID.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "event_name": "%s"}`, eventName.Name))
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects//users/%s/events", user.ID),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without user.ID.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "event_name": "%s"}`, eventName.Name))
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects/%d/users//events", project.ID),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without eventName.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(`{}`)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/%s/events", project.ID, user.ID),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent invalid project.ID.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{"event_name": "%s"}`, eventName.Name))
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects/0/users/%s/events", user.ID),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent invalid user.ID.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "event_name": "%s"}`, eventName.Name))
	req, _ = http.NewRequest("POST", fmt.Sprintf("/projects/%d/users/random123/events", project.ID),
		bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)
}
