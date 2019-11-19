package internal

import (
	"encoding/json"
	M "factors/model"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func DataServiceAdwordsAddDocumentHandler(c *gin.Context) {
	r := c.Request

	var adwordsDocument M.AdwordsDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&adwordsDocument); err != nil {
		log.WithError(err).Error("Failed to decode JSON request")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to decode JSON request."})
		return
	}

	errCode := M.CreateAdwordsDocument(&adwordsDocument)
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": "Falied creating adwords document."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Created adwords document."})
}

func DataServiceAdwordsGetLastSyncInfoHandler(c *gin.Context) {
	lastSyncInfo, status := M.GetAllAdwordsLastSyncInfoByProjectAndType()
	c.JSON(status, lastSyncInfo)
}
