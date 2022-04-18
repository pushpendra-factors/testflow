package handler

import (
	// "math/rand"
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func getOwner() string {
	// ownersData := []string{
	// 	"145127740",
	// 	"116053799",
	// }
	// return ownersData[rand.Intn(len(ownersData))]
	return "116053799"
}

func onboardingMailModoAPICall(agent *model.Agent) int {

	logCtx := log.WithFields(log.Fields{"agent": agent})

	data := map[string]interface{}{
		"email": agent.Email,
		"data":  map[string]interface{}{},
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, C.GetConfig().MailModoOnboardingURL1).
		WithHeader("Content-Type", "application/json").
		WithHeader("mmapikey", C.GetConfig().MailModoOnboardingAPIKey).
		// WithHeader("mmapikey", "TJ5JF61-44NMRN5-GAEA2WH-8Z99P4H").
		WithPostParams(data)

	req, err := rb.Build()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build request.")
		return http.StatusInternalServerError
	}

	client := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(req)
	if err != nil {
		logCtx.WithError(err).Error("Failed to make POST mail modo request 1.")
		return http.StatusInternalServerError
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logCtx.Error("Failed to execute POST mail modo request 1. -" + resp.StatusCode)
		return http.StatusInternalServerError
	}

	rb = C.NewRequestBuilderWithPrefix(http.MethodPost, C.GetConfig().MailModoOnboardingURL2).
		WithHeader("Content-Type", "application/json").
		WithHeader("mmapikey", C.GetConfig().MailModoOnboardingAPIKey).
		// WithHeader("mmapikey", "TJ5JF61-44NMRN5-GAEA2WH-8Z99P4H").
		WithPostParams(data)

	req, err = rb.Build()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build request.")
		return http.StatusInternalServerError
	}

	resp, err = client.Do(req)
	if err != nil {
		logCtx.WithError(err).Error("Failed to make POST mail modo request 2.")
		return http.StatusInternalServerError
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logCtx.Error("Failed to execute POST mail modo request 2.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func onboardingSlackAPICall(agent *model.Agent) int {

	logCtx := log.WithFields(log.Fields{"agent": agent})

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, C.GetConfig().SlackOnboardingWebhookURL).
		WithHeader("Content-Type", "application/json").
		WithPostParams(map[string]interface{}{
			"text":       "User " + agent.FirstName + " with email " + agent.Email + " just signed up",
			"username":   "Signup User Actions",
			"icon_emoji": ":golf:",
		})

	req, err := rb.Build()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build request.")
		return http.StatusInternalServerError
	}

	client := http.Client{
		Timeout: 1 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		logCtx.WithError(err).Error("Failed to make POST slack notification request.")
		return http.StatusInternalServerError
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logCtx.Error("Failed to execute POST slack notification request.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

type contactProperty struct {
	Property string `json:"property"`
	Value    string `json:"value"`
}
type contactPayload struct {
	Properties []contactProperty `json:"properties"`
}

func onboardingGetHubspotOwner(agent *model.Agent) string {

	owner := getOwner()

	if C.GetConfig().HubspotAPIOnboardingHAPIKey == "" {
		log.Warn("HubspotAPIOnboardingHAPIKey missing.")
		return owner
	}

	logCtx := log.WithFields(log.Fields{"email": agent.Email})

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("https://api.hubapi.com/contacts/v1/contact/email/%s/profile?hapikey=%s",
		agent.Email, C.GetConfig().HubspotAPIOnboardingHAPIKey))

	req, err := rb.Build()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build request.")
		return owner
	}

	client := &http.Client{
		Timeout: 1 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		logCtx.WithError(err).Error("Failed to make GET request in hubspot get contact by email handler.")
		return owner
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		return owner
	}
	var responseDetails map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseDetails)
	if err != nil {
		logCtx.Error("Unable to decode response from GET request in hubspot create contact handler : %v", resp.Body)
		return owner
	}

	properties, exists := responseDetails["properties"].(map[string]interface{})
	if exists {
		hubspotOwnerID, exists := properties["hubspot_owner_id"].(map[string]interface{})
		if exists {
			hubspotOwnerIDValue, exists := hubspotOwnerID["value"]
			if exists {
				owner = hubspotOwnerIDValue.(string)
			}
		}
	}
	return owner
}

func onboardingHubspotOwner(agent *model.Agent) int {
	owner := onboardingGetHubspotOwner(agent)

	contactData := contactPayload{
		Properties: []contactProperty{
			{
				Property: "email",
				Value:    agent.Email,
			},
			{
				Property: "firstname",
				Value:    agent.FirstName,
			},
			{
				Property: "lastname",
				Value:    agent.LastName,
			},
			{
				Property: "phone",
				Value:    agent.Phone,
			},
			{
				Property: "hubspot_owner_id",
				Value:    owner,
			},
			{
				Property: "signup_method",
				Value:    "Self-Serve Onboarding",
			},
		},
	}
	logCtx := log.WithFields(log.Fields{"contact_data": contactData})

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("https://api.hubapi.com/contacts/v1/contact/createOrUpdate/email/%s?hapikey=%s",
		agent.Email, C.GetConfig().HubspotAPIOnboardingHAPIKey)).
		WithHeader("Content-Type", "application/json").
		WithPostParams(contactData)

	req, err := rb.Build()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build request.")
		return http.StatusInternalServerError
	}

	client := &http.Client{
		Timeout: 1 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		logCtx.WithError(err).Error("Failed to make POST request in hubspot create contact handler.")
		return http.StatusInternalServerError
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		logCtx.Error("Failed to execute POST request in hubspot create contact handler.")
		return http.StatusInternalServerError
	}

	var respBody map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&respBody); err != nil {
		logCtx.WithError(err).Error("Failed to decode Json request in hubspot create contact handler.")
		return http.StatusInternalServerError
	}
	logCtx.Error(respBody)
	return http.StatusOK
}
