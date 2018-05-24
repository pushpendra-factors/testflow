package model

import (
	C "config"
	"net/http"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type EventName struct {
	// Composite primary key with projectId and random uuid.
	Name string `gorm:"primary_key:true;" json:"name"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId uint64    `gorm:"primary_key:true;" json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateEventName(eventName *EventName) (*EventName, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"eventName": &eventName}).Info("Creating event name")

	if err := db.Create(eventName).Error; err != nil {
		log.WithFields(log.Fields{"eventName": &eventName, "error": err}).Error("CreateEventName Failed")
		return nil, http.StatusInternalServerError
	} else {
		return eventName, DB_SUCCESS
	}
}

func GetEventName(name string, projectId uint64) (*EventName, int) {
	// Input Validation. (ID is to be auto generated)
	if name == "" || projectId == 0 {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventName EventName
	if err := db.Where(&EventName{Name: name, ProjectId: projectId}).First(&eventName).Error; err != nil {
		return nil, 404
	} else {
		return &eventName, DB_SUCCESS
	}
}
