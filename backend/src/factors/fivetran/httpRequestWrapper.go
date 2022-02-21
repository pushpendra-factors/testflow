package fivetran

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func HttpRequestWrapper(endpoint string, headers map[string]string, requestBody interface{}, requestMethod string) (int, map[string]interface{}, Error) {
	RootUrl := "https://api.fivetran.com/v1/"
	var reqBody []byte
	var err error
	var errResp Error
	if requestBody != nil {
		reqBody, err = json.Marshal(requestBody)
		if err != nil {
			log.WithError(err).Error("Failed to marshall request object")
			return 0, nil, errResp
		}
	}
	url := fmt.Sprintf("%s%s", RootUrl, endpoint)
	request, err := http.NewRequest(requestMethod, url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.WithError(err).Error("Failed to create request object")
		return 0, nil, errResp
	}
	for headerKey, headerValue := range headers {
		request.Header.Add(headerKey, headerValue)
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("Failed to execute request")
	}
	defer response.Body.Close()
	var respBody map[string]interface{}
	decoder := json.NewDecoder(response.Body)
	if response.StatusCode == 201 || response.StatusCode == 200 {
		if err := decoder.Decode(&respBody); err != nil {
			log.WithError(err).Error("Failed to decode response body")
		}
	} else {
		if err := decoder.Decode(&errResp); err != nil {
			log.WithError(err).Error("Failed to decode error response body")
		}
	}
	return response.StatusCode, respBody, errResp
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
