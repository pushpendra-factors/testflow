package internal

import (
	"encoding/json"
	M "factors/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func DataServiceHubspotAddDocumentHandler(c *gin.Context) {
	r := c.Request

	var hubspotDocument M.HubspotDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&hubspotDocument); err != nil {
		log.WithError(err).Error("Failed to decode Json request on hubspot upsert document handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	errCode := M.CreateHubspotDocument(hubspotDocument.ProjectId, &hubspotDocument)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to upsert hubspot document."})
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully upserted hubspot document."})
}

func DataServiceHubspotGetSyncInfoHandler(c *gin.Context) {
	syncInfo, errCode := M.GetHubspotSyncInfo()
	c.JSON(errCode, syncInfo)
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

	formDocuments, errCode := M.GetHubspotFormDocuments(projectId)
	c.JSON(errCode, formDocuments)
}
