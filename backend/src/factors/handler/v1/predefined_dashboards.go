package v1

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type PredefinedDashboardFilterValueRequest struct {
	PropertyName string `json:"pr_na"`
}

// GetPredefinedDashboardsHandler godoc
// @Summary To get predefined dashboard config.
// @Tags Predefined
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param internal_id path integer true "Internal Predefined Dashboard ID"
// @Success 200 {string} json "{"result": []interface{}"
// @Router /{project_id}/v1/predefined_dashboards/{internal_id}/config [get]
func GetPredefinedDashboardConfigsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	// reqID, projectID := getReqIDAndProjectID(c)

	dashboardInternalId, err := strconv.ParseInt(c.Params.ByName("internal_id"), 10, 64)
	if err != nil || dashboardInternalId == 0 {
		log.WithError(err).Error("Invalid internal id.")
		return gin.H{"error": "Invalid internal predefined dashboard id."}, http.StatusBadRequest, "", "", false
	}

	dashboardConfig, exists := model.MapfOfPredefinedDashboardIDToConfig[dashboardInternalId]
	if exists {
		return gin.H{"result": dashboardConfig}, http.StatusOK, "", "", false
	}

	return gin.H{"error": "Invalid internal predefined dashboard id."}, http.StatusBadRequest, "", "", false
}

// GetPredefinedDashboardFilterValues godoc
// @Summary To get predefined dashboard filter values.
// @Tags Predefined
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param internal_id path integer true "Internal Predefined Dashboard ID"
// @Param query body PredefinedDashboardFilterValueRequest true "Filter Values payload"
// @Success 200 {string} json "{"result": map[string][]string"
// @Router /{project_id}/v1/predefined_dashboards/{internal_id}/filter_values [POST]
func GetPredefinedDashboardFilterValues(c *gin.Context) (interface{}, int, string, string, bool) {
	var resultantFilterValuesResponse []string

	reqID, projectID := getReqIDAndProjectID(c)
	logCtx := log.WithField("projectID", projectID).WithField("req_id", reqID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}

	dashboardInternalId, err := strconv.ParseInt(c.Params.ByName("internal_id"), 10, 64)
	if err != nil || dashboardInternalId == 0 {
		log.WithError(err).Error("Invalid internal id.")
		return gin.H{"error": "Invalid internal predefined dashboard id."}, http.StatusBadRequest, "", "", false
	}

	request := PredefinedDashboardFilterValueRequest{}
	err1 := c.BindJSON(&request)
	if err1 != nil {
		logCtx.Warn("Error during decode of predefined filter values")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Error during decode of predefined filter values.", true
	}

	logCtx = logCtx.WithField("request", request)

	PredefinedDashboardProperties, exists1 := model.MapOfPredefinedDashboardToPropertyNameToProperties[dashboardInternalId]
	if !exists1 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "Invalid internal predefined dashboard Id.", true
	}

	predefinedDashboardProperty, exists2 := PredefinedDashboardProperties[request.PropertyName]
	if !exists2 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "Invalid filter property Name.", true
	}

	storeSelected := store.GetStore()
	if predefinedDashboardProperty.SourceEntity == model.UserEntity {
		userFilterValues, err := storeSelected.GetPropertyValuesByUserProperty(projectID, predefinedDashboardProperty.SourceProperty, model.FilterValuesOrEventNamesLimit, C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.Warn(err)
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of Predefined user FilterValues Data.", true
		}
		resultantFilterValuesResponse = userFilterValues
	} else {
		eventFilterValues, err := storeSelected.GetPropertyValuesByEventProperty(projectID, predefinedDashboardProperty.SourceEventName, predefinedDashboardProperty.SourceProperty, model.FilterValuesOrEventNamesLimit, C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.Warn(err)
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of Predefined user FilterValues Data.", true
		}
		resultantFilterValuesResponse = eventFilterValues
	}

	propertyValueLabel, err, isSourceEmpty := store.GetStore().GetPropertyValueLabel(projectID, predefinedDashboardProperty.SourceProperty, resultantFilterValuesResponse)
	if err != nil {
		logCtx.WithError(err).Error("get event properties labels and values by property name")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of Predefined PropertyValue Data.", true
	}

	if isSourceEmpty {
		logCtx.WithField("property_name", request.PropertyName).Warning("source is empty")
	}

	if len(propertyValueLabel) == 0 {
		logCtx.WithField("property_name", request.PropertyName).Error("No event property value labels Returned")
	}

	return U.FilterDisplayNameEmptyKeysAndValues(projectID, propertyValueLabel), http.StatusOK, "", "", false
}

// TODO - clarity on GBT, non GBT.
// Kind of transformations happen on results.
// Total part of computation.

// ExecutePredefinedQueryHandler godoc
// @Summary To run a predefined dashboard query.
// @Tags Predefined
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query body model.PredefinedQuery true "Query payload"
// @Success 200 {string} json "{result:[]model.QueryResult}"
// @Router /{project_id}/v1/predefined_dashboards/{internal_id}/query [post]
func ExecutePredefinedQueryHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	logCtx := log.WithField("projectID", projectID).WithField("req_id", reqID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}

	dashboardInternalId, err := strconv.ParseInt(c.Params.ByName("internal_id"), 10, 64)
	if err != nil || dashboardInternalId == 0 {
		log.WithError(err).Error("Invalid internal id.")
		return gin.H{"error": "Invalid internal predefined dashboard id."}, http.StatusBadRequest, "", "", false
	}

	request, err1 := getPredefinedDashboard(c, dashboardInternalId)
	if err1 != "" {
		logCtx.Warn("Error during decode of predefined execute query")
		return nil, http.StatusBadRequest, INVALID_INPUT, err1, true
	}

	isValid, err2 := request.IsValid()
	if !isValid {
		return nil, http.StatusBadRequest, INVALID_INPUT, err2, true
	}

	resResults, statusCode, errMsg := executePredefinedDashboard(projectID, dashboardInternalId, request)
	if errMsg != "" {
		return gin.H{"result": resResults}, statusCode, PROCESSING_FAILED, errMsg, true
	}
	return gin.H{"result": resResults}, http.StatusOK, "", "", false

}

func getPredefinedDashboard(c *gin.Context, dashboardInternalID int64) (model.PredefinedQueryGroup, string) {
	if dashboardInternalID == 1 {
		request := model.PredefWebsiteAggregationQueryGroup{}
		err := c.BindJSON(&request)
		if err != nil {
			// logCtx.Warn("Error during decode of predefined filter values")
			// return nil, http.StatusBadRequest, INVALID_INPUT, "Error during decode of predefined filter values.", true
			return nil, "Error during decode of predefined website aggregation."
		}

		return request, ""
	} else {
		return nil, "Input Dashboard ID is wrong."
	}
}

func executePredefinedDashboard(projectID, dashboardInternalId int64, request model.PredefinedQueryGroup) ([]model.QueryResult, int, string) {

	storeSelected := store.GetStore()
	if dashboardInternalId == 1 {
		queryGroup := request.(model.PredefWebsiteAggregationQueryGroup)
		return storeSelected.ExecuteQueryGroupForPredefinedWebsiteAggregation(projectID, queryGroup)
	}
	return make([]model.QueryResult, 0), http.StatusBadRequest, "Invalid dashboard internal ID"
}
