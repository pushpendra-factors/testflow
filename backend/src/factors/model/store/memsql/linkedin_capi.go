package memsql

import (
	"errors"
	"factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) NewLinkedCapiRequestPayload(properties *map[string]interface{}, linkedinCAPIConfig model.LinkedinCAPIConfig) (model.BatchLinkedinCAPIRequestPayload, error) {

	linkedinCAPIRequestPayloadBatch := make([]model.SingleLinkedinCAPIRequestPayload, 0)
	_linkedinCAPIRequestPayload := model.SingleLinkedinCAPIRequestPayload{}

	if properties == nil {
		log.Error("Failed to hash email")
		return model.BatchLinkedinCAPIRequestPayload{}, errors.New("failed to hash email")
	}

	if emailId, exists := (*properties)[U.EP_EMAIL]; exists {

		hashedEmail, err := U.GetSHA256Hash(U.GetPropertyValueAsString(emailId))
		if err != nil {
			log.WithError(err).Error("Failed to hash email")
			return model.BatchLinkedinCAPIRequestPayload{}, errors.New("failed to hash email")
		}
		_linkedinCAPIRequestPayload.User.UserIds = append(_linkedinCAPIRequestPayload.User.UserIds, model.UserId{IDType: model.SHA256_EMAIL, IDValue: hashedEmail})
	}

	if emailId, exists := (*properties)[U.UP_EMAIL]; exists {

		hashedEmail, err := U.GetSHA256Hash(U.GetPropertyValueAsString(emailId))
		if err != nil {
			log.WithError(err).Error("Failed to hash email")
			return model.BatchLinkedinCAPIRequestPayload{}, errors.New("failed to hash email")
		}
		_linkedinCAPIRequestPayload.User.UserIds = append(_linkedinCAPIRequestPayload.User.UserIds, model.UserId{IDType: model.SHA256_EMAIL, IDValue: hashedEmail})

	}

	if len(_linkedinCAPIRequestPayload.User.UserIds) == 0 {
		log.Error("no user identifier found for linked capi")
		return model.BatchLinkedinCAPIRequestPayload{}, errors.New("no user identifier found for linked capi")
	}

	if timestamp, exists := (*properties)[U.EP_TIMESTAMP]; exists {

		intTimestamp, err := U.GetPropertyValueAsInt64(timestamp)
		if err != nil {
			log.WithError(err).Error("Unable to get timestamp")
			return model.BatchLinkedinCAPIRequestPayload{}, errors.New("Unable to get timestamp")
		}

		if intTimestamp-time.Now().Unix() > 90*U.SECONDS_IN_A_DAY {
			log.WithError(err).Error("timestamp older than 90 days")
			return model.BatchLinkedinCAPIRequestPayload{}, errors.New("timestamp older than 90 days")
		}
		_linkedinCAPIRequestPayload.ConversionHappenedAt = intTimestamp * int64(1000)

	}

	if liclid, exists := (*properties)[U.EP_LICLID]; exists {

		_linkedinCAPIRequestPayload.User.UserIds = append(_linkedinCAPIRequestPayload.User.UserIds, model.UserId{IDType: model.LINKEDIN_FIRST_PARTY_ADS_TRACKING_UUID, IDValue: U.GetPropertyValueAsString(liclid)})

	}

	if len(linkedinCAPIConfig.Conversions.LinkedInCAPIConversionsResponseList) == 0 {
		log.Error("no conversions found for linked capi")
		return model.BatchLinkedinCAPIRequestPayload{}, errors.New("no conversions found for linked capi")
	}

	for _, conversion := range linkedinCAPIConfig.Conversions.LinkedInCAPIConversionsResponseList {

		linkedinCAPIRequestPayload := _linkedinCAPIRequestPayload
		linkedinCAPIRequestPayload.Conversion = fmt.Sprintf("urn:lla:llaPartnerConversion:%d", conversion.ConversionsId)
		linkedinCAPIRequestPayloadBatch = append(linkedinCAPIRequestPayloadBatch, linkedinCAPIRequestPayload)

	}

	if len(linkedinCAPIRequestPayloadBatch) == 0 {
		return model.BatchLinkedinCAPIRequestPayload{}, errors.New("no batch found for linkedin ad accounts")
	}
	return model.BatchLinkedinCAPIRequestPayload{LinkedinCAPIRequestPayloadList: linkedinCAPIRequestPayloadBatch}, nil
}

func (store *MemSQL) GetLinkedInCAPICofigByWorkflowId(projectID int64, workflowID string) (model.LinkedinCAPIConfig, error) {

	logCtx := log.WithFields(log.Fields{
		"projectID":  projectID,
		"workflowID": workflowID,
	})
	wf, _, err := store.GetWorkflowById(projectID, workflowID)
	if err != nil {
		logCtx.Error("Failed to get workflow")
		return model.LinkedinCAPIConfig{}, errors.New("Failed to get workflow")
	}

	var workflowAlertBody model.WorkflowAlertBody
	err = U.DecodePostgresJsonbToStructType(wf.AlertBody, &workflowAlertBody)
	if err != nil {
		logCtx.Error("Failed to decode struct")
		return model.LinkedinCAPIConfig{}, errors.New("Failed to decode struct")
	}

	var linkedinCAPIConfig model.LinkedinCAPIConfig
	err = U.DecodePostgresJsonbToStructType(workflowAlertBody.AdditonalConfigurations, &linkedinCAPIConfig)
	if err != nil {
		logCtx.Error("Failed to decode struct")
		return model.LinkedinCAPIConfig{}, errors.New("Failed to decode struct")
	}

	return linkedinCAPIConfig, nil
}

func (store *MemSQL) FillConfigurationValuesForLinkedinCAPIWorkFlow(projectId int64, workflowAlertBody *model.WorkflowAlertBody) error {

	logCtx := log.WithFields(log.Fields{
		"projectID": projectId,
	})
	linkedInWorkflowConfig := model.LinkedinCAPIConfig{}
	linkedInCAPIConversionsResponseList := []model.SingleLinkedInCAPIConversionsResponse{}
	settings, errCode := store.GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return errors.New("project settings not found")
	}

	if settings.IntLinkedinAccessToken == "" || settings.IntLinkedinAdAccount == "" {
		logCtx.Error("unable to fetch linkedin account info ")
		return errors.New("unable to fetch linkedin account info ")
	}

	err := U.DecodePostgresJsonbToStructType(workflowAlertBody.AdditonalConfigurations, &linkedInCAPIConversionsResponseList)
	if err != nil {
		logCtx.Error(err)
		return err
	}

	linkedInWorkflowConfig.Conversions = model.BatchLinkedInCAPIConversionsResponse{LinkedInCAPIConversionsResponseList: linkedInCAPIConversionsResponseList}
	linkedInWorkflowConfig.LinkedInAccessToken = settings.IntLinkedinAccessToken
	linkedInWorkflowConfig.LinkedInAdAccounts = config.GetTokensFromStringListAsString(settings.IntLinkedinAdAccount)

	if len(linkedInWorkflowConfig.Conversions.LinkedInCAPIConversionsResponseList) == 0 {
		logCtx.Error("No conversions for linkedin capi")
		return errors.New("No conversions for linkedin capi")
	}

	linkedInWorkflowConfig.IsLinkedInCAPI = true

	additonalConfigurations, err := U.EncodeStructTypeToPostgresJsonb(linkedInWorkflowConfig)
	if err != nil {
		logCtx.Error(err)
		return err
	}

	workflowAlertBody.AdditonalConfigurations = additonalConfigurations

	return nil

}
