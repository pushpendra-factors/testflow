package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// GetPropertyTypeFromDB returns property type by key
func (pg *Postgres) GetPropertyTypeFromDB(projectID uint64, eventName, propertyKey string, isUserProperty bool) (int, *model.PropertyDetail) {
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
		eventNameDetails, status := pg.GetEventName(eventName, projectID)
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

// CreatePropertyDetails create configured property by user or event level
func (pg *Postgres) CreatePropertyDetails(projectID uint64, eventName, propertyKey, propertyType string, isUserProperty bool) int {
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

	status, _ := pg.GetPropertyTypeFromDB(projectID, eventName, propertyKey, isUserProperty)
	if status == http.StatusFound {
		return http.StatusConflict
	}

	if !isUserProperty {
		event, status := pg.GetEventName(eventName, projectID)
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
func (pg *Postgres) getPreConfiguredPropertyTypeByName(projectID uint64, eventName, propertyKey string, isUserProperty bool) (bool, string) {
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

	status, configuredProperty := pg.GetPropertyTypeFromDB(projectID, eventName, propertyKey, isUserProperty)
	if status != http.StatusFound {
		return false, ""
	}

	model.SetCachePropertiesType(projectID, eventName, propertyKey, configuredProperty.Type, isUserProperty, true)
	return true, configuredProperty.Type
}

// GetPropertyTypeByKeyValue returns property type by key, prioritize preconfigured type or uses type casting
func (pg *Postgres) GetPropertyTypeByKeyValue(projectID uint64, eventName string, propertyKey string, propertyValue interface{}, isUserProperty bool) string {

	enabledPropertyTypeCheckFromDB := C.IsEnabledPropertyDetailFromDB() && C.IsEnabledPropertyDetailByProjectID(projectID)

	if enabledPropertyTypeCheckFromDB {
		if propertyKey != "" {
			if preConfigured, pType := pg.getPreConfiguredPropertyTypeByName(projectID, eventName, propertyKey, isUserProperty); preConfigured {
				if _, err := U.GetPropertyValueAsFloat64(propertyValue); err != nil {
					log.WithFields(log.Fields{"project_id": projectID, "event_name": eventName, "property_key": propertyKey, "is_user_property": isUserProperty}).
						Error("Failed to convert configured property value.")
					return ""
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
