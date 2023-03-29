package teams

import (
	"bytes"
	"encoding/json"
	"errors"
	// "errors"
	C "factors/config"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type oauthState struct {
	ProjectID int64   `json:"project_id"`
	AgentUUID *string `json:"agent_uuid"`
}
type TeamsMessage struct {
	Body struct {
		Content string `json:"content"`
	} `json:"body"`
}

// teams and channels related structs
type Team struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type Group struct {
	Id      string `json:"id"`
	WebUrl  string `json:"webUrl"`
	IsTeam  bool   `json:"isTeam"`
	GroupId string `json:"groupId"`
}

type Channel struct {
	Id   string `json:"id"`
	Name string `json:"displayName"`
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
	url := fmt.Sprintf(`https://login.microsoftonline.com/%s/oauth2/v2.0/authorize?client_id=%s&response_type=code&response_mode=query&scope=offline_access ChannelMessage.Send Channel.ReadBasic.All Team.ReadBasic.All&state=%s`, tenantID, clientID, state)
	return url
}
func TeamsCallbackHandler(c *gin.Context) {
	// Extract the code from the query string
	code := c.Query("code")
	if code == "" {
		log.Error("Failed to get auth code for teams")
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
	// form.Add("scope", "ChannelMessage.Send")
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
	// TODO: handle other error gracefully, eg : client secret not matched.
	var tokens model.TeamsAccessTokens

	err = json.Unmarshal(body, &tokens)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal json response")
		redirectURL := buildRedirectURL("AUTH_ERROR")
		c.Redirect(http.StatusPermanentRedirect, redirectURL)
	}
	// store token in db.
	// store the access token in the database
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
func SendTeamsMessage(projectID int64, agentUUID, teamID, channelID, message string) error {
	tokens, err := store.GetStore().GetTeamsAuthTokens(projectID, agentUUID)
	if err != nil {
		return errors.New("Failed to get access tokens for teams")
	}
	teamsMessage := TeamsMessage{Body: struct {
		Content string "json:\"content\""
	}{Content: message}}

	jsonMessage, err := json.Marshal(teamsMessage)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/teams/%s/channels/%s/messages", teamID, channelID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonMessage))
	if err != nil {
		return err
	}
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonMessage))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return errors.New(fmt.Sprintf("failed to send Teams message: %v", resp.Status))
	}

	return nil
}
func GetAllTeamsHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if projectID == 0 || loggedInAgentUUID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project id or agent id"})
		return
	}
	teams, err := getAllTeams(projectID, loggedInAgentUUID)
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "agent_uuid": loggedInAgentUUID})

	if err != nil {
		logCtx.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errors": err})
	}

	c.JSON(http.StatusFound, teams)
}

// func to get list of teams
func getAllTeams(projectID int64, agentUUID string) ([]Team, error) {
	tokens, err := store.GetStore().GetTeamsAuthTokens(projectID, agentUUID)
	if err != nil {
		return []Team{}, errors.New("Failed to get access tokens for teams")
	}

	url := "https://graph.microsoft.com/v1.0/me/joinedTeams"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get Teams teams: %v", resp.Status)
	}

	var teamList struct {
		Value []Team `json:"value"`
	}

	err = json.NewDecoder(resp.Body).Decode(&teamList)
	if err != nil {
		return nil, err
	}

	return teamList.Value, nil
}

func GetTeamsChannelsHandler(c *gin.Context) {
	teamID := c.Query("teams_id")
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

	channels, err := getTeamsChannels(projectID, loggedInAgentUUID, teamID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusAccepted, channels)

}

// func to get list of channels in a team.
func getTeamsChannels(projectID int64, agentUUID, teamID string) ([]Channel, error) {
	tokens, err := store.GetStore().GetTeamsAuthTokens(projectID, agentUUID)
	if err != nil {
		return []Channel{}, errors.New("Failed to get access tokens for teams")
	}
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/teams/%s/channels", teamID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get Teams channels: %v", resp.Status)
	}

	var channelList struct {
		Value []Channel `json:"value"`
	}

	err = json.NewDecoder(resp.Body).Decode(&channelList)
	if err != nil {
		return nil, err
	}

	return channelList.Value, nil
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
type PublisherDomain struct {
    AssociatedApplications []struct {
        ApplicationId string `json:"applicationId"`
    } `json:"associatedApplications"`
}
func VerifyPublisherDomainStaging(c *gin.Context) {
	var data PublisherDomain
	if C.GetConfig().Env == C.STAGING {
		data = PublisherDomain{
			AssociatedApplications: []struct {
				ApplicationId string `json:"applicationId"`
			}{
				{
					ApplicationId: "b60cd2fb-e869-4940-9aad-098d551f8217",
				},
			},
		}
	
	}
	
	// json := `{
	// 	"associatedApplications": [
	// 	  {
	// 		"applicationId": "b60cd2fb-e869-4940-9aad-098d551f8217"
	// 	  }
	// 	]
	//   }`
	c.JSON(200,data)
}
