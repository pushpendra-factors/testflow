package v1

import(
	"encoding/json"
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
	slack "factors/integration/slack"
)
type oauthState struct {
	ProjectID int64   `json:"project_id"`
	AgentUUID *string `json:"agent_uuid"`
	Source    int     `json:"source"`
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
		ProjectID: projectId,
		AgentUUID: &currentAgentUUID,
	}

	source := c.Query("source")

	if source == "2" {
		state.Source = 2
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
	var oauthState oauthState
	state := c.Query("state")
	err := json.Unmarshal([]byte(state), &oauthState)
	if err != nil || oauthState.ProjectID == 0 || *oauthState.AgentUUID == "" {
		redirectURL := buildRedirectURL("invalid values in state", oauthState.Source)
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
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
	redirectURL := buildRedirectURL("", oauthState.Source)
	c.Redirect(http.StatusPermanentRedirect, redirectURL)
	defer resp.Body.Close()
}
func buildRedirectURL(errMsg string, source int) string {
	if source == 2 {
		return C.GetProtocol() + C.GetAPPDomain() + "/welcome/visitoridentification/3?error=" + url.QueryEscape(errMsg)
	}
	return C.GetProtocol() + C.GetAPPDomain() + "/settings/integration?error=" + url.QueryEscape(errMsg)
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
	var channels []interface{}
	if _, exists := jsonResponse["channels"]; !exists {
		log.Error("Error while reading channels from json Response for Project ", projectID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading channels from json Response"})
		return
	}
	if jsonResponse["channels"] == nil {
		log.Error("Error while reading channels from json Response for Project, nil response found ", projectID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading channels from json Response"})
		return
	}
	if v, ok := jsonResponse["channels"].([]interface{}); ok {
		channels = v
	} else {
		log.Error("Error while reading channels from json Response for Project ", projectID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading channels from json Response"})
		return
	}
	var responseMetadata map[string]interface{}
	if _, exists := jsonResponse["response_metadata"]; !exists {
		log.Error("Error while reading response metadata from json Response for Project ", projectID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading response metadata from json Response"})
		return
	}
	if jsonResponse["response_metadata"] == nil {
		log.Error("Error while reading response metadata from json Response for Project, nil response found ", projectID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading response metadata from json Response"})
		return
	}
	if v, ok := jsonResponse["response_metadata"].(map[string]interface{}); ok {
		responseMetadata = v
	} else {
		log.Error("Error while reading response metadata from json Response for Project ", projectID)
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
		newChannels := jsonResponse["channels"].([]interface{})
		channels = append(channels, newChannels...)
		responseMetadata = jsonResponse["response_metadata"].(map[string]interface{})
		nextCursor = responseMetadata["next_cursor"].(string)
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
	err := store.GetStore().DeleteSlackIntegration(projectID, loggedInAgentUUID)
	if err != nil {
		log.Error("Failed to delete slack integration")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete slack integration"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "Slack integration deleted successfully"})
}