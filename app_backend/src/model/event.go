package model

import (
	C "config"
	"net/http"
	"time"

	_ "github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	// Composite primary key with project_id and uuid.
	ID string `gorm:"primary_key:true;type:uuid;default:uuid_generate_v4()" json:"id"`

	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	// (project_id, user_id) -> users(project_id, id)
	// (project_id, event_name) -> events(project_id, name)
	ProjectId uint64 `gorm:"primary_key:true;" json:"project_id"`
	UserId    string `json:"user_id"`
	EventName string `json:"event_name"`

	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Properties postgres.Jsonb `json:"properties,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func CreateEvent(event *Event) (*Event, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"event": &event}).Info("Creating event")

	// Input Validation. (ID is to be auto generated)
	if event.ID != "" {
		log.Error("CreateEvent Failed. Id provided.")
		return nil, http.StatusBadRequest
	}

	if err := db.Create(event).Error; err != nil {
		log.WithFields(log.Fields{"event": &event, "error": err}).Error("CreateEvent Failed")
		return nil, http.StatusInternalServerError
	} else {
		return event, DB_SUCCESS
	}
}

func GetEvent(projectId uint64, userId string, id string) (*Event, int) {
	db := C.GetServices().Db

	var event Event
	if err := db.Where("id = ?", id).Where("project_id = ?", projectId).Where("user_id = ?", userId).First(&event).Error; err != nil {
		return nil, 404
	} else {
		return &event, DB_SUCCESS
	}
}
