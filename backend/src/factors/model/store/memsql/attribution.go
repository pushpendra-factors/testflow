package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Maps the {attribution key} to the session properties field
func getQuerySessionProperty(attributionKey string) (string, error) {
	if attributionKey == model.AttributionKeyCampaign {
		return U.EP_CAMPAIGN, nil
	} else if attributionKey == model.AttributionKeySource {
		return U.EP_SOURCE, nil
	} else if attributionKey == model.AttributionKeyAdgroup {
		return U.EP_ADGROUP, nil
	} else if attributionKey == model.AttributionKeyKeyword {
		return U.EP_KEYWORD, nil
	}
	return "", errors.New("invalid query properties")
}

// Adds common column names and linked events as header to the result rows.
func addHeadersByAttributionKey(result *model.QueryResult, query *model.AttributionQuery) {
	attributionKey := query.AttributionKey
	result.Headers = append(append(result.Headers, attributionKey), model.AttributionFixedHeaders...)
	conversionEventUsers := fmt.Sprintf("%s - Users", query.ConversionEvent.Name)
	costPerConversion := fmt.Sprintf("Cost Per Conversion")
	conversionEventCompareUsers := fmt.Sprintf("Compare - Users")
	compareCostPerConversion := fmt.Sprintf("Compare Cost Per Conversion")
	result.Headers = append(result.Headers, conversionEventUsers, costPerConversion,
		conversionEventCompareUsers, compareCostPerConversion)
	if len(query.LinkedEvents) > 0 {
		for _, event := range query.LinkedEvents {
			result.Headers = append(result.Headers, fmt.Sprintf("%s - Users", event.Name))
			result.Headers = append(result.Headers, fmt.Sprintf("%s - CPC", event.Name))
		}
	}
}

func isValidAttributionKeyValueAND(attributionKeyType string, keyValue string,
	filters []model.AttributionKeyFilter) bool {

	for _, filter := range filters {
		// supports AND and treats blank operator as AND
		if filter.LogicalOp == "OR" {
			continue
		}
		filterResult := applyOperator(attributionKeyType, keyValue, filter)
		// AND is false for any false.
		if !filterResult {
			return false
		}
	}
	return true
}

func isValidAttributionKeyValueOR(attributionKeyType string, keyValue string,
	filters []model.AttributionKeyFilter) bool {

	for _, filter := range filters {
		if filter.LogicalOp != "OR" {
			continue
		}
		filterResult := applyOperator(attributionKeyType, keyValue, filter)
		// OR is true for any true
		if filterResult {
			return true
		}
	}
	return false
}

func applyOperator(attributionKeyType string, keyValue string,
	filter model.AttributionKeyFilter) bool {

	filterResult := true
	// Currently only supporting matching key filters
	if filter.AttributionKey == attributionKeyType {
		switch filter.Operator {
		case model.EqualsOpStr:
			if keyValue != filter.Value {
				filterResult = false
			}
		case model.NotEqualOpStr:
			if keyValue == filter.Value {
				filterResult = false
			}
		case model.ContainsOpStr:
			if !strings.Contains(keyValue, filter.Value) {
				filterResult = false
			}
		case model.NotContainsOpStr:
			if strings.Contains(keyValue, filter.Value) {
				filterResult = false
			}
		default:
			log.Error("invalid filter operation found: " + filter.Operator)
			filterResult = false
		}
	}
	return filterResult
}

// ExecuteAttributionQuery Executes the Attribution using following steps:
//	1. Get all the sessions data (userId, attributionId, timestamp) for given period by attribution key
// 	2. Add the website visitor info using session data from step 1
//	3. i) 	Find out users who hit conversion event applying filter
//	  ii)	Using users from 3.i) find out users who hit linked funnel event applying filter
//	4. Apply attribution methodology
//	5. Add performance data by attributionId
func (store *MemSQL) ExecuteAttributionQuery(projectID uint64, queryOriginal *model.AttributionQuery) (*model.QueryResult, error) {

	var query *model.AttributionQuery
	U.DeepCopy(queryOriginal, &query)
	// for existing queries and backward support
	if query.QueryType == "" {
		query.QueryType = model.AttributionQueryTypeConversionBased
	}
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, errors.New("failed to get project Settings")
	}
	if projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "" {
		return nil, errors.New("execute attribution query failed as no ad-words customer account id found")
	}
	sessionEventNameID, eventNameToIDList, err := store.getEventInformation(projectID, query)
	if err != nil {
		return nil, err
	}

	// 1. Get all the sessions (userId, attributionId, timestamp) for given period by attribution key
	_sessions, sessionUsers, err := store.getAllTheSessions(projectID, sessionEventNameID, query,
		*projectSetting.IntAdwordsCustomerAccountId)
	if err != nil {
		return nil, err
	}
	usersInfo, err := store.GetCoalesceIDFromUserIDs(sessionUsers, projectID)
	if err != nil {
		return nil, err
	}
	sessions := updateSessionsMapWithCoalesceID(_sessions, usersInfo)

	isCompare := false
	// Default conversion for AttributionQueryTypeConversionBased.
	conversionFrom := query.From
	conversionTo := query.To
	// Extend the campaign window for engagement based attribution.
	if query.QueryType == model.AttributionQueryTypeEngagementBased {
		conversionFrom = query.From
		conversionTo = lookbackAdjustedTo(query.To, query.LookbackDays)
	}
	var attributionData map[string]*model.AttributionData
	if query.AttributionMethodologyCompare != "" {
		// Two AttributionMethodologies comparison
		isCompare = true
		attributionData, err = store.RunAttributionForMethodologyComparison(projectID,
			conversionFrom, conversionTo, query, eventNameToIDList, sessions)

	} else if query.ConversionEventCompare.Name != "" {
		// Two events comparison
		isCompare = true
		attributionData, err = store.runAttribution(projectID,
			conversionFrom, conversionTo, query.ConversionEvent, query, eventNameToIDList, sessions)

		if err != nil {
			return nil, err
		}
		// Running for ConversionEventCompare.
		attributionCompareData, err := store.runAttribution(projectID,
			conversionFrom, conversionTo, query.ConversionEventCompare, query, eventNameToIDList, sessions)

		if err != nil {
			return nil, err
		}

		// Merge compare data into attributionData.
		for key := range attributionData {
			if _, exists := attributionCompareData[key]; exists {
				attributionData[key].ConversionEventCompareCount = attributionCompareData[key].ConversionEventCount
			} else {
				attributionData[key].ConversionEventCompareCount = 0
			}
		}
		// Filling any non-matched touch points.
		for missingKey := range attributionCompareData {
			if _, exists := attributionData[missingKey]; !exists {
				attributionData[missingKey] = &model.AttributionData{}
				attributionData[missingKey].ConversionEventCompareCount = attributionCompareData[missingKey].ConversionEventCount
			}
		}
	} else {
		// Single event attribution.
		attributionData, err = store.runAttribution(projectID,
			conversionFrom, conversionTo, query.ConversionEvent,
			query, eventNameToIDList, sessions)
	}

	if err != nil {
		return nil, err
	}

	addWebsiteVisitorsInfo(attributionData, sessions, len(query.LinkedEvents))

	// 5. Add the performance information
	currency, err := store.AddAdwordsPerformanceReportInfo(projectID, attributionData, query.From, query.To,
		*projectSetting.IntAdwordsCustomerAccountId, query.AttributionKey, query.Timezone)
	if err != nil {
		return nil, err
	}

	if projectSetting.IntFacebookAdAccount != "" {
		err := store.AddFacebookPerformanceReportInfo(projectID, attributionData, query.From, query.To,
			projectSetting.IntFacebookAdAccount, query.AttributionKey, query.Timezone)
		if err != nil {
			return nil, err
		}
	}

	if projectSetting.IntLinkedinAdAccount != "" {
		err := store.AddLinkedinPerformanceReportInfo(projectID, attributionData, query.From, query.To,
			projectSetting.IntLinkedinAdAccount, query.AttributionKey, query.Timezone)
		if err != nil {
			return nil, err
		}
	}

	// Merging same name key's into single row
	dataRows := getRowsByMaps(attributionData, query.LinkedEvents, isCompare)
	mergedDataRows := model.MergeDataRowsHavingSameKey(dataRows)

	// sort the rows by conversionEvent
	conversionIndex := 5
	sort.Slice(mergedDataRows, func(i, j int) bool {
		return mergedDataRows[i][conversionIndex].(float64) > mergedDataRows[j][conversionIndex].(float64)
	})

	result := &model.QueryResult{}
	addHeadersByAttributionKey(result, query)
	result.Rows = mergedDataRows
	result.Meta.Currency = currency
	return result, nil
}

func (store *MemSQL) RunAttributionForMethodologyComparison(projectID uint64,
	conversionFrom, conversionTo int64, query *model.AttributionQuery, eventNameToIDList map[string][]interface{},
	sessions map[string]map[string]model.UserSessionTimestamp) (map[string]*model.AttributionData, error) {

	// Empty linkedEvents as they are not analyzed in compare events.
	var linkedEvents []model.QueryEventWithProperties

	var userIDToInfoConverted map[string]model.UserInfo
	var coalescedIDToInfoConverted map[string][]model.UserIDPropID
	var coalUserIdConversionTimestamp map[string]int64
	var err error
	// Fetch users who hit conversion event.
	if C.ShouldUseUserPropertiesTableForRead(projectID) {
		userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err = store.GetConvertedUsers(projectID,
			query.ConversionEvent.Name, query.ConversionEvent.Properties, conversionFrom, conversionTo,
			eventNameToIDList)

	} else {
		userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err = store.GetConvertedUsersWithFilter(projectID,
			query.ConversionEvent.Name, query.ConversionEvent.Properties, conversionFrom, conversionTo,
			eventNameToIDList)
	}
	if err != nil {
		return nil, err
	}

	// Add users who hit conversion event.
	var usersToBeAttributed []model.UserEventInfo
	for key := range coalescedIDToInfoConverted {
		usersToBeAttributed = append(usersToBeAttributed, model.UserEventInfo{CoalUserID: key,
			EventName: query.ConversionEvent.Name})
	}

	if C.ShouldUseUserPropertiesTableForRead(projectID) {
		err = store.GetLinkedFunnelEventUsers(projectID, conversionFrom, conversionTo,
			linkedEvents, eventNameToIDList, userIDToInfoConverted,
			&usersToBeAttributed)
	} else {
		err = store.GetLinkedFunnelEventUsersFilter(projectID, conversionFrom, conversionTo,
			linkedEvents, eventNameToIDList, userIDToInfoConverted,
			&usersToBeAttributed)
	}

	if err != nil {
		return nil, err
	}

	// Attribution based on given attribution methodology.
	userConversionHit, _, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodology,
		query.ConversionEvent.Name, usersToBeAttributed, sessions, coalUserIdConversionTimestamp,
		query.LookbackDays, query.From, query.To)
	if err != nil {
		return nil, err
	}

	attributionData := addUpConversionEventCount(userConversionHit)

	// Attribution based on given attributionMethodologyCompare methodology.
	userConversionCompareHit, _, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodologyCompare,
		query.ConversionEvent.Name, usersToBeAttributed, sessions, coalUserIdConversionTimestamp,
		query.LookbackDays, query.From, query.To)
	if err != nil {
		return nil, err
	}
	attributionDataCompare := addUpConversionEventCount(userConversionCompareHit)

	// Merge compare data into attributionData.
	for key := range attributionData {
		if _, exists := attributionDataCompare[key]; exists {
			attributionData[key].ConversionEventCompareCount = attributionDataCompare[key].ConversionEventCount
		} else {
			attributionData[key].ConversionEventCompareCount = 0
		}
	}
	// filling any non-matched touch points
	for missingKey := range attributionDataCompare {
		if _, exists := attributionData[missingKey]; !exists {
			attributionData[missingKey] = &model.AttributionData{}
			attributionData[missingKey].ConversionEventCompareCount = attributionDataCompare[missingKey].ConversionEventCount
		}
	}
	return attributionData, nil
}

func (store *MemSQL) runAttribution(projectID uint64,
	conversionFrom, conversionTo int64, goalEvent model.QueryEventWithProperties,
	query *model.AttributionQuery, eventNameToIDList map[string][]interface{},
	sessions map[string]map[string]model.UserSessionTimestamp) (map[string]*model.AttributionData, error) {

	goalEventName := goalEvent.Name
	goalEventProperties := goalEvent.Properties

	// 3. Fetch users who hit conversion event
	var userIDToInfoConverted map[string]model.UserInfo
	var coalescedIDToInfoConverted map[string][]model.UserIDPropID
	var coalUserIdConversionTimestamp map[string]int64
	var err error
	// Fetch users who hit conversion event.
	if C.ShouldUseUserPropertiesTableForRead(projectID) {
		userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err = store.GetConvertedUsers(projectID,
			goalEventName, goalEventProperties, conversionFrom, conversionTo, eventNameToIDList)

	} else {
		userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err = store.GetConvertedUsersWithFilter(projectID,
			goalEventName, goalEventProperties, conversionFrom, conversionTo, eventNameToIDList)
	}
	if err != nil {
		return nil, err
	}

	// Add users who hit conversion event
	var usersToBeAttributed []model.UserEventInfo
	for key := range coalescedIDToInfoConverted {
		usersToBeAttributed = append(usersToBeAttributed, model.UserEventInfo{CoalUserID: key,
			EventName: goalEventName})
	}

	if C.ShouldUseUserPropertiesTableForRead(projectID) {
		err = store.GetLinkedFunnelEventUsers(projectID, conversionFrom, conversionTo, query.LinkedEvents,
			eventNameToIDList, userIDToInfoConverted, &usersToBeAttributed)
	} else {
		err = store.GetLinkedFunnelEventUsersFilter(projectID, conversionFrom, conversionTo, query.LinkedEvents,
			eventNameToIDList, userIDToInfoConverted, &usersToBeAttributed)
	}

	if err != nil {
		return nil, err
	}

	// 4. Apply attribution based on given attribution methodology
	userConversionHit, userLinkedFEHit, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodology,
		goalEventName, usersToBeAttributed, sessions, coalUserIdConversionTimestamp,
		query.LookbackDays, query.From, query.To)
	if err != nil {
		return nil, err
	}

	attributionData := make(map[string]*model.AttributionData)
	attributionData = addUpConversionEventCount(userConversionHit)
	addUpLinkedFunnelEventCount(query.LinkedEvents, attributionData, userLinkedFEHit)
	return attributionData, nil
}

// getLinkedEventColumnAsInterfaceList return interface list having linked event count and CPC
func getLinkedEventColumnAsInterfaceList(spend float64, data []float64) []interface{} {
	var list []interface{}
	for _, val := range data {
		cpc := 0.0
		if val != 0.0 {
			cpc, _ = U.FloatRoundOffWithPrecision(spend/val, U.DefaultPrecision)
		}
		list = append(list, val, cpc)
	}
	return list
}

// Returns result in from of metrics. For empty attribution id, the values are accumulated into "$none".
func getRowsByMaps(attributionData map[string]*model.AttributionData,
	linkedEvents []model.QueryEventWithProperties, isCompare bool) [][]interface{} {

	rows := make([][]interface{}, 0)
	nonMatchingRow := []interface{}{"none", int64(0), int64(0), float64(0), int64(0), float64(0), float64(0),
		float64(0), float64(0)}
	for i := 0; i < len(linkedEvents); i++ {
		nonMatchingRow = append(nonMatchingRow, float64(0), float64(0))
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
				cpc, _ = U.FloatRoundOffWithPrecision(data.Spend/data.ConversionEventCount, U.DefaultPrecision)
			}
			if isCompare {
				cpcCompare := 0.0
				if data.ConversionEventCompareCount != 0.0 {
					cpcCompare, _ = U.FloatRoundOffWithPrecision(data.Spend/data.ConversionEventCompareCount, U.DefaultPrecision)
				}
				row = append(row, attributionIdName, data.Impressions, data.Clicks, data.Spend,
					data.WebsiteVisitors, data.ConversionEventCount, cpc,
					data.ConversionEventCompareCount, cpcCompare)
			} else {
				row = append(row, attributionIdName, data.Impressions, data.Clicks, data.Spend,
					data.WebsiteVisitors, data.ConversionEventCount, cpc, float64(0), float64(0))
			}
			row = append(row, getLinkedEventColumnAsInterfaceList(row[3].(float64), data.LinkedEventsCount)...)
			rows = append(rows, row)
		}
	}
	rows = append(rows, nonMatchingRow)
	return rows
}

// Groups all unique users by attributionId and adds it to attributionData
func addUpConversionEventCount(usersIdAttributionIdMap map[string][]string) map[string]*model.AttributionData {
	attributionData := make(map[string]*model.AttributionData)
	for _, attributionKeys := range usersIdAttributionIdMap {
		weight := 1 / float64(len(attributionKeys))
		for _, key := range attributionKeys {
			if _, exists := attributionData[key]; !exists {
				attributionData[key] = &model.AttributionData{}
			}
			attributionData[key].ConversionEventCount += weight
		}
	}
	return attributionData
}

// Attribute each user to the conversion event and linked event by attribution Id.
func addUpLinkedFunnelEventCount(linkedEvents []model.QueryEventWithProperties,
	attributionData map[string]*model.AttributionData, linkedUserAttributionData map[string]map[string][]string) {

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
	// Update linked up events with event hit count.
	for linkedEventName, userIdAttributionIdMap := range linkedUserAttributionData {
		for _, attributionKeys := range userIdAttributionIdMap {
			weight := 1 / float64(len(attributionKeys))
			for _, key := range attributionKeys {
				if attributionData[key] != nil {
					attributionData[key].LinkedEventsCount[linkedEventToPositionMap[linkedEventName]] += weight
				}
			}
		}
	}
}

// GetCoalesceIDFromUserIDs returns the map of coalesce userId for given list of users
func (store *MemSQL) GetCoalesceIDFromUserIDs(userIDs []string, projectID uint64) (map[string]model.UserInfo, error) {

	userIDsInBatches := U.GetStringListAsBatch(userIDs, model.UserBatchSize)
	logCtx := log.WithFields(log.Fields{"ProjectId": projectID})
	userIDToCoalUserIDMap := make(map[string]model.UserInfo)
	for _, users := range userIDsInBatches {
		placeHolder := U.GetValuePlaceHolder(len(users))
		value := U.GetInterfaceList(users)
		queryUserIDCoalID := "SELECT id, COALESCE(users.customer_user_id,users.id) AS coal_user_id, " +
			" properties_id FROM users WHERE id IN (" + placeHolder + ")"
		rows, err := store.ExecQueryWithContext(queryUserIDCoalID, value)
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
			userIDToCoalUserIDMap[userID] = model.UserInfo{CoalUserID: coalesceID, PropertiesID: propertiesID}
		}
	}
	return userIDToCoalUserIDMap, nil
}

// Returns the all the sessions (userId,attributionId,minTimestamp,maxTimestamp) for given
// users from given period including lookback
func (store *MemSQL) getAllTheSessions(projectId uint64, sessionEventNameId uint64,
	query *model.AttributionQuery, adwordsAccountId string) (map[string]map[string]model.UserSessionTimestamp, []string, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})
	effectiveFrom := lookbackAdjustedFrom(query.From, query.LookbackDays)
	effectiveTo := query.To
	// extend the campaign window for engagement based attribution
	if query.QueryType == model.AttributionQueryTypeEngagementBased {
		effectiveFrom = lookbackAdjustedFrom(query.From, query.LookbackDays)
		effectiveTo = lookbackAdjustedTo(query.To, query.LookbackDays)
	}
	gclIDBasedCampaign, err := store.GetGCLIDBasedCampaignInfo(projectId, effectiveFrom, effectiveTo, adwordsAccountId)
	if err != nil {
		return nil, nil, err
	}

	attributionEventKey, err := getQuerySessionProperty(query.AttributionKey)
	if err != nil {
		return nil, nil, err
	}

	attributedSessionsByUserId := make(map[string]map[string]model.UserSessionTimestamp)
	userIdMap := make(map[string]bool)
	var userIdsWithSession []string

	caseSelectStmt := "CASE WHEN JSON_EXTRACT_STRING(sessions.properties, ?) IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(sessions.properties, ?) = '' THEN ? ELSE JSON_EXTRACT_STRING(sessions.properties, ?) END"

	queryUserSessionTimeRange := "SELECT sessions.user_id, " + caseSelectStmt + " AS attribution_id, " + caseSelectStmt + " AS gcl_id, " +
		" sessions.timestamp FROM events AS sessions " +
		" WHERE sessions.project_id=? AND sessions.event_name_id=? AND sessions.timestamp BETWEEN ? AND ?"
	var qParams []interface{}
	qParams = append(qParams, attributionEventKey, model.PropertyValueNone, attributionEventKey, model.PropertyValueNone,
		attributionEventKey, U.EP_GCLID, model.PropertyValueNone, U.EP_GCLID, model.PropertyValueNone, U.EP_GCLID, projectId,
		sessionEventNameId, effectiveFrom, effectiveTo)
	rows, err := store.ExecQueryWithContext(queryUserSessionTimeRange, qParams)
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
		if !isValidAttributionKeyValueAND(query.AttributionKey,
			attributionId, query.AttributionKeyFilter) && !isValidAttributionKeyValueOR(query.AttributionKey,
			attributionId, query.AttributionKeyFilter) {
			continue
		}
		if _, ok := userIdMap[userID]; !ok {
			userIdsWithSession = append(userIdsWithSession, userID)
			userIdMap[userID] = true
		}

		// Override GCLID based campaign info if presents
		if gclID != model.PropertyValueNone {
			attributionIdBasedOnGclID := getGCLIDAttributionValue(gclIDBasedCampaign, gclID, attributionEventKey)
			// In cases where GCLID is present in events, but not in adwords report (as users tend to bookmark expired URLs),
			// fallback is attributionId
			if attributionIdBasedOnGclID != model.PropertyValueNone {
				attributionId = attributionIdBasedOnGclID
			}
		}

		// add session info uniquely for user-attributionId pair
		if _, ok := attributedSessionsByUserId[userID]; ok {

			if timeRange, ok := attributedSessionsByUserId[userID][attributionId]; ok {
				timeRange.MinTimestamp = U.Min(timeRange.MinTimestamp, timestamp)
				timeRange.MaxTimestamp = U.Max(timeRange.MaxTimestamp, timestamp)
				timeRange.TimeStamps = append(timeRange.TimeStamps, timestamp)
				timeRange.WithinQueryPeriod = timeRange.WithinQueryPeriod || timestamp >= query.From && timestamp <= query.To
				attributedSessionsByUserId[userID][attributionId] = timeRange
			} else {
				sessionRange := model.UserSessionTimestamp{MinTimestamp: timestamp,
					MaxTimestamp: timestamp, TimeStamps: []int64{timestamp},
					WithinQueryPeriod: timestamp >= query.From && timestamp <= query.To}
				attributedSessionsByUserId[userID][attributionId] = sessionRange
			}
		} else {
			attributedSessionsByUserId[userID] = make(map[string]model.UserSessionTimestamp)
			sessionRange := model.UserSessionTimestamp{MinTimestamp: timestamp,
				MaxTimestamp: timestamp, TimeStamps: []int64{timestamp},
				WithinQueryPeriod: timestamp >= query.From && timestamp <= query.To}
			attributedSessionsByUserId[userID][attributionId] = sessionRange
		}
	}
	return attributedSessionsByUserId, userIdsWithSession, nil
}

// Returns the matching value for GCLID, if not found returns $none
func getGCLIDAttributionValue(gclIDBasedCampaign map[string]model.CampaignInfo, gclID string, attributionKey string) string {

	if value, ok := gclIDBasedCampaign[gclID]; ok {
		switch attributionKey {
		case U.EP_ADGROUP:
			return value.AdgroupName
		case U.EP_CAMPAIGN:
			return value.CampaignName
		default:
			// No enrichment for Source and Keyword via GCLID
			return model.PropertyValueNone
		}
	}
	return model.PropertyValueNone
}

// Returns the concatenated list of conversion event + funnel events names
func buildEventNamesPlaceholder(query *model.AttributionQuery) []string {
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
func (store *MemSQL) getEventInformation(projectId uint64,
	query *model.AttributionQuery) (uint64, map[string][]interface{}, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})
	names := buildEventNamesPlaceholder(query)
	conversionAndFunnelEventMap := make(map[string]bool)
	for _, name := range names {
		conversionAndFunnelEventMap[name] = true
	}
	if _, exists := conversionAndFunnelEventMap[U.EVENT_NAME_SESSION]; !exists {
		names = append(names, U.EVENT_NAME_SESSION)
	}
	eventNames, errCode := store.GetEventNamesByNames(projectId, names)
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

// GetLinkedFunnelEventUsersFilter Adds users who hit funnel event with given {event/user properties} to usersToBeAttributed
func (store *MemSQL) GetLinkedFunnelEventUsersFilter(projectID uint64, queryFrom, queryTo int64,
	linkedEvents []model.QueryEventWithProperties, eventNameToId map[string][]interface{},
	userIDInfo map[string]model.UserInfo, usersToBeAttributed *[]model.UserEventInfo) error {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID})
	var usersHitConversion []string
	for key := range userIDInfo {
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
		userPropertiesIdsInBatches := U.GetStringListAsBatch(usersHitConversion, model.UserBatchSize)
		for _, users := range userPropertiesIdsInBatches {

			// add user batching
			usersPlaceHolder := U.GetValuePlaceHolder(len(users))
			value := U.GetInterfaceList(users)
			queryEventHits := "SELECT user_id, timestamp FROM events WHERE events.project_id=? AND " +
				" timestamp >= ? AND timestamp <=? AND events.event_name_id IN (" + eventsPlaceHolder + ") " +
				" AND user_id = ANY (VALUES " + usersPlaceHolder + " ) "
			qParams := []interface{}{projectID, queryFrom, queryTo}
			qParams = append(qParams, linkedEventNameIDs...)
			qParams = append(qParams, value...)

			// add event filter
			wStmtEvent, wParamsEvent, err := getFilterSQLStmtForEventProperties(projectID, linkedEvent.Properties) // query.ConversionEvent.Properties)
			if err != nil {
				return err
			}
			if wStmtEvent != "" {
				queryEventHits = queryEventHits + " AND " + fmt.Sprintf("( %s )", wStmtEvent)
				qParams = append(qParams, wParamsEvent...)
			}

			// add user filter
			wStmtUser, wParamsUser, err := getFilterSQLStmtForUserProperties(projectID, linkedEvent.Properties) // query.ConversionEvent.Properties)
			if err != nil {
				return err
			}
			if wStmtUser != "" {
				queryEventHits = queryEventHits + " AND " + fmt.Sprintf("( %s )", wStmtUser)
				qParams = append(qParams, wParamsUser...)
			}

			// fetch query results
			rows, err := store.ExecQueryWithContext(queryEventHits, qParams)
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

		// Part-III add the filtered users with eventId usersToBeAttributed
		for _, userId := range userIDList {
			*usersToBeAttributed = append(*usersToBeAttributed,
				model.UserEventInfo{CoalUserID: userIDInfo[userId].CoalUserID, EventName: linkedEvent.Name})
		}
	}
	return nil
}

// GetLinkedFunnelEventUsers Adds users who hit funnel event with given {event/user properties} to usersToBeAttributed
func (store *MemSQL) GetLinkedFunnelEventUsers(projectID uint64, queryFrom, queryTo int64,
	linkedEvents []model.QueryEventWithProperties, eventNameToId map[string][]interface{},
	userIDInfo map[string]model.UserInfo, usersToBeAttributed *[]model.UserEventInfo) error {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID})
	var usersHitConversion []string
	for key := range userIDInfo {
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
		userPropertiesIdsInBatches := U.GetStringListAsBatch(usersHitConversion, model.UserBatchSize)
		for _, users := range userPropertiesIdsInBatches {

			// add user batching
			usersPlaceHolder := U.GetValuePlaceHolder(len(users))
			value := U.GetInterfaceList(users)
			queryEventHits := "SELECT user_id, timestamp FROM events WHERE events.project_id=? AND " +
				" timestamp >= ? AND timestamp <=? AND events.event_name_id IN (" + eventsPlaceHolder + ") " +
				" AND user_id IN (" + usersPlaceHolder + ") "
			qParams := []interface{}{projectID, queryFrom, queryTo}
			qParams = append(qParams, linkedEventNameIDs...)
			qParams = append(qParams, value...)
			// add event filter
			wStmt, wParams, err := getFilterSQLStmtForEventProperties(projectID, linkedEvent.Properties)
			if err != nil {
				return err
			}
			if wStmt != "" {
				queryEventHits = queryEventHits + " AND " + fmt.Sprintf("( %s )", wStmt)
				qParams = append(qParams, wParams...)
			}
			// fetch query results
			rows, err := store.ExecQueryWithContext(queryEventHits, qParams)
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
		filteredUserIdList, err := store.ApplyUserPropertiesFilter(projectID, userIDList, userIDInfo, linkedEvent.Properties)
		if err != nil {
			logCtx.WithError(err).Error("error while applying user properties")
			return err
		}

		// Part-III add the filtered users with eventId usersToBeAttributed
		for _, userId := range filteredUserIdList {
			*usersToBeAttributed = append(*usersToBeAttributed,
				model.UserEventInfo{CoalUserID: userIDInfo[userId].CoalUserID, EventName: linkedEvent.Name})
		}
	}
	return nil
}

// ApplyUserPropertiesFilter Applies user properties filter on given set of users and returns only the filters ones those match
func (store *MemSQL) ApplyUserPropertiesFilter(projectID uint64, userIDList []string, userIDInfo map[string]model.UserInfo,
	goalEventProperties []model.QueryProperty) ([]string, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID})

	wStmt, wParams, err := getFilterSQLStmtForUserProperties(projectID, goalEventProperties)
	if err != nil {
		return nil, err
	}
	// return the same list if there's no user_properties filter
	if wStmt == "" {
		return userIDList, nil
	}

	var userPropertiesIds []string
	// Use properties Ids to speed up the search from user_properties table
	for _, userID := range userIDList {
		userPropertiesIds = append(userPropertiesIds, userIDInfo[userID].PropertiesID)
	}

	var filteredUserIdList []string
	userIdHitGoalEventTimestamp := make(map[string]bool)
	userPropertiesIdsInBatches := U.GetStringListAsBatch(userPropertiesIds, model.UserBatchSize)
	for _, users := range userPropertiesIdsInBatches {
		placeHolder := U.GetValuePlaceHolder(len(users))
		value := U.GetInterfaceList(users)

		userPropertiesRefTable := "user_properties"
		queryUserIdCoalID := fmt.Sprintf("SELECT user_id FROM %s", userPropertiesRefTable) + " " +
			"WHERE id = ANY (VALUES " + placeHolder + " ) "

		var qParams []interface{}
		qParams = append(qParams, value...)
		// add user_properties filter
		if wStmt != "" {
			queryUserIdCoalID = queryUserIdCoalID + " AND " + fmt.Sprintf("( %s )", wStmt)
			qParams = append(qParams, wParams...)
		}
		rows, err := store.ExecQueryWithContext(queryUserIdCoalID, qParams)
		if err != nil {
			logCtx.WithError(err).Error("SQL Query failed for getUserInitialSession")
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var userID string
			if err = rows.Scan(&userID); err != nil {
				logCtx.WithError(err).Error("SQL Parse failed")
				continue
			}
			if _, ok := userIdHitGoalEventTimestamp[userID]; !ok {
				filteredUserIdList = append(filteredUserIdList, userID)
				userIdHitGoalEventTimestamp[userID] = true
			}
		}
	}
	return filteredUserIdList, nil
}

// GetConvertedUsersWithFilter Returns the list of eligible users who hit conversion
// event for userProperties from events table
func (store *MemSQL) GetConvertedUsersWithFilter(projectID uint64, goalEventName string,
	goalEventProperties []model.QueryProperty, conversionFrom, conversionTo int64,
	eventNameToIdList map[string][]interface{}) (map[string]model.UserInfo,
	map[string][]model.UserIDPropID, map[string]int64, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID})

	conversionEventNameIDs := eventNameToIdList[goalEventName]
	placeHolder := "?"
	for i := 0; i < len(conversionEventNameIDs)-1; i++ {
		placeHolder += ",?"
	}
	queryEventHits := "SELECT user_id, timestamp FROM events WHERE events.project_id=? AND timestamp >= ? AND " +
		" timestamp <=? AND events.event_name_id IN (" + placeHolder + ") "
	qParams := []interface{}{projectID, conversionFrom, conversionTo}
	qParams = append(qParams, conversionEventNameIDs...)

	// add event filter
	wStmtEvent, wParamsEvent, err := getFilterSQLStmtForEventProperties(projectID, goalEventProperties) // query.ConversionEvent.Properties)
	if err != nil {
		return nil, nil, nil, err
	}
	if wStmtEvent != "" {
		queryEventHits = queryEventHits + " AND " + fmt.Sprintf("( %s )", wStmtEvent)
		qParams = append(qParams, wParamsEvent...)
	}

	// add user filter
	wStmtUser, wParamsUser, err := getFilterSQLStmtForUserProperties(projectID, goalEventProperties) // query.ConversionEvent.Properties)
	if err != nil {
		return nil, nil, nil, err
	}
	if wStmtUser != "" {
		queryEventHits = queryEventHits + " AND " + fmt.Sprintf("( %s )", wStmtUser)
		qParams = append(qParams, wParamsUser...)
	}

	// fetch query results
	rows, err := store.ExecQueryWithContext(queryEventHits, qParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed for queryEventHits")
		return nil, nil, nil, err
	}
	defer rows.Close()
	var userIDList []string
	userIdHitGoalEventTimestamp := make(map[string]int64)
	for rows.Next() {
		var userID string
		var timestamp int64
		if err = rows.Scan(&userID, &timestamp); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		if _, ok := userIdHitGoalEventTimestamp[userID]; !ok {
			userIDList = append(userIDList, userID)
			userIdHitGoalEventTimestamp[userID] = timestamp
		}
	}

	// Get coalesced Id for converted user_ids (without filter)
	userIDToCoalIDInfo, err := store.GetCoalesceIDFromUserIDs(userIDList, projectID)
	if err != nil {
		return nil, nil, nil, err
	}

	var filteredUserIdList []string
	for key := range userIDToCoalIDInfo {
		filteredUserIdList = append(filteredUserIdList, key)
	}

	filteredUserIdToUserIDInfo := make(map[string]model.UserInfo)
	filteredCoalIDToUserIDInfo := make(map[string][]model.UserIDPropID)
	coalUserIdConversionTimestamp := make(map[string]int64)

	for _, userID := range filteredUserIdList {
		timestamp := userIdHitGoalEventTimestamp[userID]
		coalUserID := userIDToCoalIDInfo[userID].CoalUserID
		propertiesID := userIDToCoalIDInfo[userID].PropertiesID
		filteredCoalIDToUserIDInfo[coalUserID] =
			append(filteredCoalIDToUserIDInfo[coalUserID],
				model.UserIDPropID{UserID: userID, PropertiesID: propertiesID, Timestamp: timestamp})
		filteredUserIdToUserIDInfo[userID] = model.UserInfo{CoalUserID: coalUserID,
			PropertiesID: propertiesID, Timestamp: timestamp}

		if _, ok := coalUserIdConversionTimestamp[coalUserID]; ok {
			if timestamp < coalUserIdConversionTimestamp[coalUserID] {
				// Considering earliest conversion.
				coalUserIdConversionTimestamp[coalUserID] = timestamp
			}
		} else {
			coalUserIdConversionTimestamp[coalUserID] = timestamp
		}
	}

	return filteredUserIdToUserIDInfo, filteredCoalIDToUserIDInfo, coalUserIdConversionTimestamp, nil
}

// GetConvertedUsers Returns the list of eligible users who hit conversion event
func (store *MemSQL) GetConvertedUsers(projectID uint64, goalEventName string,
	goalEventProperties []model.QueryProperty, conversionFrom, conversionTo int64,
	eventNameToIdList map[string][]interface{}) (map[string]model.UserInfo,
	map[string][]model.UserIDPropID, map[string]int64, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID})

	conversionEventNameIDs := eventNameToIdList[goalEventName]
	placeHolder := "?"
	for i := 0; i < len(conversionEventNameIDs)-1; i++ {
		placeHolder += ",?"
	}
	queryEventHits := "SELECT user_id, timestamp FROM events WHERE events.project_id=? AND timestamp >= ? AND " +
		" timestamp <=? AND events.event_name_id IN (" + placeHolder + ") "
	qParams := []interface{}{projectID, conversionFrom, conversionTo}
	qParams = append(qParams, conversionEventNameIDs...)

	// add event filter
	wStmt, wParams, err := getFilterSQLStmtForEventProperties(projectID, goalEventProperties) // query.ConversionEvent.Properties)
	if err != nil {
		return nil, nil, nil, err
	}
	if wStmt != "" {
		queryEventHits = queryEventHits + " AND " + fmt.Sprintf("( %s )", wStmt)
		qParams = append(qParams, wParams...)
	}
	// fetch query results
	rows, err := store.ExecQueryWithContext(queryEventHits, qParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed for queryEventHits")
		return nil, nil, nil, err
	}
	defer rows.Close()
	var userIDList []string
	userIdHitGoalEventTimestamp := make(map[string]int64)
	for rows.Next() {
		var userID string
		var timestamp int64
		if err = rows.Scan(&userID, &timestamp); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		if _, ok := userIdHitGoalEventTimestamp[userID]; !ok {
			userIDList = append(userIDList, userID)
			userIdHitGoalEventTimestamp[userID] = timestamp
		}
	}

	// Get coalesced Id for converted user_ids (without filter)
	userIDToCoalIDInfo, err := store.GetCoalesceIDFromUserIDs(userIDList, projectID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Part-II - Filter the fetched users from Part-I base on user_properties
	filteredUserIdList, err := store.ApplyUserPropertiesFilter(projectID, userIDList, userIDToCoalIDInfo,
		goalEventProperties)
	if err != nil {
		logCtx.WithError(err).Error("error while applying user properties")
		return nil, nil, nil, err
	}

	filteredUserIdToUserIDInfo := make(map[string]model.UserInfo)
	filteredCoalIDToUserIDInfo := make(map[string][]model.UserIDPropID)
	coalUserIdConversionTimestamp := make(map[string]int64)

	for _, userID := range filteredUserIdList {
		timestamp := userIdHitGoalEventTimestamp[userID]
		coalUserID := userIDToCoalIDInfo[userID].CoalUserID
		propertiesID := userIDToCoalIDInfo[userID].PropertiesID
		filteredCoalIDToUserIDInfo[coalUserID] =
			append(filteredCoalIDToUserIDInfo[coalUserID],
				model.UserIDPropID{UserID: userID, PropertiesID: propertiesID, Timestamp: timestamp})
		filteredUserIdToUserIDInfo[userID] = model.UserInfo{CoalUserID: coalUserID,
			PropertiesID: propertiesID, Timestamp: timestamp}

		if _, ok := coalUserIdConversionTimestamp[coalUserID]; ok {
			if timestamp < coalUserIdConversionTimestamp[coalUserID] {
				// Considering earliest conversion.
				coalUserIdConversionTimestamp[coalUserID] = timestamp
			}
		} else {
			coalUserIdConversionTimestamp[coalUserID] = timestamp
		}
	}

	return filteredUserIdToUserIDInfo, filteredCoalIDToUserIDInfo, coalUserIdConversionTimestamp, nil
}

// lookbackAdjustedFrom Returns the effective From timestamp considering lookback days
func lookbackAdjustedFrom(from int64, lookbackDays int) int64 {
	lookbackDaysTimestamp := int64(lookbackDays) * model.SecsInADay
	if model.LookbackCapInDays < lookbackDays {
		lookbackDaysTimestamp = int64(model.LookbackCapInDays) * model.SecsInADay
	}
	validFrom := from - lookbackDaysTimestamp
	return validFrom
}

// lookbackAdjustedTo Returns the effective To timestamp considering lookback days
func lookbackAdjustedTo(to int64, lookbackDays int) int64 {
	lookbackDaysTimestamp := int64(lookbackDays) * model.SecsInADay
	if model.LookbackCapInDays < lookbackDays {
		lookbackDaysTimestamp = int64(model.LookbackCapInDays) * model.SecsInADay
	}
	validTo := to + lookbackDaysTimestamp
	return validTo
}

// updateSessionsMapWithCoalesceID Clones a new map replacing userId by coalUserId.
func updateSessionsMapWithCoalesceID(attributedSessionsByUserID map[string]map[string]model.UserSessionTimestamp,
	usersInfo map[string]model.UserInfo) map[string]map[string]model.UserSessionTimestamp {

	newSessionsMap := make(map[string]map[string]model.UserSessionTimestamp)
	for userID, attributionIdMap := range attributedSessionsByUserID {
		userInfo := usersInfo[userID]
		for attributionID, newUserSession := range attributionIdMap {
			if _, ok := newSessionsMap[userInfo.CoalUserID]; ok {
				if existingUserSession, ok := newSessionsMap[userInfo.CoalUserID][attributionID]; ok {
					// Update the existing attribution first and last touch.
					existingUserSession.MinTimestamp = U.Min(existingUserSession.MinTimestamp, newUserSession.MinTimestamp)
					existingUserSession.MaxTimestamp = U.Max(existingUserSession.MaxTimestamp, newUserSession.MaxTimestamp)
					// Merging timestamp of same customer having 2 userIds.
					existingUserSession.TimeStamps = append(existingUserSession.TimeStamps, newUserSession.TimeStamps...)
					existingUserSession.WithinQueryPeriod = existingUserSession.WithinQueryPeriod || newUserSession.WithinQueryPeriod
					newSessionsMap[userInfo.CoalUserID][attributionID] = existingUserSession
					continue
				}
				newSessionsMap[userInfo.CoalUserID][attributionID] = newUserSession
				continue
			}
			newSessionsMap[userInfo.CoalUserID] = make(map[string]model.UserSessionTimestamp)
			newSessionsMap[userInfo.CoalUserID][attributionID] = newUserSession
		}
	}
	return newSessionsMap
}

// addWebsiteVisitorsInfo Maps the count distinct users session to campaign id and adds it to attributionData
func addWebsiteVisitorsInfo(attributionData map[string]*model.AttributionData,
	attributedSessionsByUserID map[string]map[string]model.UserSessionTimestamp, linkedEventsCount int) {
	// Creating an empty linked events row.
	emptyLinkedEventRow := make([]float64, 0)
	for i := 0; i < linkedEventsCount; i++ {
		emptyLinkedEventRow = append(emptyLinkedEventRow, float64(0))
	}

	userIDAttributionIDVisit := make(map[string]bool)
	for userID, attributionIDMap := range attributedSessionsByUserID {
		for attributionID, sessionTimestamp := range attributionIDMap {

			// Only count sessions that happened during attribution period.
			if sessionTimestamp.WithinQueryPeriod {

				if _, ok := attributionData[attributionID]; !ok {
					attributionData[attributionID] = &model.AttributionData{}
					if linkedEventsCount > 0 {
						// Init the linked events with 0.0 value.
						tempRow := emptyLinkedEventRow
						attributionData[attributionID].LinkedEventsCount = tempRow
					}
				}
				if _, ok := userIDAttributionIDVisit[getKey(userID, attributionID)]; ok {
					continue
				}
				attributionData[attributionID].WebsiteVisitors += 1
				userIDAttributionIDVisit[getKey(userID, attributionID)] = true
			}
		}
	}
}

// Merges 2 ids to create a string key
func getKey(id1 string, id2 string) string {
	return id1 + "|_|" + id2
}

// AddAdwordsPerformanceReportInfo Adds channel data to attributionData based on attribution id.
// Key id with no matching channel
// data is left with empty name parameter
//
// # ADGroup
// SELECT value->>'ad_group_id' AS ad_group_id,  value->>'ad_group_name' AS ad_group_name,
// SUM((value->>'impressions')::float) AS impressions, SUM((value->>'clicks')::float) AS clicks,
// SUM((value->>'cost')::float)/1000000 AS total_cost FROM adwords_documents where project_id = '399'
// AND customer_account_id IN ('1475899910') AND type = '10' AND timestamp between '20210220' AND '20210303'
// group by value->>'ad_group_id', ad_group_name LIMIT 5;
//
// # Campaign
// SELECT value->>'campaign_id' AS campaign_id,  value->>'campaign_name' AS campaign_name,
// SUM((value->>'impressions')::float) AS impressions, SUM((value->>'clicks')::float) AS clicks,
// SUM((value->>'cost')::float)/1000000 AS total_cost FROM adwords_documents where project_id = '399'
// AND customer_account_id IN ('1475899910') AND type = '5' AND timestamp between '20210220' AND '20210303'
// group by value->>'campaign_id', campaign_name LIMIT 5;
func (store *MemSQL) AddAdwordsPerformanceReportInfo(projectID uint64, attributionData map[string]*model.AttributionData,
	from, to int64, customerAccountID string, attributionKey string, timeZone string) (string, error) {
	logCtx := log.WithFields(log.Fields{"ProjectId": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})

	customerAccountIDs := strings.Split(customerAccountID, ",")

	reportType := model.AdwordsDocumentTypeAlias[model.CampaignPerformanceReport] // 5
	performanceQuery := "SELECT value::campaign_id AS campaign_id,  value::campaign_name AS campaign_name, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'cost'))/1000000 AS total_cost FROM adwords_documents " +
		"where project_id = ? AND customer_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by value::campaign_id, campaign_name"

	// AdGroup report for AttributionKey as AdGroup
	if attributionKey == model.AttributionKeyAdgroup {
		reportType = model.AdwordsDocumentTypeAlias[model.AdGroupPerformanceReport] // 10
		performanceQuery = "SELECT value::ad_group_id AS ad_group_id,  value::ad_group_name AS ad_group_name, " +
			"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
			"SUM(JSON_EXTRACT_STRING(value, 'cost'))/1000000 AS total_cost FROM adwords_documents " +
			"where project_id = ? AND customer_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
			"group by value::ad_group_id, ad_group_name"
	}

	rows, err := store.ExecQueryWithContext(performanceQuery, []interface{}{projectID, customerAccountIDs, reportType,
		U.GetDateAsStringZ(from, U.TimeZoneString(timeZone)),
		U.GetDateAsStringZ(to, U.TimeZoneString(timeZone))})
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var keyName string
		var keyID string
		var impressions float64
		var clicks float64
		var spend float64
		if err = rows.Scan(&keyID, &keyName, &impressions, &clicks, &spend); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		matchingID := ""
		if _, keyIDFound := attributionData[keyID]; keyIDFound {
			matchingID = keyID
		} else if _, keyNameFound := attributionData[keyName]; keyNameFound {
			matchingID = keyName
		}
		if matchingID != "" {
			attributionData[matchingID].Name = keyName
			attributionData[matchingID].Impressions += int64(impressions)
			attributionData[matchingID].Clicks += int64(clicks)
			attributionData[matchingID].Spend += spend
		}
	}

	currency, err := store.GetAdwordsCurrency(projectID, customerAccountID, from, to)
	if err != nil {
		return "", err
	}
	return currency, nil
}

// GetAdwordsCurrency Returns currency used for adwords customer_account_id
func (store *MemSQL) GetAdwordsCurrency(projectID uint64, customerAccountID string, from, to int64) (string, error) {

	customerAccountIDs := strings.Split(customerAccountID, ",")
	if len(customerAccountIDs) == 0 {
		return "", errors.New("no ad-words customer account id found")
	}
	queryCurrency := "SELECT value::currency_code AS currency FROM adwords_documents " +
		" WHERE project_id=? AND customer_account_id=? AND type=? AND timestamp BETWEEN ? AND ? " +
		" ORDER BY timestamp DESC LIMIT 1"
	logCtx := log.WithField("ProjectId", projectID)
	// Checking just for customerAccountIDs[0], we are assuming that all accounts have same currency.
	rows, err := store.ExecQueryWithContext(queryCurrency, []interface{}{projectID, customerAccountIDs[0], 9, U.GetDateOnlyFromTimestamp(from),
		U.GetDateOnlyFromTimestamp(to)})
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

// AddFacebookPerformanceReportInfo Adds Facebook channel data to attributionData based on attribution id.
// Key id with no matching channel
// data is left with empty name parameter
// # ADGroup
// SELECT value->>'adset_id' AS adset_id,  value->>'adset_name' AS adset_name, SUM((value->>'impressions')::float)
// AS impressions, SUM((value->>'clicks')::float) AS clicks, SUM((value->>'spend')::float)/1000000 AS total_spend
// FROM facebook_documents where project_id = '399' AND customer_ad_account_id IN ('act_367960820625667')
// AND type = '6' AND timestamp between '20210220' AND '20210303' group by value->>'adset_id', adset_name LIMIT 5;
//
// # Campaign
// SELECT value->>'campaign_id' AS campaign_id,  value->>'campaign_name' AS campaign_name,
// SUM((value->>'impressions')::float) AS impressions, SUM((value->>'clicks')::float) AS clicks,
// SUM((value->>'spend')::float)/1000000 AS total_spend FROM facebook_documents where project_id = '399'
// AND customer_ad_account_id IN ('act_367960820625667') AND type = '5' AND timestamp between '20210220'
// AND '20210303' group by value->>'campaign_id', campaign_name LIMIT 5;
func (store *MemSQL) AddFacebookPerformanceReportInfo(projectID uint64, attributionData map[string]*model.AttributionData,
	from, to int64, customerAccountID string, attributionKey string, timeZone string) error {
	logCtx := log.WithFields(log.Fields{"ProjectId": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})

	customerAccountIDs := strings.Split(customerAccountID, ",")

	reportType := facebookDocumentTypeAlias["campaign_insights"] // 5
	performanceQuery := "SELECT JSON_EXTRACT_STRING(value, 'campaign_id') AS campaign_id, JSON_EXTRACT_STRING(value, 'campaign_name') AS campaign_name, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'spend')) AS total_spend FROM facebook_documents " +
		"where project_id = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by JSON_EXTRACT_STRING(value, 'campaign_id'), campaign_name"

	// AdGroup report for AttributionKey as AdGroup
	if attributionKey == model.AttributionKeyAdgroup {
		reportType = facebookDocumentTypeAlias["ad_set_insights"] // 5
		performanceQuery = "SELECT JSON_EXTRACT_STRING(value, 'adset_id') AS adset_id, JSON_EXTRACT_STRING(value, 'adset_name') AS adset_name, " +
			"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
			"SUM(JSON_EXTRACT_STRING(value, 'spend')) AS total_spend FROM facebook_documents " +
			"where project_id = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
			"group by JSON_EXTRACT_STRING(value, 'adset_id'), adset_name"
	}

	rows, err := store.ExecQueryWithContext(performanceQuery, []interface{}{projectID, customerAccountIDs, reportType,
		U.GetDateAsStringZ(from, U.TimeZoneString(timeZone)),
		U.GetDateAsStringZ(to, U.TimeZoneString(timeZone))})
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var keyName string
		var keyID string
		var impressions float64
		var clicks float64
		var spend float64
		if err = rows.Scan(&keyID, &keyName, &impressions, &clicks, &spend); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		matchingID := ""
		if _, keyIdFound := attributionData[keyID]; keyIdFound {
			matchingID = keyID
		} else if _, keyNameFound := attributionData[keyName]; keyNameFound {
			matchingID = keyName
		}
		if matchingID != "" {
			// TODO (Anil) How do we resolve the conflict in same name ads across G/FB/Linkedin
			attributionData[matchingID].Name = keyName
			attributionData[matchingID].Impressions += int64(impressions)
			attributionData[matchingID].Clicks += int64(clicks)
			// TODO (Anil) Add currency or use conversion factor to set in default currency for G/FB/Linkedin
			attributionData[matchingID].Spend += spend
		}
	}
	return nil
}

// AddLinkedinPerformanceReportInfo Adds Linkedin channel data to attributionData based on attribution id.
// Key id with no matching channel
// data is left with empty name parameter
// # ADGroup
// SELECT value->>'campaign_id' AS campaign_id,  value->>'campaign_name' AS campaign_name,
// SUM((value->>'impressions')::float) AS impressions, SUM((value->>'clicks')::float) AS clicks,
// SUM((value->>'costInLocalCurrency')::float)/1000000 AS total_spend FROM linkedin_documents where
// project_id = '399' AND customer_ad_account_id IN ('506157045') AND type = '6' AND timestamp
// between '20210220' AND '20210303' group by value->>'campaign_id', campaign_name LIMIT 5;
//
// # Campaign
// SELECT value->>'campaign_group_id' AS campaign_group_id,  value->>'campaign_group_name' AS campaign_group_name,
// SUM((value->>'impressions')::float) AS impressions, SUM((value->>'clicks')::float) AS clicks,
// SUM((value->>'costInLocalCurrency')::float)/1000000 AS total_spend FROM linkedin_documents
// where project_id = '399' AND customer_ad_account_id IN ('506157045') AND type = '5' AND
// timestamp between '20210220' AND '20210303' group by value->>'campaign_group_id', campaign_group_name LIMIT 5;
func (store *MemSQL) AddLinkedinPerformanceReportInfo(projectID uint64, attributionData map[string]*model.AttributionData,
	from, to int64, customerAccountID string, attributionKey string, timeZone string) error {
	logCtx := log.WithFields(log.Fields{"ProjectId": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})

	customerAccountIDs := strings.Split(customerAccountID, ",")

	reportType := linkedinDocumentTypeAlias["campaign_group_insights"] // 5
	performanceQuery := "SELECT JSON_EXTRACT_STRING(value, 'campaign_group_id') AS campaign_group_id, JSON_EXTRACT_STRING(value, 'campaign_group_name') AS campaign_group_name, " +
		"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
		"SUM(JSON_EXTRACT_STRING(value, 'costInLocalCurrency')) AS total_spend FROM linkedin_documents " +
		"where project_id = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
		"group by JSON_EXTRACT_STRING(value, 'campaign_group_id'), campaign_group_name"

	// AdGroup report for AttributionKey as AdGroup
	if attributionKey == model.AttributionKeyAdgroup {
		reportType = linkedinDocumentTypeAlias["campaign_insights"] // 6
		performanceQuery = "SELECT JSON_EXTRACT_STRING(value, 'campaign_id') AS campaign_id,  JSON_EXTRACT_STRING(value, 'campaign_name') AS campaign_name, " +
			"SUM(JSON_EXTRACT_STRING(value, 'impressions')) AS impressions, SUM(JSON_EXTRACT_STRING(value, 'clicks')) AS clicks, " +
			"SUM(JSON_EXTRACT_STRING(value, 'costInLocalCurrency')) AS total_spend FROM linkedin_documents " +
			"where project_id = ? AND customer_ad_account_id IN (?) AND type = ? AND timestamp between ? AND ? " +
			"group by JSON_EXTRACT_STRING(value, 'campaign_id'), campaign_name"
	}

	rows, err := store.ExecQueryWithContext(performanceQuery, []interface{}{projectID, customerAccountIDs, reportType,
		U.GetDateAsStringZ(from, U.TimeZoneString(timeZone)),
		U.GetDateAsStringZ(to, U.TimeZoneString(timeZone))})
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var keyName string
		var keyID string
		var impressions float64
		var clicks float64
		var spend float64
		if err = rows.Scan(&keyID, &keyName, &impressions, &clicks, &spend); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		matchingID := ""
		if _, keyIdFound := attributionData[keyID]; keyIdFound {
			matchingID = keyID
		} else if _, keyNameFound := attributionData[keyName]; keyNameFound {
			matchingID = keyName
		}
		if matchingID != "" {
			attributionData[matchingID].Name = keyName
			attributionData[matchingID].Impressions += int64(impressions)
			attributionData[matchingID].Clicks += int64(clicks)
			// TODO (Anil) Add currency or use conversion factor to set in default currency for G/FB/Linkedin
			attributionData[matchingID].Spend += spend
		}
	}
	return nil
}
