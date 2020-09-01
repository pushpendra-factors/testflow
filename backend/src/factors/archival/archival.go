package archival

import (
	"fmt"
	"net/http"
	"time"

	M "factors/model"
	U "factors/util"
)

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

// ArchiveUsersTableFormat Users table schema in bigquery.
type ArchiveUsersTableFormat struct {
	UserID         string    `json:"user_id" bigquery:"user_id"`
	CustomerUserID string    `json:"identified_user_id" bigquery:"identified_user_id"`
	IngestionDate  time.Time `json:"ingestion_date" bigquery:"ingestion_date"`
}

func (eventFormat ArchiveEventTableFormat) GetEventTimestamp() int64 {
	return eventFormat.EventTimestamp.Unix()
}

func (eventFormat ArchiveEventTableFormat) GetEventTimestampColumnName() string {
	return "event_timestamp"
}

func (usersFormat ArchiveUsersTableFormat) GetEventTimestamp() int64 {
	return usersFormat.IngestionDate.Unix()
}

func (usersFormat ArchiveUsersTableFormat) GetEventTimestampColumnName() string {
	return "ingestion_date"
}

// SanitizeEventProperties Sanitizes event properites making ready for archival / bigquery file.
func SanitizeEventProperties(eventProperties map[string]interface{}) map[string]interface{} {
	for _, blackListedEP := range U.DISABLED_CORE_QUERY_EVENT_PROPERTIES {
		delete(eventProperties, blackListedEP)
	}
	return eventProperties
}

// SanitizeUserProperties Sanitizes user properties making ready for archival / bigquery file.
func SanitizeUserProperties(userProperties map[string]interface{}) map[string]interface{} {
	for _, blackListedUP := range U.DISABLED_CORE_QUERY_USER_PROPERTIES {
		delete(userProperties, blackListedUP)
	}
	return userProperties
}

// GetNextArchivalBatches Returns a list of EventsArchivalBatch from the give startTime to 1 day before.
func GetNextArchivalBatches(projectID uint64, startTime int64, maxLookbackDays int, hardStartTime, hardEndTime time.Time) ([]EventsArchivalBatch, error) {
	var eventsArchivalBatches []EventsArchivalBatch
	var completedBatches map[int64]int64
	endTime := time.Unix(U.GetBeginningOfDayTimestampUTC(U.TimeNowUnix())-1, 0)

	maxLookBackTime := U.GetBeginningOfDayTimestampUTC(U.TimeNow().AddDate(0, 0, -maxLookbackDays).Unix())
	if !hardStartTime.IsZero() {
		startTime = U.GetBeginningOfDayTimestampUTC(hardStartTime.Unix())
		endTime = hardEndTime.Add(time.Second * time.Duration(1))
		var status int
		completedBatches, status = M.GetCompletedArchivalBatches(projectID, hardStartTime, hardEndTime)
		if status == http.StatusInternalServerError {
			return eventsArchivalBatches, fmt.Errorf("Failed to get completed batches")
		}
	} else if maxLookBackTime > startTime {
		// If startTime older than maxLookback allowed, set it to oldest allowed date.
		startTime = maxLookBackTime
	}

	if startTime > endTime.Unix() {
		return eventsArchivalBatches, fmt.Errorf("Invalid startTime value %v for endTime %v", startTime, endTime)
	}

	batchTime := time.Unix(startTime, 0).UTC()
	for batchTime.Before(endTime) {
		nextBatchTime := batchTime.AddDate(0, 0, 1)
		batchStartTime := U.GetBeginningOfDayTimestampUTC(batchTime.Unix())
		batchEndTime := U.GetBeginningOfDayTimestampUTC(nextBatchTime.Unix()) - 1
		if !hardStartTime.IsZero() {
			_, found := completedBatches[batchStartTime]
			if found {
				// Start date part of completed batches. Skip adding to the new batches.
				batchTime = nextBatchTime
				continue
			}
		}
		eventsArchivalBatches = append(eventsArchivalBatches, EventsArchivalBatch{
			StartTime: batchStartTime,
			EndTime:   batchEndTime,
		})

		batchTime = nextBatchTime
	}

	return eventsArchivalBatches, nil
}
