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

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndGetEvent(t *testing.T) {
	// Initialize a project, user and  the event.
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	start := time.Now()

	// Test successful CreateEvent.
	event, errCode := M.CreateEvent(&M.Event{EventNameId: eventNameId, ProjectId: projectId, UserId: userId})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, len(event.ID) > 30)
	assert.Equal(t, projectId, event.ProjectId)
	assert.Equal(t, eventNameId, event.EventNameId)
	assert.Equal(t, uint64(1), event.Count)
	assert.True(t, event.Timestamp >= start.Unix())
	assert.InDelta(t, event.Timestamp, start.Unix(), 3)
	assert.True(t, event.CreatedAt.After(start))
	assert.True(t, event.UpdatedAt.After(start))
	assert.Equal(t, event.CreatedAt, event.UpdatedAt)
	assert.Equal(t, postgres.Jsonb{RawMessage: json.RawMessage(nil)}, event.Properties)
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
	assert.Equal(t, event, retEvent)
	// Test Get Event with wrong project id.
	retEvent, errCode = M.GetEvent(projectId+1, userId, event.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)
	// Test Get Event with wrong user id.
	retEvent, errCode = M.GetEvent(projectId, "randomId", event.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)
	// Test successful CreateEvent with count increment
	event, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId, ProjectId: projectId, UserId: userId})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, len(event.ID) > 30)
	assert.Equal(t, projectId, event.ProjectId)
	assert.Equal(t, eventNameId, event.EventNameId)
	assert.Equal(t, uint64(2), event.Count)
	assert.True(t, event.Timestamp >= start.Unix())
	assert.InDelta(t, event.Timestamp, start.Unix(), 3)
	assert.True(t, event.CreatedAt.After(start))
	assert.True(t, event.UpdatedAt.After(start))
	assert.Equal(t, event.CreatedAt, event.UpdatedAt)
	assert.Equal(t, postgres.Jsonb{RawMessage: json.RawMessage(nil)}, event.Properties)

	t.Run("DuplicateCustomerEventId", func(t *testing.T) {
		custEventId := U.RandomString(8)
		//projectId, userId, eventNameId, err := SetupProjectUserEventName()
		assert.Nil(t, err)

		event, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId, ProjectId: projectId, UserId: userId, CustomerEventId: &custEventId})
		assert.Equal(t, http.StatusCreated, errCode)
		_, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId, ProjectId: projectId, UserId: userId, CustomerEventId: &custEventId})
		assert.Equal(t, http.StatusFound, errCode)
	})

	// Test Get Event on non existent id.
	retEvent, errCode = M.GetEvent(projectId, userId, "9ad21963-bcfb-4563-aa02-8ea589710d1a")
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)

	// Test Create Event with properties.
	properties := json.RawMessage(`{"email": "random@example.com"}`)
	event, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId, ProjectId: projectId, UserId: userId, Properties: postgres.Jsonb{properties}})
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

	// Test Create Event with id.
	randomId := "random_id"
	event, errCode = M.CreateEvent(&M.Event{ID: randomId, EventNameId: eventNameId, ProjectId: projectId, UserId: userId})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, event)

	// Test Create Event without projectId.
	event, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId, UserId: userId})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, event)

	// Test Create Event without userId.
	event, errCode = M.CreateEvent(&M.Event{EventNameId: eventNameId, ProjectId: projectId})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, event)

	// Test Create Event without eventNameId.
	event, errCode = M.CreateEvent(&M.Event{EventNameId: 0, ProjectId: projectId, UserId: userId})
	assert.Equal(t, http.StatusInternalServerError, errCode)
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
	ts1, errCode := M.GetProjectEventTimeInfo()
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, ts1)
	assert.NotNil(t, (*ts1)[project.ID])
	assert.Equal(t, firstTimestamp, (*ts1)[project.ID].FirstEvent)
	assert.Equal(t, thirdTimestamp, (*ts1)[project.ID].LastEvent)

	// Test with increased limit timestamp
	ts1, errCode = M.GetProjectEventTimeInfo() // adds 3 secs.
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, ts1)
	assert.NotNil(t, (*ts1)[project.ID])
	assert.Equal(t, firstTimestamp, (*ts1)[project.ID].FirstEvent)
	assert.Equal(t, thirdTimestamp, (*ts1)[project.ID].LastEvent)
	assert.Nil(t, (*ts1)[999999])
}

func createEventWithTimestampAndPrperties(t *testing.T, project *M.Project, user *M.User, timestamp int64, properties json.RawMessage) (*M.EventName, *M.Event) {
	eventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: fmt.Sprintf("event_%d", timestamp)})
	assert.NotNil(t, eventName)
	event, errCode := M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID, Timestamp: timestamp, Properties: postgres.Jsonb{properties}})
	assert.Equal(t, http.StatusCreated, errCode)
	return eventName, event
}

func TestGetRecentEventPropertyKeys(t *testing.T) {
	project, user, _ := SetupProjectUserReturnDAO()
	assert.NotNil(t, project)

	t.Run("RecentPropertiesWithLimit", func(t *testing.T) {
		timestamp := time.Now().Unix()
		eventName, _ := createEventWithTimestampAndPrperties(t, project, user, timestamp, json.RawMessage(`{"rProp1": "value1", "rProp2": "1"}`))
		_, errCode1 := M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID, Timestamp: timestamp, Properties: postgres.Jsonb{json.RawMessage(`{"rProp3": "value2", "rProp4": "2"}`)}})
		assert.Equal(t, http.StatusCreated, errCode1)

		props, errCode := M.GetRecentEventPropertyKeysWithLimits(project.ID, eventName.Name, 1)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Contains(t, props, U.PropertyTypeNumerical)
		assert.Contains(t, props, U.PropertyTypeNumerical)
		assert.Len(t, props[U.PropertyTypeCategorical], 1)
		assert.Len(t, props[U.PropertyTypeNumerical], 1)
		// validates classification.
		assert.Contains(t, props[U.PropertyTypeCategorical], "rProp1")
		assert.Contains(t, props[U.PropertyTypeNumerical], "rProp2")
		// validates limit.
		assert.NotContains(t, props[U.PropertyTypeCategorical], "rProp3")
		assert.NotContains(t, props[U.PropertyTypeNumerical], "rProp4")
	})

	t.Run("PropertiesOlderThan24Hours", func(t *testing.T) {
		timestamp := U.UnixTimeBefore24Hours()
		eventName, _ := createEventWithTimestampAndPrperties(t, project, user, timestamp, json.RawMessage(`{"rProp1": "value1", "rProp2": "1"}`))

		props, errCode := M.GetRecentEventPropertyKeysWithLimits(project.ID, eventName.Name, 100)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Contains(t, props, U.PropertyTypeNumerical)
		assert.Contains(t, props, U.PropertyTypeNumerical)
		assert.Len(t, props[U.PropertyTypeCategorical], 0)
		assert.Len(t, props[U.PropertyTypeNumerical], 0)
	})
}

func TestGetRecentEventPropertyValues(t *testing.T) {
	project, user, _ := SetupProjectUserReturnDAO()
	assert.NotNil(t, project)

	t.Run("RecentPropertyValues", func(t *testing.T) {
		timestamp := time.Now().Unix()
		eventName, _ := createEventWithTimestampAndPrperties(t, project, user, timestamp, json.RawMessage(`{"rProp1": "value1"}`))
		_, errCode1 := M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID, Timestamp: timestamp, Properties: postgres.Jsonb{json.RawMessage(`{"rProp1": "value2"}`)}})
		assert.Equal(t, http.StatusCreated, errCode1)

		// limited events to 1.
		values, errCode2 := M.GetRecentEventPropertyValuesWithLimits(project.ID, eventName.Name, "rProp1", 1, 100)
		assert.Equal(t, http.StatusFound, errCode2)
		assert.Len(t, values, 1)
		assert.Contains(t, values, "value2")

		// limited values to 1.
		values1, errCode3 := M.GetRecentEventPropertyValuesWithLimits(project.ID, eventName.Name, "rProp1", 10, 1)
		assert.Equal(t, http.StatusFound, errCode3)
		assert.Len(t, values1, 1)
		assert.Contains(t, values1, "value1")
	})

	t.Run("PropertyValuesOlderThan24Hour", func(t *testing.T) {
		timestamp := U.UnixTimeBefore24Hours()
		eventName, _ := createEventWithTimestampAndPrperties(t, project, user, timestamp, json.RawMessage(`{"rProp1": "value1"}`))
		_, errCode1 := M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID, Timestamp: timestamp, Properties: postgres.Jsonb{json.RawMessage(`{"rProp1": "value2"}`)}})
		assert.Equal(t, http.StatusCreated, errCode1)

		values, errCode2 := M.GetRecentEventPropertyValuesWithLimits(project.ID, eventName.Name, "rProp1", 100, 100)
		assert.Equal(t, http.StatusFound, errCode2)
		assert.Empty(t, values)
	})
}
