package internal

import (
	"encoding/json"
	M "factors/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func DataServiceFacebookGetProjectSettings(c *gin.Context) {
	facebookProjectSettings, status := M.GetFacebookEnabledProjectSettings()
	c.JSON(status, facebookProjectSettings)
}

func DataServiceFacebookAddDocumentHandler(c *gin.Context) {
	r := c.Request

	var facebookDocument M.FacebookDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&facebookDocument); err != nil {
		log.WithError(err).Error("Failed to decode Json request on facebook upsert document handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	errCode := M.CreateFacebookDocument(facebookDocument.ProjectId, &facebookDocument)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest || errCode == http.StatusConflict {
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed to upsert facebook document."})
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully upserted facebook document."})
}

func DataServiceFacebookGetLastSyncInfoHandler(c *gin.Context) {
	r := c.Request

	var payload M.FacebookLastSyncInfoPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on facebook upsert document handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}
	projectID, _ := strconv.ParseUint(payload.ProjectId, 10, 64)
	lastSyncInfo, status := M.GetFacebookLastSyncInfo(projectID, payload.CustomerAdAccountId)
	c.JSON(status, lastSyncInfo)
}
