package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
)

type WebAnalyticsQueries struct {
	// Multiple queries  with same timerange.
	QueryNames         []string                       `json:"query_names"`
	CustomGroupQueries []WebAnalyticsCustomGroupQuery `json:"custom_group_queries"`
	From               int64                          `json:"from"`
	To                 int64                          `json:"to"`
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

type WebAnalyticsQueryResult struct {
	QueryResult            *map[string]GenericQueryResult
	CustomGroupQueryResult map[string]*GenericQueryResult
}

// NamedQueryUnit - Query structure for dashboard unit.
type NamedQueryUnit struct {
	Class     string `json:"cl"`
	Type      string `json:"type"`
	QueryName string `json:"qname"`
}

const QueryTypeNamed = "named_query"
const QueryTypeWebAnalyticsCustomGroupQuery = "wa_custom_group_query"

const DefaultDashboardWebsiteAnalytics = "Website Analytics"
const topPageReportLimit = 50
const defaultPrecision = 1
const customGroupResultLimit = 200

// Named queries for website analytics.
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

const (
	WAGroupMetricPageViews = "page_views"
	// WAGroupMetricPageViewsContributionPercentage page_views per group / total page views
	WAGroupMetricPageViewsContributionPercentage = "page_views_contribution_percentage"
	WAGroupMetricUniqueUsers                     = "unique_users"
	WAGroupMetricUniqueSessions                  = "unique_sessions"
	WAGroupMetricTotalTimeSpent                  = "total_time_spent"
	WAGroupMetricAvgTimeSpent                    = "avg_time_spent"
	WAGroupMetricTotalScrollDepth                = "total_scroll_depth"
	WAGroupMetricAvgScrollDepth                  = "avg_scroll_depth"
	WAGroupMetricExitPercentage                  = "exit_percentage"
)

var SkippableWindows = map[string]int64{
	"2MIN":  120,
	"30MIN": 1800,
}

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

type WebAnalyticsEvent struct {
	ID         string
	ProjectID  uint64
	UserID     string // coalsced user_id
	IsSession  bool
	SessionID  string
	Properties *WebAnalyticsEventProperties
}

// WebAnalyticsEventProperties - Event Properties for web analytics.
type WebAnalyticsEventProperties struct {
	PageURL           string
	PageSpentTime     string
	PageScrollPercent string
	Source            string
	Medium            string
	ReferrerDomain    string
	Campaign          string
	CampaignID        string
	Channel           string // Channel - Added at runtime.

	// session properties.
	SessionSpentTime             string
	SessionPageCount             string
	SessionInitialPageURL        string
	SessionInitialReferrerDomain string
	SessionLatestPageURL         string
}

// WebAnalyticsResultAggregate - Supporting aggregates for
// calculating different metrics.
type WebAnalyticsGeneralAggregate struct {
	NoOfSessions        int
	NoOfPageViews       int
	NoOfBouncedSessions int     // no.of sessions with $page_count = 1.
	SessionDuration     float64 // sum of $session_spent_time of session event.
	SessionPages        float64 // sum of $page_count of session event.

	NoOfUniqueUsers int // no.of coalsced users.
	UniqueUsersMap  map[string]bool
}

type WebAnalyticsPageAggregate struct {
	NoOfPageViews        int
	NoOfEntrances        int     // no.of sessions with URL as $initial_page_url
	NoOfExits            int     // no.of sessions with URL as $lastest_page_url
	NoOfBouncedEntrances int     // no.of sessions with URL as $initial_page_url and $page_count as 1.
	TotalSpentTime       float64 // sum of $page_spent_time

	NoOfUniqueUsers int
	UniqueUsersMap  map[string]bool
}

type WebAnalyticsChannelAggregate struct {
	NoOfPageViews       int
	NoOfSessions        int     // no.of sessions
	NoOfBouncedSessions int     // no.of sessions with $page_count = 1.
	SessionDuration     float64 // sum of $session_spent_time of session event with

	NoOfUniqueUsers int
	UniqueUsersMap  map[string]bool
}

type WebAnalyticsAggregate struct {
	WebAnalyticsGeneralAggregate
	PageAggregates    map[string]*WebAnalyticsPageAggregate
	ChannelAggregates map[string]*WebAnalyticsChannelAggregate
}

type WebAnalyticsCacheResult struct {
	Result      *WebAnalyticsQueryResult `json:"result"`
	From        int64                    `json:"from"`
	To          int64                    `json:"tom"`
	RefreshedAt int64                    `json:"refreshed_at"`
}

func getWebAnalyticsEnabledProjectIDs() ([]uint64, int) {
	db := C.GetServices().Db

	var projectIDs []uint64
	rows, err := db.Raw("SELECT distinct(project_id) FROM dashboards WHERE name = ?",
		DefaultDashboardWebsiteAnalytics).Rows()
	if err != nil {
		log.WithError(err).Error("Error getting web analytics enabled project ids")
		return projectIDs, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var projectID uint64
		if err := rows.Scan(&projectID); err != nil {
			log.WithError(err).Error("Error scanning web analytics enabled project ids")
			return projectIDs, http.StatusInternalServerError
		}
		projectIDs = append(projectIDs, projectID)
	}

	if len(projectIDs) == 0 {
		return projectIDs, http.StatusNotFound
	}
	return projectIDs, http.StatusFound
}

func getWebAnalyticsQueriesFromDashboardUnits(projectID uint64) (uint64, *WebAnalyticsQueries, int) {
	logCtx := log.WithFields(log.Fields{
		"Method":    "GetWebAnalyticsDashboardIDForProject",
		"ProjectID": projectID,
	})

	var dashboardUnits []DashboardUnit

	db := C.GetServices().Db
	err := db.
		Where("project_id = ? AND query->>'cl' = ?", projectID, QueryClassWeb).
		Find(&dashboardUnits).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get web analytics queries from dashboard units.")
		return 0, nil, http.StatusInternalServerError
	}

	if len(dashboardUnits) == 0 {
		return 0, nil, http.StatusNotFound
	}

	// Build web analytics queries from dashboard units.
	namedQueries := make([]string, 0, 0)
	customGroupQueries := make([]WebAnalyticsCustomGroupQuery, 0, 0)
	for _, dunit := range dashboardUnits {
		queryMap, err := U.DecodePostgresJsonb(&dunit.Query)
		if err != nil {
			logCtx.WithError(err).Error("Failed to decode web analytics dashboard unit query.")
			continue
		}

		queryTypeInf, exists := (*queryMap)["type"]
		if !exists {
			logCtx.WithError(err).
				Error("Invalid web analytics query type on dashoard unit")
			continue
		}

		queryType := queryTypeInf.(string)
		if queryType == QueryTypeNamed {
			var query NamedQueryUnit
			if err := U.DecodePostgresJsonbToStructType(&dunit.Query, &query); err != nil {
				logCtx.WithError(err).
					Error("Failed to decode named query from dashboard unit.")
				continue
			}
			namedQueries = append(namedQueries, query.QueryName)

		} else if queryType == QueryTypeWebAnalyticsCustomGroupQuery {
			var query WebAnalyticsCustomGroupQuery
			if err := U.DecodePostgresJsonbToStructType(&dunit.Query, &query); err != nil {
				logCtx.WithError(err).
					Error("Failed to decode custom group query from dashboard unit.")
				continue
			}
			// Use dashboard unit id as unique id for query.
			query.UniqueID = fmt.Sprintf("%d", dunit.ID)
			customGroupQueries = append(customGroupQueries, query)

		} else {
			logCtx.WithError(err).
				Error("Invalid web analytics query type on dashoard unit")
			continue
		}
	}

	// Todo: Return all dashboard ids which has web
	// analytics unit, for caching.
	return dashboardUnits[0].DashboardId,
		&WebAnalyticsQueries{
			QueryNames:         namedQueries,
			CustomGroupQueries: customGroupQueries,
		},
		http.StatusFound
}

func getWebAnalyticsQueryResultCacheKey(projectID, dashboardID uint64,
	from, to int64) (*cacheRedis.Key, error) {

	prefix := "dashboard:query:web"
	var suffix string
	if from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) {
		// Query for today's dashboard. Use to as 'now'.
		suffix = fmt.Sprintf("did:%d:from:%d:to:now", dashboardID, from)
	} else {
		suffix = fmt.Sprintf("did:%d:from:%d:to:%d", dashboardID, from, to)
	}
	return cacheRedis.NewKey(projectID, prefix, suffix)
}

func isWebAnalyticsDashboardAlreadyCached(projectID, dashboardID uint64, from, to int64) bool {
	if from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) {
		// If from time is of today's beginning, refresh today's everytime a request is received.
		return false
	}
	cacheKey, err := getWebAnalyticsQueryResultCacheKey(projectID, dashboardID, from, to)
	if err != nil {
		log.WithError(err).Errorf("Failed to get cache key")
		return false
	}
	exists, err := cacheRedis.ExistsPersistent(cacheKey)
	if err != nil {
		log.WithError(err).Errorf("Redis error on exists")
		return false
	}
	return exists
}

func getQueryForNamedQueryUnit(class, queryName string) (*postgres.Jsonb, error) {
	return U.EncodeStructTypeToPostgresJsonb(NamedQueryUnit{Class: class,
		Type: QueryTypeNamed, QueryName: queryName})
}

func addWebAnalyticsDefaultDashboardUnits(projectId uint64,
	agentUUID string, dashboardId uint64) int {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "agent_uuid": agentUUID,
		"dashboard_id": dashboardId})

	hasFailure := false
	for queryName, presentation := range DefaultWebAnalyticsQueries {
		queryJsonb, err := getQueryForNamedQueryUnit(QueryClassWeb, queryName)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to named query on add web analytics dashboard units.")
			return http.StatusInternalServerError
		}

		_, errCode, errMsg := CreateDashboardUnit(projectId, agentUUID,
			&DashboardUnit{
				DashboardId:  dashboardId,
				Title:        U.GetSnakeCaseToTitleString(queryName),
				Presentation: presentation,
				Query:        *queryJsonb,
			})

		if errCode != http.StatusCreated {
			logCtx.WithField("err_msg", errMsg).WithField("query_name", queryName).
				Error("Failed to add web analytics dashboard unit.")
			hasFailure = true
		}
	}

	if hasFailure {
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func CreateWebAnalyticsDefaultDashboardWithUnits(projectId uint64, agentUUID string) int {
	logCtx := log.WithField("project_id", projectId).WithField("agent_uuid", agentUUID)

	dashboard, errCode := CreateDashboard(projectId, agentUUID, &Dashboard{
		Name: DefaultDashboardWebsiteAnalytics,
		Type: DashboardTypeProjectVisible,
	})
	if errCode != http.StatusCreated {
		logCtx.Error("Failed to create web analytics default dashboard.")
		return errCode
	}

	errCode = addWebAnalyticsDefaultDashboardUnits(projectId, agentUUID, dashboard.ID)
	if errCode != http.StatusCreated {
		return errCode
	}

	return http.StatusCreated
}

func hasCampaign(campaign, campaignID string) bool {
	return campaign != "" || campaignID != ""
}

func isSearchReferrer(referrerDomain string) bool {
	referrerDomain = strings.ToLower(referrerDomain)

	return U.IsContainsAnySubString(referrerDomain, "google", "bing")
}

func isSocialReferrer(referrerDomain string) bool {
	referrerDomain = strings.ToLower(referrerDomain)

	return U.IsContainsAnySubString(referrerDomain,
		"facebook", "twitter", "linkedin", "instagram", "tiktok")
}

func getChannel(wep *WebAnalyticsEventProperties, isSessionEvent bool) string {
	var referrerDomain string
	if isSessionEvent {
		referrerDomain = wep.SessionInitialReferrerDomain
	} else {
		// Properties like source, referrer and medium
		// are using the same name as event property on session event.
		referrerDomain = wep.ReferrerDomain
	}

	// Property values cleanup.
	wep.Medium = strings.ToLower(wep.Medium)
	referrerDomain = strings.ToLower(referrerDomain)

	// NOTE: Order of conditions is CRITICAL.
	// Ensure any change on the order is
	// planned and reviewed before.

	if !hasCampaign(wep.Campaign, wep.CampaignID) &&
		wep.Source == "" && referrerDomain == "" && wep.Medium == "" {
		return "Direct"
	}

	if !hasCampaign(wep.Campaign, wep.CampaignID) &&
		isSearchReferrer(referrerDomain) {
		return "Organic Search"
	}

	if hasCampaign(wep.Campaign, wep.CampaignID) &&
		isSearchReferrer(referrerDomain) {
		return "Paid Search"
	}

	if !hasCampaign(wep.Campaign, wep.CampaignID) &&
		isSocialReferrer(referrerDomain) {
		return "Organic Social"
	}

	if hasCampaign(wep.Campaign, wep.CampaignID) &&
		isSocialReferrer(referrerDomain) {
		return "Paid Social"
	}

	if wep.Medium == "referral" {
		return "Referral"
	}

	if wep.Medium == "email" {
		return "Email"
	}

	if wep.Medium == "affiliate" {
		return "Affiliates"
	}

	return "Other"
}

// getUniqueGroupKeyAndPropertyValues - Returns key aware unique
// string as group_key and the original values of properties.
func getUniqueGroupKeyAndPropertyValues(queryPropertyKeys []string,
	propertyKeyValues *map[string]string) (string, []interface{}) {

	var groupKey string
	groupPropertyValues := make([]interface{}, 0, 0)
	for i := range queryPropertyKeys {
		var propertyValue string
		value, exists := (*propertyKeyValues)[queryPropertyKeys[i]]
		if exists && value != "" && value != "NULL" {
			propertyValue = value
		} else {
			propertyValue = PropertyValueNone
		}

		if groupKey != "" {
			// Append key:value set seperator.
			groupKey = groupKey + "||"
		}

		groupKey = groupKey + fmt.Sprintf("%s:%s", queryPropertyKeys[i], propertyValue)
		groupPropertyValues = append(groupPropertyValues, propertyValue)
	}

	return groupKey, groupPropertyValues
}

func initWACustomGroupMetricValue(m *map[string]*WebAnalyticsCustomGroupMetricValue,
	metricName string) {

	if metricValue, exists := (*m)[metricName]; !exists || metricValue == nil {
		(*m)[metricName] = &WebAnalyticsCustomGroupMetricValue{
			UniqueMap: map[string]bool{},
		}
	}
}

func buildWebAnalyticsAggregateForPageEvent(
	webEvent *WebAnalyticsEvent,
	aggrState *WebAnalyticsAggregate,
	customGroupQueries []WebAnalyticsCustomGroupQuery,
	customGroupPropertiesMap *map[string]string,
	customGroupAggrState *map[string]map[string]*WebAnalyticsCustomGroupMetric,
) int {

	logCtx := log.WithField("project_id", webEvent.ProjectID).
		WithField("event_id", webEvent.ID)

	pageURL := webEvent.Properties.PageURL
	if pageURL == "" {
		logCtx.Error("Missing page_url property on event.")
		return http.StatusInternalServerError
	}

	if _, exists := aggrState.PageAggregates[pageURL]; !exists {
		aggrState.PageAggregates[pageURL] = &WebAnalyticsPageAggregate{}
	}

	aggrState.NoOfPageViews++

	if _, exists := aggrState.UniqueUsersMap[webEvent.UserID]; !exists {
		aggrState.UniqueUsersMap[webEvent.UserID] = true
		aggrState.NoOfUniqueUsers++
	}

	aggrState.PageAggregates[pageURL].NoOfPageViews++

	pageSpentTime, err := U.GetPropertyValueAsFloat64(webEvent.Properties.PageSpentTime)
	if err != nil {
		logCtx.WithError(err).
			Error("Failed converting page_spent_time property value to float64.")
	}
	aggrState.PageAggregates[pageURL].TotalSpentTime += pageSpentTime

	pageScrollPercent, err := U.GetPropertyValueAsFloat64(webEvent.Properties.PageScrollPercent)
	if err != nil {
		logCtx.WithError(err).
			Error("Failed converting page_scroll_percent property value to float64.")
	}

	if aggrState.PageAggregates[pageURL].UniqueUsersMap == nil {
		aggrState.PageAggregates[pageURL].UniqueUsersMap = make(map[string]bool)
	}

	if _, exists := aggrState.PageAggregates[pageURL].UniqueUsersMap[webEvent.UserID]; !exists {
		aggrState.PageAggregates[pageURL].UniqueUsersMap[webEvent.UserID] = true
		aggrState.PageAggregates[pageURL].NoOfUniqueUsers++
	}

	aggrState.ChannelAggregates[webEvent.Properties.Channel].NoOfPageViews++

	if aggrState.ChannelAggregates[webEvent.Properties.Channel].UniqueUsersMap == nil {
		aggrState.ChannelAggregates[webEvent.Properties.Channel].UniqueUsersMap = make(map[string]bool)
	}

	if _, exists := aggrState.ChannelAggregates[webEvent.Properties.Channel].
		UniqueUsersMap[webEvent.UserID]; !exists {

		aggrState.ChannelAggregates[webEvent.Properties.Channel].UniqueUsersMap[webEvent.UserID] = true
		aggrState.ChannelAggregates[webEvent.Properties.Channel].NoOfUniqueUsers++
	}

	for i := range customGroupQueries {
		query := customGroupQueries[i]

		groupKey, groupPropertyValues := getUniqueGroupKeyAndPropertyValues(
			query.GroupByProperties, customGroupPropertiesMap)

		if _, exists := (*customGroupAggrState)[query.UniqueID][groupKey]; !exists {
			(*customGroupAggrState)[query.UniqueID][groupKey] = &WebAnalyticsCustomGroupMetric{
				GroupValues: groupPropertyValues,
				MetricValue: map[string]*WebAnalyticsCustomGroupMetricValue{},
			}
		}

		// Add page_views metric by default.
		initWACustomGroupMetricValue(
			&(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue,
			WAGroupMetricPageViews,
		)
		(*customGroupAggrState)[query.UniqueID][groupKey].
			MetricValue[WAGroupMetricPageViews].Value++

		for _, metric := range query.Metrics {
			if metric == WAGroupMetricUniqueUsers {
				initWACustomGroupMetricValue(
					&(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue,
					metric,
				)

				_, exists := (*customGroupAggrState)[query.UniqueID][groupKey].
					MetricValue[metric].UniqueMap[webEvent.UserID]
				if !exists {
					// Count and add unique users to unique map for deduplication.
					(*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[metric].Value++

					(*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[metric].UniqueMap[webEvent.UserID] = true
				}
			}

			if metric == WAGroupMetricUniqueSessions {
				initWACustomGroupMetricValue(
					&(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue,
					metric,
				)

				_, exists := (*customGroupAggrState)[query.UniqueID][groupKey].
					MetricValue[metric].UniqueMap[webEvent.SessionID]
				if !exists {
					// Count and add unique sessions to unique map for deduplication.
					(*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[metric].Value++

					(*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[metric].UniqueMap[webEvent.SessionID] = true
				}
			}

			if metric == WAGroupMetricTotalTimeSpent || metric == WAGroupMetricAvgTimeSpent {
				initWACustomGroupMetricValue(
					&(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue,
					WAGroupMetricTotalTimeSpent,
				)

				(*customGroupAggrState)[query.UniqueID][groupKey].
					MetricValue[WAGroupMetricTotalTimeSpent].Value += pageSpentTime
			}

			if metric == WAGroupMetricTotalScrollDepth || metric == WAGroupMetricAvgScrollDepth {
				initWACustomGroupMetricValue(
					&(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue,
					WAGroupMetricTotalScrollDepth,
				)

				(*customGroupAggrState)[query.UniqueID][groupKey].
					MetricValue[WAGroupMetricTotalScrollDepth].Value += pageScrollPercent
			}
		}
	}

	return http.StatusOK
}

func buildWebAnalyticsAggregateForSessionEvent(
	webEvent *WebAnalyticsEvent,
	aggrState *WebAnalyticsAggregate,
	customGroupQueries []WebAnalyticsCustomGroupQuery,
	customGroupPropertiesMap *map[string]string,
	customGroupAggrState *map[string]map[string]*WebAnalyticsCustomGroupMetric,
) int {

	logCtx := log.WithField("project_id", webEvent.ProjectID).
		WithField("event_id", webEvent.ID)

	aggrState.NoOfSessions++
	aggrState.ChannelAggregates[webEvent.Properties.Channel].NoOfSessions++

	sessionSpentTime, err := U.GetPropertyValueAsFloat64(webEvent.Properties.SessionSpentTime)
	if err != nil {
		logCtx.WithError(err).
			Error("Failed converting session_spent_time property value to float64.")
	}
	aggrState.SessionDuration += sessionSpentTime
	aggrState.ChannelAggregates[webEvent.Properties.Channel].SessionDuration += sessionSpentTime

	sessionPageCount, err := U.GetPropertyValueAsFloat64(webEvent.Properties.SessionPageCount)
	if err != nil {
		logCtx.WithError(err).
			Error("Failed converting session_page_count property value to float64.")
	}
	aggrState.SessionPages += sessionPageCount

	if sessionPageCount == 1 {
		aggrState.NoOfBouncedSessions++
		aggrState.ChannelAggregates[webEvent.Properties.Channel].NoOfBouncedSessions++
	}

	sessionInitialPageURL := webEvent.Properties.SessionInitialPageURL
	if sessionInitialPageURL == "" {
		logCtx.Error("Missing $initial_page_url on session on build_web_analytic_aggregate.")
		return http.StatusInternalServerError
	}

	if _, exists := aggrState.PageAggregates[sessionInitialPageURL]; !exists {
		aggrState.PageAggregates[sessionInitialPageURL] = &WebAnalyticsPageAggregate{}
	}

	aggrState.PageAggregates[sessionInitialPageURL].NoOfEntrances++

	if sessionPageCount == 1 {
		aggrState.PageAggregates[sessionInitialPageURL].NoOfBouncedEntrances++
	}

	sessionLatestPageURL := webEvent.Properties.SessionLatestPageURL
	if sessionLatestPageURL != "" &&
		sessionInitialPageURL != "" &&
		sessionInitialPageURL == sessionLatestPageURL {

		aggrState.PageAggregates[sessionInitialPageURL].NoOfExits++
	}

	return http.StatusOK
}

// buildWebAnalyticsAggregate - Builds aggregate with each event sent as param.
func buildWebAnalyticsAggregate(
	webEvent *WebAnalyticsEvent,
	aggrState *WebAnalyticsAggregate,
	customGroupQueries []WebAnalyticsCustomGroupQuery,
	customGroupPropertiesMap *map[string]string,
	customGroupAggrState *map[string]map[string]*WebAnalyticsCustomGroupMetric,
) int {

	if aggrState.UniqueUsersMap == nil {
		aggrState.UniqueUsersMap = make(map[string]bool)
	}

	if aggrState.PageAggregates == nil {
		aggrState.PageAggregates = make(map[string]*WebAnalyticsPageAggregate)
	}

	if aggrState.ChannelAggregates == nil {
		aggrState.ChannelAggregates = make(map[string]*WebAnalyticsChannelAggregate)
	}

	if _, exists := aggrState.ChannelAggregates[webEvent.Properties.Channel]; !exists {
		aggrState.ChannelAggregates[webEvent.Properties.Channel] =
			&WebAnalyticsChannelAggregate{}
	}

	for i := range customGroupQueries {
		groupQuery := customGroupQueries[i]
		if _, exists := (*customGroupAggrState)[groupQuery.UniqueID]; !exists {
			(*customGroupAggrState)[groupQuery.UniqueID] =
				make(map[string]*WebAnalyticsCustomGroupMetric, 0)
		}
	}

	if !webEvent.IsSession {
		return buildWebAnalyticsAggregateForPageEvent(webEvent, aggrState,
			customGroupQueries, customGroupPropertiesMap, customGroupAggrState)
	}

	return buildWebAnalyticsAggregateForSessionEvent(webEvent, aggrState,
		customGroupQueries, customGroupPropertiesMap, customGroupAggrState)
}

func getTopPagesReportAsWebAnalyticsResult(
	webAggr *WebAnalyticsAggregate) GenericQueryResult {

	headers := []string{
		"Page URL",
		"Page Views",
		"Unique Users",
		"Avg Time on Page",
		"Entrances",
		"Exits",
		"Bounced Entrances",
	}

	rows := make([][]interface{}, 0, len(webAggr.PageAggregates))
	for url, aggr := range webAggr.PageAggregates {
		var fmtAvgPageSpentTimeOfPage string
		if aggr.NoOfPageViews > 0 {
			avgPageSpentTimeOfPage, err := U.FloatRoundOffWithPrecision(
				aggr.TotalSpentTime/float64(aggr.NoOfPageViews), defaultPrecision)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"total_page_spent_time": aggr.TotalSpentTime,
						"no_of_page_views": aggr.NoOfPageViews}).
					Error("Failed to format average spent time on getTopPagesReportAsWebAnalyticsResult.")

			}

			fmtAvgPageSpentTimeOfPage = GetFormattedTime(avgPageSpentTimeOfPage)
		}

		row := []interface{}{
			url,
			aggr.NoOfPageViews,
			aggr.NoOfUniqueUsers,
			fmtAvgPageSpentTimeOfPage,
			aggr.NoOfEntrances,
			aggr.NoOfExits,
			aggr.NoOfBouncedEntrances,
		}
		rows = append(rows, row)
	}

	// sort by NoOfPageViews.
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][1].(int) > rows[j][1].(int)
	})

	var rowsLimit int
	if len(rows) < topPageReportLimit {
		rowsLimit = len(rows)
	} else {
		rowsLimit = topPageReportLimit
	}

	return GenericQueryResult{Headers: headers, Rows: rows[:rowsLimit]}
}

func getTrafficChannelReport(webAggr *WebAnalyticsAggregate) GenericQueryResult {
	headers := []string{
		"Channel",
		"Page Views",
		"Unique Users",
		"Sessions",
		"Bounce Rate",
		"Avg Session Duration",
	}

	rows := make([][]interface{}, 0, len(webAggr.ChannelAggregates))
	for channel, aggr := range webAggr.ChannelAggregates {
		var avgSessionDurationInSecs float64
		var bounceRateAsFloat float64
		var err error

		if aggr.NoOfSessions > 0 {
			avgSessionDurationInSecs, err = U.FloatRoundOffWithPrecision(
				aggr.SessionDuration/float64(aggr.NoOfSessions), defaultPrecision)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"total_session_duration": aggr.SessionDuration,
						"no_of_sessions": aggr.NoOfSessions}).
					Error("Failed to format avg session duration on getTrafficChannelReport.")
			}

			bounceRateAsFloat = (float64(aggr.NoOfBouncedSessions) / float64(aggr.NoOfSessions)) * 100
		}
		// Formatted value string.
		avgSessionDuration := GetFormattedTime(avgSessionDurationInSecs)
		bounceRateAsPercentage := getFormattedPercentage(bounceRateAsFloat)

		row := []interface{}{
			channel,
			aggr.NoOfPageViews,
			aggr.NoOfUniqueUsers,
			aggr.NoOfSessions,
			bounceRateAsPercentage,
			avgSessionDuration,
		}
		rows = append(rows, row)
	}

	// sort by NoOfPageViews.
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][1].(int) > rows[j][1].(int)
	})

	return GenericQueryResult{Headers: headers, Rows: rows}
}

// Converts single value aggregate to rows and headers format for compatibility.
func fillValueAsWebAnalyticsResult(queryResultByName *map[string]GenericQueryResult,
	queryName string, value interface{}) {

	webAResult := GenericQueryResult{
		Headers: []string{"count"},
		Rows:    [][]interface{}{[]interface{}{value}},
	}

	(*queryResultByName)[queryName] = webAResult
}

// GetFormattedTime - Converts seconds into hh mm ss format.
func GetFormattedTime(totalSeconds float64) string {
	var fmtTime string

	totalSecondsInInt := int64(totalSeconds)
	if totalSecondsInInt > 3600 {
		fmtTime = fmt.Sprintf("%dh ", totalSecondsInInt/3600)
	}

	if totalSecondsInInt > 60 {
		fmtTime = fmtTime + fmt.Sprintf("%dm ", (totalSecondsInInt%3600)/60)
	}

	if totalSecondsInInt > 0 {
		fmtTime = fmtTime + fmt.Sprintf("%ds", totalSecondsInInt%60)
	}

	// Add milliseconds,  only if  total seconds has
	// upto 3 decimal points, which is millseconds.
	millSeconds := int64(totalSeconds*1000) % 1000
	if totalSecondsInInt == 0 && millSeconds > 0 {
		fmtTime = fmtTime + fmt.Sprintf("%dms", millSeconds)
	}

	if totalSecondsInInt == 0 && millSeconds == 0 {
		fmtTime = "0s" // In seconds, intentional.
	}

	return fmtTime
}

// getFormattedPercentage - Creates percentage string with required precision.
func getFormattedPercentage(value float64) string {
	if value == 0 {
		return "0%"
	}
	if value <= 5 {
		return fmt.Sprintf("%0.2f%%", value)
	}

	return fmt.Sprintf("%0.*f%%", defaultPrecision, value)
}

func getWebAnalyticsQueryResultByName(webAggrState *WebAnalyticsAggregate) (
	queryResultByName *map[string]GenericQueryResult) {

	queryResultByName = &map[string]GenericQueryResult{}

	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameSessions, webAggrState.NoOfSessions)
	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameTotalPageViews, webAggrState.NoOfPageViews)

	// Bounce Rate as in percent
	var percentageBouncedSessions float64
	if webAggrState.NoOfSessions > 0 {
		percentageBouncedSessions = float64(webAggrState.NoOfBouncedSessions) /
			float64(webAggrState.NoOfSessions) * 100
	}

	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameBounceRate, getFormattedPercentage(percentageBouncedSessions))

	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameUniqueUsers, webAggrState.NoOfUniqueUsers)

	var avgSessionDuration, avgPagesPerSession float64
	if webAggrState.NoOfSessions > 0 {
		avgSessionDuration = webAggrState.SessionDuration / float64(webAggrState.NoOfSessions)
		avgPagesPerSession = webAggrState.SessionPages / float64(webAggrState.NoOfSessions)
	}

	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameAvgSessionDuration, GetFormattedTime(avgSessionDuration))

	precisionedAvgPagesPerSession, _ :=
		U.FloatRoundOffWithPrecision(avgPagesPerSession, defaultPrecision)
	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameAvgPagesPerSession, precisionedAvgPagesPerSession)

	(*queryResultByName)[QueryNameTopPagesReport] =
		getTopPagesReportAsWebAnalyticsResult(webAggrState)
	(*queryResultByName)[QueryNameTrafficChannelReport] =
		getTrafficChannelReport(webAggrState)

	return queryResultByName
}

// getResultForCustomGroupQuery - Returns result for each unique_id,
// after ordering by first metric value and limiting.
func getResultForCustomGroupQuery(
	groupQueries []WebAnalyticsCustomGroupQuery,
	customGroupAggrState *map[string]map[string]*WebAnalyticsCustomGroupMetric,
	webAggrState *WebAnalyticsAggregate,
) (queryResult map[string]*GenericQueryResult) {

	queryResult = map[string]*GenericQueryResult{}

	for i := range groupQueries {
		query := groupQueries[i]

		headers := make([]string, 0, 0)
		headers = append(headers, query.GroupByProperties...)
		// Add readable metric names.
		for i := range query.Metrics {
			headers = append(headers, U.GetSnakeCaseToTitleString(query.Metrics[i]))
		}

		if _, exists := queryResult[query.UniqueID]; !exists {
			queryResult[query.UniqueID] = &GenericQueryResult{}
		}
		queryResult[query.UniqueID].Headers = headers

		rows := make([][]interface{}, 0, 0)
		for groupKey, group := range (*customGroupAggrState)[query.UniqueID] {
			row := make([]interface{}, 0, 0)
			row = append(row, group.GroupValues...)

			// Page view is available by default.
			pageViews := (*customGroupAggrState)[query.UniqueID][groupKey].
				MetricValue[WAGroupMetricPageViews].Value

			for i := range query.Metrics {
				metric := query.Metrics[i]

				// Add 0 values to metric, if group key doesn't exist.
				if _, exists := (*customGroupAggrState)[query.UniqueID][groupKey]; !exists {
					row = append(row, 0)
					continue
				}

				// Avoid nil pointer exception if metric value is not available,
				// For metrics like WAGroupMetricAvgTimeSpent, we will have only the
				// dependent metric WAGroupMetricTotalTimeSpent instead of the original.
				if (*customGroupAggrState)[query.UniqueID][groupKey].MetricValue[metric] == nil {
					(*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[metric] = &WebAnalyticsCustomGroupMetricValue{}
				}

				var value interface{}
				// Metrics/Aggregates which can be added directly to result.
				if metric == WAGroupMetricPageViews || metric == WAGroupMetricUniqueUsers ||
					metric == WAGroupMetricUniqueSessions {
					value = (*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[metric].Value
				}

				if metric == WAGroupMetricPageViewsContributionPercentage {
					var contributionPercentage float64
					if webAggrState.NoOfPageViews > 0 {
						contributionPercentage = (pageViews / float64(webAggrState.NoOfPageViews)) * 100
					}
					value = getFormattedPercentage(contributionPercentage)
				}

				if metric == WAGroupMetricTotalTimeSpent {
					totalTimeSpent := (*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[WAGroupMetricTotalTimeSpent].Value
					value = GetFormattedTime(totalTimeSpent)
				}

				if metric == WAGroupMetricAvgTimeSpent {
					var avgTimeSpent float64
					if pageViews > 0 && (*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[WAGroupMetricTotalTimeSpent] != nil {

						avgTimeSpent = (*customGroupAggrState)[query.UniqueID][groupKey].
							MetricValue[WAGroupMetricTotalTimeSpent].Value / pageViews
					}

					value = GetFormattedTime(avgTimeSpent)
				}

				if metric == WAGroupMetricTotalScrollDepth {
					totalScrollDepth := (*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[WAGroupMetricTotalScrollDepth].Value
					value = getFormattedPercentage(totalScrollDepth)
				}

				if metric == WAGroupMetricAvgScrollDepth {
					var avgScrollDepth float64
					if pageViews > 0 && (*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[WAGroupMetricTotalScrollDepth] != nil {
						avgScrollDepth = (*customGroupAggrState)[query.UniqueID][groupKey].
							MetricValue[WAGroupMetricTotalScrollDepth].Value / pageViews
					}

					value = getFormattedPercentage(avgScrollDepth)
				}

				row = append(row, value)
			}

			rows = append(rows, row)
		}

		indexOfFirstMetric := len(query.GroupByProperties)
		sort.SliceStable(rows, func(i, j int) bool {
			return U.GetSortWeightFromAnyType(rows[i][indexOfFirstMetric]) >
				U.GetSortWeightFromAnyType(rows[j][indexOfFirstMetric])
		})

		var rowsLimit int
		if len(rows) < customGroupResultLimit {
			rowsLimit = len(rows)
		} else {
			rowsLimit = customGroupResultLimit
		}
		queryResult[query.UniqueID].Rows = rows[:rowsLimit]
	}

	return queryResult
}

// ExecuteWebAnalyticsQueries - executes the web analytics query and returns result by query_name.
func ExecuteWebAnalyticsQueries(projectId uint64, queries *WebAnalyticsQueries) (
	queryResult *WebAnalyticsQueryResult, errCode int) {
	funcStartTimestamp := U.TimeNowUnix()

	webAggrState := WebAnalyticsAggregate{}

	queryResult = &WebAnalyticsQueryResult{}
	queryResult.QueryResult = getWebAnalyticsQueryResultByName(&webAggrState)
	queryResult.CustomGroupQueryResult = make(map[string]*GenericQueryResult, 0)

	// map[unique_id]map[group_key]*WebAnalyticsCustomGroupMetric
	customGroupAggrState := make(map[string]map[string]*WebAnalyticsCustomGroupMetric, 0)

	if projectId == 0 || queries == nil {
		return queryResult, http.StatusBadRequest
	}
	logCtx := log.WithField("project_id", projectId).WithField("query", queries)

	sessionEventName, errCode := GetSessionEventName(projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get session event name on execute_web_analytics_query.")
		return queryResult, http.StatusInternalServerError
	}

	// selectProperties - Columns will be scanned
	// on the same order. Update scanner if order changed.
	selectProperties := []string{
		U.EP_PAGE_URL,
		U.EP_PAGE_SPENT_TIME,
		U.EP_PAGE_SCROLL_PERCENT,
		U.EP_SOURCE,
		U.EP_REFERRER_DOMAIN,
		U.EP_MEDIUM,
		U.EP_CAMPAIGN,
		U.EP_CAMPAIGN_ID,

		U.SP_SPENT_TIME,
		U.SP_PAGE_COUNT,
		U.SP_INITIAL_PAGE_URL,
		U.SP_INITIAL_REFERRER_DOMAIN,
		U.SP_LATEST_PAGE_URL,
	}
	var selectPropertiesStmnt string
	for _, property := range selectProperties {
		if selectPropertiesStmnt != "" {
			selectPropertiesStmnt = selectPropertiesStmnt + ","
		}

		selectPropertiesStmnt = fmt.Sprintf("%s events.properties->>'%s'",
			selectPropertiesStmnt, property)
	}

	var customGroupPropertySelectStmnt string
	var customGroupPropertySelectParams []interface{}
	addedCustomGroupPropertiesMap := make(map[string]bool, 0)

	for _, customQuery := range queries.CustomGroupQueries {
		// Validation
		if customQuery.UniqueID == "" {
			logCtx.Error("Unique id is not assigned to custom group query. Invalid query.")
			return queryResult, http.StatusInternalServerError
		}

		for _, groupByProperty := range customQuery.GroupByProperties {
			// Avoid adding to select if exists already.
			if _, exists := addedCustomGroupPropertiesMap[groupByProperty]; exists {
				continue
			}

			if customGroupPropertySelectStmnt != "" {
				customGroupPropertySelectStmnt = customGroupPropertySelectStmnt + ","
			}

			customGroupPropertySelectStmnt = fmt.Sprintf(
				"%s events.properties->>?",
				customGroupPropertySelectStmnt,
			)

			customGroupPropertySelectParams = append(
				customGroupPropertySelectParams,
				groupByProperty,
			)

			addedCustomGroupPropertiesMap[groupByProperty] = true
		}
	}

	queryParams := make([]interface{}, 0, 0)
	selectStmnt := "events.id, events.project_id," + " " +
		"COALESCE(users.customer_user_id, users.id) as user_id," + " " +
		"events.session_id, events.event_name_id, event_names.name as event_name," + " " +
		"event_names.type as event_name_type," + " " +
		selectPropertiesStmnt

	if customGroupPropertySelectStmnt != "" {
		// Read custom group properties into an array as keys are dynamic.
		customGroupPropertySelectStmnt = "ARRAY[" + customGroupPropertySelectStmnt + "]"

		selectStmnt = selectStmnt + ", " + customGroupPropertySelectStmnt
		queryParams = append(queryParams, customGroupPropertySelectParams...)
	}

	queryStmnt := "SELECT" + " " + selectStmnt + " " +
		"FROM events LEFT JOIN event_names ON events.event_name_id=event_names.id" + " " +
		"LEFT JOIN users ON events.user_id=users.id" + " " +
		"WHERE events.project_id = ? AND events.timestamp BETWEEN ? AND ?" + " " +
		"AND (events.properties->>'$page_raw_url' IS NOT NULL OR events.event_name_id = ?)"
	queryParams = append(queryParams, projectId, queries.From, queries.To, sessionEventName.ID)

	queryStartTimestamp := U.TimeNowUnix()
	db := C.GetServices().Db
	rows, err := db.Raw(queryStmnt, queryParams...).Rows()
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to execute raw query to download events on execute_web_analytics_query.")
		return queryResult, http.StatusInternalServerError
	}
	defer rows.Close()
	logCtx = logCtx.WithField("query_exec_time_in_secs", U.TimeNowUnix()-queryStartTimestamp)

	var rowCount int
	for rows.Next() {
		var id string
		var projectID uint64
		var userID string
		var sessionID sql.NullString
		var eventName string
		var eventNameID uint64
		var eventNameType string
		// properties
		var eventPropertyPageURL sql.NullString
		var eventPropertyPageSpentTime sql.NullString
		var eventPropertyPageScrollPercent sql.NullString
		var eventPropertySource sql.NullString
		var eventPropertyReferrerDomain sql.NullString
		var eventPropertyMedium sql.NullString
		var eventPropertyCampaign sql.NullString
		var eventPropertyCampaignID sql.NullString
		// session properties
		var sessionPropertySpentTime sql.NullString
		var sessionPropertyPageCount sql.NullString
		var sessionPropertyInitialPageURL sql.NullString
		var sessionPropertyInitialReferrerDomain sql.NullString
		var sessionPropertyLatestPageURL sql.NullString

		var customGroupProperties []sql.NullString

		destinations := []interface{}{
			&id, &projectID, &userID, &sessionID, &eventNameID, &eventName, &eventNameType,
			&eventPropertyPageURL, &eventPropertyPageSpentTime, &eventPropertyPageScrollPercent,
			&eventPropertySource, &eventPropertyReferrerDomain, &eventPropertyMedium, &eventPropertyCampaign,
			&eventPropertyCampaignID, &sessionPropertySpentTime, &sessionPropertyPageCount,
			&sessionPropertyInitialPageURL, &sessionPropertyInitialReferrerDomain,
			&sessionPropertyLatestPageURL,
		}
		if customGroupPropertySelectStmnt != "" {
			destinations = append(destinations, pq.Array(&customGroupProperties))
		}

		err = rows.Scan(destinations...)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to scan row to download events on execute_web_analytics_query.")
			return queryResult, http.StatusInternalServerError
		}

		// Todo: Use a property to check instead of is_url(event_name).
		// and remove event_name join.
		isPageEvent := U.IsURLStable(eventName) || eventNameType == TYPE_FILTER_EVENT_NAME

		// Using event_name_id instead of event_name for session check.
		// To remove event_name join later.
		isSessionEvent := eventNameID == sessionEventName.ID
		if !isPageEvent && !isSessionEvent {
			continue
		}

		webEventProperties := &WebAnalyticsEventProperties{
			PageURL:           eventPropertyPageURL.String,
			PageSpentTime:     eventPropertyPageSpentTime.String,
			PageScrollPercent: eventPropertyPageScrollPercent.String,
			Source:            eventPropertySource.String,
			ReferrerDomain:    eventPropertyReferrerDomain.String,
			Medium:            eventPropertyMedium.String,
			Campaign:          eventPropertyCampaign.String,
			CampaignID:        eventPropertyCampaignID.String,

			SessionSpentTime:             sessionPropertySpentTime.String,
			SessionPageCount:             sessionPropertyPageCount.String,
			SessionInitialPageURL:        sessionPropertyInitialPageURL.String,
			SessionInitialReferrerDomain: sessionPropertyInitialReferrerDomain.String,
			SessionLatestPageURL:         sessionPropertyLatestPageURL.String,
		}

		// Add channel property at run time.
		webEventProperties.Channel = getChannel(webEventProperties, isSessionEvent)

		// Creates a map of all custom group properties with values.
		customGroupPropertiesMap := make(map[string]string, 0)
		for i := range customGroupPropertySelectParams {
			groupPropertyKey := customGroupPropertySelectParams[i]

			var propertyValue string
			if groupPropertyKey.(string) == U.EP_CHANNEL {
				propertyValue = webEventProperties.Channel
			} else {
				propertyValue = customGroupProperties[i].String
			}

			customGroupPropertiesMap[groupPropertyKey.(string)] = propertyValue
		}

		webEvent := WebAnalyticsEvent{
			ID:         id,
			ProjectID:  projectID,
			UserID:     userID,
			SessionID:  sessionID.String,
			IsSession:  isSessionEvent,
			Properties: webEventProperties,
		}

		// build all needed aggregates in one scan of events.
		buildWebAnalyticsAggregate(&webEvent, &webAggrState,
			queries.CustomGroupQueries, &customGroupPropertiesMap,
			&customGroupAggrState)

		rowCount++
	}

	queryResult.QueryResult = getWebAnalyticsQueryResultByName(&webAggrState)
	queryResult.CustomGroupQueryResult = getResultForCustomGroupQuery(
		queries.CustomGroupQueries, &customGroupAggrState, &webAggrState)

	logCtx.WithField("no_of_events", rowCount).
		WithField("total_time_taken_in_secs", U.TimeNowUnix()-funcStartTimestamp).
		Info("Executed web analytics query.")

	return queryResult, http.StatusOK
}

func cacheWebsiteAnalyticsForProjectID(projectID uint64, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	dashboardID, webAnalyticsQueries, errCode := getWebAnalyticsQueriesFromDashboardUnits(projectID)
	if errCode != http.StatusFound {
		return
	}
	log.WithFields(log.Fields{"ProjectID": projectID}).
		Info("Starting web analytics dashboard caching")

	var dashboardWaitGroup sync.WaitGroup
	for _, rangeFunction := range U.QueryDateRangePresets {
		from, to := rangeFunction()
		dashboardWaitGroup.Add(1)

		go cacheWebsiteAnalyticsForDateRange(projectID, dashboardID,
			from, to, webAnalyticsQueries, &dashboardWaitGroup)
	}
	dashboardWaitGroup.Wait()
}

func cacheWebsiteAnalyticsForDateRange(projectID, dashboardID uint64, from, to int64,
	queries *WebAnalyticsQueries, waitGroup *sync.WaitGroup) {

	defer waitGroup.Done()
	if isWebAnalyticsDashboardAlreadyCached(projectID, dashboardID, from, to) {
		return
	}

	queriesWithTimeRange := &WebAnalyticsQueries{
		QueryNames:         queries.QueryNames,
		CustomGroupQueries: queries.CustomGroupQueries,
		From:               from,
		To:                 to,
	}

	queryResult, errCode := ExecuteWebAnalyticsQueries(
		projectID, queriesWithTimeRange)
	if errCode != http.StatusOK {
		return
	}

	SetCacheResultForWebAnalyticsDashboard(queryResult, projectID, dashboardID, from, to)
}

func GetCacheResultForWebAnalyticsDashboard(projectID, dashboardID uint64,
	from, to int64) (WebAnalyticsCacheResult, int) {

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

	cacheKey, err := getWebAnalyticsQueryResultCacheKey(projectID, dashboardID, from, to)
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
	projectID, dashboardID uint64, from, to int64) {

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

	cacheKey, err := getWebAnalyticsQueryResultCacheKey(projectID, dashboardID, from, to)
	if err != nil {
		logCtx.WithError(err).Error("Error getting cache key for web analytics dashboard")
	}
	dashboardCacheResult := WebAnalyticsCacheResult{
		Result:      result,
		From:        from,
		To:          to,
		RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
	}

	dashboardCacheResultJSON, err := json.Marshal(&dashboardCacheResult)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode dashboardCacheResult")
		return
	}
	err = cacheRedis.SetPersistent(cacheKey, string(dashboardCacheResultJSON), DashboardCachingDurationInSeconds)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache for channel query")
		return
	}
}

// CacheWebsiteAnalyticsForProjects Runs for all the projectIDs passed as comma separated.
func CacheWebsiteAnalyticsForProjects(stringProjectsIDs string, numRoutines int) {
	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(stringProjectsIDs, "")
	projectIDs := C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)
	if allProjects {
		var errCode int
		projectIDs, errCode = getWebAnalyticsEnabledProjectIDs()
		if errCode != http.StatusFound {
			return
		}
	}

	var waitGroup sync.WaitGroup
	count := 0
	for _, projectID := range projectIDs {
		waitGroup.Add(1)
		count++
		go cacheWebsiteAnalyticsForProjectID(projectID, &waitGroup)

		if count%numRoutines == 0 {
			waitGroup.Wait()
		}
	}
	waitGroup.Wait()
}
