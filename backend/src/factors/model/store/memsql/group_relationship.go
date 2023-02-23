package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateGroupRelationship(projectID int64, leftGroupName, leftGroupUserID,
	rightGroupName, rightGroupUserID string) (*model.GroupRelationship, int) {
	logFields := log.Fields{
		"project_id":          projectID,
		"left_group_name":     leftGroupName,
		"left_group_user_id":  leftGroupUserID,
		"right_group_name":    rightGroupName,
		"right_group_user_id": rightGroupUserID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if projectID < 1 || leftGroupName == "" || leftGroupUserID == "" || rightGroupName == "" || rightGroupUserID == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	leftGroup, status := store.GetGroup(projectID, leftGroupName)
	if status != http.StatusFound {
		logCtx.WithField("err_code", status).Error("Failed to get left group name id.")
		if status == http.StatusNotFound {
			return nil, http.StatusBadRequest
		}

		return nil, http.StatusInternalServerError
	}
	rightGroup, status := store.GetGroup(projectID, rightGroupName)
	if status != http.StatusFound {
		logCtx.WithField("err_code", status).Error("Failed to get right group name id.")
		if status == http.StatusNotFound {
			return nil, http.StatusBadRequest
		}

		return nil, http.StatusInternalServerError
	}

	db := C.GetServices().Db

	groupRelationship := model.GroupRelationship{
		ProjectID:        projectID,
		LeftGroupNameID:  leftGroup.ID,
		LeftGroupUserID:  leftGroupUserID,
		RightGroupNameID: rightGroup.ID,
		RightGroupUserID: rightGroupUserID,
	}

	if err := db.Create(&groupRelationship).Error; err != nil {

		if IsDuplicateRecordError(err) {
			return nil, http.StatusConflict
		}

		logCtx.WithError(err).Error("Failed to create group relationship.")
		return nil, http.StatusInternalServerError
	}

	return &groupRelationship, http.StatusCreated
}

func (store *MemSQL) GetGroupRelationshipByUserID(projectID int64, leftGroupUserID string) ([]model.GroupRelationship, int) {
	logFields := log.Fields{
		"project_id":         projectID,
		"left_group_user_id": leftGroupUserID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if projectID < 1 || leftGroupUserID == "" {
		logCtx.Error("Invalid parameters")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	var groupRelationships []model.GroupRelationship

	if err := db.Where("project_id = ? AND left_group_user_id = ?", projectID, leftGroupUserID).
		Find(&groupRelationships).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get group_relationship for left_group_user_id.")
		return nil, http.StatusInternalServerError
	}

	if len(groupRelationships) < 1 {
		return nil, http.StatusNotFound
	}

	return groupRelationships, http.StatusFound
}
