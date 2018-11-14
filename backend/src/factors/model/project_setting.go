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
	ProjectId uint64 `gorm:"primary_key:true" json:"project_id"`
	// Defaults to AUTO_TRACK_DISABLED.
	AutoTrack uint8 `gorm:"default:0" json:"auto_track"`
}

// Enum AutoTrack.
const (
	AUTO_TRACK_DISABLED = 0
	AUTO_TRACK_ENABLED  = 1
)

// Validate creates and updates.
func (projectSetting ProjectSetting) Validate(db *gorm.DB) {
	// Project scope validation.
	if projectSetting.ProjectId == 0 {
		db.AddError(ErrInvalidProjectScope)
	}
}

func GetProjectSetting(projectId uint64) (*ProjectSetting, int) {
	db := C.GetServices().Db

	if valid := isValidProjectScope(projectId); !valid {
		return nil, http.StatusBadRequest
	}

	var projectSetting ProjectSetting
	if err := db.Where("project_id = ?", projectId).First(&projectSetting).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// http.BadRequest because of non-existing project id given.
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &projectSetting, DB_SUCCESS
}

func CreateProjectSetting(projectSetting *ProjectSetting) (*ProjectSetting, int) {
	db := C.GetServices().Db

	if err := db.Create(projectSetting).Error; err != nil {
		if isInvalidProjectScopeError(err) {
			return nil, http.StatusBadRequest
		}

		log.WithFields(log.Fields{"ProjectSetting": projectSetting, "error": err}).Error("Failed creating ProjectSetting.")
		return nil, http.StatusInternalServerError
	}

	return projectSetting, DB_SUCCESS
}
