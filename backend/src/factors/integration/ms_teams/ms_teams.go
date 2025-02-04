package teams

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

type oauthState struct {
	ProjectID int64   `json:"project_id"`
	AgentUUID *string `json:"agent_uuid"`
}
type TeamsMessage struct {
	Body struct {
		ContentType string `json:"contentType"`
		Content     string `json:"content"`
	} `json:"body"`
}

// teams and channels related structs
type Team struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type Group struct {
	Id      string `json:"id"`
	WebUrl  string `json:"webUrl"`
	IsTeam  bool   `json:"isTeam"`
	GroupId string `json:"groupId"`
}

type Channel struct {
	Id   string `json:"id"`
	Name string `json:"displayName"`
}

func SendTeamsMessage(projectID int64, agentUUID, teamID, channelID, message string) (map[string]interface{}, error) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"agent_uuid": agentUUID,
		"team_id":    teamID,
		"channel_id": channelID,
	})

	tokens, err := store.GetStore().GetTeamsAuthTokens(projectID, agentUUID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get access token for teams")
		return nil, errors.New("failed to get access tokens for teams")
	}
	teamsMessage := TeamsMessage{Body: struct {
		ContentType string "json:\"contentType\""
		Content     string "json:\"content\""
	}{ContentType: "html",
		Content: message}}

	jsonMessage, err := json.Marshal(teamsMessage)
	if err != nil {
		logCtx.WithError(err).Error("Failed to marshal teams message.")
		return nil, err
	}
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/teams/%s/channels/%s/messages", teamID, channelID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonMessage))
	if err != nil {
		logCtx.WithError(err).Error("Failed to create request for sending teams message.")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logCtx.WithError(err).Error("Failed to send request")
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logCtx.WithError(err).Error("Failed to read response body")
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		var errorResponse map[string]interface{}
		json.Unmarshal(body, &errorResponse)
		errorCode, ok := errorResponse["error"].(map[string]interface{})["code"].(string)
		if ok && errorCode == "InvalidAuthenticationToken" {
			token, err := RefreshAndGetTeamsAccessToken(projectID, agentUUID)
			if err != nil {
				logCtx.WithError(err).Error("Failed to refresh/get access token for teams")
				return errorResponse, err
			}
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err = client.Do(req)
			if err != nil {
				logCtx.WithError(err).Error("Failed to make send message request.")
				return errorResponse, err
			}
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				logCtx.WithError(err).Error("Failed to read response body")
				return errorResponse, err
			}

			defer resp.Body.Close()

		} else {
			logCtx.WithField("error_code", errorCode).Error("Error in making request to teams.")
			// add healthcheck/sentry notifcation to re integrate
			return errorResponse, fmt.Errorf("teams failure: status - %v", resp.Status)
		}
	}

	return nil, nil
}

// func to get list of teams
func GetAllTeams(projectID int64, agentUUID string) (interface{}, error) {
	tokens, err := store.GetStore().GetTeamsAuthTokens(projectID, agentUUID)
	if err != nil {
		return []Team{}, errors.New("failed to get access tokens for teams")
	}

	url := "https://graph.microsoft.com/v1.0/me/joinedTeams"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("Failed to read response body")
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		var errorResponse map[string]interface{}
		json.Unmarshal(body, &errorResponse)
		errorCode, ok := errorResponse["error"].(map[string]interface{})["code"].(string)
		if ok && errorCode == "InvalidAuthenticationToken" {
			accessToken, err := RefreshAndGetTeamsAccessToken(projectID, agentUUID)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", "Bearer "+accessToken)

			client := &http.Client{}
			resp, err = client.Do(req)
			if err != nil {
				return nil, err
			}
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				log.WithError(err).Error("Failed to read response body")
				return nil, err
			}

		} else {
			//
			log.Error("Error in making request to teams " + errorCode)
			// add healthcheck/sentry notifcation to re integrate
			return nil, errors.New("Error in making request to teams " + errorCode)
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get Teams teams: %v", resp.Status)
	}

	var teamList struct {
		Value []Team `json:"value"`
	}
	if err := json.Unmarshal(body, &teamList); err != nil {
		return "", err
	}
	return teamList.Value, nil
}

// func to get list of channels in a team.
func GetTeamsChannels(projectID int64, agentUUID, teamID string) ([]Channel, error) {
	tokens, err := store.GetStore().GetTeamsAuthTokens(projectID, agentUUID)
	if err != nil {
		return []Channel{}, errors.New("failed to get access tokens for teams")
	}
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/teams/%s/channels", teamID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get Teams channels: %v", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("Failed to read response body")
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		var errorResponse map[string]interface{}
		json.Unmarshal(body, &errorResponse)
		errorCode, ok := errorResponse["error"].(map[string]interface{})["code"].(string)
		if ok && errorCode == "InvalidAuthenticationToken" {
			token, err := RefreshAndGetTeamsAccessToken(projectID, agentUUID)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err = client.Do(req)
			if err != nil {
				return nil, err
			}
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				log.WithError(err).Error("Failed to read response body")
				return nil, err
			}

		} else {
			//
			log.Error("Error in making request to teams " + errorCode)
			// add healthcheck/sentry notifcation to re integrate
			return nil, errors.New("Error in making request to teams " + errorCode)
		}
	}
	var channelList struct {
		Value []Channel `json:"value"`
	}

	if err := json.Unmarshal(body, &channelList); err != nil {
		return nil, err
	}

	return channelList.Value, nil
}

func RefreshAndGetTeamsAccessToken(projectID int64, agentUUID string) (string, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"agent_id": agentUUID,
	}
	logCtx := log.WithFields(logFields)

	accessTokens, err := store.GetStore().GetTeamsAuthTokens(projectID, agentUUID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get access token for teams")
		return "", err
	}
	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("client_id", C.GetTeamsClientID())
	form.Add("client_secret", C.GetTeamsClientSecret())
	form.Add("refresh_token", accessTokens.RefreshToken)

	request, err := http.NewRequest("POST", "https://login.microsoftonline.com/common/oauth2/v2.0/token", strings.NewReader(form.Encode()))
	if err != nil {
		logCtx.WithError(err).Error("Failed to create token refresh request.")
		return "", err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logCtx.WithError(err).Error("Failed to make token refresh request.")
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logCtx.WithError(err).Error("Failed to get refresh token for teams.")
		return "", fmt.Errorf("refresh token request failed with status code %d", resp.StatusCode)
	}

	var tokens model.TeamsAccessTokens
	err = json.NewDecoder(resp.Body).Decode(&tokens)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode token for teams.")
		return "", err
	}

	tokens.LastRefreshedAt = time.Now()
	err = store.GetStore().SetAuthTokenforTeamsIntegration(projectID, agentUUID, tokens)
	if err != nil {
		log.WithError(err).Error("Failed to update access tokens for teams after refresh")
	}

	return tokens.AccessToken, nil
}
