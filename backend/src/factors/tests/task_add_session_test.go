package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	C "factors/config"
	M "factors/model"
	SDK "factors/sdk"
	TaskSession "factors/task/session"
	U "factors/util"
)

func assertAssociatedSession(t *testing.T, projectId uint64, eventIdsInOrder []string,
	skipSessionEventIds []string, message string) (sessionEvent *M.Event) {

	var firstEvent *M.Event
	for i, eventId := range eventIdsInOrder {
		event, errCode := M.GetEventById(projectId, eventId)
		assert.Equal(t, http.StatusFound, errCode, message)

		if i == 0 {
			assert.NotNil(t, event.SessionId, message)
			firstEvent = event
		}

		if i > 0 {
			var skipped bool
			for _, seid := range skipSessionEventIds {
				if seid == eventId {
					assert.Nil(t, event.SessionId, message)
					skipped = true
				}
			}

			if skipped {
				continue
			}

			// all event should have same session id.
			assert.Equal(t, *firstEvent.SessionId, *event.SessionId, message)
		}
	}

	// check session event
	sessionEvent, errCode := M.GetEventById(projectId, *firstEvent.SessionId)
	assert.Equal(t, http.StatusFound, errCode, message)
	assert.Equal(t, firstEvent.Timestamp-1, sessionEvent.Timestamp, message)

	return sessionEvent
}

func TestAddSession(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	// skip realtime session creation for project.
	C.GetConfig().SkipSessionProjectIds = fmt.Sprintf("%d", project.ID)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
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
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId := response.EventId
	userId := response.UserId

	// no session created.
	_, errCode := M.GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Name:      "non_web_event",
		Timestamp: timestamp,
		UserId:    userId,
	}
	// skip session event.
	status, response = SDK.Track(project.ID, &trackPayload, true, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	skipSessionEventId := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 30, 1)
	assert.Nil(t, err)

	U.SanitizeProperties(&trackEventProperties)
	sessionEvent1 := assertAssociatedSession(t, project.ID, []string{eventId, skipSessionEventId},
		[]string{skipSessionEventId}, "Session 1")
	// session event properties added from event properties.
	lsEventProperties1, err := U.DecodePostgresJsonb(&sessionEvent1.Properties)
	assert.Nil(t, err)
	assert.Equal(t, trackEventProperties[U.EP_REFERRER], (*lsEventProperties1)[U.SP_INITIAL_REFERRER])
	assert.Equal(t, trackEventProperties[U.EP_PAGE_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_URL])
	assert.Equal(t, trackEventProperties[U.EP_PAGE_RAW_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_RAW_URL])
	assert.Equal(t, float64(1), (*lsEventProperties1)[U.SP_PAGE_COUNT])
	assert.Equal(t, float64(1), (*lsEventProperties1)[U.SP_SPENT_TIME])
	// session event properties added from user properties.
	assert.Equal(t, trackUserProperties[U.UP_OS], (*lsEventProperties1)[U.UP_OS])
	assert.Equal(t, trackUserProperties[U.UP_OS_VERSION], (*lsEventProperties1)[U.UP_OS_VERSION])

	// check session count so far.
	event, errCode := M.GetEventById(project.ID, eventId)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesRecord, errCode := M.GetUserPropertiesRecord(project.ID, event.UserId, event.UserPropertiesId)
	userPropertiesMap, err := U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(1), (*userPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(1), (*userPropertiesMap)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(1), (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])

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
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
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
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	// inactivity.
	timestamp = timestamp + (35 * 60) // + 35 mins
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
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Name:      "non_web_event",
		Timestamp: timestamp,
		UserId:    userId,
	}
	status, response = SDK.Track(project.ID, &trackPayload, true, SDK.SourceJSSDK) // skip session.
	assert.Equal(t, http.StatusOK, status)
	skipSessionEventId1 := response.EventId

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
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 30, 1)
	assert.Nil(t, err)
	U.SanitizeProperties(&trackEventProperties2)

	// should have continue session for event 1 and 2. should have created new session for
	// event 3 and 4 because of inactivity.

	// event 1 and 2 should have continued session.
	sessionEvent1 = assertAssociatedSession(t, project.ID, []string{eventId, eventId1, eventId2},
		[]string{}, "Session 1 continued.")
	// last session's properties should be updated after continuing the same session.
	lsEventProperties1, err = U.DecodePostgresJsonb(&sessionEvent1.Properties)
	assert.Nil(t, err)
	// should have intial event's referrer, before continuing session.
	assert.Equal(t, trackEventProperties[U.EP_REFERRER], (*lsEventProperties1)[U.SP_INITIAL_REFERRER])
	// should have lastest event's page_url after continuing session.
	assert.Equal(t, trackEventProperties2[U.EP_PAGE_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_URL])
	assert.Equal(t, trackEventProperties2[U.EP_PAGE_RAW_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_RAW_URL])
	assert.Equal(t, float64(3), (*lsEventProperties1)[U.SP_PAGE_COUNT])
	assert.Equal(t, float64(5), (*lsEventProperties1)[U.SP_SPENT_TIME])

	// event 3 and skip session event 1 and event 4 should create new session,
	// without considering skip session event 1.
	sessionEvent2 := assertAssociatedSession(t, project.ID, []string{eventId3, skipSessionEventId1, eventId4},
		[]string{skipSessionEventId1}, "Session 2")
	assert.NotEqual(t, sessionEvent1.ID, sessionEvent2.ID)
	// event properties of new session created after inactivity.
	lsEventProperties2, err := U.DecodePostgresJsonb(&sessionEvent2.Properties)
	assert.Nil(t, err)

	U.SanitizeProperties(&trackEventProperties4)
	// should have intial event's referrer, before continuing session.
	assert.Equal(t, trackEventProperties3[U.EP_REFERRER], (*lsEventProperties2)[U.SP_INITIAL_REFERRER])
	// should have lastest event's page_url after continuing session.
	assert.Equal(t, trackEventProperties4[U.EP_PAGE_URL], (*lsEventProperties2)[U.SP_LATEST_PAGE_URL])
	assert.Equal(t, trackEventProperties4[U.EP_PAGE_RAW_URL], (*lsEventProperties2)[U.SP_LATEST_PAGE_RAW_URL])
	assert.Equal(t, float64(2), (*lsEventProperties2)[U.SP_PAGE_COUNT])
	assert.Equal(t, float64(4), (*lsEventProperties2)[U.SP_SPENT_TIME])

	// check session count so far.
	event4, errCode := M.GetEventById(project.ID, eventId4)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesRecord, errCode = M.GetUserPropertiesRecord(project.ID, event4.UserId, event4.UserPropertiesId)
	userPropertiesMap, err = U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(2), (*userPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(3), (*userPropertiesMap)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(5), (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])

	// Test: Create new session for event with marketing property,
	// followed by other events, even though there was continuos
	// activity from previous session.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "summer_sale",
		},
	}

	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	timestamp = timestamp + 2
	randomEventName = U.RandomLowerAphaNumString(10)
	trackPayload = SDK.TrackPayload{
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 30, 1)
	assert.Nil(t, err)

	// should have created session as campaign property exist.
	sessionEvent3 := assertAssociatedSession(t, project.ID, []string{eventId5, eventId6},
		[]string{}, "Session 3")
	assert.NotEqual(t, sessionEvent2.ID, sessionEvent3.ID)

	// check session count so far.
	event6, errCode := M.GetEventById(project.ID, eventId6)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesRecord, errCode = M.GetUserPropertiesRecord(project.ID, event6.UserId, event6.UserPropertiesId)
	userPropertiesMap, err = U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(3), (*userPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(5), (*userPropertiesMap)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(8), (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])

	// Test: Last event with marketing property.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId7 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 30, 1)
	assert.Nil(t, err)
	event7, errCode := M.GetEventById(project.ID, eventId7)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotNil(t, event7.SessionId)

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 30, 1)
	assert.Nil(t, err)
	// New session should be created after a new event.
	sessionEvent4 := assertAssociatedSession(t, project.ID, []string{eventId5, eventId6, eventId7}, []string{}, "Session 4")
	assert.Equal(t, sessionEvent3.ID, sessionEvent4.ID)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 30, 1)
	assert.Nil(t, err)
	assert.Equal(t, statusMap[project.ID].Status, "not_modified")
}

func TestAddSessionCreationBufferTime(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	C.GetConfig().SkipSessionProjectIds = fmt.Sprintf("%d", project.ID)

	// Event before session buffer time.
	timestamp := U.UnixTimeBeforeDuration(time.Minute * 35)
	randomEventName := U.RandomLowerAphaNumString(10)
	trackPayload := SDK.TrackPayload{
		Name:      randomEventName,
		Timestamp: timestamp,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK) // true: skips session.
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
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId

	timestamp = time.Now().Unix()
	trackPayload = SDK.TrackPayload{
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK) // true: skips session.
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

func TestDerivedUserPropertiesFromSession(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	C.GetConfig().SkipSessionProjectIds = fmt.Sprintf("%d", project.ID)

	// Test derived user_properties for user having multiple
	// user_properites state before adding session.
	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	randomEventName := U.RandomLowerAphaNumString(10)
	trackPayload := SDK.TrackPayload{
		Name:      randomEventName,
		Timestamp: timestamp,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
		},
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId

	event1, errCode := M.GetEventById(project.ID, eventId1)
	assert.Equal(t, errCode, http.StatusFound)

	timestamp = timestamp + 1
	trackPayload = SDK.TrackPayload{
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    response.UserId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
		},
		UserProperties: U.PropertiesMap{
			"property1": "value1",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 30, 1)
	assert.Nil(t, err)

	event2, errCode := M.GetEventById(project.ID, eventId2)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotNil(t, event2.SessionId)

	assert.NotEqual(t, event1.UserPropertiesId, event2.UserPropertiesId)

	userPropertiesRecord, errCode := M.GetUserPropertiesRecord(project.ID, event2.UserId, event2.UserPropertiesId)
	userPropertiesMap, err := U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(2), (*userPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(2), (*userPropertiesMap)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(2), (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
}
