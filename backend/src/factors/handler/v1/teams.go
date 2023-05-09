package v1

import (

	// "errors"
	"encoding/json"
	C "factors/config"
	teams "factors/integration/ms_teams"
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
	url := fmt.Sprintf(`https://login.microsoftonline.com/common/oauth2/v2.0/authorize?client_id=%s&response_type=code&response_mode=query&scope=Team.ReadBasic.All Channel.ReadBasic.All ChannelMessage.Send User.Read offline_access&state=%s&redirect_uri=%s`, clientID, state, getTeamsCallbackURL())
	return url
}
func TeamsCallbackHandler(c *gin.Context) {
	// Extract the code from the query string
	code := c.Query("code")
	if code == "" {
		log.Error("Failed to get auth code for teams")
		redirectURL := buildTeamsRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	var oauthState oauthState
	// Validate the state parameter to ensure it matches what was originally sent
	state := c.Query("state")
	err := json.Unmarshal([]byte(state), &oauthState)
	if err != nil || oauthState.ProjectID == 0 || *oauthState.AgentUUID == "" {
		redirectURL := buildTeamsRedirectURL("invalid values in state")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	_, status := store.GetStore().GetProjectAgentMapping(oauthState.ProjectID, *oauthState.AgentUUID)
	if status != http.StatusFound {
		redirectURL := buildTeamsRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	logCtx := log.WithFields(log.Fields{"project_id": oauthState.ProjectID, "agent_uuid": oauthState.AgentUUID})

	// Use the code to get an access token from Microsoft's authorization server
	form := url.Values{}
	form.Add("client_id", C.GetTeamsClientID())
	// form.Add("scope", "ChannelMessage.Send")
	form.Add("code", code)
	form.Add("redirect_uri", getTeamsCallbackURL())
	form.Add("grant_type", "authorization_code")
	form.Add("client_secret", C.GetTeamsClientSecret())

	request, err := http.NewRequest("POST", fmt.Sprintf("https://login.microsoftonline.com/common/oauth2/v2.0/token"), strings.NewReader(form.Encode()))
	if err != nil {
		log.Error("Failed to create request to get auth code")
		redirectURL := buildTeamsRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logCtx.Error("Failed to make request to get auth code")
		redirectURL := buildTeamsRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logCtx.Error("Failed to read response body.")
		redirectURL := buildTeamsRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	if resp.StatusCode != http.StatusOK {
		logCtx.Error("Error while requesting auth token ", string(body))
		redirectURL := buildTeamsRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	// TODO: handle other error gracefully, eg : client secret not matched.
	var tokens model.TeamsAccessTokens
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal json response")
		redirectURL := buildTeamsRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	// store token in db.
	// store the access token in the database
	err = store.GetStore().SetAuthTokenforTeamsIntegration(oauthState.ProjectID, *oauthState.AgentUUID, tokens)
	if err != nil {
		logCtx.Error("Failed to store access token for teams")
		redirectURl := buildTeamsRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURl)
	}
	redirectURL := buildTeamsRedirectURL("")
	c.Redirect(http.StatusPermanentRedirect, redirectURL)
	defer resp.Body.Close()
}
func getTeamsCallbackURL() string {
	return C.GetProtocol() + C.GetAPIDomain() + "/integrations/teams/callback"
}
func buildTeamsRedirectURL(errMsg string) string {
	return C.GetProtocol() + C.GetAPPDomain() + "/settings/integration?error=" + url.QueryEscape(errMsg)
}
func DeleteTeamsIntegrationHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if loggedInAgentUUID == "" || projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid agent id. or project id"})
		return
	}
	err := store.GetStore().DeleteTeamsIntegration(projectID, loggedInAgentUUID)
	if err != nil {
		log.Error("Failed to delete teams integration")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete teams integration"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "Teams integration deleted successfully"})
}
func GetAllTeamsHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if projectID == 0 || loggedInAgentUUID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project id or agent id"})
		return
	}
	teams, err := teams.GetAllTeams(projectID, loggedInAgentUUID)
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "agent_uuid": loggedInAgentUUID})

	if err != nil {
		logCtx.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errors": err.Error()})
	}

	c.JSON(http.StatusFound, teams)
}

func GetTeamsChannelsHandler(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid team id"})
		return
	}
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if projectID == 0 || loggedInAgentUUID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project id or agent id"})
		return
	}

	channels, err := teams.GetTeamsChannels(projectID, loggedInAgentUUID, teamID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, channels)

}

type PublisherDomain struct {
	AssociatedApplications []struct {
		ApplicationId string `json:"applicationId"`
	} `json:"associatedApplications"`
}

func VerifyPublisherDomainTeams(c *gin.Context) {
	var data PublisherDomain

	data = PublisherDomain{
		AssociatedApplications: []struct {
			ApplicationId string `json:"applicationId"`
		}{
			{
				ApplicationId: C.GetTeamsApplicationID(),
			},
		},
	}
	c.JSON(200, data)
}
