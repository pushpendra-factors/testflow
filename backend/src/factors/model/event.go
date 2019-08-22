package model

import (
	C "factors/config"
	U "factors/util"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	// Composite primary key with project_id and uuid.
	ID              string  `gorm:"primary_key:true;type:uuid;default:uuid_generate_v4()" json:"id"`
	CustomerEventId *string `json:"customer_event_id"`

	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	// (project_id, user_id) -> users(project_id, id)
	// (project_id, event_name_id) -> event_names(project_id, id)
	ProjectId        uint64 `gorm:"primary_key:true;" json:"project_id"`
	UserId           string `json:"user_id"`
	UserPropertiesId string `json:"user_properties_id"`
	EventNameId      uint64 `json:"event_name_id"`
	Count            uint64 `json:"count"`
	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Properties postgres.Jsonb `json:"properties,omitempty"`
	// unix epoch timestamp in seconds.
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EventTimestamp struct {
	FirstEvent  int64
	LastEvent   int64
	ProjectName string
}

type EventOccurrence struct {
	EventNameId uint64
	Count       int
}

const error_Duplicate_event_customerEventID = "pq: duplicate key value violates unique constraint \"project_id_customer_event_id_unique_idx\""
const eventsLimitForProperites = 50000

func isDuplicateCustomerEventIdError(err error) bool {
	return err.Error() == error_Duplicate_event_customerEventID
}

func CreateEvent(event *Event) (*Event, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"event": &event}).Info("Creating event")

	// Input Validation. (ID is to be auto generated)
	if event.ID != "" {
		log.Error("CreateEvent Failed. Id provided.")
		return nil, http.StatusBadRequest
	}

	if event.ProjectId == 0 || event.UserId == "" {
		log.Error("CreateEvent Failed. Invalid projectId or userId.")
		return nil, http.StatusBadRequest
	}

	// Increamenting count based on EventNameId, not by EventName.
	var count uint64
	if err := db.Model(&Event{}).Where("project_id = ? AND user_id = ? AND event_name_id = ?",
		event.ProjectId, event.UserId, event.EventNameId).Count(&count).Error; err != nil {
		return nil, http.StatusInternalServerError
	}
	event.Count = count + 1

	if event.Timestamp <= 0 {
		event.Timestamp = time.Now().Unix()
	}

	transTime := gorm.NowFunc()
	rows, err := db.Raw("INSERT INTO events (customer_event_id,project_id,user_id,user_properties_id,event_name_id,count,properties,timestamp,created_at,updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING events.id",
		event.CustomerEventId, event.ProjectId, event.UserId, event.UserPropertiesId, event.EventNameId, event.Count, event.Properties, event.Timestamp, transTime, transTime).Rows()
	if err != nil {
		if isDuplicateCustomerEventIdError(err) {
			log.WithError(err).Info("CreateEvent Failed, duplicate customerEventId")
			return nil, http.StatusFound
		}

		log.WithFields(log.Fields{"event": &event}).WithError(err).Error("CreateEvent Failed")
		return nil, http.StatusInternalServerError
	}

	var eventId string
	for rows.Next() {
		if err = rows.Scan(&eventId); err != nil {
			log.WithError(err).Error("CreateEvent Failed. Failed to read event id.")
			return nil, http.StatusInternalServerError
		}
	}
	event.ID = eventId
	event.CreatedAt = transTime
	event.UpdatedAt = transTime

	return event, http.StatusCreated
}

func GetEvent(projectId uint64, userId string, id string) (*Event, int) {
	db := C.GetServices().Db

	var event Event
	if err := db.Where("id = ?", id).Where("project_id = ?", projectId).Where("user_id = ?", userId).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &event, http.StatusFound
}

func GetEventById(projectId uint64, id string) (*Event, int) {
	db := C.GetServices().Db

	var event Event
	if err := db.Where("project_id = ?", projectId).Where("id = ?", id).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &event, http.StatusFound
}

func GetProjectEventTimeInfo() (*(map[uint64]*EventTimestamp), int) {
	db := C.GetServices().Db

	rows, err := db.Raw("SELECT projects.id, projects.name, min(events.timestamp) as first_timestamp, max(events.timestamp) as last_timestamp FROM events LEFT JOIN projects on events.project_id = projects.id GROUP BY projects.id").Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get events timestamp info.")
		return nil, http.StatusInternalServerError
	}
	defer rows.Close()

	projectEventsTime := make(map[uint64]*EventTimestamp, 0)

	count := 0
	for rows.Next() {
		var projectId uint64
		var firstTimestamp, lastTimestamp int64
		var projectName string
		if err = rows.Scan(&projectId, &projectName, &firstTimestamp, &lastTimestamp); err != nil {
			return nil, http.StatusInternalServerError
		}

		if firstTimestamp > 0 {
			projectEventsTime[projectId] = &EventTimestamp{
				FirstEvent: firstTimestamp, LastEvent: lastTimestamp, ProjectName: projectName}
		}

		count++
	}
	if count == 0 {
		return nil, http.StatusNotFound
	}

	return &projectEventsTime, http.StatusFound
}

// GetRecentEventPropertyKeys - Returns unique event property
// keys from last 24 hours.
func GetRecentEventPropertyKeysWithLimits(projectId uint64, eventName string, eventsLimit int) (map[string][]string, int) {
	db := C.GetServices().Db

	eventsAfterTimestamp := U.UnixTimeBefore24Hours()
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "events_after_timestamp": eventsAfterTimestamp})

	queryStr := "SELECT distinct(properties) AS keys FROM events WHERE project_id = ?" +
		" " + "AND event_name_id IN (SELECT id FROM event_names WHERE project_id = ? AND name = ?)" +
		" " + "AND timestamp > ? AND properties != 'null' LIMIT ?"

	rows, err := db.Raw(queryStr, projectId, projectId, eventName, eventsAfterTimestamp, eventsLimit).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get event properties.")
		return nil, http.StatusInternalServerError
	}
	defer rows.Close()

	propertiesMap := make(map[string]map[interface{}]bool, 0)
	for rows.Next() {
		var propertiesJson []byte
		rows.Scan(&propertiesJson)

		err := U.FillPropertyKvsFromPropertiesJson(propertiesJson, &propertiesMap, U.SamplePropertyValuesLimit)
		if err != nil {
			log.WithError(err).WithField("properties_json",
				string(propertiesJson)).Error("Failed to unmarshal json properties.")
			return nil, http.StatusInternalServerError
		}
	}

	err = rows.Err()
	if err != nil {
		logCtx.WithError(err).Error("Failed to scan recent property keys.")
		return nil, http.StatusInternalServerError
	}

	propsByType, err := U.ClassifyPropertiesByType(&propertiesMap)
	if err != nil {
		logCtx.WithError(err).Error("Failed to classify properties on get recent property keys.")
		return nil, http.StatusInternalServerError
	}

	return propsByType, http.StatusFound
}

func GetRecentEventPropertyKeys(projectId uint64, eventName string) (map[string][]string, int) {
	return GetRecentEventPropertyKeysWithLimits(projectId, eventName, eventsLimitForProperites)
}

// GetRecentEventPropertyValues - Returns unique event property
// values of given property from last 24 hours.
func GetRecentEventPropertyValuesWithLimits(projectId uint64, eventName string,
	property string, eventsLimit, valuesLimit int) ([]string, int) {

	db := C.GetServices().Db

	logCtx := log.WithFields(log.Fields{"projectId": projectId, "eventName": eventName, "property": property})

	eventsAfterTimestamp := U.UnixTimeBefore24Hours()
	values := make([]string, 0, 0)
	queryStr := "SELECT DISTINCT(value) FROM" +
		" " + "(SELECT properties->? AS value FROM events WHERE project_id = ? AND event_name_id IN" +
		" " + "(SELECT id FROM event_names WHERE project_id = ? AND name = ?) AND timestamp > ? AND properties->? IS NOT NULL LIMIT ?)" +
		" " + "AS property_values LIMIT ?"

	rows, err := db.Raw(queryStr, property, projectId, projectId, eventName,
		eventsAfterTimestamp, property, eventsLimit, valuesLimit).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get recent property keys.")
		return values, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var value string
		rows.Scan(&value)
		value = U.TrimQuotes(value)
		values = append(values, value)
	}

	err = rows.Err()
	if err != nil {
		logCtx.WithError(err).Error("Failed scanning property value on type classifcation.")
		return values, http.StatusInternalServerError
	}
	return values, http.StatusFound
}

func GetRecentEventPropertyValues(projectId uint64, eventName string, property string) ([]string, int) {
	return GetRecentEventPropertyValuesWithLimits(projectId, eventName, property, eventsLimitForProperites, 2000)
}

func GetEventsOccurrenceCount(projectId uint64) ([]EventOccurrence, int) {
	db := C.GetServices().Db

	eventsAfterTimestamp := U.UnixTimeBeforeAWeek()

	logCtx := log.WithFields(log.Fields{"projectId": projectId, "eventsAfterTimestamp": eventsAfterTimestamp})
	queryStr := "SELECT event_name_id, COUNT(*) FROM events WHERE project_id=? AND timestamp > ?" +
		" " + "GROUP BY event_name_id ORDER BY count DESC LIMIT ?"

	eventsOccurrence := make([]EventOccurrence, 0, 0)
	rows, err := db.Raw(queryStr, projectId, eventsAfterTimestamp, 100000).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to read rows on get events occurrence count.")
		return eventsOccurrence, http.StatusInternalServerError
	}

	for rows.Next() {
		var eventNameId uint64
		var count int
		if err := rows.Scan(&eventNameId, &count); err != nil {
			logCtx.WithError(err).Error("Failed to read rows on get events occurrence count.")
			return eventsOccurrence, http.StatusInternalServerError
		}

		eventsOccurrence = append(eventsOccurrence,
			EventOccurrence{EventNameId: eventNameId, Count: count})
	}

	return eventsOccurrence, http.StatusFound
}
