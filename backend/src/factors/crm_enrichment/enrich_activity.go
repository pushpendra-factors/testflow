package crm_enrichment

import (
	"encoding/json"
	"errors"
	"factors/model/model"
	"factors/model/store"
	"factors/sdk"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func getActivityAssociatedUserID(projectID uint64, source model.CRMSource, config *CRMSourceConfig, crmActivity *model.CRMActivity) (string, error) {
	tableName := ""

	if config.userTypes[crmActivity.ActorType] == true {
		tableName = TableNameCRMusers
	}

	if tableName == TableNameCRMusers {
		crmUser, status := store.GetStore().GetCRMUserByTypeAndAction(projectID, source, crmActivity.ActorID, crmActivity.ActorType, model.CRMActionCreated)
		if status != http.StatusFound {
			if status == http.StatusNotFound {
				return "", errors.New("failed to get crm user record for activity association")
			}

			return "", errors.New("failed to query crm user record")
		}

		if crmUser.Synced == false {
			return "", errors.New("crm user not processed for activity association")
		}

		return crmUser.UserID, nil
	}

	return "", errors.New("invalid activity association")
}

func enrichAllCRMActivity(project *model.Project, config *CRMSourceConfig, crmActivity []model.CRMActivity) map[string]bool {
	typeFailure := make(map[string]bool)
	for idx := range crmActivity {
		typeAlias, _ := config.GetCRMObjectTypeAlias(crmActivity[idx].Type)

		status := enrichCRMActivity(project, config, &crmActivity[idx])
		if status != http.StatusOK {
			typeFailure[typeAlias] = true
		} else if typeFailure[typeAlias] != true {
			typeFailure[typeAlias] = false
		}
	}

	return typeFailure
}

func getActivityEventName(sourceAlias, name string) string {
	return fmt.Sprintf("$%s_%s", sourceAlias, name)
}
func enrichCRMActivity(project *model.Project, config *CRMSourceConfig, crmActivity *model.CRMActivity) int {
	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "crm_source_config": config, "crm_activity": crmActivity})
	if project.ID == 0 {
		logCtx.Error("Missing project_id.")
		return http.StatusBadRequest
	}

	typeAlias, err := config.GetCRMObjectTypeAlias(crmActivity.Type)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get activity type alias.")
		return http.StatusInternalServerError
	}

	var properties map[string]interface{}
	err = json.Unmarshal(crmActivity.Properties.RawMessage, &properties)
	if err != nil {
		logCtx.Error("Failed to unmarshal activity properties.")
		return http.StatusInternalServerError
	}

	enProperties := getEnrichedProperties(config.sourceAlias, typeAlias, &properties)

	trackPayload := &sdk.TrackPayload{
		Name:            getActivityEventName(config.sourceAlias, crmActivity.Name),
		ProjectId:       project.ID,
		EventProperties: *enProperties,
		RequestSource:   model.UserSourceMap[config.sourceAlias],
		Timestamp:       crmActivity.Timestamp,
	}

	source, _ := model.GetCRMSourceByAliasName(config.sourceAlias)

	userID, err := getActivityAssociatedUserID(project.ID, source, config, crmActivity)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get user for associating activity.")
		return http.StatusInternalServerError
	}

	trackPayload.UserId = userID

	status, trackResponse := sdk.Track(project.ID, trackPayload, true, config.sourceAlias, typeAlias)
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithFields(log.Fields{"error": trackResponse.Error, "message": trackResponse.Message}).
			Error("Failed to create activity event.")
		return http.StatusInternalServerError
	}

	syncID := trackResponse.EventId
	if trackResponse.UserId != "" {
		userID = trackResponse.UserId
	}

	_, status = store.GetStore().UpdateCRMActivityAsSynced(project.ID, source, crmActivity, userID, syncID)
	if status != http.StatusAccepted {
		logCtx.Error("Failed to mark crm activity as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
