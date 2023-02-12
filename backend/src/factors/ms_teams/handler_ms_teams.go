package teams

import (
	// "bytes"
	"encoding/json"
	// "errors"
	C "factors/config"
	mid "factors/middleware"
	// "factors/model/model"
	// "factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
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
	redirectURL := GetTeamsAuthorisationURL(C.GetTeamsClientID(), url.QueryEscape(string(enOAuthState)))
	c.JSON(http.StatusOK, gin.H{"redirectURL": redirectURL})

}
func GetTeamsAuthorisationURL(clientID string, state string) string {
	url := fmt.Sprintf(`https://login.microsoftonline.com/organizations/oauth2/v2.0/authorize?client_id=%s&scope=group.ReadWriteAll,user.ReadAll&state=%s`, clientID, state)
	return url
}
func TeamsCallbackHandler(c *gin.Context) {
	// Extract the code from the query string
	code := c.Query()
	state := c.Query()

	// Validate the state parameter to ensure it matches what was originally sent
	// TODO: Implement state validation logic here

	// Use the code to get an access token from Microsoft's authorization server
	// TODO: Implement the logic to get the access token using the code here

	// Use the access token to call the Microsoft Graph API
	// TODO: Implement the logic to call the Microsoft Graph API here

	log.Info("Redirect handler complete! ")
}
