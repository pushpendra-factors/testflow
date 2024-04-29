package model

import (
	"bytes"
	"encoding/json"
	"factors/config"
	"factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const MAILMODO_TRIGGER_CAMPAIGN_BASE_URL = "https://api.mailmodo.com/api/v1/triggerCampaign/"
const MAILMODO_GET_CONTACT_DETAILS_URL = "https://api.mailmodo.com/api/v1/getContactDetails?email="

type MailmodoTriggerCampaignRequestPayload struct {
	ReceiverEmail string                 `json:"email"`
	Data          map[string]interface{} `json:"data"`
}

type MailmodoTriggerCampaignAPIResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ReferenceId string `json:"ref"`
}

type MailmodoGetContactDetailsAPIResponse struct {
	ContactIdentifier      string   `json:"contactIdentifier"`
	Email                  string   `json:"email"`
	Blocked                bool     `json:"blocked"`
	Unsubscribed           bool     `json:"unsubscribed"`
	UnsubscribedReason     string   `json:"unsubscribedReason"`
	UnsubscribedEmailTypes []string `json:"unsubscribedEmailTypes"`
}

func FormMailmodoTriggerCampaignRequest(campaignId string, reqPayload []byte) (*http.Request, error) {

	apiKey := config.GetMailmodoTriggerCampaignAPIResponse()
	url := MAILMODO_TRIGGER_CAMPAIGN_BASE_URL + campaignId

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(reqPayload))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("mmApiKey", apiKey)

	return request, nil

}

func GetMailmodoGetContactDetailsResponse(email string) (*http.Response, error) {

	logCtx := log.WithField("email_id", email)

	apiKey := config.GetMailmodoTriggerCampaignAPIResponse()
	url := MAILMODO_GET_CONTACT_DETAILS_URL + email

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logCtx.Error("Failed to form mailmodo get contact details request.")
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("mmApiKey", apiKey)

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		logCtx.Error("Failed to get mailmodo get contact details response.")
		return nil, err
	}

	return response, nil
}

// IsReceipentAllowedMailmodo checks if the receipent has blocked the emails or have unsubscribed the email categorised by emailTypes.
func IsReceipentAllowedMailmodo(email string, emailType string) (bool, error) {

	logCtx := log.WithField("email_id", email)
	response, err := GetMailmodoGetContactDetailsResponse(email)
	if err != nil {
		logCtx.Error("Failed to get mailmodo get contact details response.")
		return false, err
	}

	defer response.Body.Close()

	// case of new email id
	if response.StatusCode != http.StatusOK {
		logCtx.Info("Passing email check")
		return true, nil
	}

	var contactDetailsResponse MailmodoGetContactDetailsAPIResponse
	err = json.NewDecoder(response.Body).Decode(&contactDetailsResponse)
	if err != nil {
		logCtx.Error("GetContactDetails API Response decode failed.")
		return false, err
	}

	if contactDetailsResponse.Blocked {
		return false, nil
	} else if contactDetailsResponse.Unsubscribed {
		if util.StringValueIn(emailType, contactDetailsResponse.UnsubscribedEmailTypes) {
			return false, nil
		}
	}

	return true, nil

}
