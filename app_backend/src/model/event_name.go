package model

import (
	C "config"
	"net/http"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type EventName struct {
	// Composite primary key with account_id and random uuid.
	Name string `gorm:"primary_key:true;" json:"name"`
	// Below are the foreign key constraints added in creation script.
	// account_id -> accounts(id)
	AccountId uint64    `gorm:"primary_key:true;" json:"account_id"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateEventName(event_name *EventName) (*EventName, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"event_name": &event_name}).Info("Creating event name")

	if err := db.Create(event_name).Error; err != nil {
		log.WithFields(log.Fields{"event_name": &event_name, "error": err}).Error("CreateEventName Failed")
		return nil, http.StatusInternalServerError
	} else {
		return event_name, -1
	}
}
