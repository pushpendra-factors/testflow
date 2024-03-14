package leadsquared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func HttpRequestWrapper(rUrl string, endpoint string, headers map[string]string, requestBody interface{}, requestMethod string, urlParams map[string]string) (int, interface{}, error) {
	rootUrl := rUrl
	if !strings.HasPrefix(rUrl, "http") {
		rootUrl = "https://" + rUrl
	}

	var reqBody []byte
	var err error
	var errResp error
	if requestBody != nil {
		reqBody, err = json.Marshal(requestBody)
		if err != nil {
			log.WithError(err).Error("Failed to marshall request object")
			return 0, nil, errResp
		}
	}
	urlParamString := ""
	for key, value := range urlParams {
		if urlParamString != "" {
			urlParamString = urlParamString + "&"
		}
		urlParamString = urlParamString + fmt.Sprintf("%s=%s", key, value)
	}
	url := fmt.Sprintf("%s%s?%s", rootUrl, endpoint, urlParamString)
	request, err := http.NewRequest(requestMethod, url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.WithError(err).Error("Failed to create request object")
		return 0, nil, errResp
	}
	for headerKey, headerValue := range headers {
		request.Header.Add(headerKey, headerValue)
	}
	client := &http.Client{}
	var respBody interface{}
	response, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("Failed to execute request")
		return http.StatusInternalServerError, respBody, err
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.WithError(err).Error("Failed to read response as bytes.")
	}
	logCtx := log.WithField("response_body", string(responseBytes)).
		WithField("response_status", response.Status)

	if response.StatusCode == http.StatusOK {
		if err := json.Unmarshal(responseBytes, &respBody); err != nil {
			logCtx.WithError(err).Error("Failed to decode response body.")
			return http.StatusInternalServerError, respBody, err
		}
	} else {
		if err := json.Unmarshal(responseBytes, &respBody); err != nil {
			logCtx.WithError(err).Error("Failed to decode error response body.")
			return http.StatusInternalServerError, respBody, err
		}
		logCtx.Warn("Received error response on http request.")
	}
	return response.StatusCode, respBody, errResp
}
