package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type EventName struct {
	// Composite primary key with projectId.
	ID   uint64 `gorm:"primary_key:true;" json:"id"`
	Name string `json:"name"`
	// auto_name Defaults to user_created, if not supplied.
	AutoName string `gorm:"default:'$UCEN'" json:"auto_name";`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId uint64    `gorm:"primary_key:true;" json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Special autoname.
const USER_CREATED_EVENT_NAME = "$UCEN"

var ALLOWED_AUTONAMES = [...]string{USER_CREATED_EVENT_NAME}

func CreateOrGetEventName(eventName *EventName) (*EventName, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"eventName": &eventName}).Info("Create or get event_name")

	// Validation.
	if eventName.ProjectId == 0 || IsValidName(eventName.Name) != nil || isValidAutoName(eventName.AutoName) != nil {
		return nil, http.StatusBadRequest
	}

	if err := db.FirstOrInit(&eventName, &eventName).Error; err != nil {
		log.WithFields(log.Fields{"eventName": &eventName, "error": err}).Error("CreateEventName Failed")
		return nil, http.StatusInternalServerError
	}

	// Checks new record or not.
	if !eventName.CreatedAt.IsZero() {
		log.WithFields(log.Fields{"eventName": &eventName}).Info("Event Name already exists.")
		return eventName, http.StatusConflict
	} else if err := db.Create(eventName).Error; err != nil {
		log.WithFields(log.Fields{"eventName": &eventName, "error": err}).Error("CreateEventName Failed")
		return nil, http.StatusInternalServerError
	}
	return eventName, http.StatusCreated
}

func isValidAutoName(autoName string) error {
	// Allows only allowed autonames.
	if strings.HasPrefix(autoName, U.NAME_PREFIX) {
		for _, allowedAutoName := range ALLOWED_AUTONAMES {
			if autoName != allowedAutoName {
				return errors.New("invalid autoname")
			}
		}
	}
	return nil
}

// Create or Get user created EventName.
func CreateOrGetUserCreatedEventName(eventName *EventName) (*EventName, int) {
	eventName.AutoName = USER_CREATED_EVENT_NAME
	return CreateOrGetEventName(eventName)
}

func GetEventName(name string, projectId uint64) (*EventName, int) {
	// Input Validation. (ID is to be auto generated)
	if name == "" || projectId == 0 {
		log.Error("GetEventName Failed. Missing name or projectId")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventName EventName
	if err := db.Where(&EventName{Name: name, ProjectId: projectId}).First(&eventName).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &eventName, http.StatusFound
}

func GetEventNameByFilter(filter *EventName) (*EventName, int) {
	db := C.GetServices().Db

	var eventName EventName
	if err := db.First(&eventName, &filter).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &eventName, http.StatusFound
}

func GetEventNames(projectId uint64) ([]EventName, int) {
	if projectId == 0 {
		log.Error("GetEventNames Failed. Missing projectId")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventNames []EventName
	if err := db.Where("project_id = ?", projectId).Find(&eventNames).Error; err != nil {
		return nil, http.StatusInternalServerError
	}
	if len(eventNames) == 0 {
		return nil, http.StatusNotFound
	}
	return eventNames, http.StatusFound
}
