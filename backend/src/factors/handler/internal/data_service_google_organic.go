package internal

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func DataServiceGoogleOrganicAddDocumentHandler(c *gin.Context) {
	r := c.Request
	log.Warn("Inside GoogleOrganic Handler - add document.")
	var gscDocument model.GoogleOrganicDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&gscDocument); err != nil {
		log.WithError(err).Error("Failed to decode JSON request")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to decode JSON request."})
		return
	}

	errCode := store.GetStore().CreateGoogleOrganicDocument(&gscDocument)
	if errCode == http.StatusConflict {
		log.WithField("document", gscDocument).Error("Failed to insert the search console document on create document.")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Duplicate documents."})
		return
	}

	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed creating search console document."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Created search console document."})
}

// DataServiceGoogleOrganicAddMultipleDocumentsHandler can help bulk insert of 10
func DataServiceGoogleOrganicAddMultipleDocumentsHandler(c *gin.Context) {
	r := c.Request

	var gscDocuments []model.GoogleOrganicDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&gscDocuments); err != nil {
		log.WithError(err).Error("Failed to decode JSON request")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to decode JSON request."})
		return
	}
	errCode := store.GetStore().CreateMultipleGoogleOrganicDocument(gscDocuments)
	if errCode == http.StatusConflict {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Duplicate documents."})
		return
	}

	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed creating search console document."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Created search console document."})
}

func DataServiceGoogleOrganicGetLastSyncForProjectInfoHandler(c *gin.Context) {
	r := c.Request

	var payload model.GoogleOrganicLastSyncInfoPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on GoogleOrganic last sync info document handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}
	lastSyncInfo, status := store.GetStore().GetGoogleOrganicLastSyncInfoForProject(payload.ProjectId)
	c.JSON(status, lastSyncInfo)
}

// DataServiceGoogleOrganicGetLastSyncInfoHandler ...
func DataServiceGoogleOrganicGetLastSyncInfoHandler(c *gin.Context) {
	lastSyncInfo, status := store.GetStore().GetAllGoogleOrganicLastSyncInfoForAllProjects()
	log.Info(lastSyncInfo)
	c.JSON(status, lastSyncInfo)
}
