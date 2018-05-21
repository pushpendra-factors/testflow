package tests

import (
	"bytes"
	C "config"
	"encoding/json"
	H "handler"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCreateAndGetEvent(t *testing.T) {
	// Initialize routes.
	r := gin.Default()
	H.InitRoutes(r)

	// Test CreateEvent.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(`{ "account_id": "1", "user_id": "1", "event_name": "login"}`)
	req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, "1", jsonResponseMap["account_id"].(string))
	assert.Equal(t, "1", jsonResponseMap["user_id"].(string))
	assert.Equal(t, "login", jsonResponseMap["event_name"].(string))
	assert.Nil(t, jsonResponseMap["attributes"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	log.Info(string(jsonResponse))
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
	assert.Equal(t, "1", jsonResponseMap["account_id"].(string))
	assert.Equal(t, "1", jsonResponseMap["user_id"].(string))
	assert.Equal(t, "login", jsonResponseMap["event_name"].(string))
	assert.Nil(t, jsonResponseMap["attributes"])
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	log.Info(string(jsonResponse))

	// Test GetEvent on random id.
	id = "r4nd0m!234"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/events/"+id, nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusNotFound)

}

func TestCreateEventBadRequest(t *testing.T) {
	// Initialize routes.
	r := gin.Default()
	H.InitRoutes(r)

	// Test CreateEvent with id.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(`{ "id": "1234", "account_id": "1", "user_id": "1", "event_name": "login"}`)
	req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without account_id.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(`{ "id": "1234", "account_id": "", "user_id": "1", "event_name": "login"}`)
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without user_id.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(`{ "id": "1234", "account_id": "1", "user_id": "", "event_name": "login"}`)
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without event_name.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(`{ "id": "1234", "account_id": "1", "user_id": "1"}`)
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)
}

func TestMain(m *testing.M) {
	// Setup.
	// Initialize configs and connections.
	err := C.Init()
	if err != nil {
		log.Fatal("Failed to initialize config and services.")
		os.Exit(1)
	}
	retCode := m.Run()
	os.Exit(retCode)
}
