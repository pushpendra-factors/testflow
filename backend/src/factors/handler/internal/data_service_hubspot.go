package internal

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func DataServiceHubspotAddDocumentHandler(c *gin.Context) {
	r := c.Request

	var hubspotDocument model.HubspotDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&hubspotDocument); err != nil {
		log.WithError(err).Error("Failed to decode Json request on hubspot upsert document handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	newBytes := U.RemoveNullCharacterBytes(hubspotDocument.Value.RawMessage)

	if len(newBytes) != len(hubspotDocument.Value.RawMessage) {
		log.WithFields(log.Fields{"document_id": hubspotDocument.ID, "project_id": hubspotDocument.ProjectId,
			"raw_message":    string(hubspotDocument.Value.RawMessage),
			"sliced_message": string(newBytes)}).Warn("Using new sliced bytes for null character.")
		hubspotDocument.Value.RawMessage = newBytes
	}

	errCode := store.GetStore().CreateHubspotDocument(hubspotDocument.ProjectId, &hubspotDocument)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to upsert hubspot document."})
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully upserted hubspot document."})
}

func DataServiceHubspotGetSyncInfoHandler(c *gin.Context) {
	isFirstTime := c.Query("is_first_time") == "true"

	if isFirstTime {
		syncInfo, errCode := store.GetStore().GetHubspotFirstSyncProjectsInfo()
		c.JSON(errCode, syncInfo)
		return
	}

	syncInfo, errCode := store.GetStore().GetHubspotSyncInfo()
	c.JSON(errCode, syncInfo)
	return
}

type HubspotFirstTimeSyncRequestPayload struct {
	Status   string                           `json:"status"`
	Failures []model.HubspotProjectSyncStatus `json:"failures"`
	Success  []model.HubspotProjectSyncStatus `json:"success"`
}

func DataServiceHubspotUpdateFirstTimeSyncInfo(c *gin.Context) {
	r := c.Request

	var hubspotFirstTimeSyncRequestPayload HubspotFirstTimeSyncRequestPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&hubspotFirstTimeSyncRequestPayload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on hubspot first time sync update.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	status := store.GetStore().UpdateHubspotProjectSettingsBySyncStatus(hubspotFirstTimeSyncRequestPayload.Success,
		hubspotFirstTimeSyncRequestPayload.Failures)
	if status != http.StatusAccepted {
		c.Status(status)
		return
	}

	c.Status(http.StatusOK)
}

func DataServiceGetHubspotFormDocumentsHandler(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil || projectId == 0 {
		log.WithError(err).Error(
			"Failed to get project_id on get hubspot form documents.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid project id."})
		return
	}

	formDocuments, errCode := store.GetStore().GetHubspotFormDocuments(projectId)
	c.JSON(errCode, formDocuments)
}
