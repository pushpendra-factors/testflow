package hubspot

import (
	"encoding/json"
	"errors"
	"factors/model/model"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

const HubspotOAuthURL = "app.hubspot.com/oauth/authorize"
const HubspotTokenURL = "api.hubapi.com/oauth/v1/token"

var HubspotAuthAppRequiredScopes = []string{
	"crm.objects.deals.read",
	"crm.schemas.deals.read",
	"crm.objects.contacts.read",
	"crm.schemas.contacts.read",
	"crm.objects.companies.read",
	"crm.schemas.companies.read",
	"tickets",
	"sales-email-read",
	"forms",
	"content",
	"business-intelligence",
	"crm.lists.read",
	"crm.objects.owners.read",
}

// GetHubspotAuthorizationURL return the auth URL for granting access
func GetHubspotAuthorizationURL(clientID, redirectURL, state string) string {
	baseURL := "https://" + HubspotOAuthURL
	scopeStr := strings.Join(HubspotAuthAppRequiredScopes, " ")

	urlParams := fmt.Sprintf("client_id=%s&redirect_uri=%s&scope=%s&state=%s", clientID, redirectURL, url.QueryEscape(scopeStr), url.QueryEscape(state))

	return baseURL + "?" + urlParams
}

func GetHubspotOAuthUserCredentials(clientID, clientSecret, redirectURL, code string) (*model.HubspotOAuthUserCredentials, error) {
	baseURL := "https://" + HubspotTokenURL
	urlParams := fmt.Sprintf("grant_type=%s&client_id=%s&client_secret=%s&redirect_uri=%s&code=%s", "authorization_code", clientID, clientSecret, redirectURL, code)
	tokenURL := baseURL + "?" + urlParams

	resp, err := model.ActionHubspotRequestHandler("POST", tokenURL, "", "", "application/x-www-form-urlencoded;charset=utf-8", nil)
	if err != nil {
		log.WithError(err).Error("Failed to generate POST request for GetHubspotOAuthUserCredentials.")
		return nil, errors.New("failed to make POST request on GetHubspotRefreshToken")
	}

	if resp.StatusCode != http.StatusOK {
		var body interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		log.WithFields(log.Fields{"respone_body": body}).Error("Failed to get hubspot user credentials.")
		return nil, errors.New("fetching hubspot user credentials failed")
	}

	var credentials model.HubspotOAuthUserCredentials
	err = json.NewDecoder(resp.Body).Decode(&credentials)
	if err != nil {
		log.WithError(err).Error("Failed to decode hubspout auth user credentials.")
		return nil, errors.New("failed to decode hubspot auth token response")
	}

	if credentials.RefreshToken == "" {
		return nil, errors.New("empty refresh token on hubspot auth")
	}

	return &credentials, nil
}
