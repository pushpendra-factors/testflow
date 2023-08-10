package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	webhooks "factors/webhooks"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type InternalStatus struct {
	Status string `json:"status"`
}

func GetEventTriggerAlertsByProjectHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		return nil, http.StatusForbidden, "", "Get request failed. Invalid project ID.", true
	}
	trigger, errCode := store.GetStore().GetAllEventTriggerAlertsByProject(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, "", "Get Saved Queries failed.", true
	}

	return trigger, http.StatusOK, "", "", false

}

func CreateEventTriggerAlertHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	userID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Create failed. Invalid project id."})
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	var alert model.EventTriggerAlertConfig
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&alert); err != nil {
		errMsg := "Create TriggerAlert failed. Invalid JSON."
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}
	oldID := ""
	obj, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(userID, oldID, projectID, &alert, userID, userID, false)
	if errCode != http.StatusCreated {
		log.WithFields(log.Fields{"document": alert, "err-message": errMsg}).Error("Failed to create alert in handler.")
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	return obj.EventTriggerAlert, http.StatusCreated, "", "", false
}

func DeleteEventTriggerAlertHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Delete TriggerAlert failed. Invalid project."})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Delete failed. Invalid id provided."})
		return
	}

	errCode, errMsg := store.GetStore().DeleteEventTriggerAlert(projectID, id)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(errCode, gin.H{"Status": "OK"})
}

func EditEventTriggerAlertHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Edit TriggerAlert failed. Invalid project."})
		return nil, http.StatusBadRequest, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Edit failed. Invalid id provided."})
		return nil, http.StatusBadRequest, PROCESSING_FAILED, ErrorMessages[PROCESSING_FAILED], true
	}
	userID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	var alert model.EventTriggerAlertConfig

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&alert); err != nil {
		errMsg := "Edit TriggerAlert failed. Invalid JSON."
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}

	existingAlert, err := store.GetStore().GetEventTriggerAlertByID(id)
	if err != http.StatusFound {
		return nil, http.StatusBadRequest, "Invalid ID - Alert not found", "", true
	}
	var existingAlertPayload model.EventTriggerAlertConfig
	errObj := U.DecodePostgresJsonbToStructType(existingAlert.EventTriggerAlert, &existingAlertPayload)
	if errObj != nil {
		log.WithError(errObj).Error("Problem deserializing event_trigger_alerts query.")
		return nil, http.StatusBadRequest, "Problem deserializing event_trigger_alerts query.", "", true
	}

	var existingSlackChannels []model.SlackChannel
	if existingAlertPayload.Slack == true {
		errObj = U.DecodePostgresJsonbToStructType(existingAlertPayload.SlackChannels, &existingSlackChannels)
		if errObj != nil {
			log.WithError(errObj).Error("failed to decode slack channels")
			return nil, http.StatusBadRequest, "failed to decode slack channels", "", true
		}
	}
	var newSlackChannels []model.SlackChannel
	if alert.Slack == true {
		errObj = U.DecodePostgresJsonbToStructType(alert.SlackChannels, &newSlackChannels)
		if errObj != nil {
			log.WithError(errObj).Error("failed to decode slack channels")
			return nil, http.StatusBadRequest, "failed to decode slack channels", "", true
		}
	}

	slackAssociatedUserId := existingAlert.SlackChannelAssociatedBy
	if len(existingSlackChannels) == len(newSlackChannels) {
		existingChannelNameMap := make(map[string]bool)
		existingChannelIDMap := make(map[string]bool)
		for _, channel := range existingSlackChannels {
			existingChannelNameMap[channel.Name] = true
			existingChannelIDMap[channel.Id] = true
		}
		for _, channel := range newSlackChannels {
			if existingChannelNameMap[channel.Name] == false {
				slackAssociatedUserId = userID
				break
			}
			if existingChannelIDMap[channel.Id] == false {
				slackAssociatedUserId = userID
				break
			}
		}

	} else {
		slackAssociatedUserId = userID
	}

	var existingTeamChannels model.Team
	if existingAlertPayload.Teams {
		errObj = U.DecodePostgresJsonbToStructType(existingAlertPayload.TeamsChannelsConfig, &existingTeamChannels)
		if errObj != nil {
			log.WithError(errObj).Error("failed to decode team channels")
			return nil, http.StatusBadRequest, "failed to decode team channels", "", true
		}
	}
	var newTeamChannels model.Team
	if alert.Teams {
		errObj = U.DecodePostgresJsonbToStructType(alert.TeamsChannelsConfig, &newTeamChannels)
		if errObj != nil {
			log.WithError(errObj).Error("failed to decode new team channels")
			return nil, http.StatusBadRequest, "failed to decode new team channels", "", true
		}
	}
	teamAssociatedUserId := existingAlert.TeamsChannelAssociatedBy
	if len(existingTeamChannels.TeamsChannelList) == len(newTeamChannels.TeamsChannelList) {
		existingChannelNameMap := make(map[string]bool)
		existingChannelIDMap := make(map[string]bool)
		for _, channel := range existingTeamChannels.TeamsChannelList {
			existingChannelNameMap[channel.ChannelName] = true
			existingChannelIDMap[channel.ChannelId] = true

		}
		for _, channel := range newTeamChannels.TeamsChannelList {
			if existingChannelNameMap[channel.ChannelName] == false {
				teamAssociatedUserId = userID
				break
			}
			if existingChannelIDMap[channel.ChannelId] == false {
				teamAssociatedUserId = userID
				break
			}
		}
	} else {
		teamAssociatedUserId = userID
	}

	isPausedAlert := false
	if existingAlert.InternalStatus == model.Disabled {
		isPausedAlert = true
	}

	eta, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(userID, id, projectID, &alert, slackAssociatedUserId, teamAssociatedUserId, isPausedAlert)
	if errMsg != "" || errCode != http.StatusCreated || eta == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Edit TriggerAlert failed while updating db"})
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, ErrorMessages[PROCESSING_FAILED], true
	}

	errCode, errMsg = store.GetStore().DeleteEventTriggerAlert(projectID, id)
	if errCode != http.StatusAccepted || errMsg != "" {
		log.WithFields(log.Fields{"project_id": projectID}).Error("Cannot find any alert to update")
		return nil, http.StatusBadRequest, "Cannot find any alert to update", "", true
	}

	return alert, http.StatusAccepted, "", "", false
}

func TestWebhookforEventTriggerAlerts(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	userID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Create failed. Invalid project id."})
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	if userID == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Create failed. Invalid user id."})
		return nil, http.StatusUnauthorized, "INVALID_USER", ErrorMessages["INVALID_USER"], true
	}

	var webhook model.EventTriggerWebhook
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&webhook); err != nil {
		errMsg := "Test TriggerAlert failed. Invalid JSON."
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}

	messageProperties := make([]model.QueryGroupByProperty, 0)
	if webhook.MessageProperty != nil {
		err := U.DecodePostgresJsonbToStructType(webhook.MessageProperty, &messageProperties)
		if err != nil {
			errMsg := "Jsonb decoding to struct failure"
			log.WithError(err).Error(errMsg)
			return nil, http.StatusBadRequest, errMsg, "", true
		}
	}
	msgPropMap := make(U.PropertiesMap, 0)
	for i, mp := range messageProperties {
		var val interface{}
		if mp.Type == "datetime" {
			val = "01-01-1970"
		} else if mp.Type == "numerical" {
			val = 1234
		} else {
			val = "test"
		}
		msgPropMap[fmt.Sprintf("%d", i)] = model.MessagePropMapStruct{
			DisplayName: U.CreateVirtualDisplayName(mp.Property),
			PropValue:   val,
		}
	}
	payload := model.EventTriggerAlertMessage{
		Title:           webhook.Title,
		Event:           webhook.Event,
		Message:         webhook.Message,
		MessageProperty: msgPropMap,
	}

	response, err := webhooks.DropWebhook(webhook.Url, webhook.Secret, payload)
	if err != nil {
		errMsg := "failed to send test_webhook"
		log.WithFields(log.Fields{"project_id": projectID, "response": response, "url": webhook.Url}).WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, PROCESSING_FAILED, errMsg, true
	}

	return response, http.StatusAccepted, "", "", false
}

func GetInternalStatusForEventTriggerAlertHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		return "", http.StatusForbidden, ErrorMessages[INVALID_PROJECT], "Get internal status request failed. Invalid project ID.", true
	}
	id := c.Param("id")
	if id == "" {
		errMsg := "Get internal status failed. Invalid id provided."
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return "", http.StatusBadRequest, INVALID_INPUT, errMsg, true
	}
	status, errCode, err := store.GetStore().GetInternalStatusForEventTriggerAlert(projectID, id)
	if err != nil || errCode != http.StatusFound {
		return "", errCode, ErrorMessages[PROCESSING_FAILED], err.Error(), true
	}

	return status, http.StatusOK, "", "", false
}

func UpdateEventTriggerAlertInternalStatusHandler(c *gin.Context) (interface{}, int, string, string, bool)  {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		return "", http.StatusForbidden, ErrorMessages[INVALID_PROJECT], "Get internal status request failed. Invalid project ID.", true
	}
	id := c.Param("id")
	if id == "" {
		errMsg := "Get internal status failed. Invalid id."
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return "", http.StatusBadRequest, INVALID_INPUT, errMsg, true
	}

	var is InternalStatus
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&is); err != nil {
		errMsg := "Internal status update failed. Invalid JSON."
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}

	// Check if the status received is one of the known values
	if is.Status != model.Active && is.Status != model.Paused {
		errMsg := "Internal status update failed. Unkown value received for status"
		log.WithFields(log.Fields{"project_id": projectID, "alert_id": id}).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}

	updatedInternalStatus := ""

	if is.Status == model.Active {
		updatedInternalStatus = model.Active
	} else if is.Status == model.Paused {
		updatedInternalStatus = model.Disabled
	}

	field := map[string]interface{}{
		"internal_status": updatedInternalStatus,
	}

	errCode, err := store.GetStore().UpdateEventTriggerAlertField(projectID, id, field)
	if err != nil || errCode != http.StatusAccepted {
		errMsg := "Internal status update failed"
		log.WithFields(log.Fields{"project_id": projectID, "alert_id": id}).WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, errMsg, true
	}

	return nil, http.StatusAccepted, "", "", false
}