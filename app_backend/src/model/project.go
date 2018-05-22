package model

import (
	C "config"
	"net/http"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type Project struct {
	ID        uint64    `gorm:"primary_key:true;" json:"id"`
	Name      string    `gorm:"not null;unique"`
	CreatedAt time.Time `json:"created_at"`
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
		return project, -1
	}
}
