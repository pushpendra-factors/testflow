package model

import (
	C "factors/config"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type ProjectSetting struct {
	// Foreign key constraing project_id -> projects(id)
	// Used project_id as primary key also, becase of 1-1 relationship.
	ProjectId uint64 `gorm:"primary_key:true" json:"-"` // exclude on JSON response.
	// Using pointers to avoid update by default value.
	// omit empty to avoid nil(filelds not updated) on resp json.
	AutoTrack  *bool     `gorm:"not null;default:false" json:"auto_track,omitempty"`
	IntSegment *bool     `gorm:"not null;default:false" json:"int_segment,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

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

	return &projectSetting, http.StatusFound
}

func createProjectSetting(ps *ProjectSetting) (*ProjectSetting, int) {
	db := C.GetServices().Db

	if valid := isValidProjectScope(ps.ProjectId); !valid {
		return nil, http.StatusBadRequest
	}

	if err := db.Create(ps).Error; err != nil {
		log.WithFields(log.Fields{"ProjectSetting": ps,
			"error": err}).Error("Failed creating ProjectSetting.")
		return nil, http.StatusInternalServerError
	}

	return ps, http.StatusCreated
}

func UpdateProjectSettings(projectId uint64, settings *ProjectSetting) (*ProjectSetting, int) {
	db := C.GetServices().Db

	if projectId == 0 {
		return nil, http.StatusBadRequest
	}

	var updatedProjectSetting ProjectSetting
	if err := db.Model(&updatedProjectSetting).Where("project_id = ?",
		projectId).Updates(settings).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		log.WithFields(log.Fields{"ProjectSetting": settings,
			"error": err}).Error("Failed updating ProjectSettings.")
		return nil, http.StatusInternalServerError
	}

	return &updatedProjectSetting, http.StatusAccepted
}

func IsPSettingsIntSegmentEnabled(projectId uint64) bool {
	settings, errCode := GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		log.WithFields(log.Fields{"project_id": projectId, "err_code": errCode}).Error("Failed fetching project settings.")
		return false
	}

	return *settings.IntSegment
}
