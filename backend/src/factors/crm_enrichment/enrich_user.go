package crm_enrichment

import (
	"encoding/json"
	"errors"
	"factors/model/model"
	"factors/model/store"
	"factors/sdk"
	SDK "factors/sdk"
	U "factors/util"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func getEnrichedProperties(sourceAlias, typeAlias string, properties *map[string]interface{}) *map[string]interface{} {
	enProperties := make(map[string]interface{})
	for key, value := range *properties {
		if value == "" || value == nil {
			value = ""
		}

		enKey := model.GetCRMEnrichPropertyKeyByType(sourceAlias, typeAlias, key)
		enProperties[enKey] = value
	}

	return &enProperties
}

func getUserCustomerUserID(projectID uint64, crmUser *model.CRMUser) string {
	for _, indentifier := range model.GetIdentifierPrecendenceOrderByProjectID(projectID) {
		if indentifier == model.IdentificationTypePhone && crmUser.Phone != "" {
			return crmUser.Phone
		}

		if indentifier == model.IdentificationTypeEmail && crmUser.Email != "" {
			return crmUser.Email
		}
	}

	return ""
}

func enrichAllCRMUser(project *model.Project, config *CRMSourceConfig, crmUsers []model.CRMUser) map[string]bool {
	typeFailure := make(map[string]bool)
	for idx := range crmUsers {
		typeAlias, _ := config.GetCRMObjectTypeAlias(crmUsers[idx].Type)

		status := enrichUser(project, config, &crmUsers[idx])
		if status != http.StatusOK {
			typeFailure[typeAlias] = true
		} else if typeFailure[typeAlias] != true {
			typeFailure[typeAlias] = false
		}
	}

	return typeFailure
}

func GetCRMEventNameByAction(source, objectType string, action model.CRMAction) string {
	// for backward compatiblity with "$sf_" event name
	if source == U.CRM_SOURCE_NAME_SALESFORCE &&
		(objectType == model.SalesforceDocumentTypeNameLead || objectType == model.SalesforceDocumentTypeNameContact) {
		source = "sf"
	}

	if action == model.CRMActionCreated {
		return fmt.Sprintf("%s_%s_%s", U.NAME_PREFIX+source, objectType, "created")
	}
	if action == model.CRMActionUpdated {
		return fmt.Sprintf("%s_%s_%s", U.NAME_PREFIX+source, objectType, "updated")
	}

	return ""
}

func createOrGetUserByAction(projectID uint64, sourceAlias string, id string, userType int, action model.CRMAction, timestamp int64,
	customerUserID string) (string, error) {
	if action == model.CRMActionCreated {
		createUserID, status := store.GetStore().CreateUser(&model.User{
			ProjectId:      projectID,
			CustomerUserId: customerUserID,
			JoinTimestamp:  timestamp,
			Source:         model.GetRequestSourcePointer(model.UserSourceMap[sourceAlias]),
		})
		if status != http.StatusCreated {
			return "", errors.New("failed to create user for crm user")
		}

		return createUserID, nil
	}

	userID := ""
	if action == model.CRMActionUpdated {
		source, _ := model.GetCRMSourceByAliasName(sourceAlias)
		createdUser, status := store.GetStore().GetCRMUserByTypeAndAction(projectID, source, id, userType, model.CRMActionCreated)
		if status != http.StatusFound {
			return "", errors.New("failed to get user from crm user record")
		}

		if createdUser.UserID == "" {
			return "", errors.New("empty user id on created record")
		}

		userID = createdUser.UserID
	}

	if customerUserID != "" {

		status, _ := SDK.Identify(projectID, &SDK.IdentifyPayload{
			UserId:         userID,
			CustomerUserId: customerUserID,
			Timestamp:      timestamp,
			RequestSource:  model.UserSourceMap[sourceAlias],
		}, false)

		if status != http.StatusOK {
			return "", errors.New("failed indentifying crm user")
		}
	}

	return userID, nil
}

func enrichUser(project *model.Project, config *CRMSourceConfig, crmUser *model.CRMUser) int {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "crm_source_config": config, "crm_user": crmUser})
	if project.ID == 0 {
		logCtx.Error("Missing project_id.")
		return http.StatusBadRequest
	}

	userTypeAlias, err := config.GetCRMObjectTypeAlias(crmUser.Type)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get crm user type alias")
		return http.StatusInternalServerError
	}

	var properties map[string]interface{}
	err = json.Unmarshal(crmUser.Properties.RawMessage, &properties)
	if err != nil {
		logCtx.Error("Failed to unmarshal crm user properties.")
		return http.StatusInternalServerError
	}

	enProperties := getEnrichedProperties(config.sourceAlias, userTypeAlias, &properties)

	eventName := GetCRMEventNameByAction(config.sourceAlias, userTypeAlias, crmUser.Action)

	trackPayload := &sdk.TrackPayload{
		ProjectId:       project.ID,
		Name:            eventName,
		EventProperties: *enProperties,
		UserProperties:  *enProperties,
		RequestSource:   model.UserSourceMap[config.sourceAlias],
		Timestamp:       crmUser.Timestamp,
	}

	customerUserID := getUserCustomerUserID(project.ID, crmUser)

	userID, err := createOrGetUserByAction(project.ID, config.sourceAlias, crmUser.ID, crmUser.Type, crmUser.Action, crmUser.Timestamp, customerUserID)

	if err != nil {
		logCtx.WithError(err).Error("Failed to get user id from crm user")
		return http.StatusInternalServerError
	}

	trackPayload.UserId = userID

	status, trackResponse := sdk.Track(project.ID, trackPayload, true, config.sourceAlias, userTypeAlias)
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithFields(log.Fields{"message": trackResponse.Error, "event_name": eventName}).Error("Failed to create crm user event")
		return http.StatusInternalServerError
	}

	syncID := trackResponse.EventId
	if trackResponse.UserId != "" {
		userID = trackResponse.UserId
	}

	source, _ := model.GetCRMSourceByAliasName(config.sourceAlias)
	_, status = store.GetStore().UpdateCRMUserAsSynced(project.ID, source, crmUser, userID, syncID)
	if status != http.StatusAccepted {
		logCtx.Error("Failed to mark crm user as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
