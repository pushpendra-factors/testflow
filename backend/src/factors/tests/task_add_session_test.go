package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	M "factors/model"
	SDK "factors/sdk"
	TaskSession "factors/task/session"
	U "factors/util"
)

func TestAddSession(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	// Test: New user with one event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	randomEventName := U.RandomLowerAphaNumString(10)
	trackEventProperties := U.PropertiesMap{
		U.EP_REFERRER:     "www.google.com",
		U.EP_PAGE_URL:     "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example.com/1/2?x=1",
	}
	trackUserProperties := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload := SDK.TrackPayload{
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties,
		UserProperties:  trackUserProperties,
	}
	status, response := SDK.Track(project.ID, &trackPayload, true) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId := response.EventId
	userId := response.UserId

	// no session created.
	_, errCode := M.GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	_, err = TaskSession.AddSession([]uint64{project.ID}, 60, 0, 1)
	assert.Nil(t, err)

	// session event_name should have been created.
	sessionEventName, errCode := M.GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusFound, errCode)

	latestSessionEvent1, errCode := M.GetLatestEventOfUserByEventNameId(project.ID,
		response.UserId, sessionEventName.ID, timestamp-3, timestamp+3)
	assert.Equal(t, http.StatusFound, errCode)
	// session event timestamp should be less than first event
	// timestamp of session.
	assert.True(t, latestSessionEvent1.Timestamp < timestamp)
	assert.Equal(t, timestamp-1, latestSessionEvent1.Timestamp)
	// session_id should not be set to session event.
	assert.Nil(t, latestSessionEvent1.SessionId)

	// session event properties added from event properties.
	lsEventProperties1, err := U.DecodePostgresJsonb(&latestSessionEvent1.Properties)
	assert.Nil(t, err)
	assert.Equal(t, trackEventProperties[U.EP_REFERRER], (*lsEventProperties1)[U.SP_INITIAL_REFERRER])
	assert.Equal(t, trackEventProperties[U.EP_PAGE_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_URL])
	assert.Equal(t, trackEventProperties[U.EP_PAGE_RAW_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_RAW_URL])
	assert.Equal(t, float64(1), (*lsEventProperties1)[U.SP_PAGE_COUNT])
	assert.Equal(t, float64(1), (*lsEventProperties1)[U.SP_SPENT_TIME])
	// session event properties added from user properties.
	assert.Equal(t, trackUserProperties[U.UP_OS], (*lsEventProperties1)[U.UP_OS])
	assert.Equal(t, trackUserProperties[U.UP_OS_VERSION], (*lsEventProperties1)[U.UP_OS_VERSION])

	// event should have been associated with session_id.
	event, errCode := M.GetEventById(project.ID, eventId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, latestSessionEvent1.ID, *event.SessionId)

	// Test: New events without session for existing user with session.
	// Since there is continious activity, last session should be continued.
	timestamp = timestamp + 1
	randomEventName = U.RandomLowerAphaNumString(10)
	trackEventProperties1 := U.PropertiesMap{
		U.EP_REFERRER:     "www.yahoo.com",
		U.EP_PAGE_URL:     "https://example1.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example1.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: trackEventProperties1,
	}
	status, response = SDK.Track(project.ID, &trackPayload, true) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId

	timestamp = timestamp + 1
	randomEventName = U.RandomLowerAphaNumString(10)
	trackEventProperties2 := U.PropertiesMap{
		U.EP_REFERRER:     "www.facebook.com",
		U.EP_PAGE_URL:     "https://example2.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example2.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: trackEventProperties2,
	}
	status, response = SDK.Track(project.ID, &trackPayload, true) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	// inactivity.
	timestamp = timestamp + (35 * 60) // + 35 mins
	event3Timestamp := timestamp
	randomEventName = U.RandomLowerAphaNumString(10)
	trackEventProperties3 := U.PropertiesMap{
		U.EP_REFERRER:     "www.bing.com",
		U.EP_PAGE_URL:     "https://example3.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example3.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: trackEventProperties3,
	}
	status, response = SDK.Track(project.ID, &trackPayload, true) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	timestamp = timestamp + 1
	randomEventName = U.RandomLowerAphaNumString(10)
	trackEventProperties4 := U.PropertiesMap{
		U.EP_REFERRER:     "www.hacker.com",
		U.EP_PAGE_URL:     "https://example4.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example4.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: trackEventProperties4,
	}
	status, response = SDK.Track(project.ID, &trackPayload, true) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, 60, 0, 1)
	assert.Nil(t, err)

	// event1 should have been associated with latest session_id.
	event1, errCode := M.GetEventById(project.ID, eventId1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, latestSessionEvent1.ID, *event1.SessionId)

	// event2 should have been associated with latest session_id.
	event2, errCode := M.GetEventById(project.ID, eventId2)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, latestSessionEvent1.ID, *event2.SessionId)

	// last session's properties should be updated after continuing the same session.
	latestSessionEvent1, errCode = M.GetEventById(project.ID, latestSessionEvent1.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, latestSessionEvent1)
	lsEventProperties1, err = U.DecodePostgresJsonb(&latestSessionEvent1.Properties)
	assert.Nil(t, err)
	// should have intial event's referrer, before continuing session.
	assert.Equal(t, trackEventProperties[U.EP_REFERRER], (*lsEventProperties1)[U.SP_INITIAL_REFERRER])
	// should have lastest event's page_url after continuing session.
	assert.Equal(t, trackEventProperties2[U.EP_PAGE_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_URL])
	assert.Equal(t, trackEventProperties2[U.EP_PAGE_RAW_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_RAW_URL])
	assert.Equal(t, float64(3), (*lsEventProperties1)[U.SP_PAGE_COUNT])
	assert.Equal(t, float64(3), (*lsEventProperties1)[U.SP_SPENT_TIME])

	// session created after in-activity.
	latestSessionEvent2, errCode := M.GetLatestEventOfUserByEventNameId(project.ID,
		userId, sessionEventName.ID, event3Timestamp-10, event3Timestamp+10)
	assert.Equal(t, http.StatusFound, errCode)
	assert.True(t, latestSessionEvent2.Timestamp < event3Timestamp)
	assert.Equal(t, event3Timestamp-1, latestSessionEvent2.Timestamp)
	assert.Nil(t, latestSessionEvent2.SessionId)

	// event 3 and 4 should get new session_id as this last set of
	// events on right side of inactivity.
	event3, errCode := M.GetEventById(project.ID, eventId3)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, latestSessionEvent2.ID, *event3.SessionId)
	event4, errCode := M.GetEventById(project.ID, eventId4)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, latestSessionEvent2.ID, *event4.SessionId)

	// event properties of new session created after inactivity.
	lsEventProperties2, err := U.DecodePostgresJsonb(&latestSessionEvent2.Properties)
	assert.Nil(t, err)
	// should have intial event's referrer, before continuing session.
	assert.Equal(t, trackEventProperties3[U.EP_REFERRER], (*lsEventProperties2)[U.SP_INITIAL_REFERRER])
	// should have lastest event's page_url after continuing session.
	assert.Equal(t, trackEventProperties4[U.EP_PAGE_URL], (*lsEventProperties2)[U.SP_LATEST_PAGE_URL])
	assert.Equal(t, trackEventProperties4[U.EP_PAGE_RAW_URL], (*lsEventProperties2)[U.SP_LATEST_PAGE_RAW_URL])
	assert.Equal(t, float64(3), (*lsEventProperties1)[U.SP_PAGE_COUNT])
	assert.Equal(t, float64(3), (*lsEventProperties1)[U.SP_SPENT_TIME])

	// Test: Create new session for event with marketing property
	// even though there was activity.
	timestamp = timestamp + 1
	trackPayload = SDK.TrackPayload{
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.EP_CAMPAIGN: "summer_sale",
		},
	}

	status, response = SDK.Track(project.ID, &trackPayload, true) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	// Should not create session for last event timestmap  - 30 mins.
	_, err = TaskSession.AddSession([]uint64{project.ID}, 60, 0, 1)
	assert.Nil(t, err)

	event5, errCode := M.GetEventById(project.ID, eventId5)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, event5.SessionId)
	// New session should have been created.
	assert.NotEqual(t, latestSessionEvent2.SessionId, event5.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, 60, 0, 1)
	assert.Nil(t, err)
	assert.Equal(t, statusMap[project.ID].Status, "not_modified")
}

func TestAddSessionCreationBufferTime(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	// Event before session buffer time.
	timestamp := U.UnixTimeBeforeDuration(time.Minute * 35)
	randomEventName := U.RandomLowerAphaNumString(10)
	trackPayload := SDK.TrackPayload{
		Name:      randomEventName,
		Timestamp: timestamp,
	}
	status, response := SDK.Track(project.ID, &trackPayload, true) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId := response.EventId
	userId := response.UserId

	timestamp = U.UnixTimeBeforeDuration(time.Minute * 15)
	trackPayload = SDK.TrackPayload{
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
	}
	status, response = SDK.Track(project.ID, &trackPayload, true) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId

	timestamp = time.Now().Unix()
	trackPayload = SDK.TrackPayload{
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
	}
	status, response = SDK.Track(project.ID, &trackPayload, true) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	// Should not create session for last event timestmap  - 30 mins.
	_, err = TaskSession.AddSession([]uint64{project.ID}, 60, 30, 1)
	assert.Nil(t, err)

	event, errCode := M.GetEventById(project.ID, eventId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, event.SessionId)

	// events within buffer time.
	event1, errCode := M.GetEventById(project.ID, eventId1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Nil(t, event1.SessionId)

	event2, errCode := M.GetEventById(project.ID, eventId2)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Nil(t, event2.SessionId)
}

func TestGetAddSessionAllowedProjects(t *testing.T) {
	project1, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	project2, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	allowedProjectIds, errCode := TaskSession.GetAddSessionAllowedProjects(
		fmt.Sprintf("%d,%d", project1.ID, project2.ID), "")
	assert.Equal(t, http.StatusFound, errCode)
	assert.Contains(t, allowedProjectIds, project1.ID)
	assert.Contains(t, allowedProjectIds, project2.ID)

	allowedProjectIds, errCode = TaskSession.GetAddSessionAllowedProjects(
		fmt.Sprintf("%d", project1.ID), fmt.Sprintf("%d", project2.ID))
	assert.Equal(t, http.StatusFound, errCode)
	assert.Contains(t, allowedProjectIds, project1.ID)
	assert.NotContains(t, allowedProjectIds, project2.ID)

	allowedProjectIds, errCode = TaskSession.GetAddSessionAllowedProjects(
		"*", fmt.Sprintf("%d", project2.ID))
	assert.Equal(t, http.StatusFound, errCode)
	assert.Contains(t, allowedProjectIds, project1.ID)
	assert.NotContains(t, allowedProjectIds, project2.ID)
}
