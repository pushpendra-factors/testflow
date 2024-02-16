package model

import (
	"bytes"
	"factors/config"
	"net/http"
)

const MAILMODO_TRIGGER_CAMPAIGN_BASE_URL = "https://api.mailmodo.com/api/v1/triggerCampaign/"

type MailmodoTriggerCampaignRequestPayload struct {
	ReceiverEmail string                 `json:"email"`
	Data          map[string]interface{} `json:"data"`
}

type MailmodoTriggerCampaignAPIResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ReferenceId string `json:"ref"`
}

func FormMailmodoTriggerCampaignRequest(campaignId string, reqPayload []byte) (*http.Request, error) {

	//TODO: @Roshan add the flag to fetch the apin key in the required workloads.
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
