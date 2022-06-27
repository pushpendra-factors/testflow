package model

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

type WebAnalyticsQueries struct {
	// Multiple queries  with same timerange.
	QueryNames         []string                       `json:"query_names"`
	CustomGroupQueries []WebAnalyticsCustomGroupQuery `json:"custom_group_queries"`
	From               int64                          `json:"from"`
	To                 int64                          `json:"to"`
	Timezone           string                         `json:"time_zone"`
}

type WebAnalyticsCustomGroupQuery struct {
	UniqueID          string   `json:"unique_id"`
	GroupByProperties []string `json:"gbp"`
	Metrics           []string `json:"metrics"`
}

type WebAnalyticsCustomGroupMetricValue struct {
	Value     float64
	UniqueMap map[string]bool // For deduplication
}

type WebAnalyticsCustomGroupMetric struct {
	GroupValues []interface{}                                  // Original values of group key
	MetricValue map[string]*WebAnalyticsCustomGroupMetricValue // Map[metric]value
}

type WebAnalyticsCustomGroupPrevGroup struct {
	GroupKey  string
	Timestamp int64
	Value     *float64
}

type WebAnalyticsQueryResult struct {
	QueryResult            *map[string]GenericQueryResult
	CustomGroupQueryResult map[string]*GenericQueryResult
}

const DefaultDashboardWebsiteAnalytics = "Website Analytics"

type WebAnalyticsCacheResult struct {
	Result      *WebAnalyticsQueryResult `json:"result"`
	From        int64                    `json:"from"`
	To          int64                    `json:"tom"`
	Timezone    string                   `json:"timezone"`
	RefreshedAt int64                    `json:"refreshed_at"`
}

// Named queries for website
const (
	QueryNameSessions           = "sessions"
	QueryNameTotalPageViews     = "total_page_view"
	QueryNameBounceRate         = "bounce_rate"
	QueryNameUniqueUsers        = "unique_users"
	QueryNameAvgSessionDuration = "avg_session_duration"
	QueryNameAvgPagesPerSession = "avg_pages_per_session"

	QueryNameTopPagesReport       = "top_pages_report"
	QueryNameTrafficChannelReport = "traffic_channel_report"
)

// DefaultWebAnalyticsQueries -  Named queries and corresponding presentation.
var DefaultWebAnalyticsQueries = map[string]string{
	QueryNameSessions:             PresentationCard,
	QueryNameTotalPageViews:       PresentationCard,
	QueryNameBounceRate:           PresentationCard,
	QueryNameUniqueUsers:          PresentationCard,
	QueryNameAvgSessionDuration:   PresentationCard,
	QueryNameAvgPagesPerSession:   PresentationCard,
	QueryNameTopPagesReport:       PresentationTable,
	QueryNameTrafficChannelReport: PresentationTable,
}

func getWebAnalyticsQueryResultCacheKey(projectID uint64, dashboardID int64,
	from, to int64, timezoneString U.TimeZoneString) (*cacheRedis.Key, error) {

	prefix := "dashboard:query:web"
	var suffix string
	if U.IsStartOfTodaysRangeIn(from, timezoneString) {
		// Query for today's dashboard. Use to as 'now'.
		suffix = fmt.Sprintf("did:%d:from:%d:to:now", dashboardID, from)
	} else if U.Is30MinutesTimeRange(from, to) {
		// Query for last 30mins.
		suffix = fmt.Sprintf("did:%d:30mins", dashboardID)
	} else {
		suffix = fmt.Sprintf("did:%d:from:%d:to:%d", dashboardID, from, to)
	}
	return cacheRedis.NewKey(projectID, prefix, suffix)
}

var SkippableWindows = map[string]int64{
	"2MIN": 120,
}

// GetFormattedTime - Converts seconds into hh mm ss format.
func GetFormattedTime(totalSeconds float64) string {
	var fmtTime string

	totalSecondsInInt := int64(totalSeconds)
	paramHours := totalSecondsInInt / 3600
	if paramHours >= 1 {
		fmtTime = fmt.Sprintf("%dh", paramHours)
	}

	paramMinutes := (totalSecondsInInt % 3600) / 60
	if paramMinutes > 0 {
		if fmtTime != "" {
			fmtTime = fmtTime + " "
		}

		fmtTime = fmtTime + fmt.Sprintf("%dm", paramMinutes)
	}

	paramSeconds := totalSecondsInInt % 60
	if paramSeconds > 0 {
		if fmtTime != "" {
			fmtTime = fmtTime + " "
		}

		fmtTime = fmtTime + fmt.Sprintf("%ds", paramSeconds)
	}

	// Use only milliseconds, if no other params available.
	paramMilliSeconds := int64(totalSeconds*1000) % 1000
	if totalSecondsInInt < 1 && paramMilliSeconds > 0 {
		fmtTime = fmt.Sprintf("%dms", paramMilliSeconds)
	}

	if totalSecondsInInt == 0 && paramMilliSeconds == 0 {
		fmtTime = "0s" // In seconds, intentional.
	}

	return fmtTime
}

const (
	WAGroupMetricPageViews = "page_views"
	// WAGroupMetricPageViewsContributionPercentage page_views per group / total page views
	WAGroupMetricPageViewsContributionPercentage = "page_views_contribution_percentage"
	WAGroupMetricUniqueUsers                     = "unique_users"
	WAGroupMetricUniqueSessions                  = "unique_sessions"
	WAGroupMetricUniquePages                     = "unique_pages"
	WAGroupMetricTotalTimeSpent                  = "total_time_spent"
	WAGroupMetricAvgTimeSpent                    = "avg_time_spent"
	WAGroupMetricTotalScrollDepth                = "total_scroll_depth"
	WAGroupMetricAvgScrollDepth                  = "avg_scroll_depth"
	WAGroupMetricTotalExits                      = "total_exits"
	WAGroupMetricExitPercentage                  = "exit_percentage"
)

// WebAnalyticsCachePayload Payload for web analytics cache method.
type WebAnalyticsCachePayload struct {
	ProjectID   uint64
	DashboardID int64
	From, To    int64
	Timezone    U.TimeZoneString
	Queries     *WebAnalyticsQueries
}

type DashboardUnitWebAnalyticsQueryName struct {
	UnitID    uint64 `json:"unit_id"`
	QueryName string `json:"query_name"`
}

type DashboardUnitWebAnalyticsCustomGroupQuery struct {
	UnitID            uint64   `json:"unit_id"`
	Metrics           []string `json:"metrics"`
	GroupByProperties []string `json:"gbp"`
}

// This can also depend on timezone.
type DashboardUnitsWebAnalyticsQuery struct {
	Class string `json:"cl"`
	// Units - Supports redundant metric keys with different unit_ids.
	Units []DashboardUnitWebAnalyticsQueryName `json:"units"`
	// CustomGroupUnits - Customize query with group by properties and metrics.
	CustomGroupUnits []DashboardUnitWebAnalyticsCustomGroupQuery `json:"custom_group_units"`
	From             int64                                       `json:"from"`
	To               int64                                       `json:"to"`
	Timezone         string                                      `json:"time_zone"`
}

func (q *DashboardUnitsWebAnalyticsQuery) GetClass() string {
	if q.Class == "" {
		q.Class = QueryClassWeb
	}
	return q.Class
}

func (q *DashboardUnitsWebAnalyticsQuery) GetQueryDateRange() (from, to int64) {
	return q.From, q.To
}

func (q *DashboardUnitsWebAnalyticsQuery) SetTimeZone(timezoneString U.TimeZoneString) {
	q.Timezone = string(timezoneString)
}

func (q *DashboardUnitsWebAnalyticsQuery) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(q.Timezone)
}

func (q *DashboardUnitsWebAnalyticsQuery) SetQueryDateRange(from, to int64) {
	q.From, q.To = from, to
}

func (q *DashboardUnitsWebAnalyticsQuery) GetQueryCacheHashString() (string, error) {
	queryMap, err := U.EncodeStructTypeToMap(q)
	if err != nil {
		return "", err
	}
	delete(queryMap, "from")
	delete(queryMap, "to")

	queryHash, err := U.GenerateHashStringForStruct(queryMap)
	if err != nil {
		return "", err
	}
	return queryHash, nil
}

func (q *DashboardUnitsWebAnalyticsQuery) GetQueryCacheRedisKey(projectID uint64) (*cacheRedis.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.From, q.To, U.TimeZoneString(q.Timezone))
	return cacheRedis.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *DashboardUnitsWebAnalyticsQuery) GetQueryCacheExpiry() float64 {
	return getQueryCacheResultExpiry(q.From, q.To, q.Timezone)
}

func (q *DashboardUnitsWebAnalyticsQuery) TransformDateTypeFilters() error {
	return nil
}

func (q *DashboardUnitsWebAnalyticsQuery) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	return nil
}

func (query *DashboardUnitsWebAnalyticsQuery) CheckIfNameIsPresent(nameOfQuery string) bool {
	return false
}

func GetCacheResultForWebAnalyticsDashboard(projectID uint64, dashboardID int64,
	from, to int64, timezoneString U.TimeZoneString) (WebAnalyticsCacheResult, int) {

	var cacheResult WebAnalyticsCacheResult
	if shouldSkipWindow(from, to) {
		return cacheResult, http.StatusNotFound
	}

	logCtx := log.WithFields(log.Fields{
		"Method":      "GetCacheResultForWebAnalyticsDashboard",
		"ProjectID":   projectID,
		"DashboardID": dashboardID,
	})

	if projectID == 0 || dashboardID == 0 {
		return cacheResult, http.StatusBadRequest
	}

	cacheKey, err := getWebAnalyticsQueryResultCacheKey(projectID, dashboardID, from, to, timezoneString)
	if err != nil {
		logCtx.WithError(err).Error("Error getting cache key")
		return cacheResult, http.StatusInternalServerError
	}

	result, err := cacheRedis.GetPersistent(cacheKey)
	if err == redis.ErrNil {
		return cacheResult, http.StatusNotFound
	} else if err != nil {
		logCtx.WithError(err).Error("Error getting key from redis")
		return cacheResult, http.StatusInternalServerError
	}

	err = json.Unmarshal([]byte(result), &cacheResult)
	if err != nil {
		logCtx.WithError(err).Errorf("Error decoding redis result %v", result)
		return cacheResult, http.StatusInternalServerError
	}

	if cacheResult.RefreshedAt == 0 {
		cacheResult.RefreshedAt = U.TimeNowIn(U.TimeZoneStringIST).Unix()
	}
	return cacheResult, http.StatusFound
}

func shouldSkipWindow(from, to int64) bool {
	window := to - from
	for _, definedWindow := range SkippableWindows {
		if window == definedWindow {
			return true
		}
	}
	return false
}

func SetCacheResultForWebAnalyticsDashboard(result *WebAnalyticsQueryResult,
	projectID uint64, dashboardID int64, from, to int64, timezoneString U.TimeZoneString) {

	if shouldSkipWindow(from, to) {
		return
	}

	logCtx := log.WithFields(log.Fields{
		"Method":      "SetCacheResultForWebAnalyticsDashboard",
		"ProjectID":   projectID,
		"DashboardID": dashboardID,
	})

	if projectID == 0 || dashboardID == 0 {
		return
	}

	cacheKey, err := getWebAnalyticsQueryResultCacheKey(projectID, dashboardID, from, to, timezoneString)
	if err != nil {
		logCtx.WithError(err).Error("Error getting cache key for web analytics dashboard")
	}
	dashboardCacheResult := WebAnalyticsCacheResult{
		Result:      result,
		From:        from,
		To:          to,
		Timezone:    string(timezoneString),
		RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
	}

	dashboardCacheResultJSON, err := json.Marshal(&dashboardCacheResult)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode dashboardCacheResult")
		return
	}

	err = cacheRedis.SetPersistent(cacheKey, string(dashboardCacheResultJSON), U.GetDashboardCacheResultExpiryInSeconds(from, to, timezoneString))
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache for channel query")
		return
	}
}
