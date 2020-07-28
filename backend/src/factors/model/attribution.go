package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type AttributionQueryUnit struct {
	Class string                  `json:"cl"`
	Query *AttributionQuery       `json:"query"`
	Meta  *map[string]interface{} `json:"meta"`
}

type AttributionQuery struct {
	CampaignMetrics        []string `json:"cm"`
	ConversionEvent        string   `json:"ce"`
	LinkedEvents           []string `json:"lfe"`
	AttributionKey         string   `json:"attribution_key"`
	AttributionMethodology string   `json:"attribution_methodology"`
	LoopbackDays           int      `json:"lbw"`
	From                   int64    `json:"from"`
	To                     int64    `json:"to"`
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
	ATTRIBUTION_METHOD_FIRST_TOUCH = "First_Touch"
	ATTRIBUTION_METHOD_LAST_TOUCH  = "Last_Touch"

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

// Maps the {attribution key} to the session properties field
func getQuerySessionProperty(attributionKey string) (string, error) {
	if attributionKey == ATTRIBUTION_KEY_CAMPAIGN {
		return U.EP_CAMPAIGN, nil
	} else if attributionKey == ATTRIBUTION_KEY_SOURCE {
		return U.EP_SOURCE, nil
	} // Todo (Anil) add here for adgroup, medium etc
	return "", errors.New("invalid query properties")
}

// Maps the {attribution key and attribution method} to the user properties field
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

// Attribute each user to the conversion event and linked event by attribution Id
func addUpLinkedFunnelEventCount(linkedEvents []string, attributionData map[string]*AttributionData, userAttributionData map[string]string, linkedUserAttributionData map[string]map[string]string) {

	linkedEventToPositionMap := make(map[string]int)
	for position, linkedEvent := range linkedEvents {
		linkedEventToPositionMap[linkedEvent] = position
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
	conversionEventUsers := fmt.Sprintf("%s - Users", query.ConversionEvent)
	result.Headers = append(result.Headers, conversionEventUsers)
	if len(query.LinkedEvents) > 0 {
		for _, event := range query.LinkedEvents {
			result.Headers = append(result.Headers, fmt.Sprintf("%s - Users", event))
		}
	}
}

/* Executes the Attribution using following steps:
*	1. Get all the sessions data (userId, attributionId, timestamp) for given period by attribution key
* 	2. Add the website visitor info using session data from step 1
*	3. Get all the users who hit the conversion event & Attribute the users from step 2 to the attributionId in step 1
*	4. Add performance data by attributionId
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

	// Get event_name_Id from event_names to avoid unnecessary events|event_names join
	conversionAndFunnelEventNameIdList, conversionEventNameId, sessionEventNameId, eventNameIdToEventNameMap, err := getEventInformation(projectId, query)
	if err != nil {
		return nil, err
	}

	// 1. Get all the sessions (userId, attributionId, timestamp) for given period by attribution key
	allSessions, userIdsWithSession, err := getAllTheSessions(projectId, sessionEventNameId,
		query.LoopbackDays, query.From, query.To, query.AttributionKey)

	// Map userId to COALESCE(users.customer_user_id,users.id)
	userIdCoalUserIdMap, err := getCoalesceUsersFromList(userIdsWithSession, projectId)
	if err != nil {
		return nil, err
	}

	// update all sessions with coalesce userId
	allSessionsByCoalesceId := mapAttributionFromUserIdToCoalUserId(allSessions, userIdCoalUserIdMap)

	// 2. Add website visitor information against the attribution key
	addWebsiteVisitorsByEventName(attributionData, allSessionsByCoalesceId)

	// 3. Using session data, do attribution based on given attribution methodology
	userConversionAttributionKeyData, userLinkedFunnelEventData, err := mapUserConversionEventByAttributionKey(projectId, query,
		allSessionsByCoalesceId, conversionEventNameId, conversionAndFunnelEventNameIdList, eventNameIdToEventNameMap, userIdCoalUserIdMap)
	if err != nil {
		return nil, err
	}

	// Aggregate the user count based on UserId-AttributionKey-EventName mapping
	addUpConversionEventCount(attributionData, userConversionAttributionKeyData)
	addUpLinkedFunnelEventCount(query.LinkedEvents, attributionData, userConversionAttributionKeyData, userLinkedFunnelEventData)

	// 4. Add the performance information against the attribution key
	currency, err := AddPerformanceReportByCampaign(projectId, attributionData, query.From, query.To, projectSetting.IntAdwordsCustomerAccountId)
	if err != nil {
		return nil, err
	}

	result.Rows = getRowsByMaps(attributionData, query)
	result.Meta.Currency = currency
	return result, nil
}

// converts a slice of int64 to a slice of interface
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
	nonMatchingRow := []interface{}{"none", 0, 0, float64(0), int64(0), int64(0)}
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
			row = append(row, attributionIdName, data.Impressions, data.Clicks, data.Spend, data.WebsiteVisitors, data.ConversionEventCount)
			row = append(row, getInterfaceList(data.LinkedEventsCount)...)
			rows = append(rows, row)
		} else {
			updateNonMatchingRow(&nonMatchingRow, data, query)
		}
	}
	rows = append(rows, nonMatchingRow)
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
func getCoalesceUsersFromList(userIdsWithSession []string, projectId uint64) (map[string]string, error) {

	userIdsInBatches := U.GetStringListAsBatch(userIdsWithSession, USER_BATCH_SIZE)
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"projectId": projectId})

	userIdCoalUserIdMap := make(map[string]string)
	for _, users := range userIdsInBatches {
		placeHolder := U.GetValuePlaceHolder(len(users))
		value := U.GetInterfaceList(users)
		queryUserIdCoalId := "SELECT id, COALESCE(users.customer_user_id,users.id) AS coal_user_id FROM users WHERE id = ANY (VALUES " + placeHolder + " )"
		rows, err := db.Raw(queryUserIdCoalId, value...).Rows()
		defer rows.Close()
		if err != nil {
			logCtx.WithError(err).Error("SQL Query failed for getUserInitialSession")
			return nil, err
		}
		for rows.Next() {
			var userId string
			var coalesceId string
			if err = rows.Scan(&userId, &coalesceId); err != nil {
				logCtx.WithError(err).Error("SQL Parse failed.")
				continue
			}
			userIdCoalUserIdMap[userId] = coalesceId
		}
	}
	return userIdCoalUserIdMap, nil
}

// Returns the all the sessions (userId,attributionId,minTimestamp,maxTimestamp) for given userIdsInBatch from given (from-lookback days) - to
func getAllTheSessions(projectId uint64, sessionEventNameId int, lookbackDays int,
	from int64, to int64, attributionKey string) (map[string]map[string]RangeTimestamp, []string, error) {

	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"projectId": projectId})

	sessionAttributionKey, err := getQuerySessionProperty(attributionKey)
	if err != nil {
		return nil, nil, err
	}

	attributedSessionsByUserId := make(map[string]map[string]RangeTimestamp)
	userIdMap := make(map[string]bool)
	var userIdsWithSession []string
	from = getEffectiveFrom(from, lookbackDays)

	sessionAttributionKeySelect := "CASE WHEN sessions.properties->>? IS NULL THEN ? " +
		" WHEN sessions.properties->>? = '' THEN ? ELSE sessions.properties->>? END AS attribution_id"

	queryUserSessionTimeRange := "SELECT sessions.user_id, " + sessionAttributionKeySelect + ", " +
		" sessions.timestamp FROM events AS sessions " +
		" WHERE sessions.project_id=? AND sessions.event_name_id=? AND sessions.timestamp BETWEEN ? AND ?"
	qparams := []interface{}{}
	qparams = append(qparams, sessionAttributionKey, PropertyValueNone, sessionAttributionKey, PropertyValueNone, sessionAttributionKey, projectId, sessionEventNameId, from, to)
	rows, err := db.Raw(queryUserSessionTimeRange, qparams...).Rows()
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
			logCtx.WithError(err).Error("SQL Parse failed.")
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

// Returns the concatenated list of conversion event + funnel events and [] of placeholders for SQL query
func buildEventNamesPlaceholder(query *AttributionQuery) (string, []interface{}) {
	var eventNamesStmt string
	eventNamesParams := make([]interface{}, 0)

	eventNamesStmt = eventNamesStmt + "?"
	eventNamesParams = append(eventNamesParams, query.ConversionEvent)
	for _, eventName := range query.LinkedEvents {
		eventNamesStmt = eventNamesStmt + ",?"
		eventNamesParams = append(eventNamesParams, eventName)
	}
	return eventNamesStmt, eventNamesParams
}

type UserEventInfo struct {
	coalUserId  string
	eventNameId int
	timestamp   int64
}

// Return conversion event Id, list of all event_ids(Conversion and funnel events) and a Id to name mapping
func getEventInformation(projectId uint64, query *AttributionQuery) ([]interface{}, int, int, map[int]string, error) {

	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"projectId": projectId})

	var err error

	inStmnt, inParams := buildEventNamesPlaceholder(query)
	conversionAndFunnelEventMap := make(map[string]bool)
	for _, conversionEvent := range inParams {
		conversionAndFunnelEventMap[conversionEvent.(string)] = true
	}
	if _, exists := conversionAndFunnelEventMap[U.EVENT_NAME_SESSION]; !exists {
		inStmnt = inStmnt + ",?"
		inParams = append(inParams, U.EVENT_NAME_SESSION)
	}
	qparams := []interface{}{projectId}
	qparams = append(qparams, inParams...)
	queryEventNameIds := "SELECT event_names.id, event_names.name FROM event_names WHERE event_names.project_id=? AND event_names.name IN (" + inStmnt + ")"
	rows, err := db.Raw(queryEventNameIds, qparams...).Rows()
	defer rows.Close()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed for getUserInitialSession")
		return nil, 0, 0, nil, err
	}
	eventNameEventNameIdMap := make(map[string]int)
	eventNameIdEventNameMap := make(map[int]string)
	var conversionAndFunnelEventNameIdList []interface{}
	for rows.Next() {
		var eventNameId int
		var eventName string
		if err = rows.Scan(&eventNameId, &eventName); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed.")
			continue
		}
		eventNameEventNameIdMap[eventName] = eventNameId
		eventNameIdEventNameMap[eventNameId] = eventName
		if _, exists := conversionAndFunnelEventMap[eventName]; exists {
			conversionAndFunnelEventNameIdList = append(conversionAndFunnelEventNameIdList, eventNameId)
		}
	}
	conversionEventNameId := eventNameEventNameIdMap[query.ConversionEvent]
	sessionEventNameId := eventNameEventNameIdMap[U.EVENT_NAME_SESSION]
	return conversionAndFunnelEventNameIdList, conversionEventNameId, sessionEventNameId, eventNameIdEventNameMap, nil
}

// This method maps the user to the attribution key based on users hitting conversion event/linked funnel event during given from-to period.
func mapUserConversionEventByAttributionKey(projectId uint64, query *AttributionQuery, allSessionsByCoalesceId map[string]map[string]RangeTimestamp,
	conversionEventNameId int, conversionAndFunnelEventNameIdList []interface{}, eventNameIdToEventNameMap map[int]string,
	userIdCoalUserIdMap map[string]string) (map[string]string, map[string]map[string]string, error) {

	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"projectId": projectId})

	usersAttribution := make(map[string]string)
	linkedEventUserCampaign := make(map[string]map[string]string)

	eventNameIdsPlaceHolder := "?"
	for i := 0; i < len(conversionAndFunnelEventNameIdList)-1; i++ {
		eventNameIdsPlaceHolder += ",?"
	}
	var allUserIdTimestampEventId []UserEventInfo
	userIdHitConversionEvent := make(map[string]bool)
	queryUserSessionTimeEventName := "SELECT events.user_id, events.event_name_id, " +
		" events.timestamp AS timestamp FROM events  WHERE events.project_id=? AND timestamp >= ? AND timestamp <=? " +
		" AND events.event_name_id IN (" + eventNameIdsPlaceHolder + ")"
	qparams := []interface{}{projectId, query.From, query.To}
	qparams = append(qparams, conversionAndFunnelEventNameIdList...)
	rows, err := db.Raw(queryUserSessionTimeEventName, qparams...).Rows()
	defer rows.Close()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed for queryUserSessionTimeEventName")
		return usersAttribution, linkedEventUserCampaign, err
	}
	for rows.Next() {
		var userId string
		var eventNameId int
		var timestamp int64
		if err = rows.Scan(&userId, &eventNameId, &timestamp); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed.")
			continue
		}
		if coalUserId, ok := userIdCoalUserIdMap[userId]; ok {
			allUserIdTimestampEventId = append(allUserIdTimestampEventId, UserEventInfo{coalUserId, eventNameId, timestamp})
			if eventNameId == conversionEventNameId {
				userIdHitConversionEvent[coalUserId] = true
			}
		}
	}
	// prune out the users who did not hit conversion event
	var validUserIdTimestampEventId []UserEventInfo
	for _, value := range allUserIdTimestampEventId {
		if _, ok := userIdHitConversionEvent[value.coalUserId]; ok {
			validUserIdTimestampEventId = append(validUserIdTimestampEventId, value)
		}
	}

	validFrom := getEffectiveFrom(query.From, query.LoopbackDays)
	validTo := query.To
	switch query.AttributionMethodology {

	case ATTRIBUTION_METHOD_FIRST_TOUCH:
		computeForFirstTouch(validFrom, validTo, query.ConversionEvent, validUserIdTimestampEventId, eventNameIdToEventNameMap,
			allSessionsByCoalesceId, &usersAttribution, &linkedEventUserCampaign)
		break

	case ATTRIBUTION_METHOD_LAST_TOUCH:
		computeForLastTouch(validFrom, validTo, query.ConversionEvent, validUserIdTimestampEventId, eventNameIdToEventNameMap,
			allSessionsByCoalesceId, &usersAttribution, &linkedEventUserCampaign)
		break

	default:
		break
	}
	return usersAttribution, linkedEventUserCampaign, nil
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

// FIRST_TOUCH based attribution attributes user session to the attributionKey by the MaxTimestamp
func computeForFirstTouch(from int64, to int64, conversionEvent string, userIdUserPropIdTimestampEventId []UserEventInfo,
	eventNameIdToEventNameMap map[int]string, attributedSessionsByUserId map[string]map[string]RangeTimestamp, usersAttribution *map[string]string,
	linkedEventUserCampaign *map[string]map[string]string) {

	userIdAttributionIdTimestamp := make(map[string]int64)
	for _, val := range userIdUserPropIdTimestampEventId {
		userId := val.coalUserId
		eventName := eventNameIdToEventNameMap[val.eventNameId]
		for attributionKey, minMaxTimestamp := range attributedSessionsByUserId[userId] {

			if minMaxTimestamp.MinTimestamp >= from && minMaxTimestamp.MinTimestamp <= to {
				if eventName == conversionEvent {
					if _, exists := (*usersAttribution)[userId]; exists {
						// Here if we have already attributed this user to some attribution key,
						// update only when the current minTime is lower than last minTime
						minTimestamp := userIdAttributionIdTimestamp[getKey(userId, attributionKey)]
						if minTimestamp < minMaxTimestamp.MinTimestamp {
							continue
						}
					}
					(*usersAttribution)[userId] = attributionKey
					userIdAttributionIdTimestamp[getKey(userId, attributionKey)] = minMaxTimestamp.MinTimestamp
				} else {
					if _, exist := (*linkedEventUserCampaign)[eventName]; !exist {
						(*linkedEventUserCampaign)[eventName] = make(map[string]string)
					}
					(*linkedEventUserCampaign)[eventName][userId] = attributionKey
				}
			}
		}
	}
}

// LAST_TOUCH based attribution attributes user session to the attributionKey by the MaxTimestamp
func computeForLastTouch(from int64, to int64, conversionEvent string, userIdUserPropIdTimestampEventId []UserEventInfo,
	eventNameIdToEventNameMap map[int]string, userInitialSession map[string]map[string]RangeTimestamp, usersAttribution *map[string]string,
	linkedEventUserCampaign *map[string]map[string]string) {

	userIdAttributionIdTimestamp := make(map[string]int64)
	for _, val := range userIdUserPropIdTimestampEventId {
		userId := val.coalUserId
		eventId := val.eventNameId
		eventName := eventNameIdToEventNameMap[eventId]
		for attributionKey, minMaxTimestamp := range userInitialSession[userId] {

			if minMaxTimestamp.MaxTimestamp >= from && minMaxTimestamp.MaxTimestamp <= to {
				if eventName == conversionEvent {
					if _, exists := (*usersAttribution)[userId]; exists {
						// Here if we have already attributed this user to some attribution key,
						// update only when the current maxTime is larger than last maxTime
						maxTimestamp := userIdAttributionIdTimestamp[getKey(userId, attributionKey)]
						if maxTimestamp > minMaxTimestamp.MaxTimestamp {
							continue
						}
					}
					(*usersAttribution)[userId] = attributionKey
					userIdAttributionIdTimestamp[getKey(userId, attributionKey)] = minMaxTimestamp.MaxTimestamp
				} else {
					if _, exist := (*linkedEventUserCampaign)[eventName]; !exist {
						(*linkedEventUserCampaign)[eventName] = make(map[string]string)
					}
					(*linkedEventUserCampaign)[eventName][userId] = attributionKey
				}
			}
		}
	}
}

// clones a new map replacing userId by coalUserId
func mapAttributionFromUserIdToCoalUserId(attributedSessionsByUserId map[string]map[string]RangeTimestamp, userIdCoalUserIdMap map[string]string) map[string]map[string]RangeTimestamp {

	newMapForAttributedSessionsByCoalUserId := make(map[string]map[string]RangeTimestamp)
	for userId, attributionIdMap := range attributedSessionsByUserId {
		coalUserId := userIdCoalUserIdMap[userId]
		for attributionId, timeRange := range attributionIdMap {
			if _, ok := newMapForAttributedSessionsByCoalUserId[coalUserId]; !ok {
				newMapForAttributedSessionsByCoalUserId[coalUserId] = make(map[string]RangeTimestamp)
			}
			newMapForAttributedSessionsByCoalUserId[coalUserId][attributionId] = timeRange
		}
	}
	return newMapForAttributedSessionsByCoalUserId
}

// Counts distinct users session by campaign id and adds it to attributionData
func addWebsiteVisitorsByEventName(attributionData map[string]*AttributionData,
	attributedSessionsByUserId map[string]map[string]RangeTimestamp) {

	userIdAttributionIdVisit := make(map[string]bool)
	for userId, attributionIdMap := range attributedSessionsByUserId {
		for attributionId, _ := range attributionIdMap {
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

func getKey(id1 string, id2 string) string {
	return id1 + "|_|" + id2
}

// Adds channel data to attributionData based on campaign id. Campaign id with no matching channel data is left with empty name parameter
func AddPerformanceReportByCampaign(projectId uint64, attributionData map[string]*AttributionData, from, to int64,
	customerAccountId *string) (string, error) {
	db := C.GetServices().Db

	logCtx := log.WithFields(log.Fields{"projectId": projectId, "range": fmt.Sprintf("%d - %d", from, to)})

	// TODO (Anil) break this query after discussion
	rows, err := db.Raw("select value->>'campaign_id' AS campaign_id,  value->>'campaign_name' AS campaign_name, "+
		"SUM((value->>'impressions')::float) AS impressions, SUM((value->>'clicks')::float) AS clicks, "+
		"SUM((value->>'cost')::float)/1000000 AS total_cost FROM adwords_documents "+
		"where project_id = ? AND customer_account_id = ? AND type = ? AND timestamp between ? AND ? "+
		"group by value->>'campaign_id', campaign_name", projectId, customerAccountId, 5, getDateOnlyFromTimestamp(from), getDateOnlyFromTimestamp(to)).Rows()

	defer rows.Close()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed.")
		return "", err
	}
	for rows.Next() {
		var campaignName string
		var campaignId string
		var impressions int
		var clicks int
		var spend float64
		if err = rows.Scan(&campaignId, &campaignName, &impressions, &clicks, &spend); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed.")
			continue
		}
		if _, exists := attributionData[campaignId]; exists { // only matching campaign id is filled
			attributionData[campaignId].Name = campaignName
			attributionData[campaignId].Impressions = impressions
			attributionData[campaignId].Clicks = clicks
			attributionData[campaignId].Spend = spend
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
func getAdwordsCurrency(projectId uint64, customerAccountId *string, from, to int64) (string, error) {
	stmnt := "SELECT value->>'currency_code' AS currency FROM adwords_documents " +
		" WHERE project_id=? AND customer_account_id=? AND type=? AND timestamp BETWEEN ? AND ? " +
		" ORDER BY timestamp DESC LIMIT 1"

	logCtx := log.WithField("project_id", projectId)

	db := C.GetServices().Db
	rows, err := db.Raw(stmnt,
		projectId, customerAccountId, 9, getDateOnlyFromTimestamp(from), getDateOnlyFromTimestamp(to)).Rows()

	if err != nil {
		logCtx.WithError(err).Error("Failed to build meta for channel query result.")
		return "", err
	}
	defer rows.Close()

	var currency string
	for rows.Next() {
		if err = rows.Scan(&currency); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed.")
			return "", err
		}
	}

	return currency, nil
}
