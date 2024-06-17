package linkedin_capi

import (
	"encoding/json"
	"errors"
	"factors/cache"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const EventsConversionApiURL = "https://api.linkedin.com/rest/conversionEvents"
const GetConversionEventApiURL = "https://api.linkedin.com/rest/conversions"

type LinkedInCapiInfo struct {
	EventsConversionApiURL   string
	GetConversionEventApiURL string
}

func GetLinkedInCapi() LinkedInCapi {

	linkedInCapi := LinkedInCapiInfo{}
	linkedInCapi.EventsConversionApiURL = EventsConversionApiURL
	linkedInCapi.GetConversionEventApiURL = GetConversionEventApiURL
	return &linkedInCapi
}

type LinkedInCapi interface {
	SendEventsToLinkedCAPI(config model.LinkedinCAPIConfig, body model.BatchLinkedinCAPIRequestPayload) (map[string]interface{}, error)
	GetConversionFromLinkedCAPI(config model.LinkedinCAPIConfig) (model.BatchLinkedInCAPIConversionsResponse, error)
	SendHelper(key *cache.Key, cachedWorkflow *model.CachedEventTriggerAlert, workflowID string, retry bool, sendTo string, logCtx log.Entry) (map[string]interface{}, error)
}

type LinkedInCapiMock struct {
	SendEventsToLinkedCAPICalls []struct {
		Config model.LinkedinCAPIConfig
		Body   model.BatchLinkedinCAPIRequestPayload
	}
	SendEventsToLinkedCAPIData  map[string]interface{}
	SendEventsToLinkedCAPIError error

	GetConversionFromLinkedCAPICalls []struct {
		Config model.LinkedinCAPIConfig
	}
	GetConversionFromLinkedCAPIData  model.BatchLinkedInCAPIConversionsResponse
	GetConversionFromLinkedCAPIError error
}

func (m *LinkedInCapiMock) SendEventsToLinkedCAPI(config model.LinkedinCAPIConfig, body model.BatchLinkedinCAPIRequestPayload) (map[string]interface{}, error) {
	m.SendEventsToLinkedCAPICalls = append(m.SendEventsToLinkedCAPICalls, struct {
		Config model.LinkedinCAPIConfig
		Body   model.BatchLinkedinCAPIRequestPayload
	}{config, body})
	//do all checks on config and body

	return m.SendEventsToLinkedCAPIData, m.SendEventsToLinkedCAPIError
}

func (m *LinkedInCapiMock) GetConversionFromLinkedCAPI(config model.LinkedinCAPIConfig) (model.BatchLinkedInCAPIConversionsResponse, error) {
	m.GetConversionFromLinkedCAPICalls = append(m.GetConversionFromLinkedCAPICalls, struct {
		Config model.LinkedinCAPIConfig
	}{config})
	//do all checks on config

	return m.GetConversionFromLinkedCAPIData, m.GetConversionFromLinkedCAPIError
}

func (li *LinkedInCapiInfo) SendEventsToLinkedCAPI(config model.LinkedinCAPIConfig, body model.BatchLinkedinCAPIRequestPayload) (map[string]interface{}, error) {

	logCtx := log.WithFields(
		log.Fields{"config": config,
			"body": body},
	)

	payloadJson, err := U.EncodeStructTypeToPostgresJsonb(body)
	if err != nil {
		log.WithError(err).Error("failed to encode payload for linkedin capi.")
		return nil, errors.New("failed to encode payload for linkedin capi")
	}
	rb := U.NewRequestBuilder(http.MethodPost, li.EventsConversionApiURL).
		WithPostParams(payloadJson).
		WithHeader("Content-Type", "application/json").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", config.LinkedInAccessToken)).
		WithHeader("X-Restli-Protocol-Version", "2.0.0").
		WithHeader("LinkedIn-Version", model.LINKEDIN_VERSION).
		WithHeader("X-RestLi-Method", "BATCH_CREATE")

	request, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("error sending events to linkedin capi.")
		return nil, errors.New("error sending events to linkedin capi.")
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Error("Failed to send events linkedin")
		return nil, errors.New("failed to send events linkedin")
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {

		_, err = handleErrorForBatchCreateResponse(resp, logCtx)
		if err != nil {
			return nil, err
		}

		return nil, errors.New("failed to execute POST request in linkedin")
	}

	//type required for workflow sender
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		log.Error("failed to decode json response", err)
		return nil, errors.New("failed to decode json response")
	}

	jsonResponse["status"] = "success"

	logCtx.WithField("jsonResponse", jsonResponse).Info("Linkedin CAPI TEST - 25")

	return jsonResponse, nil
}

func (li *LinkedInCapiInfo) SendHelper(key *cache.Key, cachedWorkflow *model.CachedEventTriggerAlert,
	workflowID string, retry bool, sendTo string, logCtx log.Entry) (map[string]interface{}, error) {

	config, err := store.GetStore().GetLinkedInCAPICofigByWorkflowId(key.ProjectID, workflowID)
	if err != nil {
		logCtx.WithError(err).Error("failed  to get linkedin configuration")
	}

	var linkedCAPIPayloadBatch model.BatchLinkedinCAPIRequestPayload
	linkedinCAPIPayloadString := U.GetPropertyValueAsString(cachedWorkflow.Message.MessageProperty["linkedCAPI_payload"])

	err = U.DecodeJSONStringToStructType(linkedinCAPIPayloadString, &linkedCAPIPayloadBatch)
	if err != nil {
		logCtx.WithError(err).Error("failed to decode linkedin capi payload")
	}

	response, err := GetLinkedInCapi().SendEventsToLinkedCAPI(config, linkedCAPIPayloadBatch)
	if err != nil {
		logCtx.WithFields(log.Fields{"server_response": response}).WithError(err).Error("LinkedIn CAPI Workflow failure.")
	}
	logCtx.WithField("cached_workflow", cachedWorkflow).WithField("response", response).Info("LinkedIn CAPI workflow sent.")

	return response, nil
}

func (li *LinkedInCapiInfo) GetConversionFromLinkedCAPI(config model.LinkedinCAPIConfig) (model.BatchLinkedInCAPIConversionsResponse, error) {
	var finalJsonResponse model.BatchLinkedInCAPIConversionsResponse

	_rb := *U.NewRequestBuilder(http.MethodGet, li.GetConversionEventApiURL).
		WithHeader("Content-Type", "application/json").
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", config.LinkedInAccessToken)).
		WithHeader("LinkedIn-Version", model.LINKEDIN_VERSION).
		WithHeader("X-Restli-Protocol-Version", "2.0.0")

	for _, adAccount := range config.LinkedInAdAccounts {

		logCtx := log.WithFields(
			log.Fields{"adAccount": adAccount},
		)

		isEndReached := false
		start, count := 0, 1000
		for !isEndReached {

			rb := _rb
			rb.WithQueryParams(map[string]string{
				"q":       "account",
				"account": "urn:li:sponsoredAccount:" + adAccount,
				"start":   U.GetPropertyValueAsString(start),
				"count":   U.GetPropertyValueAsString(count),
			})

			request, err := rb.Build()
			if err != nil {
				log.WithError(err).Error("failed to get linkedin capi conversion list.")
			}

			client := &http.Client{}
			resp, err := client.Do(request)
			if err != nil {
				log.Error("Failed to get linkedin conversions list")
				continue

			}

			if resp.StatusCode != http.StatusOK {
				_, err = handleErrorForBatchCreateResponse(resp, logCtx)

				log.WithError(err).Error("failed to get list from linkedin capi")
				break
			}
			var jsonResponse model.BatchLinkedInCAPIConversionsResponse
			err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
			if err != nil {
				log.Error("failed to decode json response", err)
				continue
			}

			if len(jsonResponse.LinkedInCAPIConversionsResponseList) == 0 {
				isEndReached = true
				log.WithField("count", count).WithField("adAccount", adAccount).Info("count for given ad account linkedin capi")
				break
			}

			for _, singleResponse := range jsonResponse.LinkedInCAPIConversionsResponseList {
				if singleResponse.IsEnabled {
					finalJsonResponse.LinkedInCAPIConversionsResponseList = append(finalJsonResponse.LinkedInCAPIConversionsResponseList, singleResponse)
				}
			}

			start += count

		}

	}

	if len(finalJsonResponse.LinkedInCAPIConversionsResponseList) == 0 {
		return finalJsonResponse, errors.New("no conversions found for ad accounts ")
	}

	return finalJsonResponse, nil
}

type BatchError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func handleErrorForBatchCreateResponse(resp *http.Response, logCtx *log.Entry) (io.ReadCloser, error) {

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read response body")
	}

	var errorResponse BatchError

	logCtx.WithField("body", body).Info("Linkedin capi batch_create type error response")
	err = json.Unmarshal(body, &errorResponse)
	if err != nil {
		return nil, errors.New("failed to decode error response")
	}

	if errorResponse.Status >= http.StatusBadRequest && errorResponse.Status <= http.StatusUnprocessableEntity {

		logCtx.WithField("errorBatch", errorResponse.Message).Error("batch creation failed: conflict errors encountered")
		return nil, errors.New("batch creation failed: conflict errors encountered")
	}

	// Handle other error codes or generic error message

	return nil, errors.New("batch creation failed: " + errorResponse.Message)

}
