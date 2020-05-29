package archival

import (
	"time"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

// ARCHIVE_BLACKLISTED_EP Standard properties to be removed from event properties for archival.
var ARCHIVE_BLACKLISTED_EP []string = []string{
	U.EP_INTERNAL_IP,
	U.EP_LOCATION_LATITUDE,
	U.EP_LOCATION_LONGITUDE,
	U.EP_SEGMENT_EVENT_VERSION,
}

// ARCHIVE_BLACKLISTED_UP Standard properties to be removed from user properties for archival.
var ARCHIVE_BLACKLISTED_UP []string = []string{
	U.UP_DEVICE_ADTRACKING_ENABLED,
	U.UP_NETWORK_BLUETOOTH,
	U.UP_NETWORK_CARRIER,
	U.UP_NETWORK_CELLULAR,
	U.UP_NETWORK_WIFI,
	U.UP_SEGMENT_CHANNEL,
	U.UP_DEVICE_ADVERTISING_ID,
	U.UP_DEVICE_ID,
}

// EventsArchivalBatch Object to store the Events Archival batch data.
// Each batch will be scheduled as a separate task.
type EventsArchivalBatch struct {
	StartTime int64
	EndTime   int64
}

// EventFileFormat Generic interface for allowing multiple FileFormats to be written to file.
type EventFileFormat interface {
	GetEventTimestamp() int64
	GetEventTimestampColumnName() string
}

// ArchiveEventTableFormat Format to write files for archival job which gets pushed to Bigquery.
type ArchiveEventTableFormat struct {
	// TODO(prateek): Finalize and update schema.
	EventID           string    `json:"event_id" bigquery:"event_id"`
	UserID            string    `json:"user_id" bigquery:"user_id"`
	UserJoinTimestamp int64     `json:"user_join_timestamp" bigquery:"user_join_timestamp"`
	EventName         string    `json:"event_name" bigquery:"event_name"`
	EventTimestamp    time.Time `json:"event_timestamp" bigquery:"event_timestamp"`
	SessionID         string    `json:"session_id" bigquery:"session_id"`
	EventProperties   string    `json:"event_properties" bigquery:"event_properties"`
	UserProperties    string    `json:"user_properties" bigquery:"user_properties"`
}

func (eventFormat ArchiveEventTableFormat) GetEventTimestamp() int64 {
	return eventFormat.EventTimestamp.Unix()
}

func (eventFormat ArchiveEventTableFormat) GetEventTimestampColumnName() string {
	return "event_timestamp"
}

// SanitizeEventProperties Sanitizes event properites making ready for archival / bigquery file.
func SanitizeEventProperties(eventProperties map[string]interface{}) map[string]interface{} {
	for _, blackListedEP := range ARCHIVE_BLACKLISTED_EP {
		delete(eventProperties, blackListedEP)
	}
	return eventProperties
}

// SanitizeUserProperties Sanitizes user properties making ready for archival / bigquery file.
func SanitizeUserProperties(userProperties map[string]interface{}) map[string]interface{} {
	for _, blackListedUP := range ARCHIVE_BLACKLISTED_UP {
		delete(userProperties, blackListedUP)
	}
	return userProperties
}

// GetNextArchivalBatches Returns a list of EventsArchivalBatch from the give startTime to 1 day before.
func GetNextArchivalBatches(projectID uint64, startTime int64, maxLookbackDays int) []EventsArchivalBatch {
	var eventsArchivalBatches []EventsArchivalBatch
	logCtx := log.WithFields(log.Fields{
		"Prefix":    "Archival#GetNextArchivalBatches",
		"ProjectID": projectID,
	})

	maxLookBackTime := U.GetBeginningOfDayTimestampUTC(U.TimeNow().AddDate(0, 0, -maxLookbackDays).Unix())
	if maxLookBackTime > startTime {
		// If startTime older than maxLookback allowed, set it to oldest allowed date.
		startTime = maxLookBackTime
	}

	endTime := U.TimeNow()
	endDate := endTime.Format(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN)

	if startTime > endTime.Unix() {
		logCtx.Errorf("Invalid startTime value %v", startTime)
		return eventsArchivalBatches
	}

	batchTime := time.Unix(startTime, 0).UTC()
	batchDate := batchTime.Format(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN)

	for batchDate != endDate {
		nextBatchTime := batchTime.AddDate(0, 0, 1)
		eventsArchivalBatches = append(eventsArchivalBatches, EventsArchivalBatch{
			StartTime: U.GetBeginningOfDayTimestampUTC(batchTime.Unix()),
			EndTime:   U.GetBeginningOfDayTimestampUTC(nextBatchTime.Unix()) - 1,
		})

		batchTime = nextBatchTime
		batchDate = batchTime.Format(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN)
	}

	return eventsArchivalBatches
}
