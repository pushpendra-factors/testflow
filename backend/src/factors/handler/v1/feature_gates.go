package v1

import (
	"factors/model/store"
	U "factors/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	mid "factors/middleware"
)

func UpdateFeatureStatusHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	toBeUpdated := c.Query("update_status")
	featureName := c.Query("feature_name")
	intStatus, err := strconv.Atoi(toBeUpdated)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse incoming status to be updated"})
	}
	status, err := store.GetStore().UpdateStatusForFeature(projectId, featureName, intStatus)
	if err != nil || status != http.StatusAccepted {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feature status "})
	}
	c.JSON(http.StatusAccepted, "success")
	return 
}
