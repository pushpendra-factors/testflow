package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sort"
	"strings"
)

type AttributionQueryUnit struct {
	Class string                  `json:"cl"`
	Query *AttributionQuery       `json:"query"`
	Meta  *map[string]interface{} `json:"meta"`
}

type AttributionQuery struct {
	CampaignMetrics               []string                   `json:"cm"`
	ConversionEvent               QueryEventWithProperties   `json:"ce"`
	ConversionEventCompare        QueryEventWithProperties   `json:"ce_c"`
	LinkedEvents                  []QueryEventWithProperties `json:"lfe"`
	AttributionKey                string                     `json:"attribution_key"`
	AttributionKeyFilter          []AttributionKeyFilter     `json:"attribution_key_f"`
	AttributionMethodology        string                     `json:"attribution_methodology"`
	AttributionMethodologyCompare string                     `json:"attribution_methodology_c"`
	LookbackDays                  int                        `json:"lbw"`
	From                          int64                      `json:"from"`
	To                            int64                      `json:"to"`
}

type AttributionKeyFilter struct {
	AttributionKey string   `json:"attribution_key"`
	Operator       string   `json:"op"` // contains or notContains
	Values         []string `json:"va"`
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
	ATTRIBUTION_METHOD_LINEAR                 = "Linear"
	ATTRIBUTION_KEY_CAMPAIGN                  = "Campaign"
	ATTRIBUTION_KEY_SOURCE                    = "Source"
	ATTRIBUTION_KEY_ADGROUP                   = "AdGroup"
	ATTRIBUTION_KEY_KEYWORD                   = "Keyword"

	ADWORDS_CLICK_REPORT_TYPE    = 4
	ADWORDS_CAMPAIGN_REPORT_TYPE = 5

	SECS_IN_A_DAY        = int64(86400)
	LOOKBACK_CAP_IN_DAYS = 180
	USER_BATCH_SIZE      = 3000
)

var ATTRIBUTION_FIXED_HEADERS = []string{"Impressions", "Clicks", "Spend", "Website Visitors"}

type AttributionData struct {
	Name                        string
	Impressions                 int
	Clicks                      int
	Spend                       float64
	WebsiteVisitors             int64
	ConversionEventCount        float64
	LinkedEventsCount           []float64
	ConversionEventCompareCount float64
}

type UserInfo struct {
	CoalUserID   string
	PropertiesID string
	Timestamp    int64
}

type UserIDPropID struct {
	UserID       string
	PropertiesID string
	Timestamp    int64
}

type UserEventInfo struct {
	CoalUserID string
	EventName  string
}

type CampaignInfo struct {
	AdgroupName  string
	CampaignName string
	AdID         string
}

// Maps the {attribution key} to the session properties field
func getQuerySessionProperty(attributionKey string) (string, error) {
	if attributionKey == ATTRIBUTION_KEY_CAMPAIGN {
		return U.EP_CAMPAIGN, nil
	} else if attributionKey == ATTRIBUTION_KEY_SOURCE {
		return U.EP_SOURCE, nil
	} else if attributionKey == ATTRIBUTION_KEY_ADGROUP {
		return U.EP_ADGROUP, nil
	} else if attributionKey == ATTRIBUTION_KEY_KEYWORD {
		return U.EP_KEYWORD, nil
	}
	return "", errors.New("invalid query properties")
}

// Adds common column names and linked events as header to the result rows.
func addHeadersByAttributionKey(result *QueryResult, query *AttributionQuery) {
	attributionKey := query.AttributionKey
	result.Headers = append(append(result.Headers, attributionKey), ATTRIBUTION_FIXED_HEADERS...)
	conversionEventUsers := fmt.Sprintf("%s - Users", query.ConversionEvent.Name)
	costPerConversion := fmt.Sprintf("Cost Per Conversion")
	conversionEventCompareUsers := fmt.Sprintf("Compare - Users")
	compareCostPerConversion := fmt.Sprintf("Compare Cost Per Conversion")
	result.Headers = append(result.Headers, conversionEventUsers, costPerConversion,
		conversionEventCompareUsers, compareCostPerConversion)
	if len(query.LinkedEvents) > 0 {
		for _, event := range query.LinkedEvents {
			result.Headers = append(result.Headers, fmt.Sprintf("%s - Users", event.Name))
		}
	}
}

func isValidAttributionKeyValue(attributionKeyType string, keyValue string, filters []AttributionKeyFilter) bool {

	for _, filter := range filters {
		// Currently only supporting matching key filters only
		if filter.AttributionKey == attributionKeyType {
			if filter.Operator == ContainsOpStr {
				if !U.StringValueIn(keyValue, filter.Values) {
					return false
				}
			} else if filter.Operator == NotContainsOpStr {
				if U.StringValueIn(keyValue, filter.Values) {
					return false
				}
			}
		}
	}
	return true
}

/* Executes the Attribution using following steps:
	1. Get all the sessions data (userId, attributionId, timestamp) for given period by attribution key
 	2. Add the website visitor info using session data from step 1
	3. i) 	Find out users who hit conversion event applying filter
	  ii)	Using users from 3.i) find out users who hit linked funnel event applying filter
	4. Apply attribution methodology
	5. Add performance data by attributionId
*/
func ExecuteAttributionQuery(projectID uint64, query *AttributionQuery) (*QueryResult, error) {

	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, errors.New("failed to get project Settings")
	}
	if projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "" {
		return nil, errors.New("execute attribution query failed as no ad-words customer account id found")
	}
	sessionEventNameID, eventNameToIDList, err := getEventInformation(projectID, query)
	if err != nil {
		return nil, err
	}

	// 1. Get all the sessions (userId, attributionId, timestamp) for given period by attribution key
	_sessions, sessionUsers, err := getAllTheSessions(projectID, sessionEventNameID, query,
		*projectSetting.IntAdwordsCustomerAccountId)
	usersInfo, err := getCoalesceIDFromUserIDs(sessionUsers, projectID)
	if err != nil {
		return nil, err
	}
	sessions := updateSessionsMapWithCoalesceID(_sessions, usersInfo)

	isCompare := false
	var attributionData map[string]*AttributionData
	if query.AttributionMethodologyCompare != "" {
		// Two AttributionMethodologies comparison
		isCompare = true
		attributionData, err = runAttributionForMethodologyComparison(projectID,
			query.From, query.To,
			query.ConversionEvent.Name,
			query.ConversionEvent.Properties,
			query.AttributionMethodology,
			query.AttributionMethodologyCompare, // run for AttributionMethodologyCompare
			eventNameToIDList, sessions, query.LookbackDays)

	} else if query.ConversionEventCompare.Name != "" {
		// Two events comparison
		isCompare = true
		attributionData, err = runAttribution(projectID,
			query.From, query.To,
			query.ConversionEvent.Name,
			query.ConversionEvent.Properties,
			query.LinkedEvents,
			query.AttributionMethodology,
			eventNameToIDList, sessions, query.LookbackDays)

		if err != nil {
			return nil, err
		}

		attributionCompareData, err := runAttribution(projectID,
			query.From, query.To,
			query.ConversionEventCompare.Name, // run for ConversionEventCompare
			query.ConversionEventCompare.Properties,
			query.LinkedEvents,
			query.AttributionMethodology,
			eventNameToIDList, sessions, query.LookbackDays)

		if err != nil {
			return nil, err
		}

		// merge compare data into attributionData
		for key, _ := range attributionData {
			if _, exists := attributionCompareData[key]; exists {
				attributionData[key].ConversionEventCompareCount = attributionCompareData[key].ConversionEventCount
			} else {
				attributionData[key].ConversionEventCompareCount = 0
			}
		}
	} else {
		// single event attribution
		attributionData, err = runAttribution(projectID,
			query.From, query.To,
			query.ConversionEvent.Name,
			query.ConversionEvent.Properties,
			query.LinkedEvents,
			query.AttributionMethodology,
			eventNameToIDList, sessions, query.LookbackDays)
	}

	if err != nil {
		return nil, err
	}

	addWebsiteVisitorsInfo(query.From, query.To, attributionData, sessions)

	// 5. Add the performance information
	currency, err := AddPerformanceReportInfo(projectID, attributionData, query.From, query.To,
		*projectSetting.IntAdwordsCustomerAccountId)
	if err != nil {
		return nil, err
	}

	result := &QueryResult{}
	addHeadersByAttributionKey(result, query)
	result.Rows = getRowsByMaps(attributionData, query.LinkedEvents, isCompare)
	result.Meta.Currency = currency
	return result, nil
}

func runAttributionForMethodologyComparison(projectID uint64, from, to int64,
	goalEventName string, goalEventProperties []QueryProperty, attributionMethodology,
	attributionMethodologyCompare string, eventNameToIDList map[string][]interface{},
	sessions map[string]map[string]RangeTimestamp, lookbackDays int) (map[string]*AttributionData, error) {

	// empty linkedEvents as they are not analyzed in compare events.
	var linkedEvents []QueryEventWithProperties

	// 3. Fetch users who hit conversion event
	// coalUserIdConversionTimestamp := make(map[string]int64)
	userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err := getConvertedUsers(projectID,
		goalEventName, goalEventProperties, from, to,
		eventNameToIDList)
	if err != nil {
		return nil, err
	}

	// Add users who hit conversion event
	var usersToBeAttributed []UserEventInfo
	for key, _ := range coalescedIDToInfoConverted {
		usersToBeAttributed = append(usersToBeAttributed, UserEventInfo{key,
			goalEventName})
	}

	err = getLinkedFunnelEventUsers(projectID, from, to, linkedEvents, eventNameToIDList, userIDToInfoConverted,
		&usersToBeAttributed)
	if err != nil {
		return nil, err
	}

	// attribution based on given attribution methodology
	userConversionHit, _, err := ApplyAttribution(attributionMethodology, goalEventName, usersToBeAttributed, sessions, coalUserIdConversionTimestamp, 0)
	if err != nil {
		return nil, err
	}

	attributionData := addUpConversionEventCount(userConversionHit)

	// attribution based on given attributionMethodologyCompare methodology
	userConversionCompareHit, _, err := ApplyAttribution(attributionMethodologyCompare,
		goalEventName, usersToBeAttributed, sessions, coalUserIdConversionTimestamp,
		lookbackDays)
	if err != nil {
		return nil, err
	}
	attributionDataCompare := addUpConversionEventCount(userConversionCompareHit)

	// merge compare data into attributionData
	for key, _ := range attributionData {
		if _, exists := attributionDataCompare[key]; exists {
			attributionData[key].ConversionEventCompareCount = attributionDataCompare[key].ConversionEventCount
		} else {
			attributionData[key].ConversionEventCompareCount = 0
		}
	}

	return attributionData, nil
}

func runAttribution(projectID uint64,
	from, to int64,
	goalEventName string,
	goalEventProperties []QueryProperty,
	linkedEvents []QueryEventWithProperties,
	attributionMethodology string,
	eventNameToIDList map[string][]interface{},
	sessions map[string]map[string]RangeTimestamp,
	lookbackDays int) (map[string]*AttributionData, error) {

	// 3. Fetch users who hit conversion event
	userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err := getConvertedUsers(projectID,
		goalEventName, goalEventProperties, from, to,
		eventNameToIDList)
	if err != nil {
		return nil, err
	}

	// Add users who hit conversion event
	var usersToBeAttributed []UserEventInfo
	for key, _ := range coalescedIDToInfoConverted {
		usersToBeAttributed = append(usersToBeAttributed, UserEventInfo{key,
			goalEventName})
	}

	err = getLinkedFunnelEventUsers(projectID, from, to, linkedEvents, eventNameToIDList, userIDToInfoConverted,
		&usersToBeAttributed)
	if err != nil {
		return nil, err
	}

	// 4. Apply attribution based on given attribution methodology
	userConversionHit, userLinkedFEHit, err := ApplyAttribution(attributionMethodology,
		goalEventName,
		usersToBeAttributed, sessions, coalUserIdConversionTimestamp, lookbackDays)
	if err != nil {
		return nil, err
	}

	attributionData := make(map[string]*AttributionData)
	attributionData = addUpConversionEventCount(userConversionHit)
	addUpLinkedFunnelEventCount(linkedEvents, attributionData, userLinkedFEHit)
	return attributionData, nil
}

// Converts a slice of float64 to a slice of interface.
func getInterfaceListFromFloat64(data []float64) []interface{} {
	var list []interface{}
	for _, val := range data {
		list = append(list, []interface{}{val})
	}
	return list
}

// Returns result in from of metrics. For empty attribution id, the values are accumulated into "$none".
func getRowsByMaps(attributionData map[string]*AttributionData,
	linkedEvents []QueryEventWithProperties, isCompare bool) [][]interface{} {

	rows := make([][]interface{}, 0)
	nonMatchingRow := []interface{}{"none", 0, 0, float64(0), int64(0), float64(0), float64(0),
		float64(0), float64(0)}
	for i := 0; i < len(linkedEvents); i++ {
		nonMatchingRow = append(nonMatchingRow, float64(0))
	}
	for key, data := range attributionData {
		attributionIdName := data.Name
		if attributionIdName == "" {
			attributionIdName = key
		}
		if attributionIdName != "" {
			var row []interface{}
			cpc := 0.0
			if data.ConversionEventCount != 0.0 {
				cpc = data.Spend / data.ConversionEventCount
			}
			if isCompare {
				cpcCompare := 0.0
				if data.ConversionEventCompareCount != 0.0 {
					cpcCompare = data.Spend / data.ConversionEventCompareCount
				}
				row = append(row, attributionIdName, data.Impressions, data.Clicks, data.Spend,
					data.WebsiteVisitors, data.ConversionEventCount, cpc,
					data.ConversionEventCompareCount, cpcCompare)
			} else {
				row = append(row, attributionIdName, data.Impressions, data.Clicks, data.Spend,
					data.WebsiteVisitors, data.ConversionEventCount, cpc, 0.0, 0.0)
			}
			row = append(row, getInterfaceListFromFloat64(data.LinkedEventsCount)...)
			rows = append(rows, row)
		} else {
			updateNonMatchingRow(&nonMatchingRow, data, linkedEvents)
		}
	}
	rows = append(rows, nonMatchingRow)
	// sort the rows by conversionEvent
	sort.Slice(rows, func(i, j int) bool {
		return rows[i][5].(float64) > rows[j][5].(float64)
	})
	return rows
}

func updateNonMatchingRow(nonMatchingRow *[]interface{}, data *AttributionData,
	linkedEvents []QueryEventWithProperties) {
	(*nonMatchingRow)[1] = (*nonMatchingRow)[1].(int) + data.Impressions
	(*nonMatchingRow)[2] = (*nonMatchingRow)[2].(int) + data.Clicks
	(*nonMatchingRow)[3] = (*nonMatchingRow)[3].(float64) + data.Spend
	(*nonMatchingRow)[4] = (*nonMatchingRow)[4].(int64) + data.WebsiteVisitors
	(*nonMatchingRow)[5] = (*nonMatchingRow)[5].(float64) + data.ConversionEventCount
	(*nonMatchingRow)[6] = (*nonMatchingRow)[6].(float64) + 0.0
	(*nonMatchingRow)[7] = (*nonMatchingRow)[7].(float64) + 0.0
	(*nonMatchingRow)[8] = (*nonMatchingRow)[8].(float64) + 0.0
	for index := 0; index < len(linkedEvents); index++ {
		if index < len(data.LinkedEventsCount) {
			(*nonMatchingRow)[5+index+1] = (*nonMatchingRow)[5+index+1].(float64) + data.LinkedEventsCount[index]
		}
	}
}

// Groups all unique users by attributionId and adds it to attributionData
func addUpConversionEventCount(usersIdAttributionIdMap map[string][]string) map[string]*AttributionData {
	attributionData := make(map[string]*AttributionData)
	for _, attributionKeys := range usersIdAttributionIdMap {
		weight := 1 / float64(len(attributionKeys))
		for _, key := range attributionKeys {
			if _, exists := attributionData[key]; !exists {
				attributionData[key] = &AttributionData{}
			}
			attributionData[key].ConversionEventCount += weight
		}
	}
	return attributionData
}

// Attribute each user to the conversion event and linked event by attribution Id.
func addUpLinkedFunnelEventCount(linkedEvents []QueryEventWithProperties,
	attributionData map[string]*AttributionData, linkedUserAttributionData map[string]map[string][]string) {

	linkedEventToPositionMap := make(map[string]int)
	for position, linkedEvent := range linkedEvents {
		linkedEventToPositionMap[linkedEvent.Name] = position
	}
	// fill up all the linked events count with 0 value
	for _, attributionRow := range attributionData {
		if attributionRow != nil {
			for len(attributionRow.LinkedEventsCount) < len(linkedEvents) {
				attributionRow.LinkedEventsCount = append(attributionRow.LinkedEventsCount, 0.0)
			}
		}
	}
	// update linked up events with event hit count
	for linkedEventName, userIdAttributionIdMap := range linkedUserAttributionData {
		for _, attributionKeys := range userIdAttributionIdMap {
			weight := 1 / float64(len(attributionKeys))
			for _, key := range attributionKeys {
				attributionRow := attributionData[key]
				if attributionRow != nil {
					attributionRow.LinkedEventsCount[linkedEventToPositionMap[linkedEventName]] += weight
				}
			}
		}
	}
}

// getCoalesceIDFromUserIDs returns the map of coalesce userId for given list of users
func getCoalesceIDFromUserIDs(userIDs []string, projectID uint64) (map[string]UserInfo, error) {

	userIDsInBatches := U.GetStringListAsBatch(userIDs, USER_BATCH_SIZE)
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectId": projectID})
	userIDToCoalUserIDMap := make(map[string]UserInfo)
	for _, users := range userIDsInBatches {
		placeHolder := U.GetValuePlaceHolder(len(users))
		value := U.GetInterfaceList(users)
		queryUserIDCoalID := "SELECT id, COALESCE(users.customer_user_id,users.id) AS coal_user_id, " +
			" properties_id FROM users WHERE id = ANY (VALUES " + placeHolder + " )"
		rows, err := db.Raw(queryUserIDCoalID, value...).Rows()
		if err != nil {
			logCtx.WithError(err).Error("SQL Query failed for getUserInitialSession")
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var userID string
			var coalesceID string
			var propertiesID string

			if err = rows.Scan(&userID, &coalesceID, &propertiesID); err != nil {
				logCtx.WithError(err).Error("SQL Parse failed")
				continue
			}
			userIDToCoalUserIDMap[userID] = UserInfo{coalesceID, propertiesID, 0}
		}
	}
	return userIDToCoalUserIDMap, nil
}

// Returns the all the sessions (userId,attributionId,minTimestamp,maxTimestamp) for given
// users from given period including lookback
func getAllTheSessions(projectId uint64, sessionEventNameId uint64,
	query *AttributionQuery, adwordsAccountId string) (map[string]map[string]RangeTimestamp, []string, error) {

	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})
	gclIDBasedCampaign, err := GetGCLIDBasedCampaignInfo(projectId, query.From, query.To, adwordsAccountId)
	if err != nil {
		return nil, nil, err
	}

	attributionEventKey, err := getQuerySessionProperty(query.AttributionKey)
	if err != nil {
		return nil, nil, err
	}

	attributedSessionsByUserId := make(map[string]map[string]RangeTimestamp)
	userIdMap := make(map[string]bool)
	var userIdsWithSession []string
	from := getEffectiveFrom(query.From, query.LookbackDays)

	caseSelectStmt := "CASE WHEN sessions.properties->>? IS NULL THEN ? " +
		" WHEN sessions.properties->>? = '' THEN ? ELSE sessions.properties->>? END"

	queryUserSessionTimeRange := "SELECT sessions.user_id, " + caseSelectStmt + " AS attribution_id, " + caseSelectStmt + " AS gcl_id, " +
		" sessions.timestamp FROM events AS sessions " +
		" WHERE sessions.project_id=? AND sessions.event_name_id=? AND sessions.timestamp BETWEEN ? AND ?"
	var qParams []interface{}
	qParams = append(qParams, attributionEventKey, PropertyValueNone, attributionEventKey, PropertyValueNone,
		attributionEventKey, U.EP_GCLID, PropertyValueNone, U.EP_GCLID, PropertyValueNone, U.EP_GCLID, projectId,
		sessionEventNameId, from, query.To)
	rows, err := db.Raw(queryUserSessionTimeRange, qParams...).Rows()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return attributedSessionsByUserId, userIdsWithSession, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		var attributionId string
		var gclID string
		var timestamp int64
		if err = rows.Scan(&userID, &attributionId, &gclID, &timestamp); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		// apply filter at extracting session level itself
		if !isValidAttributionKeyValue(query.AttributionKey, attributionId, query.AttributionKeyFilter) {
			continue
		}
		if _, ok := userIdMap[userID]; !ok {
			userIdsWithSession = append(userIdsWithSession, userID)
			userIdMap[userID] = true
		}

		// Override GCLID based campaign info if presents
		if gclID != PropertyValueNone {
			attributionIdBasedOnGclID := getGCLIDAttributionValue(gclIDBasedCampaign, gclID, attributionEventKey)
			// In cases where GCLID is present in events, but not in adwords report (as users tend to bookmark expired URLs),
			// fallback is attributionId
			if attributionIdBasedOnGclID != PropertyValueNone {
				attributionId = attributionIdBasedOnGclID
			}
		}

		// add session info uniquely for user-attributionId pair
		if _, ok := attributedSessionsByUserId[userID]; ok {

			if timeRange, ok := attributedSessionsByUserId[userID][attributionId]; ok {
				timeRange.MinTimestamp = U.Min(timeRange.MinTimestamp, timestamp)
				timeRange.MaxTimestamp = U.Max(timeRange.MaxTimestamp, timestamp)
				attributedSessionsByUserId[userID][attributionId] = timeRange
			} else {
				sessionRange := RangeTimestamp{MinTimestamp: timestamp, MaxTimestamp: timestamp}
				attributedSessionsByUserId[userID][attributionId] = sessionRange
			}
		} else {
			attributedSessionsByUserId[userID] = make(map[string]RangeTimestamp)
			sessionRange := RangeTimestamp{MinTimestamp: timestamp, MaxTimestamp: timestamp}
			attributedSessionsByUserId[userID][attributionId] = sessionRange
		}
	}
	return attributedSessionsByUserId, userIdsWithSession, nil
}

// Returns the matching value for GCLID, if not found returns $none
func getGCLIDAttributionValue(gclIDBasedCampaign map[string]CampaignInfo, gclID string, attributionKey string) string {

	if value, ok := gclIDBasedCampaign[gclID]; ok {
		switch attributionKey {
		case U.EP_ADGROUP:
			return value.AdgroupName
		case U.EP_CAMPAIGN:
			return value.CampaignName
		default:
			// No enrichment for Source and Keyword via GCLID
			return PropertyValueNone
		}
	}
	return PropertyValueNone
}

// Returns the concatenated list of conversion event + funnel events names
func buildEventNamesPlaceholder(query *AttributionQuery) []string {
	enames := make([]string, 0)
	enames = append(enames, query.ConversionEvent.Name)
	// add name of compare event if given
	if query.ConversionEventCompare.Name != "" {
		enames = append(enames, query.ConversionEventCompare.Name)
	}
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
		return 0, nil, errors.New("$Session Name Id not found")
	}
	if len(eventNameToId[query.ConversionEvent.Name]) == 0 {
		logCtx.Error("conversion event name : " + query.ConversionEvent.Name + " not found")
		return 0, nil, errors.New("conversion event name : " + query.ConversionEvent.Name + " not found")
	}
	for _, linkedEvent := range query.LinkedEvents {
		if len(eventNameToId[linkedEvent.Name]) == 0 {
			logCtx.Error("linked event name : " + linkedEvent.Name + " not found")
			return 0, nil, errors.New("linked event name : " + linkedEvent.Name + " not found")
		}
	}
	sessionEventNameId := eventNameToId[U.EVENT_NAME_SESSION][0].(uint64)
	return sessionEventNameId, eventNameToId, nil
}

// Adds users who hit funnel event with given {event/user properties} to usersToBeAttributed
func getLinkedFunnelEventUsers(projectId uint64, from, to int64, linkedEvents []QueryEventWithProperties,
	eventNameToId map[string][]interface{},
	userIDInfo map[string]UserInfo, usersToBeAttributed *[]UserEventInfo) error {

	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})
	var usersHitConversion []string
	for key, _ := range userIDInfo {
		usersHitConversion = append(usersHitConversion, key)
	}

	for _, linkedEvent := range linkedEvents {
		// Part I - Fetch Users base on Event Hit satisfying events.properties
		linkedEventNameIDs := eventNameToId[linkedEvent.Name]
		eventsPlaceHolder := "?"
		for i := 0; i < len(linkedEventNameIDs)-1; i++ {
			eventsPlaceHolder += ",?"
		}
		var userIDList []string
		userIDHitGoalEventTimestamp := make(map[string]int64)
		userPropertiesIdsInBatches := U.GetStringListAsBatch(usersHitConversion, USER_BATCH_SIZE)
		for _, users := range userPropertiesIdsInBatches {

			// add user batching
			usersPlaceHolder := U.GetValuePlaceHolder(len(users))
			value := U.GetInterfaceList(users)
			queryEventHits := "SELECT user_id, timestamp FROM events WHERE events.project_id=? AND " +
				" timestamp >= ? AND timestamp <=? AND events.event_name_id IN (" + eventsPlaceHolder + ") " +
				" AND user_id = ANY (VALUES " + usersPlaceHolder + " ) "
			qParams := []interface{}{projectId, from, to}
			qParams = append(qParams, linkedEventNameIDs...)
			qParams = append(qParams, value...)
			// add event filter
			wStmt, wParams, err := getFilterSQLStmtForEventProperties(linkedEvent.Properties)
			if err != nil {
				return err
			}
			if wStmt != "" {
				queryEventHits = queryEventHits + " AND " + fmt.Sprintf("( %s )", wStmt)
				qParams = append(qParams, wParams...)
			}
			// fetch query results
			rows, err := db.Raw(queryEventHits, qParams...).Rows()
			if err != nil {
				logCtx.WithError(err).Error("SQL Query failed for queryEventHits")
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var userID string
				var timestamp int64
				if err = rows.Scan(&userID, &timestamp); err != nil {
					logCtx.WithError(err).Error("SQL Parse failed")
					continue
				}
				if _, ok := userIDHitGoalEventTimestamp[userID]; !ok {
					userIDList = append(userIDList, userID)
					userIDHitGoalEventTimestamp[userID] = timestamp
				}
			}
		}

		// Part-II - Filter the fetched users from Part-I base on user_properties
		filteredUserIdList, err := applyUserPropertiesFilter(projectId, userIDList, userIDInfo, linkedEvent.Properties)
		if err != nil {
			logCtx.WithError(err).Error("error while applying user properties")
			return err
		}

		// Part-III add the filtered users with eventId usersToBeAttributed
		for _, userId := range filteredUserIdList {
			*usersToBeAttributed = append(*usersToBeAttributed,
				UserEventInfo{userIDInfo[userId].CoalUserID, linkedEvent.Name})
		}
	}
	return nil
}

// Applies user properties filter on given set of users and returns only the filters ones those match
func applyUserPropertiesFilter(projectId uint64, userIdList []string, userIdInfo map[string]UserInfo,
	goalEventProperties []QueryProperty) ([]string, error) {

	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})

	wStmt, wParams, err := getFilterSQLStmtForUserProperties(goalEventProperties)
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
		userPropertiesIds = append(userPropertiesIds, userIdInfo[userId].PropertiesID)
	}

	var filteredUserIdList []string
	userIdHitGoalEventTimestamp := make(map[string]bool)
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
		if err != nil {
			logCtx.WithError(err).Error("SQL Query failed for getUserInitialSession")
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var userId string
			if err = rows.Scan(&userId); err != nil {
				logCtx.WithError(err).Error("SQL Parse failed")
				continue
			}
			if _, ok := userIdHitGoalEventTimestamp[userId]; !ok {
				filteredUserIdList = append(filteredUserIdList, userId)
				userIdHitGoalEventTimestamp[userId] = true
			}
		}
	}
	return filteredUserIdList, nil
}

// getConvertedUsers Returns the list of eligible users who hit conversion event
func getConvertedUsers(projectID uint64, goalEventName string, goalEventProperties []QueryProperty,
	from, to int64,
	eventNameToIdList map[string][]interface{}) (map[string]UserInfo, map[string][]UserIDPropID,
	map[string]int64, error) {

	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectId": projectID})

	conversionEventNameIds := eventNameToIdList[goalEventName]
	placeHolder := "?"
	for i := 0; i < len(conversionEventNameIds)-1; i++ {
		placeHolder += ",?"
	}
	queryEventHits := "SELECT user_id, timestamp FROM events WHERE events.project_id=? AND timestamp >= ? AND " +
		" timestamp <=? AND events.event_name_id IN (" + placeHolder + ") "
	qParams := []interface{}{projectID, from, to}
	qParams = append(qParams, conversionEventNameIds...)

	// add event filter
	wStmt, wParams, err := getFilterSQLStmtForEventProperties(goalEventProperties) // query.ConversionEvent.Properties)
	if err != nil {
		return nil, nil, nil, err
	}
	if wStmt != "" {
		queryEventHits = queryEventHits + " AND " + fmt.Sprintf("( %s )", wStmt)
		qParams = append(qParams, wParams...)
	}
	// fetch query results
	rows, err := db.Raw(queryEventHits, qParams...).Rows()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed for queryEventHits")
		return nil, nil, nil, err
	}
	defer rows.Close()
	var userIDList []string
	userIdHitGoalEventTimestamp := make(map[string]int64)
	for rows.Next() {
		var userId string
		var timestamp int64
		if err = rows.Scan(&userId, &timestamp); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		if _, ok := userIdHitGoalEventTimestamp[userId]; !ok {
			userIDList = append(userIDList, userId)
			userIdHitGoalEventTimestamp[userId] = timestamp
		}
	}

	// Get coalesced Id for converted user_ids (without filter)
	userIDToCoalIDInfo, err := getCoalesceIDFromUserIDs(userIDList, projectID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Part-II - Filter the fetched users from Part-I base on user_properties
	filteredUserIdList, err := applyUserPropertiesFilter(projectID, userIDList, userIDToCoalIDInfo,
		goalEventProperties)
	if err != nil {
		logCtx.WithError(err).Error("error while applying user properties")
		return nil, nil, nil, err
	}

	filteredUserIdToUserIDInfo := make(map[string]UserInfo)
	filteredCoalIDToUserIDInfo := make(map[string][]UserIDPropID)
	coalUserIdConversionTimestamp := make(map[string]int64)

	for _, userID := range filteredUserIdList {
		timestamp := userIdHitGoalEventTimestamp[userID]
		coalUserID := userIDToCoalIDInfo[userID].CoalUserID
		propertiesID := userIDToCoalIDInfo[userID].PropertiesID
		filteredCoalIDToUserIDInfo[coalUserID] =
			append(filteredCoalIDToUserIDInfo[coalUserID],
				UserIDPropID{userID, propertiesID, timestamp})
		filteredUserIdToUserIDInfo[userID] = UserInfo{coalUserID, propertiesID, timestamp}

		if _, ok := coalUserIdConversionTimestamp[coalUserID]; ok {
			if timestamp < coalUserIdConversionTimestamp[coalUserID] {
				// TODO (Anil) : discuss which T1, (T2, T3,) T4 -> () = Attribution lookback period.
				// Which T? should be considered?
				// considering earliest conversion
				coalUserIdConversionTimestamp[coalUserID] = timestamp
			}
		} else {
			coalUserIdConversionTimestamp[coalUserID] = timestamp
		}
	}

	return filteredUserIdToUserIDInfo, filteredCoalIDToUserIDInfo, coalUserIdConversionTimestamp, nil
}

// getEffectiveFrom Returns the effective From timestamp considering lookback days
func getEffectiveFrom(from int64, lookbackDays int) int64 {
	lookbackDaysTimestamp := int64(lookbackDays) * SECS_IN_A_DAY
	if LOOKBACK_CAP_IN_DAYS < lookbackDays {
		lookbackDaysTimestamp = int64(LOOKBACK_CAP_IN_DAYS) * SECS_IN_A_DAY
	}
	validFrom := from - lookbackDaysTimestamp
	return validFrom
}

// updateSessionsMapWithCoalesceID Clones a new map replacing userId by coalUserId.
func updateSessionsMapWithCoalesceID(attributedSessionsByUserId map[string]map[string]RangeTimestamp,
	usersInfo map[string]UserInfo) map[string]map[string]RangeTimestamp {

	newSessionsMap := make(map[string]map[string]RangeTimestamp)
	for userId, attributionIdMap := range attributedSessionsByUserId {
		userInfo := usersInfo[userId]
		for attributionId, newTimeRange := range attributionIdMap {
			if _, ok := newSessionsMap[userInfo.CoalUserID]; ok {
				if existingTimeRange, ok := newSessionsMap[userInfo.CoalUserID][attributionId]; ok {
					// update the existing attribution first and last touch
					existingTimeRange.MinTimestamp = U.Min(existingTimeRange.MinTimestamp, newTimeRange.MinTimestamp)
					existingTimeRange.MaxTimestamp = U.Max(existingTimeRange.MaxTimestamp, newTimeRange.MaxTimestamp)
					newSessionsMap[userInfo.CoalUserID][attributionId] = existingTimeRange
					continue
				}
				newSessionsMap[userInfo.CoalUserID][attributionId] = newTimeRange
				continue
			}
			newSessionsMap[userInfo.CoalUserID] = make(map[string]RangeTimestamp)
			newSessionsMap[userInfo.CoalUserID][attributionId] = newTimeRange
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

	customerAccountIds := strings.Split(customerAccountId, ",")
	performanceQuery := "SELECT value->>'campaign_id' AS campaign_id,  value->>'campaign_name' AS campaign_name, " +
		"SUM((value->>'impressions')::float) AS impressions, SUM((value->>'clicks')::float) AS clicks, " +
		"SUM((value->>'cost')::float)/1000000 AS total_cost FROM adwords_documents " +
		"where project_id = ? AND customer_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by value->>'campaign_id', campaign_name"
	rows, err := db.Raw(performanceQuery, projectId, customerAccountIds, ADWORDS_CAMPAIGN_REPORT_TYPE,
		U.GetDateOnlyFromTimestamp(from),
		U.GetDateOnlyFromTimestamp(to)).Rows()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var campaignName string
		var campaignId string
		var impressions float64
		var clicks float64
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
			attributionData[matchingId].Impressions = int(impressions)
			attributionData[matchingId].Clicks = int(clicks)
			attributionData[matchingId].Spend = spend
		}
	}

	currency, err := getAdwordsCurrency(projectId, customerAccountId, from, to)
	if err != nil {
		return "", err
	}
	return currency, nil
}

// Returns currency used for adwords customer_account_id
func getAdwordsCurrency(projectId uint64, customerAccountId string, from, to int64) (string, error) {

	customerAccountIds := strings.Split(customerAccountId, ",")
	if len(customerAccountIds) == 0 {
		return "", errors.New("no ad-words customer account id found")
	}
	queryCurrency := "SELECT value->>'currency_code' AS currency FROM adwords_documents " +
		" WHERE project_id=? AND customer_account_id=? AND type=? AND timestamp BETWEEN ? AND ? " +
		" ORDER BY timestamp DESC LIMIT 1"
	logCtx := log.WithField("ProjectId", projectId)
	db := C.GetServices().Db
	// checking just for customerAccountIds[0], we are assuming that all accounts have same currency
	rows, err := db.Raw(queryCurrency, projectId, customerAccountIds[0], 9, U.GetDateOnlyFromTimestamp(from),
		U.GetDateOnlyFromTimestamp(to)).Rows()
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
