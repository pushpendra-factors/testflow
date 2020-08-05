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
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
)

type WebAnalyticsQueries struct {
	// Multiple queries  with same timerange.
	QueryNames []string `json:"query_names"`
	From       int64    `json:"from"`
	To         int64    `json:"to"`
}

// Todo(Dinesh): Define a standard query result struct.
// Use it everywhere, like channel_analytics.
type WebAnalyticsQueryResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}

// NamedQueryUnit - Query structure for dashboard unit.
type NamedQueryUnit struct {
	Class     string `json:"cl"`
	Type      string `json:"type"`
	QueryName string `json:"qname"`
}

const QueryTypeNamed = "named_query"
const DefaultDashboardWebsiteAnalytics = "Website Analytics"
const topPageReportLimit = 50
const defaultPrecision = 1

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
	Properties *WebAnalyticsEventProperties
}

// WebAnalyticsEventProperties - Event Properties for web analytics.
type WebAnalyticsEventProperties struct {
	PageURL        string
	PageSpentTime  string
	Source         string
	Medium         string
	ReferrerDomain string
	Campaign       string
	CampaignID     string

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
	NoOfBouncedSessions int     // no.of sessions with $page_count = 1.
	SessionDuration     float64 // sum of $session_spent_time of session event with

	NoOfUniqueUsers int
	UniqueUsersMap  map[string]bool

	NoOfSessions      int // no.of sessions
	UniqueSessionsMap map[string]bool
}

type WebAnalyticsAggregate struct {
	WebAnalyticsGeneralAggregate
	PageAggregates    map[string]*WebAnalyticsPageAggregate
	ChannelAggregates map[string]*WebAnalyticsChannelAggregate
}

type WebAnalyticsCacheResult struct {
	Result      map[string]WebAnalyticsQueryResult `json:"result"`
	From        int64                              `json:"from"`
	To          int64                              `json:"tom"`
	RefreshedAt int64                              `json:"refreshed_at"`
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

func getWebAnalyticsDashboardIDForProject(projectID uint64) (uint64, int) {
	logCtx := log.WithFields(log.Fields{
		"Method":    "GetWebAnalyticsDashboardIDForProject",
		"ProjectID": projectID,
	})
	db := C.GetServices().Db

	var dashboardUnit DashboardUnit
	if err := db.Where("project_id = ? AND query->>'cl' = ?", projectID, QueryClassWeb).
		First(&dashboardUnit).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return 0, http.StatusNotFound
		}
		logCtx.WithError(err).Error("Failed to get dashboard_id for web analytics dashboard")
		return 0, http.StatusInternalServerError
	}
	return dashboardUnit.DashboardId, http.StatusFound
}

func getWebAnalyticsQueryResultCacheKey(projectID, dashboardID uint64, from, to int64) (*cacheRedis.Key, error) {
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

func addWebAnalyticsDefaultDashboardUnits(projectId uint64, agentUUID string, dashboardId uint64) int {
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

func isSearchReferrer(referrer string) bool {
	return strings.Contains(referrer, "google") || strings.Contains(referrer, "bing")
}

func isSocialReferrer(referrer string) bool {
	return strings.Contains(referrer, "facebook") ||
		strings.Contains(referrer, "linkedin") ||
		strings.Contains(referrer, "twitter") ||
		strings.Contains(referrer, "instagram") ||
		strings.Contains(referrer, "tiktok")
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

// Builds aggregate by each event sent.
func buildWebAnalyticsAggregate(webEvent *WebAnalyticsEvent, aggrState *WebAnalyticsAggregate) int {
	logCtx := log.WithField("project_id", webEvent.ProjectID).WithField("event_id", webEvent.ID)

	if aggrState.UniqueUsersMap == nil {
		aggrState.UniqueUsersMap = make(map[string]bool)
	}

	if aggrState.PageAggregates == nil {
		aggrState.PageAggregates = make(map[string]*WebAnalyticsPageAggregate)
	}

	if aggrState.ChannelAggregates == nil {
		aggrState.ChannelAggregates = make(map[string]*WebAnalyticsChannelAggregate)
	}

	channel := getChannel(webEvent.Properties, webEvent.IsSession)
	if _, exists := aggrState.ChannelAggregates[channel]; !exists {
		aggrState.ChannelAggregates[channel] = &WebAnalyticsChannelAggregate{}
	}

	// non-session page event aggregates.
	if !webEvent.IsSession {
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

		if aggrState.PageAggregates[pageURL].UniqueUsersMap == nil {
			aggrState.PageAggregates[pageURL].UniqueUsersMap = make(map[string]bool)
		}

		if _, exists := aggrState.PageAggregates[pageURL].UniqueUsersMap[webEvent.UserID]; !exists {
			aggrState.PageAggregates[pageURL].UniqueUsersMap[webEvent.UserID] = true
			aggrState.PageAggregates[pageURL].NoOfUniqueUsers++
		}

		aggrState.ChannelAggregates[channel].NoOfPageViews++

		if aggrState.ChannelAggregates[channel].UniqueUsersMap == nil {
			aggrState.ChannelAggregates[channel].UniqueUsersMap = make(map[string]bool)
		}

		if _, exists := aggrState.ChannelAggregates[channel].UniqueUsersMap[webEvent.UserID]; !exists {
			aggrState.ChannelAggregates[channel].UniqueUsersMap[webEvent.UserID] = true
			aggrState.ChannelAggregates[channel].NoOfUniqueUsers++
		}

		return http.StatusOK
	}

	aggrState.NoOfSessions++
	aggrState.ChannelAggregates[channel].NoOfBouncedSessions++

	sessionSpentTime, err := U.GetPropertyValueAsFloat64(webEvent.Properties.SessionSpentTime)
	if err != nil {
		logCtx.WithError(err).
			Error("Failed converting session_spent_time property value to float64.")
	}
	aggrState.SessionDuration += sessionSpentTime
	aggrState.ChannelAggregates[channel].SessionDuration += sessionSpentTime

	sessionPageCount, err := U.GetPropertyValueAsFloat64(webEvent.Properties.SessionPageCount)
	if err != nil {
		logCtx.WithError(err).
			Error("Failed converting session_page_count property value to float64.")
	}
	aggrState.SessionPages += sessionPageCount

	if sessionPageCount == 1 {
		aggrState.NoOfBouncedSessions++
		aggrState.ChannelAggregates[channel].NoOfBouncedSessions++
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

func getTopPagesReportAsWebAnalyticsResult(
	webAggr *WebAnalyticsAggregate) WebAnalyticsQueryResult {

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
			avgPageSpentTimeOfPage, _ := U.FloatRoundOffWithPrecision(
				aggr.TotalSpentTime/float64(aggr.NoOfPageViews), defaultPrecision)
			fmtAvgPageSpentTimeOfPage = getFormattedTime(int64(avgPageSpentTimeOfPage))
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

	return WebAnalyticsQueryResult{Headers: headers, Rows: rows[:rowsLimit]}
}

func getTrafficChannelReport(webAggr *WebAnalyticsAggregate) WebAnalyticsQueryResult {
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
		var avgSessionDuration, bounceRate string
		if aggr.NoOfSessions > 0 {
			avgSessionDurationInSecs, _ := U.FloatRoundOffWithPrecision(
				aggr.SessionDuration/float64(aggr.NoOfSessions), defaultPrecision)
			avgSessionDuration = getFormattedTime(int64(avgSessionDurationInSecs))

			bounceRateAsInt := (aggr.NoOfBouncedSessions / aggr.NoOfSessions) * 100
			bounceRate = fmt.Sprintf("%d%%", bounceRateAsInt)
		} else {
			avgSessionDuration = "0s"
			bounceRate = "0%"
		}

		row := []interface{}{
			channel,
			aggr.NoOfPageViews,
			aggr.NoOfUniqueUsers,
			aggr.NoOfSessions,
			bounceRate,
			avgSessionDuration,
		}
		rows = append(rows, row)
	}

	// sort by NoOfPageViews.
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][1].(int) > rows[j][1].(int)
	})

	return WebAnalyticsQueryResult{Headers: headers, Rows: rows}
}

// Converts single value aggregate to rows and headers format for compatibility.
func fillValueAsWebAnalyticsResult(queryResultByName *map[string]WebAnalyticsQueryResult,
	queryName string, value interface{}) {

	webAResult := WebAnalyticsQueryResult{
		Headers: []string{"count"},
		Rows:    [][]interface{}{[]interface{}{value}},
	}

	(*queryResultByName)[queryName] = webAResult
}

// Converts seconds into hh mm ss format
func getFormattedTime(totalSeconds int64) string {
	fmtTime := ""
	if totalSeconds > 3600 {
		fmtTime = fmt.Sprintf("%dh ", totalSeconds/3600)
	}
	if totalSeconds > 60 {
		fmtTime = fmtTime + fmt.Sprintf("%dm ", (totalSeconds%3600)/60)
	}
	if totalSeconds > 0 {
		fmtTime = fmtTime + fmt.Sprintf("%ds", totalSeconds%60)
	} else {
		return "0s"
	}
	return fmtTime
}

func getResultByNameAsWebAnalyticsResult(webAggrState *WebAnalyticsAggregate) (
	queryResultByName *map[string]WebAnalyticsQueryResult) {

	queryResultByName = &map[string]WebAnalyticsQueryResult{}

	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameSessions, webAggrState.NoOfSessions)
	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameTotalPageViews, webAggrState.NoOfPageViews)

	// Bounce Rate as in percent
	var percentageBouncedSessions float64
	var precisionedBounceRate string
	if webAggrState.NoOfSessions > 0 {
		percentageBouncedSessions = float64(webAggrState.NoOfBouncedSessions) /
			float64(webAggrState.NoOfSessions) * 100

		if percentageBouncedSessions == 0 {
			precisionedBounceRate = "0%"
		} else if percentageBouncedSessions <= 5 {
			precisionedBounceRate = fmt.Sprintf("%0.2f%%", percentageBouncedSessions)
		} else {
			precisionedBounceRate = fmt.Sprintf("%0.*f%%",
				defaultPrecision, percentageBouncedSessions)
		}
	} else {
		precisionedBounceRate = "0%"
	}

	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameBounceRate, precisionedBounceRate)

	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameUniqueUsers, webAggrState.NoOfUniqueUsers)

	var avgSessionDuration, avgPagesPerSession float64
	if webAggrState.NoOfSessions > 0 {
		avgSessionDuration = webAggrState.SessionDuration / float64(webAggrState.NoOfSessions)
		avgPagesPerSession = webAggrState.SessionPages / float64(webAggrState.NoOfSessions)
	}

	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameAvgSessionDuration, getFormattedTime(int64(avgSessionDuration)))

	precisionedAvgPagesPerSession, _ := U.FloatRoundOffWithPrecision(avgPagesPerSession, defaultPrecision)
	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameAvgPagesPerSession, precisionedAvgPagesPerSession)

	(*queryResultByName)[QueryNameTopPagesReport] = getTopPagesReportAsWebAnalyticsResult(webAggrState)
	(*queryResultByName)[QueryNameTrafficChannelReport] = getTrafficChannelReport(webAggrState)

	return queryResultByName
}

// ExecuteWebAnalyticsQueries - executes the web analytics query and returns result by query_name.
func ExecuteWebAnalyticsQueries(projectId uint64, queries *WebAnalyticsQueries) (
	queryResultByName *map[string]WebAnalyticsQueryResult, errCode int) {

	funcStartTimestamp := U.TimeNowUnix()

	webAggrState := WebAnalyticsAggregate{}
	queryResultByName = getResultByNameAsWebAnalyticsResult(&webAggrState)

	if projectId == 0 || queries == nil {
		return queryResultByName, http.StatusBadRequest
	}

	logCtx := log.WithField("project_id", projectId).WithField("query", queries)

	sessionEventName, errCode := GetSessionEventName(projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get session event name on execute_web_analytics_query.")
		return queryResultByName, http.StatusInternalServerError
	}

	queryStartTimestamp := U.TimeNowUnix()

	// selectProperties - Columns will be scanned
	// on the same order. Update scanner if order changed.
	selectProperties := []string{
		U.EP_PAGE_URL,
		U.EP_PAGE_SPENT_TIME,
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

		selectPropertiesStmnt = fmt.Sprintf("%s events.Properties->>'%s'",
			selectPropertiesStmnt, property)
	}

	selectStmnt := "events.id, events.project_id, COALESCE(users.customer_user_id, users.id) as user_id," + " " +
		"events.event_name_id, event_names.name as event_name, event_names.type as event_name_type," + " " +
		selectPropertiesStmnt

	queryStmnt := "SELECT" + " " + selectStmnt + " " +
		"FROM events LEFT JOIN event_names ON events.event_name_id=event_names.id" + " " +
		"LEFT JOIN users ON events.user_id=users.id" + " " +
		"WHERE events.project_id = ? AND events.timestamp BETWEEN ? AND ?" + " " +
		"AND (events.properties->>'$page_raw_url' IS NOT NULL OR events.event_name_id = ?)"

	db := C.GetServices().Db
	rows, err := db.Raw(queryStmnt, projectId, queries.From, queries.To, sessionEventName.ID).Rows()
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to execute raw query to download events on execute_web_analytics_query.")
		return queryResultByName, http.StatusInternalServerError
	}
	defer rows.Close()

	logCtx = logCtx.WithField("query_exec_time_in_secs", U.TimeNowUnix()-queryStartTimestamp)

	var rowCount int
	for rows.Next() {
		var id string
		var projectID uint64
		var userID string
		var eventName string
		var eventNameID uint64
		var eventNameType string
		// properties
		var eventPropertyPageURL sql.NullString
		var eventPropertyPageSpentTime sql.NullString
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

		err = rows.Scan(&id, &projectID, &userID, &eventNameID, &eventName, &eventNameType,
			&eventPropertyPageURL, &eventPropertyPageSpentTime, &eventPropertySource, &eventPropertyReferrerDomain,
			&eventPropertyMedium, &eventPropertyCampaign, &eventPropertyCampaignID,
			&sessionPropertySpentTime, &sessionPropertyPageCount, &sessionPropertyInitialPageURL,
			&sessionPropertyInitialReferrerDomain, &sessionPropertyLatestPageURL)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to scan row to download events on execute_web_analytics_query.")
			return queryResultByName, http.StatusInternalServerError
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
			PageURL:        eventPropertyPageURL.String,
			PageSpentTime:  eventPropertyPageSpentTime.String,
			Source:         eventPropertySource.String,
			ReferrerDomain: eventPropertyReferrerDomain.String,
			Medium:         eventPropertyMedium.String,
			Campaign:       eventPropertyCampaign.String,
			CampaignID:     eventPropertyCampaignID.String,

			SessionSpentTime:             sessionPropertySpentTime.String,
			SessionPageCount:             sessionPropertyPageCount.String,
			SessionInitialPageURL:        sessionPropertyInitialPageURL.String,
			SessionInitialReferrerDomain: sessionPropertyInitialReferrerDomain.String,
			SessionLatestPageURL:         sessionPropertyLatestPageURL.String,
		}

		webEvent := WebAnalyticsEvent{
			ID:         id,
			ProjectID:  projectID,
			UserID:     userID,
			Properties: webEventProperties,
			IsSession:  isSessionEvent,
		}

		// build all needed aggregates in one scan of events.
		buildWebAnalyticsAggregate(&webEvent, &webAggrState)
		rowCount++
	}

	queryResultByName = getResultByNameAsWebAnalyticsResult(&webAggrState)

	logCtx.WithField("no_of_events", rowCount).
		WithField("total_time_taken_in_secs", U.TimeNowUnix()-funcStartTimestamp).
		Info("Executed web analytics query.")

	// Todo: build query result by name using aggregates and return.
	return queryResultByName, http.StatusOK
}

func cacheWebsiteAnalyticsForProjectID(projectID uint64, queryNames []string, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	dashboardID, errCode := getWebAnalyticsDashboardIDForProject(projectID)
	if errCode != http.StatusFound {
		return
	}
	log.WithFields(log.Fields{"ProjectID": projectID}).Info("Starting web analytics dashboard caching")

	var dashboardWaitGroup sync.WaitGroup
	for _, rangeFunction := range U.QueryDateRangePresets {
		from, to := rangeFunction()
		dashboardWaitGroup.Add(1)
		go cacheWebsiteAnalyticsForDateRange(projectID, dashboardID, from, to, queryNames, &dashboardWaitGroup)
	}
	dashboardWaitGroup.Wait()
}

func cacheWebsiteAnalyticsForDateRange(projectID, dashboardID uint64, from, to int64,
	queryNames []string, waitGroup *sync.WaitGroup) {

	defer waitGroup.Done()
	if isWebAnalyticsDashboardAlreadyCached(projectID, dashboardID, from, to) {
		return
	}

	queryResultsByName, errCode := ExecuteWebAnalyticsQueries(projectID,
		&WebAnalyticsQueries{
			QueryNames: queryNames,
			From:       from,
			To:         to,
		})
	if errCode != http.StatusOK {
		return
	}
	SetCacheResultForWebAnalyticsDashboard(*queryResultsByName, projectID, dashboardID, from, to)
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

func SetCacheResultForWebAnalyticsDashboard(result map[string]WebAnalyticsQueryResult,
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

	var queryNames []string
	for queyrName := range DefaultWebAnalyticsQueries {
		queryNames = append(queryNames, queyrName)
	}

	var waitGroup sync.WaitGroup
	count := 0
	for _, projectID := range projectIDs {
		waitGroup.Add(1)
		count++
		go cacheWebsiteAnalyticsForProjectID(projectID, queryNames, &waitGroup)

		if count%numRoutines == 0 {
			waitGroup.Wait()
		}
	}
	waitGroup.Wait()
}
