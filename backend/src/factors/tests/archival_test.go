package tests

import (
	"testing"

	A "factors/archival"
	U "factors/util"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeEventProperties(t *testing.T) {
	eventProperties := make(map[string]interface{})
	for _, property := range U.SDK_ALLOWED_EVENT_PROPERTIES {
		eventProperties[property] = U.RandomString(10)
	}
	eventProperties = A.SanitizeEventProperties(eventProperties)

	// Blacklisted properties must have been removed.
	for _, property := range A.ARCHIVE_BLACKLISTED_EP {
		_, okay := eventProperties[property]
		assert.False(t, okay)
	}

	// Rest all properties must still be present.
	for _, property := range U.StringSliceDiff(U.SDK_ALLOWED_EVENT_PROPERTIES[:], A.ARCHIVE_BLACKLISTED_EP) {
		_, okay := eventProperties[property]
		assert.True(t, okay)
	}
}

func TestSanitizeUserProperties(t *testing.T) {
	userProperties := make(map[string]interface{})
	for _, property := range U.SDK_ALLOWED_USER_PROPERTIES {
		userProperties[property] = U.RandomString(10)
	}
	userProperties = A.SanitizeUserProperties(userProperties)

	// Blacklisted properties must have been removed.
	for _, property := range A.ARCHIVE_BLACKLISTED_UP {
		_, okay := userProperties[property]
		assert.False(t, okay)
	}

	// Rest all properties must still be present.
	for _, property := range U.StringSliceDiff(U.SDK_ALLOWED_USER_PROPERTIES[:], A.ARCHIVE_BLACKLISTED_UP) {
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
	batches, _ := A.GetNextArchivalBatches(project.ID, startTime, maxLookbackDays)
	assert.Empty(t, batches)

	// Corner case of beginning of the day timestamp. Empty batches must be returned.
	startTime = U.GetBeginningOfDayTimestampUTC(U.TimeNowUnix())
	batches, _ = A.GetNextArchivalBatches(project.ID, startTime, maxLookbackDays)
	assert.Empty(t, batches)

	// For a day before, only one entry.
	startTime = U.GetBeginningOfDayTimestampUTC(U.TimeNowUnix()) - U.SECONDS_IN_A_DAY
	batches, _ = A.GetNextArchivalBatches(project.ID, startTime, maxLookbackDays)
	assert.Equal(t, 1, len(batches))

	// When startTime is older than maxLookbackDays.
	startTime = U.GetBeginningOfDayTimestampUTC(U.TimeNowUnix()) - 10*U.SECONDS_IN_A_DAY
	maxLookbackDays = 5
	batches, _ = A.GetNextArchivalBatches(project.ID, startTime, maxLookbackDays)
	assert.Equal(t, maxLookbackDays, len(batches))
	effectiveStartTime := U.GetBeginningOfDayTimestampUTC(U.TimeNowUnix()) - 5*U.SECONDS_IN_A_DAY

	// Batches must be in sorted order of dates.
	for index, batch := range batches {
		expectedStartTime := effectiveStartTime + int64(index)*U.SECONDS_IN_A_DAY
		expectedEndTime := effectiveStartTime + int64(index+1)*U.SECONDS_IN_A_DAY - 1
		assert.Equal(t, expectedStartTime, batch.StartTime)
		assert.Equal(t, expectedEndTime, batch.EndTime)
	}
}
