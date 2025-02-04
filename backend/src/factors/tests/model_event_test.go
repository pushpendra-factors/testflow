package tests

import (
	"database/sql"
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	"factors/task/event_user_cache"

	pCache "factors/cache/persistent"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestPullEventRowsV2(t *testing.T) {
	// Initialize a project, user and  the event.
	start := time.Now()
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	// Test successful CreateEvent.
	newEvent := &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: 101,
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "1😄"}`)}}
	event, errCode := store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, event.CreatedAt.After(start))

	newEvent = &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: 102,
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "2😄"}`)}}
	event, errCode = store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, event.CreatedAt.After(start))

	newEvent = &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: 103,
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "3😄"}`)}}
	event, errCode = store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, event.CreatedAt.After(start))

	newEvent = &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: 104,
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "4😄"}`)}}
	event, errCode = store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, event.CreatedAt.After(start))

	newEvent = &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: 105,
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "5😄"}`)}}
	event, errCode = store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, event.CreatedAt.After(start))
	assert.True(t, time.Now().Unix() >= start.Unix())
	rows, tx, err := store.GetStore().PullEventRowsV2(projectId, start.Unix()+19800, time.Now().Unix()+19800+2, false, 0)
	assert.Nil(t, err)

	numRows := 0
	for rows.Next() {
		var userID string
		var eventName string
		var eventTimestamp int64
		var userJoinTimestamp int64
		var eventCardinality uint
		var eventProperties *postgres.Jsonb
		var userProperties *postgres.Jsonb
		var is_group_user_null sql.NullBool
		var group_1_user_id_null sql.NullString
		var group_2_user_id_null sql.NullString
		var group_3_user_id_null sql.NullString
		var group_4_user_id_null sql.NullString
		var group_5_user_id_null sql.NullString
		var group_6_user_id_null sql.NullString
		var group_7_user_id_null sql.NullString
		var group_8_user_id_null sql.NullString
		var group_1_id_null sql.NullString
		var group_2_id_null sql.NullString
		var group_3_id_null sql.NullString
		var group_4_id_null sql.NullString
		var group_5_id_null sql.NullString
		var group_6_id_null sql.NullString
		var group_7_id_null sql.NullString
		var group_8_id_null sql.NullString
		err = rows.Scan(&userID, &eventName, &eventTimestamp, &eventCardinality, &eventProperties, &userJoinTimestamp, &userProperties,
			&is_group_user_null, &group_1_user_id_null, &group_2_user_id_null, &group_3_user_id_null, &group_4_user_id_null,
			&group_5_user_id_null, &group_6_user_id_null, &group_7_user_id_null, &group_8_user_id_null,
			&group_1_id_null, &group_2_id_null, &group_3_id_null, &group_4_id_null,
			&group_5_id_null, &group_6_id_null, &group_7_id_null, &group_8_id_null)
		assert.Nil(t, err)

		assert.InDelta(t, 103, eventTimestamp, 3)
		numRows++
	}

	err = rows.Err()
	assert.Nil(t, err)
	U.CloseReadQuery(rows, tx)
	assert.Equal(t, 5, numRows)
}

func TestDBCreateAndGetEvent(t *testing.T) {
	// Initialize a project, user and  the event.
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	start := time.Now()

	// Test successful CreateEvent.
	newEvent := &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: start.Unix(),
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "The Impact of Using Emojis 😄 😍 💗 in Push Notifications"}`)}}
	event, errCode := store.GetStore().CreateEvent(newEvent)
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
	user, errCode := store.GetStore().GetUser(projectId, userId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, event.Timestamp, user.LastEventAt.Unix())

	// Test Get Event on the created.
	retEvent, errCode := store.GetStore().GetEvent(projectId, userId, event.ID)
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
	timezoneString := U.TimeZoneStringUTC
	if C.IsIngestionTimezoneEnabled(event.ProjectId) {
		var statusCode int
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(event.ProjectId)
		assert.Equal(t, http.StatusFound, statusCode)
	}
	timeWithTimezone := U.ConvertTimeIn(time.Unix(event.Timestamp, 0), timezoneString)
	assert.True(t, (*eventProperties)["$day_of_week"] != "" && (*eventProperties)["$day_of_week"] == timeWithTimezone.Weekday().String())
	hr, _, _ := timeWithTimezone.Clock()
	assert.True(t, (*eventProperties)["$hour_of_day"] != 0 && (*eventProperties)["$hour_of_day"] == float64(hr))
	assert.True(t, (*eventProperties)["$timestamp"].(float64) != 0 && (*eventProperties)["$timestamp"].(float64) == float64(event.Timestamp))

	// Test Get Event with wrong project id.
	retEvent, errCode = store.GetStore().GetEvent(projectId+1, userId, event.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)
	// Test Get Event with wrong user id.
	retEvent, errCode = store.GetStore().GetEvent(projectId, "randomId", event.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)
	// Test successful CreateEvent with count increment
	event, errCode = store.GetStore().CreateEvent(&model.Event{EventNameId: eventNameId,
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
	timezoneString = U.TimeZoneStringUTC
	if C.IsIngestionTimezoneEnabled(event.ProjectId) {
		var statusCode int
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(event.ProjectId)
		assert.Equal(t, http.StatusFound, statusCode)
	}
	timeWithTimezone = U.ConvertTimeIn(time.Unix(event.Timestamp, 0), timezoneString)
	assert.True(t, (*eventProperties)["$day_of_week"] != "" && (*eventProperties)["$day_of_week"] == timeWithTimezone.Weekday().String())
	hr, _, _ = timeWithTimezone.Clock()
	assert.True(t, (*eventProperties)["$hour_of_day"] != 0 && (*eventProperties)["$hour_of_day"] == float64(hr))
	assert.True(t, (*eventProperties)["$timestamp"].(float64) != 0 && (*eventProperties)["$timestamp"].(float64) == float64(event.Timestamp))

	t.Run("DuplicateCustomerEventId", func(t *testing.T) {
		custEventId := U.RandomString(8)

		event, errCode = store.GetStore().CreateEvent(&model.Event{EventNameId: eventNameId, ProjectId: projectId,
			UserId: userId, CustomerEventId: &custEventId, Timestamp: time.Now().Unix()})
		assert.Equal(t, http.StatusCreated, errCode)
		_, errCode = store.GetStore().CreateEvent(&model.Event{EventNameId: eventNameId, ProjectId: projectId,
			UserId: userId, CustomerEventId: &custEventId, Timestamp: time.Now().Unix()})
		assert.Equal(t, http.StatusNotAcceptable, errCode)
	})

	// Test Get Event on non existent id.
	retEvent, errCode = store.GetStore().GetEvent(projectId, userId, "9ad21963-bcfb-4563-aa02-8ea589710d1a")
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)

	// Test Create Event with properties.
	properties := json.RawMessage(`{"email": "random@example.com"}`)
	event, errCode = store.GetStore().CreateEvent(&model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Properties: postgres.Jsonb{RawMessage: properties}, Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, event)
	assert.NotEqual(t, postgres.Jsonb{RawMessage: json.RawMessage(nil)}, event.Properties)
	// Retrieve and validate properties.
	retEvent, errCode = store.GetStore().GetEvent(projectId, userId, event.ID)
	eventPropertiesBytes, err := retEvent.Properties.Value()
	var eventPropertiesMap map[string]interface{}
	json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
	assert.Equal(t, "random@example.com", eventPropertiesMap["email"])
	// Test Get Event on wrong format of id.
	retEvent, errCode = store.GetStore().GetEvent(projectId, userId, "r4nd0m!234")
	assert.Equal(t, http.StatusInternalServerError, errCode)
	assert.Nil(t, retEvent)

	// Test Create Event with external id.
	eventId := uuid.New().String()
	event, errCode = store.GetStore().CreateEvent(&model.Event{ID: eventId, EventNameId: eventNameId,
		ProjectId: projectId, UserId: userId, Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, event)
	assert.Equal(t, eventId, event.ID)

	// Test Create Event with invalid uuid as id.
	eventId = U.RandomLowerAphaNumString(10)
	event, errCode = store.GetStore().CreateEvent(&model.Event{ID: eventId, EventNameId: eventNameId,
		ProjectId: projectId, UserId: userId, Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusInternalServerError, errCode)
	assert.Nil(t, event)

	// Test Create Event without projectId.
	event, errCode = store.GetStore().CreateEvent(&model.Event{EventNameId: eventNameId, UserId: userId,
		Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, event)

	// Test Create Event without userId.
	event, errCode = store.GetStore().CreateEvent(&model.Event{EventNameId: eventNameId,
		ProjectId: projectId, Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, event)

	// Test Create Event without eventNameId.
	event, errCode = store.GetStore().CreateEvent(&model.Event{EventNameId: "", ProjectId: projectId, UserId: userId,
		Timestamp: time.Now().Unix()})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, event)
}

func TestWeekOfDay(t *testing.T) {
	// Initialize a project, user and  the event.
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	// Test successful CreateEvent.
	newEvent := &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: 1641078000,
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "The Impact of Using Emojis 😄 😍 💗 in Push Notifications"}`)}}
	event, errCode := store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)

	assert.True(t, event.Timestamp != 0)
	assert.Equal(t, event.Timestamp, event.PropertiesUpdatedTimestamp)
	eventProperties, _ := U.DecodePostgresJsonb(&event.Properties)
	timezoneString := U.TimeZoneStringUTC
	if C.IsIngestionTimezoneEnabled(event.ProjectId) {
		var statusCode int
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(event.ProjectId)
		assert.Equal(t, http.StatusFound, statusCode)
	}
	timeWithTimezone := U.ConvertTimeIn(time.Unix(event.Timestamp, 0), timezoneString)
	assert.True(t, (*eventProperties)["$day_of_week"] != "" && (*eventProperties)["$day_of_week"] == timeWithTimezone.Weekday().String())
	hr, _, _ := timeWithTimezone.Clock()
	assert.True(t, (*eventProperties)["$hour_of_day"] != 0 && (*eventProperties)["$hour_of_day"] == float64(hr))
}

func createEventWithTimestampAndPrperties(t *testing.T, project *model.Project, user *model.User, timestamp int64, properties json.RawMessage) (*model.EventName, *model.Event) {
	eventName, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: fmt.Sprintf("event_%d", timestamp)})
	assert.NotNil(t, eventName)
	event, errCode := store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID,
		Timestamp: timestamp, Properties: postgres.Jsonb{RawMessage: properties}})
	assert.Equal(t, http.StatusCreated, errCode)
	eventProperties, _ := U.DecodePostgresJsonb(&event.Properties)
	assert.True(t, (*eventProperties)["$timestamp"].(float64) != 0 && (*eventProperties)["$timestamp"].(float64) == float64(event.Timestamp))
	return eventName, event
}

func TestGetRecentEventPropertyKeys(t *testing.T) {
	project, user, _ := SetupProjectUserReturnDAO()
	assert.NotNil(t, project)

	t.Run("RecentPropertiesWithLimit", func(t *testing.T) {
		timestamp := time.Now().Unix()
		eventName, _ := createEventWithTimestampAndPrperties(t, project, user, timestamp, json.RawMessage(`{"rProp1": "value1", "rProp2": 1}`))
		_, errCode1 := store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID,
			Timestamp: timestamp, Properties: postgres.Jsonb{json.RawMessage(`{"rProp3": "value2", "rProp4": 2}`)}})
		assert.Equal(t, http.StatusCreated, errCode1)
		props, err := store.GetStore().GetRecentEventPropertyKeysWithLimits(project.ID, eventName.Name, time.Now().AddDate(0, 0, -30).Unix(), timestamp, 100)
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

		props, err := store.GetStore().GetRecentEventPropertyKeysWithLimits(project.ID, eventName.Name, time.Now().AddDate(0, 0, -1).Unix(), timestamp, 100)
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
		_, errCode1 := store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID,
			Timestamp: timestamp, Properties: postgres.Jsonb{RawMessage: json.RawMessage(`{"rProp1": "value2"}`)}})
		assert.Equal(t, http.StatusCreated, errCode1)
		_, errCode1 = store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID,
			Timestamp: timestamp, Properties: postgres.Jsonb{RawMessage: json.RawMessage(`{"rProp1": "value1", "rProp2": 1}`)}})
		assert.Equal(t, http.StatusCreated, errCode1)

		values, category, err := store.GetStore().GetRecentEventPropertyValuesWithLimits(project.ID, eventName.Name, "rProp1", 10, 100, time.Now().AddDate(0, 0, -1).Unix(), timestamp)
		assert.Equal(t, nil, err)
		assert.Len(t, values, 2)
		assert.Equal(t, values[0].Value, "value1")
		assert.Equal(t, values[1].Value, "value2")
		assert.Equal(t, category, U.PropertyTypeCategorical)

		// limited values to 1.
		values1, category, err := store.GetStore().GetRecentEventPropertyValuesWithLimits(project.ID, eventName.Name, "rProp1", 1, 100, time.Now().AddDate(0, 0, -30).Unix(), timestamp)
		assert.Equal(t, nil, err)
		assert.Len(t, values1, 1)
		assert.Equal(t, values1[0].Value, "value1")
		assert.Equal(t, category, U.PropertyTypeCategorical)

		values2, category, err := store.GetStore().GetRecentEventPropertyValuesWithLimits(project.ID, eventName.Name, "rProp2", 10, 100, time.Now().AddDate(0, 0, -30).Unix(), timestamp)
		assert.Equal(t, nil, err)
		assert.Len(t, values2, 1)
		assert.Equal(t, values2[0].Value, "1")
		assert.Equal(t, category, U.PropertyTypeNumerical)
	})

	t.Run("PropertyValuesOlderThan24Hour", func(t *testing.T) {
		timestamp := U.UnixTimeBeforeDuration(24 * time.Hour)
		eventName, _ := createEventWithTimestampAndPrperties(t, project, user, timestamp, json.RawMessage(`{"rProp1": "value1"}`))
		event, errCode1 := store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID,
			Timestamp: timestamp, Properties: postgres.Jsonb{RawMessage: json.RawMessage(`{"rProp1": "value2"}`)}})
		assert.Equal(t, http.StatusCreated, errCode1)
		eventProperties, _ := U.DecodePostgresJsonb(&event.Properties)
		assert.True(t, (*eventProperties)["$timestamp"].(float64) != 0 && (*eventProperties)["$timestamp"].(float64) == float64(event.Timestamp))

		values, category, err := store.GetStore().GetRecentEventPropertyValuesWithLimits(project.ID, eventName.Name, "rProp1", 100, 100, time.Now().AddDate(0, 0, -1).Unix(), timestamp)
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

	properties := &U.PropertiesMap{
		"$page_spent_time": 1.346, "$page_load_time": 1.594, "$page_scroll_percent": 97.54}
	// should add properties if not exist.
	errCode := store.GetStore().UpdateEventProperties(project.ID, event.ID, event.UserId, properties, time.Now().Unix(), nil)

	assert.Equal(t, http.StatusAccepted, errCode)
	updatedEvent, errCode := store.GetStore().GetEventById(project.ID, event.ID, event.UserId)
	assert.Equal(t, http.StatusFound, errCode)
	eventProperties, err := U.DecodePostgresJsonb(&updatedEvent.Properties)
	assert.Nil(t, err)
	assert.Contains(t, *eventProperties, "$page_spent_time")
	assert.Contains(t, *eventProperties, "$page_load_time")
	assert.Contains(t, *eventProperties, "$page_scroll_percent")
	assert.Contains(t, *eventProperties, "rProp1") // should not remove old properties.
	// values should be unchanged.
	assert.Equal(t, float64(1.594), (*eventProperties)["$page_load_time"])
	assert.Equal(t, float64(1.346), (*eventProperties)["$page_spent_time"])
	assert.Equal(t, "value1", (*eventProperties)["rProp1"])
	assert.LessOrEqual(t, (*eventProperties)["$page_scroll_percent"], float64(100)) // value must be <= 100

	// should update properties if exist.
	errCode = store.GetStore().UpdateEventProperties(project.ID, event.ID, event.UserId, &U.PropertiesMap{
		"$page_spent_time": 3, "$page_scroll_percent": 150.87}, time.Now().Unix(), nil)
	assert.Equal(t, http.StatusAccepted, errCode)
	updatedEvent, errCode = store.GetStore().GetEventById(project.ID, event.ID, event.UserId)
	assert.Equal(t, http.StatusFound, errCode)
	eventProperties, _ = U.DecodePostgresJsonb(&updatedEvent.Properties)
	assert.Contains(t, *eventProperties, "$page_spent_time")
	assert.Contains(t, *eventProperties, "$page_load_time")
	assert.Contains(t, *eventProperties, "rProp1") // should not remove old properties.
	// should update the property alone.
	assert.Equal(t, float64(3), (*eventProperties)["$page_spent_time"])
	assert.Equal(t, float64(1.594), (*eventProperties)["$page_load_time"])
	assert.Equal(t, "value1", (*eventProperties)["rProp1"])
	assert.LessOrEqual(t, (*eventProperties)["$page_scroll_percent"], float64(100)) // value must be <= 100

	// should update properties if exist.
	errCode = store.GetStore().UpdateEventProperties(project.ID, event.ID, event.UserId, &U.PropertiesMap{
		"$page_spent_time": 5, "$page_scroll_percent": "207.98"}, time.Now().Unix(), nil)
	assert.Equal(t, http.StatusAccepted, errCode)
	updatedEvent, errCode = store.GetStore().GetEventById(project.ID, event.ID, event.UserId)
	assert.Equal(t, http.StatusFound, errCode)
	eventProperties, _ = U.DecodePostgresJsonb(&updatedEvent.Properties)
	assert.Contains(t, *eventProperties, "$page_spent_time")
	assert.Contains(t, *eventProperties, "$page_load_time")
	assert.Contains(t, *eventProperties, "rProp1") // should not remove old properties.
	// should update the property alone.
	assert.Equal(t, float64(5), (*eventProperties)["$page_spent_time"])
	assert.Equal(t, float64(1.594), (*eventProperties)["$page_load_time"])
	assert.Equal(t, "value1", (*eventProperties)["rProp1"])
	assert.LessOrEqual(t, (*eventProperties)["$page_scroll_percent"], float64(100)) // value must be <= 100
}

func TestCacheEvent(t *testing.T) {
	for i := 0; i < 30; i++ {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.Nil(t, err)

		eventId := U.RandomString(10)
		timestamp := time.Now().Unix() - 100
		err = model.SetCacheUserLastEvent(project.ID, user.ID,
			&model.CacheEvent{ID: eventId, Timestamp: timestamp})
		assert.Nil(t, err)

		cacheEvent, err := model.GetCacheUserLastEvent(project.ID, user.ID)
		assert.NotNil(t, cacheEvent)
		assert.Equal(t, eventId, cacheEvent.ID)
		assert.Equal(t, timestamp, cacheEvent.Timestamp)
		assert.Nil(t, err)

		event, errCode := store.GetStore().CreateEvent(&model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: user.ID, Timestamp: time.Now().Unix()})
		assert.Equal(t, http.StatusCreated, errCode)
		eventProperties, _ := U.DecodePostgresJsonb(&event.Properties)
		assert.True(t, (*eventProperties)["$timestamp"].(float64) != 0 && (*eventProperties)["$timestamp"].(float64) == float64(event.Timestamp))

		cacheEvent, err = model.GetCacheUserLastEvent(project.ID, user.ID)
		assert.NotNil(t, cacheEvent)
		assert.Equal(t, event.ID, cacheEvent.ID)
		assert.True(t, cacheEvent.Timestamp > timestamp)
		assert.Nil(t, err)
	}
}

func TestOverwriteEventProperties(t *testing.T) {
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	event, errCode := store.GetStore().CreateEvent(&model.Event{EventNameId: eventNameId,
		ProjectId: projectId, UserId: userId, Timestamp: time.Now().Unix()})

	eventPropertiesMap, err := U.DecodePostgresJsonb(&event.Properties)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Nil(t, (*eventPropertiesMap)["Hello"])
	(*eventPropertiesMap)["Hello"] = "World"

	eventPropertiesJSONb, err := U.EncodeToPostgresJsonb(eventPropertiesMap)
	assert.Nil(t, err)

	errCode = store.GetStore().OverwriteEventProperties(projectId, userId, event.ID, eventPropertiesJSONb)
	assert.Equal(t, errCode, http.StatusAccepted)

	rEvent, errCode := store.GetStore().GetEvent(projectId, userId, event.ID)
	assert.Equal(t, http.StatusFound, errCode)

	rEventPropertiesMap, err := U.DecodePostgresJsonb(&rEvent.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*rEventPropertiesMap)["Hello"], "World")
}

func TestGetEventCountOfUserByEventName(t *testing.T) {
	project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
	assert.Nil(t, err)

	// No events, should return found and zero count.
	count, errCode := store.GetStore().GetEventCountOfUserByEventName(project.ID, user.ID, eventName.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, uint64(0), count)

	newEvent := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
		UserId: user.ID, Timestamp: time.Now().Unix()}
	event, errCode := store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, event)

	count, errCode = store.GetStore().GetEventCountOfUserByEventName(project.ID, user.ID, eventName.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, uint64(1), count)
}

func TestGetDatesForNextEventsArchivalBatch(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	userID, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, userID)
	assert.Equal(t, http.StatusCreated, status)

	timeNow := U.TimeNowZ()
	timeNowUnix := timeNow.Unix()

	user, errCode := store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, errCode)

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
	datesEventCountMap, status := store.GetStore().GetDatesForNextEventsArchivalBatch(project.ID, timeNowUnix)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 0, len(datesEventCountMap))

	// Must return 3 days.
	datesEventCountMap, status = store.GetStore().GetDatesForNextEventsArchivalBatch(project.ID,
		timeNowUnix-10*U.SECONDS_IN_A_DAY)
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

func TestDeleteEventByIDs(t *testing.T) {
	// Initialize a project, user and  the event.
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	start := time.Now()
	newEvent := &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: start.Unix()}
	event, errCode := store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = store.GetStore().DeleteEventByIDs(projectId, eventNameId, []string{event.ID})
	assert.Equal(t, http.StatusAccepted, errCode)

	_, errCode = store.GetStore().GetEvent(projectId, userId, event.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
}

func TestGetUserEventsByEventNameId(t *testing.T) {
	// Initialize a project, user and  the event.
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	event1Timestamp := time.Now().Unix()
	event1 := &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: event1Timestamp}
	_, errCode := store.GetStore().CreateEvent(event1)
	assert.Equal(t, http.StatusCreated, errCode)

	event2Timestamp := event1Timestamp + 1000
	event2 := &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: event2Timestamp}
	_, errCode = store.GetStore().CreateEvent(event2)
	assert.Equal(t, http.StatusCreated, errCode)

	event3Timestamp := event2Timestamp + 1000
	event3 := &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: event3Timestamp}
	_, errCode = store.GetStore().CreateEvent(event3)
	assert.Equal(t, http.StatusCreated, errCode)

	events, errCode := store.GetStore().GetUserEventsByEventNameId(projectId, userId, eventNameId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Len(t, events, 3)
	assert.Equal(t, event3.ID, events[0].ID)
	assert.Equal(t, event2.ID, events[1].ID)
	assert.Equal(t, event1.ID, events[2].ID)
	assert.Greater(t, events[0].Timestamp, events[1].Timestamp)
	assert.Greater(t, events[1].Timestamp, events[2].Timestamp)
}

func TestGetEventByIdWithoutEventAndUserProperties(t *testing.T) {
	// Initialize a project, user and  the event.
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	start := time.Now()
	newEvent := &model.Event{EventNameId: eventNameId, ProjectId: projectId,
		UserId: userId, Timestamp: start.Unix()}
	event, errCode := store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)

	event, errCode = store.GetStore().GetEventById(projectId, event.ID, userId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, event.Properties)
	assert.NotNil(t, event.UserProperties)

	eventID, userID, errCode := store.GetStore().GetUserIdFromEventId(projectId, event.ID, "")
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, event.ID, eventID)
	assert.Equal(t, newEvent.UserId, userID)

	eventID, userID, errCode = store.GetStore().GetUserIdFromEventId(projectId, event.ID, userId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, event.ID, eventID)
	assert.Equal(t, newEvent.UserId, userID)
}

func TestPrependEvent(t *testing.T) {
	e1 := model.Event{ID: "e1"}
	e2 := model.Event{ID: "e2"}

	events := make([]model.Event, 0, 0)
	events = append(events, e1, e2)

	e3 := model.Event{ID: "e3"}

	events = model.PrependEvent(e3, events)
	assert.Equal(t, "e3", events[0].ID)
	assert.Equal(t, "e1", events[1].ID)
	assert.Equal(t, "e2", events[2].ID)
}

func TestGetLatestTimestampByEventNameId(t *testing.T) {
	// Initialize a project, user and  the event.
	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	start := time.Now()

	for i := 0; i < 5; i++ {

		start = start.AddDate(0, 0, 1)

		newEvent := &model.Event{EventNameId: eventNameId, ProjectId: projectId,
			UserId: userId, Timestamp: start.Unix(),
			Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "The Impact of Using Emojis 😄 😍 💗 in Push Notifications"}`)}}

		_, errCode := store.GetStore().CreateEvent(newEvent)
		assert.Equal(t, http.StatusCreated, errCode)

	}

	latestTimestamp, errCode := store.GetStore().GetLatestEventTimeStampByEventNameId(projectId, eventNameId, start.Unix()-10*U.SECONDS_IN_A_DAY, start.Unix()+10*U.SECONDS_IN_A_DAY)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, start.Unix(), latestTimestamp)

	latestTimestamp1, errCode := store.GetStore().GetLatestEventTimeStampByEventNameId(projectId, eventNameId, start.Unix()+20*U.SECONDS_IN_A_DAY, start.Unix()+30*U.SECONDS_IN_A_DAY)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Equal(t, int64(0), latestTimestamp1)

}

func TestIsOTPKeyUniqueWithQuery(t *testing.T) {
	// Initialize a project, user and  the event.
	projectId, _, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	start := time.Now()

	createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: projectId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("OTPKeyIsUnique", func(t *testing.T) {
		ruleID := U.RandomLowerAphaNumString(6)
		eventID := U.RandomLowerAphaNumString(6)
		otpUniqueKey := createdUserID + ruleID + eventID

		newEvent := &model.Event{EventNameId: eventNameId, ProjectId: projectId,
			UserId: createdUserID, Timestamp: start.Unix(),
			Properties: postgres.Jsonb{RawMessage: []byte(fmt.Sprintf(`{"$otp_unique_key":123}`))}}

		_, errCode = store.GetStore().CreateEvent(newEvent)
		assert.Equal(t, http.StatusCreated, errCode)

		isUnique, _ := store.GetStore().IsOTPKeyUniqueWithQuery(projectId, createdUserID, eventNameId, otpUniqueKey)
		assert.Equal(t, true, isUnique)
	})

	t.Run("OTPKeyIsNotUnique", func(t *testing.T) {
		ruleID := U.RandomLowerAphaNumString(6)
		eventID := U.RandomLowerAphaNumString(6)
		otpUniqueKey := createdUserID + ruleID + eventID

		newEvent := &model.Event{EventNameId: eventNameId, ProjectId: projectId,
			UserId: createdUserID, Timestamp: start.Unix(),
			Properties: postgres.Jsonb{RawMessage: []byte(fmt.Sprintf(`{"%s":"%s"}`, U.EP_OTP_UNIQUE_KEY, otpUniqueKey))}}

		_, errCode = store.GetStore().CreateEvent(newEvent)
		assert.Equal(t, http.StatusCreated, errCode)

		isUnique, _ := store.GetStore().IsOTPKeyUniqueWithQuery(projectId, createdUserID, eventNameId, otpUniqueKey)
		assert.Equal(t, false, isUnique)
	})

}

func TestEvaluateRulesV1(t *testing.T) {

	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	var eventPropertiesMap map[string]interface{}
	err = json.Unmarshal([]byte(`{"$salesforce_campaignmember_firstrespondeddate": "100",
									"$salesforce_campaignmember_lastmodifieddate": "100",
										"$salesforce_campaignmember_createddate": "100",
											"$salesforce_campaign_name":"Webinar",
												"$salesforce_campaignmember_hasresponded":true,
													"$salesforce_campaignmember_status":"Attended Live"}`), &eventPropertiesMap)
	assert.Equal(t, err, nil)

	_, status := store.GetStore().CreateOrGetOfflineTouchPointEventName(project.ID)
	assert.Equal(t, http.StatusCreated, status)

	t.Run("WithCorrectSetOfRule", func(t *testing.T) {
		sfEvent := model.EventIdToProperties{
			Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN,
			EventProperties: eventPropertiesMap,
		}

		filter1 := model.TouchPointFilter{
			Property:  "$salesforce_campaign_name",
			Operator:  "contains",
			Value:     "Webinar",
			LogicalOp: "AND",
		}
		filter2 := model.TouchPointFilter{
			LogicalOp: "AND",
			Operator:  "equals",
			Property:  "$salesforce_campaignmember_status",
			Value:     "Attended Live",
		}
		filter3 := model.TouchPointFilter{
			LogicalOp: "OR",
			Operator:  "equals",
			Property:  "$salesforce_campaignmember_status",
			Value:     "Attended On-demand",
		}
		filter4 := model.TouchPointFilter{
			LogicalOp: "OR",
			Operator:  "equals",
			Property:  "$salesforce_campaignmember_status",
			Value:     "Attended Live and on-Demand",
		}

		rulePropertyMap := make(map[string]model.TouchPointPropertyValue)
		rulePropertyMap["$campaign"] = model.TouchPointPropertyValue{Type: model.TouchPointPropertyValueAsProperty, Value: "$salesforce_campaignmember_campaignname"}
		rulePropertyMap["$channel"] = model.TouchPointPropertyValue{Type: model.TouchPointPropertyValueAsConstant, Value: "Other"}

		f, _ := json.Marshal([]model.TouchPointFilter{filter1, filter2, filter3, filter4})
		rPM, _ := json.Marshal(rulePropertyMap)

		rule := model.OTPRule{
			Filters:           postgres.Jsonb{json.RawMessage(f)},
			TouchPointTimeRef: model.SFCampaignMemberResponded, // SFCampaignMemberResponded
			PropertiesMap:     postgres.Jsonb{json.RawMessage(rPM)},
		}

		rule.ID = U.RandomLowerAphaNumString(5)

		passed := model.EvaluateOTPFilterV1(rule, sfEvent, nil)
		assert.Equal(t, true, passed)

	})

	t.Run("WithWrongSetOfRule", func(t *testing.T) {
		sfEvent := model.EventIdToProperties{
			Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN,
			EventProperties: eventPropertiesMap,
		}

		filter1 := model.TouchPointFilter{
			Property:  "$salesforce_campaign_name",
			Operator:  "contains",
			Value:     "Event",
			LogicalOp: "AND",
		}
		filter2 := model.TouchPointFilter{
			LogicalOp: "AND",
			Operator:  "equals",
			Property:  "$salesforce_campaignmember_status",
			Value:     "Attended Live",
		}
		filter3 := model.TouchPointFilter{
			LogicalOp: "OR",
			Operator:  "equals",
			Property:  "$salesforce_campaignmember_status",
			Value:     "Attended On-demand",
		}
		filter4 := model.TouchPointFilter{
			LogicalOp: "OR",
			Operator:  "equals",
			Property:  "$salesforce_campaignmember_status",
			Value:     "Attended Live and on-Demand",
		}

		rulePropertyMap := make(map[string]model.TouchPointPropertyValue)
		rulePropertyMap["$campaign"] = model.TouchPointPropertyValue{Type: model.TouchPointPropertyValueAsProperty, Value: "$salesforce_campaignmember_campaignname"}
		rulePropertyMap["$channel"] = model.TouchPointPropertyValue{Type: model.TouchPointPropertyValueAsConstant, Value: "Other"}

		f, _ := json.Marshal([]model.TouchPointFilter{filter1, filter2, filter3, filter4})
		rPM, _ := json.Marshal(rulePropertyMap)

		rule := model.OTPRule{
			Filters:           postgres.Jsonb{json.RawMessage(f)},
			TouchPointTimeRef: model.SFCampaignMemberResponded, // SFCampaignMemberResponded
			PropertiesMap:     postgres.Jsonb{json.RawMessage(rPM)},
		}

		rule.ID = U.RandomLowerAphaNumString(5)

		rule.ID = U.RandomLowerAphaNumString(5)

		passed := model.EvaluateOTPFilterV1(rule, sfEvent, nil)
		assert.Equal(t, false, passed)

	})
}

func TestGetDomainNamesByProjectID(t *testing.T) {

	projectId, userId, eventNameId, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	newEvent := &model.Event{EventNameId: eventNameId, ProjectId: projectId, UserId: userId, Timestamp: time.Now().Unix(),
		Properties: postgres.Jsonb{RawMessage: []byte(fmt.Sprintf(`{"$is_page_view":true , "$page_url":"app.factors.ai/accounts"}`))}}

	_, errCode := store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)

	newEvent = &model.Event{EventNameId: eventNameId, ProjectId: projectId, UserId: userId, Timestamp: time.Now().Unix(),
		Properties: postgres.Jsonb{RawMessage: []byte(fmt.Sprintf(`{"$is_page_view":true , "$page_url":"app.factors.ai/analyze"}`))}}

	_, errCode = store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)

	newEvent = &model.Event{EventNameId: eventNameId, ProjectId: projectId, UserId: userId, Timestamp: time.Now().Unix(),
		Properties: postgres.Jsonb{RawMessage: []byte(fmt.Sprintf(`{"$is_page_view":true , "$page_url":"app.xyz.ai/accounts"}`))}}

	_, errCode = store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)

	domains, errCode := store.GetStore().GetDomainNamesByProjectID(projectId)
	assert.Equal(t, len(domains), 2)
	assert.True(t, U.ContainsStringInArray(domains, "app.factors.ai"))
	assert.True(t, U.ContainsStringInArray(domains, "app.xyz.ai"))
	assert.Equal(t, http.StatusFound, errCode)
}

func TestEventPropertyValuesAggregate(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	// Initialize a project, user and  the event.
	project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
	assert.Nil(t, err)
	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+1343545")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	configs := make(map[string]interface{})
	configs["rollupLookback"] = 4
	configs["deleteRollupAfterAddingToAggregate"] = 1

	start := time.Now()

	newEvent := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
		UserId: user.ID, Timestamp: start.Unix(),
		CreatedAt:  time.Now().AddDate(0, 0, -3),
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "value1"}`)}}
	store.GetStore().AddEventDetailsToCacheWithTime(project.ID, newEvent, false, time.Now().AddDate(0, 0, -3))

	// Run rollup.
	event_user_cache.DoRollUpSortedSet(configs)

	eventPropertyValuesAggCacheKey, err := model.GetValuesByEventPropertyRollUpAggregateCacheKey(project.ID, eventName.Name, "value")
	assert.Nil(t, err)
	existingAggCache, aggCacheExists, err := pCache.GetIfExists(eventPropertyValuesAggCacheKey, true)
	assert.Nil(t, err)
	assert.True(t, aggCacheExists)
	var existingAggregate U.CacheEventPropertyValuesAggregate
	err = json.Unmarshal([]byte(existingAggCache), &existingAggregate)
	assert.Nil(t, err)
	assert.Equal(t, "value1", existingAggregate.NameCountTimestampCategoryList[0].Name)
	assert.Equal(t, int64(1), existingAggregate.NameCountTimestampCategoryList[0].Count)

	newEvent = &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
		UserId: user.ID, Timestamp: start.Unix(),
		CreatedAt:  time.Now().AddDate(0, 0, -2),
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "value1"}`)}}
	store.GetStore().AddEventDetailsToCacheWithTime(project.ID, newEvent, false, time.Now().AddDate(0, 0, -2))

	event_user_cache.DoRollUpSortedSet(configs)

	eventPropertyValuesAggCacheKey, err = model.GetValuesByEventPropertyRollUpAggregateCacheKey(project.ID, eventName.Name, "value")
	assert.Nil(t, err)
	existingAggCache, aggCacheExists, err = pCache.GetIfExists(eventPropertyValuesAggCacheKey, true)
	assert.Nil(t, err)
	assert.True(t, aggCacheExists)
	err = json.Unmarshal([]byte(existingAggCache), &existingAggregate)
	assert.Nil(t, err)
	assert.Equal(t, "value1", existingAggregate.NameCountTimestampCategoryList[0].Name)
	assert.Equal(t, int64(2), existingAggregate.NameCountTimestampCategoryList[0].Count)

	newEvent = &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
		UserId: user.ID, Timestamp: start.Unix(),
		CreatedAt:  time.Now().AddDate(0, 0, -1),
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "value1"}`)}}
	store.GetStore().AddEventDetailsToCacheWithTime(project.ID, newEvent, false, time.Now().AddDate(0, 0, -1))

	event_user_cache.DoRollUpSortedSet(configs)

	eventPropertyValuesAggCacheKey, err = model.GetValuesByEventPropertyRollUpAggregateCacheKey(project.ID, eventName.Name, "value")
	assert.Nil(t, err)
	existingAggCache, aggCacheExists, err = pCache.GetIfExists(eventPropertyValuesAggCacheKey, true)
	assert.Nil(t, err)
	assert.True(t, aggCacheExists)
	err = json.Unmarshal([]byte(existingAggCache), &existingAggregate)
	assert.Nil(t, err)
	assert.Equal(t, "value1", existingAggregate.NameCountTimestampCategoryList[0].Name)
	assert.Equal(t, int64(3), existingAggregate.NameCountTimestampCategoryList[0].Count)

	// Current day event.
	newEvent = &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
		UserId: user.ID, Timestamp: start.Unix(),
		CreatedAt:  time.Now(),
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "value1"}`)}}
	store.GetStore().AddEventDetailsToCacheWithTime(project.ID, newEvent, false, time.Now())

	event_user_cache.DoRollUpSortedSet(configs)

	eventPropertyValuesAggCacheKey, err = model.GetValuesByEventPropertyRollUpAggregateCacheKey(project.ID, eventName.Name, "value")
	assert.Nil(t, err)
	existingAggCache, aggCacheExists, err = pCache.GetIfExists(eventPropertyValuesAggCacheKey, true)
	assert.Nil(t, err)
	assert.True(t, aggCacheExists)
	err = json.Unmarshal([]byte(existingAggCache), &existingAggregate)
	assert.Nil(t, err)
	assert.Equal(t, "value1", existingAggregate.NameCountTimestampCategoryList[0].Name)
	assert.Equal(t, int64(3), existingAggregate.NameCountTimestampCategoryList[0].Count)

	// Validate even name property values through api.
	w := sendGetEventPropertyValues(project.ID, "login", "value", false, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)
	var propertyValues []string
	jsonResponse, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	json.Unmarshal(jsonResponse, &propertyValues)
	assert.Equal(t, "value1", propertyValues[0])
}

func TestServingFromEventPropertyValueRollupsOnTheSameDay(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	// Initialize a project, user and  the event.
	project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
	assert.Nil(t, err)
	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+1343545")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	configs := make(map[string]interface{})
	configs["rollupLookback"] = 3

	start := time.Now()

	newEvent := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
		UserId: user.ID, Timestamp: start.Unix(),
		CreatedAt:  time.Now(),
		Properties: postgres.Jsonb{RawMessage: []byte(`{"value": "value1"}`)}}
	event, errCode := store.GetStore().CreateEvent(newEvent)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, event)

	event_user_cache.DoRollUpSortedSet(configs)

	// Validate even name property values through api.
	w := sendGetEventPropertyValues(project.ID, "login", "value", false, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)
	var propertyValues []string
	jsonResponse, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	json.Unmarshal(jsonResponse, &propertyValues)
	assert.Equal(t, "value1", propertyValues[0])
}
