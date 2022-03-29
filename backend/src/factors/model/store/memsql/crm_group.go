package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// isExistCRMGroupByID check for existing user external id, object type and source
func isExistCRMGroupByID(projectID uint64, source model.CRMSource, groupType int, id string) (int, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"group_type": groupType,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || groupType <= 0 || id == "" || source <= 0 {
		logCtx.Error("Missing required parameters.")
		return http.StatusBadRequest, errors.New("missing required fields project_id, group_type, id, source")
	}

	var crmGroup model.CRMGroup
	db := C.GetServices().Db
	err := db.Model(&model.CRMGroup{}).Where("project_id = ? AND source = ? "+
		"AND type = ? AND id = ? AND action = ? ",
		projectID, source, groupType, id, model.CRMActionCreated).Select("id").Limit(1).Find(&crmGroup).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, nil
		}

		logCtx.WithError(err).Error("Failed to get user from crm_group table.")
		return http.StatusInternalServerError, err
	}

	if crmGroup.ID == "" {
		return http.StatusNotFound, nil
	}

	return http.StatusFound, nil
}

func (store *MemSQL) CreateCRMGroup(crmGroup *model.CRMGroup) (int, error) {
	logFields := log.Fields{
		"project_id": crmGroup.ProjectID,
		"id":         crmGroup.ID,
		"source":     crmGroup.Source,
		"group_type": crmGroup.Type,
		"timestamp":  crmGroup.Timestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if !model.AllowedCRMBySource(crmGroup.Source) {
		logCtx.Error("Invalid source.")
		return http.StatusBadRequest, errors.New("invalid source")
	}

	/*
		Id - External group object id, unique for a group object. Used to track changes on the group object
		Type - Group object name
		Properties - Group object properties as key and value only
	*/
	if crmGroup.ProjectID == 0 || crmGroup.ID == "" || crmGroup.Type <= 0 || crmGroup.Properties == nil {
		logCtx.Error("Missing required parameters.")
		return http.StatusBadRequest, errors.New("missing required fields project_id, id, group_type, properties")
	}

	// timestamp is when the group objects state occured
	if crmGroup.Timestamp <= 0 {
		logCtx.Error("Missing crm group timestamp.")
		return http.StatusBadRequest, errors.New("missing timestamp")
	}

	if U.IsEmptyPostgresJsonb(crmGroup.Properties) {
		logCtx.Error("Empty group properties.")
		return http.StatusBadRequest, errors.New("empty properties")
	}

	// metadata for any information other than group properties, which required for special processing
	if crmGroup.Metadata != nil && U.IsEmptyPostgresJsonb(crmGroup.Properties) {
		logCtx.Error("Empty group metadata.")
		return http.StatusBadRequest, errors.New("empty metadata")
	}

	status, err := isExistCRMGroupByID(crmGroup.ProjectID, crmGroup.Source, crmGroup.Type, crmGroup.ID)
	if status == http.StatusInternalServerError {
		logCtx.WithError(err).Error("Failed to check existing crm group.")
		return http.StatusInternalServerError, err
	}

	// first state of a group object should be stored as action created representing new group object, followed by only updated
	// To track the action used for group refer the same object
	isNew := status == http.StatusNotFound
	if isNew {
		crmGroup.Action = model.CRMActionCreated
	} else {
		crmGroup.Action = model.CRMActionUpdated
	}

	db := C.GetServices().Db
	err = db.Create(&crmGroup).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			return http.StatusConflict, nil
		}

		logCtx.WithError(err).Error("Failed to insert crm group.")
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}
