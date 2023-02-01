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
		logCtx.Warnf("Decode failed on request to Custom Metrics struct. %v", requestAsMap)
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
// @Router /{project_id}/v1/custom_metrics/config/v1 [get]
func GetCustomMetricsConfigV1(c *gin.Context) {
	reqID, projectID := getReqIDAndProjectID(c)
	finalConfigs := make([]model.CustomMetricConfigV1, 0)

	for _, sectionDisplayCategory := range model.CustomKPIProfileSectionDisplayCategories {
		// CustomMetricConfigObjectV1
		currentConfigV1 := model.CustomMetricConfigV1{}
		currentConfigV1.SectionDisplayCategory = sectionDisplayCategory
		currentConfigV1.ObjectType = currentConfigV1.SectionDisplayCategory
		currentConfigV1.TypeOfQuery = model.ProfileQueryType
		currentConfigV1.TypeOfQueryDisplayName = model.ProfileQueryTypeDisplayName
		currentConfigV1.AggregateFunctions = model.CustomMetricProfilesAggregateFunctions
		currentConfigV1.Properties = getPropertiesFunctionBasedOnSectionDisplayCategory(sectionDisplayCategory)(projectID, reqID)

		finalConfigs = append(finalConfigs, currentConfigV1)
	}

	customEventConfig := model.CustomMetricConfigV1{}
	customEventConfig.SectionDisplayCategory = model.EventsBasedDisplayCategory
	customEventConfig.ObjectType = customEventConfig.SectionDisplayCategory
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
		logCtx.Warn("GetCustomMetrics - Custom metrics projectID is not given.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Invalid project id provided.", true
	}
	customMetrics, errMsg, statusCode := store.GetStore().GetCustomMetricsByProjectId(projectID)
	if statusCode != http.StatusFound {
		logCtx.WithField("message finder", "Failed to get custom metrics").Warn(errMsg)
		return nil, statusCode, PROCESSING_FAILED, "Failed to get custom metrics", true
	}
	return customMetrics, http.StatusOK, "", "", false
}

// GetPropertiesForCustomKPIEventBased Test command.
// curl -i -X GET http://localhost:8080/v1/:project_id/kpi/:custom_event_kpi/properties
// GetPropertiesForCustomKPIEventBased godoc
// @Summary To get properties for a given custom kpi event based.
// @Tags CustomMetric
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param custom_event_kpi path string true "Custom KPI event based"
// @Success 200 {string} json "map[string]string"
// @Router v1/{project_id}/kpi/{custom_event_kpi}/properties [get]
func GetPropertiesForCustomKPIEventBased(c *gin.Context) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})

	customKPIEventBased := c.Params.ByName("custom_event_kpi")
	if customKPIEventBased == "" {
		logCtx.WithField("custom_event_kpi", customKPIEventBased).Error("null custom_event_kpi")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	customKPI, _, statusCode := store.GetStore().GetEventBasedCustomMetricByProjectIdName(projectID, customKPIEventBased)
	if statusCode == http.StatusNotFound {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if statusCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	eventName, err := customKPI.GetEventName()
	if err != "" {
		noProperties := make([]map[string]string, 0)
		c.JSON(http.StatusOK, gin.H{"properties": noProperties})
		return
	}

	propertiesFromCache, statusCode := store.GetStore().GetEventNamesAndModifyResultsForNonExplain(projectID, eventName)
	if statusCode != http.StatusOK {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	displayNamesOp := store.GetStore().GetDisplayNamesForEventName(projectID, propertiesFromCache, eventName)

	kpiConfig := model.TransformEventPropertiesToKPIConfigProperties(propertiesFromCache, displayNamesOp)
	// Handling both error and NotFound.
	if statusCode != http.StatusFound {
		c.JSON(http.StatusOK, gin.H{"properties": kpiConfig})
		return
	}
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

	derivedKPINames := make([]string, 0)
	if customMetric.TypeOfQuery != model.DerivedQueryType {
		derivedKPINames = store.GetStore().GetDerivedKPIsHavingNameInInternalQueries(projectID, customMetric.Name)
	}

	if len(dashboardUnitNames) == 0 && len(alertNames) == 0 && len(derivedKPINames) == 0 {
		statusCode = store.GetStore().DeleteCustomMetricByID(projectID, customMetricsID)
		if statusCode != http.StatusAccepted {
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "", true
		} else {
			return nil, http.StatusOK, "", "", false
		}
	} else {
		errorMessage := BuildDependentsErrorMessage("Custom KPI", [][]string{dashboardUnitNames, alertNames, derivedKPINames}, []string{"dashboard unit", "alert", "derived KPI"})
		return nil, http.StatusBadRequest, DEPENDENT_RECORD_PRESENT, errorMessage, true
	}
}

func getPropertiesFunctionBasedOnSectionDisplayCategory(sectionDisplayCategory string) func(int64, string) []map[string]string {
	if model.GetGroupNameByMetricSectionDisplayCategory(sectionDisplayCategory) == model.GROUP_NAME_HUBSPOT_COMPANY {
		return store.GetStore().GetPropertiesForHubspotCompanies
	} else if model.GetGroupNameByMetricSectionDisplayCategory(sectionDisplayCategory) == model.GROUP_NAME_HUBSPOT_DEAL {
		return store.GetStore().GetPropertiesForHubspotDeals
	} else if strings.Contains(sectionDisplayCategory, U.CRM_SOURCE_NAME_HUBSPOT) {
		return store.GetStore().GetPropertiesForHubspotContacts
	} else if model.GetGroupNameByMetricSectionDisplayCategory(sectionDisplayCategory) == model.GROUP_NAME_SALESFORCE_OPPORTUNITY {
		return store.GetStore().GetPropertiesForSalesforceOpportunities
	} else if model.GetGroupNameByMetricSectionDisplayCategory(sectionDisplayCategory) == model.GROUP_NAME_SALESFORCE_ACCOUNT {
		return store.GetStore().GetPropertiesForSalesforceAccounts
	} else if strings.Contains(sectionDisplayCategory, U.CRM_SOURCE_NAME_SALESFORCE) {
		return store.GetStore().GetPropertiesForSalesforceUsers
	} else if strings.Contains(sectionDisplayCategory, U.CRM_SOURCE_NAME_MARKETO) {
		return store.GetStore().GetPropertiesForMarketo
	} else if strings.Contains(sectionDisplayCategory, U.CRM_SOURCE_NAME_LEADSQUARED) {
		return store.GetStore().GetPropertiesForLeadSquared
	}
	return nil
}
