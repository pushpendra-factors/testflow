package v1

import (
	"errors"
	C "factors/config"
	H "factors/handler/helpers"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type KPIFilterValuesRequest struct {
	Category          string `json:"category"`
	DisplayCategory   string `json:"display_category"`
	ObjectType        string `json:"object_type"` // used for channels ex. object_type = campaign and for events. ex. object_type = event_name
	PropertyName      string `json:"property_name"`
	Entity            string `json:"entity"`
	Metric            string `json:"me"`
	IsPropertyMapping bool   `json:"is_property_mapping"` // true if requested property_name is a property mapping
}

func (req *KPIFilterValuesRequest) isValid() bool {
	if req == nil {
		return false
	}
	if !req.IsPropertyMapping && (req.Category == "" || !U.ContainsStringInArray([]string{model.ChannelCategory, model.CustomChannelCategory, model.EventCategory, model.ProfileCategory}, req.Category) ||
		req.Entity == "" || !U.ContainsStringInArray([]string{model.EventEntity, model.UserEntity}, req.Entity) ||
		req.ObjectType == "") || req.PropertyName == "" {
		return false
	}
	return true
}

func getReqIDAndProjectID(c *gin.Context) (string, int64) {
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	return reqID, projectID
}

// GetKPIConfigHandler godoc
// @Summary To get config for the required kpi.
// @Tags KPIQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "{"result": map[string]interface{}"
// @Router /{project_id}/v1/kpi/config [get]
func GetKPIConfigHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}

	includeDerivedKPIsString := c.Query("include_derived_kpis")
	log.WithField("includeDerivedKPIsString", includeDerivedKPIsString).Warn("QueryParam")
	includeDerivedKPIs := false

	if includeDerivedKPIsString != "" {
		includeDerivedKPIs, _ = strconv.ParseBool(includeDerivedKPIsString)
	}

	storeSelected := store.GetStore()
	resultantResultConfigFunctions := make([]func(int64, string, bool) (map[string]interface{}, int), 0)
	resultantConfigs := make([]map[string]interface{}, 0)
	configForStaticSubSectionsFunctions := []func(int64, string, bool) (map[string]interface{}, int){
		storeSelected.GetKPIConfigsForWebsiteSessions,
		storeSelected.GetKPIConfigsForFormSubmissions,
		storeSelected.GetKPIConfigsForHubspotContacts,
		storeSelected.GetKPIConfigsForHubspotCompanies,
		storeSelected.GetKPIConfigsForHubspotDeals,
		storeSelected.GetKPIConfigsForSalesforceUsers,
		storeSelected.GetKPIConfigsForSalesforceAccounts,
		storeSelected.GetKPIConfigsForSalesforceOpportunities,
		storeSelected.GetKPIConfigsForAdwords, storeSelected.GetKPIConfigsForGoogleOrganic,
		storeSelected.GetKPIConfigsForFacebook, storeSelected.GetKPIConfigsForLinkedin,
		storeSelected.GetKPIConfigsForLinkedinCompanyEngagements,
		storeSelected.GetKPIConfigsForAllChannels, storeSelected.GetKPIConfigsForBingAds, storeSelected.GetKPIConfigsForMarketoLeads,
		storeSelected.GetKPIConfigsForLeadSquaredLeads,
	}
	configFunctionsForCustomAds := []func(int64, string, bool) ([]map[string]interface{}, int){
		storeSelected.GetKPIConfigsForCustomAds,
	}

	resultantResultConfigFunctions = append(resultantResultConfigFunctions, configForStaticSubSectionsFunctions...)

	if includeDerivedKPIs {
		resultantResultConfigFunctions = append(resultantResultConfigFunctions, storeSelected.GetKPIConfigsForPageViews)
	}

	for _, configFunction := range resultantResultConfigFunctions {
		currentConfig, errCode := configFunction(projectID, reqID, includeDerivedKPIs)
		if errCode != http.StatusOK {
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of KPI Config Data.", true
		}
		if currentConfig != nil {
			resultantConfigs = append(resultantConfigs, currentConfig)
		}
	}

	for _, configFunction := range configFunctionsForCustomAds {
		currentConfigs, errCode := configFunction(projectID, reqID, includeDerivedKPIs)
		if errCode != http.StatusOK {
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of KPI Custom Config Data.", true
		}
		if len(currentConfigs) > 0 {
			resultantConfigs = append(resultantConfigs, currentConfigs...)
		}
	}

	// for custom events
	currentConfig, errCode := storeSelected.GetKPIConfigsForCustomEvents(projectID, model.EventsBasedDisplayCategory, includeDerivedKPIs)
	if errCode != http.StatusOK {
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of KPI Custom Config Data.", true
	}
	if currentConfig != nil {
		resultantConfigs = append(resultantConfigs, currentConfig)
	}

	if includeDerivedKPIs {
		currentConfig, errCode := storeSelected.GetKPIConfigsForOthers(projectID, model.OthersDisplayCategory, includeDerivedKPIs)
		if errCode != http.StatusOK {
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of KPI Custom Config Data.", true
		}
		if currentConfig != nil {
			resultantConfigs = append(resultantConfigs, currentConfig)
		}
	}

	return resultantConfigs, http.StatusOK, "", "", false
}

// GetKPIFilterValuesHandler godoc
// @Summary To filter on values for kpi query.
// @Tags KPIanalytics, KPIQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query body model.KPIFilterValuesRequest true "Filter Values payload"
// @Success 200 {string} json "{"result": interface{}}"
// @Router /{project_id}/v1/kpi/filter_values [get]
func GetKPIFilterValuesHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	logCtx := log.WithField("projectID", projectID).WithField("req_id", reqID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}
	request := KPIFilterValuesRequest{}
	err := c.BindJSON(&request)
	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, "Error during validation of KPI FilterValues Data.", true
	}
	if !request.isValid() {
		logCtx.Warn(err)
		return nil, http.StatusInternalServerError, INVALID_INPUT, "Error during validation of KPI FilterValues Data.", true
	}

	var kpiFilterValues interface{}
	var status int
	var errString, errMsg string
	var failures bool
	// If property mapping is requested, then get the union of individual property values
	if request.IsPropertyMapping {
		kpiFilterValues, status, errString, errMsg, failures = getKpiFilterValuesForPropertyMapping(request, projectID, reqID)
	} else {
		kpiFilterValues, status, errString, errMsg, failures = getKpiFilterValuesForStaticProperty(request, projectID, reqID)
	}

	if failures {
		return kpiFilterValues, status, errString, errMsg, failures
	}

	label := c.Query("label")
	if label != "true" {
		return kpiFilterValues, http.StatusOK, "", "", false
	}

	propertyValueLabel := getPropertyValueLabel(projectID, request.PropertyName, kpiFilterValues)

	return propertyValueLabel, http.StatusOK, "", "", false
}

func getPropertyValueLabel(projectID int64, propertyName string, filterValues interface{}) map[string]string {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "property_name": propertyName, "filter_values": filterValues})

	propertyValues := []string{}
	switch values := filterValues.(type) {
	case []string:
		propertyValues = values

	case []interface{}:
		for i := range values {
			propertyValues = append(propertyValues, U.GetPropertyValueAsString(values[i]))
		}

	default:
		logCtx.Error("Failed to convert interface on KPI getPropertyValueLabel.")
		return nil
	}

	filterValueMap := make(map[string]string)
	for _, value := range propertyValues {
		filterValueMap[value] = value
	}

	var source string
	if U.IsAllowedCRMPropertyPrefix(propertyName) {
		source = strings.Split(propertyName, "_")[0]
		source = strings.TrimPrefix(source, "$")
	}

	if source == "" || propertyName == "" {
		return filterValueMap
	}

	propertyValueLabelMap, err := store.GetStore().GetPropertyLabelAndValuesByProjectIdAndPropertyKey(projectID, source, propertyName)
	if err != nil {
		logCtx.Error("Failed to GetPropertyLabelAndValuesByProjectIdAndPropertyKey in KPI filter values.")
		return filterValueMap
	}

	for value := range filterValueMap {
		if label, exists := propertyValueLabelMap[value]; exists && label != "" {
			filterValueMap[value] = label
		}
	}

	return filterValueMap
}

// Gets filter values for property mapping request as union of individual property values
func getKpiFilterValuesForPropertyMapping(request KPIFilterValuesRequest, projectID int64, reqID string) (interface{}, int, string, string, bool) {

	storeSelected := store.GetStore()
	logCtx := log.WithField("projectID", projectID).WithField("req_id", reqID).WithField("request", request)

	propertyMappings, errMsg, statusCode := storeSelected.GetPropertyMappingByProjectIDAndName(projectID, request.PropertyName)
	if statusCode != http.StatusOK {
		return nil, statusCode, PROCESSING_FAILED, errMsg, true
	}

	var properties []model.Property
	err := U.DecodePostgresJsonbToStructType(propertyMappings.Properties, &properties)
	if err != nil {
		logCtx.WithError(err).Error("Error during decode of property_mapping property")
		return nil, statusCode, PROCESSING_FAILED, "Error during decode of property_mapping property - KPIFilterValues handler.", true
	}

	// set to avoid duplicate values (to make a union of values)
	filterValuesSet := map[interface{}]struct{}{}
	for _, property := range properties {
		request := KPIFilterValuesRequest{
			Category:        property.Category,
			DisplayCategory: property.DisplayCategory,
			ObjectType:      property.ObjectType,
			PropertyName:    property.Name,
			Entity:          property.Entity,
		}
		filterValues, statusCode, err, errMsg, isError := getKpiFilterValuesForStaticProperty(request, projectID, reqID)
		if isError {
			logCtx.Error("Error during fetch of KPI Filter Values Data.")
			return nil, statusCode, err, errMsg, true
		}
		// as channel and custom channel returns []interface{} and others return []string
		if request.Category == model.ChannelCategory || request.Category == model.CustomChannelCategory {
			for _, filterValue := range filterValues.([]interface{}) {
				filterValuesSet[filterValue] = struct{}{}
			}
		} else {
			for _, filterValue := range filterValues.([]string) {
				filterValuesSet[filterValue] = struct{}{}
			}
		}
	}
	resultantFilterValuesResponse := make([]interface{}, 0, len(filterValuesSet))
	for filterValue := range filterValuesSet {
		resultantFilterValuesResponse = append(resultantFilterValuesResponse, filterValue)
	}
	return resultantFilterValuesResponse, http.StatusOK, "", "", false
}

// Gets filter values for static property request
func getKpiFilterValuesForStaticProperty(request KPIFilterValuesRequest, projectID int64, reqID string) (interface{}, int, string, string, bool) {
	storeSelected := store.GetStore()
	var resultantFilterValuesResponse interface{}

	logCtx := log.WithField("projectID", projectID).WithField("req_id", reqID).WithField("request", request)
	if request.Category == model.ChannelCategory {
		currentChannel, err := model.GetChannelFromKPIQuery(request.DisplayCategory, request.Category)
		if err != nil {
			return nil, http.StatusBadRequest, PROCESSING_FAILED, "Input display category is wrong", true
		}
		request.DisplayCategory = currentChannel
		channelsFilterValues, errCode := storeSelected.GetChannelFilterValuesV1(projectID, request.DisplayCategory, request.ObjectType,
			strings.TrimPrefix(request.PropertyName, request.ObjectType+"_"), reqID)
		if errCode != http.StatusOK && errCode != http.StatusFound {
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of KPI FilterValues Data.", true
		}
		resultantFilterValuesResponse = channelsFilterValues.FilterValues
	} else if request.Category == model.CustomChannelCategory {
		currentChannel, err := model.GetCustomChannelFromKPIQuery()
		if err != nil {
			return nil, http.StatusBadRequest, PROCESSING_FAILED, "Input display category is wrong", true
		}
		channelsFilterValues, errCode := storeSelected.GetCustomChannelFilterValuesV1(projectID, request.DisplayCategory, currentChannel, request.ObjectType,
			strings.TrimPrefix(request.PropertyName, request.ObjectType+"_"), reqID)
		if errCode != http.StatusOK && errCode != http.StatusFound {
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of KPI FilterValues Data.", true
		}
		resultantFilterValuesResponse = channelsFilterValues.FilterValues
	} else if request.Category == model.EventCategory && request.Entity == model.EventEntity {
		// For both static and custom event metrics, same method of fetching filter values is used.
		request.ObjectType = model.GetObjectTypeForFilterValues(request.DisplayCategory, request.Metric)
		eventsFilterValues, err := storeSelected.GetPropertyValuesByEventProperty(projectID, request.ObjectType, request.PropertyName, model.FilterValuesOrEventNamesLimit, C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.Warn(err)
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of KPI FilterValues Data.", true
		}
		resultantFilterValuesResponse = eventsFilterValues
	} else if groupName, isGroupProperty := U.GetGroupNameByPropertyName(request.PropertyName); isGroupProperty {
		groupFilterValues, err := storeSelected.GetPropertyValuesByGroupProperty(projectID, groupName, request.PropertyName, 2500, C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.Warn(err)
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of KPI FilterValues Data.", true
		}
		resultantFilterValuesResponse = groupFilterValues
	} else {
		userFilterValues, err := storeSelected.GetPropertyValuesByUserProperty(projectID, request.PropertyName, model.FilterValuesOrEventNamesLimit, C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.Warn(err)
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of KPI FilterValues Data.", true
		}
		resultantFilterValuesResponse = userFilterValues
	}
	return resultantFilterValuesResponse, http.StatusOK, "", "", false
}

// ExecuteKPIQueryHandler godoc
// @Summary To run a channel query.
// @Tags KPIanalytics, KPIQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id query integer false "Dashboard ID"
// @Param dashboard_unit_id query integer false "Dashboard Unit ID"
// @Param query body model.KPIQueryGroup true "Query payload"
// @Success 200 {string} json "{result:[]model.QueryResult}"
// @Router /{project_id}/v1/channels/query [post]
func ExecuteKPIQueryHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	logCtx := log.WithField("reqId", reqID)
	var timezoneString U.TimeZoneString
	var statusCode int
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}
	logCtx = logCtx.WithField("project_id", projectID).WithField("reqId", reqID)

	request, dashboardId, unitId, isDashboardQueryRequest, isDashboardQueryLocked, statusCode, errorCode, errMsg, isErr := getValidKPIQueryAndDetailsFromRequest(c.Request, c, logCtx, projectID, c.Query("dashboard_id"), c.Query("dashboard_unit_id"), c.Query("query_id"))
	if statusCode != http.StatusOK {
		return nil, statusCode, errorCode, errMsg, isErr
	}

	preset, commonQueryFrom, commonQueryTo, hardRefresh, _ := getOtherRelatedInformationFromRequest(request, projectID,
		c.Query("preset"), c.Query("refresh"), c.Query("is_query"))

	request, statusCode, errorCode, errMsg, isErr = setTimezoneForKPIRequest(logCtx, request, projectID)
	if statusCode != http.StatusOK {
		return nil, statusCode, errorCode, errMsg, isErr
	}

	err := request.TransformDateTypeFilters()
	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, err.Error(), true
	}

	allowSyncReferenceFields := C.AllowSyncReferenceFields(projectID)
	// Use JSON optimised filter for profiles query from KPI if enabled using query_param or global config.
	enableOptimisedFilterOnProfileQuery := c.Request.Header.Get(H.HeaderUserFilterOptForProfiles) == "true" ||
		C.EnableOptimisedFilterOnProfileQuery()

	enableOptimisedFilterOnEventUserQuery := c.Request.Header.Get(H.HeaderUserFilterOptForEventsAndUsers) == "true" ||
		C.EnableOptimisedFilterOnEventUserQuery()

	// Tracking dashboard query request.
	if isDashboardQueryRequest {
		if preset == "" && C.IsLastComputedWhitelisted(projectID) {
			preset = U.GetPresetNameByFromAndTo(commonQueryFrom, commonQueryTo, timezoneString)
		}
		model.SetDashboardCacheAnalytics(projectID, dashboardId, unitId, commonQueryFrom, commonQueryTo, timezoneString)
	}

	if isDashboardQueryLocked {
		var duplicatedRequest model.KPIQueryGroup
		U.DeepCopy(&request, &duplicatedRequest)
		queryResult, statusCode := store.GetStore().ExecuteKPIQueryGroup(projectID, reqID,
			duplicatedRequest, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		if statusCode != http.StatusOK {
			if statusCode == http.StatusPartialContent {
				return queryResult, statusCode, PROCESSING_FAILED, "Failed to process query from DB", true
			}
			return nil, statusCode, PROCESSING_FAILED, "Failed to process query from DB", true
		}

		if allowSyncReferenceFields {
			queryResult, err = store.GetStore().AddPropertyValueLabelToQueryResults(projectID, queryResult)
			if err != nil {
				logCtx.WithError(err).Error("Failed to set property value label.")
			}
		}

		return H.DashboardQueryResponsePayload{Result: queryResult, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(), CacheMeta: nil}, http.StatusOK, "", "", false
	}
	if !hardRefresh { // to make condition explicit
		data, statusCode, errorCode, errMsg, isErr := GetResultFromCacheOrDashboard(c, reqID, projectID, request, dashboardId, unitId, preset, commonQueryFrom, commonQueryTo, hardRefresh, timezoneString, isDashboardQueryRequest, logCtx, false)
		if statusCode != http.StatusProcessing {
			if allowSyncReferenceFields && data != nil {
				data, err = H.TransformQueryCacheResponseColumnValuesToLabel(projectID, data)
				if err != nil {
					logCtx.WithError(err).Error("Failed to set property value label.")
				}
			}
			return data, statusCode, errorCode, errMsg, isErr
		}
	}

	/*if isDashboardQueryRequest && C.DisableDashboardQueryDBExecution() && !isQuery {
		logCtx.WithField("request_payload", request).Warn("Skip hitting db for queries from dashboard, if not found on cache.")
		return nil, statusCode, PROCESSING_FAILED, "Not found in cache. Execution suspended temporarily.", true
	}*/

	model.SetQueryCachePlaceholder(projectID, &request)
	H.SleepIfHeaderSet(c)

	var duplicatedRequest model.KPIQueryGroup
	U.DeepCopy(&request, &duplicatedRequest)
	queryResult, statusCode := store.GetStore().ExecuteKPIQueryGroup(projectID, reqID,
		duplicatedRequest, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
	if statusCode != http.StatusOK {
		model.DeleteQueryCacheKey(projectID, &request)
		logCtx.Error("Failed to process query from DB")
		if statusCode == http.StatusPartialContent {
			return queryResult, statusCode, PROCESSING_FAILED, "Failed to process query from DB", true
		}
		return nil, statusCode, PROCESSING_FAILED, "Failed to process query from DB", true
	}
	meta := model.CacheMeta{
		Timezone:       string(timezoneString),
		From:           commonQueryFrom,
		To:             commonQueryTo,
		RefreshedAt:    U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		Preset:         preset,
	}
	for i, _ := range queryResult {
		queryResult[i].CacheMeta = meta
	}

	model.SetQueryCacheResult(projectID, &request, queryResult)

	// if it is a dashboard query, cache it
	if isDashboardQueryRequest {
		if C.IsLastComputedWhitelisted(projectID) {
			model.SetCacheResultByDashboardIdAndUnitIdWithPreset(queryResult, projectID, dashboardId, unitId, preset,
				commonQueryFrom, commonQueryTo, timezoneString, meta)
		} else {
			model.SetCacheResultByDashboardIdAndUnitId(queryResult, projectID, dashboardId, unitId,
				commonQueryFrom, commonQueryTo, timezoneString, meta)
		}

		if allowSyncReferenceFields {
			queryResult, err = store.GetStore().AddPropertyValueLabelToQueryResults(projectID, queryResult)
			if err != nil {
				logCtx.WithError(err).Error("Failed to set property value label.")
			}
		}

		return H.DashboardQueryResponsePayload{Result: queryResult, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(), CacheMeta: meta}, http.StatusOK, "", "", false
	}
	isQueryShareable := isQueryShareable(request)

	if allowSyncReferenceFields {
		queryResult, err = store.GetStore().AddPropertyValueLabelToQueryResults(projectID, queryResult)
		if err != nil {
			logCtx.WithError(err).Error("Failed to set property value label.")
		}
	}
	return gin.H{"result": queryResult, "query": request, "sharable": isQueryShareable}, http.StatusOK, "", "", false
}

func isQueryShareable(request model.KPIQueryGroup) bool {
	// if breakdown is not present, then it is sharable
	if request.GlobalGroupBy == nil || len(request.GlobalGroupBy) == 0 {
		return true
	}
	// if breakdown is present, then it is not sharable
	return false
}

// isDashboardQueryLocked returns if query run for dashboard is locked. isDashboardQueryRequest returns if its any dashboard query request.
func getValidKPIQueryAndDetailsFromRequest(r *http.Request, c *gin.Context, logCtx *log.Entry, projectId int64, dashboardIdParam, unitIdParam, queryIdParam string) (model.KPIQueryGroup, int64, int64, bool, bool, int, string, string, bool) {
	var unitId int64
	requestPayload, queryPayload, isDashboardQueryLocked := model.KPIQueryGroup{}, model.KPIQueryGroup{}, false
	dashboardId, err := strconv.ParseInt(dashboardIdParam, 10, 64)
	if err != nil {
		err = errors.New("Query failed. Invalid DashboardID.")
	}
	unitId, err = strconv.ParseInt(unitIdParam, 10, 64)
	if err != nil || unitId == 0 {
		err = errors.New("Query failed. Invalid DashboardUnitID.")
	}
	isDashboardQueryRequest := dashboardIdParam != "" && unitIdParam != ""

	if queryIdParam == "" {
		if err := c.BindJSON(&requestPayload); err != nil {
			logCtx.WithError(err).Error("Query failed. Json decode failed.")
			return queryPayload, dashboardId, unitId, isDashboardQueryRequest, isDashboardQueryLocked, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
		}
	}

	if isDashboardQueryRequest {

		_, query, err := store.GetStore().GetQueryFromUnitID(projectId, unitId)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return queryPayload, dashboardId, unitId, isDashboardQueryRequest, isDashboardQueryLocked, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		// If there is a difference between request and internal Query, lets log it and error out with different status Code.

		isDashboardQueryLocked = query.LockedForCacheInvalidation
		U.DecodePostgresJsonbToStructType(&query.Query, &queryPayload)
	} else if queryIdParam != "" {
		_, query, err := store.GetStore().GetQueryAndClassFromQueryIdString(queryIdParam, projectId)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return queryPayload, dashboardId, unitId, isDashboardQueryRequest, isDashboardQueryLocked, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		U.DecodePostgresJsonbToStructType(&query.Query, &queryPayload)
	} else {
		queryPayload = requestPayload
	}

	if queryIdParam == "" {
		queryPayload.SetQueryDateRange(requestPayload.Queries[0].From, requestPayload.Queries[0].To)
		if requestPayload.Queries[0].Timezone != "" {
			queryPayload.SetTimeZone(U.TimeZoneString(requestPayload.Queries[0].Timezone))
		}

		var inputGroupByTimestamp string
		for _, query := range requestPayload.Queries {
			if query.GroupByTimestamp != "" {
				inputGroupByTimestamp = query.GroupByTimestamp
			}
		}

		for index := range queryPayload.Queries {
			if queryPayload.Queries[index].GroupByTimestamp != "" {
				queryPayload.Queries[index].GroupByTimestamp = inputGroupByTimestamp
			}
		}
	}

	if len(queryPayload.Queries) == 0 {
		logCtx.Error("Query failed. Empty query group.")
		return queryPayload, dashboardId, unitId, isDashboardQueryRequest, isDashboardQueryLocked, http.StatusBadRequest, INVALID_INPUT, "Query failed. Empty query group.", true
	}
	return queryPayload, dashboardId, unitId, isDashboardQueryRequest, isDashboardQueryLocked, http.StatusOK, "", "", false
}

func setTimezoneForKPIRequest(logCtx *log.Entry, requestPayload model.KPIQueryGroup, projectId int64) (model.KPIQueryGroup, int, string, string, bool) {
	var timezoneString U.TimeZoneString
	if requestPayload.Queries[0].Timezone != "" {
		_, errCode := time.LoadLocation(string(requestPayload.Queries[0].Timezone))
		if errCode != nil {
			return requestPayload, http.StatusBadRequest, INVALID_INPUT, "Query failed. Invalid Timezone provided.", true
		}
		timezoneString = U.TimeZoneString(requestPayload.Queries[0].Timezone)
	} else {
		var statusCode int
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectId)
		if statusCode != http.StatusFound {
			logCtx.Error("Query failed. Failed to get Timezone.")
			return requestPayload, http.StatusBadRequest, INVALID_INPUT, "Query failed. Failed to get Timezone.", true
		}
	}
	requestPayload.SetTimeZone(timezoneString)
	return requestPayload, http.StatusOK, "", "", false
}

func getOtherRelatedInformationFromRequest(request model.KPIQueryGroup, projectID int64, presetParam, refreshParam, isQueryParam string) (string, int64, int64, bool, bool) {
	hardRefresh, preset := false, ""

	commonQueryFrom, commonQueryTo, timeZoneString := request.Queries[0].From, request.Queries[0].To, request.Queries[0].Timezone
	if U.PresetLookup[presetParam] != "" && C.IsLastComputedWhitelisted(projectID) {
		preset = presetParam
	}

	if refreshParam != "" {
		hardRefresh, _ = strconv.ParseBool(refreshParam)
	}
	isQuery := false
	if isQueryParam != "" {
		isQuery, _ = strconv.ParseBool(isQueryParam)
	}

	if preset == "" {
		preset = U.GetPresetNameByFromAndTo(commonQueryFrom, commonQueryTo, U.TimeZoneString(timeZoneString))
	}
	return preset, commonQueryFrom, commonQueryTo, hardRefresh, isQuery
}

func GetResultFromCacheOrDashboard(c *gin.Context, reqID string, projectID int64, request model.KPIQueryGroup,
	dashboardId int64, unitId int64, preset string, commonQueryFrom int64, commonQueryTo int64, hardRefresh bool,
	timezoneString U.TimeZoneString, isDashboardQueryRequest bool, logCtx *log.Entry, skipContextVerfication bool) (interface{}, int, string, string, bool) {

	// If refresh is passed, refresh only is Query.From is of todays beginning.
	if isDashboardQueryRequest && !H.ShouldAllowHardRefresh(commonQueryFrom, commonQueryTo, request.GetTimeZone(), hardRefresh) {
		var shouldReturn bool
		var resCode int
		var resMsg interface{}
		if C.IsLastComputedWhitelisted(projectID) {
			shouldReturn, resCode, resMsg = H.GetResponseIfCachedDashboardQueryWithPreset(reqID, projectID, dashboardId, unitId, preset, commonQueryFrom, commonQueryTo, timezoneString)
		} else {
			shouldReturn, resCode, resMsg = H.GetResponseIfCachedDashboardQuery(reqID, projectID, dashboardId, unitId, commonQueryFrom, commonQueryTo, timezoneString)
		}
		if shouldReturn {
			if resCode == http.StatusOK {
				return resMsg, resCode, "", "", false
			}
		}
	}

	var cacheResult model.ChannelResultGroupV1
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectID, &request, cacheResult, isDashboardQueryRequest, reqID, skipContextVerfication)
	if shouldReturn {
		if resCode == http.StatusOK {
			return resMsg, resCode, "", "", false
		}
		logCtx.Error("Query failed. Error Processing/Fetching data from Query cache")
		return nil, resCode, PROCESSING_FAILED, "Error Processing/Fetching data from Query cache", true
	}
	return nil, http.StatusProcessing, "", "", true
}
