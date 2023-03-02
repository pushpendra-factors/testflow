package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	"errors"
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateOrUpdateDisplayNameLabel(projectID int64, source, propertyKey, value, label string) int {
	logFields := log.Fields{
		"project_id":   projectID,
		"source":       source,
		"property_key": propertyKey,
		"value":        value,
		"label":        label,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || source == "" || propertyKey == "" || value == "" {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest
	}

	element, getErrCode, getErr := store.GetDisplayNameLabel(projectID, source, propertyKey, value)
	if getErrCode == http.StatusBadRequest || getErrCode == http.StatusInternalServerError {
		logCtx.WithError(getErr).Error("Failed on CreateOrUpdateDisplayNameLabel.")
		return getErrCode
	} else if getErrCode == http.StatusNotFound {
		status, err := store.CreateDisplayNameLabel(projectID, source, propertyKey, value, label)
		if status != http.StatusCreated {
			logCtx.WithError(err).Error("Failed on CreateOrUpdateDisplayNameLabel.")
			return status
		}
		return http.StatusCreated
	}

	if element.Label == label {
		return http.StatusConflict
	}

	db := C.GetServices().Db
	if err := db.Model(&model.DisplayNameLabel{}).
		Where("project_id = ? AND source = ? AND property_key = ? AND value = ?", projectID, source, propertyKey, value).
		Update(map[string]interface{}{"label": label}).Error; err != nil {

		logCtx.WithField("err", err).Error("Failed to update display name label.")

		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (store *MemSQL) CreateDisplayNameLabel(projectID int64, source, propertyKey, value, label string) (int, error) {
	if projectID == 0 || source == "" || propertyKey == "" || value == "" || label == "" {
		return http.StatusBadRequest, errors.New("Invalid parameters.")
	}

	displayNameLabel := model.DisplayNameLabel{
		ID:          U.GetUUID(),
		ProjectID:   projectID,
		Source:      source,
		PropertyKey: propertyKey,
		Value:       value,
		Label:       label,
	}

	db := C.GetServices().Db
	dbx := db.Create(&displayNameLabel)
	// Duplicates gracefully handled and allowed further.
	if dbx.Error != nil && !IsDuplicateRecordError(dbx.Error) {
		return http.StatusInternalServerError, errors.New("Failed to create a display name")
	}

	return http.StatusCreated, nil
}

func (store *MemSQL) GetDisplayNameLabel(projectID int64, source, propertyKey, value string) (*model.DisplayNameLabel, int, error) {
	if projectID == 0 || source == "" || propertyKey == "" || value == "" {
		return nil, http.StatusBadRequest, errors.New("Invalid parameters.")
	}

	var displayNameLabel model.DisplayNameLabel

	db := C.GetServices().Db
	dbx := db.Limit(1).Where("project_id = ? AND source = ? AND property_key = ? AND value = ?", projectID, source, propertyKey, value)

	if err := dbx.Find(&displayNameLabel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, nil
		}
		return nil, http.StatusInternalServerError, errors.New("Failed to get display name label on GetDisplayNameLabel.")
	}

	return &displayNameLabel, http.StatusFound, nil
}

func (store *MemSQL) GetDisplayNameLabelsByProjectIdAndSource(projectID int64, source string) ([]model.DisplayNameLabel, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"source":     source,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || source == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	var displayNameLabel []model.DisplayNameLabel

	db := C.GetServices().Db
	dbx := db.Where("project_id = ? AND source = ?", projectID, source)

	if err := dbx.Find(&displayNameLabel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get display name labels on GetDisplayNameLabelsByProjectIdAndSource.")
		return nil, http.StatusInternalServerError
	}

	return displayNameLabel, http.StatusFound
}
