package model

/*
import (
	C "config"
	"net/http"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type Agent struct {
	// Composite primary key with project_id and random uuid.
	ID uint `gorm:"primary_key:true;" json:"id"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId uint64    `gorm:"primary_key:true;" json:"project_id"`
	Email
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
		return user, DB_SUCCESS
	}
}
*/
