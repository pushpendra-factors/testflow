package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

const SALESFORCE_TOKEN_URL = "login.salesforce.com/services/oauth2/token"
const SALESFORCE_AUTH_URL = "login.salesforce.com/services/oauth2/authorize"
const SALESFORCE_APP_SETTINGS_URL = "/#/settings/salesforce"
const SALESFORCE_REFRESH_TOKEN = "refresh_token"
const SALESFORCE_INSTANCE_URL = "instance_url"

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

func GetSalesforceUserToken(salesforceTokenParams *SalesforceAuthParams) (map[string]interface{}, error) {
	var credentials map[string]interface{}
	urlParamsStr, err := buildQueryParamsByTagName(*salesforceTokenParams, "token_param")
	if err != nil {
		return credentials, errors.New("failed to build query parameter")
	}

	tokenUrl := fmt.Sprintf("https://%s?%s", SALESFORCE_TOKEN_URL, urlParamsStr)
	resp, err := http.Post(tokenUrl, "application/json", strings.NewReader(""))
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

func GetSalesforceAuthorizationUrl(clientId, redirectUrl, responseType, state string) string {
	baseUrl := "https://" + SALESFORCE_AUTH_URL
	urlParams := SalesforceAuthParams{
		ClientId:     clientId,
		RedirectURL:  redirectUrl,
		ResponseType: responseType,
		State:        state,
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
