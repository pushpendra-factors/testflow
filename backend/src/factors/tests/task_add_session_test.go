package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"

	C "factors/config"
	SDK "factors/sdk"
	TaskSession "factors/task/session"
	U "factors/util"
)

func assertAssociatedSession(t *testing.T, projectId uint64, eventIdsInOrder []string,
	skipSessionEventIds []string, message string) (sessionEvent *model.Event) {

	var firstEvent *model.Event
	for i, eventId := range eventIdsInOrder {
		event, errCode := store.GetStore().GetEventById(projectId, eventId)
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
	sessionEvent, errCode := store.GetStore().GetEventById(projectId, *firstEvent.SessionId)
	assert.Equal(t, http.StatusFound, errCode, message)
	assert.Equal(t, firstEvent.Timestamp-1, sessionEvent.Timestamp, message)

	return sessionEvent
}

func TestAddSessionOnUserWithContinuousEvents(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName := RandomURL()
	trackEventProperties := U.PropertiesMap{
		U.EP_REFERRER:        "www.google.com",
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
	}
	trackUserProperties := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload := SDK.TrackPayload{
		Auto:            true,
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
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
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

	// create new user_properties state, for testing session user_properties addition
	// on latest user_properties, which is not associated to any event.
	userProperties := postgres.Jsonb{RawMessage: json.RawMessage(`{"plan": "enterprise"}`)}
	newUserPropertiesID, _, errCode := store.GetStore().UpdateUserProperties(project.ID, userId, &userProperties, time.Now().Unix())
	user, _ := store.GetStore().GetUser(project.ID, userId)
	assert.NotNil(t, user)
	// new user_properties state should be the user's latest user_property state.
	assert.Equal(t, newUserPropertiesID, user.PropertiesId)

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
	assert.Equal(t, float64(10), (*lsEventProperties1)[U.SP_SPENT_TIME])
	// session event properties added from user properties.
	assert.Equal(t, trackUserProperties[U.UP_OS], (*lsEventProperties1)[U.UP_OS])
	assert.Equal(t, trackUserProperties[U.UP_OS_VERSION], (*lsEventProperties1)[U.UP_OS_VERSION])

	// check session user_properties so far, on both event associated
	// user_property and user's latest user_property.
	event, errCode := store.GetStore().GetEventById(project.ID, eventId)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err := U.DecodePostgresJsonb(event.UserProperties)
	if C.ShouldUseUserPropertiesTableForRead(project.ID) {
		userPropertiesRecord, errCode := store.GetStore().GetUserPropertiesRecord(project.ID, event.UserId, event.UserPropertiesId)
		assert.Equal(t, http.StatusFound, errCode)
		userPropertiesMap, err = U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	}
	assert.Nil(t, err)
	assert.Equal(t, float64(1), (*userPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(1), (*userPropertiesMap)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(10), (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	assert.Equal(t, trackUserProperties[U.UP_OS], (*userPropertiesMap)[U.UP_OS])
	// check latest user_properties state.
	user, _ = store.GetStore().GetUser(project.ID, event.UserId)
	lastestUserPropertiesMap, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.NotEmpty(t, user.Properties)
	if C.IsUserPropertiesTableWriteDeprecated(project.ID) {
		assert.Empty(t, event.UserPropertiesId)
	} else {
		assert.NotEmpty(t, event.UserPropertiesId)
	}
	assert.Equal(t, float64(1), (*lastestUserPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(1), (*lastestUserPropertiesMap)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(10), (*lastestUserPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	assert.Equal(t, trackUserProperties[U.UP_OS], (*lastestUserPropertiesMap)[U.UP_OS])

	// Test: New events without session for existing user with session.
	// Since there is continuous activity, last session should be continued.
	timestamp = timestamp + 1
	randomEventName = RandomURL()
	trackEventProperties1 := U.PropertiesMap{
		U.EP_REFERRER:     "www.yahoo.com",
		U.EP_PAGE_URL:     "https://example1.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example1.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: trackEventProperties1,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId

	timestamp = timestamp + 1
	randomEventName = RandomURL()
	trackEventProperties2 := U.PropertiesMap{
		U.EP_REFERRER:     "www.facebook.com",
		U.EP_PAGE_URL:     "https://example2.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example2.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
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
	randomEventName = RandomURL()
	trackEventProperties3 := U.PropertiesMap{
		U.EP_REFERRER:     "www.bing.com",
		U.EP_PAGE_URL:     "https://example3.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example3.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
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
	randomEventName = RandomURL()
	trackEventProperties4 := U.PropertiesMap{
		U.EP_REFERRER:     "www.hacker.com",
		U.EP_PAGE_URL:     "https://example4.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example4.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: trackEventProperties4,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
	// should have initial event's referrer, before continuing session.
	assert.Equal(t, trackEventProperties[U.EP_REFERRER], (*lsEventProperties1)[U.SP_INITIAL_REFERRER])
	// should have latest event's page_url after continuing session.
	assert.Equal(t, trackEventProperties2[U.EP_PAGE_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_URL])
	assert.Equal(t, trackEventProperties2[U.EP_PAGE_RAW_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_RAW_URL])
	assert.Equal(t, float64(3), (*lsEventProperties1)[U.SP_PAGE_COUNT])
	// event = 10ms, event1 = 1ms (default), event2 = 1ms (default).
	assert.Equal(t, float64(12), (*lsEventProperties1)[U.SP_SPENT_TIME])

	// event 3 and skip session event 1 and event 4 should create new session,
	// without considering skip session event 1.
	sessionEvent2 := assertAssociatedSession(t, project.ID, []string{eventId3, skipSessionEventId1, eventId4},
		[]string{skipSessionEventId1}, "Session 2")
	assert.NotEqual(t, sessionEvent1.ID, sessionEvent2.ID)
	// event properties of new session created after inactivity.
	lsEventProperties2, err := U.DecodePostgresJsonb(&sessionEvent2.Properties)
	assert.Nil(t, err)

	U.SanitizeProperties(&trackEventProperties4)
	// should have initial event's referrer, before continuing session.
	assert.Equal(t, trackEventProperties3[U.EP_REFERRER], (*lsEventProperties2)[U.SP_INITIAL_REFERRER])
	// should have latest event's page_url after continuing session.
	assert.Equal(t, trackEventProperties4[U.EP_PAGE_URL], (*lsEventProperties2)[U.SP_LATEST_PAGE_URL])
	assert.Equal(t, trackEventProperties4[U.EP_PAGE_RAW_URL], (*lsEventProperties2)[U.SP_LATEST_PAGE_RAW_URL])
	assert.Equal(t, float64(2), (*lsEventProperties2)[U.SP_PAGE_COUNT])
	assert.Equal(t, float64(2), (*lsEventProperties2)[U.SP_SPENT_TIME])

	// check session count so far.
	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err = U.DecodePostgresJsonb(event4.UserProperties)
	if C.ShouldUseUserPropertiesTableForRead(project.ID) {
		userPropertiesRecord, errCode := store.GetStore().GetUserPropertiesRecord(project.ID, event4.UserId, event4.UserPropertiesId)
		assert.Equal(t, http.StatusFound, errCode)
		userPropertiesMap, err = U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	}
	assert.Nil(t, err)

	assert.Equal(t, float64(2), (*userPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(3), (*userPropertiesMap)[U.UP_PAGE_COUNT])
	// This is because of two different user property id in the same session
	assert.Equal(t, float64(12), (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])

	// Test: Create new session for event with marketing property,
	// followed by other events, even though there was continuos
	// activity from previous session.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
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
	randomEventName = RandomURL()
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	// should have created session as campaign property exist.
	sessionEvent3 := assertAssociatedSession(t, project.ID, []string{eventId5, eventId6},
		[]string{}, "Session 3")
	assert.NotEqual(t, sessionEvent2.ID, sessionEvent3.ID)

	// check session count so far.
	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err = U.DecodePostgresJsonb(event6.UserProperties)
	if C.ShouldUseUserPropertiesTableForRead(project.ID) {
		userPropertiesRecord, errCode := store.GetStore().GetUserPropertiesRecord(project.ID, event6.UserId, event6.UserPropertiesId)
		assert.Equal(t, http.StatusFound, errCode)
		userPropertiesMap, err = U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	}
	assert.Nil(t, err)
	assert.Equal(t, float64(3), (*userPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(5), (*userPropertiesMap)[U.UP_PAGE_COUNT])
	// This is because of two different user property id in the same session
	assert.Equal(t, float64(14), (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])

	// Test: Last event with marketing property.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
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

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	// New session should be created after a new event.
	//sessionEvent4 := assertAssociatedSession(t, project.ID, []string{eventId5, eventId6}, []string{}, "Session 4")
	//assert.Equal(t, sessionEvent3.ID, sessionEvent4.ID)

	// Last event with marketing property should be process on next run of add session,
	// to avoid associating previous session.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event7.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}

func TestEventsConsiderationForAddingSession(t *testing.T) {
	t.Run("ShouldNotChangeSessionOfSessionAddedEvents", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		event3, _ := store.GetStore().GetEvent(project.ID, userId, eventId3)
		assert.Equal(t, event1.SessionId, event2.SessionId)
		assert.Equal(t, event1.SessionId, event3.SessionId)
		sessionID1 := event1.SessionId

		// Second run of add_session with same timerange.
		// Should not change the session associated.
		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// Sessions associated events should be the same.
		event1, _ = store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.Equal(t, sessionID1, event1.SessionId)
		event2, _ = store.GetStore().GetEvent(project.ID, userId, eventId2)
		event3, _ = store.GetStore().GetEvent(project.ID, userId, eventId3)
		assert.Equal(t, event1.SessionId, event2.SessionId)
		assert.Equal(t, event1.SessionId, event3.SessionId)

		// No.of sessions created so far, should be the same.
		sessionCount, _ = store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)
	})
}

func TestAddSessionDifferentCreationCases(t *testing.T) {
	t.Run("MaxLookbackTimestamp", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(32 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		// This event should not be considered for session
		// creation as it is beyond the max lookback.
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.Empty(t, event1.SessionId)
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		assert.NotEmpty(t, event2.SessionId)
	})

	t.Run("StartingWithMarketingProperty", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		event3, _ := store.GetStore().GetEvent(project.ID, userId, eventId3)
		assert.Equal(t, event1.SessionId, event2.SessionId)
		assert.Equal(t, event1.SessionId, event3.SessionId)
	})

	t.Run("ContinuingSessionCreatedWithMarketingProperty", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			EventProperties: U.PropertiesMap{
				U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
			},
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions for user should be 1.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created. Session continued.
		sessionCount, _ = store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		event3, _ := store.GetStore().GetEvent(project.ID, userId, eventId3)
		assert.Equal(t, event1.SessionId, event2.SessionId)
		assert.Equal(t, event2.SessionId, event3.SessionId)
	})

	t.Run("ContinuingSessionCreatedWithOneEvent", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions for user should be 1.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created. Session continued.
		sessionCount, _ = store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		assert.Equal(t, event1.SessionId, event2.SessionId)
	})

	t.Run("ContinuingSessionButFirstEventWithMarketingProperty", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
			EventProperties: U.PropertiesMap{
				U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
			},
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created.
		sessionCount, _ = store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(2), sessionCount)

		// Check session association.
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		assert.NotEmpty(t, event2.SessionId)
		assert.NotEqual(t, event1.SessionId, event2.SessionId)
	})

	t.Run("ContinuingSessionButFirstEventWithInactivity", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)

		timestamp = timestamp + (32 * 60) + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created.
		sessionCount, _ = store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(2), sessionCount)

		// Check session association.
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		assert.NotEmpty(t, event2.SessionId)
		assert.NotEqual(t, event1.SessionId, event2.SessionId)
	})

	t.Run("MarketingPropertyInTheMiddle", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 1
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = timestamp + 5
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
			EventProperties: U.PropertiesMap{
				U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
			},
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 6
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// Check no.of sessions created for user so far.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(2), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		event3, _ := store.GetStore().GetEvent(project.ID, userId, eventId3)
		assert.NotEmpty(t, event2.SessionId)
		assert.NotEqual(t, event1.SessionId, event2.SessionId)
		assert.Equal(t, event2.SessionId, event3.SessionId)
	})

	t.Run("InactivityImmediatelyAfterMarketingProperty", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
			EventProperties: U.PropertiesMap{
				U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
			},
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		// inactivity.
		timestamp = timestamp + (31 * 60) + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId4 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// Check no.of sessions created for user so far.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(3), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)

		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		assert.NotEmpty(t, event2.SessionId)
		assert.NotEqual(t, event1.SessionId, event2.SessionId)

		event3, _ := store.GetStore().GetEvent(project.ID, userId, eventId3)
		assert.NotEmpty(t, event3.SessionId)
		assert.NotEqual(t, event1.SessionId, event3.SessionId)
		assert.NotEqual(t, event2.SessionId, event3.SessionId)
		event4, _ := store.GetStore().GetEvent(project.ID, userId, eventId4)
		assert.Equal(t, event3.SessionId, event4.SessionId)
	})

	t.Run("InactivityImmediatelyAfterFirstEvent", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		// inactivity.
		timestamp = timestamp + (31 * 60) + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// Check no.of sessions created for user so far.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(2), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)

		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		assert.NotEmpty(t, event2.SessionId)
		assert.NotEqual(t, event1.SessionId, event2.SessionId)
		event3, _ := store.GetStore().GetEvent(project.ID, userId, eventId3)
		assert.Equal(t, event2.SessionId, event3.SessionId)
	})

	t.Run("SingleEventWithMarketingProperty", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			EventProperties: U.PropertiesMap{
				U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
			},
		}
		status, response := SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// Check no.of sessions created for user so far.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)
	})

	t.Run("LastEventWithMarketingProperty", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			EventProperties: U.PropertiesMap{
				U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
			},
			UserId: userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		assert.Equal(t, event1.SessionId, event2.SessionId)
		// New session should be created for last event and associated.
		event3, _ := store.GetStore().GetEvent(project.ID, userId, eventId3)
		assert.NotEmpty(t, event3.SessionId)
		assert.NotEqual(t, event2.SessionId, event3.SessionId)

		// Check no.of sessions created for user so far.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(2), sessionCount)
	})

	t.Run("ContinuingSessionCreatedWithLastEventQualifyForNewSession", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			EventProperties: U.PropertiesMap{
				U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
			},
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions for user should be 1.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)

		timestamp = timestamp + (32 * 60) + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created. Session continued.
		sessionCount, _ = store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(2), sessionCount)

		// Check session association.
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		assert.NotEqual(t, *event1.SessionId, *event2.SessionId)
	})

	t.Run("PageViewEventsWithUserCreatedEvents", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		// User created event with a marketing property.
		// Should create a new session.
		timestamp = timestamp + 2
		randomEventName = U.RandomLowerAphaNumString(5)
		trackPayload2 := SDK.TrackPayload{
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
			EventProperties: U.PropertiesMap{
				U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
			},
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		// User created event with inactivity.
		// Should create a new session.
		timestamp = timestamp + (32 * 60) + 2
		randomEventName = U.RandomLowerAphaNumString(5)
		trackPayload2 = SDK.TrackPayload{
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId4 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(3), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		assert.NotEqual(t, event1.SessionId, event2.SessionId)
		event3, _ := store.GetStore().GetEvent(project.ID, userId, eventId3)
		assert.NotEqual(t, event2.SessionId, event3.SessionId)
		event4, _ := store.GetStore().GetEvent(project.ID, userId, eventId4)
		assert.Equal(t, event3.SessionId, event4.SessionId)
	})

	t.Run("StartingWithUserCreatedEventMarketingProperty", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := U.RandomLowerAphaNumString(5)

		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Name:      randomEventName,
			Timestamp: timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:      true,
			Name:      randomEventName,
			Timestamp: timestamp,
			UserId:    userId,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		event3, _ := store.GetStore().GetEvent(project.ID, userId, eventId3)
		assert.Equal(t, event1.SessionId, event2.SessionId)
		assert.Equal(t, event1.SessionId, event3.SessionId)
	})

	t.Run("ContinuingSessionCheckingTotalSpentTime", func(t *testing.T) {
		project, _, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

		// Test: New user with one event and one skip_session event.
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		// Updating project timestamp to before events start timestamp.
		errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
		assert.Equal(t, http.StatusAccepted, errCode)
		randomEventName := RandomURL()

		trackEventProperties := U.PropertiesMap{
			U.EP_REFERRER:        "www.google.com",
			U.EP_PAGE_URL:        "https://example.com/1/2/",
			U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
			U.EP_PAGE_SPENT_TIME: 10,
		}
		trackUserProperties := U.PropertiesMap{
			U.UP_OS:         "Mac OSX",
			U.UP_OS_VERSION: "1.23.1",
		}
		timestamp = timestamp + 2
		trackPayload1 := SDK.TrackPayload{
			Auto:            true,
			Name:            randomEventName,
			Timestamp:       timestamp,
			EventProperties: trackEventProperties,
			UserProperties:  trackUserProperties,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions for user should be 1.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		assert.NotEmpty(t, event1.SessionId)

		user, _ := store.GetStore().GetUser(project.ID, userId)
		lastestUserPropertiesMap, err := U.DecodePostgresJsonb(&user.Properties)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), (*lastestUserPropertiesMap)[U.UP_SESSION_COUNT])
		assert.Equal(t, float64(1), (*lastestUserPropertiesMap)[U.UP_PAGE_COUNT])
		assert.Equal(t, float64(10), (*lastestUserPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
		assert.Equal(t, trackUserProperties[U.UP_OS], (*lastestUserPropertiesMap)[U.UP_OS])
		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackEventProperties = U.PropertiesMap{
			U.EP_REFERRER:        "www.google.com",
			U.EP_PAGE_URL:        "https://example.com/1/2/",
			U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
			U.EP_PAGE_SPENT_TIME: 5,
		}
		trackPayload2 := SDK.TrackPayload{
			Auto:            true,
			Name:            randomEventName,
			Timestamp:       timestamp,
			UserId:          userId,
			EventProperties: trackEventProperties,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created. Session continued.
		sessionCount, _ = store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		assert.Equal(t, uint64(1), sessionCount)

		// Check session association.
		event2, _ := store.GetStore().GetEvent(project.ID, userId, eventId2)
		assert.Equal(t, event1.SessionId, event2.SessionId)

		user, _ = store.GetStore().GetUser(project.ID, userId)
		lastestUserPropertiesMap, err = U.DecodePostgresJsonb(&user.Properties)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), (*lastestUserPropertiesMap)[U.UP_SESSION_COUNT])
		assert.Equal(t, float64(2), (*lastestUserPropertiesMap)[U.UP_PAGE_COUNT])
		assert.Equal(t, float64(15), (*lastestUserPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
		assert.Equal(t, trackUserProperties[U.UP_OS], (*lastestUserPropertiesMap)[U.UP_OS])
	})
}

func TestAddSessionCreationBufferTime(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	// Event before session buffer time.
	timestamp := U.UnixTimeBeforeDuration(time.Minute * 35)
	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName := RandomURL()
	trackPayload := SDK.TrackPayload{
		Auto:      true,
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
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId

	timestamp = time.Now().Unix()
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK) // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	// Should not create session for last event timestmap  - 30 mins.
	_, err = TaskSession.AddSession([]uint64{project.ID}, 60, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event, errCode := store.GetStore().GetEventById(project.ID, eventId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, event.SessionId)

	// events within buffer time.
	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Nil(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2)
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

func TestAddSessionMergingEventsOnCommonMarketingProperty(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName := RandomURL()
	trackEventProperties := U.PropertiesMap{
		U.EP_REFERRER:        "www.google.com",
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
	}
	trackUserProperties := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload := SDK.TrackPayload{
		Auto:            true,
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
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
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

	// create new user_properties state, for testing session user_properties addition
	// on latest user_properties, which is not associated to any event.
	userProperties := postgres.Jsonb{RawMessage: json.RawMessage(`{"plan": "enterprise"}`)}
	newUserPropertiesID, _, errCode := store.GetStore().UpdateUserProperties(project.ID, userId, &userProperties, time.Now().Unix())
	user, _ := store.GetStore().GetUser(project.ID, userId)
	assert.NotNil(t, user)
	// new user_properties state should be the user's latest user_property state.
	assert.Equal(t, newUserPropertiesID, user.PropertiesId)

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
	assert.Equal(t, float64(10), (*lsEventProperties1)[U.SP_SPENT_TIME])
	// session event properties added from user properties.
	assert.Equal(t, trackUserProperties[U.UP_OS], (*lsEventProperties1)[U.UP_OS])
	assert.Equal(t, trackUserProperties[U.UP_OS_VERSION], (*lsEventProperties1)[U.UP_OS_VERSION])

	// check session user_properties so far, on both event associated
	// user_property and user's latest user_property.
	event, errCode := store.GetStore().GetEventById(project.ID, eventId)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err := U.DecodePostgresJsonb(event.UserProperties)
	if C.ShouldUseUserPropertiesTableForRead(project.ID) {
		userPropertiesRecord, errCode := store.GetStore().GetUserPropertiesRecord(project.ID, event.UserId, event.UserPropertiesId)
		assert.Equal(t, http.StatusFound, errCode)
		userPropertiesMap, err = U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	}
	assert.Nil(t, err)
	assert.Equal(t, float64(1), (*userPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(1), (*userPropertiesMap)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(10), (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	assert.Equal(t, trackUserProperties[U.UP_OS], (*userPropertiesMap)[U.UP_OS])
	// check latest user_properties state.
	user, _ = store.GetStore().GetUser(project.ID, event.UserId)
	lastestUserPropertiesMap, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.NotEmpty(t, user.Properties)
	if C.IsUserPropertiesTableWriteDeprecated(project.ID) {
		assert.Empty(t, event.UserPropertiesId)
	} else {
		assert.NotEmpty(t, event.UserPropertiesId)
	}
	assert.Equal(t, float64(1), (*lastestUserPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(1), (*lastestUserPropertiesMap)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(10), (*lastestUserPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	assert.Equal(t, trackUserProperties[U.UP_OS], (*lastestUserPropertiesMap)[U.UP_OS])

	// ===========================================================

	// Test: New events without session for existing user with session.
	// Since there is continuous activity, last session should be continued.
	timestamp = timestamp + 1
	randomEventName = RandomURL()
	trackEventProperties1 := U.PropertiesMap{
		U.EP_REFERRER:     "www.yahoo.com",
		U.EP_PAGE_URL:     "https://example1.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example1.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: trackEventProperties1,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId

	timestamp = timestamp + 1
	randomEventName = RandomURL()
	trackEventProperties2 := U.PropertiesMap{
		U.EP_REFERRER:     "www.facebook.com",
		U.EP_PAGE_URL:     "https://example2.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example2.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
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
	randomEventName = RandomURL()
	trackEventProperties3 := U.PropertiesMap{
		U.EP_REFERRER:     "www.bing.com",
		U.EP_PAGE_URL:     "https://example3.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example3.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
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
	randomEventName = RandomURL()
	trackEventProperties4 := U.PropertiesMap{
		U.EP_REFERRER:     "www.hacker.com",
		U.EP_PAGE_URL:     "https://example4.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example4.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: trackEventProperties4,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
	// should have initial event's referrer, before continuing session.
	assert.Equal(t, trackEventProperties[U.EP_REFERRER], (*lsEventProperties1)[U.SP_INITIAL_REFERRER])
	// should have latest event's page_url after continuing session.
	assert.Equal(t, trackEventProperties2[U.EP_PAGE_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_URL])
	assert.Equal(t, trackEventProperties2[U.EP_PAGE_RAW_URL], (*lsEventProperties1)[U.SP_LATEST_PAGE_RAW_URL])
	assert.Equal(t, float64(3), (*lsEventProperties1)[U.SP_PAGE_COUNT])
	// event = 10ms, event1 = 1ms (default), event2 = 1ms (default).
	assert.Equal(t, float64(12), (*lsEventProperties1)[U.SP_SPENT_TIME])

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
	assert.Equal(t, float64(2), (*lsEventProperties2)[U.SP_SPENT_TIME])

	// check session count so far.
	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err = U.DecodePostgresJsonb(event4.UserProperties)
	if C.ShouldUseUserPropertiesTableForRead(project.ID) {
		userPropertiesRecord, errCode := store.GetStore().GetUserPropertiesRecord(project.ID, event4.UserId, event4.UserPropertiesId)
		assert.Equal(t, http.StatusFound, errCode)
		userPropertiesMap, err = U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	}
	assert.Nil(t, err)

	assert.Equal(t, float64(2), (*userPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(3), (*userPropertiesMap)[U.UP_PAGE_COUNT])
	// This is because of two different user property id in the same session
	assert.Equal(t, float64(12), (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])

	// =============================================

	// Test: Create new session for event with marketing property,
	// followed by other events, even though there was continuous
	// activity from previous session.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
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
	randomEventName = RandomURL()
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	// should have created session as campaign property exist.
	sessionEvent3 := assertAssociatedSession(t, project.ID, []string{eventId5, eventId6},
		[]string{}, "Session 3")
	assert.NotEqual(t, sessionEvent2.ID, sessionEvent3.ID)

	// check session count so far.
	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err = U.DecodePostgresJsonb(event6.UserProperties)
	if C.ShouldUseUserPropertiesTableForRead(project.ID) {
		userPropertiesRecord, errCode := store.GetStore().GetUserPropertiesRecord(project.ID, event6.UserId, event6.UserPropertiesId)
		assert.Equal(t, http.StatusFound, errCode)
		userPropertiesMap, err = U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	}
	assert.Nil(t, err)
	assert.Equal(t, float64(3), (*userPropertiesMap)[U.UP_SESSION_COUNT])
	assert.Equal(t, float64(5), (*userPropertiesMap)[U.UP_PAGE_COUNT])
	// This is because of two different user property id in the same session
	assert.Equal(t, float64(14), (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])

	// Test: event with marketing property.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale1",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId7 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale2",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId8 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale3",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	// New session should be created after a new event.
	sessionEvent4 := assertAssociatedSession(t, project.ID, []string{eventId5, eventId6}, []string{}, "Session 4")
	assert.Equal(t, sessionEvent3.ID, sessionEvent4.ID)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event7.SessionId, *event8.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event8.SessionId, *event9.SessionId)

	// ==================================

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}

func TestAddSessionMergingEventsOnDifferentMarketingProperty(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName := RandomURL()
	trackEventProperties := U.PropertiesMap{
		U.EP_REFERRER:        "www.google.com",
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
	}
	trackUserProperties := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload := SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties,
		UserProperties:  trackUserProperties,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	userId := response.UserId

	// no session created.
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	// Test: event with marketing property.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale1",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId7 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale2",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId8 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale3",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event7.SessionId, *event8.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event8.SessionId, *event9.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}

func TestAddSessionMergingEventsOnSameMarketingProperty(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName := RandomURL()

	// Test: event with marketing property.
	timestamp = timestamp + 2
	trackPayload := SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale",
		},
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId7 := response.EventId
	userId := response.UserId

	// Test:  event with same marketing property.
	timestamp = timestamp + 3
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId8 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 4
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.Equal(t, *event7.SessionId, *event8.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.Equal(t, *event8.SessionId, *event9.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}

func TestAddSessionMergingEventsOnVaryingMarketingProperty(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName := RandomURL()

	// Test: event with diff marketing property.
	timestamp = timestamp + 2
	trackPayload := SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale4",
		},
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId
	userId := response.UserId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 3
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale5",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 4
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale6",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	// Test: event with same marketing property.
	timestamp = timestamp + 5
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId7 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 6
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId8 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 7
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)

	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event5.SessionId)

	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event6.SessionId)

	assert.NotEqual(t, *event4.SessionId, *event5.SessionId)
	assert.NotEqual(t, *event5.SessionId, *event6.SessionId)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)

	assert.NotEqual(t, *event6.SessionId, *event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.Equal(t, *event7.SessionId, *event8.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.Equal(t, *event8.SessionId, *event9.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}

func TestAddSessionMergingEventsOnCommonMarketingPropertyInMiddle(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName := RandomURL()

	// Test: event with diff marketing property.
	timestamp = timestamp + 1
	trackPayload := SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale1",
		},
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId
	userId := response.UserId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale2",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 3
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale3",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	// Test: event with same marketing property.
	timestamp = timestamp + 4
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale4",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 5
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale4",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 6
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale4",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	// Test: event with diff marketing property.
	timestamp = timestamp + 7
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale7",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId7 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 8
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale8",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId8 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 9
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale9",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event2.SessionId)

	event3, errCode := store.GetStore().GetEventById(project.ID, eventId3)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event3.SessionId)

	assert.NotEqual(t, *event1.SessionId, *event2.SessionId)
	assert.NotEqual(t, *event2.SessionId, *event3.SessionId)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event3.SessionId, *event4.SessionId)

	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event4.SessionId, *event5.SessionId)

	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event5.SessionId, *event6.SessionId)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event8.SessionId, *event7.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event9.SessionId, *event8.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}

func TestAddSessionMergingEventsOnVaryingMarketingPropertyContinuous(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName := RandomURL()

	// Test: event with diff marketing property.
	timestamp = timestamp + 1
	trackPayload := SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale1",
		},
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId
	userId := response.UserId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale2",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 3
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale3",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	// Test: event with same marketing property.
	timestamp = timestamp + 4
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale4",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 5
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale4",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 6
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale4",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	// Test: event with diff marketing property.
	timestamp = timestamp + 7
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale7",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId7 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 8
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale8",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId8 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 9
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale9",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId

	// Test: event with diff marketing property.
	timestamp = timestamp + 91
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale91",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId91 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 92
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale92",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId92 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 93
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale93",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId93 := response.EventId

	// Test: event with same marketing property.
	timestamp = timestamp + 94
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale94",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId94 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 95
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale94",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId95 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 96
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale94",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId96 := response.EventId

	// Test: event with diff marketing property.
	timestamp = timestamp + 97
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale97",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId97 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 98
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale98",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId98 := response.EventId

	// Test:  event with diff marketing property.
	timestamp = timestamp + 99
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "same_winter_sale99",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId99 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event2.SessionId)

	event3, errCode := store.GetStore().GetEventById(project.ID, eventId3)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event3.SessionId)

	assert.NotEqual(t, *event1.SessionId, *event2.SessionId)
	assert.NotEqual(t, *event2.SessionId, *event3.SessionId)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event3.SessionId, *event4.SessionId)

	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event4.SessionId, *event5.SessionId)

	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event5.SessionId, *event6.SessionId)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event8.SessionId, *event7.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event9.SessionId, *event8.SessionId)

	event91, errCode := store.GetStore().GetEventById(project.ID, eventId91)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event91.SessionId)

	event92, errCode := store.GetStore().GetEventById(project.ID, eventId92)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event92.SessionId)

	event93, errCode := store.GetStore().GetEventById(project.ID, eventId93)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event93.SessionId)

	assert.NotEqual(t, *event91.SessionId, *event92.SessionId)
	assert.NotEqual(t, *event92.SessionId, *event93.SessionId)

	event94, errCode := store.GetStore().GetEventById(project.ID, eventId94)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.NotEqual(t, *event93.SessionId, *event94.SessionId)

	event95, errCode := store.GetStore().GetEventById(project.ID, eventId95)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.Equal(t, *event94.SessionId, *event95.SessionId)

	event96, errCode := store.GetStore().GetEventById(project.ID, eventId96)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.Equal(t, *event95.SessionId, *event96.SessionId)

	event97, errCode := store.GetStore().GetEventById(project.ID, eventId97)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.NotEqual(t, *event96.SessionId, *event97.SessionId)

	event98, errCode := store.GetStore().GetEventById(project.ID, eventId98)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.NotEqual(t, *event98.SessionId, *event97.SessionId)

	event99, errCode := store.GetStore().GetEventById(project.ID, eventId99)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.NotEqual(t, *event99.SessionId, *event98.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}

func TestAddSessionMergingEventsOnMissedMarketingProperty(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName := RandomURL()

	// Test: event with marketing property.
	timestamp = timestamp + 2
	trackPayload := SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "campaign_same_winter_sale",
			U.QUERY_PARAM_UTM_PREFIX + "adgroup":  "adgroup_same_winter_sale",
		},
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId
	userId := response.UserId

	// Test:  event with same marketing property.
	timestamp = timestamp + 3
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "campaign_same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 4
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "adgroup": "adgroup_same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	timestamp = timestamp + 5
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: U.PropertiesMap{},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	timestamp = timestamp + 6
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "campaign_same_winter_sale",
			U.QUERY_PARAM_UTM_PREFIX + "adgroup":  "adgroup_same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event2.SessionId)
	assert.Equal(t, *event2.SessionId, *event1.SessionId)

	event3, errCode := store.GetStore().GetEventById(project.ID, eventId3)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event3.SessionId)
	assert.Equal(t, *event3.SessionId, *event1.SessionId)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event4.SessionId, *event1.SessionId)

	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event5.SessionId)
	assert.Equal(t, *event5.SessionId, *event1.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}

func TestAddSessionMergingEventsOnMissedMarketingPropertyMultiSession(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName := RandomURL()

	// Test: event with marketing property.
	timestamp = timestamp + 2
	trackPayload := SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "campaign_same_winter_sale",
			U.QUERY_PARAM_UTM_PREFIX + "adgroup":  "adgroup_same_winter_sale",
		},
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId
	userId := response.UserId

	// Test:  event with same marketing property.
	timestamp = timestamp + 3
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "campaign_same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 4
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "adgroup": "adgroup_same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	timestamp = timestamp + 5
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: U.PropertiesMap{},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	timestamp = timestamp + 6
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "campaign_same_winter_sale",
			U.QUERY_PARAM_UTM_PREFIX + "adgroup":  "adgroup_same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	timestamp = timestamp + 6
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "campaign_same_winter_sale",
			U.QUERY_PARAM_UTM_PREFIX + "adgroup":  "adgroup_same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event2.SessionId)
	assert.Equal(t, *event2.SessionId, *event1.SessionId)

	event3, errCode := store.GetStore().GetEventById(project.ID, eventId3)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event3.SessionId)
	assert.Equal(t, *event3.SessionId, *event1.SessionId)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event4.SessionId, *event1.SessionId)

	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event5.SessionId)
	assert.Equal(t, *event5.SessionId, *event1.SessionId)

	// since user id different, should create a new session for this
	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event6.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event5.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}

func TestAddSessionMergingEventsOnMissedMarketingPropertyMultiSessionEmptyProperty(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName := RandomURL()

	// Test: event with marketing property.
	timestamp = timestamp + 2
	trackPayload := SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "campaign_same_winter_sale",
			U.QUERY_PARAM_UTM_PREFIX + "adgroup":  "adgroup_same_winter_sale",
		},
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId
	userId := response.UserId

	// Test:  event with same marketing property.
	timestamp = timestamp + 3
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "campaign_same_winter_sale",
			U.QUERY_PARAM_UTM_PREFIX + "adgroup":  "",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	// Test:  event with same marketing property.
	timestamp = timestamp + 4
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "",
			U.QUERY_PARAM_UTM_PREFIX + "adgroup":  "adgroup_same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	timestamp = timestamp + 5
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "keyword": "",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	timestamp = timestamp + 6
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "campaign_same_winter_sale",
			U.QUERY_PARAM_UTM_PREFIX + "adgroup":  "adgroup_same_winter_sale",
			U.QUERY_PARAM_UTM_PREFIX + "keyword":  "",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	timestamp = timestamp + 6
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "campaign_same_winter_sale",
			U.QUERY_PARAM_UTM_PREFIX + "adgroup":  "adgroup_same_winter_sale",
		},
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	_, err = TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event2.SessionId)
	assert.Equal(t, *event2.SessionId, *event1.SessionId)

	event3, errCode := store.GetStore().GetEventById(project.ID, eventId3)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event3.SessionId)
	assert.Equal(t, *event3.SessionId, *event1.SessionId)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event4.SessionId, *event1.SessionId)

	//
	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event5.SessionId)
	assert.Equal(t, *event5.SessionId, *event1.SessionId)

	// since user id different, should create a new session for this
	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event6.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event5.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]uint64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}
