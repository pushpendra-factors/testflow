package v1

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// CreateCustomMetric godoc
// @Summary To create custom metric.
// @Tags CustomMetric
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query body model.CustomMetric true "Custom metric payload."
// @Success 200 {string} json "{"result": model.CustomMetric}"
// @Router /{project_id}/v1/custom_metrics [post]
func CreateCustomMetric(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}
	request := model.CustomMetric{}
	err := c.BindJSON(&request)
	if err != nil {
		var requestAsMap map[string]interface{}
		c.BindJSON(&requestAsMap)
		logCtx.Warnf("Decode failed on request to profiles struct. %v", requestAsMap)
		return nil, http.StatusBadRequest, INVALID_INPUT, "Error during decode of custom metrics.", true
	}

	var customMetricTransformation model.CustomMetricTransformation
	err = U.DecodePostgresJsonbToStructType(request.Transformations, &customMetricTransformation)
	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, "Error during decode of custom metrics transformations.", true
	}
	request.ProjectID = projectID
	customMetric, errMsg, statusCode := store.GetStore().CreateCustomMetric(request)
	if statusCode != http.StatusCreated {
		logCtx.WithField("message finder", "Error during custom metric creation").Warn(errMsg)
		return customMetric, statusCode, PROCESSING_FAILED, errMsg, true
	}
	return customMetric, http.StatusOK, "", "", false
}

// GetCustomMetricsConfig godoc
// @Summary To get config for the building custom metrics on settings.
// @Tags CustomMetric
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "{"result": model.CustomMetricConfig"
// @Router /{project_id}/v1/custom_metrics/config [get]
func GetCustomMetricsConfig(c *gin.Context) {
	reqID, projectID := getReqIDAndProjectID(c)
	customMetricConfigs := model.CustomMetricConfig{}
	customMetricObjectTypesAndProperties := make([]model.CustomMetricObjectTypeAndProperties, 0)

	for _, objectType := range model.CustomMetricObjectTypeNames {
		currentObjectTypeAndProperties := model.CustomMetricObjectTypeAndProperties{}
		currentObjectTypeAndProperties.ObjectType = objectType
		currentObjectTypeAndProperties.Properties = getPropertiesFunctionBasedOnObjectType(objectType)(projectID, reqID)

		customMetricObjectTypesAndProperties = append(customMetricObjectTypesAndProperties, currentObjectTypeAndProperties)
	}

	customMetricConfigs.AggregateFunctions = model.CustomMetricAggregateFunctions
	customMetricConfigs.ObjectTypeAndProperties = customMetricObjectTypesAndProperties
	c.JSON(http.StatusOK, gin.H{"result": customMetricConfigs})
}

// GetCustomMetrics godoc
// @Summary To get custom metrics for a project id.
// @Tags CustomMetric
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "{"result": []model.CustomMetricConfig}"
// @Router /{project_id}/v1/custom_metrics [get]
func GetCustomMetrics(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "Invalid project id provided.", true
	}
	customMetrics, errMsg, statusCode := store.GetStore().GetCustomMetricsByProjectId(projectID)
	if statusCode != http.StatusFound {
		logCtx.WithField("message finder", "Failed to get custom metrics").Warn(errMsg)
		return nil, statusCode, PROCESSING_FAILED, "Failed to get custom metrics", true
	}
	return customMetrics, http.StatusOK, "", "", false
}

func getPropertiesFunctionBasedOnObjectType(objectType string) func(uint64, string) []map[string]string {
	if strings.Contains(objectType, U.CRM_SOURCE_NAME_HUBSPOT) {
		return store.GetStore().GetPropertiesForHubspot
	} else if strings.Contains(objectType, U.CRM_SOURCE_NAME_SALESFORCE) {
		return store.GetStore().GetPropertiesForSalesforce
	} else if strings.Contains(objectType, U.CRM_SOURCE_NAME_MARKETO) {
		return store.GetStore().GetPropertiesForMarketo
	}
	return nil
}
