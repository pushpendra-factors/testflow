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

type CommonPropertyMappingsRequest struct {
	Name       string `json:"name"`
	DerivedKPI bool   `json:"derived_kpi"`
}

// CreatePropertyMapping godoc.
// @Summary To create property mapping.
// @Tags PropertyMapping
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query body model.PropertyMapping true "Property mapping payload."
// @Success 200 {string} json "{"result": model.PropertyMapping}"
// @Router /{project_id}/kpi/property_mappings [post]
func CreatePropertyMapping(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)

	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}
	request := model.PropertyMapping{}
	err := c.BindJSON(&request)
	if err != nil {
		var requestAsMap map[string]interface{}
		c.BindJSON(&requestAsMap)
		logCtx.Warnf("Decode failed on request to property mapping struct. %v", requestAsMap)
		return nil, http.StatusBadRequest, INVALID_INPUT, "Error during decode of property mapping.", true
	}

	var properties []model.Property
	err = U.DecodePostgresJsonbToStructType(request.Properties, &properties)
	if err != nil {
		var requestAsMap map[string]interface{}
		c.BindJSON(&requestAsMap)
		logCtx.Warnf("Decode failed on request to property mapping properties struct. %v", requestAsMap)
		return nil, http.StatusBadRequest, INVALID_INPUT, "Error during decode of property mapping properties.", true
	}
	isValid, errMsg := request.IsValid(properties)
	if !isValid {
		return nil, http.StatusBadRequest, INVALID_INPUT, errMsg, true
	}

	request.ProjectID = projectID
	request.Name = U.CreatePropertyNameFromDisplayName(request.DisplayName)
	request.DataType = properties[0].DataType
	request.SectionBitMap, errMsg = model.GenerateSectionBitMapFromProperties(properties)
	if errMsg != "" {
		return nil, http.StatusBadRequest, INVALID_INPUT, errMsg, true
	}
	propertyMapping, errMsg, statusCode := store.GetStore().CreatePropertyMapping(request)
	if statusCode != http.StatusCreated {
		logCtx.WithField("message finder", "Error during property mapping creation").Warn(errMsg)
		if strings.Contains(errMsg, "Duplicate") {
			return nil, http.StatusConflict, DUPLICATE_RECORD, ErrorMessages[DUPLICATE_RECORD], true
		} else {
			return nil, statusCode, PROCESSING_FAILED, ErrorMessages[PROCESSING_FAILED], true
		}
	}
	return propertyMapping, http.StatusOK, "", "", false
}

// GetPropertyMappings godoc.
// @Summary To get property mappings.
// @Tags PropertyMapping
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "{"result": []model.PropertyMapping}"
// @Router /{project_id}/kpi/property_mappings [get]
func GetPropertyMappings(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)

	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}

	propertyMappings, errMsg, statusCode := store.GetStore().GetPropertyMappingsByProjectId(projectID)
	if statusCode != http.StatusOK {
		logCtx.WithField("message finder", "Error during property mapping creation").Warn(errMsg)
		return nil, statusCode, PROCESSING_FAILED, ErrorMessages[PROCESSING_FAILED], true
	}
	return propertyMappings, http.StatusOK, "", "", false
}

// GetCommonPropertyMappings godoc.
// @Summary To get property mappings with section bit map.
// @Tags PropertyMapping
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query body section list true "Section list."
// @Success 200 {string} json "{"result": []model.PropertyMapping}"
// @Router /{project_id}/kpi/property_mappings_list [post]
func GetCommonPropertyMappings(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)

	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}

	requests := []CommonPropertyMappingsRequest{}
	err := c.BindJSON(&requests)
	if err != nil {
		var requestAsMap map[string]interface{}
		c.BindJSON(&requestAsMap)
		logCtx.Warnf("Decode failed on request to string array. %v", requestAsMap)
		return nil, http.StatusBadRequest, INVALID_INPUT, "Error during decode.", true
	}

	// there could be doublicate displayCategories, but it dosen't matter for sectionBitMap generation
	displayCategories, errMsg := getDisplayCategoryFromCommonPropertyMappingsRequest(projectID, requests)
	if errMsg != "" {
		logCtx.Warnf("Error during display category extraction. %v", err)
		return nil, http.StatusBadRequest, INVALID_INPUT, "Error during display category extraction.", true
	}
	sectionBitMap, errMsg := model.GenerateSectionBitMap(displayCategories)
	if errMsg != "" {
		logCtx.Warnf("Error during section bit map creation. %v", err)
		return nil, http.StatusBadRequest, INVALID_INPUT, "Error during section bit map creation.", true
	}

	propertyMappings, errMsg, statusCode := store.GetStore().GetPropertyMappingsByProjectIdAndSectionBitMap(projectID, sectionBitMap)
	if statusCode != http.StatusOK {
		logCtx.WithField("message finder", "Error during property mapping creation").Warn(errMsg)
		return nil, statusCode, PROCESSING_FAILED, ErrorMessages[PROCESSING_FAILED], true
	}

	return getCommonPropertyMappingResponceMapFromPropertyMapping(propertyMappings), http.StatusOK, "", "", false
}

// Gets Display Categories from CommonPropertyMappingsRequest
// If DerivedKPI is true, then it gets the display categories from the derived_metric_transformations(KPIQueryGroup)
// Return array can contain duplicate values
func getDisplayCategoryFromCommonPropertyMappingsRequest(projectID int64, requests []CommonPropertyMappingsRequest) ([]string, string) {
	displayCategories := []string{}
	for _, req := range requests {
		if req.DerivedKPI {
			displayCategoriesInDerivedMetric, errMsg := store.GetStore().GetDisplayCategoriesByProjectIdAndNameFromDerivedCustomKPI(projectID, req.Name)
			if errMsg != "" {
				return displayCategories, errMsg
			}
			displayCategories = append(displayCategories, displayCategoriesInDerivedMetric...)
		} else {
			displayCategories = append(displayCategories, req.Name)
		}
	}
	return displayCategories, ""
}

// Gets results for common property mapping fetch from property mappings
// Only selected and required fields are returned
func getCommonPropertyMappingResponceMapFromPropertyMapping(propertyMappings []*model.PropertyMapping) []map[string]interface{} {
	finalResponse := []map[string]interface{}{}

	// Only required fields are sent in response(i.e. display_name, name, data_type)
	for _, propertyMapping := range propertyMappings {
		var responseMap = map[string]interface{}{
			"display_name": propertyMapping.DisplayName,
			"name":         propertyMapping.Name,
			"data_type":    propertyMapping.DataType,
		}
		finalResponse = append(finalResponse, responseMap)
	}
	return finalResponse
}

// DeletePropertyMapping godoc.
// @Summary To delete property mapping.
// @Tags PropertyMapping
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param name path integer true "id"
// @Router /{project_id}/kpi/property_mappings/{id} [delete]
func DeletePropertyMapping(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	propertyMappingId := c.Params.ByName("id")
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID).WithField("propertyMappingId", propertyMappingId)

	if projectID == 0 {
		logCtx.Warnf("Invalid project id.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}

	// Check if the property mapping with the given id exists and get it.
	propertyMapping, errMsg, statusCode := store.GetStore().GetPropertyMappingByID(projectID, propertyMappingId)
	if statusCode != http.StatusOK {
		logCtx.Error(errMsg)
		return nil, http.StatusInternalServerError, INVALID_INPUT, ErrorMessages[INVALID_INPUT], true
	}

	// Gets the dashboard unit names in which this property mapping is used for KPI Query.
	dashboardUnitNames, statusCode := store.GetStore().GetDashboardUnitNamesByProjectIdTypeAndPropertyMappingName(projectID, reqID, propertyMapping.Name)
	if statusCode == http.StatusInternalServerError {
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error Processing/Fetching GetDashboardUnitNamesByProjectIdTypeAndName", true
	}

	// Gets the alert names in which this property mapping is used for KPI Query.
	alertNames, statusCode := store.GetStore().GetAlertNamesByProjectIdTypeAndNameAndPropertyMappingName(projectID, reqID, propertyMapping.Name)
	if statusCode == http.StatusInternalServerError {
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error Processing/Fetching GetAlertNamesByProjectIdTypeAndName", true
	}

	// If the property mapping is not used in any dashboard unit, then delete it.
	if len(dashboardUnitNames) == 0 && len(alertNames) == 0 {
		statusCode = store.GetStore().DeletePropertyMappingByID(projectID, propertyMappingId)
		if statusCode != http.StatusOK {
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "", true
		} else {
			return nil, http.StatusOK, "", "", false
		}
	} else {
		errorMessage := BuildDependentsErrorMessage("Property Mapping", [][]string{dashboardUnitNames, alertNames}, []string{"dashboard unit", "alert"})
		return nil, http.StatusBadRequest, DEPENDENT_RECORD_PRESENT, errorMessage, true
	}
}
