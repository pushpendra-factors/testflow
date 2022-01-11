package v1

import (
	"bytes"
	"encoding/json"
	"factors/config"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type contactProperty struct {
	Property string `json:"property"`
	Value    string `json:"value"`
}
type contactPayload struct {
	Properties []contactProperty `json:"properties"`
}

func HubspotCreateContact(c *gin.Context) {
	if config.GetConfig().HubspotAPIOnboardingHAPIKey == "" {
		log.Warn("HubspotAPIOnboardingHAPIKey missing.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "HubspotAPIOnboardingHAPIKey missing."})
		return
	}

	r := c.Request
	var contactPayload contactPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contactPayload); err != nil {
		log.WithError(err).Error("Failed to decode Json request in hubspot create contact handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	logCtx := log.WithFields(log.Fields{"contact_payload": contactPayload})
	var email string
	exists := false
	for i := range contactPayload.Properties {
		if contactPayload.Properties[i].Property == "email" {
			email = contactPayload.Properties[i].Value
			if email != "" {
				exists = true
			}
		}
	}

	if !exists {
		logCtx.Error("Failed to get email property.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Email property not found."})
		return
	}

	reqBody, err := json.Marshal(contactPayload)
	if err != nil {
		logCtx.Error("Unable to marshal requestMap in hubspot create contact handler")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to process request."})
		return
	}

	urlCreateContact := fmt.Sprintf("https://api.hubapi.com/contacts/v1/contact/createOrUpdate/email/%s/?hapikey=%s",
		email, config.GetConfig().HubspotAPIOnboardingHAPIKey)

	req, err := http.NewRequest("POST", urlCreateContact, bytes.NewBuffer(reqBody))
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to create POST request in hubspot create contact handler.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Failed to process request."})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 1 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to make POST request in hubspot create contact handler.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Failed to process request."})
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		logCtx.Error("Failed to execute POST request in hubspot create contact handler.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to process request."})
		return
	}

	var respBody map[string]interface{}
	decoder = json.NewDecoder(resp.Body)
	if err := decoder.Decode(&respBody); err != nil {
		logCtx.WithError(err).Error("Failed to decode Json request in hubspot create contact handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to process request."})
		return
	}

	c.JSON(http.StatusOK, respBody)
}

func GetHubspotContactByEmail(c *gin.Context) {
	if config.GetConfig().HubspotAPIOnboardingHAPIKey == "" {
		log.Warn("HubspotAPIOnboardingHAPIKey missing.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Not enabled."})
		return
	}

	email := c.Query("email")
	if email == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Missing email."})
		return
	}

	logCtx := log.WithFields(log.Fields{"email": email})
	urlGetContactByEmail := fmt.Sprintf("https://api.hubapi.com/contacts/v1/contact/email/%s/profile?hapikey=%s",
		email, config.GetConfig().HubspotAPIOnboardingHAPIKey)

	req, err := http.NewRequest("GET", urlGetContactByEmail, nil)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to create GET request in hubspot get contact by email handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "GET request failed."})
		return
	}
	client := &http.Client{
		Timeout: 1 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		logCtx.WithError(err).Error("Failed to make GET request in hubspot get contact by email handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to process request."})
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		logCtx.Error("Failed to execute GET request in hubspot get contact by email handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "GET request execution failed."})
		return
	}
	var responseDetails map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseDetails)

	if err != nil {
		logCtx.Error("Unable to decode response from GET request in hubspot create contact handler : %v", resp.Body)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request."})
		return
	}
	logCtx = logCtx.WithFields(log.Fields{"contact": responseDetails})

	properties, exists := responseDetails["properties"].(map[string]interface{})
	if !exists {
		logCtx.Error("Unable to get 'properties' property from response of GET request in hubspot create contact handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request."})
		return
	}

	hubspotOwnerID, exists := properties["hubspot_owner_id"].(map[string]interface{})
	if !exists {
		logCtx.Error("Unable to get 'hubspot_owner_id' property from response of GET request in hubspot create contact handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request."})
		return
	}

	value, exists := hubspotOwnerID["value"]
	if !exists {
		logCtx.Error("Unable to get 'value' property from response of GET request in hubspot create contact handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request."})
		return
	}
	responseMap := make(map[string]interface{})
	responseMap["hubspot_owner_id"] = value
	c.JSON(http.StatusOK, responseMap)
}
