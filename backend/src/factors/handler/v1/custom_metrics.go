package v1

import (
	C "factors/config"
	DD "factors/default_data"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// TODO @kark We are using custom Metrics convention everywhere. But we intended the below to be used for custom KPI convention. This needs change.
// TODO @kark Think how we are going to transition customMetric model to customKPI interface based.

// CreateCustomMetric godoc.
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

	isValid, errMsg := request.IsValid()
	if isValid == false {
		return nil, http.StatusBadRequest, INVALID_INPUT, errMsg, true
	}

	request.SetDefaultGroupIfRequired()
	request.ProjectID = projectID
	customMetric, errMsg, statusCode := store.GetStore().CreateCustomMetric(request)
	if statusCode != http.StatusCreated {
		logCtx.WithField("message finder", "Error during custom metric creation").Warn(errMsg)
		if strings.Contains(errMsg, "Duplicate") {
			return nil, http.StatusConflict, DUPLICATE_RECORD, ErrorMessages[DUPLICATE_RECORD], true
		} else {
			return nil, statusCode, PROCESSING_FAILED, ErrorMessages[PROCESSING_FAILED], true
		}
	}
	return customMetric, http.StatusOK, "", "", false
}

// CreateMissingPreBuiltCustomKPI godoc.
// @Summary To create missing custom kpi.
// @Tags CustomMetric
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query integration string true "Integration name"
// @Success 200 {string} json "{"result": }"
// @Router /{project_id}/v1/custom_metrics/prebuilt/add_missing [get]
func CreateMissingPreBuiltCustomKPI(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}
	integrationString := c.Query("integration")
	// TODO add validation.
	if integrationString == "" {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}

	isFirstTimeIntegrationDone, statusCode := DD.CheckIfFirstTimeIntegrationDone(projectID, integrationString)
	if statusCode != http.StatusFound {
		errMsg := fmt.Sprintf("Failed during first time integration check marketo: %v", projectID)
		C.PingHealthcheckForFailure(C.HealthCheckPreBuiltCustomKPIPingID, errMsg)
	}

	if !isFirstTimeIntegrationDone {
		factory := DD.GetDefaultDataCustomKPIFactory(integrationString)
		statusCode2 := factory.Build(projectID)
		if statusCode2 != http.StatusOK {
			_, errMsg := fmt.Printf("Failed during prebuilt custom KPI creation for: %v, %v", projectID, integrationString)
			logCtx.WithField("projectID", projectID).WithField("integration", integrationString).Warn(errMsg)
			return "", http.StatusInternalServerError, "", "", false
		} else {
			statusCode3 := DD.SetFirstTimeIntegrationDone(projectID, integrationString)
			if statusCode3 != http.StatusOK {
				_, errMsg := fmt.Printf("Failed during setting first time integration done marketo: %v", projectID)
				logCtx.WithField("projectID", projectID).WithField("integration", integrationString).Warn(errMsg.Error())
				return "", http.StatusInternalServerError, "", "", false
			}
		}
	}
	return "", http.StatusOK, "", "", false
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

	for _, objectType := range model.CustomKPIProfileObjectCategories {
		currentObjectTypeAndProperties := model.CustomMetricObjectTypeAndProperties{}
		currentObjectTypeAndProperties.ObjectType = objectType
		currentObjectTypeAndProperties.Properties = getPropertiesFunctionBasedOnObjectType(objectType)(projectID, reqID)

		customMetricObjectTypesAndProperties = append(customMetricObjectTypesAndProperties, currentObjectTypeAndProperties)
	}

	customMetricConfigs.AggregateFunctions = model.CustomMetricProfilesAggregateFunctions
	customMetricConfigs.ObjectTypeAndProperties = customMetricObjectTypesAndProperties
	c.JSON(http.StatusOK, gin.H{"result": customMetricConfigs})
}

// GetCustomMetricsConfig godoc
// @Summary To get config for the building custom metrics on settings.
// @Tags CustomMetric
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "{"result": model.CustomMetricConfig"
// @Router /{project_id}/v1/custom_metrics/config/v1 [get]
func GetCustomMetricsConfigV1(c *gin.Context) {
	reqID, projectID := getReqIDAndProjectID(c)
	finalConfigs := make([]model.CustomMetricConfigV1, 0)

	for _, objectType := range model.CustomKPIProfileObjectCategories {
		// CustomMetricConfigObjectV1
		currentConfigV1 := model.CustomMetricConfigV1{}
		currentConfigV1.ObjectType = objectType
		currentConfigV1.TypeOfQuery = model.ProfileQueryType
		currentConfigV1.TypeOfQueryDisplayName = model.ProfileQueryTypeDisplayName
		currentConfigV1.AggregateFunctions = model.CustomMetricProfilesAggregateFunctions
		currentConfigV1.Properties = getPropertiesFunctionBasedOnObjectType(objectType)(projectID, reqID)

		finalConfigs = append(finalConfigs, currentConfigV1)
	}

	customEventConfig := model.CustomMetricConfigV1{}
	customEventConfig.TypeOfQuery = model.EventBasedQueryType
	customEventConfig.TypeOfQueryDisplayName = model.EventBasedQueryTypeDisplayName
	customEventConfig.AggregateFunctions = model.CustomEventsAggregateFunctions

	derivedConfig := model.CustomMetricConfigV1{}
	derivedConfig.TypeOfQueryDisplayName = model.DerivedQueryTypeDisplayName
	derivedConfig.TypeOfQuery = model.DerivedQueryType

	finalConfigs = append(finalConfigs, customEventConfig, derivedConfig)

	c.JSON(http.StatusOK, gin.H{"result": finalConfigs})
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

// DeleteCustomMetrics godoc
// @Summary To delete custom metrics for a project id.
// @Tags CustomMetric
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param name path integer true "id"
// @Router /{project_id}/v1/custom_metrics/{id} [delete]
func DeleteCustomMetrics(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	customMetricsID := c.Params.ByName("id")

	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	customMetric, _, statusCode := store.GetStore().GetCustomMetricsByID(projectID, customMetricsID)
	if statusCode != http.StatusFound {
		return nil, http.StatusInternalServerError, INVALID_INPUT, ErrorMessages[INVALID_INPUT], true
	}

	dashboardUnitNames, statusCode := store.GetStore().GetDashboardUnitNamesByProjectIdTypeAndName(projectID, reqID, model.QueryClassKPI, customMetric.Name)
	if statusCode != http.StatusFound {
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error Processing/Fetching GetDashboardUnitNamesByProjectIdTypeAndName", true
	}

	alertNames, statusCode := store.GetStore().GetAlertNamesByProjectIdTypeAndName(projectID, customMetric.Name)
	if statusCode != http.StatusFound {
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error Processing/Fetching GetAlertNamesByProjectIdTypeAndName", true
	}

	customMetricNames := make([]string, 0)
	if customMetric.TypeOfQuery == model.ProfileQueryType {
		customMetricNames = store.GetStore().GetDerivedKPIsHavingNameInInternalQueries(projectID, customMetric.Name)
	}

	if len(dashboardUnitNames) == 0 && len(alertNames) == 0 {
		statusCode = store.GetStore().DeleteCustomMetricByID(projectID, customMetricsID)
		if statusCode != http.StatusAccepted {
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "", true
		} else {
			return nil, http.StatusOK, "", "", false
		}
	} else {
		errorMessage := "This Custom KPI is part of \""
		IsPrevious := false
		if len(dashboardUnitNames) > 0 {
			errorMessage = errorMessage + strings.Join(dashboardUnitNames, "\", \"") + "\" dashboard unit"
			if len(dashboardUnitNames) > 1 {
				errorMessage = errorMessage + "s"
			}
			IsPrevious = true
		}
		if len(alertNames) > 0 {
			if IsPrevious {
				errorMessage = errorMessage + " and \""
			}
			errorMessage = errorMessage + strings.Join(alertNames, "\", \"") + "\" alert"
			if len(alertNames) > 1 {
				errorMessage = errorMessage + "s"
			}
			IsPrevious = true
		}
		if len(customMetricNames) > 0 {
			if IsPrevious {
				errorMessage = errorMessage + " and "
			}
			errorMessage = errorMessage + strings.Join(customMetricNames, "\", \"") + "\" derived KPI"
			if len(customMetricNames) > 1 {
				errorMessage = errorMessage + "s"
			}
		}

		pronoun := "it"
		if (len(dashboardUnitNames) + len(alertNames) + len(customMetricNames)) > 1 {
			pronoun = "them"
		}
		errorMessage = errorMessage + ". Please remove " + pronoun + " first."
		return nil, http.StatusBadRequest, DEPENDENT_RECORD_PRESENT, errorMessage, true
	}
}

func getPropertiesFunctionBasedOnObjectType(objectType string) func(int64, string) []map[string]string {
	if model.GetGroupNameByMetricObjectType(objectType) == model.GROUP_NAME_HUBSPOT_COMPANY {
		return store.GetStore().GetPropertiesForHubspotCompanies
	} else if model.GetGroupNameByMetricObjectType(objectType) == model.GROUP_NAME_HUBSPOT_DEAL {
		return store.GetStore().GetPropertiesForHubspotDeals
	} else if strings.Contains(objectType, U.CRM_SOURCE_NAME_HUBSPOT) {
		return store.GetStore().GetPropertiesForHubspotContacts
	} else if model.GetGroupNameByMetricObjectType(objectType) == model.GROUP_NAME_SALESFORCE_OPPORTUNITY {
		return store.GetStore().GetPropertiesForSalesforceOpportunities
	} else if model.GetGroupNameByMetricObjectType(objectType) == model.GROUP_NAME_SALESFORCE_ACCOUNT {
		return store.GetStore().GetPropertiesForSalesforceAccounts
	} else if strings.Contains(objectType, U.CRM_SOURCE_NAME_SALESFORCE) {
		return store.GetStore().GetPropertiesForSalesforceUsers
	} else if strings.Contains(objectType, U.CRM_SOURCE_NAME_MARKETO) {
		return store.GetStore().GetPropertiesForMarketo
	} else if strings.Contains(objectType, U.CRM_SOURCE_NAME_LEADSQUARED) {
		return store.GetStore().GetPropertiesForLeadSquared
	}
	return nil
}
