package internal

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func DataServiceFacebookGetProjectSettings(c *gin.Context) {
	facebookIDs, facebookProjectSettings, status := store.GetStore().GetFacebookEnabledIDsAndProjectSettings()
	projects, _ := store.GetStore().GetProjectsByIDs(facebookIDs)
	for _, project := range projects {
		for index := range facebookProjectSettings {
			if facebookProjectSettings[index].ProjectId == project.ID && facebookProjectSettings[index].Timezone == "" {
				facebookProjectSettings[index].Timezone = project.TimeZone
			}
		}
	}
	c.JSON(status, facebookProjectSettings)
}
func DataServiceFacebookAddDocumentHandler(c *gin.Context) {
	r := c.Request

	var facebookDocument model.FacebookDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&facebookDocument); err != nil {
		log.WithError(err).Error("Failed to decode Json request on facebook upsert document handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	errCode := store.GetStore().CreateFacebookDocument(facebookDocument.ProjectID, &facebookDocument)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest || errCode == http.StatusConflict {
		log.WithField("document", facebookDocument).Error("Failed to insert the facebook document on create document.")
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed to upsert facebook document."})
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully upserted facebook document."})
}

func DataServiceFacebookGetLastSyncInfoHandler(c *gin.Context) {
	r := c.Request

	var payload model.FacebookLastSyncInfoPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on facebook last sync info document handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}
	lastSyncInfo, status := store.GetStore().GetFacebookLastSyncInfo(payload.ProjectId, payload.CustomerAdAccountId)
	c.JSON(status, lastSyncInfo)
}
