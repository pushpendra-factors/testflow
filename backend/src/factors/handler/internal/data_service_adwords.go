package internal

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func DataServiceAdwordsAddDocumentHandler(c *gin.Context) {
	r := c.Request

	var adwordsDocument model.AdwordsDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&adwordsDocument); err != nil {
		log.WithError(err).Error("Failed to decode JSON request")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to decode JSON request."})
		return
	}

	errCode := store.GetStore().CreateAdwordsDocument(&adwordsDocument)
	if errCode == http.StatusConflict {
		log.WithField("document", adwordsDocument).Error("Failed to insert the adword document on create document.")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Duplicate documents."})
		return
	}

	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Failed creating adwords document."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Created adwords document."})
}

func DataServiceAdwordsGetLastSyncInfoHandler(c *gin.Context) {
	lastSyncInfo, status := store.GetStore().GetAllAdwordsLastSyncInfoByProjectCustomerAccountAndType()
	c.JSON(status, lastSyncInfo)
}
