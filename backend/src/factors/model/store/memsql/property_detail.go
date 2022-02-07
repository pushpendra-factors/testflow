package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/util"
	U "factors/util"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesPropertyDetailForeignConstraints(propertyDetail model.PropertyDetail) int {
	logFields := log.Fields{
		"property_detail": propertyDetail,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, errCode := store.GetProject(propertyDetail.ProjectID)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

// GetPropertyTypeFromDB returns property type by key
func (store *MemSQL) GetPropertyTypeFromDB(projectID uint64, eventName, propertyKey string, isUserProperty bool) (int, *model.PropertyDetail) {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": eventName,
		"property_key": propertyKey,
		"is_user_property": isUserProperty,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
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
		if status != http.StatusFound {
			return status, nil
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
		pType := ""
		for i := range propertyDetails {
			if pType != "" && propertyDetails[i].Type != pType {
				logCtx.Error("More than one match found on getPropertyTypeFromDB.")
				return http.StatusMultipleChoices, &propertyDetails[0]
			}
			pType = propertyDetails[i].Type
		}
	}

	if len(propertyDetails) < 1 {
		return http.StatusNotFound, nil
	}

	return http.StatusFound, &propertyDetails[0]
}

// CreatePropertyDetails create configured property by user or event level. Poperty type will be overwrite if allowOverWrite is true
func (store *MemSQL) CreatePropertyDetails(projectID uint64, eventName, propertyKey, propertyType string, isUserProperty bool, allowOverWrite bool) int {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": eventName,
		"property_key": propertyKey,
		"property_type": propertyType,
		"allow_over_write": allowOverWrite,
		"is_user_property": isUserProperty,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

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

		if isUserProperty {
			status = store.updatePropertyDetails(projectID, "", propertyDetail.Key, propertyDetail.Type, propertyDetail.Entity, propertyType)
			if status != http.StatusAccepted {
				logCtx.Error("Failed to update property details.")
				return http.StatusInternalServerError
			}
		} else {
			status = store.updatePropertyDetails(projectID, *propertyDetail.EventNameID, propertyDetail.Key, propertyDetail.Type, propertyDetail.Entity, propertyType)
			if status != http.StatusAccepted {
				logCtx.Error("Failed to update property details.")
				return http.StatusInternalServerError
			}
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

	if errCode := store.satisfiesPropertyDetailForeignConstraints(*configuredProperties); errCode != http.StatusOK {
		return http.StatusInternalServerError
	}

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
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": eventName,
		"property_key": propertyKey,
		"is_user_property": isUserProperty,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
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
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": eventName,
		"property_key": propertyKey,
		"property_value": propertyValue,
		"is_user_property": isUserProperty,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
					// try removing comma separated number
					cleanedValue := strings.ReplaceAll(U.GetPropertyValueAsString(propertyValue), ",", "")
					if _, err := U.GetPropertyValueAsFloat64(cleanedValue); err != nil {
						log.WithFields(log.Fields{"project_id": projectID, "event_name": eventName, "property_key": propertyKey, "property_value": propertyValue, "is_user_property": isUserProperty}).
							WithError(err).Error("Failed to convert numerical property value.")
						return U.PropertyTypeUnknown
					}

				}

				return pType
			}

		} else if strings.HasPrefix(propertyKey, U.SALESFORCE_PROPERTY_PREFIX) { // for not configured property return categorical
			model.SetCachePropertiesType(projectID, eventName, propertyKey, U.PropertyTypeCategorical, isUserProperty, false)
			return U.PropertyTypeCategorical
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
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": eventName,
		"en_key": enKey,
		"p_type": pType,
		"allow_over_write": allowOverWrite,
		"is_user_property": isUserProperty,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": eventName,
		"key": key,
		"is_user_property": isUserProperty,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

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

	entity := model.GetEntity(isUserProperty)
	whereStmnt := "project_id = ? AND `key` = ? AND entity = ?"
	whereParams := []interface{}{projectID, key, entity}

	if !isUserProperty {
		whereStmnt = whereStmnt + " AND " + " event_name_id = ?"
		whereParams = append(whereParams, *propertyDetail.EventNameID)
	}

	db := C.GetServices().Db

	if err := db.Where(whereStmnt, whereParams...).Delete(&model.PropertyDetail{}).Error; err != nil {
		logCtx.WithError(err).Error("Failed to delete property details")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// updatePropertyDetails updates property details by event_name_id or user_property
func (store *MemSQL) updatePropertyDetails(projectID uint64, eventNameID string, key, propertyType string, entity int, newPropertyType string) int {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name_id": eventNameID,
		"key": key,
		"property_type": propertyType,
		"entity": entity,
		"new_property_type": newPropertyType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectID == 0 || key == "" || propertyType == "" || newPropertyType == "" {
		logCtx.Error("Invalid parameter.")
		return http.StatusBadRequest
	}

	if entity != model.EntityUser && eventNameID == "" {
		logCtx.Error("Invalid entity.")
		return http.StatusBadRequest
	}

	updateFields := map[string]interface{}{
		"type": newPropertyType,
	}

	whereStmnt := "project_id = ? AND `key` = ? AND type = ? AND entity = ?"
	whereParams := []interface{}{projectID, key, propertyType, entity}
	if eventNameID != "" {
		whereStmnt = whereStmnt + " AND " + "event_name_id = ?"
		whereParams = append(whereParams, eventNameID)
	}

	db := C.GetServices().Db
	query := db.Model(&model.PropertyDetail{}).Where(whereStmnt, whereParams...).Updates(updateFields)

	if err := query.Error; err != nil {
		logCtx.WithError(err).Error("Failed updating property details.")
		return http.StatusInternalServerError
	}

	if query.RowsAffected == 0 {
		return http.StatusBadRequest
	}

	return http.StatusAccepted
}

func (store *MemSQL) getPropertyDetailsForSmartEventName(projectID uint64, eventNameDetails *model.EventName) (*map[string]string, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name_details": eventNameDetails,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if !model.IsEventNameTypeSmartEvent(eventNameDetails.Type) {
		logCtx.Error("Invalid smart event type.")
		return nil, http.StatusBadRequest
	}

	smartEventFilter, err := model.GetDecodedSmartEventFilterExp(eventNameDetails.FilterExpr)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get decoded smart event filter.")
		return nil, http.StatusInternalServerError
	}

	eventName := ""
	if smartEventFilter != nil {
		if smartEventFilter.Source == model.SmartCRMEventSourceSalesforce {
			eventName = model.GetSalesforceEventNameByAction(smartEventFilter.ObjectType, model.SalesforceDocumentCreated)
		}

		if smartEventFilter.Source == model.SmartCRMEventSourceHubspot {
			if smartEventFilter.ObjectType == model.HubspotDocumentTypeNameContact {
				eventName = util.EVENT_NAME_HUBSPOT_CONTACT_CREATED
			}

			if smartEventFilter.ObjectType == model.HubspotDocumentTypeNameDeal {
				eventName = util.EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED
			}
		}
	}

	if eventName == "" {
		logCtx.Error("Empty event name.")
		return nil, http.StatusInternalServerError
	}

	eventNameDetails, status := store.GetEventName(eventName, projectID)
	if status != http.StatusFound {
		if status == http.StatusNotFound {
			return nil, status
		}

		logCtx.Error("Failed to get event name on smart event property details.")
		return nil, http.StatusInternalServerError
	}

	db := C.GetServices().Db

	var propertyDetails []model.PropertyDetail
	if err := db.Where("project_id = ? AND entity = ? AND event_name_id = ?", projectID, model.EntityEvent, eventNameDetails.ID).
		Find(&propertyDetails).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error(
			"Failed to getPropertyDetailsForSmartEventName.")
		return nil, http.StatusInternalServerError
	}

	if len(propertyDetails) < 1 {
		return nil, http.StatusNotFound
	}

	propertyDetailsMap := make(map[string]string)
	for i := range propertyDetails {
		propertyDetailsMap[propertyDetails[i].Key] = propertyDetails[i].Type
	}

	return &propertyDetailsMap, http.StatusFound
}

// GetAllPropertyDetailsByProjectID returns all property details by event_name or user_property
func (store *MemSQL) GetAllPropertyDetailsByProjectID(projectID uint64, eventName string, isUserProperty bool) (*map[string]string, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": eventName,
		"is_user_property": isUserProperty,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		return nil, http.StatusBadRequest
	}

	if !isUserProperty && eventName == "" {
		return nil, http.StatusBadRequest
	}

	entity := model.GetEntity(isUserProperty)

	logCtx := log.WithFields(logFields)

	whereStmnt := " project_id = ? AND entity = ? "
	whereParams := []interface{}{projectID, entity}
	if !isUserProperty {
		eventNameDetails, status := store.GetEventName(eventName, projectID)

		if status != http.StatusFound {
			if status == http.StatusNotFound {
				return nil, status
			}

			logCtx.Error("Failed to get event name on property details")
			return nil, http.StatusInternalServerError
		}

		if model.IsEventNameTypeSmartEvent(eventNameDetails.Type) {
			return store.getPropertyDetailsForSmartEventName(projectID, eventNameDetails)
		}

		whereStmnt = whereStmnt + " AND " + " event_name_id = ? "
		whereParams = append(whereParams, eventNameDetails.ID)
	}

	db := C.GetServices().Db

	var propertyDetails []model.PropertyDetail
	if err := db.Where(whereStmnt, whereParams...).Find(&propertyDetails).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error(
			"Failed to GetPropertyTypeFromDB.")
		return nil, http.StatusInternalServerError
	}

	if len(propertyDetails) < 1 {
		return nil, http.StatusNotFound
	}

	propertyDetailsMap := make(map[string]string)
	for i := range propertyDetails {
		propertyDetailsMap[propertyDetails[i].Key] = propertyDetails[i].Type
	}

	return &propertyDetailsMap, http.StatusFound
}
