package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"factors/cache"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	cacheRedis "factors/cache/redis"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

type oauthState struct {
	ProjectID int64   `json:"project_id"`
	AgentUUID *string `json:"agent_uuid"`
	Source    int     `json:"source"`
}

func GetCacheKeyForSlackIntegration(projectID int64, agentUUID string) (*cache.Key, error) {
	prefix := "slack:auth:"
	suffix := fmt.Sprintf("agent_uid:%v", agentUUID)
	return cache.NewKey(projectID, prefix, suffix)
}

func SetCacheForSlackAuthRandomState(projectID int64, agentUUID, randAuthState string) {

	slackAuthKey, err := GetCacheKeyForSlackIntegration(projectID, agentUUID)
	if err != nil {
		return
	}
	expiry := float64(U.SLACK_AUTH_RANDOM_STATE_EXPIRY_SECS)
	err = cacheRedis.SetPersistent(slackAuthKey, randAuthState, expiry)
	if err != nil {
		log.WithError(err).Error("Failed to set randState in cache")
		return
	}
}

func GetCacheSlackAuthRandomState(projectId int64, agentUUID string) (string, int) {

	cacheKey, err := GetCacheKeyForSlackIntegration(projectId, agentUUID)
	if err != nil {
		return "", http.StatusNotFound
	}
	cacheResult, err := cacheRedis.GetPersistent(cacheKey)
	if err == redis.ErrNil {
		return cacheResult, http.StatusNotFound
	} else if err != nil {
		log.WithError(err).Error("Error getting key from redis")
		return cacheResult, http.StatusInternalServerError
	}
	return cacheResult, http.StatusFound
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

func SendSlackAlert(projectID int64, message, agentUUID string, channel model.SlackChannel) (map[string]interface{}, bool, error) {
	// get the auth token for the agent uuid and then call the POST method to send the message
	accessTokens, err := store.GetStore().GetSlackAuthToken(projectID, agentUUID)
	if err != nil {
		log.Error("Failed to get access token for slack")
		return nil, false, err
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
		return nil, false, err
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
		return nil, false, err
	}
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error("Failed to decode response from slack")
		return nil, false, err
	}
	if response["ok"] == true {
		return response, true, nil
	}
	log.WithFields(log.Fields{
		"message":  message,
		"response": response,
	}).Error("failed to send slack alert")
	defer resp.Body.Close()
	return response, false, fmt.Errorf("failed to send slack alert")
}

func GetSlackUsersList(projectID int64, agentID string) ([]model.SlackMember, int, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "agent_uuid": agentID})
	if projectID == 0 || agentID == "" {
		logCtx.Error("invalid parameters")
		return nil, http.StatusBadRequest, fmt.Errorf("invalid parameters")
	}

	jsonResponse, status, err := getSlackUsers(projectID, agentID, "")
	if err != nil {
		return nil, status, err
	}

	if !jsonResponse.Ok {
		err := jsonResponse.Error
		if err == "missing_scope" {
			logCtx.WithError(fmt.Errorf("%v", err)).Error("permission not granted for reading users, please reintegrate")
			return nil, http.StatusExpectationFailed, fmt.Errorf("permission not granted for reading users, please reintegrate")
		}
		logCtx.WithError(fmt.Errorf("%v", err)).Error("error received from slack server")
		return nil, http.StatusInternalServerError, fmt.Errorf("error received from slack server")
	}

	members := jsonResponse.Members
	if members == nil {
		logCtx.Error("nil json response found for users error")
		return nil, http.StatusInternalServerError, fmt.Errorf("nil json response found for users error")
	}

	jsonMetadata := jsonResponse.ResponseMetadata
	if jsonMetadata == nil {
		logCtx.Error("no metadata from json response error")
		return nil, http.StatusInternalServerError, fmt.Errorf("no metadata from json response error")
	}

	nextCursor := jsonMetadata["next_cursor"].(string)
	for nextCursor != "" {
		jsonResponse, status, err = getSlackUsers(projectID, agentID, nextCursor)
		if err != nil {
			return nil, status, err
		}

		if newMembers := jsonResponse.Members; newMembers != nil {
			members = append(members, newMembers...)
			if metadata := jsonResponse.ResponseMetadata; metadata != nil {
				nextCursor = metadata["next_cursor"].(string)
			} else {
				break
			}
		} else {
			break
		}
	}

	return members, http.StatusFound, nil
}

func getSlackUsers(projectID int64, agentID string, nextCursor string) (response *model.SlackGetUsersResponse, status int, err error) {

	// get the auth token from the agent_uuid and project_id map
	accessTokens, err := store.GetStore().GetSlackAuthToken(projectID, agentID)
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID, "agent_id": agentID}).Error("failed to get slack auth token")
		return nil, http.StatusBadRequest, err
	}

	request, err := http.NewRequest("GET", "https://slack.com/api/users.list", nil)
	if err != nil {
		log.Error("failed at request creation for users list")
		return nil, http.StatusInternalServerError, fmt.Errorf("failed at request creation for users list")
	}

	q := request.URL.Query()
	q.Add("limit", "200")
	if nextCursor != "" {
		q.Add("cursor", nextCursor)
	}

	request.URL.RawQuery = q.Encode()
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessTokens.UserAccessToken))
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("failed to get slack users list")
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to get slack users list")
	}
	body := resp.Body
	defer body.Close()

	var jsonResponse model.SlackGetUsersResponse
	err = json.NewDecoder(body).Decode(&jsonResponse)
	if err != nil {
		log.WithError(err).Error("failed to decode json response")
		return &jsonResponse, http.StatusInternalServerError, fmt.Errorf("failed to decode json response")
	}

	return &jsonResponse, http.StatusOK, nil
}

func UpdateSlackUsersListTable(projectID int64, agentID string) ([]model.SlackMember, int, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "agent_uuid": agentID})
	if projectID == 0 || agentID == "" {
		logCtx.Error("invalid parameters")
		return nil, http.StatusBadRequest, fmt.Errorf("invalid parameters")
	}

	members, errCode, err := GetSlackUsersList(projectID, agentID)
	if err != nil || errCode != http.StatusFound {
		if errCode == http.StatusNotFound {
			return nil, errCode, err
		}
		logCtx.WithError(err).Error("failed to fetch slack users list")
		return nil, errCode, err
	}

	memberJson, err := U.EncodeStructTypeToPostgresJsonb(members)
	if err != nil {
		logCtx.WithError(err).Error("failed to encode slack users list")
		return nil, http.StatusInternalServerError, err
	}

	fields := make(map[string]interface{})
	fields["agent_id"] = agentID
	fields["users_list"] = memberJson

	errCode, err = store.GetStore().UpdateSlackUsersListForProject(projectID, fields)
	if err != nil || errCode != http.StatusOK {
		return nil, errCode, err
	}

	return members, http.StatusOK, nil
}

func GetSlackIntegrationState(projectID int64, agentUUID string) model.IntegrationState {
	// get the auth token for the agent uuid and then call the POST method to send the message
	accessTokens, err := store.GetStore().GetSlackAuthToken(projectID, agentUUID)
	if err != nil {
		log.Error("Failed to get access token for slack")
		return model.IntegrationState{}
	}
	url := fmt.Sprintf("https://slack.com/api/auth.test")

	// create new http post request
	request, _ := http.NewRequest("POST", url, nil)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessTokens.UserAccessToken))

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Do(request)
	if err != nil {
		log.Error("Failed to make request to slack health check")
		return model.IntegrationState{}
	}
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error("Failed to decode response from slack")
		return model.IntegrationState{}
	}

	//switch case for all possible error
	if response["ok"] == true {
		return model.IntegrationState{
			State:   model.SYNCED,
			Message: U.GetPropertyValueAsString(response["error"]),
		}
	}

	return model.IntegrationState{
		State:   model.SYNC_PENDING,
		Message: U.GetPropertyValueAsString(response["error"]),
	}

}
