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
	Category        string `json:"category"`
	DisplayCategory string `json:"display_category"`
	ObjectType      string `json:"object_type"`
	PropertyName    string `json:"property_name"`
	Entity          string `json:"entity"`
	Metric          string `json:"me"`
}

func (req *KPIFilterValuesRequest) isValid() bool {
	if req == nil {
		return false
	}
	if req.Category == "" || !U.ContainsStringInArray([]string{model.ChannelCategory, model.EventCategory}, req.Category) ||
		req.Entity == "" || !U.ContainsStringInArray([]string{model.EventEntity, model.UserEntity}, req.Entity) ||
		req.ObjectType == "" || req.PropertyName == "" {
		return false
	}
	return true
}

func getReqIDAndProjectID(c *gin.Context) (string, uint64) {
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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
	storeSelected := store.GetStore()
	configFunctions := []func(uint64, string) (map[string]interface{}, int){
		storeSelected.GetKPIConfigsForWebsiteSessions,
		storeSelected.GetKPIConfigsForPageViews,
		storeSelected.GetKPIConfigsForFormSubmissions,
		storeSelected.GetKPIConfigsForHubspotContacts,
		storeSelected.GetKPIConfigsForHubspotCompanies,
		storeSelected.GetKPIConfigsForHubspotDeals,
		storeSelected.GetKPIConfigsForSalesforceUsers,
		storeSelected.GetKPIConfigsForSalesforceAccounts,
		storeSelected.GetKPIConfigsForSalesforceOpportunities,
		storeSelected.GetKPIConfigsForAdwords, storeSelected.GetKPIConfigsForGoogleOrganic,
		storeSelected.GetKPIConfigsForFacebook, storeSelected.GetKPIConfigsForLinkedin,
		storeSelected.GetKPIConfigsForAllChannels,
	}
	resultantResultConfigs := make([]map[string]interface{}, 0)
	for _, configFunction := range configFunctions {
		currentConfig, errCode := configFunction(projectID, reqID)
		if errCode != http.StatusOK {
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of KPI Config Data.", true
		}
		if currentConfig != nil {
			resultantResultConfigs = append(resultantResultConfigs, currentConfig)
		}
	}
	return resultantResultConfigs, http.StatusOK, "", "", false
}

// GetKPIFilterValuesHandler godoc
// @Summary To filter on values for kpi query.
// @Tags KPIanalytics, KPIQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query body model.KPIFilterValuesRequest true "Filter Values payload"
// @Success 200 {string} json "{"result": interface{}}"
// @Router /{project_id}/v1/channels/filter_values [get]
func GetKPIFilterValuesHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID := getReqIDAndProjectID(c)
	storeSelected := store.GetStore()
	var resultantFilterValuesResponse interface{}
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
	logCtx = logCtx.WithField("request", request)

	if request.Category == model.ChannelCategory {
		currentChannel, err := model.GetChannelFromKPIQuery(request.DisplayCategory)
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
	} else if request.Category == model.EventCategory && request.Entity == model.EventEntity {
		request.ObjectType = model.GetObjectTypeForFilterValues(request.ObjectType, request.Metric)
		eventsFilterValues, err := storeSelected.GetPropertyValuesByEventProperty(projectID, request.ObjectType, request.PropertyName, model.FilterValuesOrEventNamesLimit, C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.Warn(err)
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Error during fetch of KPI FilterValues Data.", true
		}
		resultantFilterValuesResponse = eventsFilterValues
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

	queryIdString := c.Query("query_id")
	request := model.KPIQueryGroup{}
	if queryIdString == "" {
		err := c.BindJSON(&request)
		if err != nil {
			return nil, http.StatusBadRequest, INVALID_INPUT, "Error during validation of execute KPIQuery.", true
		}
	} else {
		_, query, err := store.GetStore().GetQueryAndClassFromQueryIdString(queryIdString, projectID)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		U.DecodePostgresJsonbToStructType(&query.Query, &request)
	}

	dashboardId, unitId, commonQueryFrom, commonQueryTo, hardRefresh, isDashboardQueryRequest, _, err := getDashboardRelatedInformationFromRequest(request,
		c.Query("dashboard_id"), c.Query("dashboard_unit_id"), c.Query("refresh"), c.Query("is_query"))
	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, err.Error(), true
	}

	if request.Queries[0].Timezone != "" {
		_, errCode := time.LoadLocation(string(request.Queries[0].Timezone))
		if errCode != nil {
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Invalid Timezone provided.", true
		}

		timezoneString = U.TimeZoneString(request.Queries[0].Timezone)
	} else {
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectID)
		if statusCode != http.StatusFound {
			logCtx.Error("Query failed. Failed to get Timezone.")
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Failed to get Timezone.", true
		}
		// logCtx.WithError(err).Error("Query failed. Invalid Timezone.")
	}

	request.SetTimeZone(timezoneString)
	err = request.TransformDateTypeFilters()
	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, err.Error(), true
	}

	data, statusCode, errorCode, errMsg, isErr := getResultFromCacheOrDashboard(c, reqID, projectID, request, dashboardId, unitId, commonQueryFrom, commonQueryTo, hardRefresh, timezoneString, isDashboardQueryRequest, logCtx)
	if statusCode != http.StatusProcessing {
		return data, statusCode, errorCode, errMsg, isErr
	}

	/*if isDashboardQueryRequest && C.DisableDashboardQueryDBExecution() && !isQuery {
		logCtx.WithField("request_payload", request).Warn("Skip hitting db for queries from dashboard, if not found on cache.")
		return nil, statusCode, PROCESSING_FAILED, "Not found in cache. Execution suspended temporarily.", true
	}*/

	model.SetQueryCachePlaceholder(projectID, &request)
	H.SleepIfHeaderSet(c)

	queryResult, statusCode := store.GetStore().ExecuteKPIQueryGroup(projectID, reqID, request)
	if statusCode != http.StatusOK {
		model.DeleteQueryCacheKey(projectID, &request)
		logCtx.Error("Failed to process query from DB")
		if statusCode == http.StatusPartialContent {
			return queryResult, statusCode, PROCESSING_FAILED, "Failed to process query from DB", true
		}
		return nil, statusCode, PROCESSING_FAILED, "Failed to process query from DB", true
	}
	model.SetQueryCacheResult(projectID, &request, queryResult)

	// if it is a dashboard query, cache it
	if isDashboardQueryRequest {
		model.SetCacheResultByDashboardIdAndUnitId(queryResult, projectID, dashboardId, unitId, commonQueryFrom, commonQueryTo, request.GetTimeZone())
		return H.DashboardQueryResponsePayload{Result: queryResult, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix()}, http.StatusOK, "", "", false
	}

	return gin.H{"result": queryResult, "query": request}, http.StatusOK, "", "", false
}

func getDashboardRelatedInformationFromRequest(request model.KPIQueryGroup, dashboardIdParam, unitIdParam, refreshParam, isQueryParam string) (uint64, uint64, int64, int64, bool, bool, bool, error) {
	var dashboardId uint64
	var unitId uint64
	var err error
	hardRefresh := false

	commonQueryFrom := request.Queries[0].From
	commonQueryTo := request.Queries[0].To
	if refreshParam != "" {
		hardRefresh, _ = strconv.ParseBool(refreshParam)
	}
	isQuery := false
	if isQueryParam != "" {
		isQuery, _ = strconv.ParseBool(isQueryParam)
	}

	isDashboardQueryRequest := dashboardIdParam != "" && unitIdParam != ""
	if !isDashboardQueryRequest {
		return dashboardId, unitId, commonQueryFrom, commonQueryTo, hardRefresh, isDashboardQueryRequest, isQuery, err
	}
	dashboardId, err = strconv.ParseUint(dashboardIdParam, 10, 64)
	unitId, err = strconv.ParseUint(unitIdParam, 10, 64)
	if err != nil || unitId == 0 {
		err = errors.New("Query failed. Invalid DashboardID.")
	}
	if err != nil || unitId == 0 {
		err = errors.New("Query failed. Invalid DashboardUnitID.")
	}
	return dashboardId, unitId, commonQueryFrom, commonQueryTo, hardRefresh, isDashboardQueryRequest, isQuery, err
}

func getResultFromCacheOrDashboard(c *gin.Context, reqID string, projectID uint64, request model.KPIQueryGroup,
	dashboardId uint64, unitId uint64, commonQueryFrom int64, commonQueryTo int64, hardRefresh bool,
	timezoneString U.TimeZoneString, isDashboardQueryRequest bool, logCtx *log.Entry) (interface{}, int, string, string, bool) {

	// Tracking dashboard query request.
	if isDashboardQueryRequest {
		model.SetDashboardCacheAnalytics(projectID, dashboardId, unitId, commonQueryFrom, commonQueryTo, timezoneString)
	}

	// If refresh is passed, refresh only is Query.From is of todays beginning.
	if isDashboardQueryRequest && !H.ShouldAllowHardRefresh(commonQueryFrom, commonQueryTo, request.GetTimeZone(), hardRefresh) {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedDashboardQuery(reqID, projectID, dashboardId, unitId, commonQueryFrom, commonQueryTo, timezoneString)
		if shouldReturn {
			if resCode == http.StatusOK {
				return resMsg, resCode, "", "", false
			}
		}
	}

	var cacheResult model.ChannelResultGroupV1
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectID, &request, cacheResult, isDashboardQueryRequest, reqID)
	if shouldReturn {
		if resCode == http.StatusOK {
			return resMsg, resCode, "", "", false
		}
		logCtx.Error("Query failed. Error Processing/Fetching data from Query cache")
		return nil, resCode, PROCESSING_FAILED, "Error Processing/Fetching data from Query cache", true
	}
	return nil, http.StatusProcessing, "", "", true
}
