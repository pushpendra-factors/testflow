package model

import (
	C "config"
	"net/http"
	"time"
	U "util"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type Project struct {
	ID        uint64    `gorm:"primary_key:true;" json:"id"`
	Name      string    `gorm:"not null;" json:"name"`
	APIKey    string    `gorm:"unique" json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (project *Project) BeforeCreate() (err error) {
	// Create a new API Key.
	project.APIKey = U.RandomLowerAphaNumString(32)
	return nil
}

func CreateProject(project *Project) (*Project, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"project": &project}).Info("Creating project")

	// Input Validation. (ID is to be auto generated)
	if project.ID > 0 {
		return nil, http.StatusBadRequest
	}

	if err := db.Create(project).Error; err != nil {
		log.WithFields(log.Fields{"project": &project, "error": err}).Error("CreateProject Failed")
		return nil, http.StatusInternalServerError
	} else {
		return project, DB_SUCCESS
	}
}

func GetProject(id uint64) (*Project, int) {
	db := C.GetServices().Db

	var project Project
	if err := db.Where("id = ?", id).First(&project).Error; err != nil {
		return nil, 404
	} else {
		return &project, DB_SUCCESS
	}
}
