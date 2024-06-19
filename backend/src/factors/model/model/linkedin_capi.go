package model

import (
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

const (
	SHA256_EMAIL                           = "SHA256_EMAIL"
	LINKEDIN_FIRST_PARTY_ADS_TRACKING_UUID = "LINKEDIN_FIRST_PARTY_ADS_TRACKING_UUID"
	LINKEDIN_VERSION                       = "202403"
)

type LinkedinCAPIConfig struct {
	Conversions         BatchLinkedInCAPIConversionsResponse `json:"conversions"` // eg:- urn:lla:llaPartnerConversion:123
	IsLinkedInCAPI      bool                                 `json:"is_linkedin_capi"`
	LinkedInAccessToken string                               `json:"int_linkedin_access_token"`
	ApiKey              string                               `json:"api_key"`
	HashedEmailId       string                               `json:"hashed_email_id"`
	LinkedInClickId     string                               `json:"li_click_id"`
	LinkedInAdAccounts  []string                             `json:"int_linkedin_ad_account"`
}

type BatchLinkedinCAPIRequestPayload struct {
	LinkedinCAPIRequestPayloadList []SingleLinkedinCAPIRequestPayload `json:"elements"`
}

type SingleLinkedinCAPIRequestPayload struct {
	Conversion           string       `json:"conversion"`
	ConversionHappenedAt int64        `json:"conversionHappenedAt"`
	User                 LinkedInUser `json:"user"`
}

type LinkedInUser struct {
	UserIds []UserId `json:"userIds"`
}

type UserId struct {
	IDType  string `json:"idType"`
	IDValue string `json:"idValue"`
}

type BatchLinkedInCAPIConversionsResponse struct {
	LinkedInCAPIConversionsResponseList []SingleLinkedInCAPIConversionsResponse `json:"elements"`
}
type SingleLinkedInCAPIConversionsResponse struct {
	ConversionsId          int64  `json:"id"`
	ConversoinsDisplayName string `json:"name"`
	IsEnabled              bool   `json:"enabled"`
	AdAccount              string `json:"account"`
}

func IsLinkedInCAPICofigByWorkflow(workflowAlertBody WorkflowAlertBody) bool {

	var linkedinCAPIConfig LinkedinCAPIConfig
	err := U.DecodePostgresJsonbToStructType(workflowAlertBody.AdditonalConfigurations, &linkedinCAPIConfig)
	if err != nil {
		log.Error("Failed to decode struct")
		return false
	}

	return linkedinCAPIConfig.IsLinkedInCAPI
}
