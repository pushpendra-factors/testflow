package v1

import (
	"encoding/json"
	teams "factors/integration/ms_teams"
	slack "factors/integration/slack"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	webhooks "factors/webhooks"
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type InternalStatus struct {
	Status string `json:"status"`
}

func GetAllAlertsInOneHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		return nil, http.StatusForbidden, "", "Get request failed. Invalid project ID.", true
	}

	triggers, errCode := store.GetStore().GetAllEventTriggerAlertsByProject(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, "", "Get trigger alerts failed", true
	}

	excludeSavedQueriesFlag := true
	// negation of flag to include/exlude returning saved queries i.e KPI/Events
	includeSavedQueries := c.Query("saved_queries")
	if includeSavedQueries == "true" {
		excludeSavedQueriesFlag = false
	}

	kpis, errCode := store.GetStore().GetAlertByProjectId(projectID, excludeSavedQueriesFlag)
	if errCode != http.StatusFound {
		return nil, errCode, "", "Get kpi alerts failed", true
	}

	alerts := make([]model.AlertInfo, 0)
	alerts = append(alerts, triggers...)
	alerts = append(alerts, kpis...)

	sort.Slice(alerts, func(p, q int) bool {
		return alerts[p].CreatedAt.After(alerts[q].CreatedAt)
	})

	return alerts, http.StatusOK, "", "", false
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

	obj, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(userID, oldID, projectID, &alert, userID, userID, false, nil)
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

	eta, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(userID, id, projectID, &alert, slackAssociatedUserId, teamAssociatedUserId, isPausedAlert, existingAlert.ParagonMetadata)
	if errMsg != "" || errCode != http.StatusCreated || eta == nil {
		log.WithFields(log.Fields{"project_id": projectID}).Error(errMsg)
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, errMsg, true
	}

	errCode, errMsg = store.GetStore().DeleteEventTriggerAlert(projectID, id)
	if errCode != http.StatusAccepted || errMsg != "" {
		log.WithFields(log.Fields{"project_id": projectID}).Error(errMsg)
		return nil, http.StatusBadRequest, "Cannot find any alert to update", errMsg, true
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

	factorsUrl := ""
	if webhook.IsFactorsUrlInPayload {
		if webhook.EventLevel == model.EventLevelAccount {
			factorsUrl = "https://app.factors.ai"
		} else {
			factorsUrl = "https://app.factors.ai/profiles/people"
		}

		msgPropMap[fmt.Sprintf("%d", len(messageProperties))] = model.MessagePropMapStruct{
			DisplayName: "Factors Activity URL",
			PropValue:   factorsUrl,
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

func UpdateEventTriggerAlertInternalStatusHandler(c *gin.Context) (interface{}, int, string, string, bool) {
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

func SlackTestforEventTriggerAlerts(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	agentID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid project id."})
		return nil, http.StatusForbidden, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	if agentID == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid agent id."})
		return nil, http.StatusForbidden, "INVALID_AGENT", ErrorMessages["INVALID_AGENT"], true
	}

	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"agent_id":   agentID,
	})

	var alert model.ETAConfigForSlackTest
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&alert); err != nil {
		errMsg := "Test TriggerAlert failed. Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, INVALID_INPUT, errMsg, true
	}

	messageProperties := make([][]string, 0)
	if alert.MessageProperty != nil {
		err := U.DecodePostgresJsonbToStructType(alert.MessageProperty, &messageProperties)
		if err != nil {
			errMsg := "Jsonb decoding to struct failure"
			logCtx.WithError(err).Error(errMsg)
			return nil, http.StatusBadRequest, INVALID_INPUT, errMsg, true
		}
	}

	payload := make(U.PropertiesMap)
	for i, prop := range messageProperties {
		payload[fmt.Sprintf("%d", i)], _ = U.EncodeStructTypeToMap(model.MessagePropMapStruct{
			DisplayName: prop[0],
			PropValue:   prop[1],
		})
	}

	slackChannels := make([]model.SlackChannel, 0)
	if alert.SlackChannels == nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, "No slack channels found", true
	}
	err := U.DecodePostgresJsonbToStructType(alert.SlackChannels, &slackChannels)
	if err != nil {
		logCtx.WithError(err).Error("slack channels could not be decoded")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "slack channels could not be decoded", true
	}

	slackMentions := make([]model.SlackMember, 0)
	if alert.SlackMentions != nil {
		logCtx.WithError(err).Error("failed to decode slack mentions")
		if err := U.DecodePostgresJsonbToStructType(alert.SlackMentions, &slackMentions); err != nil {
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "failed to decode slack mentions", true
		}
	}

	slackPayload := model.EventTriggerAlertMessage{
		Title:           alert.Title,
		Event:           alert.Event,
		MessageProperty: payload,
		Message:         alert.Message,
	}

	var blockMessage, slackMentionStr string
	if slackMentions != nil || alert.SlackFieldsTag != nil {
		slackMentionStr = model.GetSlackMentionsStr(slackMentions, alert.SlackFieldsTag)
	}

	isAccountAlert := alert.EventLevel == model.EventLevelAccount
	if !alert.IsHyperlinkDisabled {
		blockMessage = model.GetSlackMsgBlock(slackPayload, slackMentionStr, isAccountAlert, "", "", "")
	} else {
		blockMessage = model.GetSlackMsgBlockWithoutHyperlinks(slackPayload, slackMentionStr)
	}

	alertErrMessage := ""
	slackSuccess := true
	for _, channel := range slackChannels {
		errMsg := ""
		response, status, err := slack.SendSlackAlert(projectID, blockMessage, agentID, channel)
		if err != nil || !status {
			slackSuccess = false
			if response["error"] != nil {
				errMsg = response["error"].(string)
			}
			slackErr, exists := model.SlackErrorStates[errMsg]
			if !exists {
				alertErrMessage += fmt.Sprintf("Slack reported %s for %s channel", errMsg, channel.Name)
			} else {
				alertErrMessage += "; " + fmt.Sprintf("%s for %s channel", slackErr, channel.Name)
			}
			logCtx.WithField("channel", channel).WithError(err).Error("failed to send slack alert")
			continue
		}
	}

	if !slackSuccess {
		return nil, http.StatusNotAcceptable, "", alertErrMessage, true
	}
	return nil, http.StatusAccepted, "", "", false
}

func TeamsTestforEventTriggerAlerts(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	agentID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid project id."})
		return nil, http.StatusForbidden, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	if agentID == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid agent id."})
		return nil, http.StatusForbidden, "INVALID_AGENT", ErrorMessages["INVALID_AGENT"], true
	}

	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"agent_id":   agentID,
	})

	var alert model.ETAConfigForTeamsTest
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&alert); err != nil {
		errMsg := "Test TriggerAlert failed. Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, INVALID_INPUT, errMsg, true
	}

	messageProperties := make([][]string, 0)
	if alert.MessageProperty != nil {
		err := U.DecodePostgresJsonbToStructType(alert.MessageProperty, &messageProperties)
		if err != nil {
			errMsg := "Jsonb decoding to struct failure"
			logCtx.WithError(err).Error(errMsg)
			return nil, http.StatusBadRequest, INVALID_INPUT, errMsg, true
		}
	}

	payload := make(U.PropertiesMap)
	for i, prop := range messageProperties {
		payload[fmt.Sprintf("%d", i)], _ = U.EncodeStructTypeToMap(model.MessagePropMapStruct{
			DisplayName: prop[0],
			PropValue:   prop[1],
		})
	}

	var teamsChannels model.Team
	if alert.TeamsChannelsConfig == nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, "No teams channels found", true
	}
	err := U.DecodePostgresJsonbToStructType(alert.TeamsChannelsConfig, &teamsChannels)
	if err != nil {
		logCtx.WithError(err).Error("teams channels could not be decoded")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "teams channels could not be decoded", true
	}

	teamsPayload := model.EventTriggerAlertMessage{
		Title:           alert.Title,
		Event:           alert.Event,
		MessageProperty: payload,
		Message:         alert.Message,
	}

	isAccountAlert := alert.EventLevel == model.EventLevelAccount
	message := model.GetTeamsMsgBlock(teamsPayload, isAccountAlert, "")
	alertErrMessage := ""
	teamsSuccess := true
	for _, channel := range teamsChannels.TeamsChannelList {
		errMsg := ""
		repsonse, err := teams.SendTeamsMessage(projectID, agentID, teamsChannels.TeamsId, channel.ChannelId, message)
		if err != nil {
			errorCode, ok := repsonse["error"].(map[string]interface{})["code"].(string)
			teamsErr, exists := model.TeamsErrorStates[errMsg]
			if !ok || !exists {
				alertErrMessage += fmt.Sprintf("Teams reported %s for %s channel", errorCode, channel.ChannelName)
			} else {
				alertErrMessage += "; " + fmt.Sprintf("%s for %s channel", teamsErr, channel.ChannelName)
			}
			logCtx.WithField("channel", channel).WithError(err).Error("failed to send teams alert")
			continue
		}
	}

	if !teamsSuccess {
		return nil, http.StatusNotAcceptable, "", alertErrMessage, true
	}
	return nil, http.StatusAccepted, "", "", false
}
