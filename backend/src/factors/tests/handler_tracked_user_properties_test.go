package tests

import (
	"encoding/json"
	"factors/handler/helpers"
	"factors/task/event_user_cache"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	C "factors/config"

	H "factors/handler"
	"factors/model/model"
	"factors/model/store"

	V1 "factors/handler/v1"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func createProjectAgentEventsTrackedUserProperty(r *gin.Engine) (uint64, *model.Agent) {

	C.GetConfig().LookbackWindowForEventUserCache = 1

	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, agent, _ := SetupProjectWithAgentDAO()

	createdUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})

	rEventName := "event1"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	rEventName = "event2"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up1": "uv1"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	rEventName = "event3"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up2": "uv2"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	configs := make(map[string]interface{})
	configs["rollupLookback"] = 1
	event_user_cache.DoRollUpSortedSet(configs)
	return project.ID, agent
}

func sendCreateTrackedUserProperty(r *gin.Engine, request V1.CreateTrackeduserPropertyParams, agent *model.Agent, projectID uint64) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/factors/tracked_user_property", projectID)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating tracked trackeduserproperty")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetAllTrackedUserPropertyRequest(r *gin.Engine, agent *model.Agent, projectID uint64) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/factors/tracked_user_property", projectID)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error getting tracked trackeduserproperty.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendRemoveTrackedUserPropertyRequest(r *gin.Engine, agent *model.Agent, projectID uint64, request V1.RemoveFactorsTrackedUserPropertyParams) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodDelete, fmt.Sprintf("/projects/%d/v1/factors/tracked_user_property/remove", projectID)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error removing trackeduserproperty")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func TestCreateTrackedUserProperty(t *testing.T) {
	successTrackedUPIds := make([]int64, 0)
	type errorObject struct {
		Error string `json:"error"`
	}
	type successObject struct {
		Id     int64  `json:"id"`
		Status string `json:"status"`
	}
	r := gin.Default()
	H.InitAppRoutes(r)
	projectId, agent := createProjectAgentEventsFactorsTrackedEvents(r)
	C.GetConfig().ActiveFactorsTrackedUserPropertiesLimit = 50

	request := V1.CreateTrackeduserPropertyParams{}
	request.UserPropertyName = "up1"

	// Happy path
	request = V1.CreateTrackeduserPropertyParams{}
	request.UserPropertyName = "up1"
	w := sendCreateTrackedUserProperty(r, request, agent, projectId)
	assert.Equal(t, http.StatusCreated, w.Code)
	var obj successObject
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &obj)
	successTrackedUPIds = append(successTrackedUPIds, obj.Id)

	// tracked user property - that doesnt exist in the database
	request = V1.CreateTrackeduserPropertyParams{}
	request.UserPropertyName = "up4"
	w = sendCreateTrackedUserProperty(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var err errorObject
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "user property Not found", err.Error)

	// tracked user property - already tracked
	request = V1.CreateTrackeduserPropertyParams{}
	request.UserPropertyName = "up1"
	w = sendCreateTrackedUserProperty(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	err = errorObject{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "Tracked user property already exist", err.Error)

	// get all tracked user property
	w = sendGetAllTrackedUserPropertyRequest(r, agent, projectId)
	assert.Equal(t, http.StatusOK, w.Code)
	trackedUserProperty := []model.FactorsTrackedUserProperty{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &trackedUserProperty)
	assert.Equal(t, successTrackedUPIds[0], int64(trackedUserProperty[0].ID))
	assert.Equal(t, true, trackedUserProperty[0].IsActive)

	// remove user property
	removeRequest := V1.RemoveFactorsTrackedUserPropertyParams{
		ID: successTrackedUPIds[0],
	}
	w = sendRemoveTrackedUserPropertyRequest(r, agent, projectId, removeRequest)
	assert.Equal(t, http.StatusOK, w.Code)
	obj = successObject{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &obj)
	assert.Equal(t, successTrackedUPIds[0], obj.Id)

	// remove user property - already inactive
	removeRequest = V1.RemoveFactorsTrackedUserPropertyParams{
		ID: successTrackedUPIds[0],
	}
	w = sendRemoveTrackedUserPropertyRequest(r, agent, projectId, removeRequest)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	err = errorObject{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "Tracked user property already deleted", err.Error)

	// get all user property
	w = sendGetAllTrackedUserPropertyRequest(r, agent, projectId)
	assert.Equal(t, http.StatusOK, w.Code)
	trackedUserProperty = []model.FactorsTrackedUserProperty{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &trackedUserProperty)
	assert.Equal(t, successTrackedUPIds[0], int64(trackedUserProperty[0].ID))
	assert.Equal(t, false, trackedUserProperty[0].IsActive)

	// Null AgentID
	id, errCode := store.GetStore().CreateFactorsTrackedUserProperty(projectId, "up2", "")
	assert.NotEqual(t, 0, id)
	assert.Equal(t, 201, errCode)

	// Limit exceeded
	C.GetConfig().ActiveFactorsTrackedUserPropertiesLimit = 0
	request = V1.CreateTrackeduserPropertyParams{}
	request.UserPropertyName = "up1"
	w = sendCreateTrackedUserProperty(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	err = errorObject{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "Tracked User Properties Count Exceeded", err.Error)
}
