package model

import (
	C "factors/config"
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

const error_Duplicate_event_customerEventID = "pq: duplicate key value violates unique constraint \"project_id_customer_event_id_unique_idx\""

func isDuplicateCustomerEventIdError(err error) bool {
	return err.Error() == error_Duplicate_event_customerEventID
}

type EventTimestamp struct {
	First int64
	Last  int64
}

func (event *Event) BeforeCreate(scope *gorm.Scope) error {
	db := C.GetServices().Db

	// Increamenting count based on EventNameId, not by EventName.
	var count uint64
	if err := db.Model(&Event{}).Where("project_id = ? AND user_id = ? AND event_name_id = ?",
		event.ProjectId, event.UserId, event.EventNameId).Count(&count).Error; err != nil {
		return err
	}
	event.Count = count + 1
	if event.Timestamp <= 0 {
		event.Timestamp = time.Now().Unix()
	}
	return nil
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

	if err := db.Create(event).Error; err != nil {
		log.WithFields(log.Fields{"event": &event, "error": err}).Error("CreateEvent Failed")
		if isDuplicateCustomerEventIdError(err) {
			log.WithError(err).Info("CreateEvent Failed, duplicate customerEventId")
			return nil, http.StatusFound
		}
		return nil, http.StatusInternalServerError
	}
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

	rows, err := db.Raw("SELECT project_id, min(timestamp) as first_timestamp, max(timestamp) as last_timestamp FROM events GROUP BY project_id").Rows()
	if err != nil {
		log.Error("Failed to get events timestamp info.")
		return nil, http.StatusInternalServerError
	}
	defer rows.Close()

	projectEventsTime := make(map[uint64]*EventTimestamp, 0)

	count := 0
	for rows.Next() {
		var projectId uint64
		var firstTimestamp, lastTimestamp int64
		if err = rows.Scan(&projectId, &firstTimestamp, &lastTimestamp); err != nil {
			return nil, http.StatusInternalServerError
		}

		if firstTimestamp > 0 {
			projectEventsTime[projectId] = &EventTimestamp{First: firstTimestamp, Last: lastTimestamp}
		}

		count++
	}
	if count == 0 {
		return nil, http.StatusNotFound
	}

	return &projectEventsTime, http.StatusFound
}
