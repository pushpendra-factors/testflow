package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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

// Default Website Analytics named queries and corresponding presentation.
var DefaultWebAnalyticsQueries = map[string]string{
	QueryNameSessions:           PresentationCard,
	QueryNameTotalPageViews:     PresentationCard,
	QueryNameBounceRate:         PresentationCard,
	QueryNameUniqueUsers:        PresentationCard,
	QueryNameAvgSessionDuration: PresentationCard,
	QueryNameAvgPagesPerSession: PresentationCard,
	QueryNameTopPagesReport:     PresentationTable,

	// Todo: Add support for traffic channel report.
	// QueryNameTrafficChannelReport: PresentationTable,
}

type WebAnalyticsEvent struct {
	ID         string
	ProjectID  uint64
	UserID     string // coalsced user_id
	IsSession  bool
	Properties *map[string]interface{}
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

type WebAnalyticsAggregate struct {
	WebAnalyticsGeneralAggregate
	PageAggregates map[string]*WebAnalyticsPageAggregate
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
	rows, err := db.Raw("SELECT distinct(project_id) FROM dashboards WHERE name = ?", DefaultDashboardWebsiteAnalytics).Rows()
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
	if err := db.Where("project_id = ? AND query->>'cl' = ?", projectID, QueryClassWeb).First(&dashboardUnit).Error; err != nil {
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

// Builds aggregate by each event sent.
func buildWebAnalyticsAggregate(webEvent *WebAnalyticsEvent, aggrState *WebAnalyticsAggregate) int {
	logCtx := log.WithField("project_id", webEvent.ProjectID).WithField("event_id", webEvent.ID)

	if aggrState.UniqueUsersMap == nil {
		aggrState.UniqueUsersMap = make(map[string]bool)
	}

	if aggrState.PageAggregates == nil {
		aggrState.PageAggregates = make(map[string]*WebAnalyticsPageAggregate)
	}

	// non-session page event aggregates.
	if !webEvent.IsSession {
		pageURL := U.GetPropertyValueAsString((*webEvent.Properties)[U.EP_PAGE_URL])
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

		pageSpentTime, err := U.GetPropertyValueAsFloat64((*webEvent.Properties)[U.EP_PAGE_SPENT_TIME])
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

		return http.StatusOK
	}

	aggrState.NoOfSessions++

	sessionSpentTime, err := U.GetPropertyValueAsFloat64((*webEvent.Properties)[U.SP_SPENT_TIME])
	if err != nil {
		logCtx.WithError(err).
			Error("Failed converting session_spent_time property value to float64.")
	}
	aggrState.SessionDuration += sessionSpentTime

	sessionPageCount, err := U.GetPropertyValueAsFloat64((*webEvent.Properties)[U.SP_PAGE_COUNT])
	if err != nil {
		logCtx.WithError(err).
			Error("Failed converting session_page_count property value to float64.")
	}
	aggrState.SessionPages += sessionPageCount

	if sessionPageCount == 1 {
		aggrState.NoOfBouncedSessions++
	}

	sessionInitialPageURL := U.GetPropertyValueAsString((*webEvent.Properties)[U.SP_INITIAL_PAGE_URL])
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

	sessionLatestPageURL := U.GetPropertyValueAsString((*webEvent.Properties)[U.SP_LATEST_PAGE_URL])
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
		var avgPageSpentTimeOfPage string
		if aggr.NoOfPageViews > 0 {
			avgPageSpentTimeOfPage = fmt.Sprintf("%0.1f", aggr.TotalSpentTime/float64(aggr.NoOfPageViews))
		}

		row := []interface{}{
			url,
			aggr.NoOfPageViews,
			aggr.NoOfUniqueUsers,
			avgPageSpentTimeOfPage,
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

// Converts single value aggregate to rows and headers format for compatibility.
func fillValueAsWebAnalyticsResult(queryResultByName *map[string]WebAnalyticsQueryResult,
	queryName string, value interface{}) {

	webAResult := WebAnalyticsQueryResult{
		Headers: []string{"count"},
		Rows:    [][]interface{}{[]interface{}{value}},
	}

	(*queryResultByName)[queryName] = webAResult
}

func getResultByNameAsWebAnalyticsResult(webAggrState *WebAnalyticsAggregate) (
	queryResultByName *map[string]WebAnalyticsQueryResult) {

	queryResultByName = &map[string]WebAnalyticsQueryResult{}

	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameSessions, webAggrState.NoOfSessions)
	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameTotalPageViews, webAggrState.NoOfPageViews)
	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameBounceRate, webAggrState.NoOfBouncedSessions)
	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameUniqueUsers, webAggrState.NoOfUniqueUsers)

	var avgSessionDuration, avgPagesPerSession float64
	if webAggrState.NoOfSessions > 0 {
		avgSessionDuration = webAggrState.SessionDuration / float64(webAggrState.NoOfSessions)
		avgPagesPerSession = webAggrState.SessionPages / float64(webAggrState.NoOfSessions)
	}

	// Todo: duration should be in x mins y secs.
	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameAvgSessionDuration, fmt.Sprintf("%0.1f", avgSessionDuration))

	fillValueAsWebAnalyticsResult(queryResultByName,
		QueryNameAvgPagesPerSession, fmt.Sprintf("%0.1f", avgPagesPerSession))

	(*queryResultByName)[QueryNameTopPagesReport] = getTopPagesReportAsWebAnalyticsResult(webAggrState)

	return queryResultByName
}

/*

Query Explanations:

1. Total Page Views: No.of events with $page_raw_url not null.
2. Bounce Rate: (No.of session events with $page_count as 1/ total sessions) * 100
3. Unique Users: Unique no.of coalsced users.
4. Average Session Duration: Sum of $session_spent_time in session events / total sessions - in x mins y secs
5. Average Session: Sum of $page_count on sessoin events / total sessions.

6. Top Pages Report: For each $page_url,
* Page URL: $page_url
* Page Views: No.of events with this as $page_url.
* Unique users: No.of coalsced users with this as $page_url.
* Avg Time on Page: Sum of $page_spent_time on this as $page_url / total pages views of this page.
* Entrances: No.of sessions with this as $initial_page_url
* Exits: No.of sessions with this as $lastest_page_url
* Bounced Entrances: No.of sessions with this as $initial_page_url and $page_count as 1

*/
func ExecuteWebAnalyticsQueries(projectId uint64, queries *WebAnalyticsQueries) (
	queryResultByName *map[string]WebAnalyticsQueryResult, errCode int) {
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

	// Todo: Select required properties directly and avoid JSON decode for each event?
	selectStmnt := "events.id, events.project_id, COALESCE(users.customer_user_id, users.id) as user_id," + " " +
		"events.properties, events.event_name_id, event_names.name as event_name, event_names.type as event_name_type"

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

	for rows.Next() {
		var id string
		var projectID uint64
		var userID string
		var eventProperties *postgres.Jsonb
		var eventName string
		var eventNameID uint64
		var eventNameType string

		err = rows.Scan(&id, &projectID, &userID, &eventProperties,
			&eventNameID, &eventName, &eventNameType)
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

		eventPropertiesMap, err := U.DecodePostgresJsonb(eventProperties)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to decode event_properties JSON on execute_web_analytics_query.")
			continue
		}

		webEvent := WebAnalyticsEvent{
			ID:         id,
			ProjectID:  projectID,
			UserID:     userID,
			Properties: eventPropertiesMap,
			IsSession:  isSessionEvent,
		}
		// build all needed aggregates in one scan of events.
		buildWebAnalyticsAggregate(&webEvent, &webAggrState)
	}

	queryResultByName = getResultByNameAsWebAnalyticsResult(&webAggrState)

	// Todo: build query result by name using aggregates and return.
	return queryResultByName, http.StatusOK
}

func cacheWebsiteAnalyticsForProjectID(projectID uint64, queryNames []string, waitGroup *sync.WaitGroup) {
	logCtx := log.WithFields(log.Fields{
		"Method":    "cacheWebsiteAnalyticsForProjectID",
		"ProjectID": projectID,
	})
	defer waitGroup.Done()

	dashboardID, errCode := getWebAnalyticsDashboardIDForProject(projectID)
	if errCode != http.StatusFound {
		return
	}

	for preset, rangeFunction := range U.QueryDateRangePresets {
		from, to := rangeFunction()
		logCtx = logCtx.WithFields(log.Fields{"Preset": preset, "From": from, "To": to})
		if isWebAnalyticsDashboardAlreadyCached(projectID, dashboardID, from, to) {
			continue
		}

		queryResultsByName, errCode := ExecuteWebAnalyticsQueries(projectID,
			&WebAnalyticsQueries{
				QueryNames: queryNames,
				From:       from,
				To:         to,
			})
		if errCode != http.StatusOK {
			continue
		}
		SetCacheResultForWebAnalyticsDashboard(*queryResultsByName, projectID, dashboardID, from, to)
	}
}

func GetCacheResultForWebAnalyticsDashboard(projectID, dashboardID uint64, from, to int64) (WebAnalyticsCacheResult, int) {
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

func SetCacheResultForWebAnalyticsDashboard(result map[string]WebAnalyticsQueryResult, projectID, dashboardID uint64, from, to int64) {
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
