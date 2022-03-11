package internal

import (
	"encoding/json"
	C "factors/config"
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

type HubspotBatchInsertPayload struct {
	ProjectID    uint64                   `json:"project_id"`
	DocTypeAlias string                   `json:"doc_type"`
	Documents    []*model.HubspotDocument `json:"documents"`
}

func DataServiceHubspotAddBatchDocumentHandler(c *gin.Context) {
	r := c.Request

	var hubspotBatchInsertPayload HubspotBatchInsertPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&hubspotBatchInsertPayload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on hubspot upsert document handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	docType, err := model.GetHubspotTypeByAlias(hubspotBatchInsertPayload.DocTypeAlias)
	if err != nil {
		log.WithField("doc_type_alias", hubspotBatchInsertPayload.DocTypeAlias).
			WithError(err).Error("invalid documet type.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid document type."})
		return
	}

	projectID := hubspotBatchInsertPayload.ProjectID
	documents := hubspotBatchInsertPayload.Documents

	if !C.UseHubspotBatchInsertByProjectID(projectID) {
		log.WithFields(log.Fields{"project_id": projectID}).Error("Batch insert not enabled for this project.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Project not enabled for batch insert."})
		return
	}

	for i := range documents {
		newBytes := U.RemoveNullCharacterBytes(documents[i].Value.RawMessage)

		if len(newBytes) != len(documents[i].Value.RawMessage) {
			log.WithFields(log.Fields{"document_id": documents[i].ID, "project_id": projectID, "doc_type": docType,
				"raw_message":    string(documents[i].Value.RawMessage),
				"sliced_message": string(newBytes)}).Warn("Using new sliced bytes for null character.")
			documents[i].Value.RawMessage = newBytes
		}
	}

	batchSize := C.GetHubspotBatchInsertBatchSize()
	errCode := store.GetStore().CreateHubspotDocumentInBatch(hubspotBatchInsertPayload.ProjectID,
		docType, documents, batchSize)
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

type HubspotSyncRequestPayload struct {
	Status   string                           `json:"status"`
	Failures []model.HubspotProjectSyncStatus `json:"failures"`
	Success  []model.HubspotProjectSyncStatus `json:"success"`
}

func DataServiceHubspotUpdateSyncInfo(c *gin.Context) {
	r := c.Request
	isFirstTime := c.Query("is_first_time") == "true"
	var hubspotSyncRequestPayload HubspotSyncRequestPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&hubspotSyncRequestPayload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on hubspot sync update.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	status := store.GetStore().UpdateHubspotProjectSettingsBySyncStatus(hubspotSyncRequestPayload.Success,
		hubspotSyncRequestPayload.Failures, isFirstTime)
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
