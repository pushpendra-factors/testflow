package model

import (
	"factors/util"
	"time"
)

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
	for _, blackListedEP := range util.DISABLED_CORE_QUERY_EVENT_PROPERTIES {
		delete(eventProperties, blackListedEP)
	}
	return eventProperties
}

// SanitizeUserProperties Sanitizes user properties making ready for archival / bigquery file.
func SanitizeUserProperties(userProperties map[string]interface{}) map[string]interface{} {
	for _, blackListedUP := range util.DISABLED_CORE_QUERY_USER_PROPERTIES {
		delete(userProperties, blackListedUP)
	}
	return userProperties
}
