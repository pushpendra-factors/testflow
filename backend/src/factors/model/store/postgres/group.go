package postgres

import (
	C "factors/config"
	"factors/model/model"
	"factors/util"
	"net/http"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) CreateGroup(projectID uint64, groupName string, allowedGroupNames map[string]bool) (*model.Group, int) {
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

	_, status := pg.GetGroup(projectID, groupName)
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

		if util.IsPostgresUniqueIndexViolationError("groups_pkey", err) {
			return nil, http.StatusConflict
		}

		logCtx.WithError(err).Error("Failed to create group.")
		return nil, http.StatusInternalServerError
	}

	return &group, http.StatusCreated
}
func (pg *Postgres) GetGroups(projectId uint64) ([]model.Group, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId})

	if projectId < 1 {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	var groups []model.Group
	db := C.GetServices().Db
	err := db.Where("project_id = ?", projectId).Find(&groups).Error
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error("Failed to get groups.")
		return groups, http.StatusInternalServerError
	}

	return groups, http.StatusFound

}
func (pg *Postgres) GetGroup(projectID uint64, groupName string) (*model.Group, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "group_name": groupName})
	if projectID < 1 || groupName == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	group := model.Group{}

	if err := db.Model(&model.Group{}).Where("project_id = ? AND name = ? ", projectID, groupName).
		Find(&group).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get group.")
		return nil, http.StatusInternalServerError
	}

	return &group, http.StatusFound
}
