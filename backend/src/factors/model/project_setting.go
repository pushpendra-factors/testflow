package model

import (
	C "factors/config"
	"net/http"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type ProjectSetting struct {
	// Foreign key constraing project_id -> projects(id)
	// Used project_id as primary key also, becase of 1-1 relationship.
	ProjectId uint64 `gorm:"primary_key:true" json:"-"` // exclude on JSON response.
	// Defaults to AUTO_TRACK_DISABLED.
	AutoTrack uint8 `gorm:"not null;default:0" json:"auto_track"`
}

// Enum AutoTrack.
const (
	AUTO_TRACK_DISABLED = 0
	AUTO_TRACK_ENABLED  = 1
)

func GetProjectSetting(projectId uint64) (*ProjectSetting, int) {
	db := C.GetServices().Db

	if valid := isValidProjectScope(projectId); !valid {
		return nil, http.StatusBadRequest
	}

	var projectSetting ProjectSetting
	if err := db.Where("project_id = ?", projectId).First(&projectSetting).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &projectSetting, DB_SUCCESS
}

func CreateProjectSetting(ps *ProjectSetting) (*ProjectSetting, int) {
	db := C.GetServices().Db

	if valid := isValidProjectScope(ps.ProjectId); !valid {
		return nil, http.StatusBadRequest
	}

	if err := db.Create(ps).Error; err != nil {
		log.WithFields(log.Fields{"ProjectSetting": ps, "error": err}).Error("Failed creating ProjectSetting.")
		return nil, http.StatusInternalServerError
	}

	return ps, DB_SUCCESS
}
