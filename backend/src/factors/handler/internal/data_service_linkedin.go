package internal

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func DataServiceLinkedinGetProjectSettings(c *gin.Context) {
	linkedinProjectSettings, status := store.GetStore().GetLinkedinEnabledProjectSettings()
	c.JSON(status, linkedinProjectSettings)
}
func DataServiceLinkedinAddDocumentHandler(c *gin.Context) {
	r := c.Request

	var linkedinDocument model.LinkedinDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&linkedinDocument); err != nil {
		log.WithError(err).Error("Failed to decode Json request on linkedin upsert document handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	errCode := store.GetStore().CreateLinkedinDocument(linkedinDocument.ProjectID, &linkedinDocument)
	if errCode != http.StatusCreated {
		log.WithField("document", linkedinDocument).Error("Failed to insert the linkedin document on create document.")
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed to upsert linkedin document."})
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully upserted linkedin document."})
}

func DataServiceLinkedinGetLastSyncInfoHandler(c *gin.Context) {
	r := c.Request

	var payload model.LinkedinLastSyncInfoPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on linkedin get last sync info handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}
	projectID, _ := strconv.ParseUint(payload.ProjectID, 10, 64)
	lastSyncInfo, status := store.GetStore().GetLinkedinLastSyncInfo(projectID, payload.CustomerAdAccountID)
	c.JSON(status, lastSyncInfo)
}
