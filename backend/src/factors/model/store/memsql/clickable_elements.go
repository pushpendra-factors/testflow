package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"

	"errors"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) UpdateButtonClickEventById(projectId int64, reqPayload *model.SDKButtonElementAttributesPayload) (int, error) {
	logCtx := log.WithField("project_id", projectId).WithField("request_payload", reqPayload)

	logFields := log.Fields{
		"project_id":      projectId,
		"request_payload": reqPayload,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if projectId == 0 || reqPayload.DisplayName == "" || reqPayload.ElementType == "" {
		logCtx.Error("Invalid parameters.")
		return http.StatusBadRequest, errors.New("Invalid parameters.")
	}

	if reqPayload.ElementAttributes.Timestamp == 0 {
		reqPayload.ElementAttributes.Timestamp = time.Now().Unix()
	}

	event, err := store.GetButtonClickEventById(projectId, reqPayload.DisplayName, reqPayload.ElementType)
	if err == http.StatusNotFound {
		return GetStore().CreateButtonClickEventById(projectId, reqPayload)
	} else if err == http.StatusBadRequest {
		logCtx.Error("Invalid parameters.")
		return http.StatusBadRequest, errors.New("Update button click failed. Invalid parameters.")
	} else if err == http.StatusInternalServerError {
		logCtx.Error("Getting button click failed.")
		return http.StatusInternalServerError, errors.New("Update button click failed. Getting button click failed.")
	}

	db := C.GetServices().Db

	event.ClickCount += 1
	event.UpdatedAt = time.Unix(reqPayload.ElementAttributes.Timestamp, 0)
	if err := db.Save(&event).Error; err != nil {
		logCtx.WithField("err", err).Error("Failed in updating button click.")
		return http.StatusInternalServerError, errors.New("Update button click failed. Failed to update button click.")
	}

	return http.StatusAccepted, nil
}

func (store *MemSQL) CreateButtonClickEventById(projectId int64, buttonClick *model.SDKButtonElementAttributesPayload) (int, error) {
	logFields := log.Fields{
		"project_id":   projectId,
		"button_click": buttonClick,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectId == 0 || buttonClick.DisplayName == "" || buttonClick.ElementType == "" {
		logCtx.Error("Invalid parameters.")
		return http.StatusBadRequest, errors.New("Failed to create a button click event. Invalid parameters.")
	}

	if buttonClick.ElementAttributes.Timestamp == 0 {
		buttonClick.ElementAttributes.Timestamp = time.Now().Unix()
	}

	_, error := store.GetButtonClickEventById(projectId, buttonClick.DisplayName, buttonClick.ElementType)
	if error == http.StatusFound {
		logCtx.Error("Duplicate.")
		return http.StatusConflict, errors.New("Failed to create a button click event. Duplicate.")
	} else if error == http.StatusBadRequest {
		logCtx.Error("Invalid parameters.")
		return http.StatusBadRequest, errors.New("Failed to create a button click event. Invalid parameters.")
	} else if error == http.StatusInternalServerError {
		logCtx.Error("Getting button click failed.")
		return http.StatusInternalServerError, errors.New("Failed to create a button click event. Getting button click failed.")
	}

	elementAttributes, err := U.EncodeStructTypeToPostgresJsonb(buttonClick.ElementAttributes)
	if err != nil {
		logCtx.Error("Cannot convert struct to json.")
		return http.StatusInternalServerError, errors.New("Failed to create a button click event. Cannot convert struct to json.")
	}

	event := model.ClickableElements{
		ProjectID:         projectId,
		Id:                U.GetUUID(),
		DisplayName:       buttonClick.DisplayName,
		ElementType:       buttonClick.ElementType,
		ElementAttributes: elementAttributes,
		ClickCount:        1,
		Enabled:           false,
		CreatedAt:         time.Unix(buttonClick.ElementAttributes.Timestamp, 0),
		UpdatedAt:         time.Unix(buttonClick.ElementAttributes.Timestamp, 0),
	}

	db := C.GetServices().Db
	dbx := db.Create(&event)
	if dbx.Error != nil {
		if IsDuplicateRecordError(dbx.Error) {
			logCtx.WithError(dbx.Error).Error("Duplicate.")
			return http.StatusConflict, errors.New("Failed to create a button click event. Duplicate.")
		}
		logCtx.WithError(dbx.Error).Error("Failed to create a button click event.")
		return http.StatusInternalServerError, errors.New("Failed to create a button click event")
	}

	return http.StatusCreated, nil
}

func (store *MemSQL) GetButtonClickEventById(projectId int64, displayName string, elementType string) (*model.ClickableElements, int) {
	logFields := log.Fields{
		"project_id":   projectId,
		"display_name": displayName,
		"element_type": elementType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectId == 0 || displayName == "" || elementType == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	var event model.ClickableElements

	db := C.GetServices().Db
	dbx := db.Limit(1).Where("project_id = ? AND display_name = ? AND element_type = ?", projectId, displayName, elementType)

	if err := dbx.Find(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// Do not log error. Log on caller, if needed.
			return nil, http.StatusNotFound
		}

		log.WithFields(log.Fields{"project_id": projectId, "display_name": displayName, "element_type": elementType}).WithError(err).Error(
			"Getting button click failed on GetButtonClickById.")
		return nil, http.StatusInternalServerError
	}

	return &event, http.StatusFound
}
