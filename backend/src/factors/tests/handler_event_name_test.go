package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	M "factors/model"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendGetEventNamesRequest(projectId uint64, agent *M.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := U.NewRequestBuilder(http.MethodGet, fmt.Sprintf("/projects/%d/event_names", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error getting event names.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func createEventWithTimestampByName(t *testing.T, project *M.Project, user *M.User, name string, timestamp int64) (*M.EventName, *M.Event) {
	eventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: name})
	assert.NotNil(t, eventName)
	event, errCode := M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID, Timestamp: timestamp})
	assert.Equal(t, http.StatusCreated, errCode)
	return eventName, event
}

func TestGetEventNamesHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	user, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.NotNil(t, user)
	assert.Equal(t, http.StatusCreated, errCode)

	timestamp := U.UnixTimeBeforeAWeek()
	timeWithinWeek := timestamp + 3600
	timeBeforeWeek := timestamp - 3600
	createEventWithTimestampByName(t, project, user, "event3", timeBeforeWeek)
	createEventWithTimestampByName(t, project, user, "event4", timeBeforeWeek)

	createEventWithTimestampByName(t, project, user, "event1", timeWithinWeek)
	createEventWithTimestampByName(t, project, user, "event1", timeWithinWeek)
	createEventWithTimestampByName(t, project, user, "event2", timeWithinWeek)
	createEventWithTimestampByName(t, project, user, "event2", timeWithinWeek)
	createEventWithTimestampByName(t, project, user, "event2", timeWithinWeek)

	w := sendGetEventNamesRequest(project.ID, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	eventNames := make([]string, 0, 0)
	json.Unmarshal(jsonResponse, &eventNames)
	assert.Len(t, eventNames, 4)

	// should contain events ordered by occurrence count.
	assert.Equal(t, "event2", eventNames[0])
	assert.Equal(t, "event1", eventNames[1])
	// should contain all event names even though not
	// occurred on the window.
	assert.Equal(t, "event3", eventNames[2])
	assert.Equal(t, "event4", eventNames[3])

}
