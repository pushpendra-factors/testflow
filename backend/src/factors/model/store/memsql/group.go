package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateGroup(projectID uint64, groupName string, allowedGroupNames map[string]bool) (*model.Group, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "group_name": groupName})
	if projectID < 1 || groupName == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	if _, allowed := allowedGroupNames[groupName]; !allowed {
		logCtx.Error("group name not allowed.")
		return nil, http.StatusBadRequest
	}

	_, status := store.GetGroup(projectID, groupName)
	if status != http.StatusNotFound {
		if status == http.StatusFound {
			return nil, http.StatusConflict
		}

		logCtx.Error("Failed to get existing groups.")
		return nil, http.StatusInternalServerError
	}

	id := struct {
		MaxID int `json:"max_id"`
	}{}

	if err := db.Table("groups").Select("max(id) as max_id").Where("project_id = ?", projectID).Find(&id).Error; err != nil {
		logCtx.WithError(err).Error("Failed to get maximum id from groups.")
		return nil, http.StatusInternalServerError
	}

	if id.MaxID >= model.AllowedGroups {
		logCtx.Error("Maximum allowed groups reached.")
		return nil, http.StatusBadRequest
	}

	group := model.Group{
		ProjectID: projectID,
		Name:      groupName,
		ID:        id.MaxID + 1,
	}

	err := db.Create(&group).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			if strings.Contains(err.Error(), "PRIMARY") {
				return nil, http.StatusConflict
			}
		}

		logCtx.WithError(err).Error("Failed to insert group.")
		return nil, http.StatusInternalServerError

	}

	return &group, http.StatusCreated
}

func (store *MemSQL) GetGroup(projectID uint64, groupName string) (*model.Group, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "group_name": groupName})
	if projectID < 1 || groupName == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	group := model.Group{}

	if err := db.Model(&model.Group{}).Where("project_id = ? AND name = ? ", projectID, groupName).Find(&group).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get group.")
		return nil, http.StatusInternalServerError
	}

	return &group, http.StatusFound
}
