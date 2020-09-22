package internal

import (
	"encoding/json"
	M "factors/model"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func DataServiceSalesforceAddDocumentHandler(c *gin.Context) {
	r := c.Request

	var salesforceDocument M.SalesforceDocument
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&salesforceDocument); err != nil {
		log.WithError(err).Error("Failed to decode Json request on salesforce upsert document handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	errCode := M.CreateSalesforceDocument(salesforceDocument.ProjectId, &salesforceDocument)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to upsert salesforce document."})
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully upserted salesforce document."})
}
