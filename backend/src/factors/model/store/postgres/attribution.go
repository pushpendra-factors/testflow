package postgres

import (
	"errors"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ExecuteAttributionQuery Executes the Attribution using following steps:
//	1. Get all the sessions data (userId, attributionId, timestamp) for given period by attribution key
// 	2. Add the website visitor info using session data from step 1
//	3. i) 	Find out users who hit conversion event applying filter
//	  ii)	Using users from 3.i) find out users who hit linked funnel event applying filter
//	4. Apply attribution methodology
//	5. Add performance data by attributionId
func (pg *Postgres) ExecuteAttributionQuery(projectID uint64, queryOriginal *model.AttributionQuery) (*model.QueryResult, error) {

	var query *model.AttributionQuery
	U.DeepCopy(queryOriginal, &query)
	// for existing queries and backward support
	if query.QueryType == "" {
		query.QueryType = model.AttributionQueryTypeConversionBased
	}
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, errors.New("failed to get project Settings")
	}
	if projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "" {
		return &model.QueryResult{}, errors.New(model.AttributionErrorIntegrationNotFound)
	}

	marketingReports, err := pg.FetchMarketingReports(projectID, *query, *projectSetting)
	if err != nil {
		return nil, err
	}

	sessionEventNameID, eventNameToIDList, err := pg.getEventInformation(projectID, query)
	if err != nil {
		return nil, err
	}

	// 1. Get all the sessions (userId, attributionId, timestamp) for given period by attribution key
	_sessions, sessionUsers, err := pg.getAllTheSessions(projectID, sessionEventNameID, query, marketingReports)
	if err != nil {
		return nil, err
	}

	usersInfo, err := pg.GetCoalesceIDFromUserIDs(sessionUsers, projectID)
	if err != nil {
		return nil, err
	}
	// coalUserId[Key][UserSessionData]
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
		attributionData, err = pg.RunAttributionForMethodologyComparison(projectID,
			conversionFrom, conversionTo, query, eventNameToIDList, sessions)

	} else if query.ConversionEventCompare.Name != "" {
		// Two events comparison
		isCompare = true
		attributionData, err = pg.runAttribution(projectID,
			conversionFrom, conversionTo, query.ConversionEvent, query, eventNameToIDList, sessions)

		if err != nil {
			return nil, err
		}
		// Running for ConversionEventCompare.
		attributionCompareData, err := pg.runAttribution(projectID,
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
		attributionData, err = pg.runAttribution(projectID,
			conversionFrom, conversionTo, query.ConversionEvent,
			query, eventNameToIDList, sessions)
	}

	if err != nil {
		return nil, err
	}

	addWebsiteVisitorsInfo(attributionData, sessions, len(query.LinkedEvents))

	// Add the Added keys
	model.AddTheAddedKeysAndMetrics(attributionData, query.AttributionKey, sessions)

	// Add the performance information
	model.AddPerformanceData(attributionData, query.AttributionKey, marketingReports)

	// Add additional metrics values
	model.ComputeAdditionalMetrics(attributionData)

	// Merging same name key's into single row
	dataRows := model.GetRowsByMaps(query.AttributionKey, attributionData, query.LinkedEvents, isCompare)
	mergedDataRows := model.MergeDataRowsHavingSameKey(dataRows, model.GetKeyIndexOrAddedKeySize(query.AttributionKey))

	result := &model.QueryResult{}
	model.AddHeadersByAttributionKey(result, query)

	// sort the rows by conversionEvent
	conversionIndex := model.GetConversionIndex(result.Headers)
	sort.Slice(mergedDataRows, func(i, j int) bool {
		return mergedDataRows[i][conversionIndex].(float64) > mergedDataRows[j][conversionIndex].(float64)
	})

	result.Rows = mergedDataRows
	currency, err := pg.GetAdwordsCurrency(projectID, *projectSetting.IntAdwordsCustomerAccountId, query.From, query.To)
	if err != nil {
		return result, err
	}
	result.Meta.Currency = currency
	return result, nil
}

func (pg *Postgres) RunAttributionForMethodologyComparison(projectID uint64,
	conversionFrom, conversionTo int64, query *model.AttributionQuery, eventNameToIDList map[string][]interface{},
	sessions map[string]map[string]model.UserSessionData) (map[string]*model.AttributionData, error) {

	// Empty linkedEvents as they are not analyzed in compare events.
	var linkedEvents []model.QueryEventWithProperties

	var userIDToInfoConverted map[string]model.UserInfo
	var coalescedIDToInfoConverted map[string][]model.UserIDPropID
	var coalUserIdConversionTimestamp map[string]int64
	var err error
	// Fetch users who hit conversion event.
	userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err = pg.GetConvertedUsersWithFilter(projectID,
		query.ConversionEvent.Name, query.ConversionEvent.Properties, conversionFrom, conversionTo,
		eventNameToIDList)
	if err != nil {
		return nil, err
	}

	// Add users who hit conversion event.
	var usersToBeAttributed []model.UserEventInfo
	for key := range coalescedIDToInfoConverted {
		usersToBeAttributed = append(usersToBeAttributed, model.UserEventInfo{CoalUserID: key,
			EventName: query.ConversionEvent.Name})
	}

	err = pg.GetLinkedFunnelEventUsersFilter(projectID, conversionFrom, conversionTo,
		linkedEvents, eventNameToIDList, userIDToInfoConverted,
		&usersToBeAttributed)

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

	attributionData := model.AddUpConversionEventCount(userConversionHit)

	// Attribution based on given attributionMethodologyCompare methodology.
	userConversionCompareHit, _, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodologyCompare,
		query.ConversionEvent.Name, usersToBeAttributed, sessions, coalUserIdConversionTimestamp,
		query.LookbackDays, query.From, query.To)
	if err != nil {
		return nil, err
	}
	attributionDataCompare := model.AddUpConversionEventCount(userConversionCompareHit)

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

func (pg *Postgres) runAttribution(projectID uint64,
	conversionFrom, conversionTo int64, goalEvent model.QueryEventWithProperties,
	query *model.AttributionQuery, eventNameToIDList map[string][]interface{},
	sessions map[string]map[string]model.UserSessionData) (map[string]*model.AttributionData, error) {

	goalEventName := goalEvent.Name
	goalEventProperties := goalEvent.Properties

	// 3. Fetch users who hit conversion event
	var userIDToInfoConverted map[string]model.UserInfo
	var coalescedIDToInfoConverted map[string][]model.UserIDPropID
	var coalUserIdConversionTimestamp map[string]int64
	var err error
	// Fetch users who hit conversion event.
	userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err = pg.GetConvertedUsersWithFilter(projectID,
		goalEventName, goalEventProperties, conversionFrom, conversionTo, eventNameToIDList)
	if err != nil {
		return nil, err
	}

	// Add users who hit conversion event
	var usersToBeAttributed []model.UserEventInfo
	for key := range coalescedIDToInfoConverted {
		usersToBeAttributed = append(usersToBeAttributed, model.UserEventInfo{CoalUserID: key,
			EventName: goalEventName})
	}

	err = pg.GetLinkedFunnelEventUsersFilter(projectID, conversionFrom, conversionTo, query.LinkedEvents,
		eventNameToIDList, userIDToInfoConverted, &usersToBeAttributed)

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
	attributionData = model.AddUpConversionEventCount(userConversionHit)
	model.AddUpLinkedFunnelEventCount(query.LinkedEvents, attributionData, userLinkedFEHit)
	return attributionData, nil
}

// GetCoalesceIDFromUserIDs returns the map of coalesce userId for given list of users
func (pg *Postgres) GetCoalesceIDFromUserIDs(userIDs []string, projectID uint64) (map[string]model.UserInfo, error) {

	userIDsInBatches := U.GetStringListAsBatch(userIDs, model.UserBatchSize)
	logCtx := log.WithFields(log.Fields{"ProjectId": projectID})
	userIDToCoalUserIDMap := make(map[string]model.UserInfo)
	for _, users := range userIDsInBatches {
		placeHolder := U.GetValuePlaceHolder(len(users))
		value := U.GetInterfaceList(users)
		queryUserIDCoalID := "SELECT id, COALESCE(users.customer_user_id,users.id) AS coal_user_id, " +
			" properties_id FROM users WHERE id = ANY (VALUES " + placeHolder + " )"
		rows, err := pg.ExecQueryWithContext(queryUserIDCoalID, value)
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
				logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
				continue
			}
			userIDToCoalUserIDMap[userID] = model.UserInfo{CoalUserID: coalesceID}
		}
	}
	return userIDToCoalUserIDMap, nil
}

// getAllTheSessions Returns the all the sessions (userId,attributionId,minTimestamp,maxTimestamp) for given
// users from given period including lookback
func (pg *Postgres) getAllTheSessions(projectId uint64, sessionEventNameId string, query *model.AttributionQuery,
	reports *model.MarketingReports) (map[string]map[string]model.UserSessionData, []string, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})
	effectiveFrom := lookbackAdjustedFrom(query.From, query.LookbackDays)
	effectiveTo := query.To
	// extend the campaign window for engagement based attribution
	if query.QueryType == model.AttributionQueryTypeEngagementBased {
		effectiveFrom = lookbackAdjustedFrom(query.From, query.LookbackDays)
		effectiveTo = lookbackAdjustedTo(query.To, query.LookbackDays)
	}

	attributionEventKey, err := model.GetQuerySessionProperty(query.AttributionKey)
	if err != nil {
		return nil, nil, err
	}

	attributedSessionsByUserId := make(map[string]map[string]model.UserSessionData)
	userIdMap := make(map[string]bool)
	var userIdsWithSession []string

	caseSelectStmt := "CASE WHEN sessions.properties->>? IS NULL THEN ? " +
		" WHEN sessions.properties->>? = '' THEN ? ELSE sessions.properties->>? END"

	queryUserSessionTimeRange := "SELECT sessions.user_id, " +
		caseSelectStmt + " AS sessionTimeSpent, " +
		caseSelectStmt + " AS pageCount, " +
		caseSelectStmt + " AS campaignName, " +
		caseSelectStmt + " AS adgroupName, " +
		caseSelectStmt + " AS keywordName, " +
		caseSelectStmt + " AS keywordMatchType, " +
		caseSelectStmt + " AS source, " +
		caseSelectStmt + " AS attribution_id, " +
		caseSelectStmt + " AS gcl_id, " +
		" sessions.timestamp FROM events AS sessions " +
		" WHERE sessions.project_id=? AND sessions.event_name_id=? AND sessions.timestamp BETWEEN ? AND ?"
	var qParams []interface{}

	qParams = append(qParams,
		U.SP_SPENT_TIME, 0, U.SP_SPENT_TIME, 0, U.SP_SPENT_TIME,
		U.SP_PAGE_COUNT, 0, U.SP_PAGE_COUNT, 0, U.SP_PAGE_COUNT,
		U.EP_CAMPAIGN, model.PropertyValueNone, U.EP_CAMPAIGN, model.PropertyValueNone, U.EP_CAMPAIGN,
		U.EP_ADGROUP, model.PropertyValueNone, U.EP_ADGROUP, model.PropertyValueNone, U.EP_ADGROUP,
		U.EP_KEYWORD, model.PropertyValueNone, U.EP_KEYWORD, model.PropertyValueNone, U.EP_KEYWORD,
		U.EP_KEYWORD_MATCH_TYPE, model.PropertyValueNone, U.EP_KEYWORD_MATCH_TYPE, model.PropertyValueNone, U.EP_KEYWORD_MATCH_TYPE,
		U.EP_SOURCE, model.PropertyValueNone, U.EP_SOURCE, model.PropertyValueNone, U.EP_SOURCE,
		attributionEventKey, model.PropertyValueNone, attributionEventKey, model.PropertyValueNone, attributionEventKey,
		U.EP_GCLID, model.PropertyValueNone, U.EP_GCLID, model.PropertyValueNone, U.EP_GCLID,
		projectId, sessionEventNameId, effectiveFrom, effectiveTo)
	rows, err := pg.ExecQueryWithContext(queryUserSessionTimeRange, qParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return attributedSessionsByUserId, userIdsWithSession, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		var sessionSpentTime float64
		var pageCount int64
		var campaignName string
		var adgroupName string
		var keywordName string
		var keywordMatchType string
		var sourceName string
		var attributionId string
		var gclID string
		var timestamp int64
		if err = rows.Scan(&userID, &sessionSpentTime, &pageCount, &campaignName, &adgroupName, &keywordName, &keywordMatchType, &sourceName, &attributionId, &gclID, &timestamp); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
			continue
		}
		// apply filter at extracting session level itself
		if !model.IsValidAttributionKeyValueAND(query.AttributionKey,
			attributionId, query.AttributionKeyFilter) && !model.IsValidAttributionKeyValueOR(query.AttributionKey,
			attributionId, query.AttributionKeyFilter) {
			continue
		}
		if _, ok := userIdMap[userID]; !ok {
			userIdsWithSession = append(userIdsWithSession, userID)
			userIdMap[userID] = true
		}
		marketingValues := model.MarketingData{CampaignName: campaignName, AdgroupName: adgroupName, KeywordName: keywordName, KeywordMatchType: keywordMatchType, Source: sourceName}
		var attributionIdBasedOnGclID string
		// Override GCLID based campaign info if presents
		if gclID != model.PropertyValueNone && !(query.AttributionKey == model.AttributionKeyKeyword && !model.IsASearchSlotKeyword(&(*reports).AdwordsGCLIDData, gclID)) {
			attributionIdBasedOnGclID, marketingValues = model.GetGCLIDAttributionValue(&(*reports).AdwordsGCLIDData, gclID, attributionEventKey, marketingValues)
			// In cases where GCLID is present in events, but not in adwords report (as users tend to bookmark expired URLs),
			// fallback is attributionId
			if attributionIdBasedOnGclID != model.PropertyValueNone && attributionIdBasedOnGclID != "" {
				attributionId = attributionIdBasedOnGclID
			}
		} else {
			// we can't enrich the values in any other way, why?
			// Because, an adgroup can belong to multiple campaign, keyword also has match type, adgroup, campaign as keys
		}
		// Name
		marketingValues.Name = attributionId
		// Add the unique attributionKey key
		marketingValues.Key = model.GetMarketingDataKey(query.AttributionKey, marketingValues)
		uniqueAttributionKey := marketingValues.Key
		// add session info uniquely for user-attributionId pair
		if _, ok := attributedSessionsByUserId[userID]; ok {

			if userSessionData, ok := attributedSessionsByUserId[userID][uniqueAttributionKey]; ok {
				userSessionData.MinTimestamp = U.Min(userSessionData.MinTimestamp, timestamp)
				userSessionData.MaxTimestamp = U.Max(userSessionData.MaxTimestamp, timestamp)
				userSessionData.SessionSpentTimes = append(userSessionData.SessionSpentTimes, sessionSpentTime)
				userSessionData.PageCounts = append(userSessionData.PageCounts, pageCount)
				userSessionData.TimeStamps = append(userSessionData.TimeStamps, timestamp)
				userSessionData.WithinQueryPeriod = userSessionData.WithinQueryPeriod || timestamp >= query.From && timestamp <= query.To
				attributedSessionsByUserId[userID][uniqueAttributionKey] = userSessionData
			} else {
				userSessionDataNew := model.UserSessionData{MinTimestamp: timestamp,
					SessionSpentTimes: []float64{sessionSpentTime},
					PageCounts:        []int64{pageCount},
					MaxTimestamp:      timestamp, TimeStamps: []int64{timestamp},
					WithinQueryPeriod: timestamp >= query.From && timestamp <= query.To, MarketingInfo: marketingValues}
				attributedSessionsByUserId[userID][uniqueAttributionKey] = userSessionDataNew
			}
		} else {
			attributedSessionsByUserId[userID] = make(map[string]model.UserSessionData)
			userSessionDataNew := model.UserSessionData{MinTimestamp: timestamp,
				SessionSpentTimes: []float64{sessionSpentTime},
				PageCounts:        []int64{pageCount},
				MaxTimestamp:      timestamp, TimeStamps: []int64{timestamp},
				WithinQueryPeriod: timestamp >= query.From && timestamp <= query.To, MarketingInfo: marketingValues}
			attributedSessionsByUserId[userID][uniqueAttributionKey] = userSessionDataNew
		}
	}
	return attributedSessionsByUserId, userIdsWithSession, nil
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
func (pg *Postgres) getEventInformation(projectId uint64,
	query *model.AttributionQuery) (string, map[string][]interface{}, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectId})
	names := buildEventNamesPlaceholder(query)
	conversionAndFunnelEventMap := make(map[string]bool)
	for _, name := range names {
		conversionAndFunnelEventMap[name] = true
	}
	if _, exists := conversionAndFunnelEventMap[U.EVENT_NAME_SESSION]; !exists {
		names = append(names, U.EVENT_NAME_SESSION)
	}
	eventNames, errCode := pg.GetEventNamesByNames(projectId, names)
	if errCode != http.StatusFound {
		logCtx.Error("failed to find event names")
		return "", nil, errors.New("failed to find event names")
	}
	// this is one to many mapping
	eventNameToId := make(map[string][]interface{})
	// this is one to one mapping
	eventNameIdToName := make(map[string]string)
	for _, event := range eventNames {
		eventNameId := event.ID
		eventName := event.Name
		eventNameIdToName[eventNameId] = eventName
		eventNameToId[eventName] = append(eventNameToId[eventName], eventNameId)
	}
	// there exists only one session event name per project
	if len(eventNameToId[U.EVENT_NAME_SESSION]) == 0 {
		logCtx.Error("$Session Name Id not found")
		return "", nil, errors.New("$Session Name Id not found")
	}
	if len(eventNameToId[query.ConversionEvent.Name]) == 0 {
		logCtx.Error("conversion event name : " + query.ConversionEvent.Name + " not found")
		return "", nil, errors.New("conversion event name : " + query.ConversionEvent.Name + " not found")
	}
	for _, linkedEvent := range query.LinkedEvents {
		if len(eventNameToId[linkedEvent.Name]) == 0 {
			logCtx.Error("linked event name : " + linkedEvent.Name + " not found")
			return "", nil, errors.New("linked event name : " + linkedEvent.Name + " not found")
		}
	}
	sessionEventNameId := eventNameToId[U.EVENT_NAME_SESSION][0].(string)
	return sessionEventNameId, eventNameToId, nil
}

// GetLinkedFunnelEventUsersFilter Adds users who hit funnel event with given {event/user properties} to usersToBeAttributed
func (pg *Postgres) GetLinkedFunnelEventUsersFilter(projectID uint64, queryFrom, queryTo int64,
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
			rows, err := pg.ExecQueryWithContext(queryEventHits, qParams)
			if err != nil {
				logCtx.WithError(err).Error("SQL Query failed for queryEventHits")
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var userID string
				var timestamp int64
				if err = rows.Scan(&userID, &timestamp); err != nil {
					logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
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

// GetConvertedUsersWithFilter Returns the list of eligible users who hit conversion
// event for userProperties from events table
func (pg *Postgres) GetConvertedUsersWithFilter(projectID uint64, goalEventName string,
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
	rows, err := pg.ExecQueryWithContext(queryEventHits, qParams)
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
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
			continue
		}
		if _, ok := userIdHitGoalEventTimestamp[userID]; !ok {
			userIDList = append(userIDList, userID)
			userIdHitGoalEventTimestamp[userID] = timestamp
		}
	}

	// Get coalesced Id for converted user_ids (without filter)
	userIDToCoalIDInfo, err := pg.GetCoalesceIDFromUserIDs(userIDList, projectID)
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
		filteredCoalIDToUserIDInfo[coalUserID] =
			append(filteredCoalIDToUserIDInfo[coalUserID],
				model.UserIDPropID{UserID: userID, Timestamp: timestamp})
		filteredUserIdToUserIDInfo[userID] = model.UserInfo{CoalUserID: coalUserID, Timestamp: timestamp}

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
func updateSessionsMapWithCoalesceID(attributedSessionsByUserID map[string]map[string]model.UserSessionData,
	usersInfo map[string]model.UserInfo) map[string]map[string]model.UserSessionData {

	newSessionsMap := make(map[string]map[string]model.UserSessionData)
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
			newSessionsMap[userInfo.CoalUserID] = make(map[string]model.UserSessionData)
			newSessionsMap[userInfo.CoalUserID][attributionID] = newUserSession
		}
	}
	return newSessionsMap
}

// addWebsiteVisitorsInfo Maps the count distinct users session to campaign id and adds it to attributionData
func addWebsiteVisitorsInfo(attributionData map[string]*model.AttributionData,
	attributedSessionsByUserID map[string]map[string]model.UserSessionData, linkedEventsCount int) {
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
				attributionData[attributionID].Sessions += 1
				userIDAttributionIDVisit[getKey(userID, attributionID)] = true
			}
		}
	}
}

// Merges 2 ids to create a string key
func getKey(id1 string, id2 string) string {
	return id1 + "|_|" + id2
}

// GetAdwordsCurrency Returns currency used for adwords customer_account_id
func (pg *Postgres) GetAdwordsCurrency(projectID uint64, customerAccountID string, from, to int64) (string, error) {

	customerAccountIDs := strings.Split(customerAccountID, ",")
	if len(customerAccountIDs) == 0 {
		return "", errors.New("no ad-words customer account id found")
	}
	queryCurrency := "SELECT value->>'currency_code' AS currency FROM adwords_documents " +
		" WHERE project_id=? AND customer_account_id=? AND type=? AND timestamp BETWEEN ? AND ? " +
		" ORDER BY timestamp DESC LIMIT 1"
	logCtx := log.WithField("ProjectId", projectID)
	// Checking just for customerAccountIDs[0], we are assuming that all accounts have same currency.
	rows, err := pg.ExecQueryWithContext(queryCurrency, []interface{}{projectID, customerAccountIDs[0], 9, U.GetDateOnlyFromTimestamp(from),
		U.GetDateOnlyFromTimestamp(to)})
	if err != nil {
		logCtx.WithError(err).Error("failed to build meta for attribution query result")
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
