package handler

import (
	"encoding/json"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type AttributionRequestPayload struct {
	Query *M.AttributionQuery `json:"query"`
}

func AttributionHandler(c *gin.Context) {
	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	r := c.Request

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Query failed. Invalid project."})
		return
	}

	var requestPayload AttributionRequestPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Query failed. Json decode failed."})
		return
	}
	result, meta, err := M.ExecuteAttributionQuery(projectId, requestPayload.Query)
	if err != nil {
		logCtx.WithError(err).Error("Query failed. Query execution failed")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Query failed. Query execution failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"metrics_breakdown": result, "Meta": gin.H{"currency": meta}})
}
