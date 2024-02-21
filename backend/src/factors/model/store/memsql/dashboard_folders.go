package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateDashboardFolder(projectId int64, folder *model.DashboardFolders) (*model.DashboardFolders, int) {

	logCtx := log.WithFields(log.Fields{"dashboard_folder": folder, "project_id": projectId})

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)
	db := C.GetServices().Db

	if projectId == 0 {
		return nil, http.StatusBadRequest
	}

	folder.Id = U.GetUUID()
	folder.ProjectId = projectId

	if err := db.Create(folder).Error; err != nil {
		logCtx.WithError(err).Error("Failed to create dashboard.")
		return nil, http.StatusInternalServerError
	}

	return folder, http.StatusCreated
}

// UpdateDashboardFolder updates the name of the folder, which also impacts the updated_at column.
func (store *MemSQL) UpdateDashboardFolder(projectId int64, folderId string, folder *model.UpdatableDashboardFolder) int {

	logCtx := log.WithFields(log.Fields{"dashboard_folder": folder, "folder_id": folderId, "project_id": projectId})

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)
	db := C.GetServices().Db

	if projectId == 0 {
		return http.StatusBadRequest
	}

	// update allowed fields.
	updateFields := make(map[string]interface{}, 0)

	if folder.Name != "" {
		updateFields["name"] = folder.Name
	}

	// nothing to update.
	if len(updateFields) == 0 {
		return http.StatusBadRequest
	}

	err := db.Model(&model.DashboardFolders{}).Where("project_id = ? AND id = ? AND is_deleted = ? AND is_default_folder=?", projectId, folderId, false, false).Update(updateFields).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update dashboard folder.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// GetDashboardFolders fetches the folders by projectId
func (store *MemSQL) GetDashboardFolders(projectId int64) ([]model.DashboardFolders, int) {

	logCtx := log.WithField("project_id", projectId)

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)
	db := C.GetServices().Db

	var folders []model.DashboardFolders
	if projectId == 0 {
		logCtx.Error("Failed to get dashboard folders. Invalid projectId.")
		return nil, http.StatusBadRequest
	}

	err := db.Order("created_at DESC").Where("project_id = ? AND is_deleted = ?", projectId, false).Find(&folders).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return folders, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get dashboard folders.")
		return folders, http.StatusInternalServerError
	}

	return folders, http.StatusFound
}

// DeleteDashboardFolder takes folderId and projectId as params and then soft delete (set is_deleted as true) the folder
func (store *MemSQL) DeleteDashboardFolder(projectId int64, folderId string) int {

	logCtx := log.WithField("project_id", projectId)

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)
	db := C.GetServices().Db

	err := db.Model(&model.DashboardFolders{}).Where("id= ? AND project_id=? AND is_default_folder=?", folderId, projectId, false).
		Update(map[string]interface{}{"is_deleted": true}).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to delete dashboard folder.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (store *MemSQL) GetAllBoardsDashboardFolder(projectId int64) (model.DashboardFolders, int) {

	logCtx := log.WithField("project_id", projectId)

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)
	db := C.GetServices().Db

	var folder model.DashboardFolders
	err := db.Where("project_id = ? AND is_default_folder=? AND is_deleted = ?", projectId, true, false).Find(&folder).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return folder, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get ALL BOARDS dashboard folders.")
		return folder, http.StatusInternalServerError
	}

	return folder, http.StatusFound
}

func (store *MemSQL) IsFolderNameAlreadyExists(projectId int64, name string) int {
	logCtx := log.WithField("project_id", projectId)

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)
	db := C.GetServices().Db

	var folder model.DashboardFolders
	err := db.Where("project_id = ? AND name=? AND is_deleted = ?", projectId, name, false).Find(&folder).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}
		logCtx.WithError(err).Error("Failed to check if folder name already exists.")
		return http.StatusInternalServerError
	}
	return http.StatusFound
}
