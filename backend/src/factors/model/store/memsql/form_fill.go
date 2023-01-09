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
		"form_id":    formFillPayload.FormId,
		"field_id":   formFillPayload.FieldId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectId == 0 || formFillPayload.UserId == "" || formFillPayload.FormId == "" || formFillPayload.FieldId == "" {
		logCtx.Error("Failed to create form fill event by ID. Invalid parameters")
		return http.StatusBadRequest, errors.New("Failed to create a form fill event. Invalid parameters.")
	}

	formFill := model.FormFill{
		ProjectID:       projectId,
		UserId:          formFillPayload.UserId,
		FormId:          formFillPayload.FormId,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Id:              U.GetUUID(),
		Value:           formFillPayload.Value,
		FieldId:         formFillPayload.FieldId,
		EventProperties: formFillPayload.EventProperties,
	}

	if formFillPayload.UpdatedAt != nil {
		formFill.UpdatedAt = *formFillPayload.UpdatedAt
	}

	db := C.GetServices().Db
	dbx := db.Create(&formFill)
	if dbx.Error != nil {
		if IsDuplicateRecordError(dbx.Error) {
			return http.StatusConflict, errors.New("Failed to create a form fill event. Duplicate.")
		}
		logCtx.WithError(dbx.Error).Error("Failed to create a form fill event.")
		return http.StatusInternalServerError, errors.New("Failed to create a form fill event")
	}

	return http.StatusCreated, nil
}

func (store *MemSQL) GetFormFillEventById(projectId int64, userId string,
	formId string, fieldId string) (*model.FormFill, int) {
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
	dbx := db.Limit(1).Where("project_id = ? AND user_id = ? AND form_id = ? AND field_id = ?",
		projectId, userId, formId, fieldId)

	var event model.FormFill
	if err := dbx.Find(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error(
			"Getting form fill failed on GetFormFillEventById.")
		return nil, http.StatusInternalServerError
	}

	return &event, http.StatusFound
}

func (store *MemSQL) GetFormFillEventsUpdatedBeforeTenMinutes(projectIds []int64) ([]model.FormFill, error) {
	logFields := log.Fields{
		"project_ids": projectIds,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var updatedTime = U.TimeNowZ().Add(-time.Minute * 10)
	var formFills []model.FormFill

	db := C.GetServices().Db
	err := db.Table("form_fills").Where("project_id IN (?) AND updated_at < ?", projectIds, updatedTime).Find(&formFills).Error
	if err != nil {
		logCtx.WithFields(log.Fields{"project_ids": projectIds}).WithError(err).Error("fetching enabled project_id failed")
		return formFills, err
	}
	return formFills, nil
}

func (store *MemSQL) DeleteFormFillProcessedRecords(projectId int64, userId string, formId string, fieldId string) (int, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"user_id":    userId,
		"form_id":    formId,
		"field_id":   fieldId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db
	err := db.Table("form_fills").
		Where("project_id = ? AND user_id = ? AND form_id = ? AND field_id = ?",
			projectId, userId, formId, fieldId).
		Delete(&model.FormFill{}).Error
	if err != nil {
		logCtx.WithError(err).Error("Delete form fill records failed.")
		return http.StatusBadRequest, err
	}
	return http.StatusAccepted, nil
}
