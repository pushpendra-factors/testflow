package tests

import (
	C "factors/config"
	"factors/handler/helpers"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendGetEventNamesApproxRequest(projectId uint64, agent *M.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	req, err := buildEventNameRequest(projectId, "approx", cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting event names.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetEventNamesExactRequest(projectId uint64, agent *M.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	req, err := buildEventNameRequest(projectId, "exact", cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting event names.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func buildEventNameRequest(projectId uint64, requestType, cookieData string) (*http.Request, error) {
	rb := U.NewRequestBuilder(http.MethodGet, fmt.Sprintf("/projects/%d/event_names?type=%s", projectId, requestType)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		return nil, err
	}
	return req, nil
}

func createEventWithTimestampByName(t *testing.T, project *M.Project, user *M.User, name string, timestamp int64) (*M.EventName, *M.Event) {
	eventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: name})
	assert.NotNil(t, eventName)
	event, errCode := M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID, Timestamp: timestamp})
	assert.Equal(t, http.StatusCreated, errCode)
	return eventName, event
}

/*func TestGetEventNamesHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	var eventNames = struct {
		EventNames []string `json:"event_names"`
		Exact      bool     `json:"exact"`
	}{}
	C.GetConfig().WhitelistedProjectIdsEventUserCache = "*"
	C.GetConfig().IsRealTimeEventUserCachingEnabled = true
	C.GetConfig().RealTimeEventUserCachingProjectIds = "*"

	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	ReinitialiseConfigForCachedEnabledProjects(fmt.Sprintf("%v", project.ID))
	w := sendGetEventNamesExactRequest(project.ID, agent, r)
	assert.Equal(t, http.StatusOK, w.Code) // Should be still 200 for no event_names with empty result set
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &eventNames)
	// should contain all event names.
	assert.Len(t, eventNames.EventNames, 0)

	user, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.NotNil(t, user)
	assert.Equal(t, http.StatusCreated, errCode)

	rEventName := "event1"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, user.ID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)
	rEventName = "event2"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, user.ID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	assert.Equal(t, http.StatusOK, w.Code)

	rEventName = "event1"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, user.ID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)

	// Test events ingested via sdk/track call
	w = sendGetEventNamesExactRequest(project.ID, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &eventNames)
	// should contain all event names.
	assert.Len(t, eventNames.EventNames, 3)
}*/
