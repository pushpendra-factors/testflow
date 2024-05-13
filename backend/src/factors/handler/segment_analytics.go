package handler

import (
	"encoding/json"
	"factors/cache"
	cacheRedis "factors/cache/redis"
	v1 "factors/handler/v1"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const FifteenMinutesTime = 15 * 60
const TwentyMinutesTime = 20 * 60

// GetSegmentAnalyticsConfigHandler godoc
// @Summary To get config for the segment analytics.
// @Tags SegmentAnalytics
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} json model.SegmentAnalyticsConfig
// @Router /{project_id}/v1/segments/analytics/config [get]
func GetSegmentAnalyticsConfigHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	widgetGroups, errMsg, statusCode := store.GetStore().GetWidgetGroupAndWidgetsForConfig(projectID)
	log.WithField("errMsg", errMsg).WithField("projectID", projectID).Warn("Failed in getting config for widget group")
	responseData := gin.H{
		"result": widgetGroups,
	}

	// Return JSON response
	c.JSON(statusCode, responseData)
}

func getReqIDProjectIDAndWidgetGroupID(c *gin.Context) (string, int64, string) {
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	widgetGroupID := c.Params.ByName("widget_group_id")
	return reqID, projectID, widgetGroupID
}

func getReqIDProjectIDAndWidgetGroupIDAndWidgetID(c *gin.Context) (string, int64, string, string) {
	reqID, projectID, widgetGroupID := getReqIDProjectIDAndWidgetGroupID(c)
	widgetID := c.Params.ByName("widget_id")
	return reqID, projectID, widgetGroupID, widgetID
}

func getReqIDSegmentIDProjectIDAndWidgetGroupID(c *gin.Context) (string, int64, string, string) {
	reqID, projectID, widgetGroupID := getReqIDProjectIDAndWidgetGroupID(c)
	segmentID := c.Params.ByName("segment_id")
	return reqID, projectID, widgetGroupID, segmentID
}

// AddNewWidgetToWidgetGroupHandler godoc.
// @Summary To add widget to segment widget group.
// @Tags SegmentAnalytics
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param widget_group_id path integer true "Widget group ID"
// @Param query body model.SegmentAnalyticsWidget true "Segment analytics widget payload."
// @Success 200 {string} json "{"result": model.SegmentAnalyticsWidget}"
// @Router /{project_id}/segments/analytics/widget_group/{widget_group_id}/widgets [post]
func AddNewWidgetToWidgetGroupHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID, widgetGroupID := getReqIDProjectIDAndWidgetGroupID(c)
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)
	if projectID == 0 {
		logCtx.Error("AddNewWidgetToWidgetGroupHandler - Add widget Segment analytics failed. Invalid project.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Invalid project id", true
	}
	if widgetGroupID == "" {
		logCtx.Error("AddNewWidgetToWidgetGroupHandler - Add widget Segment analytics failed. Invalid widget group ID.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Adding Widget failed. Invalid widget group ID", true
	}
	request := model.Widget{}
	err := c.BindJSON(&request)
	if err != nil {
		var requestAsMap map[string]interface{}
		c.BindJSON(&requestAsMap)
		logCtx.Warnf("Decode failed on request to Widget struct. %v", requestAsMap)
		return nil, http.StatusBadRequest, INVALID_INPUT, "Error during decode of widget.", true
	}

	isValid, errMsg := request.IsValid()
	if isValid == false {
		log.WithField("message finder", "Error during widget validation").Warn(errMsg)
		return nil, http.StatusBadRequest, INVALID_INPUT, errMsg, true
	}

	// Code should be in service probably.
	widgetGroup, errMsg, statusCode := store.GetStore().GetWidgetGroupByID(projectID, widgetGroupID)
	if statusCode != http.StatusFound {
		log.WithField("errMsg", errMsg).WithField("projectID", projectID).Warn("Failed in getting widget group by ID")
		return nil, statusCode, v1.PROCESSING_FAILED, "Failed during get of widget group", true
	}

	if request.QueryType == model.QueryClassKPI {
		// Currently restricting it to crm based KPI alone.
		_, errMsg, statusCode = store.GetStore().GetKpiRelatedCustomMetricsByName(projectID, request.QueryMetric)
		if errMsg != "" {
			if statusCode == http.StatusNotFound {
				logCtx.WithField("statusCode", statusCode).WithField("errMsg", errMsg).Warn("Failed in widget fetch")
				return nil, http.StatusBadRequest, INVALID_INPUT, "Invalid custom KPI", true
			}
			if statusCode != http.StatusFound {
				logCtx.WithField("statusCode", statusCode).WithField("errMsg", errMsg).Warn("Failed in widget fetch")
				return nil, statusCode, v1.PROCESSING_FAILED, errMsg, true
			}
		}
	}

	isValid, errmsg := request.ValidateConstraints(widgetGroup)
	if !isValid {
		return nil, http.StatusConflict, v1.DUPLICATE_RECORD, errmsg, true
	}

	markerError := setWidgetGroupMarker(projectID, widgetGroupID)
	if markerError != nil {
		log.WithField("projectID", projectID).WithField("widgetGroupID", widgetGroupID).Error("Failed in adding new Widget group" + markerError.Error())
	}

	request, errMsg, statusCode = store.GetStore().AddWidgetToWidgetGroup(widgetGroup, request)
	if statusCode != http.StatusCreated {
		return nil, statusCode, v1.PROCESSING_FAILED, errMsg, true
	}
	responseData := gin.H{
		"result": request,
	}
	return responseData, http.StatusOK, "", "", false
}

// EditSegmentAnalyticsWidgetHandler godoc.
// @Summary To edit segment widget.
// @Tags SegmentAnalytics
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param widget_group_id path integer true "Widget group ID"
// @Param widget_id path integer true "Widget ID"
// @Param query body model.SegmentAnalyticsWidget true "Segment analytics widget payload."
// @Success 200 {string} json "{"result": model.SegmentAnalyticsWidget}"
// @Router /{project_id}/segments/analytics/widget_group/{widget_group_id}/widgets/{widget_id} [patch]
func EditSegmentAnalyticsWidgetHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID, widgetGroupID, widgetID := getReqIDProjectIDAndWidgetGroupIDAndWidgetID(c)
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)
	if projectID == 0 {
		logCtx.Error("AddNewWidgetToWidgetGroupHandler - Edit widget Segment analytics failed. Invalid project.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Edit Widget failed. Invalid project ID", true
	}
	if widgetGroupID == "" {
		logCtx.Error("AddNewWidgetToWidgetGroupHandler - Edit widget Segment analytics failed. Invalid widget group ID.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Edit Widget failed. Invalid widget group ID", true
	}
	if widgetID == "" {
		logCtx.Error("AddNewWidgetToWidgetGroupHandler - Edit widget Segment analytics failed. Invalid widget ID.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Edit Widget failed. Invalid widget ID", true
	}

	request := model.Widget{}
	err := c.BindJSON(&request)
	if err != nil {
		var requestAsMap map[string]interface{}
		c.BindJSON(&requestAsMap)
		logCtx.Warnf("Decode failed on request to Widget struct. %v", requestAsMap)
		return nil, http.StatusBadRequest, INVALID_INPUT, "Error during decode of widget.", true
	}

	// add and validate - since its patch.
	widgetGroup, widget, errMsg, statusCode2 := store.GetStore().GetWidgetAndWidgetGroupByWidgetID(projectID, widgetGroupID, widgetID)
	if statusCode2 != http.StatusFound {
		return nil, http.StatusBadRequest, INVALID_INPUT, errMsg, true
	}
	displayNameChanged := request.DisplayName != ""
	if displayNameChanged {
		widget.DisplayName = request.DisplayName
	}
	queryMetricChanged := request.QueryMetric != ""
	// If prev type is x, the update should also contain only x.
	if queryMetricChanged {
		widget.QueryMetric = request.QueryMetric
	}

	// Currently restricting it to crm based KPI alone.
	_, errMsg, statusCode := store.GetStore().GetKpiRelatedCustomMetricsByName(projectID, widget.QueryMetric)
	if errMsg != "" {
		if statusCode == http.StatusNotFound {
			logCtx.WithField("statusCode", statusCode).WithField("errMsg", errMsg).Warn("Failed in Custom Metric fetch")
			return nil, http.StatusBadRequest, INVALID_INPUT, "Invalid custom KPI", true
		}
		if statusCode != http.StatusFound {
			logCtx.WithField("statusCode", statusCode).WithField("errMsg", errMsg).Warn("Failed in Custom Metric fetch")
			return nil, statusCode, v1.PROCESSING_FAILED, errMsg, true
		}
	}

	isValid, errmsg := widget.ValidateUpdatedConstraints(widgetGroup, displayNameChanged, queryMetricChanged)
	if !isValid {
		return nil, http.StatusConflict, v1.DUPLICATE_RECORD, errmsg, true
	}

	if queryMetricChanged {
		markerError := setWidgetGroupMarker(projectID, widgetGroupID)
		if markerError != nil {
			log.WithField("projectID", projectID).WithField("widgetGroupID", widgetGroupID).Error("Failed in Edit Widget group" + markerError.Error())
		}
	}

	widget, errMsg, statusCode = store.GetStore().UpdateWidgetToWidgetGroup(widgetGroup, widget)
	if statusCode != http.StatusOK {
		logCtx.WithField("errMsg", errMsg).Warn("Failed in updating widget to widget group")
		return nil, statusCode, v1.PROCESSING_FAILED, errMsg, true
	}

	responseData := gin.H{
		"result": widget,
	}

	return responseData, http.StatusOK, "", "", false
}

func hardRefreshPresent(c *gin.Context) bool {
	refreshParam := c.Query("refresh")
	if refreshParam != "" {
		hardRefresh, _ := strconv.ParseBool(refreshParam)
		return hardRefresh
	}
	return false
}

// ExecuteSegmentQueryHandler godoc.
// @Summary To run segment widget group query.
// @Tags SegmentAnalytics, Query
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param segment_id path integer true "Segment ID"
// @Param widget_group_id path integer true "Widget group ID"
// @Param query body model. true "Segment analytics query payload."
// @Success 200 {string} json "{"result": model.SegmentAnalyticsWidget}"
// @Router /{project_id}/segments/{segment_id}/analytics/widget_group/{widget_group_id}/query [post]
func ExecuteSegmentQueryHandler(c *gin.Context) {

	reqID, projectID, widgetGroupID, segmentID := getReqIDSegmentIDProjectIDAndWidgetGroupID(c)
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)
	hardRefresh := hardRefreshPresent(c)
	if projectID == 0 {
		logCtx.Error("AddNewWidgetToWidgetGroupHandler - Execute Segment analytics failed. Invalid project.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.SegmentResponse{
			Error: "Execute Segment failed. Invalid project ID."})
		return
	}
	if widgetGroupID == "" {
		logCtx.Error("AddNewWidgetToWidgetGroupHandler - Execute Segment analytics failed. Invalid widget group ID.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.SegmentResponse{
			Error: "Execute Segment failed. Invalid widget group ID."})
		return
	}
	if segmentID == "" {
		logCtx.Error("AddNewWidgetToWidgetGroupHandler - Execute Segment analytics failed. Invalid segment ID.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.SegmentResponse{
			Error: "Execute Segment failed. Invalid segment ID."})
		return
	}

	widgetGroup, errMsg, statusCode := store.GetStore().GetWidgetGroupByID(projectID, widgetGroupID)
	if statusCode != http.StatusFound {
		log.WithField("errMsg", errMsg).WithField("projectID", projectID).Warn("Failed in getting widget group by ID")
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.SegmentResponse{
			Error: "Failed in getting widget group by ID"})
		return
	}

	markerExists, statusCode := getIfExistsWidgetGroupMarker(projectID, widgetGroupID)
	if statusCode == http.StatusInternalServerError {
		log.WithField("projectID", projectID).WithField("widgetGroupID", widgetGroupID).Warn("Failed in getting getIfExistsWidgetGroupMarker")
		markerExists = true
	}

	requestParams := model.RequestSegmentKPI{}
	if err := c.BindJSON(&requestParams); err != nil {
		var requestAsMap map[string]interface{}
		c.BindJSON(&requestAsMap)
		logCtx.Warnf("Decode failed on request to RequestSegmentKPI struct. %v", requestAsMap)
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.SegmentResponse{
			Error: "Decode failed on request to RequestSegmentKPI struct"})
		return
	}

	// Keeping the cache expiry to 15 minutes. Not invalidating the cache.
	if !hardRefresh && !markerExists {
		log.Warn("getting values from cache")
		shouldReturn, resCode, resMsg := GetSegmentResponseIfCachedQuery(c, projectID, segmentID, widgetGroupID, requestParams.From, requestParams.To)
		if shouldReturn {
			if resCode == http.StatusOK {
				responseData := gin.H{
					"result": resMsg,
				}
				c.JSON(resCode, responseData)
				return
			}
		}
		log.Warn("setting placeholder values to cache")
		SetSegmentCachePlaceholder(projectID, segmentID, widgetGroupID, requestParams.From, requestParams.To)
	}

	results, statusCode := store.GetStore().ExecuteWidgetGroup(projectID, widgetGroup, segmentID, reqID, requestParams)
	responseData := gin.H{
		"result": results,
	}
	if statusCode != http.StatusOK {
		DeleteSegmentCache(projectID, segmentID, widgetGroupID, requestParams.From, requestParams.To)
		c.JSON(statusCode, responseData)
		return
	}
	if !markerExists {
		log.Warn("setting values from cache")
		SetSegmentCacheResult(projectID, segmentID, widgetGroupID, requestParams.From, requestParams.To, results)
	}

	c.JSON(statusCode, responseData)
}

func GetSegmentResponseIfCachedQuery(c *gin.Context, projectID int64, segmentID, widgetGroup string, from, to int64) (bool, int, interface{}) {
	cacheResult, errCode := GetSegmentResultFromCache(c, projectID, segmentID, widgetGroup, from, to)
	if errCode == http.StatusFound {
		return getSegmentCacheResponse(c, cacheResult)
	} else if errCode == http.StatusAccepted {
		// An instance of query is in progress. Poll for result.
		for {

			time.Sleep(5 * time.Second)

			cacheResult, errCode = GetSegmentResultFromCache(c, projectID, segmentID, widgetGroup, from, to)
			if errCode == http.StatusAccepted {
				continue
			} else if errCode == http.StatusFound {
				return getSegmentCacheResponse(c, cacheResult)
			} else {
				// If not in Accepted state, return with error.
				return true, http.StatusInternalServerError, errors.New("Query Cache: Failed to fetch from cache")
			}
		}
	}
	return false, errCode, errors.New("Query Cache: Failed to fetch from cache")
}

func GetSegmentResultFromCache(c *gin.Context, projectID int64, segmentID, widgetGroup string, from, to int64) (interface{}, int) {
	queryResult := model.QueryCacheResult{}

	cacheKey, _ := getSegmentWidgetGroupCacheKey(projectID, segmentID, widgetGroup, from, to)
	value, exists, err := cacheRedis.GetIfExistsPersistent(cacheKey)

	if err != nil {
		return queryResult.Result, http.StatusInternalServerError
	}
	if !exists {
		return queryResult.Result, http.StatusNotFound
	} else if value == model.QueryCacheInProgressPlaceholder {
		return queryResult.Result, http.StatusAccepted
	}

	err = json.Unmarshal([]byte(value), &queryResult)
	if err != nil {
		log.WithError(err).Error("Segment analytics : Failed to unmarshal cache result to result container")
		return queryResult, http.StatusInternalServerError
	}
	return queryResult.Result, http.StatusFound
}

func getSegmentCacheResponse(c *gin.Context, result interface{}) (bool, int, interface{}) {
	c.Header(model.QueryCacheResponseFromCacheHeader, "true")
	return true, http.StatusOK, result
}

// SetQueryCachePlaceholder To set a placeholder temporarily to indicate that query is already running.
func SetSegmentCachePlaceholder(projectID int64, segmentID, widgetGroup string, from, to int64) {
	cacheKey, _ := getSegmentWidgetGroupCacheKey(projectID, segmentID, widgetGroup, from, to)
	cacheRedis.SetPersistent(cacheKey, model.QueryCacheInProgressPlaceholder, model.QueryCachePlaceholderExpirySeconds)
}

func getWidgetGroupMarkerKey(projectID int64, widgetGroup string) (*cache.Key, error) {
	suffix := fmt.Sprintf("widgetGroup:%s", widgetGroup)
	return cache.NewKey(projectID, model.WidgetGroupMarker, suffix)
}

// Set when edit of widget group query is changed.
func setWidgetGroupMarker(projectID int64, widgetGroup string) error {
	cacheKey, _ := getWidgetGroupMarkerKey(projectID, widgetGroup)
	return cacheRedis.SetPersistent(cacheKey, string("1"), TwentyMinutesTime)
}

// Marker here implies cache is getting invalidated.
func getIfExistsWidgetGroupMarker(projectID int64, widgetGroup string) (bool, int) {
	cacheKey, _ := getWidgetGroupMarkerKey(projectID, widgetGroup)
	_, exists, err := cacheRedis.GetIfExistsPersistent(cacheKey)
	if err != nil {
		return true, http.StatusInternalServerError
	}
	if !exists {
		return false, http.StatusNotFound
	} else {
		return true, http.StatusFound
	}

}

// DeleteQueryCacheKey Delete a query cache key on error.
func DeleteSegmentCache(projectID int64, segmentID string, widgetGroup string, fr, to int64) {
	cacheKey, _ := getSegmentWidgetGroupCacheKey(projectID, segmentID, widgetGroup, fr, to)
	cacheRedis.DelPersistent(cacheKey)
}

func SetSegmentCacheResult(projectID int64, segmentID string, widgetGroup string, fr, to int64, results []model.QueryResult) {
	cacheKey, _ := getSegmentWidgetGroupCacheKey(projectID, segmentID, widgetGroup, fr, to)
	queryCache := model.QueryCacheResult{
		Result: results,
	}
	queryResultString, err := json.Marshal(queryCache)
	if err != nil {
		return
	}
	cacheRedis.SetPersistent(cacheKey, string(queryResultString), FifteenMinutesTime)
}

func getSegmentWidgetGroupCacheKey(projectID int64, segmentID string, widgetGroup string, fr, to int64) (*cache.Key, error) {
	suffix := fmt.Sprintf("segment:%s:widgetGroup:%s:from:%d:to:%d", segmentID, widgetGroup, fr, to)
	return cache.NewKey(projectID, model.QueryCacheKeyForSegmentAnalytics, suffix)
}

// Doesnt have any dependencies to be checked when deleting segment analysis.
// DeleteSegmentAnalyticsWidgethandler godoc
// @Summary To delete segment widget group.
// @Tags SegmentAnalytics
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param widget_group_id path integer true "Widget group ID"
// @Param widget_id path integer true "Widget ID"
// @Router /{project_id}/segments/analytics/widget_group/{widget_group_id}/widgets/{widget_id} [delete]
func DeleteSegmentAnalyticsWidgetHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	reqID, projectID, widgetGroupID, widgetID := getReqIDProjectIDAndWidgetGroupIDAndWidgetID(c)
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)
	if projectID == 0 {
		logCtx.Error("AddNewWidgetToWidgetGroupHandler - Delete widget Segment analytics failed. Invalid project.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Delting Widget failed. Invalid project id", true
	}
	if widgetGroupID == "" {
		logCtx.Error("AddNewWidgetToWidgetGroupHandler - Delete widget Segment analytics failed. Invalid widget group ID.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Deleting Widget failed. Invalid widget group ID", true
	}
	if widgetID == "" {
		logCtx.Error("AddNewWidgetToWidgetGroupHandler - Delete widget Segment analytics failed. Invalid widget ID.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Delete Widget failed. Invalid widget ID", true
	}

	errmsg, statusCode := store.GetStore().DeleteWidgetFromWidgetGroup(projectID, widgetGroupID, widgetID)
	if statusCode != http.StatusAccepted {
		if statusCode == http.StatusBadRequest {
			logCtx.WithField("errMsg", errmsg).Warn("Failed in deleting widget of widget group")
			return nil, statusCode, v1.INVALID_INPUT, "Failed in deleting widget", true
		} else {
			logCtx.WithField("errMsg", errmsg).Warn("Failed in deleting widget of widget group")
			return nil, statusCode, v1.PROCESSING_FAILED, "", true
		}
	}
	return nil, http.StatusOK, "", "", false
}
