package model

import (
	C "config"
	"net/http"
	"time"

	_ "github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type User struct {
	// Composite primary key with project_id and random uuid.
	ID string `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId uint64 `gorm:"primary_key:true;" json:"project_id"`
	// UserId provided by the customer.
	// An unique index is creatd on ProjectId+UserId.
	CustomerUserId string `json:"c_uid"`

	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Properties postgres.Jsonb `json:"properties,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func CreateUser(user *User) (*User, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"user": &user}).Info("Creating user")

	// Input Validation. (ID is to be auto generated).
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

func GetUser(projectId uint64, id string) (*User, int) {
	db := C.GetServices().Db

	var user User
	if err := db.Where("project_id = ?", projectId).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, 404
	} else {
		return &user, DB_SUCCESS
	}
}
