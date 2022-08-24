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

var allowedElementTypes = map[string]bool{
	"BUTTON": true,
	"ANCHOR": true,
}

func isAllowedElementType(elementType string) bool {
	_, isAllowed := allowedElementTypes[elementType]
	return isAllowed
}

func (store *MemSQL) UpsertCountAndCheckEnabledClickableElement(projectId int64,
	reqPayload *model.CaptureClickPayload) (isEnabled bool, status int, err error) {
	logCtx := log.WithField("project_id", projectId).WithField("request_payload", reqPayload)

	logFields := log.Fields{
		"project_id":      projectId,
		"request_payload": reqPayload,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if projectId == 0 || reqPayload.DisplayName == "" {
		logCtx.Error("Invalid parameters.")
		return false, http.StatusBadRequest, errors.New("Invalid parameters.")
	}

	if !isAllowedElementType(reqPayload.ElementType) {
		logCtx.Warn("Captured element type which is not part of allowed list.")
	}

	allowedAttributes := U.PropertiesMap{}
	model.AddAllowedElementAttributes(projectId, reqPayload.ElementAttributes, &allowedAttributes)
	reqPayload.ElementAttributes = allowedAttributes

	element, getErr := store.GetClickableElement(projectId, reqPayload.DisplayName, reqPayload.ElementType)
	if getErr == http.StatusNotFound {
		status, err := GetStore().CreateClickableElement(projectId, reqPayload)
		return false, status, err
	} else if getErr == http.StatusBadRequest {
		logCtx.Error("Invalid parameters.")
		return false, http.StatusBadRequest, errors.New("Update click failed. Invalid parameters.")
	} else if getErr == http.StatusInternalServerError {
		logCtx.Error("Gettingclick failed.")
		return false, http.StatusInternalServerError,
			errors.New("Updateclick failed. Getting click failed.")
	}

	db := C.GetServices().Db
	if err := db.Model(&model.ClickableElements{}).
		Where("project_id = ? AND display_name = ? AND element_type = ?", projectId, reqPayload.DisplayName, reqPayload.ElementType).
		Update(map[string]interface{}{"click_count": element.ClickCount + 1}).
		Error; err != nil {

		logCtx.WithField("err", err).Error("Failed to increment click.")

		// If enabled log and return positive, to avoid confusion.
		//click increment is secondary for enabled elements.
		if element.Enabled {
			return element.Enabled, http.StatusAccepted, nil
		}

		return element.Enabled, http.StatusInternalServerError,
			errors.New("Updateclick failed. Failed to update click.")
	}

	return element.Enabled, http.StatusAccepted, nil
}

func (store *MemSQL) CreateClickableElement(projectId int64, click *model.CaptureClickPayload) (int, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"click":      click,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectId == 0 || click.DisplayName == "" || click.ElementType == "" {
		logCtx.Error("Invalid parameters.")
		return http.StatusBadRequest, errors.New("Failed to create aclick event. Invalid parameters.")
	}

	elementAttributes, err := U.EncodeStructTypeToPostgresJsonb(click.ElementAttributes)
	if err != nil {
		logCtx.Error("Cannot convert struct to json.")
		return http.StatusInternalServerError,
			errors.New("Failed to create aclick event. Cannot convert struct to json.")
	}

	event := model.ClickableElements{
		ProjectID:         projectId,
		Id:                U.GetUUID(),
		DisplayName:       click.DisplayName,
		ElementType:       click.ElementType,
		ElementAttributes: elementAttributes,
		ClickCount:        1,
		Enabled:           false,
	}

	db := C.GetServices().Db
	dbx := db.Create(&event)
	if dbx.Error != nil {
		if IsDuplicateRecordError(dbx.Error) {
			logCtx.WithError(dbx.Error).Error("Duplicate.")
			return http.StatusConflict, errors.New("Failed to create a click event. Duplicate.")
		}
		logCtx.WithError(dbx.Error).Error("Failed to create a click event.")
		return http.StatusInternalServerError, errors.New("Failed to create aclick event")
	}

	return http.StatusCreated, nil
}

func (store *MemSQL) GetClickableElement(projectId int64, displayName string,
	elementType string) (*model.ClickableElements, int) {

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
	dbx := db.Limit(1).
		Where("project_id = ? AND display_name = ? AND element_type = ?",
			projectId, displayName, elementType)

	if err := dbx.Find(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		log.WithFields(log.Fields{"project_id": projectId, "display_name": displayName, "element_type": elementType}).
			WithError(err).
			Error("Getting click failed on get clickable element.")
		return nil, http.StatusInternalServerError
	}

	return &event, http.StatusFound
}

func (store *MemSQL) ToggleEnabledClickableElement(projectId int64, id string) int {
	logCtx := log.WithField("project_id", projectId).WithField("id", id)

	if projectId == 0 || id == "" {
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Exec("UPDATE clickable_elements SET enabled = NOT enabled WHERE project_id = ? AND id = ?", projectId, id).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to toggle enabled clickable elements")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (store *MemSQL) GetAllClickableElements(projectId int64) ([]model.ClickableElements, int) {
	logCtx := log.WithField("project_id", projectId)

	var clickableElements []model.ClickableElements

	db := C.GetServices().Db
	err := db.Model(&model.ClickableElements{}).Order("click_count DESC").
		Where("project_id = ?", projectId).Find(&clickableElements).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return clickableElements, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get clickable elements")
		return clickableElements, http.StatusInternalServerError
	}

	if len(clickableElements) == 0 {
		return clickableElements, http.StatusNotFound
	}

	return clickableElements, http.StatusFound
}
