package leadsquared

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func HttpRequestWrapper(rUrl string, endpoint string, headers map[string]string, requestBody interface{}, requestMethod string, urlParams map[string]string) (int, interface{}, error, bool) {
	rootUrl := rUrl
	if !strings.HasPrefix(rUrl, "http") {
		rootUrl = "https://" + rUrl
	}

	logCtx := log.WithField("host", rUrl).WithField("endpoint", endpoint)

	var reqBody []byte
	var err error
	var errResp error
	if requestBody != nil {
		reqBody, err = json.Marshal(requestBody)
		if err != nil {
			logCtx.WithError(err).Error("Failed to marshall request object")
			return 0, nil, errResp, false
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
		logCtx.WithError(err).Error("Failed to create request object")
		return 0, nil, errResp, false
	}
	for headerKey, headerValue := range headers {
		request.Header.Add(headerKey, headerValue)
	}
	client := &http.Client{}
	var respBody interface{}
	response, err := client.Do(request)
	if err != nil {
		logCtx.WithError(err).Error("Failed to execute request")
		return http.StatusInternalServerError, respBody, err, false
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		logCtx.WithError(err).Error("Failed to read response as bytes.")
	}
	logCtx = logCtx.WithField("response_body", string(responseBytes)).
		WithField("response_status", response.Status)

	if response.StatusCode == http.StatusOK {
		if err := json.Unmarshal(responseBytes, &respBody); err != nil {
			logCtx.WithError(err).Error("Failed to decode response body.")
			return http.StatusInternalServerError, respBody, err, false
		}
	} else {
		errResp = errors.New(string(responseBytes))
		logCtx.Error("Received error response on API request.")
		return http.StatusInternalServerError, respBody, errResp, isContinuableError(errResp)
	}

	return response.StatusCode, respBody, errResp, isContinuableError(errResp)
}

func isContinuableError(err error) bool {
	return strings.Contains(err.Error(), "does not exist")
}
