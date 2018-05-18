package model

import (
	C "config"
	"time"

	_ "github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	ID        uint   `gorm:"primary_key" json:"-"`
	AccountId string `json:"account_id"`
	UserId    string `json:"user_id"`
	EventId   string `json:"event_id"`
	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Attributes postgres.Jsonb `json:"attributes"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func CreateEvent(event *Event) (*Event, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"event": &event}).Info("Creating event")
	// Input Validation.
	if event.ID != 0 || event.AccountId == "" || event.EventId == "" {
		return nil, 422
	}

	if err := db.Create(event).Error; err != nil {
		log.WithFields(log.Fields{"event": &event, "error": err}).Error("CreateEvent Failed")
		return nil, 500
	} else {
		return event, -1
	}
}

func GetEvent(id string) (*Event, int) {
	db := C.GetServices().Db

	var event Event
	if err := db.Where("id = ?", id).First(&event).Error; err != nil {
		return nil, 404
	} else {
		return &event, -1
	}
}
