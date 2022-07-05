package handler

import (
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

// CRMResponsePayload impelements response object for crm properties
type CRMResponsePayload struct {
	Categorical []string `json:"categorical"`
	DateTime    []string `json:"datetime"`
}

// GetCRMObjectPropertiesHandler godoc
// @Summary To get crm object properties by source.
// @Tags V1ApiSmartEvent
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param crm_source path string true "Source"
// @Param object_type path string true "Object type"
// @Success 200 {object} CRMResponsePayload
// @Router /{project_id}/v1/crm/{crm_source}/{object_type}/properties [get]
func GetCRMObjectPropertiesHandler(c *gin.Context) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
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
	if source == model.SmartCRMEventSourceSalesforce {
		crmResponsePayload.Categorical, crmResponsePayload.DateTime = store.GetStore().GetSalesforceObjectPropertiesName(projectID, objectType)
	} else if source == model.SmartCRMEventSourceHubspot {
		crmResponsePayload.Categorical, crmResponsePayload.DateTime = store.GetStore().GetHubspotObjectPropertiesName(projectID, objectType)
	}

	c.JSON(http.StatusOK, crmResponsePayload)
}

// GetCRMObjectValuesByPropertyNameHandler godoc
// @Summary To get crm object values by source and property name.
// @Tags V1ApiSmartEvent
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param crm_source path string true "Source"
// @Param object_type path string true "Object type"
// @Success 200 {object} CRMResponsePayload
// @Router /{project_id}/v1/crm/{crm_source}/{object_type}/properties/{property_name}/values [get]
func GetCRMObjectValuesByPropertyNameHandler(c *gin.Context) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid project id."})
		return
	}

	source := c.Params.ByName("crm_source")
	objectType := c.Params.ByName("object_type")
	propertyName := c.Params.ByName("property_name")

	if source == "" || objectType == "" || propertyName == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid query params."})
		return
	}

	var properties []interface{}
	if source == model.SmartCRMEventSourceSalesforce {
		properties = store.GetStore().GetSalesforceObjectValuesByPropertyName(projectID, objectType, propertyName)
	} else if source == model.SmartCRMEventSourceHubspot {
		properties = store.GetStore().GetAllHubspotObjectValuesByPropertyName(projectID, objectType, propertyName)
	}

	for i, value := range properties {
		if reflect.ValueOf(value).Kind() == reflect.Bool {
			if value == true {
				properties[i] = "true"
			} else {
				properties[i] = "false"
			}
		}
	}
	properties = append([]interface{}{model.PropertyValueNone}, properties...)
	c.JSON(http.StatusOK, properties)
}
