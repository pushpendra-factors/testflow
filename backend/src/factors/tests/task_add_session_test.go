package tests

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"

	SDK "factors/sdk"
	TaskSession "factors/task/session"
	U "factors/util"
)

func assertAssociatedSession(t *testing.T, projectId int64, eventIdsInOrder []string,
	skipSessionEventIds []string, message string) (sessionEvent *model.Event) {

	var firstEvent *model.Event
	for i, eventId := range eventIdsInOrder {
		event, errCode := store.GetStore().GetEventById(projectId, eventId, "")
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
	sessionEvent, errCode := store.GetStore().GetEventById(projectId, *firstEvent.SessionId, firstEvent.UserId)
	assert.Equal(t, http.StatusFound, errCode, message)
	assert.Equal(t, firstEvent.Timestamp-1, sessionEvent.Timestamp, message)

	return sessionEvent
}

func TestAddSessionLatestUserProperties(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	timestamp := time.Now().AddDate(0, 0, -1)
	timestampUnix := timestamp.Unix()

	/*
		Test flow
		Session event - t1 -> Non session event - t2 -> -> Non session event - t3 -> Add session job
		Expected
		Event on t3 properties should be on latest user properties even after session creation for old timestamp
	*/
	randomEventName := RandomURL()
	trackEventProperties := U.PropertiesMap{
		U.EP_PAGE_URL:     "https://example.com",
		U.EP_PAGE_RAW_URL: "https://example.com",
		U.EP_CAMPAIGN_ID:  "124",
	}
	trackUserProperties := U.PropertiesMap{
		U.UP_OS: "android1",
	}
	trackPayload := SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestampUnix,
		EventProperties: trackEventProperties,
		UserProperties:  trackUserProperties,
		RequestSource:   model.UserSourceWeb,
	}
	status, res := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	userID := res.UserId
	// session not created.
	_, errCode := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	// skip session event
	trackEventProperties = U.PropertiesMap{
		U.EP_PAGE_URL:     "https://example.com/1",
		U.EP_PAGE_RAW_URL: "https://example.com/1?x=1",
		U.EP_CAMPAIGN_ID:  "123456",
	}
	trackUserProperties = U.PropertiesMap{
		U.UP_OS: "android2",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestampUnix + 100000,
		EventProperties: trackEventProperties,
		UserProperties:  trackUserProperties,
		UserId:          userID,
		RequestSource:   model.UserSourceWeb,
	}
	status, _ = SDK.Track(project.ID, &trackPayload, true, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	user, status := store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, status)
	properitesMap := make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	assert.Equal(t, "123456", properitesMap[U.UP_LATEST_CAMPAIGN_ID])
	assert.Equal(t, "android2", properitesMap[U.UP_OS])

	// skip session event
	trackEventProperties = U.PropertiesMap{
		U.EP_PAGE_URL:     "https://example.com/2",
		U.EP_PAGE_RAW_URL: "https://example.com/2?x=1",
		U.EP_CAMPAIGN_ID:  "1234567",
	}
	trackUserProperties = U.PropertiesMap{
		U.UP_OS: "android3",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestampUnix + 200000,
		EventProperties: trackEventProperties,
		UserProperties:  trackUserProperties,
		UserId:          userID,
		RequestSource:   model.UserSourceWeb,
	}
	status, _ = SDK.Track(project.ID, &trackPayload, true, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	user, status = store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, status)
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	assert.Equal(t, "1234567", properitesMap[U.UP_LATEST_CAMPAIGN_ID])
	assert.Equal(t, "android3", properitesMap[U.UP_OS])

	_, err = TaskSession.AddSession([]int64{project.ID}, 2*24*60*60, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	// session created.
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusFound, errCode)

	user, status = store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, status)

	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	assert.Equal(t, "1234567", properitesMap[U.UP_LATEST_CAMPAIGN_ID])
	assert.Equal(t, "android3", properitesMap[U.UP_OS])
}

func TestAddSessionWithChannelGroup(t *testing.T) {
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
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_REFERRER:        "",
		U.EP_REFERRER_DOMAIN: "",
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId := response.EventId
	userID := response.UserId

	randomEventName = RandomURL()
	trackEventProperties0 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_REFERRER:        "",
		U.EP_REFERRER_DOMAIN: "",
	}
	trackUserProperties0 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		UserId:          userID,
		Name:            randomEventName,
		Timestamp:       timestamp + 1,
		EventProperties: trackEventProperties0,
		UserProperties:  trackUserProperties0,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	// assert.NotEmpty(t, response.UserId)

	// no session created.
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	user, status := store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, status)

	propertiesMap := make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &propertiesMap)
	assert.Nil(t, err)
	assert.Equal(t, propertiesMap[U.UP_INITIAL_CHANNEL], model.ChannelDirect)
	assert.Equal(t, propertiesMap[U.UP_LATEST_CHANNEL], model.ChannelDirect)

	sessionEvent1 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 1")
	// session event properties added from event properties.
	lsEventProperties1, err := U.DecodePostgresJsonb(&sessionEvent1.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties1)[U.EP_CHANNEL], model.ChannelDirect)
	lsUserProperties1, err := U.DecodePostgresJsonb(sessionEvent1.UserProperties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsUserProperties1)[U.UP_INITIAL_CHANNEL], model.ChannelDirect)
	assert.Equal(t, (*lsUserProperties1)[U.UP_LATEST_CHANNEL], model.ChannelDirect)

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties1 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_GCLID:           "xyz1231",
	}
	trackUserProperties1 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		UserId:          userID,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties1,
		UserProperties:  trackUserProperties1,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	// assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	user, status = store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, status)

	propertiesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &propertiesMap)
	assert.Nil(t, err)
	assert.Equal(t, propertiesMap[U.UP_INITIAL_CHANNEL], model.ChannelDirect)
	assert.Equal(t, propertiesMap[U.UP_LATEST_CHANNEL], model.ChannelGoogleAdsNetwork)

	sessionEvent2 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 2")
	// session event properties added from event properties.
	lsEventProperties2, err := U.DecodePostgresJsonb(&sessionEvent2.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties2)[U.EP_CHANNEL], model.ChannelGoogleAdsNetwork)
	lsUserProperties2, err := U.DecodePostgresJsonb(sessionEvent2.UserProperties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsUserProperties2)[U.UP_INITIAL_CHANNEL], model.ChannelDirect)
	assert.Equal(t, (*lsUserProperties2)[U.UP_LATEST_CHANNEL], model.ChannelGoogleAdsNetwork)

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties2 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_SOURCE:          "google",
		U.EP_MEDIUM:          "paid",
	}
	trackUserProperties2 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		UserId:          userID,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties2,
		UserProperties:  trackUserProperties2,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent3 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 3")
	// session event properties added from event properties.
	lsEventProperties3, err := U.DecodePostgresJsonb(&sessionEvent3.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties3)[U.EP_CHANNEL], model.ChannelPaidSearch)
	lsUserProperties3, err := U.DecodePostgresJsonb(sessionEvent3.UserProperties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsUserProperties3)[U.UP_INITIAL_CHANNEL], model.ChannelDirect)
	assert.Equal(t, (*lsUserProperties3)[U.UP_LATEST_CHANNEL], model.ChannelPaidSearch)

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties3 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_REFERRER_DOMAIN: "bing.com",
	}
	trackUserProperties3 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		UserId:          userID,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties3,
		UserProperties:  trackUserProperties3,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent4 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 4")
	// session event properties added from event properties.
	lsEventProperties4, err := U.DecodePostgresJsonb(&sessionEvent4.Properties)
	assert.Nil(t, err)
	assert.Equal(t, model.ChannelOrganicSearch, (*lsEventProperties4)[U.EP_CHANNEL])

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties4 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_SOURCE:          "LinkedIn",
		U.EP_MEDIUM:          "paid_social",
	}
	trackUserProperties4 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		UserId:          userID,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties4,
		UserProperties:  trackUserProperties4,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent5 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 5")
	// session event properties added from event properties.
	lsEventProperties5, err := U.DecodePostgresJsonb(&sessionEvent5.Properties)
	assert.Nil(t, err)
	assert.Equal(t, model.ChannelPaidSocial, (*lsEventProperties5)[U.EP_CHANNEL])
	lsUserProperties5, err := U.DecodePostgresJsonb(sessionEvent5.UserProperties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsUserProperties5)[U.UP_INITIAL_CHANNEL], model.ChannelDirect)
	assert.Equal(t, (*lsUserProperties5)[U.UP_LATEST_CHANNEL], model.ChannelPaidSocial)

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties5 := U.PropertiesMap{
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_DOMAIN:     "example.com",
		U.EP_SOURCE:          "Linkedin",
		U.EP_MEDIUM:          "paid_social",
	}
	trackUserProperties5 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties5,
		UserProperties:  trackUserProperties5,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent6 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 6")
	// session event properties added from event properties.
	lsEventProperties6, err := U.DecodePostgresJsonb(&sessionEvent6.Properties)
	assert.Nil(t, err)
	assert.Equal(t, model.ChannelPaidSocial, (*lsEventProperties6)[U.EP_CHANNEL])

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties6 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_SOURCE:          "facebook",
		U.EP_MEDIUM:          "paidsocial",
	}
	trackUserProperties6 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties6,
		UserProperties:  trackUserProperties6,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent7 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 7")
	// session event properties added from event properties.
	lsEventProperties7, err := U.DecodePostgresJsonb(&sessionEvent7.Properties)
	assert.Nil(t, err)
	assert.Equal(t, model.ChannelPaidSocial, (*lsEventProperties7)[U.EP_CHANNEL])

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties7 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_REFERRER_DOMAIN: "facebook.com",
		U.EP_MEDIUM:          "paid",
	}
	trackUserProperties7 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties7,
		UserProperties:  trackUserProperties7,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent8 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 8")
	// session event properties added from event properties.
	lsEventProperties8, err := U.DecodePostgresJsonb(&sessionEvent8.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties8)[U.EP_CHANNEL], model.ChannelPaidSocial)

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties8 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_REFERRER_DOMAIN: "facebook.com",
		U.EP_MEDIUM:          "something",
	}
	trackUserProperties8 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties8,
		UserProperties:  trackUserProperties8,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent9 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 9")
	// session event properties added from event properties.
	lsEventProperties9, err := U.DecodePostgresJsonb(&sessionEvent9.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties9)[U.EP_CHANNEL], model.ChannelOrganicSocial)

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties9 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_REFERRER_DOMAIN: "google.com",
	}
	trackUserProperties9 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties9,
		UserProperties:  trackUserProperties9,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent10 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 10")
	// session event properties added from event properties.
	lsEventProperties10, err := U.DecodePostgresJsonb(&sessionEvent10.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties10)[U.EP_CHANNEL], model.ChannelOrganicSearch)

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties10 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_SOURCE:          "email",
	}
	trackUserProperties10 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties10,
		UserProperties:  trackUserProperties10,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent11 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 11")
	// session event properties added from event properties.
	lsEventProperties11, err := U.DecodePostgresJsonb(&sessionEvent11.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties11)[U.EP_CHANNEL], model.ChannelEmail)

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties11 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_MEDIUM:          "affiliate",
	}
	trackUserProperties11 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties11,
		UserProperties:  trackUserProperties11,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent12 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 12")
	// session event properties added from event properties.
	lsEventProperties12, err := U.DecodePostgresJsonb(&sessionEvent12.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties12)[U.EP_CHANNEL], model.ChannelAffiliate)

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties12 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_REFERRER_DOMAIN: "abc.com",
	}
	trackUserProperties12 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties12,
		UserProperties:  trackUserProperties12,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent13 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 13")
	// session event properties added from event properties.
	lsEventProperties13, err := U.DecodePostgresJsonb(&sessionEvent13.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties13)[U.EP_CHANNEL], model.ChannelReferral)

	timestamp = timestamp + 2000
	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties13 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_REFERRER_DOMAIN: "www.linkedin.com",
	}
	trackUserProperties13 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties13,
		UserProperties:  trackUserProperties13,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent14 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 15")
	// session event properties added from event properties.
	lsEventProperties14, err := U.DecodePostgresJsonb(&sessionEvent14.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties14)[U.EP_CHANNEL], model.ChannelOrganicSocial)

	timestamp = timestamp + 2000

	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties14 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_PAGE_DOMAIN:     "www.example.com",
		U.EP_REFERRER_DOMAIN: "www.example.com",
	}
	trackUserProperties14 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties14,
		UserProperties:  trackUserProperties14,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent15 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 15")
	// session event properties added from event properties.
	lsEventProperties15, err := U.DecodePostgresJsonb(&sessionEvent15.Properties)
	assert.Nil(t, err)
	assert.Equal(t, model.ChannelDirect, (*lsEventProperties15)[U.EP_CHANNEL])

	// Updating project timestamp to before events start timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)
	randomEventName = RandomURL()
	trackEventProperties15 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_PAGE_DOMAIN:     "www.example.com",
		U.EP_REFERRER_DOMAIN: "www.example.com",
		U.EP_LICLID:          "yyy123",
	}
	trackUserProperties15 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties15,
		UserProperties:  trackUserProperties15,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId = response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	sessionEvent16 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 16")
	// session event properties added from event properties.
	lsEventProperties16, err := U.DecodePostgresJsonb(&sessionEvent16.Properties)
	assert.Nil(t, err)
	assert.Equal(t, model.ChannelOthers, (*lsEventProperties16)[U.EP_CHANNEL])
}

func TestMultipleEventsWithSingleAddSessionCallWithChannelGroup(t *testing.T) {
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
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_REFERRER:        "",
		U.EP_REFERRER_DOMAIN: "",
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId := response.EventId
	userID := response.UserId

	randomEventName = RandomURL()
	trackEventProperties0 := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_GCLID:           "xyz1231",
	}
	trackUserProperties0 := U.PropertiesMap{
		U.UP_OS:         "Mac OSX",
		U.UP_OS_VERSION: "1.23.1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		UserId:          userID,
		Name:            randomEventName,
		Timestamp:       timestamp + 2000,
		EventProperties: trackEventProperties0,
		UserProperties:  trackUserProperties0,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	// assert.NotEmpty(t, response.UserId)
	eventId0 := response.EventId

	// no session created.
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	user, status := store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, status)

	propertiesMap := make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &propertiesMap)
	assert.Nil(t, err)
	assert.Equal(t, propertiesMap[U.UP_INITIAL_CHANNEL], model.ChannelDirect)
	assert.Equal(t, propertiesMap[U.UP_LATEST_CHANNEL], model.ChannelGoogleAdsNetwork)

	sessionEvent1 := assertAssociatedSession(t, project.ID, []string{eventId},
		[]string{}, "Session 1")
	sessionEvent2 := assertAssociatedSession(t, project.ID, []string{eventId0},
		[]string{}, "Session 2")
	// session event properties added from event properties.
	lsEventProperties1, err := U.DecodePostgresJsonb(&sessionEvent1.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties1)[U.EP_CHANNEL], model.ChannelDirect)
	lsUserProperties1, err := U.DecodePostgresJsonb(sessionEvent1.UserProperties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsUserProperties1)[U.UP_INITIAL_CHANNEL], model.ChannelDirect)
	assert.Equal(t, (*lsUserProperties1)[U.UP_LATEST_CHANNEL], model.ChannelDirect)

	lsUserProperties2, err := U.DecodePostgresJsonb(sessionEvent2.UserProperties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsUserProperties2)[U.UP_INITIAL_CHANNEL], model.ChannelDirect)
	assert.Equal(t, (*lsUserProperties2)[U.UP_LATEST_CHANNEL], model.ChannelGoogleAdsNetwork)
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
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=124",
		U.EP_PAGE_SPENT_TIME: 10,
		U.EP_TERM:            "term1",
		U.EP_KEYWORD:         "keyword1",
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId := response.EventId
	userId := response.UserId

	// no session created.
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Name:          "non_web_event",
		Timestamp:     timestamp,
		UserId:        userId,
		RequestSource: model.UserSourceWeb,
	}
	// skip session event.
	status, response = SDK.Track(project.ID, &trackPayload, true, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	skipSessionEventId := response.EventId

	// create new user_properties state, for testing session user_properties addition
	// on latest user_properties, which is not associated to any event.
	userProperties := postgres.Jsonb{RawMessage: json.RawMessage(`{"plan": "enterprise"}`)}
	userProperties1, errCode := store.GetStore().UpdateUserProperties(project.ID, userId, &userProperties, time.Now().Unix())
	user, _ := store.GetStore().GetUser(project.ID, userId)
	assert.NotNil(t, user)
	// new user_properties state should be the user's latest user_property state.
	properties1, _ := U.DecodePostgresJsonb(userProperties1)
	properties, _ := U.DecodePostgresJsonb(&user.Properties)
	assert.Equal(t, properties1, properties)

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
	event, errCode := store.GetStore().GetEventById(project.ID, eventId, "")
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err := U.DecodePostgresJsonb(event.UserProperties)
	assert.Nil(t, err)
	assert.Nil(t, (*userPropertiesMap)[U.UP_PAGE_COUNT])
	assert.Nil(t, (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	assert.Equal(t, trackUserProperties[U.UP_OS], (*userPropertiesMap)[U.UP_OS])
	// check latest user_properties state.
	user, _ = store.GetStore().GetUser(project.ID, event.UserId)
	lastestUserPropertiesMap, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.NotEmpty(t, user.Properties)
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
		U.EP_PAGE_RAW_URL: "https://example1.com/1/2?x=123",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: trackEventProperties1,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Name:          "non_web_event",
		Timestamp:     timestamp,
		UserId:        userId,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, true, SDK.SourceJSSDK, "") // skip session.
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
	assert.Equal(t, "term1", (*lsEventProperties1)[U.EP_TERM])
	assert.Equal(t, "keyword1", (*lsEventProperties1)[U.EP_KEYWORD])

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
	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4, "")
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err = U.DecodePostgresJsonb(event4.UserProperties)
	assert.Nil(t, err)

	assert.Nil(t, (*userPropertiesMap)[U.UP_PAGE_COUNT])
	// This is because of two different user property id in the same session
	assert.Nil(t, (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])

	user, _ = store.GetStore().GetUser(project.ID, event.UserId)
	lastestUserPropertiesMap, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.NotEmpty(t, user.Properties)
	assert.NotNil(t, (*lastestUserPropertiesMap)[U.UP_PAGE_COUNT])
	assert.NotNil(t, (*lastestUserPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	assert.Nil(t, (*lastestUserPropertiesMap)[U.EP_SESSION_COUNT])

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
		RequestSource: model.UserSourceWeb,
	}

	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	timestamp = timestamp + 2
	randomEventName = RandomURL()
	trackPayload = SDK.TrackPayload{
		Auto:          true,
		Name:          randomEventName,
		Timestamp:     timestamp,
		UserId:        userId,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	// should have created session as campaign property exist.
	sessionEvent3 := assertAssociatedSession(t, project.ID, []string{eventId5, eventId6},
		[]string{}, "Session 3")
	assert.NotEqual(t, sessionEvent2.ID, sessionEvent3.ID)

	// check session count so far.
	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6, "")
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err = U.DecodePostgresJsonb(event6.UserProperties)
	assert.Nil(t, err)
	assert.Nil(t, (*userPropertiesMap)[U.UP_PAGE_COUNT])
	// This is because of two different user property id in the same session
	assert.Nil(t, (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])

	user, _ = store.GetStore().GetUser(project.ID, event.UserId)
	lastestUserPropertiesMap, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.NotEmpty(t, user.Properties)
	assert.NotNil(t, (*lastestUserPropertiesMap)[U.UP_PAGE_COUNT])
	assert.NotNil(t, (*lastestUserPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	assert.Nil(t, (*lastestUserPropertiesMap)[U.EP_SESSION_COUNT])

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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId7 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	// New session should be created after a new event.
	//sessionEvent4 := assertAssociatedSession(t, project.ID, []string{eventId5, eventId6}, []string{}, "Session 4")
	//assert.Equal(t, sessionEvent3.ID, sessionEvent4.ID)

	// Last event with marketing property should be process on next run of add session,
	// to avoid associating previous session.
	statusMap, err := TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event7.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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

		// v1 - This event should not be considered for session creation as it is beyond the max lookback.
		// v2 - This event should be considered for session creation as it is independent of max lookback.
		trackPayload1 := SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
		assert.Nil(t, err)

		// No.of sessions created.
		sessionEventName, _ := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
		sessionCount, _ := store.GetStore().GetEventCountOfUserByEventName(project.ID, userId, sessionEventName.ID)
		if C.EnableUserLevelEventPullForAddSessionByProjectID(project.ID) {
			assert.Equal(t, uint64(2), sessionCount)
		} else {
			assert.Equal(t, uint64(1), sessionCount)
		}

		// Check session association.
		event1, _ := store.GetStore().GetEvent(project.ID, userId, eventId1)
		if C.EnableUserLevelEventPullForAddSessionByProjectID(project.ID) {
			assert.NotEmpty(t, event1.SessionId)
		} else {
			assert.Empty(t, event1.SessionId)
		}

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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
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
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 6
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
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
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		// inactivity.
		timestamp = timestamp + (31 * 60) + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId4 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		// inactivity.
		timestamp = timestamp + (31 * 60) + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
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
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
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
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		// User created event with inactivity.
		// Should create a new session.
		timestamp = timestamp + (32 * 60) + 2
		randomEventName = U.RandomLowerAphaNumString(5)
		trackPayload2 = SDK.TrackPayload{
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId4 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			Name:          randomEventName,
			Timestamp:     timestamp,
			RequestSource: model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 := SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		timestamp = timestamp + 2
		randomEventName = RandomURL()
		trackPayload2 = SDK.TrackPayload{
			Auto:          true,
			Name:          randomEventName,
			Timestamp:     timestamp,
			UserId:        userId,
			RequestSource: model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId3 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			RequestSource:   model.UserSourceWeb,
		}
		status, response := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		userId := response.UserId
		eventId1 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
			RequestSource:   model.UserSourceWeb,
		}
		status, response = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
		assert.Equal(t, http.StatusOK, status)
		eventId2 := response.EventId

		_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
		Auto:          true,
		Name:          randomEventName,
		Timestamp:     timestamp,
		RequestSource: model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "") // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId := response.EventId
	userId := response.UserId

	timestamp = U.UnixTimeBeforeDuration(time.Minute * 15)
	trackPayload = SDK.TrackPayload{
		Auto:          true,
		Name:          randomEventName,
		Timestamp:     timestamp,
		UserId:        userId,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "") // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId1 := response.EventId

	timestamp = time.Now().Unix()
	trackPayload = SDK.TrackPayload{
		Auto:          true,
		Name:          randomEventName,
		Timestamp:     timestamp,
		UserId:        userId,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "") // true: skips session.
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	// Should not create session for last event timestmap  - 30 mins.
	_, err = TaskSession.AddSession([]int64{project.ID}, 60, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event, errCode := store.GetStore().GetEventById(project.ID, eventId, "")
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, event.SessionId)

	// events within buffer time.
	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1, "")
	assert.Equal(t, http.StatusFound, errCode)
	assert.Nil(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2, "")
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId := response.EventId
	userId := response.UserId

	// no session created.
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Name:          "non_web_event",
		Timestamp:     timestamp,
		UserId:        userId,
		RequestSource: model.UserSourceWeb,
	}
	// skip session event.
	status, response = SDK.Track(project.ID, &trackPayload, true, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	skipSessionEventId := response.EventId

	// create new user_properties state, for testing session user_properties addition
	// on latest user_properties, which is not associated to any event.
	userProperties := postgres.Jsonb{RawMessage: json.RawMessage(`{"plan": "enterprise"}`)}
	newUserProperties, errCode := store.GetStore().UpdateUserProperties(project.ID, userId, &userProperties, time.Now().Unix())
	user, _ := store.GetStore().GetUser(project.ID, userId)
	assert.NotNil(t, user)
	// new user_properties state should be the user's latest user_property state.
	assert.Equal(t, DecodePostgresJsonbWithoutError(newUserProperties),
		DecodePostgresJsonbWithoutError(&user.Properties))

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
	event, errCode := store.GetStore().GetEventById(project.ID, eventId, "")
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err := U.DecodePostgresJsonb(event.UserProperties)
	assert.Nil(t, err)
	assert.Nil(t, (*userPropertiesMap)[U.UP_PAGE_COUNT])
	assert.Nil(t, (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	assert.Equal(t, trackUserProperties[U.UP_OS], (*userPropertiesMap)[U.UP_OS])
	// check latest user_properties state.
	user, _ = store.GetStore().GetUser(project.ID, event.UserId)
	lastestUserPropertiesMap, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.NotEmpty(t, user.Properties)
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Name:          "non_web_event",
		Timestamp:     timestamp,
		UserId:        userId,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, true, SDK.SourceJSSDK, "") // skip session.
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4, "")
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err = U.DecodePostgresJsonb(event4.UserProperties)
	assert.Nil(t, err)

	assert.Nil(t, (*userPropertiesMap)[U.UP_PAGE_COUNT])
	// This is because of two different user property id in the same session
	assert.Nil(t, (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])

	user, _ = store.GetStore().GetUser(project.ID, event.UserId)
	lastestUserPropertiesMap, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.NotEmpty(t, user.Properties)
	assert.NotNil(t, (*lastestUserPropertiesMap)[U.UP_PAGE_COUNT])
	assert.NotNil(t, (*lastestUserPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	assert.Nil(t, (*lastestUserPropertiesMap)[U.EP_SESSION_COUNT])

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
		RequestSource: model.UserSourceWeb,
	}

	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	timestamp = timestamp + 2
	randomEventName = RandomURL()
	trackPayload = SDK.TrackPayload{
		Auto:          true,
		Name:          randomEventName,
		Timestamp:     timestamp,
		UserId:        userId,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	// should have created session as campaign property exist.
	sessionEvent3 := assertAssociatedSession(t, project.ID, []string{eventId5, eventId6},
		[]string{}, "Session 3")
	assert.NotEqual(t, sessionEvent2.ID, sessionEvent3.ID)

	// check session count so far.
	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6, "")
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesMap, err = U.DecodePostgresJsonb(event6.UserProperties)
	assert.Nil(t, err)
	assert.Nil(t, (*userPropertiesMap)[U.UP_PAGE_COUNT])
	// This is because of two different user property id in the same session
	assert.Nil(t, (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME])

	user, _ = store.GetStore().GetUser(project.ID, event.UserId)
	lastestUserPropertiesMap, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.NotEmpty(t, user.Properties)
	assert.NotNil(t, (*lastestUserPropertiesMap)[U.UP_PAGE_COUNT])
	assert.NotNil(t, (*lastestUserPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	assert.Nil(t, (*lastestUserPropertiesMap)[U.EP_SESSION_COUNT])

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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	// New session should be created after a new event.
	sessionEvent4 := assertAssociatedSession(t, project.ID, []string{eventId5, eventId6}, []string{}, "Session 4")
	assert.Equal(t, sessionEvent3.ID, sessionEvent4.ID)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event7.SessionId, *event8.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event8.SessionId, *event9.SessionId)

	// ==================================

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
		RequestSource:   model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event7.SessionId, *event8.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.NotEqual(t, *event8.SessionId, *event9.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
		RequestSource: model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.Equal(t, *event7.SessionId, *event8.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.Equal(t, *event8.SessionId, *event9.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
		RequestSource: model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)

	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event5.SessionId)

	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event6.SessionId)

	assert.NotEqual(t, *event4.SessionId, *event5.SessionId)
	assert.NotEqual(t, *event5.SessionId, *event6.SessionId)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)

	assert.NotEqual(t, *event6.SessionId, *event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.Equal(t, *event7.SessionId, *event8.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event7.SessionId)
	assert.Equal(t, *event8.SessionId, *event9.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
		RequestSource: model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event2.SessionId)

	event3, errCode := store.GetStore().GetEventById(project.ID, eventId3, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event3.SessionId)

	assert.NotEqual(t, *event1.SessionId, *event2.SessionId)
	assert.NotEqual(t, *event2.SessionId, *event3.SessionId)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event3.SessionId, *event4.SessionId)

	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event4.SessionId, *event5.SessionId)

	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event5.SessionId, *event6.SessionId)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event8.SessionId, *event7.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event9.SessionId, *event8.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
		RequestSource: model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId99 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event2.SessionId)

	event3, errCode := store.GetStore().GetEventById(project.ID, eventId3, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event3.SessionId)

	assert.NotEqual(t, *event1.SessionId, *event2.SessionId)
	assert.NotEqual(t, *event2.SessionId, *event3.SessionId)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event3.SessionId, *event4.SessionId)

	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event4.SessionId, *event5.SessionId)

	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event5.SessionId, *event6.SessionId)

	event7, errCode := store.GetStore().GetEventById(project.ID, eventId7, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event7.SessionId)

	event8, errCode := store.GetStore().GetEventById(project.ID, eventId8, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event8.SessionId, *event7.SessionId)

	event9, errCode := store.GetStore().GetEventById(project.ID, eventId9, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.NotEqual(t, *event9.SessionId, *event8.SessionId)

	event91, errCode := store.GetStore().GetEventById(project.ID, eventId91, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event91.SessionId)

	event92, errCode := store.GetStore().GetEventById(project.ID, eventId92, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event92.SessionId)

	event93, errCode := store.GetStore().GetEventById(project.ID, eventId93, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event93.SessionId)

	assert.NotEqual(t, *event91.SessionId, *event92.SessionId)
	assert.NotEqual(t, *event92.SessionId, *event93.SessionId)

	event94, errCode := store.GetStore().GetEventById(project.ID, eventId94, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.NotEqual(t, *event93.SessionId, *event94.SessionId)

	event95, errCode := store.GetStore().GetEventById(project.ID, eventId95, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.Equal(t, *event94.SessionId, *event95.SessionId)

	event96, errCode := store.GetStore().GetEventById(project.ID, eventId96, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.Equal(t, *event95.SessionId, *event96.SessionId)

	event97, errCode := store.GetStore().GetEventById(project.ID, eventId97, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.NotEqual(t, *event96.SessionId, *event97.SessionId)

	event98, errCode := store.GetStore().GetEventById(project.ID, eventId98, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.NotEqual(t, *event98.SessionId, *event97.SessionId)

	event99, errCode := store.GetStore().GetEventById(project.ID, eventId99, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event94.SessionId)
	assert.NotEqual(t, *event99.SessionId, *event98.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
		RequestSource: model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	timestamp = timestamp + 5
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: U.PropertiesMap{},
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event2.SessionId)
	assert.Equal(t, *event2.SessionId, *event1.SessionId)

	event3, errCode := store.GetStore().GetEventById(project.ID, eventId3, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event3.SessionId)
	assert.Equal(t, *event3.SessionId, *event1.SessionId)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event4.SessionId, *event1.SessionId)

	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event5.SessionId)
	assert.Equal(t, *event5.SessionId, *event1.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
		RequestSource: model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	timestamp = timestamp + 5
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId,
		EventProperties: U.PropertiesMap{},
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event2.SessionId)
	assert.Equal(t, *event2.SessionId, *event1.SessionId)

	event3, errCode := store.GetStore().GetEventById(project.ID, eventId3, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event3.SessionId)
	assert.Equal(t, *event3.SessionId, *event1.SessionId)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event4.SessionId, *event1.SessionId)

	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event5.SessionId)
	assert.Equal(t, *event5.SessionId, *event1.SessionId)

	// since user id different, should create a new session for this
	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event6.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event5.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
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
		RequestSource: model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
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
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event1, errCode := store.GetStore().GetEventById(project.ID, eventId1, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event1.SessionId)

	event2, errCode := store.GetStore().GetEventById(project.ID, eventId2, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event2.SessionId)
	assert.Equal(t, *event2.SessionId, *event1.SessionId)

	event3, errCode := store.GetStore().GetEventById(project.ID, eventId3, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event3.SessionId)
	assert.Equal(t, *event3.SessionId, *event1.SessionId)

	event4, errCode := store.GetStore().GetEventById(project.ID, eventId4, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event4.SessionId)
	assert.Equal(t, *event4.SessionId, *event1.SessionId)

	//
	event5, errCode := store.GetStore().GetEventById(project.ID, eventId5, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event5.SessionId)
	assert.Equal(t, *event5.SessionId, *event1.SessionId)

	// since user id different, should create a new session for this
	event6, errCode := store.GetStore().GetEventById(project.ID, eventId6, "")
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, event6.SessionId)
	assert.NotEqual(t, *event6.SessionId, *event5.SessionId)

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}

func TestGetAllEventsForSessionCreationAsUserEventsMap(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	maxLookbackTimestamp := U.UnixTimeBeforeDuration(31 * 24 * time.Hour)

	// Test: New user with one event and one skip_session event.
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)

	// Updating project timestamp to before events start timestamp.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, timestamp-1)
	assert.Equal(t, http.StatusAccepted, errCode)

	randomEventName := RandomURL()
	trackEventProperties1 := U.PropertiesMap{
		U.EP_REFERRER:        "www.google.com",
		U.EP_PAGE_URL:        "https://example.com/1/2/",
		U.EP_PAGE_RAW_URL:    "https://example.com/1/2?x=1",
		U.EP_PAGE_SPENT_TIME: 10,
	}
	trackPayload := SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		EventProperties: trackEventProperties1,
		RequestSource:   model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, response.UserId)
	eventId1 := response.EventId
	userId1 := response.UserId

	// no session created.
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Name:          "non_web_event",
		Timestamp:     timestamp,
		UserId:        userId1,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, true, SDK.SourceJSSDK, "") // skip session.
	assert.Equal(t, http.StatusOK, status)
	skipSessionEventId1 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	_, errCode = store.GetStore().GetEventById(project.ID, eventId1, "")
	assert.Equal(t, http.StatusFound, errCode)

	_, errCode = store.GetStore().GetEventById(project.ID, skipSessionEventId1, "")
	assert.Equal(t, http.StatusFound, errCode)

	session1 := assertAssociatedSession(t, project.ID, []string{eventId1, skipSessionEventId1},
		[]string{skipSessionEventId1}, "Session 1")
	session1Properties, err := U.DecodePostgresJsonb(&session1.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(1), (*session1Properties)["$session_count"])
	assert.Equal(t, float64(1), (*session1Properties)["$page_count"])

	// Test: New events without session for existing user with session.
	// Since there is continuous activity, last session should be continued.
	timestamp = timestamp + 1
	randomEventName = RandomURL()
	trackEventProperties2 := U.PropertiesMap{
		U.EP_REFERRER:     "www.yahoo.com",
		U.EP_PAGE_URL:     "https://example1.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example1.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId1,
		EventProperties: trackEventProperties2,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId2 := response.EventId

	timestamp = timestamp + 1
	randomEventName = RandomURL()
	trackEventProperties3 := U.PropertiesMap{
		U.EP_REFERRER:     "www.facebook.com",
		U.EP_PAGE_URL:     "https://example2.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example2.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId1,
		EventProperties: trackEventProperties3,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId3 := response.EventId

	// inactivity.
	timestamp = timestamp + (35 * 60) // + 35 mins
	randomEventName = RandomURL()
	trackEventProperties4 := U.PropertiesMap{
		U.EP_REFERRER:     "www.bing.com",
		U.EP_PAGE_URL:     "https://example3.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example3.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId1,
		EventProperties: trackEventProperties4,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId4 := response.EventId

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Name:          "non_web_event",
		Timestamp:     timestamp,
		UserId:        userId1,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, true, SDK.SourceJSSDK, "") // skip session.
	assert.Equal(t, http.StatusOK, status)
	skipSessionEventId2 := response.EventId

	timestamp = timestamp + 1
	randomEventName = RandomURL()
	trackEventProperties5 := U.PropertiesMap{
		U.EP_REFERRER:     "www.hacker.com",
		U.EP_PAGE_URL:     "https://example4.com/1/2/",
		U.EP_PAGE_RAW_URL: "https://example4.com/1/2?x=1",
	}
	trackPayload = SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp,
		UserId:          userId1,
		EventProperties: trackEventProperties5,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId5 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	_, errCode = store.GetStore().GetEventById(project.ID, eventId2, "")
	assert.Equal(t, http.StatusFound, errCode)

	_, errCode = store.GetStore().GetEventById(project.ID, eventId3, "")
	assert.Equal(t, http.StatusFound, errCode)

	_, errCode = store.GetStore().GetEventById(project.ID, eventId4, "")
	assert.Equal(t, http.StatusFound, errCode)

	_, errCode = store.GetStore().GetEventById(project.ID, skipSessionEventId2, "")
	assert.Equal(t, http.StatusFound, errCode)

	_, errCode = store.GetStore().GetEventById(project.ID, eventId5, "")
	assert.Equal(t, http.StatusFound, errCode)

	// Order of events - event1, skip session event 1, event2, event3
	// Session 1 - event1, event2, event3
	session1Continued := assertAssociatedSession(t, project.ID, []string{eventId1, skipSessionEventId1, eventId2, eventId3},
		[]string{skipSessionEventId1}, "Session 1 continued.")
	session1ContinuedProperties, err := U.DecodePostgresJsonb(&session1Continued.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(1), (*session1ContinuedProperties)["$session_count"])
	assert.Equal(t, float64(3), (*session1ContinuedProperties)["$page_count"])

	// Order of events - event4, skip session event 2, event5
	// Session 2 - event4, event5
	session2 := assertAssociatedSession(t, project.ID, []string{eventId4, skipSessionEventId2, eventId5},
		[]string{skipSessionEventId2}, "Session 2")
	session2Properties, err := U.DecodePostgresJsonb(&session2.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(2), (*session2Properties)["$session_count"])
	assert.Equal(t, float64(2), (*session2Properties)["$page_count"])

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId1,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "summer_sale",
		},
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId6 := response.EventId

	timestamp = timestamp + 2
	randomEventName = RandomURL()
	trackPayload = SDK.TrackPayload{
		Auto:      true,
		Name:      randomEventName,
		Timestamp: timestamp,
		UserId:    userId1,
		EventProperties: U.PropertiesMap{
			U.QUERY_PARAM_UTM_PREFIX + "campaign": "winter_sale",
		},
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId7 := response.EventId

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:          true,
		Name:          randomEventName,
		Timestamp:     timestamp,
		UserId:        userId1,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId8 := response.EventId

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:          true,
		Name:          randomEventName,
		Timestamp:     timestamp,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId9 := response.EventId
	userId2 := response.UserId

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:          true,
		Name:          randomEventName,
		Timestamp:     timestamp,
		UserId:        userId1,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId10 := response.EventId

	timestamp = timestamp + 2
	trackPayload = SDK.TrackPayload{
		Auto:          true,
		Name:          randomEventName,
		Timestamp:     timestamp,
		UserId:        userId2,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	eventId11 := response.EventId

	_, err = TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	_, errCode = store.GetStore().GetEventById(project.ID, eventId6, "")
	assert.Equal(t, http.StatusFound, errCode)

	// Should create new session as campaign property exist.
	session3 := assertAssociatedSession(t, project.ID, []string{eventId6},
		[]string{}, "Session 3")
	session3Properties, err := U.DecodePostgresJsonb(&session3.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(3), (*session3Properties)["$session_count"])
	assert.Equal(t, float64(1), (*session3Properties)["$page_count"])

	_, errCode = store.GetStore().GetEventById(project.ID, eventId7, "")
	assert.Equal(t, http.StatusFound, errCode)

	_, errCode = store.GetStore().GetEventById(project.ID, eventId8, "")
	assert.Equal(t, http.StatusFound, errCode)

	_, errCode = store.GetStore().GetEventById(project.ID, eventId10, "")
	assert.Equal(t, http.StatusFound, errCode)

	// Should create new session as campaign property changed.
	session4 := assertAssociatedSession(t, project.ID, []string{eventId7, eventId8, eventId10},
		[]string{}, "Session 4")
	session4Properties, err := U.DecodePostgresJsonb(&session4.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(4), (*session4Properties)["$session_count"])
	assert.Equal(t, float64(3), (*session4Properties)["$page_count"])

	_, errCode = store.GetStore().GetEventById(project.ID, eventId9, "")
	assert.Equal(t, http.StatusFound, errCode)

	_, errCode = store.GetStore().GetEventById(project.ID, eventId11, "")
	assert.Equal(t, http.StatusFound, errCode)

	// Should create new session for new user.
	newSession1 := assertAssociatedSession(t, project.ID, []string{eventId9, eventId11},
		[]string{}, "New Session 1")
	newSession1Properties, err := U.DecodePostgresJsonb(&newSession1.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(1), (*newSession1Properties)["$session_count"])
	assert.Equal(t, float64(2), (*newSession1Properties)["$page_count"])

	// Test: Project with no events and all events with session already.
	statusMap, err := TaskSession.AddSession([]int64{project.ID}, maxLookbackTimestamp, 0, 0, 30, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, "not_modified", statusMap[project.ID].Status)
}

func TestAddSessionRemoveEventLevelUserProperties(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	timestamp := time.Now().AddDate(0, 0, -1)
	timestampUnix := timestamp.Unix()

	randomEventName := RandomURL()
	trackEventProperties := U.PropertiesMap{
		U.EP_PAGE_URL:     "https://example.com",
		U.EP_PAGE_RAW_URL: "https://example.com",
		U.EP_CAMPAIGN_ID:  "124",
	}
	trackUserProperties := U.PropertiesMap{
		U.UP_OS: "android1",
	}

	trackPayload := SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestampUnix,
		EventProperties: trackEventProperties,
		UserProperties:  trackUserProperties,
		RequestSource:   model.UserSourceWeb,
	}
	status, res := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	userID := res.UserId
	// session not created.
	_, errCode := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	_, err = TaskSession.AddSession([]int64{project.ID}, 2*24*60*60, 0, 0, 30, 1, 1)

	assert.Nil(t, err)

	// session created.
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusFound, errCode)

	sessionEvent := assertAssociatedSession(t, project.ID, []string{res.EventId},
		[]string{}, "Session")

	// session event properties .
	lsEventProperties, err := U.DecodePostgresJsonb(&sessionEvent.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties)[U.SP_PAGE_COUNT], float64(1))
	assert.Equal(t, (*lsEventProperties)[U.SP_SPENT_TIME], float64(1))
	lsUserProperties, err := U.DecodePostgresJsonb(sessionEvent.UserProperties)
	assert.Nil(t, err)
	assert.NotEqual(t, (*lsUserProperties)[U.UP_PAGE_COUNT], float64(1))
	assert.NotEqual(t, (*lsUserProperties)[U.UP_TOTAL_SPENT_TIME], float64(1))

	//user properties
	user, status := store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, status)
	properitesMap := make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	assert.Equal(t, properitesMap[U.UP_PAGE_COUNT], float64(1))
	assert.Equal(t, properitesMap[U.UP_TOTAL_SPENT_TIME], float64(1))
	assert.Nil(t, properitesMap[U.EP_SESSION_COUNT])
}

func TestPageCountAndSessionSpentTime(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	timestamp := time.Now().AddDate(0, 0, -1)
	timestampUnix := timestamp.Unix()

	randomEventName := RandomURL()
	trackEventProperties := U.PropertiesMap{
		U.EP_PAGE_URL:        "https://example.com",
		U.EP_PAGE_RAW_URL:    "https://example.com",
		U.EP_CAMPAIGN_ID:     "124",
		U.EP_PAGE_SPENT_TIME: 10,
	}
	trackUserProperties := U.PropertiesMap{
		U.UP_OS: "android1",
	}

	trackPayload := SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestampUnix,
		EventProperties: trackEventProperties,
		UserProperties:  trackUserProperties,
		RequestSource:   model.UserSourceWeb,
	}
	status, res := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	userID := res.UserId
	eventId1 := res.EventId

	timestampUnix = timestampUnix + 2

	trackEventProperties1 := U.PropertiesMap{
		U.EP_PAGE_URL:     "https://example.com",
		U.EP_PAGE_RAW_URL: "https://example.com",
		U.EP_CAMPAIGN_ID:  "124",
	}
	trackUserProperties1 := U.PropertiesMap{
		U.UP_OS: "android1",
	}

	RandomString := U.RandomString(5)
	trackPayload = SDK.TrackPayload{
		Auto:            false,
		Name:            RandomString,
		UserId:          userID,
		Timestamp:       timestampUnix,
		EventProperties: trackEventProperties1,
		UserProperties:  trackUserProperties1,
		RequestSource:   model.UserSourceWeb,
	}

	status, res = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)

	// session not created.
	_, errCode := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	_, err = TaskSession.AddSession([]int64{project.ID}, 2*24*60*60, 0, 0, 30, 1, 1)

	assert.Nil(t, err)

	// session created.
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusFound, errCode)

	sessionEvent := assertAssociatedSession(t, project.ID, []string{eventId1},
		[]string{}, "Session")

	// session event properties .
	lsEventProperties, err := U.DecodePostgresJsonb(&sessionEvent.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties)[U.SP_PAGE_COUNT], float64(1))
	assert.Equal(t, (*lsEventProperties)[U.SP_SPENT_TIME], float64(10))

	//user properties
	user, status := store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, status)
	properitesMap := make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	assert.Equal(t, properitesMap[U.UP_PAGE_COUNT], float64(1))
	assert.Equal(t, properitesMap[U.UP_TOTAL_SPENT_TIME], float64(10))
	assert.Nil(t, properitesMap[U.EP_SESSION_COUNT])
}

func TestAddSessionUserRealSessionProperties(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	timestamp := time.Now().AddDate(0, 0, -1)
	timestampUnix := timestamp.Unix()

	randomEventName := RandomURL()
	trackEventProperties := U.PropertiesMap{
		U.EP_PAGE_URL:     "https://example.com",
		U.EP_PAGE_RAW_URL: "https://example.com",
		U.EP_CAMPAIGN_ID:  "124",
	}
	trackUserProperties := U.PropertiesMap{
		U.UP_OS: "android1",
	}

	trackPayload := SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestampUnix,
		EventProperties: trackEventProperties,
		UserProperties:  trackUserProperties,
		RequestSource:   model.UserSourceWeb,
	}
	trackPayload1 := trackPayload
	status, res := SDK.Track(project.ID, &trackPayload1, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	userID1 := res.UserId
	eventID1 := res.EventId

	status, _ = SDK.Identify(project.ID, &SDK.IdentifyPayload{
		UserId:         userID1,
		CustomerUserId: "abc1@abc1.com",
		JoinTimestamp:  timestampUnix,
		RequestSource:  model.UserSourceWeb,
	}, false)
	assert.Equal(t, http.StatusOK, status)

	// session not created.
	_, errCode := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusNotFound, errCode)

	_, err = TaskSession.AddSession([]int64{project.ID}, 2*24*60*60, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	// session event name created.
	_, errCode = store.GetStore().GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusFound, errCode)

	event, status := store.GetStore().GetEvent(project.ID, userID1, eventID1)
	assert.Equal(t, http.StatusFound, errCode)

	sessionEvent, status := store.GetStore().GetEvent(project.ID, userID1, *event.SessionId)
	assert.Equal(t, http.StatusFound, errCode)

	// session event properties .
	lsEventProperties, err := U.DecodePostgresJsonb(&sessionEvent.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*lsEventProperties)[U.SP_PAGE_COUNT], float64(1))
	assert.Equal(t, (*lsEventProperties)[U.SP_SPENT_TIME], float64(1))
	assert.Nil(t, (*lsEventProperties)[U.UP_REAL_PAGE_SPENT_TIME])
	assert.Nil(t, (*lsEventProperties)[U.UP_REAL_PAGE_COUNT])
	lsUserProperties, err := U.DecodePostgresJsonb(sessionEvent.UserProperties)
	assert.Nil(t, err)
	assert.NotEqual(t, (*lsUserProperties)[U.UP_PAGE_COUNT], float64(1))
	assert.NotEqual(t, (*lsUserProperties)[U.UP_TOTAL_SPENT_TIME], float64(1))
	assert.Equal(t, float64(0), (*lsUserProperties)[U.UP_REAL_PAGE_SPENT_TIME])
	assert.Equal(t, float64(0), (*lsUserProperties)[U.UP_REAL_PAGE_COUNT])

	trackUserProperties = U.PropertiesMap{
		U.UP_OS: "android2",
	}

	trackPayload2 := trackPayload
	trackPayload2.UserProperties = trackUserProperties
	status, res = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	userID2 := res.UserId
	eventID2 := res.EventId

	status, _ = SDK.Identify(project.ID, &SDK.IdentifyPayload{
		UserId:         userID2,
		CustomerUserId: "abc1@abc1.com",
		JoinTimestamp:  timestampUnix,
		RequestSource:  model.UserSourceWeb,
	}, false)
	assert.Equal(t, http.StatusOK, status)

	properties := &U.PropertiesMap{
		"$page_spent_time": 1.34, "$page_load_time": 1.594, "$page_scroll_percent": 97.54}
	// add page spent time
	errCode = store.GetStore().UpdateEventProperties(project.ID, eventID2, userID2, properties, time.Now().Unix(), nil)
	assert.Equal(t, http.StatusAccepted, errCode)

	trackPayload2 = trackPayload
	trackPayload2.Name = randomEventName
	trackPayload2.UserId = userID2
	status, _ = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)

	_, err = TaskSession.AddSession([]int64{project.ID}, 2*24*60*60, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	event, status = store.GetStore().GetEvent(project.ID, userID2, eventID2)
	assert.Equal(t, http.StatusFound, status)

	sessionEvent, status = store.GetStore().GetEvent(project.ID, userID2, *event.SessionId)
	assert.Equal(t, http.StatusFound, status)

	// session event properties .
	lsEventProperties, err = U.DecodePostgresJsonb(&sessionEvent.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(2), (*lsEventProperties)[U.SP_PAGE_COUNT])
	assert.Equal(t, float64(2.34), (*lsEventProperties)[U.SP_SPENT_TIME])
	assert.Nil(t, (*lsEventProperties)[U.UP_REAL_PAGE_SPENT_TIME])
	assert.Nil(t, (*lsEventProperties)[U.UP_REAL_PAGE_COUNT])
	lsUserProperties, err = U.DecodePostgresJsonb(sessionEvent.UserProperties)
	assert.Nil(t, err)
	assert.NotEqual(t, (*lsUserProperties)[U.UP_PAGE_COUNT], float64(1))
	assert.NotEqual(t, (*lsUserProperties)[U.UP_TOTAL_SPENT_TIME], float64(1))
	assert.Equal(t, float64(0), (*lsUserProperties)[U.UP_REAL_PAGE_SPENT_TIME])
	assert.Equal(t, float64(0), (*lsUserProperties)[U.UP_REAL_PAGE_COUNT])

	// real properties for each user is kept as it is without merging
	user, status := store.GetStore().GetUser(project.ID, userID1)
	assert.Equal(t, http.StatusFound, status)
	userProperties, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(1), (*userProperties)[U.UP_REAL_PAGE_SPENT_TIME])
	assert.Equal(t, float64(1), (*userProperties)[U.UP_REAL_PAGE_COUNT])
	assert.Equal(t, float64(3), (*userProperties)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(3.34), (*userProperties)[U.UP_TOTAL_SPENT_TIME])
	assert.Equal(t, "android2", (*userProperties)[U.UP_OS])

	user, status = store.GetStore().GetUser(project.ID, userID2)
	assert.Equal(t, http.StatusFound, status)
	userProperties, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(2.34), (*userProperties)[U.UP_REAL_PAGE_SPENT_TIME])
	assert.Equal(t, float64(2), (*userProperties)[U.UP_REAL_PAGE_COUNT])
	assert.Equal(t, float64(3), (*userProperties)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(3.34), (*userProperties)[U.UP_TOTAL_SPENT_TIME])
	assert.Equal(t, "android2", (*userProperties)[U.UP_OS])

	// user 2 getting another page view event, should add 1 for page spent time and page count
	trackPayload2 = trackPayload
	trackPayload2.Name = randomEventName
	trackPayload2.UserId = userID2
	status, _ = SDK.Track(project.ID, &trackPayload2, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)

	_, err = TaskSession.AddSession([]int64{project.ID}, 2*24*60*60, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	user, status = store.GetStore().GetUser(project.ID, userID2)
	assert.Equal(t, http.StatusFound, status)
	userProperties, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(3.34), (*userProperties)[U.UP_REAL_PAGE_SPENT_TIME])
	assert.Equal(t, float64(3), (*userProperties)[U.UP_REAL_PAGE_COUNT])
	assert.Equal(t, float64(4), (*userProperties)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(4.34), (*userProperties)[U.UP_TOTAL_SPENT_TIME])
	assert.Equal(t, "android2", (*userProperties)[U.UP_OS])

	// test when old user without real session spent time gets updated
	trackPayload3 := trackPayload
	status, res = SDK.Track(project.ID, &trackPayload3, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	userID3 := res.UserId

	user, status = store.GetStore().GetUser(project.ID, userID3)
	assert.Equal(t, http.StatusFound, status)
	userProperties, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	delete(*userProperties, U.UP_REAL_PAGE_SPENT_TIME)
	delete(*userProperties, U.UP_REAL_PAGE_COUNT)

	userPropertiesJsonB, err := U.EncodeToPostgresJsonb(userProperties)
	assert.Nil(t, err)

	status = store.GetStore().OverwriteUserPropertiesByID(project.ID, userID3, nil, userPropertiesJsonB, false, 0, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusAccepted, status)
	_, err = TaskSession.AddSession([]int64{project.ID}, 2*24*60*60, 0, 0, 30, 1, 1)
	assert.Nil(t, err)

	user, status = store.GetStore().GetUser(project.ID, userID3)
	assert.Equal(t, http.StatusFound, status)
	userProperties, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(2), (*userProperties)[U.UP_REAL_PAGE_SPENT_TIME]) // adds 1 for all previous data
	assert.Equal(t, float64(2), (*userProperties)[U.UP_REAL_PAGE_COUNT])      // adds 1 for all previous data
	assert.Equal(t, float64(2), (*userProperties)[U.UP_PAGE_COUNT])
	assert.Equal(t, float64(2), (*userProperties)[U.UP_TOTAL_SPENT_TIME])
	assert.Equal(t, "android1", (*userProperties)[U.UP_OS])
}
