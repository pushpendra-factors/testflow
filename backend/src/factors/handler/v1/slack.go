package v1

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	C "factors/config"
	slack "factors/integration/slack"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type oauthState struct {
	ProjectID   int64   `json:"project_id"`
	AgentUUID   *string `json:"agent_uuid"`
	Source      int     `json:"source"`
	RandomState string  `json:"randomState"`
}

func SlackAuthRedirectHandler(c *gin.Context) {
	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if currentAgentUUID == "" || projectId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid agent id."})
		return
	}
	state := oauthState{
		ProjectID:   projectId,
		AgentUUID:   &currentAgentUUID,
		RandomState: U.RandomLowerAphaNumString(20),
	}

	source := c.Query("source")

	if source == "2" {
		state.Source = 2
	}

	randAuthState, err := json.Marshal(state)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	slack.SetCacheForSlackAuthRandomState(projectId, currentAgentUUID, state.RandomState)

	redirectURL := GetSlackAuthorisationURL(C.GetSlackClientID(), url.QueryEscape(base64.StdEncoding.EncodeToString(randAuthState)))
	c.JSON(http.StatusOK, gin.H{"redirectURL": redirectURL})
}
func GetSlackAuthorisationURL(clientID string, state string) string {

	url := fmt.Sprintf(`https://slack.com/oauth/v2/authorize?client_id=%s&scope=channels:read,chat:write,chat:write.public,users:read,users:read.email&user_scope=channels:read,chat:write,groups:read,mpim:read,users:read,users:read.email&state=%s`, clientID, state)
	return url
}
func SlackCallbackHandler(c *gin.Context) {
	var oauthState oauthState
	enState := c.Query("state")

	state, err := base64.StdEncoding.DecodeString(enState)
	if err != nil {
		redirectURL := buildRedirectURL("fail to decode state", 0)
		c.Redirect(http.StatusInternalServerError, redirectURL)
	}

	err = json.Unmarshal([]byte(state), &oauthState)
	if err != nil || oauthState.ProjectID == 0 || *oauthState.AgentUUID == "" || *&oauthState.RandomState == "" {
		redirectURL := buildRedirectURL("invalid values in state", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}

	stateFromCache, errCode := slack.GetCacheSlackAuthRandomState(*&oauthState.ProjectID, *oauthState.AgentUUID)
	if errCode != http.StatusFound {
		return
	}

	if stateFromCache != *&oauthState.RandomState {
		log.Error("Failed to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusUnauthorized, redirectURL)
	}

	code := c.Query("code")
	if code == "" {
		log.Error("Failed to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	_, status := store.GetStore().GetProjectAgentMapping(oauthState.ProjectID, *oauthState.AgentUUID)
	if status != http.StatusFound {
		redirectURL := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	logCtx := log.WithFields(log.Fields{"project_id": oauthState.ProjectID, "agent_uuid": oauthState.AgentUUID})
	request, err := http.NewRequest("POST", fmt.Sprintf("https://slack.com/api/oauth.v2.access?client_id=%s&client_secret=%s&code=%s", C.GetSlackClientID(), C.GetSlackClientSecret(), code), nil)
	if err != nil {
		log.Error("Failed to create request to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logCtx.Error("Failed to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		logCtx.Error("failed to decode json response", err)
		redirectURL := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	if jsonResponse["ok"] != true {
		logCtx.Error("Failed to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	if jsonResponse["access_token"] == nil || jsonResponse["authed_user"] == nil {
		logCtx.Error("Failed to get access token")
		redirectURL := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	access_token := jsonResponse["access_token"].(string)
	authed_user := jsonResponse["authed_user"].(map[string]interface{})
	if authed_user["access_token"] == nil {
		logCtx.Error("Failed to get access token")
		redirectURL := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	user_access_token := authed_user["access_token"].(string)

	var tokens model.SlackAccessTokens
	tokens.BotAccessToken = access_token
	tokens.UserAccessToken = user_access_token

	//store the access token in the database
	err = store.GetStore().SetAuthTokenforSlackIntegration(oauthState.ProjectID, *oauthState.AgentUUID, tokens)
	if err != nil {
		logCtx.Error("Failed to store access token for slack")
		redirectURl := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURl)
	}

	team := jsonResponse["team"].(map[string]interface{})
	if team["id"] == nil {
		logCtx.Error("Failed to get team id")
		redirectURL := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	errCode = store.GetStore().SetSlackTeamIdForProjectAgentMappings(oauthState.ProjectID, *oauthState.AgentUUID, team["id"].(string))
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to store access token for slack")
		redirectURl := buildRedirectURL("AUTH_ERROR", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURl)
	}

	redirectURL := buildRedirectURL("", oauthState.Source)
	c.Redirect(http.StatusPermanentRedirect, redirectURL)
	defer resp.Body.Close()
}
func buildRedirectURL(errMsg string, source int) string {
	if source == 2 {
		return C.GetProtocol() + C.GetAPPDomain() + "/welcome/visitoridentification/3?error=" + url.QueryEscape(errMsg)
	}
	return C.GetProtocol() + C.GetAPPDomain() + "/callback/integration/slack?error=" + url.QueryEscape(errMsg)
}

func GetSlackChannelsListHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if projectID == 0 || loggedInAgentUUID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project id or agent id"})
		return
	}
	accessTokens, err := store.GetStore().GetSlackAuthToken(projectID, loggedInAgentUUID)
	if err != nil {
		log.Error("Failed to get slack auth token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get slack auth token"})
		return
	}
	jsonResponse, status, err := slack.GetSlackChannels(accessTokens, "")
	if err != nil {
		c.JSON(status, gin.H{"error": err})
	}

	logCtx := log.WithField("project_id", projectID)
	var channels []interface{}
	if _, exists := jsonResponse["channels"]; !exists {
		logCtx.Error("Error while reading channels from json Response for Project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading channels from json Response"})
		return
	}
	if jsonResponse["channels"] == nil {
		logCtx.Error("Error while reading channels from json Response for Project, nil response found")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading channels from json Response"})
		return
	}
	if v, ok := jsonResponse["channels"].([]interface{}); ok {
		channels = v
	} else {
		logCtx.Error("Error while reading channels from json Response for Project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading channels from json Response"})
		return
	}
	var responseMetadata map[string]interface{}
	if _, exists := jsonResponse["response_metadata"]; !exists {
		logCtx.Error("Error while reading response metadata from json Response for Project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading response metadata from json Response"})
		return
	}
	if jsonResponse["response_metadata"] == nil {
		logCtx.Error("Error while reading response metadata from json Response for Project, nil response found")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading response metadata from json Response"})
		return
	}
	if v, ok := jsonResponse["response_metadata"].(map[string]interface{}); ok {
		responseMetadata = v
	} else {
		logCtx.Error("Error while reading response metadata from json Response for Project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading response metadata from json Response"})
		return
	}
	nextCursor := responseMetadata["next_cursor"].(string)
	for nextCursor != "" {
		jsonResponse, status, err = slack.GetSlackChannels(accessTokens, nextCursor)
		if err != nil {
			c.JSON(status, gin.H{"error": err})
			return
		}
		if v, ok := jsonResponse["channels"]; ok {
			newChannels := v.([]interface{})
			channels = append(channels, newChannels...)
			if metadata, ok := jsonResponse["response_metadata"]; ok {
				responseMetadata = metadata.(map[string]interface{})
				nextCursor = responseMetadata["next_cursor"].(string)
			} else {
				break
			}

		} else {
			break
		}

	}

	c.JSON(http.StatusOK, channels)
}
func DeleteSlackIntegrationHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if loggedInAgentUUID == "" || projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid agent id. or project id"})
		return
	}
	err := slack.DeleteSlackIntegration(projectID, loggedInAgentUUID)
	if err != nil {
		log.Error("Failed to delete slack integration")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete slack integration"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "Slack integration deleted successfully"})
}

func GetSlackUsersListHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if projectID == 0 || loggedInAgentUUID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project id or agent id"})
		return
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "agent_id": loggedInAgentUUID})

	users, errCode, err := store.GetStore().GetSlackUsersListFromDb(projectID, loggedInAgentUUID)
	if err != nil || errCode != http.StatusFound {
		if errCode == http.StatusNotFound {
			users, errCode, err = slack.UpdateSlackUsersListTable(projectID, loggedInAgentUUID)
			if err != nil || errCode != http.StatusOK || users == nil {
				logCtx.WithError(err).Error("failed to fetch slack users list")
				c.AbortWithStatusJSON(errCode, gin.H{"error": "failed to fetch slack users list"})
				return
			}
			//Sending only human users to the UI
			humanUsers := make([]model.SlackMember, 0)
			for _, user := range users {
				if !user.IsBot && !strings.Contains(user.Name, "slackbot") {
					humanUsers = append(humanUsers, user)
				}
			}

			c.JSON(http.StatusOK, humanUsers)
			return
		}
		logCtx.WithError(err).Error("failed to fetch slack users list from db")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "failed to fetch slack users list from db"})
		return
	}

	//Sending only human users to the UI
	humanUsers := make([]model.SlackMember, 0)
	for _, user := range users {
		if !user.IsBot && !strings.Contains(user.Name, "slackbot") {
			humanUsers = append(humanUsers, user)
		}
	}

	c.JSON(http.StatusOK, humanUsers)
}

func SlackEventListnerHandler(c *gin.Context) {

	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithError(http.StatusBadRequest, errors.New("Invalid request. Request body unavailable."))
		return
	}

	var jsonPayload map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&jsonPayload)
	if err != nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithError(http.StatusBadRequest, errors.New("Invalid request. Request body unavailable."))
		return
	}

	if jsonPayload["type"] == nil {
		logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")

		c.AbortWithError(http.StatusInternalServerError, errors.New("Tracking failed. Json Decoding failed."))
		return
	}

	if jsonPayload["type"] == "url_verification" {

		if jsonPayload["challenge"] == nil {
			logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")
			c.AbortWithError(http.StatusInternalServerError, errors.New("Tracking failed. Json Decoding failed."))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"challenge": jsonPayload["challenge"],
		})
	} else {

		if jsonPayload["event"] == nil {
			logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")
			c.AbortWithError(http.StatusInternalServerError, errors.New("Tracking failed. Json Decoding failed."))
			return
		}

		slackEvent := jsonPayload["event"].(map[string]interface{})
		if slackEvent["type"] == nil {
			logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")
			c.AbortWithError(http.StatusInternalServerError, errors.New("Tracking failed. Json Decoding failed."))
			return
		}

		if slackEvent["type"] == "app_uninstalled" {

			if jsonPayload["team_id"] == nil {
				logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")
				c.AbortWithError(http.StatusInternalServerError, errors.New("Tracking failed. Json Decoding failed."))
				return
			}
			teamId := U.GetPropertyValueAsString(jsonPayload["team_id"])

			pamList, errCode := store.GetStore().GetProjectAgentMappingFromSlackTeamId(teamId)
			if errCode != http.StatusFound {
				logCtx.WithError(err).Error("No associated agents found")
				c.AbortWithError(http.StatusInternalServerError, errors.New("No associated agents found"))
				return
			}
			for _, pam := range pamList {
				err := slack.DeleteSlackIntegration(pam.ProjectID, pam.AgentUUID)
				if err != nil {
					logCtx.WithError(err).Error("Slack accsess Token removal failed.")
					c.AbortWithError(http.StatusInternalServerError, errors.New("Slack accsess Token removal failed."))
					return
				}
			}

			c.JSON(http.StatusOK, nil)
		}

	}

}
