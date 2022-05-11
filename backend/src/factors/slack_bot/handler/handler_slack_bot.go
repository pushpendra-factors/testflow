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
	if currentAgentUUID == "" {
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
	url := fmt.Sprintf(`https://slack.com/oauth/v2/authorize?client_id=%s&scope=channels:read,chat:write,chat:write.public,im:read&user_scope=channels:read,chat:write,groups:read,mpim:read,im:read&state=%s`, clientID, state)
	return url
}
func SlackCallbackHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		log.Error("Failed to get auth code")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get auth code"})
		return
	}
	var oauthState oauthState
	state := c.Query("state")
	err := json.Unmarshal([]byte(state), &oauthState)
	if err != nil || oauthState.ProjectID == 0 || *oauthState.AgentUUID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx := log.WithFields(log.Fields{"project_id": oauthState.ProjectID, "agent_uuid": oauthState.AgentUUID})
	request, err := http.NewRequest("POST", fmt.Sprintf("https://slack.com/api/oauth.v2.access?client_id=%s&client_secret=%s&code=%s", C.GetSlackClientID(), C.GetSlackClientSecret(), code), nil)
	if err != nil {
		log.Error("Failed to create request to get auth code")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request to get auth code"})
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logCtx.Error("Failed to get auth code")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get auth code"})
		return
	}
	var jsonResponse map[string]interface{}
	// remove this after seeing structure of json response
	fmt.Println("resp.Body", resp.Body)
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		logCtx.Error("failed to decode json response", err)
		return
	}
	access_token := jsonResponse["access_token"].(string)

	var tokens model.SlackAccessTokens
	tokens.BotAccessToken = access_token
	tokens.UserAccessToken = ""

	//store the access token in the database
	err = store.GetStore().SetAuthTokenforSlackIntegration(oauthState.ProjectID, *oauthState.AgentUUID, tokens)
	if err != nil {
		logCtx.Error("Failed to store access token for slack")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store access token for slack"})
		return
	}
	redirectURL := C.GetProtocol() + C.GetAPPDomain()
	c.Redirect(http.StatusPermanentRedirect, redirectURL)
	defer resp.Body.Close()
}
func GetSlackChannelsListHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	authToken, err := store.GetStore().GetSlackAuthToken(loggedInAgentUUID)
	if err != nil {
		log.Error("Failed to get slack auth token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get slack auth token"})
		return
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("https://slack.com/api/conversations.list"), nil)
	if err != nil {
		log.Error("Failed to create request to get slack channels list")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request to get slack channels list"})
		return
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken[projectId].BotAccessToken))
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
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if loggedInAgentUUID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid agent id."})
		return
	}
	err := store.GetStore().DeleteSlackIntegration(loggedInAgentUUID)
	if err != nil {
		log.Error("Failed to delete slack integration")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete slack integration"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "Slack integration deleted successfully"})

}

func SendSlackAlert(projectID uint64, message, agentUUID, channelID string) (bool, error) {
	// get the auth token for the agent uuid and then call the POST method to send the message
	authToken, err := store.GetStore().GetSlackAuthToken(agentUUID)
	if err != nil {
		log.Error("Failed to get access token for slack")
		return false, err
	}
	url := fmt.Sprintf("https://slack.com/api/chat.postMessage")
	// create new http post request
	reqBody := map[string]interface{}{
		"channel": channelID,
		"text":    message,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Error("Failed to marshal request body for slack")
		return false, err
	}
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	accessToken := authToken[projectID].BotAccessToken
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
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
