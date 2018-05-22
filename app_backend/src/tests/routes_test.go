package tests

import (
	"bytes"
	C "config"
	"encoding/json"
	"fmt"
	H "handler"
	"io/ioutil"
	M "model"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	U "util"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var account_id uint64
var user_id string
var event_name string

func TestCreateAndGetEvent(t *testing.T) {
	// Initialize routes.
	r := gin.Default()
	H.InitRoutes(r)

	// Test CreateEvent.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{ "account_id": %d, "user_id": "%s", "event_name": "%s"}`,
		account_id, user_id, event_name))
	req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(account_id), jsonResponseMap["account_id"].(float64))
	assert.Equal(t, user_id, jsonResponseMap["user_id"].(string))
	assert.Equal(t, event_name, jsonResponseMap["event_name"].(string))
	assert.Nil(t, jsonResponseMap["attributes"])
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
	assert.Equal(t, float64(account_id), jsonResponseMap["account_id"].(float64))
	assert.Equal(t, user_id, jsonResponseMap["user_id"].(string))
	assert.Equal(t, event_name, jsonResponseMap["event_name"].(string))
	assert.Nil(t, jsonResponseMap["attributes"])
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

func TestCreateEventWithAttributes(t *testing.T) {
	// Initialize routes.
	r := gin.Default()
	H.InitRoutes(r)

	// Test CreateEvent.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{ "account_id": %d, "user_id": "%s", "event_name": "%s", "attributes": {"ip": "10.0.0.1", "mobile": true, "code": 1}}`,
		account_id, user_id, event_name))
	req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, len(jsonResponseMap["id"].(string)))
	assert.Equal(t, float64(account_id), jsonResponseMap["account_id"].(float64))
	assert.Equal(t, user_id, jsonResponseMap["user_id"].(string))
	assert.Equal(t, event_name, jsonResponseMap["event_name"].(string))
	assert.NotNil(t, jsonResponseMap["created_at"].(string))
	assert.NotNil(t, jsonResponseMap["updated_at"].(string))
	assert.Equal(t, jsonResponseMap["created_at"].(string), jsonResponseMap["updated_at"].(string))
	assert.NotNil(t, jsonResponseMap["attributes"])
	attributesMap := jsonResponseMap["attributes"].(map[string]interface{})
	assert.Equal(t, "10.0.0.1", attributesMap["ip"].(string))
	assert.Equal(t, true, attributesMap["mobile"].(bool))
	assert.Equal(t, 1.0, attributesMap["code"].(float64))
	assert.Equal(t, 7, len(jsonResponseMap))
}

func TestCreateEventBadRequest(t *testing.T) {
	// Initialize routes.
	r := gin.Default()
	H.InitRoutes(r)

	// Test CreateEvent with id.
	w := httptest.NewRecorder()
	var reqBodyStr = []byte(fmt.Sprintf(`{ "id": "a745814b-a820-4f34-a01a-34e623b9c1a2", "account_id": %d, "user_id": "%s", "event_name": "%s"}`,
		account_id, user_id, event_name))
	req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without account_id.
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
	reqBodyStr = []byte(fmt.Sprintf(`{ "account_id": %d, "user_id": "", "event_name": "%s"}`,
		account_id, event_name))
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent without event_name.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "account_id": %d, "user_id": "%s"}`,
		account_id, user_id))
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent invalid account_id.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "account_id": 0, "user_id": "%s", "event_name": "%s"}`,
		user_id, event_name))
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent invalid user_id.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "account_id": %d, "user_id": "random1234", "event_name": "%s"}`,
		account_id, event_name))
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)

	// Test CreateEvent invalid event_name.
	w = httptest.NewRecorder()
	reqBodyStr = []byte(fmt.Sprintf(`{ "account_id": %d, "user_id": "%s", "event_name": "random1234"}`,
		account_id, user_id))
	req, _ = http.NewRequest("POST", "/events", bytes.NewBuffer(reqBodyStr))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, []byte{}, jsonResponse)
}

func TestMain(m *testing.M) {
	// Setup.
	// Initialize configs and connections.
	if err := C.Init(); err != nil {
		log.Fatal("Failed to initialize config and services.")
		os.Exit(1)
	}
	if C.GetConfig().Env != C.DEVELOPMENT {
		log.Fatal("Environment is not Development.")
		os.Exit(1)
	}
	// Create random account and a corresponding event_name and user.
	random_account_name := U.RandomLowerAphaNumString(15)
	account, err_code := M.CreateAccount(&M.Account{Name: random_account_name})
	if err_code != -1 {
		log.Fatal("Account Creation failed.")
		os.Exit(1)
	}
	user, err_code := M.CreateUser(&M.User{AccountId: account.ID})
	if err_code != -1 {
		log.Fatal("User Creation failed.")
		os.Exit(1)
	}
	en, err_code := M.CreateEventName(&M.EventName{AccountId: account.ID, Name: "login"})
	if err_code != -1 {
		log.Fatal("EventName Creation failed.")
		os.Exit(1)
	}
	account_id = account.ID
	user_id = user.ID
	event_name = en.Name

	retCode := m.Run()
	os.Exit(retCode)
}
