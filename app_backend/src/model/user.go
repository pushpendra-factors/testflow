package model

import (
	C "config"
	"net/http"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type User struct {
	// Composite primary key with account_id and random uuid.
	ID string `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"`
	// Below are the foreign key constraints added in creation script.
	// account_id -> accounts(id)
	AccountId uint64    `gorm:"primary_key:true;" json:"account_id"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateUser(user *User) (*User, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"user": &user}).Info("Creating user")

	// Input Validation. (ID is to be auto generated)
	if user.ID != "" {
		return nil, http.StatusBadRequest
	}

	if err := db.Create(user).Error; err != nil {
		log.WithFields(log.Fields{"user": &user, "error": err}).Error("CreateUser Failed")
		return nil, http.StatusInternalServerError
	} else {
		return user, -1
	}
}
