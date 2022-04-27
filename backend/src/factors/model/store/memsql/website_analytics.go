package memsql

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/model/model"
	U "factors/util"
)

const QueryTypeNamed = "named_query"
const QueryTypeWebAnalyticsCustomGroupQuery = "wa_custom_group_query"

const topPageReportLimit = 50
const customGroupResultLimit = 200

const (
	// Yourstory - Constants for Articles Publised metric.
	WAYourstoryGroupMetricArticlesPublished    = "articles_published"
	CustomPropertyNameYourStoryPublicationDate = "publicationDate"
)

type WebAnalyticsEvent struct {
	ID         string
	ProjectID  uint64
	UserID     string // coalsced user_id
	IsSession  bool
	SessionID  string
	Timestamp  int64
	Properties *WebAnalyticsEventProperties
}

// WebAnalyticsEventProperties - Event Properties for web model.
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

	CustomPropertyYourStoryPublicationDate string
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

func getWebAnalyticsEnabledProjectIDs() ([]uint64, int) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	db := C.GetServices().Db

	var projectIDs []uint64
	rows, err := db.Raw("SELECT distinct(project_id) FROM dashboards WHERE name = ?",
		model.DefaultDashboardWebsiteAnalytics).Rows()
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

func (store *MemSQL) GetWebAnalyticsQueriesFromDashboardUnits(projectID uint64) (uint64, *model.WebAnalyticsQueries, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	dashboardUnits, errCode := store.GetDashboardUnitsForProjectID(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get web analytics queries from dashboard units.")
		return 0, nil, http.StatusInternalServerError
	}

	if len(dashboardUnits) == 0 {
		return 0, nil, http.StatusNotFound
	}

	// Build web analytics queries from dashboard units.
	namedQueries := make([]string, 0, 0)
	webAnalyticsDashboardIDCandidates := make(map[uint64]int64)
	customGroupQueries := make([]model.WebAnalyticsCustomGroupQuery, 0, 0)
	for i := range dashboardUnits {
		dunit := dashboardUnits[i]

		savedQuery, errCode := store.GetQueryWithQueryId(projectID, dunit.QueryId)
		if errCode != http.StatusFound {
			logCtx.Errorf("Failed to fetch query from query_id %d", dunit.QueryId)
			continue
		}

		queryMap, err := U.DecodePostgresJsonb(&savedQuery.Query)
		if err != nil {
			logCtx.WithError(err).WithField("unidID", dunit.ID).
				Error("Failed to decode web analytics dashboard unit query.")
			continue
		}

		// ignoring other queries (query_group or core queries)
		queryClass, exists := (*queryMap)["cl"]
		if !exists || queryClass != model.QueryClassWeb {
			continue
		}

		queryTypeInf, exists := (*queryMap)["type"]
		if !exists {
			logCtx.WithError(err).
				Error("Invalid web analytics query type on dashboard unit")
			continue
		}

		queryType := queryTypeInf.(string)
		if queryType == QueryTypeNamed {
			var query model.NamedQueryUnit
			if err := U.DecodePostgresJsonbToStructType(&savedQuery.Query, &query); err != nil {
				logCtx.WithError(err).
					Error("Failed to decode named query from dashboard unit.")
				continue
			}
			namedQueries = append(namedQueries, query.QueryName)

		} else if queryType == QueryTypeWebAnalyticsCustomGroupQuery {
			var query model.WebAnalyticsCustomGroupQuery
			if err := U.DecodePostgresJsonbToStructType(&savedQuery.Query, &query); err != nil {
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
		webAnalyticsDashboardIDCandidates[dunit.DashboardId]++
	}

	if len(webAnalyticsDashboardIDCandidates) == 0 {
		// No units found with cl='web'. Website analytics is not enabled for project.
		return 0, nil, http.StatusNotFound
	}

	// Get the dashboardID with max web analytics type units.
	var webAnalyticsDashboardID uint64
	var maxCount int64
	for did, count := range webAnalyticsDashboardIDCandidates {
		if count > maxCount {
			maxCount = count
			webAnalyticsDashboardID = did
		}
	}

	// Todo: Return all dashboard ids which has web
	// analytics unit, for caching.
	return webAnalyticsDashboardID,
		&model.WebAnalyticsQueries{
			QueryNames:         namedQueries,
			CustomGroupQueries: customGroupQueries,
		},
		http.StatusFound
}

func getQueryForNamedQueryUnit(class, queryName string) (*postgres.Jsonb, error) {
	logFields := log.Fields{
		"class":      class,
		"query_name": queryName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return U.EncodeStructTypeToPostgresJsonb(model.NamedQueryUnit{Class: class,
		Type: QueryTypeNamed, QueryName: queryName})
}

func (store *MemSQL) addWebAnalyticsDefaultDashboardUnits(projectId uint64,
	agentUUID string, dashboardId uint64) int {
	logFields := log.Fields{
		"project_id":   projectId,
		"agent_uuid":   agentUUID,
		"dashboard_id": dashboardId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	hasFailure := false
	for queryName, presentation := range model.DefaultWebAnalyticsQueries {
		queryJsonb, err := getQueryForNamedQueryUnit(model.QueryClassWeb, queryName)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to named query on add web analytics dashboard units.")
			return http.StatusInternalServerError
		}

		// creating query
		query, errCode, errMsg := store.CreateQuery(projectId,
			&model.Queries{
				Query: *queryJsonb,
				Title: U.GetSnakeCaseToTitleString(queryName),
				Type:  model.QueryTypeDashboardQuery,
			})
		if errCode != http.StatusCreated {
			log.WithFields(log.Fields{"dashboardId": dashboardId,
				"project_id": projectId}).Error(errMsg)
			return http.StatusInternalServerError
		}

		// creating dashboard unit for query created just above
		_, errCode, errMsg = store.CreateDashboardUnitForDashboardClass(projectId, agentUUID,
			&model.DashboardUnit{
				DashboardId:  dashboardId,
				Presentation: presentation,
				QueryId:      query.ID,
			}, model.DashboardClassWebsiteAnalytics)

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

func (store *MemSQL) CreateWebAnalyticsDefaultDashboardWithUnits(projectId uint64, agentUUID string) int {
	logFields := log.Fields{
		"project_id": projectId,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	dashboard, errCode := store.CreateDashboard(projectId, agentUUID, &model.Dashboard{
		Name:  model.DefaultDashboardWebsiteAnalytics,
		Type:  model.DashboardTypeProjectVisible,
		Class: model.DashboardClassWebsiteAnalytics,
	})
	if errCode != http.StatusCreated {
		logCtx.Error("Failed to create web analytics default dashboard.")
		return errCode
	}

	errCode = store.addWebAnalyticsDefaultDashboardUnits(projectId, agentUUID, dashboard.ID)
	if errCode != http.StatusCreated {
		return errCode
	}

	return http.StatusCreated
}

func hasCampaign(campaign, campaignID string) bool {
	logFields := log.Fields{
		"campaign":    campaign,
		"campaign_id": campaignID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return campaign != "" || campaignID != ""
}

func isSearchReferrer(referrerDomain string) bool {
	logFields := log.Fields{
		"referrer_domain": referrerDomain,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	referrerDomain = strings.ToLower(referrerDomain)

	return U.IsContainsAnySubString(referrerDomain, "google", "bing")
}

func isSocialReferrer(referrerDomain string) bool {
	logFields := log.Fields{
		"referrer_domain": referrerDomain,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	referrerDomain = strings.ToLower(referrerDomain)

	return U.IsContainsAnySubString(referrerDomain,
		"facebook", "twitter", "linkedin", "instagram", "tiktok")
}

func getChannel(wep *WebAnalyticsEventProperties, isSessionEvent bool) string {
	logFields := log.Fields{
		"wep":              wep,
		"is_session_event": isSessionEvent,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"query_property_keys": queryPropertyKeys,
		"property_key_values": propertyKeyValues,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var groupKey string
	groupPropertyValues := make([]interface{}, 0, 0)
	for i := range queryPropertyKeys {
		var propertyValue string
		value, exists := (*propertyKeyValues)[queryPropertyKeys[i]]
		if exists && value != "" && value != "NULL" {
			propertyValue = value
		} else {
			propertyValue = model.PropertyValueNone
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

func initWACustomGroupMetricValue(m *map[string]*model.WebAnalyticsCustomGroupMetricValue,
	metricName string) {
	logFields := log.Fields{
		"metric_name": metricName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if metricValue, exists := (*m)[metricName]; !exists || metricValue == nil {
		(*m)[metricName] = &model.WebAnalyticsCustomGroupMetricValue{
			UniqueMap: map[string]bool{},
		}
	}
}

func buildWebAnalyticsAggregateForPageEvent(
	webEvent *WebAnalyticsEvent,
	aggrState *WebAnalyticsAggregate,
	customGroupQueries []model.WebAnalyticsCustomGroupQuery,
	customGroupPropertiesMap *map[string]string,
	customGroupAggrState *map[string]map[string]*model.WebAnalyticsCustomGroupMetric,
	customGroupPrevGroupBySession *map[string]map[string]*model.WebAnalyticsCustomGroupPrevGroup,
	from, to int64,
) int {
	logFields := log.Fields{
		"web_event":            webEvent,
		"custom_group_queries": customGroupQueries,
		"from":                 from,
		"to":                   to,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

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
			(*customGroupAggrState)[query.UniqueID][groupKey] = &model.WebAnalyticsCustomGroupMetric{
				GroupValues: groupPropertyValues,
				MetricValue: map[string]*model.WebAnalyticsCustomGroupMetricValue{},
			}
		}

		// Add page_views metric by default.
		initWACustomGroupMetricValue(
			&(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue,
			model.WAGroupMetricPageViews,
		)
		(*customGroupAggrState)[query.UniqueID][groupKey].
			MetricValue[model.WAGroupMetricPageViews].Value++

		for _, metric := range query.Metrics {
			if metric == model.WAGroupMetricUniqueUsers {
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

			if metric == model.WAGroupMetricUniqueSessions {
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

			if metric == model.WAGroupMetricUniquePages {
				initWACustomGroupMetricValue(
					&(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue,
					metric,
				)

				_, exists := (*customGroupAggrState)[query.UniqueID][groupKey].
					MetricValue[metric].UniqueMap[webEvent.Properties.PageURL]
				if !exists {
					// Count and add unique pages to unique map for deduplication.
					(*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[metric].Value++

					(*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[metric].UniqueMap[webEvent.Properties.PageURL] = true
				}
			}

			if metric == WAYourstoryGroupMetricArticlesPublished {
				initWACustomGroupMetricValue(
					&(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue,
					metric,
				)

				_, exists := (*customGroupAggrState)[query.UniqueID][groupKey].
					MetricValue[metric].UniqueMap[webEvent.Properties.PageURL]
				if !exists {
					publishedDate := webEvent.Properties.
						CustomPropertyYourStoryPublicationDate
					if publishedDate == "" {
						continue
					}

					publishedDateParsed, err := time.Parse(time.RFC3339, publishedDate)
					if err != nil {
						log.WithField("publication_date", publishedDate).
							Warn("Failed to parse yourstory publication date.")
						continue
					}
					publishedTimestamp := publishedDateParsed.UTC().Unix()

					if publishedTimestamp >= from && publishedTimestamp <= to {
						(*customGroupAggrState)[query.UniqueID][groupKey].
							MetricValue[metric].Value++

						(*customGroupAggrState)[query.UniqueID][groupKey].
							MetricValue[metric].UniqueMap[webEvent.Properties.PageURL] = true
					}
				}
			}

			if metric == model.WAGroupMetricTotalTimeSpent || metric == model.WAGroupMetricAvgTimeSpent {
				initWACustomGroupMetricValue(
					&(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue,
					model.WAGroupMetricTotalTimeSpent,
				)

				(*customGroupAggrState)[query.UniqueID][groupKey].
					MetricValue[model.WAGroupMetricTotalTimeSpent].Value += pageSpentTime
			}

			if metric == model.WAGroupMetricTotalScrollDepth || metric == model.WAGroupMetricAvgScrollDepth {
				initWACustomGroupMetricValue(
					&(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue,
					model.WAGroupMetricTotalScrollDepth,
				)

				(*customGroupAggrState)[query.UniqueID][groupKey].
					MetricValue[model.WAGroupMetricTotalScrollDepth].Value += pageScrollPercent
			}

			// As we don't have initial or latest properties state for all custom properties, on session.
			// We have used session_id on page_view, to get initial or latest properties state per group.
			// For example, Total Exits on group (author+brand) = no.of sessions with latest group
			// as current group (author+brand).
			if metric == model.WAGroupMetricTotalExits || metric == model.WAGroupMetricExitPercentage {
				initWACustomGroupMetricValue(
					&(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue,
					model.WAGroupMetricTotalExits,
				)

				if _, exists := (*customGroupPrevGroupBySession)[webEvent.SessionID]; !exists {
					metricPrevGroup := make(map[string]*model.WebAnalyticsCustomGroupPrevGroup, 0)
					(*customGroupPrevGroupBySession)[webEvent.SessionID] = metricPrevGroup
				}

				if _, exists := (*customGroupPrevGroupBySession)[webEvent.SessionID][model.WAGroupMetricTotalExits]; !exists {
					previousGroupBySession := &model.WebAnalyticsCustomGroupPrevGroup{}
					(*customGroupPrevGroupBySession)[webEvent.SessionID][model.WAGroupMetricTotalExits] = previousGroupBySession
				}

				// Proceed only if it is a different group_key for session, and the timestamp is greater than
				// previous for achieving latest of group_key.
				if (*customGroupPrevGroupBySession)[webEvent.SessionID][model.WAGroupMetricTotalExits].GroupKey != groupKey &&
					webEvent.Timestamp > (*customGroupPrevGroupBySession)[webEvent.SessionID][model.WAGroupMetricTotalExits].Timestamp {

					(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue[model.WAGroupMetricTotalExits].Value++

					// As one session should be attributed to one group_key, we remove the attribution
					// on previous group, by holding pointer of value.
					prevGroupKeyValue := (*customGroupPrevGroupBySession)[webEvent.SessionID][model.WAGroupMetricTotalExits].Value
					if prevGroupKeyValue != nil && *prevGroupKeyValue > 0 {
						// TODO(Dinesh): Use exact previous value to subtract.
						// This only support metrics with +1 increament.
						(*prevGroupKeyValue)--
					}

					// Update previous group_key for session, after increamenting, for next group_key.
					(*customGroupPrevGroupBySession)[webEvent.SessionID][model.WAGroupMetricTotalExits].Timestamp = webEvent.Timestamp
					(*customGroupPrevGroupBySession)[webEvent.SessionID][model.WAGroupMetricTotalExits].GroupKey = groupKey
					currentValueAddress := &(*customGroupAggrState)[query.UniqueID][groupKey].MetricValue[model.WAGroupMetricTotalExits].Value
					(*customGroupPrevGroupBySession)[webEvent.SessionID][model.WAGroupMetricTotalExits].Value = currentValueAddress
				}

			}
		}
	}

	return http.StatusOK
}

func buildWebAnalyticsAggregateForSessionEvent(
	webEvent *WebAnalyticsEvent,
	aggrState *WebAnalyticsAggregate,
	customGroupQueries []model.WebAnalyticsCustomGroupQuery,
	customGroupPropertiesMap *map[string]string,
	customGroupAggrState *map[string]map[string]*model.WebAnalyticsCustomGroupMetric,
) int {

	logFields := log.Fields{}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

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
	if sessionInitialPageURL != "" {
		if _, exists := aggrState.PageAggregates[sessionInitialPageURL]; !exists {
			aggrState.PageAggregates[sessionInitialPageURL] = &WebAnalyticsPageAggregate{}
		}

		aggrState.PageAggregates[sessionInitialPageURL].NoOfEntrances++

		if sessionPageCount == 1 {
			aggrState.PageAggregates[sessionInitialPageURL].NoOfBouncedEntrances++
		}
	} else {
		logCtx.Error("Missing $initial_page_url on session on build_web_analytic_aggregate.")
	}

	sessionLatestPageURL := webEvent.Properties.SessionLatestPageURL
	if sessionLatestPageURL != "" {
		if _, exists := aggrState.PageAggregates[sessionLatestPageURL]; !exists {
			aggrState.PageAggregates[sessionLatestPageURL] = &WebAnalyticsPageAggregate{}
		}

		aggrState.PageAggregates[sessionLatestPageURL].NoOfExits++
	} else {
		logCtx.Error("Missing $latest_page_url on session on build_web_analytic_aggregate.")
	}

	return http.StatusOK
}

// buildWebAnalyticsAggregate - Builds aggregate with each event sent as param.
func buildWebAnalyticsAggregate(
	webEvent *WebAnalyticsEvent,
	aggrState *WebAnalyticsAggregate,
	customGroupQueries []model.WebAnalyticsCustomGroupQuery,
	customGroupPropertiesMap *map[string]string,
	customGroupAggrState *map[string]map[string]*model.WebAnalyticsCustomGroupMetric,
	customGroupPrevGroupBySession *map[string]map[string]*model.WebAnalyticsCustomGroupPrevGroup,
	from, to int64,
) int {
	logFields := log.Fields{
		"custom_group_queries": customGroupQueries,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
				make(map[string]*model.WebAnalyticsCustomGroupMetric, 0)
		}
	}

	if !webEvent.IsSession {
		return buildWebAnalyticsAggregateForPageEvent(webEvent, aggrState,
			customGroupQueries, customGroupPropertiesMap, customGroupAggrState,
			customGroupPrevGroupBySession, from, to)
	}

	return buildWebAnalyticsAggregateForSessionEvent(webEvent, aggrState,
		customGroupQueries, customGroupPropertiesMap, customGroupAggrState)
}

func getTopPagesReportAsWebAnalyticsResult(
	webAggr *WebAnalyticsAggregate) model.GenericQueryResult {
	logFields := log.Fields{}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
				aggr.TotalSpentTime/float64(aggr.NoOfPageViews), U.DefaultPrecision)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"total_page_spent_time": aggr.TotalSpentTime,
						"no_of_page_views": aggr.NoOfPageViews}).
					Error("Failed to format average spent time on getTopPagesReportAsWebAnalyticsResult.")

			}

			fmtAvgPageSpentTimeOfPage = model.GetFormattedTime(avgPageSpentTimeOfPage)
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

	return model.GenericQueryResult{Headers: headers, Rows: rows[:rowsLimit]}
}

func getTrafficChannelReport(webAggr *WebAnalyticsAggregate) model.GenericQueryResult {
	logFields := log.Fields{
		"web_aggr": webAggr,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
				aggr.SessionDuration/float64(aggr.NoOfSessions), U.DefaultPrecision)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"total_session_duration": aggr.SessionDuration,
						"no_of_sessions": aggr.NoOfSessions}).
					Error("Failed to format avg session duration on getTrafficChannelReport.")
			}

			bounceRateAsFloat = (float64(aggr.NoOfBouncedSessions) / float64(aggr.NoOfSessions)) * 100
		}
		// Formatted value string.
		avgSessionDuration := model.GetFormattedTime(avgSessionDurationInSecs)
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

	return model.GenericQueryResult{Headers: headers, Rows: rows}
}

// Converts single value aggregate to rows and headers format for compatibility.
func fillValueAsWebAnalyticsResult(queryResultByName *map[string]model.GenericQueryResult,
	queryName string, value interface{}) {
	logFields := log.Fields{
		"query_result_by_name": queryResultByName,
		"query_name":           queryName,
		"value":                value,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	webAResult := model.GenericQueryResult{
		Headers: []string{"count"},
		Rows:    [][]interface{}{[]interface{}{value}},
	}

	(*queryResultByName)[queryName] = webAResult
}

// getFormattedPercentage - Creates percentage string with required precision.
func getFormattedPercentage(value float64) string {
	logFields := log.Fields{
		"value": value,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if value == 0 {
		return "0%"
	}
	if value <= 5 {
		return fmt.Sprintf("%0.2f%%", value)
	}

	return fmt.Sprintf("%0.*f%%", U.DefaultPrecision, value)
}

func getWebAnalyticsQueryResultByName(webAggrState *WebAnalyticsAggregate) (
	queryResultByName *map[string]model.GenericQueryResult) {

	logFields := log.Fields{
		"web_aggr_state":       webAggrState,
		"query_result_by_name": queryResultByName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	queryResultByName = &map[string]model.GenericQueryResult{}

	fillValueAsWebAnalyticsResult(queryResultByName,
		model.QueryNameSessions, webAggrState.NoOfSessions)
	fillValueAsWebAnalyticsResult(queryResultByName,
		model.QueryNameTotalPageViews, webAggrState.NoOfPageViews)

	// Bounce Rate as in percent
	var percentageBouncedSessions float64
	if webAggrState.NoOfSessions > 0 {
		percentageBouncedSessions = float64(webAggrState.NoOfBouncedSessions) /
			float64(webAggrState.NoOfSessions) * 100
	}

	fillValueAsWebAnalyticsResult(queryResultByName,
		model.QueryNameBounceRate, getFormattedPercentage(percentageBouncedSessions))

	fillValueAsWebAnalyticsResult(queryResultByName,
		model.QueryNameUniqueUsers, webAggrState.NoOfUniqueUsers)

	var avgSessionDuration, avgPagesPerSession float64
	if webAggrState.NoOfSessions > 0 {
		avgSessionDuration = webAggrState.SessionDuration / float64(webAggrState.NoOfSessions)
		avgPagesPerSession = webAggrState.SessionPages / float64(webAggrState.NoOfSessions)
	}

	fillValueAsWebAnalyticsResult(queryResultByName,
		model.QueryNameAvgSessionDuration, model.GetFormattedTime(avgSessionDuration))

	precisionedAvgPagesPerSession, _ :=
		U.FloatRoundOffWithPrecision(avgPagesPerSession, U.DefaultPrecision)
	fillValueAsWebAnalyticsResult(queryResultByName,
		model.QueryNameAvgPagesPerSession, precisionedAvgPagesPerSession)

	(*queryResultByName)[model.QueryNameTopPagesReport] =
		getTopPagesReportAsWebAnalyticsResult(webAggrState)
	(*queryResultByName)[model.QueryNameTrafficChannelReport] =
		getTrafficChannelReport(webAggrState)

	return queryResultByName
}

// getResultForCustomGroupQuery - Returns result for each unique_id,
// after ordering by first metric value and limiting.
func getResultForCustomGroupQuery(
	groupQueries []model.WebAnalyticsCustomGroupQuery,
	customGroupAggrState *map[string]map[string]*model.WebAnalyticsCustomGroupMetric,
	webAggrState *WebAnalyticsAggregate,
) (queryResult map[string]*model.GenericQueryResult) {
	logFields := log.Fields{
		"group_queries":           groupQueries,
		"web_aggr_state":          webAggrState,
		"custom_group_aggr_state": customGroupAggrState,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	queryResult = map[string]*model.GenericQueryResult{}

	for i := range groupQueries {
		query := groupQueries[i]

		headers := make([]string, 0, 0)
		headers = append(headers, query.GroupByProperties...)
		// Add readable metric names.
		for i := range query.Metrics {
			headers = append(headers, U.GetSnakeCaseToTitleString(query.Metrics[i]))
		}

		if _, exists := queryResult[query.UniqueID]; !exists {
			queryResult[query.UniqueID] = &model.GenericQueryResult{}
		}
		queryResult[query.UniqueID].Headers = headers

		rows := make([][]interface{}, 0, 0)
		for groupKey, group := range (*customGroupAggrState)[query.UniqueID] {
			row := make([]interface{}, 0, 0)
			row = append(row, group.GroupValues...)

			// Page view is available by default.
			pageViews := (*customGroupAggrState)[query.UniqueID][groupKey].
				MetricValue[model.WAGroupMetricPageViews].Value

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
						MetricValue[metric] = &model.WebAnalyticsCustomGroupMetricValue{}
				}

				var value interface{}
				// Metrics/Aggregates which can be added directly to result.
				// TODO: Use in_list method.
				if metric == model.WAGroupMetricPageViews || metric == model.WAGroupMetricUniqueUsers ||
					metric == model.WAGroupMetricUniqueSessions || metric == model.WAGroupMetricUniquePages ||
					metric == model.WAGroupMetricTotalExits ||
					metric == WAYourstoryGroupMetricArticlesPublished {

					value = (*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[metric].Value
				}

				if metric == model.WAGroupMetricPageViewsContributionPercentage {
					var contributionPercentage float64
					if webAggrState.NoOfPageViews > 0 {
						contributionPercentage = (pageViews / float64(webAggrState.NoOfPageViews)) * 100
					}
					value = getFormattedPercentage(contributionPercentage)
				}

				if metric == model.WAGroupMetricTotalTimeSpent {
					totalTimeSpent := (*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[model.WAGroupMetricTotalTimeSpent].Value
					value = model.GetFormattedTime(totalTimeSpent)
				}

				if metric == model.WAGroupMetricAvgTimeSpent {
					var avgTimeSpent float64
					if pageViews > 0 && (*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[model.WAGroupMetricTotalTimeSpent] != nil {

						avgTimeSpent = (*customGroupAggrState)[query.UniqueID][groupKey].
							MetricValue[model.WAGroupMetricTotalTimeSpent].Value / pageViews
					}

					value = model.GetFormattedTime(avgTimeSpent)
				}

				if metric == model.WAGroupMetricTotalScrollDepth {
					totalScrollDepth := (*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[model.WAGroupMetricTotalScrollDepth].Value
					value = getFormattedPercentage(totalScrollDepth)
				}

				if metric == model.WAGroupMetricAvgScrollDepth {
					var avgScrollDepth float64
					if pageViews > 0 && (*customGroupAggrState)[query.UniqueID][groupKey].
						MetricValue[model.WAGroupMetricTotalScrollDepth] != nil {
						avgScrollDepth = (*customGroupAggrState)[query.UniqueID][groupKey].
							MetricValue[model.WAGroupMetricTotalScrollDepth].Value / pageViews
					}

					value = getFormattedPercentage(avgScrollDepth)
				}

				if metric == model.WAGroupMetricExitPercentage {
					var exitPercentage float64
					if webAggrState.NoOfSessions > 0 {
						exitPercentage = ((*customGroupAggrState)[query.UniqueID][groupKey].
							MetricValue[model.WAGroupMetricTotalExits].Value / float64(webAggrState.NoOfSessions)) * 100
					}
					value = getFormattedPercentage(exitPercentage)

					// TODO(Dinesh): Remove log after fixing zero percentage issue.
					if value == "0%" {
						log.WithField("total_exits", (*customGroupAggrState)[query.UniqueID][groupKey].
							MetricValue[model.WAGroupMetricTotalExits].Value).
							WithField("no_of_sessions", float64(webAggrState.NoOfSessions)).
							Info("Exit percentage is zero.")
					}
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
func (store *MemSQL) ExecuteWebAnalyticsQueries(projectId uint64, queries *model.WebAnalyticsQueries) (
	queryResult *model.WebAnalyticsQueryResult, errCode int) {
	logFields := log.Fields{
		"project_id": projectId,
		"queries":    queries,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	funcStartTimestamp := U.TimeNowUnix()

	webAggrState := WebAnalyticsAggregate{}

	queryResult = &model.WebAnalyticsQueryResult{}
	queryResult.QueryResult = getWebAnalyticsQueryResultByName(&webAggrState)
	queryResult.CustomGroupQueryResult = make(map[string]*model.GenericQueryResult, 0)

	// map[unique_id]map[group_key]*WebAnalyticsCustomGroupMetric
	customGroupAggrState := make(map[string]map[string]*model.WebAnalyticsCustomGroupMetric, 0)
	// map[session_id]map[metric]*WebAnalyticsCustomGroupPrevGroup
	customGroupPrevGroupBySession := make(map[string]map[string]*model.WebAnalyticsCustomGroupPrevGroup)

	if projectId == 0 || queries == nil {
		return queryResult, http.StatusBadRequest
	}
	logCtx := log.WithFields(logFields)

	sessionEventName, errCode := store.GetSessionEventName(projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get session event name on execute_web_analytics_query.")
		return queryResult, http.StatusInternalServerError
	}

	// selectProperties - Columns will be scanned
	// on the same order. Update scanner if order changed.
	selectProperties := []string{
		U.EP_IS_PAGE_VIEW,
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

		CustomPropertyNameYourStoryPublicationDate,
	}

	var selectPropertiesStmnt string
	for _, property := range selectProperties {
		if selectPropertiesStmnt != "" {
			selectPropertiesStmnt = selectPropertiesStmnt + ","
		}

		selectPropertiesStmnt = fmt.Sprintf("%s JSON_EXTRACT_STRING(events.properties, '%s')",
			selectPropertiesStmnt, property)
	}

	var customGroupPropertySelectStmnt string
	var customGroupPropertySelectParams []interface{}
	addedCustomGroupPropertiesMap := make(map[string]bool, 0)

	var customGroupQueryCount int
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
			customGroupQueryCount++

			if customGroupPropertySelectStmnt != "" {
				customGroupPropertySelectStmnt = customGroupPropertySelectStmnt + ","
			}

			// TODO: Use the additional table for properties, if this is slow.
			customGroupPropertySelectStmnt = fmt.Sprintf(
				"%s JSON_EXTRACT_STRING(events.properties, ?)",
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
	selectStmnt := "events.id, events.project_id, events.timestamp," + " " +
		"COALESCE(users.customer_user_id, users.id) as user_id," + " " +
		"events.session_id, events.event_name_id," + " " +
		selectPropertiesStmnt

	if customGroupPropertySelectStmnt != "" {
		selectStmnt = selectStmnt + ", " + customGroupPropertySelectStmnt
		queryParams = append(queryParams, customGroupPropertySelectParams...)
	}

	queryStmnt := "SELECT" + " " + selectStmnt + " " + "FROM events" + " " +
		"LEFT JOIN users ON events.user_id=users.id AND users.project_id = ?" + " " +
		"WHERE events.project_id = ? AND events.timestamp BETWEEN ? AND ?" + " " +
		// Filter session and page view event.
		"AND (JSON_EXTRACT_STRING(events.properties, ?)=? OR events.event_name_id = ?)"
	queryParams = append(queryParams, projectId, projectId, queries.From, queries.To,
		U.EP_IS_PAGE_VIEW, "true", sessionEventName.ID)

	queryStartTimestamp := U.TimeNowUnix()
	rows, tx, err := store.ExecQueryWithContext(queryStmnt, queryParams)
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to execute raw query to download events on execute_web_analytics_query.")
		return queryResult, http.StatusInternalServerError
	}
	defer U.CloseReadQuery(rows, tx)
	logCtx = logCtx.WithField("query_exec_time_in_secs", U.TimeNowUnix()-queryStartTimestamp)

	var rowCount int
	for rows.Next() {
		var id string
		var projectID uint64
		var timestamp int64
		var userID string
		var sessionID sql.NullString
		var eventNameID string
		// properties
		var eventPropertyIsPageView sql.NullBool
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
		// customer specific property.
		var customPropertyYourStoryPublicationDate sql.NullString

		var customGroupProperties []sql.NullString

		destinations := []interface{}{
			&id, &projectID, &timestamp, &userID, &sessionID, &eventNameID, &eventPropertyIsPageView,
			&eventPropertyPageURL, &eventPropertyPageSpentTime, &eventPropertyPageScrollPercent,
			&eventPropertySource, &eventPropertyReferrerDomain, &eventPropertyMedium, &eventPropertyCampaign,
			&eventPropertyCampaignID, &sessionPropertySpentTime, &sessionPropertyPageCount,
			&sessionPropertyInitialPageURL, &sessionPropertyInitialReferrerDomain,
			&sessionPropertyLatestPageURL, &customPropertyYourStoryPublicationDate,
		}
		if customGroupPropertySelectStmnt != "" {
			customGroupProperties = make([]sql.NullString, customGroupQueryCount, customGroupQueryCount)
			for i := 0; i < customGroupQueryCount; i++ {
				destinations = append(destinations, &customGroupProperties[i])
			}
		}

		err = rows.Scan(destinations...)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to scan row to download events on execute_web_analytics_query.")
			return queryResult, http.StatusInternalServerError
		}

		isPageViewEvent := eventPropertyIsPageView.Valid && eventPropertyIsPageView.Bool
		isSessionEvent := eventNameID == sessionEventName.ID
		// Check is part of sql query filter too.
		if !isPageViewEvent && !isSessionEvent {
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

			CustomPropertyYourStoryPublicationDate: customPropertyYourStoryPublicationDate.String,
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
			Timestamp:  timestamp,
			UserID:     userID,
			SessionID:  sessionID.String,
			IsSession:  isSessionEvent,
			Properties: webEventProperties,
		}

		// build all needed aggregates in one scan of events.
		buildWebAnalyticsAggregate(&webEvent, &webAggrState,
			queries.CustomGroupQueries, &customGroupPropertiesMap,
			&customGroupAggrState, &customGroupPrevGroupBySession,
			queries.From, queries.To)

		rowCount++
	}

	logCtx = logCtx.WithField("no_of_events", rowCount).
		WithField("total_time_taken_in_secs", U.TimeNowUnix()-funcStartTimestamp)

	if err := rows.Err(); err != nil {
		logCtx.WithError(err).Error("Failed to scan rows of web analytics query result.")
		return nil, http.StatusInternalServerError
	}

	queryResult.QueryResult = getWebAnalyticsQueryResultByName(&webAggrState)
	queryResult.CustomGroupQueryResult = getResultForCustomGroupQuery(
		queries.CustomGroupQueries, &customGroupAggrState, &webAggrState)

	logCtx.Info("Executed web analytics query.")

	return queryResult, http.StatusOK
}

// GetWebAnalyticsCachePayloadsForProject Returns web analytics cache payloads with date range and queries.
func (store *MemSQL) GetWebAnalyticsCachePayloadsForProject(projectID uint64) ([]model.WebAnalyticsCachePayload, int, string) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	dashboardID, webAnalyticsQueries, errCode := store.GetWebAnalyticsQueriesFromDashboardUnits(projectID)
	if errCode != http.StatusFound {
		errMsg := fmt.Sprintf("Failed to get web analytics queries for project %d", projectID)
		return []model.WebAnalyticsCachePayload{}, errCode, errMsg
	}
	timezoneString, statusCode := store.GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		errMsg := fmt.Sprintf("Failed to get project Timezone for %d", projectID)
		return []model.WebAnalyticsCachePayload{}, statusCode, errMsg
	}

	var cachePayloads []model.WebAnalyticsCachePayload
	for _, rangeFunction := range U.QueryDateRangePresets {
		from, to, errCode := rangeFunction(timezoneString)
		if errCode != nil {
			errMsg := fmt.Sprintf("Failed to get proper project Timezone for %d", projectID)
			return []model.WebAnalyticsCachePayload{}, http.StatusNotFound, errMsg
		}
		cachePayloads = append(cachePayloads, model.WebAnalyticsCachePayload{
			ProjectID:   projectID,
			DashboardID: dashboardID,
			From:        from,
			To:          to,
			Timezone:    timezoneString,
			Queries:     webAnalyticsQueries,
		})
	}
	return cachePayloads, http.StatusFound, ""
}

func (store *MemSQL) cacheWebsiteAnalyticsForProjectID(projectID uint64, waitGroup *sync.WaitGroup, reportCollector *sync.Map) {
	logFields := log.Fields{
		"project_id":       projectID,
		"wait_group":       waitGroup,
		"report_collector": reportCollector,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
	}

	cachePayloads, errCode, errMsg := store.GetWebAnalyticsCachePayloadsForProject(projectID)
	if errCode != http.StatusFound {
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return
	}

	var dashboardWaitGroup sync.WaitGroup
	for index := range cachePayloads {
		if C.GetIsRunningForMemsql() == 0 {
			dashboardWaitGroup.Add(1)
			go store.cacheWebsiteAnalyticsForDateRange(cachePayloads[index], &dashboardWaitGroup, reportCollector)
		} else {
			store.cacheWebsiteAnalyticsForDateRange(cachePayloads[index], &dashboardWaitGroup, reportCollector)
		}
	}
	if C.GetIsRunningForMemsql() == 0 {
		dashboardWaitGroup.Wait()
	}
}

// CacheWebsiteAnalyticsForDateRange Cache web analytics dashboard for with given payload.
func (store *MemSQL) CacheWebsiteAnalyticsForDateRange(cachePayload model.WebAnalyticsCachePayload) (int, model.CachingUnitReport) {
	logFields := log.Fields{
		"cache_payload": cachePayload,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	projectID := cachePayload.ProjectID
	dashboardID := cachePayload.DashboardID
	from, to := cachePayload.From, cachePayload.To
	queries := cachePayload.Queries
	logCtx := log.WithFields(logFields)
	timezoneString := cachePayload.Timezone

	uniqueId := ""
	for _, unit := range queries.CustomGroupQueries {
		uniqueId = uniqueId + unit.UniqueID + "-"
	}
	uniqueId = uniqueId + "--"
	for _, name := range queries.QueryNames {
		uniqueId = uniqueId + name + "-"
	}

	unitReport := model.CachingUnitReport{
		UnitType:    model.CachingUnitWebAnalytics,
		ProjectId:   projectID,
		DashboardID: dashboardID,
		UnitID:      0,
		QueryClass:  uniqueId,
		Status:      model.CachingUnitStatusNotComputed,
		From:        from,
		To:          to,
		QueryRange:  U.SecondsToHMSString(to - from),
	}

	if !model.ShouldRefreshDashboardUnit(projectID, dashboardID, 0, from, to, timezoneString, true) {
		return http.StatusOK, unitReport
	}

	queriesWithTimeRange := &model.WebAnalyticsQueries{
		QueryNames:         queries.QueryNames,
		CustomGroupQueries: queries.CustomGroupQueries,
		From:               from,
		To:                 to,
	}

	startTime := time.Now().Unix()
	queryResult, errCode := store.ExecuteWebAnalyticsQueries(
		projectID, queriesWithTimeRange)
	if errCode != http.StatusOK {
		unitReport.Status = model.CachingUnitStatusFailed
		unitReport.TimeTaken = U.TimeNowUnix() - startTime
		unitReport.TimeTakenStr = U.SecondsToHMSString(unitReport.TimeTaken)
		return http.StatusInternalServerError, unitReport
	}

	timeTaken := time.Now().Unix() - startTime
	timeTakenString := U.SecondsToHMSString(timeTaken)
	logCtx.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).Info("Completed website analytics query.")

	model.SetCacheResultForWebAnalyticsDashboard(queryResult, projectID, dashboardID, from, to, timezoneString)
	unitReport.Status = model.CachingUnitStatusPassed
	unitReport.TimeTaken = timeTaken
	unitReport.TimeTakenStr = timeTakenString
	return http.StatusOK, unitReport
}

func (store *MemSQL) cacheWebsiteAnalyticsForDateRange(cachePayload model.WebAnalyticsCachePayload,
	waitGroup *sync.WaitGroup, reportCollector *sync.Map) {
	logFields := log.Fields{
		"cache_payload":    cachePayload,
		"wait_group":       waitGroup,
		"report_collector": reportCollector,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
	}
	logCtx := log.WithFields(logFields)
	errCode, report := store.CacheWebsiteAnalyticsForDateRange(cachePayload)
	reportCollector.Store(model.GetCachingUnitReportUniqueKey(report), report)
	if errCode != http.StatusOK {
		logCtx.Errorf("Error while running Web analytics query")
		return
	}
	logCtx.Info("Completed caching for WebsiteAnalytics unit")
}

// GetWebAnalyticsEnabledProjectIDsFromList Returns only project ids for which web analytics is enabled.
func (store *MemSQL) GetWebAnalyticsEnabledProjectIDsFromList(stringProjectIDs, excludeProjectIDs string) []uint64 {
	logFields := log.Fields{
		"string_project_ids":  stringProjectIDs,
		"exclude_project_ids": excludeProjectIDs,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	allProjects, projectIDsMap, excludeProjectIDsMap := C.GetProjectsFromListWithAllProjectSupport(stringProjectIDs, excludeProjectIDs)
	allWebAnalyticsProjectIDs, errCode := getWebAnalyticsEnabledProjectIDs()
	if errCode != http.StatusFound {
		return []uint64{}
	}

	var projectIDs []uint64
	if allProjects {
		projectIDs = allWebAnalyticsProjectIDs
	} else {
		projectIDs = C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)
	}

	// Add only those projects for which website analytics is enabled. Exclude marked ones.
	var projectIDsToRun []uint64
	for _, projectID := range projectIDs {
		_, shouldExclude := excludeProjectIDsMap[projectID]
		if U.Uint64ValueIn(projectID, allWebAnalyticsProjectIDs) && !shouldExclude {
			projectIDsToRun = append(projectIDsToRun, projectID)
		}
	}
	return projectIDsToRun
}

// CacheWebsiteAnalyticsForProjects Runs for all the projectIDs passed as comma separated.
func (store *MemSQL) CacheWebsiteAnalyticsForProjects(stringProjectsIDs, excludeProjectIDs string, numRoutines int, reportCollector *sync.Map) {
	logFields := log.Fields{
		"string_projects_ids": stringProjectsIDs,
		"exclude_project_ids": excludeProjectIDs,
		"num_routines":        numRoutines,
		"report_collector":    reportCollector,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	projectIDsToRun := store.GetWebAnalyticsEnabledProjectIDsFromList(stringProjectsIDs, excludeProjectIDs)

	var waitGroup sync.WaitGroup
	count := 0
	if C.GetIsRunningForMemsql() == 0 {
		waitGroup.Add(U.MinInt(len(projectIDsToRun), numRoutines))
	}
	for _, projectID := range projectIDsToRun {
		count++
		log.WithFields(log.Fields{"ProjectID": projectID}).Info("Starting web analytics dashboard caching")
		if C.GetIsRunningForMemsql() == 0 {
			go store.cacheWebsiteAnalyticsForProjectID(projectID, &waitGroup, reportCollector)
			if count%numRoutines == 0 {
				waitGroup.Wait()
				waitGroup.Add(U.MinInt(len(projectIDsToRun)-count, numRoutines))
			}
		} else {
			store.cacheWebsiteAnalyticsForProjectID(projectID, &waitGroup, reportCollector)
		}
	}
	if C.GetIsRunningForMemsql() == 0 {
		waitGroup.Wait()
	}
}

// CacheWebsiteAnalyticsForMonthlyRange Cache monthly range dashboards for website analytics.
func (store *MemSQL) CacheWebsiteAnalyticsForMonthlyRange(projectIDs, excludeProjectIDs string, numMonths, numRoutines int, reportCollector *sync.Map) {
	logFields := log.Fields{
		"project_ids":         projectIDs,
		"exclude_project_ids": excludeProjectIDs,
		"num_routines":        numRoutines,
		"report_collector":    reportCollector,
		"num_months":          numMonths,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	projectIDsToRun := store.GetWebAnalyticsEnabledProjectIDsFromList(projectIDs, excludeProjectIDs)
	for _, projectID := range projectIDsToRun {
		timezoneString, statusCode := store.GetTimezoneForProject(projectID)
		if statusCode != http.StatusFound {
			errMsg := fmt.Sprintf("Failed to get project Timezone for %d", projectID)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}
		monthlyRanges := U.GetMonthlyQueryRangesTuplesZ(numMonths, timezoneString)
		dashboardID, webAnalyticsQueries, errCode := store.GetWebAnalyticsQueriesFromDashboardUnits(projectID)
		if errCode != http.StatusFound {
			errMsg := fmt.Sprintf("Failed to get web analytics queries for project %d", projectID)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}

		var waitGroup sync.WaitGroup
		if C.GetIsRunningForMemsql() == 0 {
			waitGroup.Add(U.MinInt(len(monthlyRanges), numRoutines))
		}
		count := 0
		for _, monthlyRange := range monthlyRanges {
			count++
			from, to := monthlyRange.First, monthlyRange.Second
			cachePayload := model.WebAnalyticsCachePayload{
				ProjectID:   projectID,
				DashboardID: dashboardID,
				From:        from,
				To:          to,
				Queries:     webAnalyticsQueries,
			}
			if C.GetIsRunningForMemsql() == 0 {
				go store.cacheWebsiteAnalyticsForDateRange(cachePayload, &waitGroup, reportCollector)
				if count%numRoutines == 0 {
					waitGroup.Wait()
					waitGroup.Add(U.MinInt(len(monthlyRanges)-count, numRoutines))

				}
			} else {
				store.cacheWebsiteAnalyticsForDateRange(cachePayload, &waitGroup, reportCollector)
			}
		}
		if C.GetIsRunningForMemsql() == 0 {
			waitGroup.Wait()
		}
	}
}
