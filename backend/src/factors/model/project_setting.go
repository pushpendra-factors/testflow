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
	AutoTrack bool `gorm:"not null;default:false" json:"auto_track"`
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

	return &projectSetting, DB_SUCCESS
}

func createProjectSetting(ps *ProjectSetting) (*ProjectSetting, int) {
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

func UpdateProjectSettings(projectId uint64, ps *ProjectSetting) (*ProjectSetting, int) {
	db := C.GetServices().Db

	if projectId == 0 {
		return nil, http.StatusBadRequest
	}

	var updatedProjectSetting ProjectSetting

	// Todo(Dinesh): Create a method to convert Model object to map[string]interface{}
	// and use reuse it for all updates.
	updateFields := map[string]interface{}{"auto_track": ps.AutoTrack}

	// Note: '.Updates(Project{AutoTrack: false})' won't trigger an update query to the backend.
	// Issue with updating with default value. Ref: https://github.com/jinzhu/gorm/issues/314
	if err := db.Model(&updatedProjectSetting).Where("project_id = ?", projectId).Updates(updateFields).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		log.WithFields(log.Fields{"ProjectSetting": ps, "error": err, "update_fields": updateFields}).Error("Failed updating ProjectSettings.")
		return nil, http.StatusInternalServerError
	}

	return &updatedProjectSetting, DB_SUCCESS
}
