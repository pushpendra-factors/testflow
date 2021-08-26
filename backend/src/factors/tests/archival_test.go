package tests

import (
	"testing"
	"time"

	"factors/model/model"
	"factors/model/store"
	U "factors/util"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeEventProperties(t *testing.T) {
	eventProperties := make(map[string]interface{})
	for _, property := range U.SDK_ALLOWED_EVENT_PROPERTIES {
		eventProperties[property] = U.RandomString(10)
	}
	eventProperties = model.SanitizeEventProperties(eventProperties)

	// Blacklisted properties must have been removed.
	for _, property := range U.DISABLED_CORE_QUERY_EVENT_PROPERTIES {
		_, okay := eventProperties[property]
		assert.False(t, okay)
	}

	// Rest all properties must still be present.
	for _, property := range U.StringSliceDiff(U.SDK_ALLOWED_EVENT_PROPERTIES[:], U.DISABLED_CORE_QUERY_EVENT_PROPERTIES[:]) {
		_, okay := eventProperties[property]
		assert.True(t, okay)
	}
}

func TestSanitizeUserProperties(t *testing.T) {
	userProperties := make(map[string]interface{})
	for _, property := range U.SDK_ALLOWED_USER_PROPERTIES {
		userProperties[property] = U.RandomString(10)
	}
	userProperties = model.SanitizeUserProperties(userProperties)

	// Blacklisted properties must have been removed.
	for _, property := range U.DISABLED_CORE_QUERY_USER_PROPERTIES {
		_, okay := userProperties[property]
		assert.False(t, okay)
	}

	// Rest all properties must still be present.
	for _, property := range U.StringSliceDiff(U.SDK_ALLOWED_USER_PROPERTIES[:], U.DISABLED_CORE_QUERY_USER_PROPERTIES[:]) {
		_, okay := userProperties[property]
		assert.True(t, okay)
	}
}

func TestGetNextArchivalBatches(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	maxLookbackDays := 365

	// Should not return any batches for the same day.
	startTime := U.TimeNowUnix()
	batches, _ := store.GetStore().GetNextArchivalBatches(project.ID, startTime, maxLookbackDays, time.Time{}, time.Time{})
	assert.Empty(t, batches)

	// Corner case of beginning of the day timestamp. Empty batches must be returned.
	startTime = U.GetBeginningOfDayTimestampZ(U.TimeNowUnix())
	batches, _ = store.GetStore().GetNextArchivalBatches(project.ID, startTime, maxLookbackDays, time.Time{}, time.Time{})
	assert.Empty(t, batches)

	// For a day before, only one entry.
	startTime = U.GetBeginningOfDayTimestampZ(U.TimeNowUnix()) - U.SECONDS_IN_A_DAY
	batches, _ = store.GetStore().GetNextArchivalBatches(project.ID, startTime, maxLookbackDays, time.Time{}, time.Time{})
	assert.Equal(t, 1, len(batches))

	// When startTime is older than maxLookbackDays.
	startTime = U.GetBeginningOfDayTimestampZ(U.TimeNowUnix()) - 10*U.SECONDS_IN_A_DAY
	maxLookbackDays = 5
	batches, _ = store.GetStore().GetNextArchivalBatches(project.ID, startTime, maxLookbackDays, time.Time{}, time.Time{})
	assert.Equal(t, maxLookbackDays, len(batches))
	effectiveStartTime := U.GetBeginningOfDayTimestampZ(U.TimeNowUnix()) - 5*U.SECONDS_IN_A_DAY

	// Batches must be in sorted order of dates.
	for index, batch := range batches {
		expectedStartTime := effectiveStartTime + int64(index)*U.SECONDS_IN_A_DAY
		expectedEndTime := effectiveStartTime + int64(index+1)*U.SECONDS_IN_A_DAY - 1
		assert.Equal(t, expectedStartTime, batch.StartTime)
		assert.Equal(t, expectedEndTime, batch.EndTime)
	}
}
