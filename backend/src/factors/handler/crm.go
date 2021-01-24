package handler

import (
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CRMResponsePayload impelements response object for crm properties
type CRMResponsePayload struct {
	Categorical []string `json:"categorical"`
	DateTime    []string `json:"datetime"`
}

// GetCRMObjectProperties godoc
// @Summary To get crm object properties by source.
// @Tags V1ApiSmartEvent
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param crm_source path string true "Source"
// @Param object_type path string true "Object type"
// @Success 200 {object} CRMResponsePayload
// @Router /{project_id}/v1/crm/{crm_source}/{object_type}/properties [get]
func GetCRMObjectProperties(c *gin.Context) {

	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid project id."})
		return
	}

	source := c.Params.ByName("crm_source")
	objectType := c.Params.ByName("object_type")

	if source == "" || objectType == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid query params."})
		return
	}

	var crmResponsePayload CRMResponsePayload
	if source == M.SmartCRMEventSourceSalesforce {
		crmResponsePayload.Categorical, crmResponsePayload.DateTime = M.GetSalesforceObjectPropertiesName(projectID, objectType)
	} else if source == M.SmartCRMEventSourceHubspot {
		crmResponsePayload.Categorical, crmResponsePayload.DateTime = M.GetHubspotObjectPropertiesName(projectID, objectType)
	}

	c.JSON(http.StatusOK, crmResponsePayload)
}
