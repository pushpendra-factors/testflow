package tests

import (
	"encoding/json"
	M "factors/model"
	U "factors/util"
	"fmt"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndGetEvent(t *testing.T) {
	// Initialize a project, user and  the event.
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	start := time.Now()

	// Test successful CreateEvent.
	newEvent := &M.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: start.Unix()}
	event, errCode := M.CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, len(event.ID) > 30)
	assert.Equal(t, projectId, event.ProjectId)
	assert.Equal(t, eventNameId, event.EventNameId)
	assert.Equal(t, uint64(1), event.Count)
	assert.True(t, event.Timestamp >= start.Unix())
	assert.InDelta(t, event.Timestamp, start.Unix(), 3)
	assert.Equal(t, event.Timestamp, event.PropertiesUpdatedTimestamp)
	assert.True(t, event.CreatedAt.After(start))
	assert.True(t, event.UpdatedAt.After(start))
	assert.Equal(t, event.CreatedAt, event.UpdatedAt)

	// Test Get Event on the created.
	retEvent, errCode := M.GetEvent(projectId, userId, event.ID)
	assert.Equal(t, http.StatusFound, errCode)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(event.CreatedAt.Sub(retEvent.CreatedAt).Seconds()) < 0.1)
	assert.True(t, math.Abs(event.UpdatedAt.Sub(retEvent.UpdatedAt).Seconds()) < 0.1)
	event.CreatedAt = time.Time{}
	event.UpdatedAt = time.Time{}
	retEvent.CreatedAt = time.Time{}
	retEvent.UpdatedAt = time.Time{}

	assert.Equal(t, newEvent.EventNameId, retEvent.EventNameId)
	assert.Equal(t, newEvent.ProjectId, retEvent.ProjectId)
	assert.Equal(t, newEvent.UserId, retEvent.UserId)
	assert.True(t, event.Timestamp != 0)
	assert.Equal(t, event.Timestamp, event.PropertiesUpdatedTimestamp)
	eventProperties, _ := U.DecodePostgresJsonb(&event.Properties)
	assert.True(t, (*eventProperties)["$day_of_week"] != "" && (*eventProperties)["$day_of_week"] == time.Unix(event.Timestamp, 0).Weekday().String())
	hr, _, _ := time.Unix(event.Timestamp, 0).Clock()
	assert.True(t, (*eventProperties)["$hour_of_day"] != 0 && (*eventProperties)["$hour_of_day"] == float64(hr))

	// Test Get Event with wrong project id.
	retEvent, errCode = M.GetEvent(projectId+1, userId, event.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)
	// Test Get Event with wrong user id.
	retEvent, errCode = M.GetEvent(projectId, "randomId", event.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)
	// Test successful CreateEvent with count increment
	event, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId,
		ProjectId: projectId, UserId: userId, Timestamp: start.Unix()})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, len(event.ID) > 30)
	assert.Equal(t, projectId, event.ProjectId)
	assert.Equal(t, eventNameId, event.EventNameId)
	assert.Equal(t, uint64(2), event.Count)
	assert.True(t, event.Timestamp >= start.Unix())
	assert.InDelta(t, event.Timestamp, start.Unix(), 3)
	assert.Equal(t, event.Timestamp, event.PropertiesUpdatedTimestamp)
	assert.True(t, event.CreatedAt.After(start))
	assert.True(t, event.UpdatedAt.After(start))
	assert.Equal(t, event.CreatedAt, event.UpdatedAt)
	assert.True(t, event.Timestamp != 0)
	eventProperties, err = U.DecodePostgresJsonb(&event.Properties)
	assert.Equal(t, err, nil)
	assert.Equal(t, (*eventProperties)["$day_of_week"], time.Unix(event.Timestamp, 0).Weekday().String())
	hr, _, _ = time.Unix(event.Timestamp, 0).Clock()
	assert.Equal(t, (*eventProperties)["$hour_of_day"], float64(hr))

	t.Run("DuplicateCustomerEventId", func(t *testing.T) {
		custEventId := U.RandomString(8)
		//projectId, userId, eventNameId, err := SetupProjectUserEventName()
		assert.Nil(t, err)

		event, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId, ProjectId: projectId,
			UserId: userId, CustomerEventId: &custEventId, Timestamp: time.Now().Unix()})
		assert.Equal(t, http.StatusCreated, errCode)
		_, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId, ProjectId: projectId,
			UserId: userId, CustomerEventId: &custEventId, Timestamp: time.Now().Unix()})
		assert.Equal(t, http.StatusNotAcceptable, errCode)
	})

	// Test Get Event on non existent id.
	retEvent, errCode = M.GetEvent(projectId, userId, "9ad21963-bcfb-4563-aa02-8ea589710d1a")
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)

	// Test Create Event with properties.
	properties := json.RawMessage(`{"email": "random@example.com"}`)
	event, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Properties: postgres.Jsonb{properties}, Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, event)
	assert.NotEqual(t, postgres.Jsonb{RawMessage: json.RawMessage(nil)}, event.Properties)
	// Retrieve and validate properties.
	retEvent, errCode = M.GetEvent(projectId, userId, event.ID)
	eventPropertiesBytes, err := retEvent.Properties.Value()
	var eventPropertiesMap map[string]interface{}
	json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
	assert.Equal(t, "random@example.com", eventPropertiesMap["email"])
	// Test Get Event on wrong format of id.
	retEvent, errCode = M.GetEvent(projectId, userId, "r4nd0m!234")
	assert.Equal(t, http.StatusInternalServerError, errCode)
	assert.Nil(t, retEvent)

	// Test Create Event with external id.
	eventId := uuid.New().String()
	event, errCode = M.CreateEvent(&M.Event{ID: eventId, EventNameId: eventNameId,
		ProjectId: projectId, UserId: userId, Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, event)
	assert.Equal(t, eventId, event.ID)

	// Test Create Event with invalid uuid as id.
	eventId = U.RandomLowerAphaNumString(10)
	event, errCode = M.CreateEvent(&M.Event{ID: eventId, EventNameId: eventNameId,
		ProjectId: projectId, UserId: userId, Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusInternalServerError, errCode)
	assert.Nil(t, event)

	// Test Create Event without projectId.
	event, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId, UserId: userId,
		Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, event)

	// Test Create Event without userId.
	event, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId,
		ProjectId: projectId, Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, event)

	// Test Create Event without eventNameId.
	event, errCode = M.CreateEvent(&M.Event{EventNameId: 0, ProjectId: projectId, UserId: userId,
		Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, event)
}

func createEventWithTimestamp(t *testing.T, project *M.Project, user *M.User, timestamp int64) (*M.EventName, *M.Event) {
	eventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: fmt.Sprintf("event_%d", timestamp)})
	assert.NotNil(t, eventName)
	event, errCode := M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID, Timestamp: timestamp})
	assert.Equal(t, http.StatusCreated, errCode)
	return eventName, event
}

func TestGetFirstLastEventTimestamp(t *testing.T) {
	project, user, _ := SetupProjectUserReturnDAO()
	assert.NotNil(t, project)

	var firstTimestamp int64 = 1393632004
	var secondTimestamp int64 = 1393633007
	var thirdTimestamp int64 = 1393634005

	createEventWithTimestamp(t, project, user, firstTimestamp)
	createEventWithTimestamp(t, project, user, secondTimestamp)
	createEventWithTimestamp(t, project, user, thirdTimestamp)

	// Test with exact limit timestamp
	ts1, errCode := M.GetProjectEventsInfo()
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, ts1)
	assert.NotNil(t, (*ts1)[project.ID])
	assert.Equal(t, firstTimestamp, (*ts1)[project.ID].FirstEventTimestamp)
	assert.Equal(t, thirdTimestamp, (*ts1)[project.ID].LastEventTimestamp)

	// Test with increased limit timestamp
	ts1, errCode = M.GetProjectEventsInfo() // adds 3 secs.
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, ts1)
	assert.NotNil(t, (*ts1)[project.ID])
	assert.Equal(t, firstTimestamp, (*ts1)[project.ID].FirstEventTimestamp)
	assert.Equal(t, thirdTimestamp, (*ts1)[project.ID].LastEventTimestamp)
	assert.Nil(t, (*ts1)[999999])
	assert.Equal(t, 3, (*ts1)[project.ID].EventsCount)
}

func createEventWithTimestampAndPrperties(t *testing.T, project *M.Project, user *M.User, timestamp int64, properties json.RawMessage) (*M.EventName, *M.Event) {
	eventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: fmt.Sprintf("event_%d", timestamp)})
	assert.NotNil(t, eventName)
	event, errCode := M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID,
		Timestamp: timestamp, Properties: postgres.Jsonb{properties}})
	assert.Equal(t, http.StatusCreated, errCode)
	return eventName, event
}

func TestGetRecentEventPropertyKeys(t *testing.T) {
	project, user, _ := SetupProjectUserReturnDAO()
	assert.NotNil(t, project)

	t.Run("RecentPropertiesWithLimit", func(t *testing.T) {
		timestamp := time.Now().Unix()
		eventName, _ := createEventWithTimestampAndPrperties(t, project, user, timestamp, json.RawMessage(`{"rProp1": "value1", "rProp2": 1}`))
		_, errCode1 := M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID,
			Timestamp: timestamp, Properties: postgres.Jsonb{json.RawMessage(`{"rProp3": "value2", "rProp4": 2}`)}})
		assert.Equal(t, http.StatusCreated, errCode1)
		props, err := M.GetRecentEventPropertyKeysWithLimits(project.ID, eventName.Name, time.Now().AddDate(0, 0, -30).Unix(), timestamp, 100)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(props) >= 4, true)
		propertyMap := make(map[string]bool)
		for _, property := range props {
			propertyMap[property.Key] = true
		}
		assert.Equal(t, propertyMap["rProp1"], true)
		assert.Equal(t, propertyMap["rProp2"], true)
		assert.Equal(t, propertyMap["rProp3"], true)
		assert.Equal(t, propertyMap["rProp4"], true)
	})

	t.Run("PropertiesOlderThan24Hours", func(t *testing.T) {
		timestamp := U.UnixTimeBeforeDuration(24 * time.Hour)
		eventName, _ := createEventWithTimestampAndPrperties(t, project, user, timestamp, json.RawMessage(`{"rProp1": "value1", "rProp2": 1}`))

		props, err := M.GetRecentEventPropertyKeysWithLimits(project.ID, eventName.Name, time.Now().AddDate(0, 0, -1).Unix(), timestamp, 100)
		assert.Equal(t, nil, err)
		assert.Len(t, props, 0)
	})
}

func TestGetRecentEventPropertyValues(t *testing.T) {
	project, user, _ := SetupProjectUserReturnDAO()
	assert.NotNil(t, project)

	t.Run("RecentPropertyValues", func(t *testing.T) {
		timestamp := time.Now().Unix()
		eventName, _ := createEventWithTimestampAndPrperties(t, project, user, timestamp, json.RawMessage(`{"rProp1": "value1"}`))
		_, errCode1 := M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID,
			Timestamp: timestamp, Properties: postgres.Jsonb{json.RawMessage(`{"rProp1": "value2"}`)}})
		assert.Equal(t, http.StatusCreated, errCode1)
		_, errCode1 = M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID,
			Timestamp: timestamp, Properties: postgres.Jsonb{json.RawMessage(`{"rProp1": "value1", "rProp2": 1}`)}})
		assert.Equal(t, http.StatusCreated, errCode1)

		values, category, err := M.GetRecentEventPropertyValuesWithLimits(project.ID, eventName.Name, "rProp1", 10, 100, time.Now().AddDate(0, 0, -1).Unix(), timestamp)
		assert.Equal(t, nil, err)
		assert.Len(t, values, 2)
		assert.Equal(t, values[0].Value, "value1")
		assert.Equal(t, values[1].Value, "value2")
		assert.Equal(t, category, U.PropertyTypeCategorical)

		// limited values to 1.
		values1, category, err := M.GetRecentEventPropertyValuesWithLimits(project.ID, eventName.Name, "rProp1", 1, 100, time.Now().AddDate(0, 0, -30).Unix(), timestamp)
		assert.Equal(t, nil, err)
		assert.Len(t, values1, 1)
		assert.Equal(t, values1[0].Value, "value1")
		assert.Equal(t, category, U.PropertyTypeCategorical)

		values2, category, err := M.GetRecentEventPropertyValuesWithLimits(project.ID, eventName.Name, "rProp2", 10, 100, time.Now().AddDate(0, 0, -30).Unix(), timestamp)
		assert.Equal(t, nil, err)
		assert.Len(t, values2, 1)
		assert.Equal(t, values2[0].Value, "1")
		assert.Equal(t, category, U.PropertyTypeNumerical)
	})

	t.Run("PropertyValuesOlderThan24Hour", func(t *testing.T) {
		timestamp := U.UnixTimeBeforeDuration(24 * time.Hour)
		eventName, _ := createEventWithTimestampAndPrperties(t, project, user, timestamp, json.RawMessage(`{"rProp1": "value1"}`))
		_, errCode1 := M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID,
			Timestamp: timestamp, Properties: postgres.Jsonb{json.RawMessage(`{"rProp1": "value2"}`)}})
		assert.Equal(t, http.StatusCreated, errCode1)

		values, category, err := M.GetRecentEventPropertyValuesWithLimits(project.ID, eventName.Name, "rProp1", 100, 100, time.Now().AddDate(0, 0, -1).Unix(), timestamp)
		assert.Equal(t, nil, err)
		assert.Empty(t, values)
		assert.Equal(t, category, "")
	})
}

func TestUpdateEventProperties(t *testing.T) {
	project, user, _ := SetupProjectUserReturnDAO()
	assert.NotNil(t, project)

	timestamp := time.Now().Unix()
	_, event := createEventWithTimestampAndPrperties(t, project, user, timestamp,
		json.RawMessage(`{"rProp1": "value1", "rProp2": 1}`))

	// should add properties if not exist.
	errCode := M.UpdateEventProperties(project.ID, event.ID, &U.PropertiesMap{
		"$page_spent_time": 1.346, "$page_load_time": 1.594}, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, errCode)
	updatedEvent, errCode := M.GetEventById(project.ID, event.ID)
	assert.Equal(t, http.StatusFound, errCode)
	eventProperties, err := U.DecodePostgresJsonb(&updatedEvent.Properties)
	assert.Nil(t, err)
	assert.Contains(t, *eventProperties, "$page_spent_time")
	assert.Contains(t, *eventProperties, "$page_load_time")
	assert.Contains(t, *eventProperties, "rProp1") // should not remove old properties.
	// values should be unchanged.
	assert.Equal(t, float64(1.594), (*eventProperties)["$page_load_time"])
	assert.Equal(t, float64(1.346), (*eventProperties)["$page_spent_time"])
	assert.Equal(t, "value1", (*eventProperties)["rProp1"])

	// should update properties if exist.
	errCode = M.UpdateEventProperties(project.ID, event.ID, &U.PropertiesMap{
		"$page_spent_time": 3}, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, errCode)
	updatedEvent, errCode = M.GetEventById(project.ID, event.ID)
	assert.Equal(t, http.StatusFound, errCode)
	eventProperties, err = U.DecodePostgresJsonb(&updatedEvent.Properties)
	assert.Contains(t, *eventProperties, "$page_spent_time")
	assert.Contains(t, *eventProperties, "$page_load_time")
	assert.Contains(t, *eventProperties, "rProp1") // should not remove old properties.
	// should update the property alone.
	assert.Equal(t, float64(3), (*eventProperties)["$page_spent_time"])
	assert.Equal(t, float64(1.594), (*eventProperties)["$page_load_time"])
	assert.Equal(t, "value1", (*eventProperties)["rProp1"])
}

func TestGetLatestAnyEventOfUserInDuration(t *testing.T) {
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	// no event exist in 10 secs.
	event, errCode := M.GetLatestAnyEventOfUserForSession(projectId, userId,
		U.UnixTimeBeforeDuration(30*time.Minute))
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, event)

	createEventTimestamp := U.UnixTimeBeforeDuration(time.Minute * 100)
	// user active.
	createdEvent, errCode := M.CreateEvent(&M.Event{EventNameId: eventNameId,
		ProjectId: projectId, UserId: userId, Timestamp: createEventTimestamp})
	assert.Equal(t, http.StatusCreated, errCode)
	event, errCode = M.GetLatestAnyEventOfUserForSession(projectId, userId,
		createEventTimestamp+300) // after 5 mins of last activity.
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, errCode)
	assert.Equal(t, createdEvent.ID, event.ID)

	// inactivity.
	_, errCode = M.GetLatestAnyEventOfUserForSession(projectId, userId,
		createEventTimestamp+1800) // after 30 mins of last activity.
	assert.Equal(t, http.StatusNotFound, errCode)
}

func TestCacheEvent(t *testing.T) {
	for i := 0; i < 30; i++ {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.Nil(t, err)

		eventId := U.RandomString(10)
		timestamp := time.Now().Unix() - 100
		err = M.SetCacheUserLastEvent(project.ID, user.ID,
			&M.CacheEvent{ID: eventId, Timestamp: timestamp})
		assert.Nil(t, err)

		cacheEvent, err := M.GetCacheUserLastEvent(project.ID, user.ID)
		assert.NotNil(t, cacheEvent)
		assert.Equal(t, eventId, cacheEvent.ID)
		assert.Equal(t, timestamp, cacheEvent.Timestamp)
		assert.Nil(t, err)

		event, errCode := M.CreateEvent(&M.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: user.ID, Timestamp: time.Now().Unix()})
		assert.Equal(t, http.StatusCreated, errCode)

		cacheEvent, err = M.GetCacheUserLastEvent(project.ID, user.ID)
		assert.NotNil(t, cacheEvent)
		assert.Equal(t, event.ID, cacheEvent.ID)
		assert.True(t, cacheEvent.Timestamp > timestamp)
		assert.Nil(t, err)
	}
}

func TestGetLatestAnyEventOfUserInDurationFromCache(t *testing.T) {
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	// no event exist in 10 secs.
	event, errCode := M.GetLatestAnyEventOfUserForSessionFromCache(projectId, userId,
		U.UnixTimeBeforeDuration(30*time.Minute))
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, event)

	createEventTimestamp := U.UnixTimeBeforeDuration(time.Minute * 100)
	// user active.
	createdEvent, errCode := M.CreateEvent(&M.Event{EventNameId: eventNameId,
		ProjectId: projectId, UserId: userId, Timestamp: createEventTimestamp})
	assert.Equal(t, http.StatusCreated, errCode)
	event, errCode = M.GetLatestAnyEventOfUserForSessionFromCache(projectId,
		userId, createEventTimestamp+300) // after 5 mins of last activity.
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, errCode)
	assert.Equal(t, createdEvent.ID, event.ID)

	// inactivity.
	_, errCode = M.GetLatestAnyEventOfUserForSessionFromCache(projectId, userId,
		createEventTimestamp+1800) // after 30 mins of last activity.
	assert.Equal(t, http.StatusNotFound, errCode)
}

func TestCreateOrGetSessionEvent(t *testing.T) {
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	user, _ := M.GetUser(projectId, userId)
	userPropertiesId := user.PropertiesId
	sessionEventTimestamp := U.UnixTimeBeforeDuration(time.Minute * 32)

	t.Run("ShouldCreateNewSessionAsNoEventInLast30Mins", func(t *testing.T) {
		session, errCode := M.CreateOrGetSessionEvent(projectId, userId, false, false, sessionEventTimestamp,
			&U.PropertiesMap{U.EP_PAGE_LOAD_TIME: 0.10}, &U.PropertiesMap{}, userPropertiesId)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, session)

		// Session event should exist with initial event properites.
		sessionEvent, errCode := M.GetEvent(projectId, userId, session.ID)
		assert.Equal(t, http.StatusFound, errCode)
		eventPropertiesBytes, err := sessionEvent.Properties.Value()
		assert.Nil(t, err)
		var eventPropertiesMap map[string]interface{}
		json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
		assert.NotNil(t, eventPropertiesMap[U.UP_INITIAL_PAGE_LOAD_TIME])

		userPropertiesMap, errCode := M.GetUserPropertiesAsMap(projectId, userId)
		assert.Equal(t, errCode, http.StatusFound)
		assert.Nil(t, (*userPropertiesMap)[U.UP_LATEST_PAGE_LOAD_TIME])
		assert.Nil(t, (*userPropertiesMap)[U.UP_LATEST_CAMPAIGN])
	})

	t.Run("ShouldReturnLatestSessionAsUserWasActive", func(t *testing.T) {
		_, errCode := M.CreateEvent(&M.Event{EventNameId: eventNameId,
			ProjectId: projectId, UserId: userId, Timestamp: sessionEventTimestamp + 10})

		session, errCode := M.CreateOrGetSessionEvent(projectId, userId, false,
			false, sessionEventTimestamp+20, &U.PropertiesMap{}, &U.PropertiesMap{}, userPropertiesId)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, session)

		userProperties, _ := M.GetUserProperties(projectId, userId, userPropertiesId)
		userPropertiesMap, _ := U.DecodePostgresJsonb(userProperties)
		assert.NotNil(t, (*userPropertiesMap)[U.UP_SESSION_COUNT])

		userPropertiesMap, errCode = M.GetUserPropertiesAsMap(projectId, userId)
		assert.Equal(t, errCode, http.StatusFound)
		assert.Nil(t, (*userPropertiesMap)[U.UP_LATEST_PAGE_LOAD_TIME])
		assert.Nil(t, (*userPropertiesMap)[U.UP_LATEST_CAMPAIGN])
	})
}

func TestOverwriteEventProperties(t *testing.T) {
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	event, errCode := M.CreateEvent(&M.Event{EventNameId: eventNameId,
		ProjectId: projectId, UserId: userId, Timestamp: time.Now().Unix()})

	eventPropertiesMap, err := U.DecodePostgresJsonb(&event.Properties)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Nil(t, (*eventPropertiesMap)["Hello"])
	(*eventPropertiesMap)["Hello"] = "World"

	eventPropertiesJSONb, err := U.EncodeToPostgresJsonb(eventPropertiesMap)
	assert.Nil(t, err)

	errCode = M.OverwriteEventProperties(projectId, userId, event.ID, eventPropertiesJSONb)
	assert.Equal(t, errCode, http.StatusAccepted)

	rEvent, errCode := M.GetEvent(projectId, userId, event.ID)
	assert.Equal(t, http.StatusFound, errCode)

	rEventPropertiesMap, err := U.DecodePostgresJsonb(&rEvent.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*rEventPropertiesMap)["Hello"], "World")
}

func TestGetEventCountOfUserByEventName(t *testing.T) {
	project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
	assert.Nil(t, err)

	// No events, should return found and zero count.
	count, errCode := M.GetEventCountOfUserByEventName(project.ID, user.ID, eventName.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, uint64(0), count)

	newEvent := &M.Event{EventNameId: eventName.ID, ProjectId: project.ID,
		UserId: user.ID, Timestamp: time.Now().Unix()}
	event, errCode := M.CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, event)

	count, errCode = M.GetEventCountOfUserByEventName(project.ID, user.ID, eventName.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, uint64(1), count)
}

func TestGetDatesForNextEventsArchivalBatch(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	user, status := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.NotNil(t, user)
	assert.Equal(t, http.StatusCreated, status)

	timeNow := U.TimeNow()
	timeNowUnix := timeNow.Unix()

	createEventWithTimestampByName(t, project, user, "event1", timeNowUnix)
	// 1 Day older events.
	createEventWithTimestampByName(t, project, user, "event2", timeNowUnix-U.SECONDS_IN_A_DAY)
	createEventWithTimestampByName(t, project, user, "event3", timeNowUnix-U.SECONDS_IN_A_DAY-1)
	// 3 days older events.
	createEventWithTimestampByName(t, project, user, "event4", timeNowUnix-3*U.SECONDS_IN_A_DAY)
	createEventWithTimestampByName(t, project, user, "event5", timeNowUnix-3*U.SECONDS_IN_A_DAY)
	// 4 days older events.
	createEventWithTimestampByName(t, project, user, "event6", timeNowUnix-4*U.SECONDS_IN_A_DAY)

	// Should be empty for todays startTime.
	datesEventCountMap, status := M.GetDatesForNextEventsArchivalBatch(project.ID, timeNowUnix)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 0, len(datesEventCountMap))

	// Must return 3 days.
	datesEventCountMap, status = M.GetDatesForNextEventsArchivalBatch(project.ID, timeNowUnix-10*U.SECONDS_IN_A_DAY)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 3, len(datesEventCountMap))
	expectedDateCount := map[string]int64{
		timeNow.AddDate(0, 0, -1).Format(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN): 2, // 1 day before.
		timeNow.AddDate(0, 0, -3).Format(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN): 2, // 3 days before.
		timeNow.AddDate(0, 0, -4).Format(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN): 1, // 4 days before.
	}

	for expectedDate, expectedCount := range expectedDateCount {
		value, found := datesEventCountMap[expectedDate]
		assert.True(t, found)
		assert.Equal(t, expectedCount, value)
	}
}
