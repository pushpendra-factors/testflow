package handler

import (
	"factors/integration/linkedin_capi"
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLinkedinCAPIConversionsList(c *gin.Context) {

	workflowId := c.Query("workflow_id")

	if workflowId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid workflow_id."})
		return
	}

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{"error": " Invalid project."})
		return
	}

	config, err := store.GetStore().GetLinkedInCAPICofigByWorkflowId(projectId, workflowId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	response, err := linkedin_capi.GetConversionFromLinkedCAPI(config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to get conversions "})
		return
	}

	if len(response.LinkedInCAPIConversionsResponseList) == 0 {
		c.AbortWithStatusJSON(http.StatusExpectationFailed, gin.H{"error": "nil response from paragon"})
		return
	}

	c.JSON(http.StatusOK, response)
	return

}
