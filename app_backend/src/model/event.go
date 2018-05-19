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
	// Random string (uuid_v4) as primary key and id.
	ID        string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	AccountId string `json:"account_id"`
	UserId    string `json:"user_id"`
	EventName string `json:"event_name"`
	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Attributes postgres.Jsonb `json:"attributes"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func CreateEvent(event *Event) (*Event, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"event": &event}).Info("Creating event")
	// Input Validation.
	if event.ID != "" || event.AccountId == "" || event.UserId == "" || event.EventName == "" {
		return nil, http.StatusBadRequest
	}

	if err := db.Create(event).Error; err != nil {
		log.WithFields(log.Fields{"event": &event, "error": err}).Error("CreateEvent Failed")
		return nil, http.StatusInternalServerError
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
