package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"

	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) CreateOrUpdateDisplayName(projectID uint64, eventName, propertyName, displayName, tag string) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "property_name": propertyName, "event_name": eventName, "display_name": displayName, "tag": tag})

	if displayName == "" || projectID == 0 {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db

	var entityType int
	if eventName != "" && propertyName != "" {
		entityType = model.DisplayNameEventPropertyEntityType
	} else if eventName != "" {
		entityType = model.DisplayNameEventEntityType
	} else if propertyName != "" {
		entityType = model.DisplayNameUserPropertyEntityType
	} else {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest
	}

	displayNameObj := model.DisplayName{
		ProjectID:    projectID,
		EventName:    eventName,
		PropertyName: propertyName,
		Tag:          tag,
		EntityType:   entityType,
		DisplayName:  displayName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(&displayNameObj).Error; err != nil {
		if U.IsPostgresUniqueIndexViolationError("display_names_project_id_event_name_property_name_tag_unique_id", err) {
			updateFields := map[string]interface{}{
				"display_name": displayName,
			}
			query := db.Model(&model.DisplayName{}).Where("project_id = ? AND event_name = ? AND property_name = ? AND  tag = ? AND entity_type = ?",
				projectID, eventName, propertyName, tag, entityType).Updates(updateFields)
			if err := query.Error; err != nil {
				logCtx.WithError(err).Error("Failed updating property details.")
				return http.StatusInternalServerError
			}

			if query.RowsAffected == 0 {
				return http.StatusInternalServerError
			}
		} else if U.IsPostgresUniqueIndexViolationError("display_names_project_id_object_group_entity_tag_unique_idx", err) {
			return http.StatusConflict
		} else {
			return http.StatusInternalServerError
		}
	}
	return http.StatusCreated
}

func (pg *Postgres) CreateOrUpdateDisplayNameByObjectType(projectID uint64, propertyName, objectType, displayName, group string) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "property_name": propertyName, "object_type": objectType})

	if objectType == "" || propertyName == "" || displayName == "" || group == "" || projectID == 0 {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db

	displayNameObj := model.DisplayName{
		ProjectID:       projectID,
		PropertyName:    propertyName,
		GroupObjectName: objectType,
		Tag:             "Source",
		GroupName:       group,
		EntityType:      model.DisplayNameObjectEntityType,
		DisplayName:     displayName,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := db.Create(&displayNameObj).Error; err != nil {
		if U.IsPostgresUniqueIndexViolationError("display_names_project_id_event_name_property_name_tag_unique_id", err) {
			updateFields := map[string]interface{}{
				"display_name": displayName,
			}
			query := db.Model(&model.DisplayName{}).Where("project_id = ? AND property_name = ? AND group_object_name = ? AND group_name = ? AND tag = ? AND entity_type = ?",
				projectID, propertyName, objectType, group, "Source", model.DisplayNameObjectEntityType).Updates(updateFields)
			if err := query.Error; err != nil {
				logCtx.WithError(err).Error("Failed updating property details.")
				return http.StatusInternalServerError
			}

			if query.RowsAffected == 0 {
				return http.StatusInternalServerError
			}
		} else if U.IsPostgresUniqueIndexViolationError("display_names_project_id_object_group_entity_tag_unique_idx", err) {
			return http.StatusConflict
		} else {
			return http.StatusInternalServerError
		}
	}
	return http.StatusCreated
}

func (pg *Postgres) GetDisplayNamesForAllEvents(projectID uint64) (int, map[string]string) {
	if projectID == 0 {
		return http.StatusBadRequest, nil
	}

	entityType := model.DisplayNameEventEntityType

	displayNameFilter := &model.DisplayName{
		ProjectID:  projectID,
		EntityType: entityType,
	}

	db := C.GetServices().Db

	var displayNames []model.DisplayName
	if err := db.Where(displayNameFilter).Find(&displayNames).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, nil
		}
		log.WithFields(log.Fields{"projectId": projectID}).WithError(err).Error(
			"Failed to GetDisplayName.")
		return http.StatusInternalServerError, nil
	}

	displayNamesMap := make(map[string]string)
	for _, displayName := range displayNames {
		displayNamesMap[displayName.EventName] = displayName.DisplayName
	}

	return http.StatusFound, displayNamesMap
}

func (pg *Postgres) GetDisplayNamesForAllEventProperties(projectID uint64, eventName string) (int, map[string]string) {
	if projectID == 0 {
		return http.StatusBadRequest, nil
	}

	entityType := model.DisplayNameEventPropertyEntityType

	displayNameFilter := &model.DisplayName{
		ProjectID:  projectID,
		EntityType: entityType,
		EventName:  eventName,
	}

	db := C.GetServices().Db

	var displayNames []model.DisplayName
	if err := db.Where(displayNameFilter).Find(&displayNames).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, nil
		}
		log.WithFields(log.Fields{"projectId": projectID}).WithError(err).Error(
			"Failed to GetDisplayName.")
		return http.StatusInternalServerError, nil
	}

	displayNamesMap := make(map[string]string)
	for _, displayName := range displayNames {
		displayNamesMap[displayName.PropertyName] = displayName.DisplayName
	}

	return http.StatusFound, displayNamesMap
}

func (pg *Postgres) GetDisplayNamesForAllUserProperties(projectID uint64) (int, map[string]string) {
	if projectID == 0 {
		return http.StatusBadRequest, nil
	}

	entityType := model.DisplayNameUserPropertyEntityType

	displayNameFilter := &model.DisplayName{
		ProjectID:  projectID,
		EntityType: entityType,
	}

	db := C.GetServices().Db

	var displayNames []model.DisplayName
	if err := db.Where(displayNameFilter).Find(&displayNames).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, nil
		}
		log.WithFields(log.Fields{"projectId": projectID}).WithError(err).Error(
			"Failed to GetDisplayName.")
		return http.StatusInternalServerError, nil
	}

	displayNamesMap := make(map[string]string)
	for _, displayName := range displayNames {
		displayNamesMap[displayName.PropertyName] = displayName.DisplayName
	}

	return http.StatusFound, displayNamesMap
}

func (pg *Postgres) GetDisplayNamesForObjectEntities(projectID uint64) (int, map[string]string) {
	if projectID == 0 {
		return http.StatusBadRequest, nil
	}

	entityType := model.DisplayNameObjectEntityType

	displayNameFilter := &model.DisplayName{
		ProjectID:  projectID,
		EntityType: entityType,
	}

	db := C.GetServices().Db

	var displayNames []model.DisplayName
	if err := db.Where(displayNameFilter).Find(&displayNames).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, nil
		}
		log.WithFields(log.Fields{"projectId": projectID}).WithError(err).Error(
			"Failed to GetDisplayName.")
		return http.StatusInternalServerError, nil
	}

	displayNamesMap := make(map[string]string)
	for _, displayName := range displayNames {
		if displayName.GroupName != "" {
			displayNamesMap[displayName.PropertyName] = fmt.Sprintf("%s ", displayName.GroupName)
		}
		if displayName.GroupObjectName != "" {
			displayNamesMap[displayName.PropertyName] = fmt.Sprintf("%s%s ", displayNamesMap[displayName.PropertyName], displayName.GroupObjectName)
		}
		displayNamesMap[displayName.PropertyName] = fmt.Sprintf("%s%s", displayNamesMap[displayName.PropertyName], displayName.DisplayName)
	}

	return http.StatusFound, displayNamesMap
}
