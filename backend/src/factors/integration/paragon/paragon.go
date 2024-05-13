package paragon

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

const (
	Expiry                              int64  = math.MaxInt64
	nameOfEvent                         string = "EventName" //`GS [Search -> -> Create or Update]`
	getParagonProjectIntegrationsUrl    string = "https://api.useparagon.com/projects/%s/sdk/integrations"
	getParagonUserUrl                   string = "https://api.useparagon.com/projects/%s/sdk/me"
	getParagonUserConnectCredentialsUrl string = "https://api.useparagon.com/projects/%s/sdk/credentials"
	getParagonIntegrationMetadataUrl    string = "https://api.useparagon.com/projects/%s/sdk/metadata"
	triggerParagonWorkflowUrl           string = "https://api.useparagon.com/projects/%s/sdk/triggers/%s"
	deleteParagonProjectIntegrationUrl  string = "https://api.useparagon.com/projects/%s/sdk/integrations/%s"
	disableParagonWorkflowForUserUrl    string = "https://api.useparagon.com/projects/%s/sdk/workflows/%s"
	LinkedInAudienceId                  string = "LinkedIn Audience ID"
	ParagonLinkedInSuffix               string = "PLI"
)

func GenerateJWTTokenForProject(projectID int64) (string, error) {

	if projectID == 0 {
		log.WithField("project_id", projectID).Error("invalid parameter")
		return "", fmt.Errorf("invalid parameter")
	}

	factorsSigningKey := config.GetParagonTokenSigningKey()
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(factorsSigningKey))
	if err != nil {
		log.WithField("project_id", projectID).WithError(err).Error("private key could not be found")
		return "", err
	}

	return GetSignedToken(key, projectID)
}

func GetSignedToken(signingKey *rsa.PrivateKey, projectID int64) (string, error) {
	// issued at
	iat := &jwt.NumericDate{
		Time: time.Now(),
	}
	// expiry set to maximum possible int
	exp := &jwt.NumericDate{
		Time: time.Unix(math.MaxInt64, 0),
	}
	claims := &jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", projectID),
		IssuedAt:  iat,
		ExpiresAt: exp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(signingKey)
}

func SendPayloadToParagonForTheAlert(projectID int64, alertID string, alert *model.EventTriggerAlertConfig, payload *model.CachedEventTriggerAlert) (map[string]interface{}, error) {

	if projectID == 0 || alertID == "" || alert == nil || payload == nil {
		return nil, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id": projectID,
		"alert_id":   alertID,
		"payload":    *payload,
	}
	logCtx := log.WithFields(logFields)

	metadata, errCode, err := store.GetStore().GetParagonMetadataForEventTriggerAlert(projectID, alertID)
	if err != nil || errCode != http.StatusFound {
		logCtx.WithError(err).Error("no metadata found")
		return nil, err
	}

	if strings.HasSuffix(alert.Title, ParagonLinkedInSuffix) {
		size := len(payload.Message.MessageProperty)
		payload.Message.MessageProperty[fmt.Sprintf("%d", size)] = model.MessagePropMapStruct{
			DisplayName: LinkedInAudienceId,
			PropValue:   metadata[LinkedInAudienceId],
		}
	}

	nameOfParagonEvent := U.GetPropertyValueAsString(metadata[nameOfEvent])
	newPayload := transformEventTriggerToParagonPayload(payload, nameOfParagonEvent)

	signedToken, errCode, err := store.GetStore().GetParagonTokenFromProjectSetting(projectID)
	if err != nil || errCode != http.StatusFound {
		log.WithError(err).Error("no auth token found for paragon")
		return nil, err
	}

	response, err := SendParagonEventRequest(alert.WebhookURL, signedToken, newPayload)
	if err != nil {
		log.WithError(err).Error("failed to make request for paragon event")
		return response, err
	}

	return response, nil
}

type ParagonPayload struct {
	Name    string      `json:"name"`
	Payload interface{} `json:"payload"`
}

func transformEventTriggerToParagonPayload(payload *model.CachedEventTriggerAlert, nameofEvent string) ParagonPayload {

	var paragonLoad = ParagonPayload{
		Name:    nameofEvent,
		Payload: payload.Message,
	}

	return paragonLoad
}

func SendParagonEventRequest(url, token string, payload interface{}) (map[string]interface{}, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	bearer := fmt.Sprintf("Bearer %s", token)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("failed to make request for paragon event")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	response := make(map[string]interface{})
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		response["status"] = "success"
		response["body"] = string(bodyBytes)
	} else {
		log.WithField("request", request).Error("Failed to send webhook request")
		response["status"] = "failure"
		response["error"] = string(bodyBytes)
		response["statuscode"] = resp.StatusCode
	}
	return response, nil
}

func GetParagonUserAPI(token, paragonProjectID string) (map[string]interface{}, error) {

	request, err := http.NewRequest("GET", fmt.Sprintf(getParagonUserUrl, paragonProjectID), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	bearer := fmt.Sprintf("Bearer %s", token)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("failed to make request for paragon event")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	response := make(map[string]interface{})
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		response["status"] = "success"
		response["body"] = string(bodyBytes)
	} else {
		log.WithField("request", request).Error("Failed to send webhook request")
		response["status"] = "failure"
		response["error"] = string(bodyBytes)
		response["statuscode"] = resp.StatusCode
	}

	return response, nil
}

func GetParagonIntegrationMetadataAPI(token, paragonProjectID string) (map[string]interface{}, error) {

	request, err := http.NewRequest("GET", fmt.Sprintf(getParagonIntegrationMetadataUrl, paragonProjectID), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	bearer := fmt.Sprintf("Bearer %s", token)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("failed to make request for paragon event")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	response := make(map[string]interface{})
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		response["status"] = "success"
		response["body"] = string(bodyBytes)
	} else {
		log.WithField("request", request).Error("Failed to send webhook request")
		response["status"] = "failure"
		response["error"] = string(bodyBytes)
		response["statuscode"] = resp.StatusCode
	}

	return response, nil
}

func GetParagonUsersConnectCredentialsAPI(token, paragonProjectID string) (map[string]interface{}, error) {

	request, err := http.NewRequest("GET", fmt.Sprintf(getParagonUserConnectCredentialsUrl, paragonProjectID), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	bearer := fmt.Sprintf("Bearer %s", token)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("failed to make request for paragon event")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	response := make(map[string]interface{})
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		response["status"] = "success"
		response["body"] = string(bodyBytes)
	} else {
		log.WithField("request", request).Error("Failed to send webhook request")
		response["status"] = "failure"
		response["error"] = string(bodyBytes)
		response["statuscode"] = resp.StatusCode
	}

	return response, nil
}

func GetParagonProjectIntegrationsAPI(token, paragonProjectID string) (map[string]interface{}, error) {

	request, err := http.NewRequest("GET", fmt.Sprintf(getParagonProjectIntegrationsUrl, paragonProjectID), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	bearer := fmt.Sprintf("Bearer %s", token)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("failed to make request for paragon event")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	response := make(map[string]interface{})
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		response["status"] = "success"
		response["body"] = string(bodyBytes)
	} else {
		log.WithField("request", request).Error("Failed to send webhook request")
		response["status"] = "failure"
		response["error"] = string(bodyBytes)
		response["statuscode"] = resp.StatusCode
	}

	return response, nil
}

func DeleteParagonProjectIntegrationAPI(token, paragonProjectID, integrationID string) (map[string]interface{}, error) {

	request, err := http.NewRequest("DELETE", fmt.Sprintf(deleteParagonProjectIntegrationUrl, paragonProjectID, integrationID), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	bearer := fmt.Sprintf("Bearer %s", token)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("failed to make request for paragon event")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	response := make(map[string]interface{})
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		response["status"] = "success"
		response["body"] = string(bodyBytes)
	} else {
		log.WithField("request", request).Error("Failed to send webhook request")
		response["status"] = "failure"
		response["error"] = string(bodyBytes)
		response["statuscode"] = resp.StatusCode
	}

	return response, nil
}

func TriggerParagonWorkflowAPI(token, paragonProjectID, workflowID string, payload interface{}) (map[string]interface{}, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", fmt.Sprintf(triggerParagonWorkflowUrl, paragonProjectID, workflowID), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	bearer := fmt.Sprintf("Bearer %s", token)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("failed to make request for paragon event")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	response := make(map[string]interface{})
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		response["status"] = "success"
		response["body"] = string(bodyBytes)
	} else {
		log.WithField("request", request).Error("Failed to send webhook request")
		response["status"] = "failure"
		response["error"] = string(bodyBytes)
		response["statuscode"] = resp.StatusCode
	}

	return response, nil
}

func DisableParagonWorkflowForUserAPI(token, paragonProjectID, workflowID string) (map[string]interface{}, error) {

	request, err := http.NewRequest("DELETE", fmt.Sprintf(disableParagonWorkflowForUserUrl, paragonProjectID, workflowID), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	bearer := fmt.Sprintf("Bearer %s", token)
	request.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("failed to make request for paragon event")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	response := make(map[string]interface{})
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		response["status"] = "success"
		response["body"] = string(bodyBytes)
	} else {
		log.WithField("request", request).Error("Failed to send webhook request")
		response["status"] = "failure"
		response["error"] = string(bodyBytes)
		response["statuscode"] = resp.StatusCode
	}

	return response, nil
}

func SendPayloadToParagonWorkflow(projectID int64, url string, payload *model.CachedEventTriggerAlert) (map[string]interface{}, error) {

	if projectID == 0 || payload == nil {
		return nil, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id": projectID,
		"payload":    *payload,
	}
	logCtx := log.WithFields(logFields)


	signedToken, errCode, err := store.GetStore().GetParagonTokenFromProjectSetting(projectID)
	if err != nil || errCode != http.StatusFound {
		logCtx.WithError(err).Error("Failed to fetch auth token for paragon.")
		return nil, err
	}

	var newPayload model.WorkflowParagonPayload
	err = U.DecodeInterfaceMapToStructType(payload.Message.MessageProperty, &newPayload)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode payload. Workflow trigger cancelled.")
		return nil, err
	}

	response, err := SendParagonEventRequest(url, signedToken, newPayload)
	if err != nil {
		logCtx.WithError(err).Error("Failed to trigger workflow.")
		return response, err
	}

	return response, nil
}