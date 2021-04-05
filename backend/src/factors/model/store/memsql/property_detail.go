package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// GetPropertyTypeFromDB returns property type by key
func (store *MemSQL) GetPropertyTypeFromDB(projectID uint64, eventName, propertyKey string, isUserProperty bool) (int, *model.PropertyDetail) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_name": eventName, "property_key": propertyKey})
	if projectID == 0 || propertyKey == "" {
		return http.StatusBadRequest, nil
	}

	if !isUserProperty && eventName == "" {
		return http.StatusBadRequest, nil
	}

	propertyLevel := model.GetEntity(isUserProperty)

	propertyDetail := &model.PropertyDetail{
		ProjectID: projectID,
		Key:       propertyKey,
		Entity:    propertyLevel,
	}

	if !isUserProperty {
		eventNameDetails, status := store.GetEventName(eventName, projectID)
		if status == http.StatusNotFound {
			return http.StatusBadRequest, nil
		}
		propertyDetail.EventNameID = &eventNameDetails.ID
	}

	db := C.GetServices().Db

	var propertyDetails []model.PropertyDetail
	if err := db.Where(propertyDetail).Find(&propertyDetails).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectID, "event_name": eventName, "property_key": propertyKey}).WithError(err).Error(
			"Failed to GetPropertyTypeFromDB.")
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, nil
		}
		return http.StatusInternalServerError, nil
	}

	if len(propertyDetails) > 1 {
		logCtx.Error("More than one match found on getPropertyTypeFromDB.")
		return http.StatusMultipleChoices, nil
	}

	if len(propertyDetails) < 1 {
		return http.StatusNotFound, nil
	}

	return http.StatusFound, &propertyDetails[0]
}

// CreatePropertyDetails create configured property by user or event level. Poperty type will be overwrite if allowOverWrite is true
func (store *MemSQL) CreatePropertyDetails(projectID uint64, eventName, propertyKey, propertyType string, isUserProperty bool, allowOverWrite bool) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_name": eventName, "property_key": propertyKey})

	if propertyKey == "" || projectID == 0 ||
		(propertyType != U.PropertyTypeDateTime && propertyType != U.PropertyTypeNumerical) {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest
	}

	if !isUserProperty && eventName == "" {
		logCtx.Error("Missing event_name.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	configuredProperties := &model.PropertyDetail{
		ProjectID: projectID,
		Key:       propertyKey,
		Type:      propertyType,
	}

	status, propertyDetail := store.GetPropertyTypeFromDB(projectID, eventName, propertyKey, isUserProperty)
	if status == http.StatusFound {
		if propertyDetail.Type == propertyType || !allowOverWrite {
			return http.StatusConflict
		}

		status = store.updatePropertyDetails(projectID, *propertyDetail.EventNameID, propertyDetail.Key, propertyDetail.Type, propertyDetail.Entity, propertyType)
		if status != http.StatusAccepted {
			logCtx.Error("Failed to update property details.")
			return http.StatusInternalServerError
		}

		return http.StatusAccepted
	}

	if !isUserProperty {
		event, status := store.GetEventName(eventName, projectID)
		if status != http.StatusFound {
			logCtx.Error("Failed to get event name.")
			return http.StatusBadRequest
		}
		configuredProperties.EventNameID = &event.ID
	}

	configuredProperties.Entity = model.GetEntity(isUserProperty)

	if err := db.Create(configuredProperties).Error; err != nil {

		if U.IsPostgresUniqueIndexViolationError("configured_properties_pkey", err) {
			return http.StatusConflict
		}

		logCtx.WithError(err).Error("Failed to CreatePropertyDetails.")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

// getPreConfiguredPropertyTypeByName returns if property is configured and property type from cache or DB. Only returns datetime, numerical or unknown.
func (store *MemSQL) getPreConfiguredPropertyTypeByName(projectID uint64, eventName, propertyKey string, isUserProperty bool) (bool, string) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_name": eventName, "property_key": propertyKey})
	if projectID == 0 || propertyKey == "" {
		logCtx.Error("Missing required field.")
		return false, ""
	}

	if !isUserProperty && eventName == "" {
		logCtx.Error("Missing event name field.")
		return false, ""
	}

	propertyStatus := model.GetCachePropertiesType(projectID, eventName, propertyKey, isUserProperty)
	if propertyStatus != model.TypeMissingConfiguredProperties {
		if propertyStatus == model.TypeNonConfiguredProperties {
			return false, ""
		}

		if propertyStatus == model.TypeConfiguredDatetimeProperties {
			return true, U.PropertyTypeDateTime
		}

		if propertyStatus == model.TypeConfiguredNumericalProperties {
			return true, U.PropertyTypeNumerical
		}

		return false, ""
	}

	status, configuredProperty := store.GetPropertyTypeFromDB(projectID, eventName, propertyKey, isUserProperty)
	if status != http.StatusFound {
		return false, ""
	}

	model.SetCachePropertiesType(projectID, eventName, propertyKey, configuredProperty.Type, isUserProperty, true)
	return true, configuredProperty.Type
}

// GetPropertyTypeByKeyValue returns property type by key, prioritize preconfigured type or uses type casting
func (store *MemSQL) GetPropertyTypeByKeyValue(projectID uint64, eventName string, propertyKey string, propertyValue interface{}, isUserProperty bool) string {

	enabledPropertyTypeCheckFromDB := C.IsEnabledPropertyDetailFromDB() && C.IsEnabledPropertyDetailByProjectID(projectID)

	if enabledPropertyTypeCheckFromDB && propertyKey != "" {
		if preConfigured, pType := store.getPreConfiguredPropertyTypeByName(projectID, eventName, propertyKey, isUserProperty); preConfigured {

			if pType == U.PropertyTypeDateTime {
				err := model.ValidateDateTimeProperty(propertyKey, propertyValue)

				if err != nil {
					if err == model.ErrorUsingSalesforceDatetimeTemplate {
						log.WithFields(log.Fields{"project_id": projectID, "event_name": eventName, "property_key": propertyKey, "property_value": propertyValue, "is_user_property": isUserProperty}).
							Warn(err)
						return pType
					}

					log.WithFields(log.Fields{"project_id": projectID, "event_name": eventName, "property_key": propertyKey, "property_value": propertyValue, "is_user_property": isUserProperty}).
						WithError(err).Error("Failed to convert configured property value.")
					return U.PropertyTypeUnknown
				}

				return pType
			}

			if pType == U.PropertyTypeNumerical {
				if _, err := U.GetPropertyValueAsFloat64(propertyValue); err != nil {
					log.WithFields(log.Fields{"project_id": projectID, "event_name": eventName, "property_key": propertyKey, "property_value": propertyValue, "is_user_property": isUserProperty}).
						WithError(err).Error("Failed to convert numerical property value.")
					return U.PropertyTypeUnknown
				}

				return pType
			}

		}
	}

	pType, internalProperty := U.GetPropertyTypeByKeyORValue(projectID, eventName, propertyKey, propertyValue, isUserProperty)

	if enabledPropertyTypeCheckFromDB && !internalProperty { // do not add to cache if it is internal property
		model.SetCachePropertiesType(projectID, eventName, propertyKey, pType, isUserProperty, false)
	}

	return pType
}

/*
CreateOrDeletePropertyDetails creates or delete property details by type.
WARNING Unkown type would be deleted from DB is existed
*/
func (store *MemSQL) CreateOrDeletePropertyDetails(projectID uint64, eventName, enKey, pType string, isUserProperty, allowOverWrite bool) error {
	if projectID == 0 || enKey == "" || pType == "" {
		return errors.New("missing required field")
	}

	if !isUserProperty && eventName == "" {
		return errors.New("missing event_name")
	}

	if pType == U.PropertyTypeUnknown { // unknown type would be deleted

		status := store.deletePropertyDetailsIfExist(projectID, eventName, enKey, isUserProperty)
		if status != http.StatusOK && status != http.StatusNotFound {
			return errors.New("failed to delete property details for created event")
		}

		return nil
	}

	status := store.CreatePropertyDetails(projectID, eventName, enKey, pType, isUserProperty, allowOverWrite)
	if status != http.StatusCreated && status != http.StatusConflict && status != http.StatusAccepted {
		return errors.New("failed to create created event property details")
	}

	return nil
}

// deletePropertyDetailsIfExist delete property details by event_name_id or user_property if exists
func (store *MemSQL) deletePropertyDetailsIfExist(projectID uint64, eventName, key string, isUserProperty bool) int {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "property_key": key, "event_name": eventName})

	if projectID == 0 || key == "" {
		return http.StatusBadRequest
	}

	if !isUserProperty && eventName == "" {
		return http.StatusBadRequest
	}

	status, propertyDetail := store.GetPropertyTypeFromDB(projectID, eventName, key, isUserProperty)
	if status != http.StatusFound {
		return http.StatusNotFound
	}

	db := C.GetServices().Db

	if err := db.Delete(propertyDetail).Error; err != nil {
		logCtx.WithError(err).Error("Failed to delete property details")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// updatePropertyDetails updates property details by event_name_id or user_property
func (store *MemSQL) updatePropertyDetails(projectID uint64, eventNameID uint64, key, propertyType string, entity int, newPropertyType string) int {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "property_key": key, "property_type": propertyType, "event_name_id": eventNameID})

	if projectID == 0 || key == "" || propertyType == "" || newPropertyType == "" {
		return http.StatusBadRequest
	}

	if entity != model.EntityUser && eventNameID == 0 {
		return http.StatusBadRequest
	}

	updateFields := map[string]interface{}{
		"type": newPropertyType,
	}

	db := C.GetServices().Db
	query := db.Model(&model.PropertyDetail{}).Where("project_id = ? AND `key` = ? AND `type` = ? ",
		projectID, key, propertyType).Updates(updateFields)

	if err := query.Error; err != nil {
		logCtx.WithError(err).Error("Failed updating property details.")
		return http.StatusInternalServerError
	}

	if query.RowsAffected == 0 {
		return http.StatusBadRequest
	}

	return http.StatusAccepted
}

// GetAllPropertyDetailsByProjectID returns all property details by event_name or user_property
func (store *MemSQL) GetAllPropertyDetailsByProjectID(projectID uint64, eventName string, isUserProperty bool) (int, *map[string]string) {
	if projectID == 0 {
		return http.StatusBadRequest, nil
	}

	if !isUserProperty && eventName == "" {
		return http.StatusBadRequest, nil
	}

	entity := model.GetEntity(isUserProperty)

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_name": eventName, "entity": entity})

	whereStmnt := " project_id = ? AND entity = ? "
	whereParams := []interface{}{projectID, entity}
	if !isUserProperty {
		eventNameDetails, status := store.GetEventName(eventName, projectID)

		if status != http.StatusFound {
			if status == http.StatusNotFound {
				return status, nil
			}

			logCtx.Error("Failed to get event name on property details")
			return http.StatusInternalServerError, nil
		}

		whereStmnt = whereStmnt + " AND " + " event_name_id = ? "
		whereParams = append(whereParams, eventNameDetails.ID)
	}

	db := C.GetServices().Db

	var propertyDetails []model.PropertyDetail
	if err := db.Where(whereStmnt, whereParams...).Find(&propertyDetails).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, nil
		}

		logCtx.WithError(err).Error(
			"Failed to GetPropertyTypeFromDB.")
		return http.StatusInternalServerError, nil
	}

	if len(propertyDetails) < 1 {
		return http.StatusNotFound, nil
	}

	propertyDetailsMap := make(map[string]string)
	for i := range propertyDetails {
		propertyDetailsMap[propertyDetails[i].Key] = propertyDetails[i].Type
	}

	return http.StatusFound, &propertyDetailsMap
}
