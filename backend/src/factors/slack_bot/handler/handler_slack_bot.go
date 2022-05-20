package slack_alert

import (
	"bytes"
	"encoding/json"
	"errors"
	C "factors/config"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

type oauthState struct {
	ProjectID uint64  `json:"project_id"`
	AgentUUID *string `json:"agent_uuid"`
}

func SlackAuthRedirectHandler(c *gin.Context) {
	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if currentAgentUUID == "" || projectId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid agent id."})
		return
	}
	state := oauthState{
		ProjectID: projectId,
		AgentUUID: &currentAgentUUID,
	}

	enOAuthState, err := json.Marshal(state)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	redirectURL := GetSlackAuthorisationURL(C.GetSlackClientID(), url.QueryEscape(string(enOAuthState)))
	c.JSON(http.StatusOK, gin.H{"redirectURL": redirectURL})
}
func GetSlackAuthorisationURL(clientID string, state string) string {
	url := fmt.Sprintf(`https://slack.com/oauth/v2/authorize?client_id=%s&scope=channels:read,chat:write,chat:write.public,im:read&user_scope=channels:read,chat:write,groups:read,mpim:read&state=%s`, clientID, state)
	return url
}
func SlackCallbackHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		log.Error("Failed to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	var oauthState oauthState
	state := c.Query("state")
	err := json.Unmarshal([]byte(state), &oauthState)
	if err != nil || oauthState.ProjectID == 0 || *oauthState.AgentUUID == "" {
		redirectURL := buildRedirectURL("invalid values in state")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	_, status := store.GetStore().GetProjectAgentMapping(oauthState.ProjectID, *oauthState.AgentUUID)
	if status != http.StatusFound {
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	logCtx := log.WithFields(log.Fields{"project_id": oauthState.ProjectID, "agent_uuid": oauthState.AgentUUID})
	request, err := http.NewRequest("POST", fmt.Sprintf("https://slack.com/api/oauth.v2.access?client_id=%s&client_secret=%s&code=%s", C.GetSlackClientID(), C.GetSlackClientSecret(), code), nil)
	if err != nil {
		log.Error("Failed to create request to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logCtx.Error("Failed to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		logCtx.Error("failed to decode json response", err)
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	if jsonResponse["ok"] != true {
		logCtx.Error("Failed to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	if jsonResponse["access_token"] == nil || jsonResponse["authed_user"] == nil {
		logCtx.Error("Failed to get access token")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	access_token := jsonResponse["access_token"].(string)
	authed_user := jsonResponse["authed_user"].(map[string]interface{})
	if authed_user["access_token"] == nil {
		logCtx.Error("Failed to get access token")
		redirectURL := buildRedirectURL("AUTH_ERROR")
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
		redirectURl := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURl)
	}
	redirectURL := buildRedirectURL("")
	c.Redirect(http.StatusPermanentRedirect, redirectURL)
	defer resp.Body.Close()
}
func buildRedirectURL(errMsg string) string {
	return C.GetProtocol() + C.GetAPPDomain() + "/settings/integration?error=" + url.QueryEscape(errMsg)
}
func GetSlackChannelsListHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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
	// add query params
	request, err := http.NewRequest("GET", fmt.Sprintf("https://slack.com/api/conversations.list"), nil)
	if err != nil {
		log.Error("Failed to create request to get slack channels list")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request to get slack channels list"})
		return
	}
	q := request.URL.Query()
	q.Add("types", "public_channel,private_channel,mpim")
	q.Add("limit", "2000")
	request.URL.RawQuery = q.Encode()
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessTokens.UserAccessToken))
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Error("Failed to get slack channels list")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get slack channels list"})
		return
	}
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		log.Error("failed to decode json response", err)
	}
	channels := jsonResponse["channels"].([]interface{})
	c.JSON(http.StatusOK, channels)
}
func DeleteSlackIntegrationHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if loggedInAgentUUID == "" || projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid agent id. or project id"})
		return
	}
	err := store.GetStore().DeleteSlackIntegration(projectID, loggedInAgentUUID)
	if err != nil {
		log.Error("Failed to delete slack integration")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete slack integration"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "Slack integration deleted successfully"})

}

func SendSlackAlert(projectID uint64, message, agentUUID string, channel model.SlackChannel) (bool, error) {
	// get the auth token for the agent uuid and then call the POST method to send the message
	accessTokens, err := store.GetStore().GetSlackAuthToken(projectID, agentUUID)
	if err != nil {
		log.Error("Failed to get access token for slack")
		return false, err
	}
	url := fmt.Sprintf("https://slack.com/api/chat.postMessage")
	// create new http post request
	reqBody := map[string]interface{}{
		"channel": channel.ChannelID,
		"text":    message,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Error("Failed to marshal request body for slack")
		return false, err
	}
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	if channel.IsPrivate {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessTokens.UserAccessToken))
	} else {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessTokens.BotAccessToken))
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Error("Failed to send slack alert")
		return false, err
	}
	var response string
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error("Failed to decode response from slack")
		return false, err
	}
	defer resp.Body.Close()
	if response == "ok" {
		return true, nil
	}
	return false, errors.New("Failed to send slack alert")
}
