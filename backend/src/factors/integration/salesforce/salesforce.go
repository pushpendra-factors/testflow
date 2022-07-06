package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

const (
	salesforceTokenURL = "login.salesforce.com/services/oauth2/token"
	salesforceAuthURL  = "login.salesforce.com/services/oauth2/authorize"
	// RefreshTokenURL URL for salesforce refresh token
	RefreshTokenURL = "https://login.salesforce.com/services/oauth2/token"
	// AppSettingsURL URL for factors salesforce settings page
	AppSettingsURL = "/#/settings/salesforce"
	// RefreshToken refresh_token
	RefreshToken = "refresh_token"
)

// OAuthState represent the state parameter for oAuth flow
type OAuthState struct {
	ProjectID int64   `json:"pid"`
	AgentUUID *string `json:"aid"`
}

// AuthParams common struct throughout auth
type AuthParams struct {
	GrantType    string `token_param:"grant_type"`
	AccessCode   string `token_param:"code"`
	ClientSecret string `token_param:"client_secret"`
	ClientID     string `token_param:"client_id" auth_param:"client_id" `
	RedirectURL  string `token_param:"redirect_uri" auth_param:"redirect_uri"`
	ResponseType string `auth_param:"response_type"`
	State        string `auth_param:"state"`
}

// GetSalesforceUserToken return security credentials for a user
func GetSalesforceUserToken(salesforceTokenParams *AuthParams) (map[string]interface{}, error) {
	var credentials map[string]interface{}
	urlParamsStr, err := buildQueryParamsByTagName(*salesforceTokenParams, "token_param")
	if err != nil {
		return credentials, errors.New("failed to build query parameter")
	}

	tokenURL := fmt.Sprintf("https://%s?%s", salesforceTokenURL, urlParamsStr)
	resp, err := http.Post(tokenURL, "application/json", strings.NewReader(""))
	if err != nil {
		return credentials, errors.New("failed to build request to salesforce tokenUrl")
	}

	if resp.StatusCode != http.StatusOK {
		return credentials, errors.New("fetching salesforce user credentials failed")
	}

	err = json.NewDecoder(resp.Body).Decode(&credentials)
	if err != nil {
		return credentials, errors.New("failed to decode salesforce token response")
	}
	return credentials, nil
}

// GetSalesforceAuthorizationURL return the auth URL for granting access
func GetSalesforceAuthorizationURL(clientID, redirectURL, responseType, state string) string {
	baseURL := "https://" + salesforceAuthURL
	urlParams := AuthParams{
		ClientID:     clientID,
		RedirectURL:  redirectURL,
		ResponseType: responseType,
		State:        state,
	}

	urlParamsStr, err := buildQueryParamsByTagName(urlParams, "auth_param")
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s?%s", baseURL, urlParamsStr)
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
