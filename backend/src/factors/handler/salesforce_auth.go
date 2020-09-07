package handler

import (
	"encoding/json"
	"errors"
	C "factors/config"
	M "factors/model"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type OAuthState struct {
	ProjectId uint64  `json:"pid"`
	AgentUUID *string `json:"aid"`
}

// SalesforceAuthParams common struct throughout auth
type SalesforceAuthParams struct {
	GrantType    string `token_param:"grant_type"`
	AccessCode   string `token_param:"code"`
	ClientSecret string `token_param:"client_secret"`
	ClientId     string `token_param:"client_id" auth_param:"client_id" `
	RedirectURL  string `token_param:"redirect_uri" auth_param:"redirect_uri"`
	ResponseType string `auth_param:"response_type"`
	State        string `auth_param:"state"`
}

const SALESFORCE_TOKEN_URL = "login.salesforce.com/services/oauth2/token"
const SALESFORCE_AUTH_URL = "login.salesforce.com/services/oauth2/authorize"
const SALESFORCE_CALLBACK_URL = "/salesforce/auth/callback"
const SALESFORCE_APP_SETTINGS_URL = "/#/settings/salesforce"
const SALESFORCE_REFRESH_TOKEN = "refresh_token"
const SALESFORCE_INSTANCE_URL = "instance_url"

func getSalesforceRedirectURL() string {
	return C.GetProtocol() + C.GetAPIDomain() + SALESFORCE_CALLBACK_URL
}

// SalesforceCallbackHandler handles the callback url from salesforce auth redirect url and requests access token
func SalesforceCallbackHandler(c *gin.Context) {
	var oauthState OAuthState
	accessCode := c.Query("code")
	state := c.Query("state")
	err := json.Unmarshal([]byte(state), &oauthState)
	if err != nil || oauthState.ProjectId == 0 || *oauthState.AgentUUID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{"project_id": oauthState.ProjectId, "agent_uuid": oauthState.AgentUUID})
	salesforceTokenParams := SalesforceAuthParams{
		GrantType:    "authorization_code",
		AccessCode:   accessCode,
		ClientId:     C.GetSalesforceAppId(),
		ClientSecret: C.GetSalesforceAppSecret(),
		RedirectURL:  getSalesforceRedirectURL(),
	}

	urlParamsStr, err := buildQueryParamsByTagName(salesforceTokenParams, "token_param")
	if err != nil {
		logCtx.WithError(err).Error("Failed to build query parameter")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	tokenUrl := fmt.Sprintf("https://%s?%s", SALESFORCE_TOKEN_URL, urlParamsStr)
	resp, err := http.Post(tokenUrl, "application/json", strings.NewReader(""))
	if err != nil {
		logCtx.WithError(err).Error("Failed to make request to salesforce tokenUrl")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var tokenResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode salesforce token respone")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	refreshToken, instancUrl := getRequiredSalesforceCredentials(tokenResponse)
	if refreshToken == "" || instancUrl == "" {
		logCtx.Error("Failed to getRequiredSalesforceCredentials")
		c.AbortWithStatus(http.StatusBadRequest)
	}

	errCode := M.UpdateAgentIntSalesforce(*oauthState.AgentUUID,
		refreshToken,
		instancUrl,
	)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce properties for agent.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed updating saleforce properties for agent"})
		return
	}

	_, errCode = M.UpdateProjectSettings(oauthState.ProjectId,
		&M.ProjectSetting{IntSalesforceEnabledAgentUUID: oauthState.AgentUUID},
	)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update project settings salesforce enable agent uuid.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed updating salesforce enabled agent uuid project settings"})
		return
	}

	redirectURL := C.GetProtocol() + C.GetAPPDomain() + SALESFORCE_APP_SETTINGS_URL
	c.Redirect(http.StatusPermanentRedirect, redirectURL)
}

func getRequiredSalesforceCredentials(credentials map[string]interface{}) (string, string) {
	if refreshToken, rValid := credentials[SALESFORCE_REFRESH_TOKEN].(string); rValid { //could lead to error if refresh token not set on auth scope
		if instancUrl, iValid := credentials[SALESFORCE_REFRESH_TOKEN].(string); iValid {
			if refreshToken != "" && instancUrl != "" {
				return refreshToken, instancUrl
			}

		}
	}
	return "", ""
}

// SalesforceAuthRedirect redirects to Salesforce oauth page
func SalesforceAuthRedirect(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Query("pid"), 10, 64)
	if err != nil || projectId == 0 {
		log.WithError(err).Error(
			"Failed to get project_id on get SalesforceAuthRedirect.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid project id."})
		return
	}

	agentUUID := c.Query("aid")
	if agentUUID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	oAuthState := &OAuthState{
		ProjectId: projectId,
		AgentUUID: &agentUUID,
	}

	enOAuthState, err := json.Marshal(oAuthState)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	redirectURL := getSalesforceAuthorizationUrl(C.GetSalesforceAppId(), getSalesforceRedirectURL(), "code", url.QueryEscape(string(enOAuthState)))
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func getSalesforceAuthorizationUrl(clientId, redirectUrl, responseType, state string) string {
	baseUrl := "https://" + SALESFORCE_AUTH_URL
	urlParams := SalesforceAuthParams{
		ClientId:    clientId,
		RedirectURL: redirectUrl,
		State:       state,
	}

	urlParamsStr, err := buildQueryParamsByTagName(urlParams, "auth_param")
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s?%s", baseUrl, urlParamsStr)
}

// buildQueryParamsByTagName generates url parameters by struct tags
func buildQueryParamsByTagName(params interface{}, tag string) (string, error) {
	rParams := reflect.ValueOf(params)
	if rParams.Kind() != reflect.Struct {
		return "", errors.New("params must be struct type")
	}

	var urlParams string
	paramsTyp := rParams.Type()
	for i := 0; i < rParams.NumField(); i++ {
		paramField := paramsTyp.Field(i)
		if tagName := paramField.Tag.Get(tag); tagName != "" {
			if urlParams == "" {
				urlParams = tagName + "=" + rParams.Field(i).Interface().(string)
			} else {
				urlParams = urlParams + "&" + tagName + "=" + rParams.Field(i).Interface().(string)
			}
		}

	}
	return urlParams, nil
}
