package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"factors/model/model"
	"factors/model/store"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type oauthState struct {
	ProjectID int64   `json:"project_id"`
	AgentUUID *string `json:"agent_uuid"`
	Source    int     `json:"source"`
}


func GetSlackChannels(accessTokens model.SlackAccessTokens, nextCursor string) (response map[string]interface{}, status int, err error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("https://slack.com/api/conversations.list"), nil)
	if err != nil {
		log.Error("Failed to create request to get slack channels list")
		return nil, http.StatusInternalServerError, errors.New("Failed to create request to get slack channels list")
	}
	q := request.URL.Query()
	q.Add("types", "public_channel,private_channel,mpim")
	q.Add("limit", "2000")
	if nextCursor != "" {
		q.Add("cursor", nextCursor)
	}
	request.URL.RawQuery = q.Encode()
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessTokens.UserAccessToken))
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Error("Failed to get slack channels list")
		return nil, http.StatusInternalServerError, errors.New("Failed to get slack channels list")
	}
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		log.Error("failed to decode json response", err)
		return jsonResponse, http.StatusInternalServerError, errors.New("failed to decode json response.")
	}
	return jsonResponse, http.StatusOK, nil
}

func SendSlackAlert(projectID int64, message, agentUUID string, channel model.SlackChannel) (bool, error) {
	// get the auth token for the agent uuid and then call the POST method to send the message
	accessTokens, err := store.GetStore().GetSlackAuthToken(projectID, agentUUID)
	if err != nil {
		log.Error("Failed to get access token for slack")
		return false, err
	}
	url := fmt.Sprintf("https://slack.com/api/chat.postMessage")
	// create new http post request
	reqBody := map[string]interface{}{
		"channel":      channel.Id,
		"blocks":       message,
		"unfurl_links": false,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Error("Failed to marshal request body for slack")
		return false, err
	}
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	if channel.IsPrivate {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessTokens.UserAccessToken))
	} else {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessTokens.BotAccessToken))
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Error("Failed to make request to slack for sending alert")
		return false, err
	}
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error("Failed to decode response from slack")
		return false, err
	}
	if response["ok"] == true {
		return true, nil
	}
	log.Error("failed to send slack alert ", message, response)
	defer resp.Body.Close()
	return false, fmt.Errorf("failure response: %v", response)
}
