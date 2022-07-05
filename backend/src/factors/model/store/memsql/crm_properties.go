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

func isCurrentPropertyType(projectID int64, incomingProperty *model.CRMProperty) (int, error) {
	logFields := log.Fields{"project_id": projectID, "crm_property": incomingProperty}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	var currentProperty model.CRMProperty
	err := db.Model(&model.CRMProperty{}).Where("project_id = ? AND source = ? AND type = ? and name = ? ",
		projectID, incomingProperty.Source, incomingProperty.Type, incomingProperty.Name).Order("timestamp desc").Limit(1).
		Find(&currentProperty).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, nil
		}
		log.WithFields(log.Fields{"project_id": projectID, "crm_property": currentProperty}).WithError(err).
			Error("Failed to check for current crm property.")
		return http.StatusInternalServerError, err
	}

	if currentProperty.ID == "" ||
		incomingProperty.ExternalDataType != currentProperty.ExternalDataType {
		return http.StatusNotFound, nil
	}

	return http.StatusFound, nil
}

func (store *MemSQL) CreateCRMProperties(crmProperty *model.CRMProperty) (int, error) {
	logFields := log.Fields{"crm_properties": crmProperty}
	logCtx := log.WithFields(logFields)

	if crmProperty.ProjectID == 0 {
		logCtx.Error("Missing project id.")
		return http.StatusBadRequest, errors.New("missing project id")
	}

	if crmProperty.Source == 0 {
		logCtx.Error("Missing source.")
		return http.StatusBadRequest, errors.New("missing source")
	}

	if crmProperty.Type == 0 || crmProperty.Name == "" || crmProperty.ExternalDataType == "" {
		logCtx.Error("Missing crm property required fields.")
		return http.StatusBadRequest, errors.New("missing properties required fields type,name,mapped_data_type")
	}

	if !model.IsValidCRMMappedDataType(crmProperty.MappedDataType) {
		logCtx.Error("Invalid mapped data type.")
		return http.StatusBadRequest, errors.New("invalid mapped data type")
	}

	status, err := isCurrentPropertyType(crmProperty.ProjectID, crmProperty)
	if err != nil {
		logCtx.WithError(err).Error("Failed to check for current property type.")
		return http.StatusInternalServerError, err
	}

	if status != http.StatusNotFound {
		if status == http.StatusFound {
			return http.StatusConflict, nil
		}

		return http.StatusInternalServerError, errors.New("failed to check for existence of property type")
	}

	crmProperty.Timestamp = U.TimeNowZ().Unix()
	crmProperty.ID = U.GetUUID()

	db := C.GetServices().Db
	if err := db.Create(crmProperty).Error; err != nil {
		if isDuplicateRecord(err) {
			return http.StatusConflict, nil
		}
		logCtx.WithError(err).Error("Failed to insert crm properties.")
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

func (store *MemSQL) GetCRMPropertiesForSync(projectID int64) ([]model.CRMProperty, int) {
	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 {
		logCtx.Error("Invalid project id.")
	}

	db := C.GetServices().Db

	var properties []model.CRMProperty
	err := db.Model(model.CRMProperty{}).Where("project_id = ? AND synced = false", projectID).
		Order("timestamp").Find(&properties).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get crm properties for sync.")
		return nil, http.StatusInternalServerError
	}

	if len(properties) == 0 {
		return nil, http.StatusNotFound
	}

	return properties, http.StatusFound
}

func (store *MemSQL) UpdateCRMProperyAsSynced(projectID int64, source U.CRMSource, crmProperty *model.CRMProperty) (*model.CRMProperty, int) {

	logFields := log.Fields{"crm_properties": crmProperty, "source": source, "project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || source == 0 || crmProperty.ID == "" || crmProperty.Type == 0 ||
		crmProperty.Name == "" || crmProperty.Timestamp == 0 {
		logCtx.Error("Invalid parameters")
		return nil, http.StatusBadRequest
	}

	updates := make(map[string]interface{})
	updates["synced"] = true

	db := C.GetServices().Db
	err := db.Model(&model.CRMProperty{}).Where("project_id = ? AND source = ? AND id = ? AND type= ? AND name = ? AND timestamp= ? ",
		projectID, source, crmProperty.ID, crmProperty.Type, crmProperty.Name, crmProperty.Timestamp).Updates(updates).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update crm user as synced.")
		return nil, http.StatusInternalServerError
	}

	crmProperty.Synced = true

	return crmProperty, http.StatusAccepted
}
