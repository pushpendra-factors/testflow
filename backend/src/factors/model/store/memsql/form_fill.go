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

func (store *MemSQL) CreateFormFillEventById(projectId int64, formFillPayload *model.SDKFormFillPayload) (int, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"form_fill":  formFillPayload,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectId == 0 || formFillPayload.FormId == "" || formFillPayload.FieldId == "" {
		logCtx.Error("Failed to create form fill event by ID. Invalid parameters")
		return http.StatusBadRequest, errors.New("Failed to create a form fill event. Invalid parameters.")
	}

	if formFillPayload.TimeSpent == 0 {
		logCtx.WithFields(log.Fields{"project_id": projectId, "form_id": formFillPayload.FormId, "field_id": formFillPayload.FieldId}).Info("Changing value of time_spent_on_field from 0 to 100")
		formFillPayload.TimeSpent = 100
	}

	if formFillPayload.FirstUpdatedTime == 0 {
		var firstUpdatedTimeValue = time.Now().Unix()
		logCtx.WithFields(log.Fields{"project_id": projectId, "form_id": formFillPayload.FormId, "field_id": formFillPayload.FieldId}).Info("Changing value of first_updated_time from 0 to ", firstUpdatedTimeValue)
		formFillPayload.FirstUpdatedTime = firstUpdatedTimeValue
	}

	if formFillPayload.LastUpdatedTime == 0 {
		var lastUpdatedTimeValue = time.Now().Unix()
		logCtx.WithFields(log.Fields{"project_id": projectId, "form_id": formFillPayload.FormId, "field_id": formFillPayload.FieldId}).Info("Changing value of last_updated_time from 0 to ", lastUpdatedTimeValue)
		formFillPayload.LastUpdatedTime = lastUpdatedTimeValue
	}

	_, error := store.GetFormFillEventById(projectId, formFillPayload.FormId, formFillPayload.FieldId)
	if error == http.StatusFound {
		return http.StatusConflict, nil
	} else if error == http.StatusBadRequest {
		return http.StatusBadRequest, errors.New("Invalid parameters on create form fill event.")
	} else if error == http.StatusInternalServerError {
		return http.StatusInternalServerError, errors.New("Failed to create a form fill event. Getting form fill event failed.")
	}

	event := model.FormFill{
		ProjectID:        projectId,
		FormId:           formFillPayload.FormId,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Id:               U.GetUUID(),
		Value:            formFillPayload.Value,
		TimeSpentOnField: formFillPayload.TimeSpent,
		FirstUpdatedTime: formFillPayload.FirstUpdatedTime,
		LastUpdatedTime:  formFillPayload.LastUpdatedTime,
		FieldId:          formFillPayload.FieldId,
	}

	db := C.GetServices().Db
	dbx := db.Create(&event)
	if dbx.Error != nil {
		if IsDuplicateRecordError(dbx.Error) {
			return http.StatusConflict, errors.New("Failed to create a form fill event. Duplicate.")
		}
		logCtx.WithError(dbx.Error).Error("Failed to create a form fill event.")
		return http.StatusInternalServerError, errors.New("Failed to create a form fill event")
	}

	return http.StatusCreated, nil
}

func (store *MemSQL) GetFormFillEventById(projectId int64, formId string, fieldId string) (*model.FormFill, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"form_id":    formId,
		"field_id":   fieldId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectId == 0 || formId == "" || fieldId == "" {
		logCtx.Error("Failed to get form fill event by ID. Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	dbx := db.Limit(1).Where("project_id = ? AND form_id = ? AND field_id = ?", projectId, formId, fieldId)

	var event model.FormFill
	if err := dbx.Find(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithFields(log.Fields{"project_id": projectId, "form_id": formId, "field_id": fieldId}).WithError(err).Error(
			"Getting form fill failed on GetFormFillEventById.")
		return nil, http.StatusInternalServerError
	}

	return &event, http.StatusFound
}
