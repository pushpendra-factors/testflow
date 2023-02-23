package teams

import (
	// "bytes"
	"encoding/json"
	// "errors"
	C "factors/config"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type oauthState struct {
	ProjectID int64   `json:"project_id"`
	AgentUUID *string `json:"agent_uuid"`
}

func TeamsAuthRedirectHandler(c *gin.Context) {
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

	enOAuthState, err := json.Marshal(state)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	redirectURL := GetTeamsAuthorisationURL(C.GetTeamsTenantID(), C.GetTeamsClientID(), url.QueryEscape(string(enOAuthState)))
	c.JSON(http.StatusOK, gin.H{"redirectURL": redirectURL})

}
func GetTeamsAuthorisationURL(tenantID, clientID, state string) string {
	url := fmt.Sprintf(`https://login.microsoftonline.com/%s/organizations/oauth2/v2.0/authorize?client_id=%s&scope=group.ReadWriteAll,user.ReadAll&state=%s`, tenantID, clientID, state)
	return url
}
func TeamsCallbackHandler(c *gin.Context) {
	// Extract the code from the query string
	code := c.Query("code")
	if code == "" {
		log.Error("Failed to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	var oauthState oauthState
	// Validate the state parameter to ensure it matches what was originally sent
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

	// Use the code to get an access token from Microsoft's authorization server
	form := url.Values{}
	form.Add("client_id", C.GetTeamsClientID())
	form.Add("scope", "group.ReadWriteAll user.ReadAll")
	form.Add("code", code)
	// form.Add("redirect_uri", "http://localhost/myapp/")
	form.Add("grant_type", "authorization_code")
	form.Add("client_secret", C.GetTeamsClientSecret())

	request, err := http.NewRequest("POST", fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", C.GetTeamsTenantID()), strings.NewReader(form.Encode()))
	if err != nil {
		log.Error("Failed to create request to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	if err != nil {
		logCtx.Error("Failed to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	resp, err := client.Do(request)
	if err != nil {
		logCtx.Error("Failed to make request to get auth code")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logCtx.Error("Failed to read response body.")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	var tokens model.TeamsAccessTokens

	err = json.Unmarshal(body, &tokens)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal json response")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	// store token in db.
	//store the access token in the database
	err = store.GetStore().SetAuthTokenforTeamsIntegration(oauthState.ProjectID, *oauthState.AgentUUID, tokens)
	if err != nil {
		logCtx.Error("Failed to store access token for teams")
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
