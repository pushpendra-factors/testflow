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
)

func GenerateJWTTokenForProject(projectID int64) (string, error) {

	if projectID == 0 {
		log.WithField("project_id", projectID).Error("invalid parameter")
		return "", fmt.Errorf("invalid parameter")
	}

	factorsSigningKey := "-----BEGIN PRIVATE KEY-----\n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\n5bbe7sHDmrpvM7CzAJGl8/y0ydfx\n-----END PRIVATE KEY-----"
	configKey := config.GetParagonTokenSigningKey()
	log.Info("####Config key - ", configKey)
	log.Info("$$$$$Look here - ", factorsSigningKey)
	log.Info("@@@@@Equality - ", configKey == factorsSigningKey)
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
