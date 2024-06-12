package linkedin_capi

import (
	"bytes"
	"encoding/json"
	"errors"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func SendEventsToLinkedCAPI(config model.LinkedinCAPIConfig, body model.BatchLinkedinCAPIRequestPayload) (map[string]interface{}, error) {

	logCtx := log.WithFields(
		log.Fields{"config": config,
			"body": body},
	)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Error("Failed to marshal request body for linkedinCAPI")
		return nil, err
	}

	request, err := http.NewRequest("POST", "https://api.linkedin.com/rest/conversions", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Error("Failed to create request to get slack channels list")
		return nil, errors.New("failed to create request to send events linkedin")
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.LinkedInAccessToken))
	request.Header.Set("LinkedIn-Version", model.LINKEDIN_VERSION)
	request.Header.Set("X-Restli-Protocol-Version", "2.0.0")
	request.Header.Set("X-RestLi-Method", "BATCH_CREATE")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Error("Failed to send events linkedin")
		return nil, errors.New("failed to send events linkedin")
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {

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

	jsonResponse["stat"] = "success"

	return jsonResponse, nil
}

func GetConversionFromLinkedCAPI(config model.LinkedinCAPIConfig) (model.BatchLinkedInCAPIConversionsResponse, error) {
	var finalJsonResponse model.BatchLinkedInCAPIConversionsResponse
	request, err := http.NewRequest("GET", "https://api.linkedin.com/rest/conversions", nil)
	if err != nil {
		log.Error("Failed to create request to get slack channels list")
		return model.BatchLinkedInCAPIConversionsResponse{}, errors.New("failed to create request to get linkedin conversions list")
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.LinkedInAccessToken))
	request.Header.Set("LinkedIn-Version", model.LINKEDIN_VERSION)
	request.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	for _, adAccount := range config.LinkedInAdAccounts {

		logCtx := log.WithFields(
			log.Fields{"adAccount": adAccount},
		)

		isEndReached := false
		start, count := 0, 1000
		for !isEndReached {

			q := request.URL.Query()
			q.Add("q", "account")
			q.Add("account", "urn%3Ali%3AsponsoredAccount%3A"+adAccount)
			q.Add("start", U.GetPropertyValueAsString(start))
			q.Add("count", U.GetPropertyValueAsString(count))

			request.URL.RawQuery = q.Encode()

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

			finalJsonResponse.LinkedInCAPIConversionsResponseList = append(finalJsonResponse.LinkedInCAPIConversionsResponseList, jsonResponse.LinkedInCAPIConversionsResponseList...)

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
