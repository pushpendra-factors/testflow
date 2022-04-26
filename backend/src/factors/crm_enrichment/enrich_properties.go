package crm_enrichment

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func getBatchedPropertiesByTable(projectID uint64, properties []model.CRMProperty, config *CRMSourceConfig) ([]*model.CRMProperty, []*model.CRMProperty) {
	userProperties, activityProperties := make([]*model.CRMProperty, 0), make([]*model.CRMProperty, 0)
	for i := range properties {

		if config.activityTypes[properties[i].Type] {
			activityProperties = append(activityProperties, &properties[i])
		}

		if config.userTypes[properties[i].Type] {
			userProperties = append(userProperties, &properties[i])
		}
	}

	return userProperties, activityProperties
}

func SyncProperties(projectID uint64, sourceConfig *CRMSourceConfig) []EnrichStatus {

	properties, status := store.GetStore().GetCRMPropertiesForSync(projectID)
	if status != http.StatusFound {
		if status == http.StatusNotFound {
			return nil
		}

		return []EnrichStatus{{TableName: "crm_properties", Status: U.CRM_SYNC_STATUS_FAILURES}}
	}

	userProperties, activityProperties := getBatchedPropertiesByTable(projectID, properties, sourceConfig)

	overAlltableSyncStatus := []EnrichStatus{}
	for tableName, properties := range map[string][]*model.CRMProperty{
		TableNameCRMusers:      userProperties,
		TableNameCRMActivities: activityProperties,
	} {
		if len(properties) < 1 {
			continue
		}

		tableSyncStatus := EnrichStatus{TableName: tableName}
		tablefailure := false
		switch tableName {
		case TableNameCRMusers:
			tablefailure = syncAllUserProperty(projectID, userProperties, sourceConfig)
		case TableNameCRMActivities:
			tablefailure = syncAllActivityProperties(projectID, activityProperties, sourceConfig)
		}

		if tablefailure {
			tableSyncStatus.Status = U.CRM_SYNC_STATUS_FAILURES
		} else {
			tableSyncStatus.Status = U.CRM_SYNC_STATUS_SUCCESS
		}
		overAlltableSyncStatus = append(overAlltableSyncStatus, tableSyncStatus)
	}

	return overAlltableSyncStatus
}

func syncAllUserProperty(projectID uint64, properties []*model.CRMProperty, sourceConfig *CRMSourceConfig) bool {
	anyFailure := false
	for i := range properties {
		status := syncUserProperty(projectID, properties[i], sourceConfig)
		if status != http.StatusOK {
			anyFailure = true
		}
	}

	return anyFailure
}

func syncUserProperty(projectID uint64, property *model.CRMProperty, sourceConfig *CRMSourceConfig) int {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "crm_property": property, "source_config": sourceConfig})

	if projectID == 0 {
		logCtx.Error("Invalid project id.")
		return http.StatusBadRequest
	}

	if !sourceConfig.userTypes[property.Type] {
		logCtx.Error("Invalid property type.")
		return http.StatusBadRequest
	}

	objectTypeAlias, err := sourceConfig.GetCRMObjectTypeAlias(property.Type)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get crm object alias.")
		return http.StatusInternalServerError
	}

	eventNames := []string{}
	for _, action := range []model.CRMAction{model.CRMActionCreated, model.CRMActionUpdated} {
		eventNames = append(eventNames, GetCRMEventNameByAction(sourceConfig.sourceAlias, objectTypeAlias, action))
	}

	enKey := model.GetCRMEnrichPropertyKeyByType(sourceConfig.sourceAlias, objectTypeAlias, property.Name)
	for _, eventName := range eventNames {
		// create event name before creating properties
		_, status := store.GetStore().CreateOrGetEventName(&model.EventName{
			ProjectId: projectID,
			Name:      eventName,
			Type:      model.TYPE_USER_CREATED_EVENT_NAME,
		})

		if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
			logCtx.Error("Failed to create event name on sync crm properties.")
			return http.StatusInternalServerError
		}

		err = store.GetStore().CreateOrDeletePropertyDetails(projectID, eventName, enKey, property.MappedDataType, false, true)
		if err != nil {
			logCtx.WithFields(log.Fields{"enriched_property_key": enKey, "event_name": eventName}).WithError(err).
				Error("Failed to crated event property details.")
			return http.StatusInternalServerError
		}
	}

	err = store.GetStore().CreateOrDeletePropertyDetails(projectID, "", enKey, property.MappedDataType, true, true)
	if err != nil {
		logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).WithError(err).
			Error("Failed to create user property details.")
		return http.StatusInternalServerError
	}

	if property.Label != "" {
		logCtx.Info("Inserting display names")
		status := store.GetStore().CreateOrUpdateDisplayNameByObjectType(projectID, enKey,
			objectTypeAlias, property.Label, sourceConfig.sourceAlias)
		if status != http.StatusCreated && status != http.StatusConflict {
			logCtx.Error("Failed to create or update display name")
			return http.StatusInternalServerError
		}
	}

	source, _ := model.GetCRMSourceByAliasName(sourceConfig.sourceAlias)

	_, status := store.GetStore().UpdateCRMProperyAsSynced(projectID, source, property)

	if status != http.StatusAccepted {
		logCtx.Error("Failed to mark crm properties as synced.")
		return status
	}

	return http.StatusOK
}

func getAllPropertiesObjectType(properties []*model.CRMProperty) []int {
	propertyObjectTypeMap := make(map[int]bool)
	propertyObjectTypes := []int{}
	for i := range properties {
		if _, exist := propertyObjectTypeMap[properties[i].Type]; !exist {
			propertyObjectTypeMap[properties[i].Type] = true
			propertyObjectTypes = append(propertyObjectTypes, properties[i].Type)
		}
	}

	return propertyObjectTypes
}

func syncAllActivityProperties(projectID uint64, properties []*model.CRMProperty, sourceConfig *CRMSourceConfig) bool {
	types := getAllPropertiesObjectType(properties)
	typeNames, status := store.GetStore().GetActivitiesDistinctEventNamesByType(projectID, types)
	if status != http.StatusFound {
		log.WithFields(log.Fields{"project_id": projectID, "properties": properties, "err_code": status}).
			Error("Failed to get name for activites properties.")
		return true
	}

	anyFailure := false
	for i := range properties {
		status := syncActivityProperty(projectID, properties[i], typeNames[properties[i].Type], sourceConfig)
		if status != http.StatusOK {
			anyFailure = true
		}
	}

	return anyFailure
}

func syncActivityProperty(projectID uint64, property *model.CRMProperty, activityName []string, sourceConfig *CRMSourceConfig) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "crm_property": property, "source_config": sourceConfig})

	if !sourceConfig.activityTypes[property.Type] {
		logCtx.Error("Invalid activity property type.")
		return http.StatusBadRequest
	}

	if projectID == 0 {
		logCtx.Error("Invalid project id.")
		return http.StatusBadRequest
	}

	objectTypeAlias, err := sourceConfig.GetCRMObjectTypeAlias(property.Type)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get crm object alias.")
		return http.StatusInternalServerError
	}

	enKey := model.GetCRMEnrichPropertyKeyByType(sourceConfig.sourceAlias, objectTypeAlias, property.Name)
	// only update event property data type
	for _, name := range activityName {
		eventName := getActivityEventName(sourceConfig.sourceAlias, name)

		// create event name before adding property details
		_, status := store.GetStore().CreateOrGetEventName(&model.EventName{
			ProjectId: projectID,
			Name:      eventName,
			Type:      model.TYPE_USER_CREATED_EVENT_NAME,
		})

		if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
			logCtx.Error("Failed to create event name on sync crm properties.")
			return http.StatusInternalServerError
		}

		err = store.GetStore().CreateOrDeletePropertyDetails(projectID, eventName, enKey, property.MappedDataType, false, true)
		if err != nil {
			logCtx.WithFields(log.Fields{"enriched_property_key": enKey, "event_name": eventName}).WithError(err).
				Error("Failed to crated event property details.")
			return http.StatusInternalServerError
		}
	}

	if property.Label != "" {
		logCtx.Info("Inserting display names")
		status := store.GetStore().CreateOrUpdateDisplayNameByObjectType(projectID, enKey,
			objectTypeAlias, property.Label, sourceConfig.sourceAlias)
		if status != http.StatusCreated && status != http.StatusConflict {
			logCtx.Error("Failed to create or update display name on activity properties.")
			return http.StatusInternalServerError
		}
	}

	source, _ := model.GetCRMSourceByAliasName(sourceConfig.sourceAlias)
	_, status := store.GetStore().UpdateCRMProperyAsSynced(projectID, source, property)

	if status != http.StatusAccepted {
		logCtx.Error("Failed to mark crm properties as synced.")
		return status
	}

	return http.StatusOK
}
