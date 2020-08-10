package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

type AttributionQueryUnit struct {
	Class string                  `json:"cl"`
	Query *AttributionQuery       `json:"query"`
	Meta  *map[string]interface{} `json:"meta"`
}

type AttributionQuery struct {
	CampaignMetrics        []string                   `json:"cm"`
	ConversionEvent        QueryEventWithProperties   `json:"ce"`
	LinkedEvents           []QueryEventWithProperties `json:"lfe"`
	AttributionKey         string                     `json:"attribution_key"`
	AttributionMethodology string                     `json:"attribution_methodology"`
	LookbackDays           int                        `json:"lbw"`
	From                   int64                      `json:"from"`
	To                     int64                      `json:"to"`
}
type RangeTimestamp struct {
	MinTimestamp int64
	MaxTimestamp int64
}

func (q *AttributionQueryUnit) GetClass() string {
	return q.Class
}

func (q *AttributionQueryUnit) GetQueryDateRange() (from, to int64) {
	return q.Query.From, q.Query.To
}

func (q *AttributionQueryUnit) SetQueryDateRange(from, to int64) {
	q.Query.From, q.Query.To = from, to
}

const (
	ATTRIBUTION_METHOD_FIRST_TOUCH            = "First_Touch"
	ATTRIBUTION_METHOD_FIRST_TOUCH_NON_DIRECT = "First_Touch_ND"
	ATTRIBUTION_METHOD_LAST_TOUCH             = "Last_Touch"
	ATTRIBUTION_METHOD_LAST_TOUCH_NON_DIRECT  = "Last_Touch_ND"

	ATTRIBUTION_KEY_CAMPAIGN = "Campaign"
	ATTRIBUTION_KEY_SOURCE   = "Source"

	SECS_IN_A_DAY        = int64(86400)
	LOOKBACK_CAP_IN_DAYS = 180
	USER_BATCH_SIZE      = 3000
)

var ATTRIBUTION_FIXED_HEADERS = []string{"Impressions", "Clicks", "Spend", "Website Visitors"}

type AttributionData struct {
	Name                 string
	Impressions          int
	Clicks               int
	Spend                float64
	WebsiteVisitors      int64
	ConversionEventCount int64
	LinkedEventsCount    []int64
}

type UserInfo struct {
	coalUserId   string
	propertiesId string
}

type UserEventInfo struct {
	coalUserId string
	eventName  string
}

// Maps the {attribution key} to the session properties field
func getQuerySessionProperty(attributionKey string) (string, error) {
	if attributionKey == ATTRIBUTION_KEY_CAMPAIGN {
		return U.EP_CAMPAIGN, nil
	} else if attributionKey == ATTRIBUTION_KEY_SOURCE {
		return U.EP_SOURCE, nil
	} // Todo (Anil) add here for adgroup, medium etc
	return "", errors.New("invalid query properties")
}

// Maps the {attribution key and attribution method} to the user properties field.
func GetQueryUserProperty(query *AttributionQuery) (string, error) {
	if query.AttributionKey == ATTRIBUTION_KEY_CAMPAIGN {
		if query.AttributionMethodology == ATTRIBUTION_METHOD_FIRST_TOUCH {
			return U.UP_INITIAL_CAMPAIGN, nil
		} else if query.AttributionMethodology == ATTRIBUTION_METHOD_LAST_TOUCH {
			return U.UP_LATEST_CAMPAIGN, nil
		}
	} else if query.AttributionKey == ATTRIBUTION_KEY_SOURCE {
		if query.AttributionMethodology == ATTRIBUTION_METHOD_FIRST_TOUCH {
			return U.UP_INITIAL_SOURCE, nil
		} else if query.AttributionMethodology == ATTRIBUTION_METHOD_LAST_TOUCH {
			return U.UP_LATEST_SOURCE, nil
		}
	}
	return "", errors.New("invalid query properties")
}

// Attribute each user to the conversion event and linked event by attribution Id.
func addUpLinkedFunnelEventCount(linkedEvents []QueryEventWithProperties,
	attributionData map[string]*AttributionData, linkedUserAttributionData map[string]map[string]string) {

	linkedEventToPositionMap := make(map[string]int)
	for position, linkedEvent := range linkedEvents {
		linkedEventToPositionMap[linkedEvent.Name] = position
	}
	// fill up all the linked events count with 0 value
	for _, attributionRow := range attributionData {
		if attributionRow != nil {
			for len(attributionRow.LinkedEventsCount) < len(linkedEvents) {
				attributionRow.LinkedEventsCount = append(attributionRow.LinkedEventsCount, 0)
			}
		}
	}
	// update linked up events with event hit count
	for linkedEventName, userIdAttributionIdMap := range linkedUserAttributionData {
		for _, attributionId := range userIdAttributionIdMap {
			attributionRow := attributionData[attributionId]
			if attributionRow != nil {
				attributionRow.LinkedEventsCount[linkedEventToPositionMap[linkedEventName]] += 1
			}
		}
	}
}

// Adds common column names and linked events as header to the result rows.
func addHeadersByAttributionKey(result *QueryResult, query *AttributionQuery) {
	attributionKey := query.AttributionKey
	result.Headers = append(append(result.Headers, attributionKey), ATTRIBUTION_FIXED_HEADERS...)
	conversionEventUsers := fmt.Sprintf("%s - Users", query.ConversionEvent.Name)
	costPerConversion := fmt.Sprintf("Cost Per Conversion")
	result.Headers = append(result.Headers, conversionEventUsers, costPerConversion)
	if len(query.LinkedEvents) > 0 {
		for _, event := range query.LinkedEvents {
			result.Headers = append(result.Headers, fmt.Sprintf("%s - Users", event.Name))
		}
	}
}

/* Executes the Attribution using following steps:
	1. Get all the sessions data (userId, attributionId, timestamp) for given period by attribution key
 	2. Add the website visitor info using session data from step 1
	3. i) 	Find out users who hit conversion event applying filter
	  ii)	Using users from 3.i) find out users who hit linked funnel event applying filter
	4. Apply attribution methodology
	5. Add performance data by attributionId
*/
func ExecuteAttributionQuery(projectId uint64, query *AttributionQuery) (*QueryResult, error) {

	result := &QueryResult{}
	attributionData := make(map[string]*AttributionData)
	projectSetting, errCode := GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return nil, errors.New("failed to get project Settings")
	}
	if projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "" {
		return nil, errors.New("execute attribution query failed as no ad-words customer account id found")
	}
	addHeadersByAttributionKey(result, query)
	sessionEventNameId, eventNameToIdList, err := getEventInformation(projectId, query)
	if err != nil {
		return nil, err
	}

	// 1. Get all the sessions (userId, attributionId, timestamp) for given period by attribution key
	_sessions, sessionUsers, err := getAllTheSessions(projectId, sessionEventNameId, query)
	usersInfo, err := getCoalesceUsersFromList(sessionUsers, projectId)
	if err != nil {
		return nil, err
	}
	sessions := updateSessionsByCoalesceId(_sessions, usersInfo)

	// 2. Add website visitor information against the attribution key
	addWebsiteVisitorsInfo(query.From, query.To, attributionData, sessions)

	// 3. Fetch users who hit conversion event
	var usersToBeAttributed []UserEventInfo
	usersHitConversion, err := getConvertedUsers(projectId, query, eventNameToIdList, usersInfo,
		&usersToBeAttributed)
	if err != nil {
		return nil, err
	}
	err = getLinkedFunnelEventUsers(projectId, query, eventNameToIdList, usersInfo,
		&usersToBeAttributed, usersHitConversion)
	if err != nil {
		return nil, err
	}

	// 4. Apply attribution based on given attribution methodology
	userConversionHit, userLinkedFEHit, err := applyAttribution(query.AttributionMethodology,
		query.ConversionEvent.Name,
		usersToBeAttributed, sessions)
	if err != nil {
		return nil, err
	}

	addUpConversionEventCount(attributionData, userConversionHit)
	addUpLinkedFunnelEventCount(query.LinkedEvents, attributionData, userLinkedFEHit)

	// 5. Add the performance information
	currency, err := AddPerformanceReportInfo(projectId, attributionData, query.From, query.To,
		*projectSetting.IntAdwordsCustomerAccountId)
	if err != nil {
		return nil, err
	}

	result.Rows = getRowsByMaps(attributionData, query)
	result.Meta.Currency = currency
	return result, nil
}

// Converts a slice of int64 to a slice of interface.
func getInterfaceList(data []int64) []interface{} {
	var list []interface{}
	for _, val := range data {
		list = append(list, []interface{}{val})
	}
	return list
}

// Returns result in from of metrics. For empty attribution id, the values are accumulated into "$none".
func getRowsByMaps(attributionData map[string]*AttributionData, query *AttributionQuery) [][]interface{} {
	rows := make([][]interface{}, 0)
	nonMatchingRow := []interface{}{"none", 0, 0, float64(0), int64(0), int64(0), float64(0)}
	for i := 0; i < len(query.LinkedEvents); i++ {
		nonMatchingRow = append(nonMatchingRow, int64(0))
	}
	for key, data := range attributionData {
		attributionIdName := data.Name
		if attributionIdName == "" {
			attributionIdName = key
		}
		if attributionIdName != "" {
			var row []interface{}
			cpc := 0.0
			if data.ConversionEventCount != 0 {
				cpc = data.Spend / float64(data.ConversionEventCount)
			}
			row = append(row, attributionIdName, data.Impressions, data.Clicks, data.Spend, data.WebsiteVisitors,
				data.ConversionEventCount, cpc)
			row = append(row, getInterfaceList(data.LinkedEventsCount)...)
			rows = append(rows, row)
		} else {
			updateNonMatchingRow(&nonMatchingRow, data, query)
		}
	}
	rows = append(rows, nonMatchingRow)
	// sort the rows by conversionEvent
	sort.Slice(rows, func(i, j int) bool {
		return rows[i][5].(int64) > rows[j][5].(int64)
	})
	return rows
}

func updateNonMatchingRow(nonMatchingRow *[]interface{}, data *AttributionData, query *AttributionQuery) {
	(*nonMatchingRow)[1] = (*nonMatchingRow)[1].(int) + data.Impressions
	(*nonMatchingRow)[2] = (*nonMatchingRow)[2].(int) + data.Clicks
	(*nonMatchingRow)[3] = (*nonMatchingRow)[3].(float64) + data.Spend
	(*nonMatchingRow)[4] = (*nonMatchingRow)[4].(int64) + data.WebsiteVisitors
	(*nonMatchingRow)[5] = (*nonMatchingRow)[5].(int64) + data.ConversionEventCount
	for index := 0; index < len(query.LinkedEvents); index++ {
		if index < len(data.LinkedEventsCount) {
			(*nonMatchingRow)[5+index+1] = (*nonMatchingRow)[5+index+1].(int64) + data.LinkedEventsCount[index]
		}
	}
}

// Groups all unique users by attributionId and adds it to attributionData
func addUpConversionEventCount(attributionData map[string]*AttributionData, usersIdAttributionIdMap map[string]string) {
	for _, attributionId := range usersIdAttributionIdMap {
		if _, exists := attributionData[attributionId]; !exists {
			attributionData[attributionId] = &AttributionData{}
		}
		attributionData[attributionId].ConversionEventCount += 1
	}
}

// Returns the map of coalesce userId for given list of users
func getCoalesceUsersFromList(userIdsWithSession []string, projectId uint64) (map[string]UserInfo, error) {
	userIdsInBatches := U.GetStringListAsBatch(userIdsWithSession, USER_BATCH_SIZE)
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})
	userIdCoalUserIdMap := make(map[string]UserInfo)
	for _, users := range userIdsInBatches {
		placeHolder := U.GetValuePlaceHolder(len(users))
		value := U.GetInterfaceList(users)
		queryUserIdCoalId := "SELECT id, COALESCE(users.customer_user_id,users.id) AS coal_user_id, " +
			" properties_id FROM users WHERE id = ANY (VALUES " + placeHolder + " )"
		rows, err := db.Raw(queryUserIdCoalId, value...).Rows()
		defer rows.Close()
		if err != nil {
			logCtx.WithError(err).Error("SQL Query failed for getUserInitialSession")
			return nil, err
		}
		for rows.Next() {
			var userId string
			var coalesceId string
			var propertiesId string
			if err = rows.Scan(&userId, &coalesceId, &propertiesId); err != nil {
				logCtx.WithError(err).Error("SQL Parse failed")
				continue
			}
			userIdCoalUserIdMap[userId] = UserInfo{coalesceId, propertiesId}
		}
	}
	return userIdCoalUserIdMap, nil
}

// Returns the all the sessions (userId,attributionId,minTimestamp,maxTimestamp) for given
// users from given period including lookback
func getAllTheSessions(projectId uint64, sessionEventNameId uint64,
	query *AttributionQuery) (map[string]map[string]RangeTimestamp, []string, error) {

	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})
	sessionAttributionKey, err := getQuerySessionProperty(query.AttributionKey)
	if err != nil {
		return nil, nil, err
	}
	attributedSessionsByUserId := make(map[string]map[string]RangeTimestamp)
	userIdMap := make(map[string]bool)
	var userIdsWithSession []string
	from := getEffectiveFrom(query.From, query.LookbackDays)

	sessionAttributionKeySelect := "CASE WHEN sessions.properties->>? IS NULL THEN ? " +
		" WHEN sessions.properties->>? = '' THEN ? ELSE sessions.properties->>? END AS attribution_id"

	queryUserSessionTimeRange := "SELECT sessions.user_id, " + sessionAttributionKeySelect + ", " +
		" sessions.timestamp FROM events AS sessions " +
		" WHERE sessions.project_id=? AND sessions.event_name_id=? AND sessions.timestamp BETWEEN ? AND ?"
	var qParams []interface{}
	qParams = append(qParams, sessionAttributionKey, PropertyValueNone, sessionAttributionKey, PropertyValueNone,
		sessionAttributionKey, projectId, sessionEventNameId, from, query.To)
	rows, err := db.Raw(queryUserSessionTimeRange, qParams...).Rows()
	defer rows.Close()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return attributedSessionsByUserId, userIdsWithSession, err
	}
	for rows.Next() {
		var userId string
		var attributionId string
		var timestamp int64
		if err = rows.Scan(&userId, &attributionId, &timestamp); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		if _, ok := userIdMap[userId]; !ok {
			userIdsWithSession = append(userIdsWithSession, userId)
			userIdMap[userId] = true
		}
		if _, ok := attributedSessionsByUserId[userId]; ok {

			if timeRange, ok := attributedSessionsByUserId[userId][attributionId]; ok {
				timeRange.MinTimestamp = U.Min(timeRange.MinTimestamp, timestamp)
				timeRange.MaxTimestamp = U.Max(timeRange.MaxTimestamp, timestamp)
				attributedSessionsByUserId[userId][attributionId] = timeRange
			} else {
				sessionRange := RangeTimestamp{MinTimestamp: timestamp, MaxTimestamp: timestamp}
				attributedSessionsByUserId[userId][attributionId] = sessionRange
			}
		} else {
			attributedSessionsByUserId[userId] = make(map[string]RangeTimestamp)
			sessionRange := RangeTimestamp{MinTimestamp: timestamp, MaxTimestamp: timestamp}
			attributedSessionsByUserId[userId][attributionId] = sessionRange
		}
	}
	return attributedSessionsByUserId, userIdsWithSession, nil
}

// Returns the concatenated list of conversion event + funnel events names
func buildEventNamesPlaceholder(query *AttributionQuery) []string {
	enames := make([]string, 0)
	enames = append(enames, query.ConversionEvent.Name)
	for _, linkedEvent := range query.LinkedEvents {
		enames = append(enames, linkedEvent.Name)
	}
	return enames
}

// Return conversion event Id, list of all event_ids(Conversion and funnel events) and a Id to name mapping
func getEventInformation(projectId uint64, query *AttributionQuery) (uint64, map[string][]interface{}, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})
	names := buildEventNamesPlaceholder(query)
	conversionAndFunnelEventMap := make(map[string]bool)
	for _, name := range names {
		conversionAndFunnelEventMap[name] = true
	}
	if _, exists := conversionAndFunnelEventMap[U.EVENT_NAME_SESSION]; !exists {
		names = append(names, U.EVENT_NAME_SESSION)
	}
	eventNames, errCode := GetEventNamesByNames(projectId, names)
	if errCode != http.StatusFound {
		logCtx.Error("failed to find event names")
		return 0, nil, errors.New("failed to find event names")
	}
	// this is one to many mapping
	eventNameToId := make(map[string][]interface{})
	// this is one to one mapping
	eventNameIdToName := make(map[uint64]string)
	for _, event := range eventNames {
		eventNameId := event.ID
		eventName := event.Name
		eventNameIdToName[eventNameId] = eventName
		eventNameToId[eventName] = append(eventNameToId[eventName], eventNameId)
	}
	// there exists only one session event name per project
	if len(eventNameToId[U.EVENT_NAME_SESSION]) == 0 {
		logCtx.Error("$Session Name Id not found")
		return 0, nil, nil
	}
	sessionEventNameId := eventNameToId[U.EVENT_NAME_SESSION][0].(uint64)
	return sessionEventNameId, eventNameToId, nil
}

// Adds users who hit funnel event with given {event/user properties} to usersToBeAttributed
func getLinkedFunnelEventUsers(projectId uint64, query *AttributionQuery, eventNameToId map[string][]interface{},
	userIdInfo map[string]UserInfo, usersToBeAttributed *[]UserEventInfo, usersHitConversion []string) error {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})
	for _, linkedEvent := range query.LinkedEvents {
		// Part I - Fetch Users base on Event Hit satisfying events.properties
		linkedEventNameIds := eventNameToId[linkedEvent.Name]
		eventsPlaceHolder := "?"
		for i := 0; i < len(linkedEventNameIds)-1; i++ {
			eventsPlaceHolder += ",?"
		}
		var userIdList []string
		userIdHitEvent := make(map[string]bool)
		userPropertiesIdsInBatches := U.GetStringListAsBatch(usersHitConversion, USER_BATCH_SIZE)
		for _, users := range userPropertiesIdsInBatches {

			// add user batching
			usersPlaceHolder := U.GetValuePlaceHolder(len(users))
			value := U.GetInterfaceList(users)
			queryEventHits := "SELECT user_id FROM events WHERE events.project_id=? AND " +
				" timestamp >= ? AND timestamp <=? AND events.event_name_id IN (" + eventsPlaceHolder + ") " +
				" AND user_id = ANY (VALUES " + usersPlaceHolder + " ) "
			qParams := []interface{}{projectId, query.From, query.To}
			qParams = append(qParams, linkedEventNameIds...)
			qParams = append(qParams, value...)
			// add event filter
			wStmt, wParams, err := getFilterSQLStmtForEventProperties(query.ConversionEvent.Properties)
			if err != nil {
				return err
			}
			if wStmt != "" {
				queryEventHits = queryEventHits + " AND " + fmt.Sprintf("( %s )", wStmt)
				qParams = append(qParams, wParams...)
			}
			// fetch query results
			rows, err := db.Raw(queryEventHits, qParams...).Rows()
			defer rows.Close()
			if err != nil {
				logCtx.WithError(err).Error("SQL Query failed for queryEventHits")
				return err
			}
			for rows.Next() {
				var userId string
				if err = rows.Scan(&userId); err != nil {
					logCtx.WithError(err).Error("SQL Parse failed")
					continue
				}
				if _, ok := userIdHitEvent[userId]; !ok {
					userIdList = append(userIdList, userId)
					userIdHitEvent[userId] = true
				}
			}
		}

		// Part-II - Filter the fetched users from Part-I base on user_properties
		filteredUserIdList, err := applyUserPropertiesFilter(projectId, userIdList, userIdInfo, query)
		if err != nil {
			logCtx.WithError(err).Error("error while applying user properties")
			return err
		}

		// Part-III add the filtered users with eventId usersToBeAttributed
		for _, userId := range filteredUserIdList {
			*usersToBeAttributed = append(*usersToBeAttributed,
				UserEventInfo{userIdInfo[userId].coalUserId, linkedEvent.Name})
		}
	}
	return nil
}

// Applies user properties filter on given set of users and returns only the filters ones those match
func applyUserPropertiesFilter(projectId uint64, userIdList []string, userIdInfo map[string]UserInfo,
	query *AttributionQuery) ([]string, error) {

	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})

	wStmt, wParams, err := getFilterSQLStmtForUserProperties(query.ConversionEvent.Properties)
	if err != nil {
		return nil, err
	}
	// return the same list if there's no user_properties filter
	if wStmt == "" {
		return userIdList, nil
	}

	var userPropertiesIds []string
	// Use properties Ids to speed up the search from user_properties table
	for _, userId := range userIdList {
		userPropertiesIds = append(userPropertiesIds, userIdInfo[userId].propertiesId)
	}

	var filteredUserIdList []string
	filteredUserIdHitEvent := make(map[string]bool)
	userPropertiesIdsInBatches := U.GetStringListAsBatch(userPropertiesIds, USER_BATCH_SIZE)
	for _, users := range userPropertiesIdsInBatches {
		placeHolder := U.GetValuePlaceHolder(len(users))
		value := U.GetInterfaceList(users)
		queryUserIdCoalId := "SELECT user_id FROM user_properties WHERE id = ANY (VALUES " + placeHolder + " ) "
		var qParams []interface{}
		qParams = append(qParams, value...)
		// add user_properties filter
		if wStmt != "" {
			queryUserIdCoalId = queryUserIdCoalId + " AND " + fmt.Sprintf("( %s )", wStmt)
			qParams = append(qParams, wParams...)
		}
		rows, err := db.Raw(queryUserIdCoalId, qParams...).Rows()
		defer rows.Close()
		if err != nil {
			logCtx.WithError(err).Error("SQL Query failed for getUserInitialSession")
			return nil, err
		}
		for rows.Next() {
			var userId string
			if err = rows.Scan(&userId); err != nil {
				logCtx.WithError(err).Error("SQL Parse failed")
				continue
			}
			if _, ok := filteredUserIdHitEvent[userId]; !ok {
				filteredUserIdList = append(filteredUserIdList, userId)
				filteredUserIdHitEvent[userId] = true
			}
		}
	}
	return filteredUserIdList, nil
}

// Returns the list of eligible users who hit conversion event
func getConvertedUsers(projectId uint64, query *AttributionQuery, eventNameToIdList map[string][]interface{},
	userIdInfo map[string]UserInfo, usersToBeAttributed *[]UserEventInfo) ([]string, error) {

	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})

	conversionEventNameIds := eventNameToIdList[query.ConversionEvent.Name]
	placeHolder := "?"
	for i := 0; i < len(conversionEventNameIds)-1; i++ {
		placeHolder += ",?"
	}
	queryEventHits := "SELECT user_id FROM events WHERE events.project_id=? AND timestamp >= ? AND " +
		" timestamp <=? AND events.event_name_id IN (" + placeHolder + ") "
	qParams := []interface{}{projectId, query.From, query.To}
	qParams = append(qParams, conversionEventNameIds...)

	// add event filter
	wStmt, wParams, err := getFilterSQLStmtForEventProperties(query.ConversionEvent.Properties)
	if err != nil {
		return nil, err
	}
	if wStmt != "" {
		queryEventHits = queryEventHits + " AND " + fmt.Sprintf("( %s )", wStmt)
		qParams = append(qParams, wParams...)
	}
	// fetch query results
	rows, err := db.Raw(queryEventHits, qParams...).Rows()
	defer rows.Close()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed for queryEventHits")
		return nil, err
	}
	var userIdList []string
	userIdHitEvent := make(map[string]bool)
	for rows.Next() {
		var userId string
		if err = rows.Scan(&userId); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		if _, ok := userIdHitEvent[userId]; !ok {
			userIdList = append(userIdList, userId)
			userIdHitEvent[userId] = true
		}
	}

	// Part-II - Filter the fetched users from Part-I base on user_properties
	filteredUserIdList, err := applyUserPropertiesFilter(projectId, userIdList, userIdInfo, query)
	if err != nil {
		logCtx.WithError(err).Error("error while applying user properties")
		return nil, err
	}

	// add the filtered users against eventId
	for _, userId := range filteredUserIdList {
		*usersToBeAttributed = append(*usersToBeAttributed,
			UserEventInfo{userIdInfo[userId].coalUserId, query.ConversionEvent.Name})
	}
	return filteredUserIdList, nil
}

// Returns the effective From timestamp considering lookback days
func getEffectiveFrom(from int64, lookbackDays int) int64 {
	lookbackDaysTimestamp := int64(lookbackDays) * SECS_IN_A_DAY
	if LOOKBACK_CAP_IN_DAYS < lookbackDays {
		lookbackDaysTimestamp = int64(LOOKBACK_CAP_IN_DAYS) * SECS_IN_A_DAY
	}
	validFrom := from - lookbackDaysTimestamp
	return validFrom
}

// Clones a new map replacing userId by coalUserId.
func updateSessionsByCoalesceId(attributedSessionsByUserId map[string]map[string]RangeTimestamp,
	usersInfo map[string]UserInfo) map[string]map[string]RangeTimestamp {

	newSessionsMap := make(map[string]map[string]RangeTimestamp)
	for userId, attributionIdMap := range attributedSessionsByUserId {
		userInfo := usersInfo[userId]
		for attributionId, newTimeRange := range attributionIdMap {
			if _, ok := newSessionsMap[userInfo.coalUserId]; ok {
				if existingTimeRange, ok := newSessionsMap[userInfo.coalUserId][attributionId]; ok {
					// update the existing attribution first and last touch
					existingTimeRange.MinTimestamp = U.Min(existingTimeRange.MinTimestamp, newTimeRange.MinTimestamp)
					existingTimeRange.MaxTimestamp = U.Max(existingTimeRange.MaxTimestamp, newTimeRange.MaxTimestamp)
					newSessionsMap[userInfo.coalUserId][attributionId] = existingTimeRange
					continue
				}
				newSessionsMap[userInfo.coalUserId][attributionId] = newTimeRange
				continue
			}
			newSessionsMap[userInfo.coalUserId] = make(map[string]RangeTimestamp)
			newSessionsMap[userInfo.coalUserId][attributionId] = newTimeRange
		}
	}
	return newSessionsMap
}

// Maps the count distinct users session to campaign id and adds it to attributionData
func addWebsiteVisitorsInfo(from int64, to int64, attributionData map[string]*AttributionData,
	attributedSessionsByUserId map[string]map[string]RangeTimestamp) {

	userIdAttributionIdVisit := make(map[string]bool)
	for userId, attributionIdMap := range attributedSessionsByUserId {
		for attributionId, rangeTimestamp := range attributionIdMap {

			// only count sessions that happened during attribution period
			if rangeTimestamp.MaxTimestamp >= from && rangeTimestamp.MaxTimestamp <= to {

				if _, ok := attributionData[attributionId]; !ok {
					attributionData[attributionId] = &AttributionData{}
				}
				if _, ok := userIdAttributionIdVisit[getKey(userId, attributionId)]; ok {
					continue
				}
				attributionData[attributionId].WebsiteVisitors += 1
				userIdAttributionIdVisit[getKey(userId, attributionId)] = true
			}
		}
	}
}

// Merges 2 ids to create a string key
func getKey(id1 string, id2 string) string {
	return id1 + "|_|" + id2
}

// Adds channel data to attributionData based on campaign id. Campaign id with no matching channel data is left with empty name parameter
func AddPerformanceReportInfo(projectId uint64, attributionData map[string]*AttributionData,
	from, to int64, customerAccountId string) (string, error) {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectId": projectId, "Range": fmt.Sprintf("%d - %d", from, to)})
	performanceQuery := "SELECT value->>'campaign_id' AS campaign_id,  value->>'campaign_name' AS campaign_name, " +
		"SUM((value->>'impressions')::float) AS impressions, SUM((value->>'clicks')::float) AS clicks, " +
		"SUM((value->>'cost')::float)/1000000 AS total_cost FROM adwords_documents " +
		"where project_id = ? AND customer_account_id = ? AND type = ? AND timestamp between ? AND ? " +
		"group by value->>'campaign_id', campaign_name"
	rows, err := db.Raw(performanceQuery, projectId, customerAccountId, 5, getDateOnlyFromTimestamp(from),
		getDateOnlyFromTimestamp(to)).Rows()
	defer rows.Close()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return "", err
	}
	for rows.Next() {
		var campaignName string
		var campaignId string
		var impressions int
		var clicks int
		var spend float64
		if err = rows.Scan(&campaignId, &campaignName, &impressions, &clicks, &spend); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		matchingId := ""
		if _, campaignIdFound := attributionData[campaignId]; campaignIdFound {
			matchingId = campaignId
		} else if _, campaignNameFound := attributionData[campaignName]; campaignNameFound {
			matchingId = campaignName
		}
		if matchingId != "" {
			attributionData[matchingId].Name = campaignName
			attributionData[matchingId].Impressions = impressions
			attributionData[matchingId].Clicks = clicks
			attributionData[matchingId].Spend = spend
		}
	}

	currency, err := getAdwordsCurrency(projectId, customerAccountId, from, to)
	if err != nil {
		return "", err
	}
	return currency, nil
}

// Returns date in YYYYMMDD format
func getDateOnlyFromTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("20060102")
}

// Returns currency used for adwords customer_account_id
func getAdwordsCurrency(projectId uint64, customerAccountId string, from, to int64) (string, error) {

	queryCurrency := "SELECT value->>'currency_code' AS currency FROM adwords_documents " +
		" WHERE project_id=? AND customer_account_id=? AND type=? AND timestamp BETWEEN ? AND ? " +
		" ORDER BY timestamp DESC LIMIT 1"
	logCtx := log.WithField("ProjectId", projectId)
	db := C.GetServices().Db
	rows, err := db.Raw(queryCurrency, projectId, customerAccountId, 9, getDateOnlyFromTimestamp(from),
		getDateOnlyFromTimestamp(to)).Rows()
	if err != nil {
		logCtx.WithError(err).Error("failed to build meta for attribution query result")
		return "", err
	}
	defer rows.Close()
	var currency string
	for rows.Next() {
		if err = rows.Scan(&currency); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			return "", err
		}
	}
	return currency, nil
}
