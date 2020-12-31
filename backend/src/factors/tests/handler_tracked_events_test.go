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

	V1 "factors/handler/v1"
	M "factors/model"
	U "factors/util"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func createProjectAgentEventsFactorsTrackedEvents(r *gin.Engine) (uint64, *M.Agent) {

	C.GetConfig().LookbackWindowForEventUserCache = 1

	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, agent, _ := SetupProjectWithAgentDAO()

	user, _ := M.CreateUser(&M.User{ProjectId: project.ID})

	rEventName := "event1"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, user.ID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	rEventName = "event2"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up1": "uv1"}}`, user.ID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	rEventName = "event3"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up2": "uv2"}}`, user.ID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	eventsLimit, propertyLimit, valueLimit, rollBackWindow := 1000, 10000, 10000, 1
	event_user_cache.DoRollUpAndCleanUp(&eventsLimit, &propertyLimit, &valueLimit, &rollBackWindow)
	return project.ID, agent
}

func sendCreateFactorsTrackedEvent(r *gin.Engine, request V1.CreateFactorsTrackedEventParams, agent *M.Agent, projectID uint64) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := U.NewRequestBuilder(http.MethodPost, fmt.Sprintf("/projects/%d/v1/factors/tracked_event", projectID)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating tracked event")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetAllFactorsTrackedEventRequest(r *gin.Engine, agent *M.Agent, projectID uint64) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	rb := U.NewRequestBuilder(http.MethodGet, fmt.Sprintf("/projects/%d/v1/factors/tracked_event", projectID)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error getting tracked event.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendRemoveFactorsTrackedEventRequest(r *gin.Engine, agent *M.Agent, projectID uint64, request V1.RemoveFactorsTrackedEventParams) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := U.NewRequestBuilder(http.MethodDelete, fmt.Sprintf("/projects/%d/v1/factors/tracked_event/remove", projectID)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error removing trackedevent")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func TestCreateFactorsTrackedEvent(t *testing.T) {
	successFactorsTrackedEventIds := make([]int64, 0)
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
	C.GetConfig().ActiveFactorsTrackedEventsLimit = 50

	// Non Admin
	request := V1.CreateFactorsTrackedEventParams{}
	request.EventName = "event1"

	// Happy path
	request = V1.CreateFactorsTrackedEventParams{}
	request.EventName = "event1"
	w := sendCreateFactorsTrackedEvent(r, request, agent, projectId)
	assert.Equal(t, http.StatusCreated, w.Code)
	var obj successObject
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &obj)
	successFactorsTrackedEventIds = append(successFactorsTrackedEventIds, obj.Id)

	// tracked event - event that doesnt exist in the database
	request = V1.CreateFactorsTrackedEventParams{}
	request.EventName = "event4"
	w = sendCreateFactorsTrackedEvent(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var err errorObject
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "Event Not found", err.Error)

	// tracked event - already tracked
	request = V1.CreateFactorsTrackedEventParams{}
	request.EventName = "event1"
	w = sendCreateFactorsTrackedEvent(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	err = errorObject{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "Tracked Event already exist", err.Error)

	// get all tracked events
	w = sendGetAllFactorsTrackedEventRequest(r, agent, projectId)
	assert.Equal(t, http.StatusOK, w.Code)
	trackedEvent := []M.FactorsTrackedEventInfo{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &trackedEvent)
	assert.Equal(t, successFactorsTrackedEventIds[0], int64(trackedEvent[0].ID))
	assert.Equal(t, true, trackedEvent[0].IsActive)
	assert.Equal(t, "event1", trackedEvent[0].Name)

	// remove event
	removeRequest := V1.RemoveFactorsTrackedEventParams{
		ID: successFactorsTrackedEventIds[0],
	}
	w = sendRemoveFactorsTrackedEventRequest(r, agent, projectId, removeRequest)
	assert.Equal(t, http.StatusOK, w.Code)
	obj = successObject{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &obj)
	assert.Equal(t, successFactorsTrackedEventIds[0], obj.Id)

	// get all events
	w = sendGetAllFactorsTrackedEventRequest(r, agent, projectId)
	assert.Equal(t, http.StatusOK, w.Code)
	trackedEvent = []M.FactorsTrackedEventInfo{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &trackedEvent)
	assert.Equal(t, successFactorsTrackedEventIds[0], int64(trackedEvent[0].ID))
	assert.Equal(t, false, trackedEvent[0].IsActive)
	assert.Equal(t, "event1", trackedEvent[0].Name)

	// Null AgentID
	id, errCode := M.CreateFactorsTrackedEvent(projectId, "event2", "")
	assert.NotEqual(t, 0, id)
	assert.Equal(t, 201, errCode)

	// Limit exceeded
	C.GetConfig().ActiveFactorsTrackedEventsLimit = 0
	request = V1.CreateFactorsTrackedEventParams{}
	request.EventName = "event1"
	w = sendCreateFactorsTrackedEvent(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	err = errorObject{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "Tracked Events Count Exceeded", err.Error)

}
