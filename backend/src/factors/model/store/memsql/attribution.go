package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// ExecuteAttributionQueryV0 Executes the Attribution using following steps:
//	1. Get all the sessions data (userId, attributionId, timestamp) for given period by attribution key
// 	2. Add the website visitor info using session data from step 1
//	3. i) 	Find out users who hit conversion event applying filter
//	  ii)	Using users from 3.i) find out users who hit linked funnel event applying filter
//	4. Apply attribution methodology
//	5. Add performance data by attributionId
func (store *MemSQL) ExecuteAttributionQueryV0(projectID int64, queryOriginal *model.AttributionQuery,
	debugQueryKey string, enableOptimisedFilterOnProfileQuery,
	enableOptimisedFilterOnEventUserQuery bool) (*model.QueryResult, error) {

	logFields := log.Fields{
		"project_id":        projectID,
		"debug_query_key":   debugQueryKey,
		"attribution_query": true,
	}

	logCtx := log.WithFields(logFields)
	if C.GetAttributionDebug() == 1 {
		logCtx.Info("Hitting ExecuteAttributionQueryV0")
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	queryStartTime := time.Now().UTC().Unix()

	var query *model.AttributionQuery
	U.DeepCopy(queryOriginal, &query)

	// pulling project setting to build attribution query
	settings, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, errors.New("failed to get project settings during attribution call")
	}
	// enrich RunType for attribution query
	err := model.EnrichRequestUsingAttributionConfig(query, settings, logCtx)
	if err != nil {
		return nil, err
	}

	// supporting existing old/saved queries
	model.AddDefaultAnalyzeType(query)
	model.AddDefaultKeyDimensionsToAttributionQuery(query)
	model.AddDefaultMarketingEventTypeTacticOffer(query)

	if query.AttributionKey == model.AttributionKeyLandingPage && query.TacticOfferType == model.MarketingEventTypeTactic {
		return nil, errors.New("can not get landing page level report for Tactic")
	}

	if query.AttributionKey == model.AttributionKeyAllPageView && query.TacticOfferType == model.MarketingEventTypeTactic {
		return nil, errors.New("can not get all page view level report for Tactic")
	}

	// for existing queries and backward support
	if query.QueryType == "" {
		query.QueryType = model.AttributionQueryTypeConversionBased
	}
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, errors.New("failed to get project Settings")
	}

	marketingReports, err := store.FetchMarketingReports(projectID, *query, *projectSetting)
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Fetch marketing report took time")
	}
	if C.GetAttributionDebug() == 1 && projectID == 12384898978000017 {
		log.WithFields(log.Fields{"Attribution": "Debug",
			"AdwordsCampaignIDData":  marketingReports.AdwordsCampaignIDData,
			"AdwordsCampaignKeyData": marketingReports.AdwordsCampaignKeyData}).Info("FetchMarketingReports in ExecuteAttributionQueryV0")
	}
	queryStartTime = time.Now().UTC().Unix()

	if err != nil {
		return nil, err
	}

	err = store.PullCustomDimensionData(projectID, query.AttributionKey, marketingReports, *logCtx)
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Pull Custom dimension data took time")
	}
	queryStartTime = time.Now().UTC().Unix()

	if err != nil {
		return nil, err
	}

	sessionEventNameID, eventNameToIDList, err := store.getEventInformation(projectID, query, *logCtx)
	if err != nil {
		return nil, err
	}

	var contentGroupNamesList []string
	if query.AttributionKey == model.AttributionKeyLandingPage || query.AttributionKey == model.AttributionKeyAllPageView {
		contentGroups, errCode := store.GetAllContentGroups(projectID)
		if errCode != http.StatusFound {
			return nil, errors.New("failed to get content groups")
		}
		for _, contentGroup := range contentGroups {
			contentGroupNamesList = append(contentGroupNamesList, contentGroup.ContentGroupName)
		}
	}

	var usersIDsToAttribute []string
	var kpiData map[string]model.KPIInfo

	// Default conversion for AttributionQueryTypeConversionBased.
	conversionFrom := query.From
	conversionTo := query.To
	// Extend the campaign window for engagement based attribution.
	if query.QueryType == model.AttributionQueryTypeEngagementBased {
		conversionFrom = query.From
		conversionTo = model.LookbackAdjustedTo(query.To, query.LookbackDays)
	}

	coalUserIdConversionTimestamp, userInfo, kpiData, usersIDsToAttribute, err3 := store.PullConvertedUsers(projectID, query, conversionFrom, conversionTo, eventNameToIDList,
		debugQueryKey, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery, logCtx)

	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"KPIAttribution": "Debug",
			"kpiData":                       kpiData,
			"coalUserIdConversionTimestamp": coalUserIdConversionTimestamp,
			"usersIDsToAttribute":           usersIDsToAttribute}).Info("Attributable users list - ConvertedUsers")
	}

	if err3 != nil {
		return nil, err3
	}

	var userData map[string]map[string]model.UserSessionData
	var err4 error
	if query.AttributionKey == model.AttributionKeyAllPageView {
		userData, err4 = store.PullPagesOfConvertedUsers(projectID, query, sessionEventNameID, usersIDsToAttribute, marketingReports, contentGroupNamesList, logCtx)
	} else {
		userData, err4 = store.PullSessionsOfConvertedUsers(projectID, query, sessionEventNameID, usersIDsToAttribute, marketingReports, contentGroupNamesList, logCtx)
	}
	if err4 != nil {
		return nil, err4
	}

	if query.AnalyzeType == model.AnalyzeTypeUserKPI {
		log.WithFields(log.Fields{"UserKPIAttribution": "Debug", "sessions": userData}).Info("UserKPI Attribution sessions")
	}

	if C.GetAttributionDebug() == 1 && projectID == 12384898978000017 {
		log.WithFields(log.Fields{"Attribution": "Debug",
			"CampaignChannelGroupMapping": marketingReports.CampaignChannelGroupMapping,
			"sessions":                    userData}).Info("CampaignChannelGroupMapping after PullSessions")
	}

	// Pull Offline touch points for all the cases: "Tactic",  "Offer", "TacticOffer"
	store.AppendOTPSessions(projectID, query, &userData, *logCtx)
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Pull Offline touch points user data took time")
	}
	queryStartTime = time.Now().UTC().Unix()

	attributionData, isCompare, err2 := store.GetAttributionData(query, userData, userInfo, coalUserIdConversionTimestamp, marketingReports, kpiData, logCtx)
	if err2 != nil {
		return nil, err2
	}

	// Filter out the key values from query (apply filter after performance enrichment)
	model.ApplyFilter(attributionData, query)
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Metrics, Performance report, filter took time")
	}
	queryStartTime = time.Now().UTC().Unix()

	result := ProcessAttributionDataToResult(projectID, query, attributionData, isCompare, queryStartTime, marketingReports, kpiData, logCtx)
	result.Meta.Currency = ""
	if projectSetting.IntAdwordsCustomerAccountId != nil && *projectSetting.IntAdwordsCustomerAccountId != "" {
		currency, _ := store.GetAdwordsCurrency(projectID, *projectSetting.IntAdwordsCustomerAccountId, query.From, query.To, *logCtx)
		result.Meta.Currency = currency
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Total query took time")
	}
	model.SanitizeResult(result)
	return result, nil
}

func (store *MemSQL) AppendOTPSessions(projectID int64, query *model.AttributionQuery,
	sessions *map[string]map[string]model.UserSessionData, logCtx log.Entry) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	otpEvent, err := store.getOfflineEventData(projectID, logCtx)
	if err != nil {
		logCtx.Info("no OTP events/sessions found. Skipping computation")
		return
	}

	_sessionsOTP, sessionOTPUsers, err := store.fetchOTPSessions(projectID, otpEvent.ID, query, logCtx)
	if err != nil {
		logCtx.Info("fetchOTPSessions failed for OTP events/sessions")
		return
	}

	usersInfoOTP, _, err := store.GetCoalesceIDFromUserIDs(sessionOTPUsers, projectID, logCtx)
	if err != nil {
		logCtx.Info("no users found for OTP events/sessions found. Skipping computation")
		return
	}

	model.UpdateSessionsMapWithCoalesceID(_sessionsOTP, usersInfoOTP, sessions)
}

// GetCoalesceIDFromUserIDs returns the map of coalesce userId for given list of users
func (store *MemSQL) GetCoalesceIDFromUserIDs(userIDs []string, projectID int64,
	logCtx log.Entry) (map[string]model.UserInfo, []string, error) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	userIDsInBatches := U.GetStringListAsBatch(userIDs, model.UserBatchSize)
	userIDToCoalUserIDMap := make(map[string]model.UserInfo)
	batch := 1
	for _, users := range userIDsInBatches {
		placeHolder := U.GetValuePlaceHolder(len(users))
		value := U.GetInterfaceList(users)
		queryUserIDCoalID := "SELECT id, COALESCE(users.customer_user_id,users.id) AS coal_user_id" + " " +
			"FROM users WHERE project_id=? and id IN (" + placeHolder + ")"
		var gULParams []interface{}
		gULParams = append(gULParams, projectID)
		gULParams = append(gULParams, value...)
		rows, tx, err, reqID := store.ExecQueryWithContext(queryUserIDCoalID, gULParams)
		if err != nil {
			logCtx.WithError(err).Error("SQL Query failed for GetCoalesceIDFromUserIDs")
			return nil, nil, err
		}
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"Batch": batch}).Info("Executing GetCoalesceIDFromUserIDs")
		}
		batch++
		startReadTime := time.Now()
		for rows.Next() {
			var userID string
			var coalesceID string

			if err = rows.Scan(&userID, &coalesceID); err != nil {
				logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
				continue
			}
			userIDToCoalUserIDMap[userID] = model.UserInfo{CoalUserID: coalesceID}
		}
		err = rows.Err()
		if err != nil {
			// Error from DB is captured eg: timeout error
			logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in GetCoalesceIDFromUserIDs")
			return nil, nil, err
		}
		U.CloseReadQuery(rows, tx)
		U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &log.Fields{"project_id": projectID})
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"user_count": len(userIDToCoalUserIDMap)}).Info("GetCoalesceIDFromUserIDs 3")
	}

	// Get all CoalIDs
	_allCoalIds := make(map[string]int)
	var coalIDsList []string
	// Get user IDs for Revenue Attribution
	for _, v := range userIDToCoalUserIDMap {
		_allCoalIds[v.CoalUserID] = 1
	}

	for id := range _allCoalIds {
		coalIDsList = append(coalIDsList, id)
	}

	return userIDToCoalUserIDMap, coalIDsList, nil
}

// getOfflineEventData returns  offline touch point event id
func (store *MemSQL) getOfflineEventData(projectID int64, logCtx log.Entry) (model.EventName, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	names := []string{U.EVENT_NAME_OFFLINE_TOUCH_POINT}

	eventNames, errCode := store.GetEventNamesByNames(projectID, names)
	if errCode != http.StatusFound || len(eventNames) != 1 {
		if C.GetAttributionDebug() == 1 {
			logCtx.Info("failed to find offline touch point event names, skipping OTP attribution computation")
		}
		return model.EventName{}, errors.New("failed to find offline touch point event names, skipping OTP attribution computation")
	}
	return eventNames[0], nil
}

// getEventInformation Returns conversion event Id, list of all event_ids(Conversion and funnel events) and a Id to name mapping
func (store *MemSQL) getEventInformation(projectId int64,
	query *model.AttributionQuery, logCtx log.Entry) (string, map[string][]interface{}, error) {

	names := model.BuildEventNamesPlaceholder(query)
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
		if _, exists := eventNameToId[eventName]; !exists {
			eventNameToId[eventName] = []interface{}{}
		}
		eventNameToId[eventName] = append(eventNameToId[eventName], eventNameId)
	}
	// there exists only one session event name per project
	if len(eventNameToId[U.EVENT_NAME_SESSION]) == 0 {
		logCtx.Error("$Session Name Id not found")
		return "", nil, errors.New("$Session Name Id not found")
	}
	if len(eventNameToId[query.ConversionEvent.Name]) == 0 && query.AnalyzeType == model.AnalyzeTypeUsers {
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
func (store *MemSQL) GetLinkedFunnelEventUsersFilter(projectID int64, queryFrom, queryTo int64,
	linkedEvents []model.QueryEventWithProperties, eventNameToId map[string][]interface{},
	userIDInfo map[string]model.UserInfo, logCtx log.Entry) (error, []model.UserEventInfo) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	var usersToBeAttributed []model.UserEventInfo
	var coalUserIdsHitConversion []string
	for _, v := range userIDInfo {
		coalUserIdsHitConversion = append(coalUserIdsHitConversion, v.CoalUserID)
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
		userPropertiesIdsInBatches := U.GetStringListAsBatch(coalUserIdsHitConversion, model.UserBatchSize)
		for _, coalUserIds := range userPropertiesIdsInBatches {

			// add user batching
			coalUserIdPlaceHolder := U.GetValuePlaceHolder(len(coalUserIds))
			value := U.GetInterfaceList(coalUserIds)

			selectEventHits := "SELECT events.user_id, events.timestamp FROM events" +
				" JOIN users ON events.user_id = users.id AND users.project_id= ? "
			whereEventHits := " WHERE events.project_id=? AND " +
				" timestamp >= ? AND timestamp <=? AND events.event_name_id IN (" + eventsPlaceHolder + ") " +
				" AND COALESCE(users.customer_user_id,users.id) IN ( " + coalUserIdPlaceHolder + " ) "

			qParams := []interface{}{projectID, projectID, queryFrom, queryTo}
			qParams = append(qParams, linkedEventNameIDs...)
			qParams = append(qParams, value...)

			// add event filter
			wStmtEvent, wParamsEvent, _, err := getFilterSQLStmtForEventProperties(
				projectID, linkedEvent.Properties, queryFrom)
			if err != nil {
				return err, nil
			}

			if wStmtEvent != "" {
				whereEventHits = whereEventHits + " AND " + fmt.Sprintf("( %s )", wStmtEvent)
				qParams = append(qParams, wParamsEvent...)
			}

			// add user filter
			wStmtUser, wParamsUser, _, err := getFilterSQLStmtForUserProperties(projectID, linkedEvent.Properties, queryFrom)
			if err != nil {
				return err, nil
			}

			if wStmtUser != "" {
				whereEventHits = whereEventHits + " AND " + fmt.Sprintf("( %s )", wStmtUser)
				qParams = append(qParams, wParamsUser...)
			}

			queryEventHits := selectEventHits + " " + whereEventHits

			// fetch query results
			rows, tx, err, reqID := store.ExecQueryWithContext(queryEventHits, qParams)
			if err != nil {
				logCtx.WithError(err).Error("SQL Query failed for queryEventHits")
				return err, nil
			}

			startReadTime := time.Now()
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
				} else {
					// record the fist occurrence of the event by userID
					if timestamp < userIDHitGoalEventTimestamp[userID] {
						userIDHitGoalEventTimestamp[userID] = timestamp
					}
				}
			}
			err = rows.Err()
			if err != nil {
				// Error from DB is captured eg: timeout error
				logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in GetLinkedFunnelEventUsersFilter")
				return err, nil
			}

			U.CloseReadQuery(rows, tx)
			U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &log.Fields{"project_id": projectID})
		}
		// Get coalesced ID for Funnel Event user_ids
		userIDToCoalIDInfo, _, err := store.GetCoalesceIDFromUserIDs(userIDList, projectID, logCtx)
		if err != nil {
			return err, nil
		}
		// add the filtered users with eventId usersToBeAttributed
		for _, userId := range userIDList {
			usersToBeAttributed = append(usersToBeAttributed,
				model.UserEventInfo{CoalUserID: userIDToCoalIDInfo[userId].CoalUserID, EventName: linkedEvent.Name,
					Timestamp: userIDHitGoalEventTimestamp[userId], EventType: model.EventTypeLinkedFunnelEvent})
		}
	}
	return nil, usersToBeAttributed
}

// GetAdwordsCurrency Returns currency used for Google Ads (adwords) customer_account_id
func (store *MemSQL) GetAdwordsCurrency(projectID int64, customerAccountID string, from, to int64, logCtx log.Entry) (string, error) {
	logFields := log.Fields{
		"project_id":          projectID,
		"customer_account_id": customerAccountID,
		"from":                from,
		"to":                  to,
	}
	logCtx = *logCtx.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	// Check for no-adwords account linked
	if customerAccountID == "" {
		return "", errors.New("no ad-words customer account id found")
	}
	customerAccountIDs := strings.Split(customerAccountID, ",")
	if len(customerAccountIDs) == 0 {
		return "", errors.New("no ad-words customer account id found")
	}
	queryCurrency := "SELECT JSON_EXTRACT_STRING(value, 'currency_code') AS currency FROM adwords_documents " +
		" WHERE project_id=? AND customer_account_id=? AND type=? AND timestamp BETWEEN ? AND ? " +
		" ORDER BY timestamp DESC LIMIT 1"

	// Checking just for customerAccountIDs[0], we are assuming that all accounts have same currency.
	rows, tx, err, reqID := store.ExecQueryWithContext(queryCurrency,
		[]interface{}{projectID, customerAccountIDs[0], 9, U.GetDateOnlyFromTimestampZ(from),
			U.GetDateOnlyFromTimestampZ(to)})
	if err != nil {
		logCtx.WithError(err).Error("failed to build meta for attribution query result")
		return "", err
	}
	defer U.CloseReadQuery(rows, tx)

	startReadTime := time.Now()
	var currency string
	for rows.Next() {
		if err = rows.Scan(&currency); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed.")
			return "", err
		}
	}
	err = rows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in GetAdwordsCurrency")
		return currency, err
	}
	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &logFields)

	return currency, nil
}

// Returns the all the sessions (userId,attributionId,minTimestamp,maxTimestamp) for given
// users from given period including look back
func (store *MemSQL) getAllThePages(projectId int64, sessionEventNameId string, query *model.AttributionQuery, usersToPullSessionFor []string,
	reports *model.MarketingReports, contentGroupNamesList []string, logCtx log.Entry) (map[string]map[string]model.UserSessionData, []string, error) {
	logFields := log.Fields{
		"project_id":            projectId,
		"session_event_name_id": sessionEventNameId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx = *logCtx.WithFields(logFields)
	effectiveFrom := model.LookbackAdjustedFrom(query.From, query.LookbackDays)
	effectiveTo := query.To
	// extend the campaign window for engagement based attribution
	if query.QueryType == model.AttributionQueryTypeEngagementBased {
		effectiveFrom = model.LookbackAdjustedFrom(query.From, query.LookbackDays)
		effectiveTo = model.LookbackAdjustedTo(query.To, query.LookbackDays)
	}
	attributionEventKey, err := model.GetQuerySessionProperty(query.AttributionKey)
	if err != nil {
		return nil, nil, err
	}

	attributedSessionsByUserId := make(map[string]map[string]model.UserSessionData)
	var userIdsWithSession []string

	userIDsInBatches := U.GetStringListAsBatch(usersToPullSessionFor, model.UserBatchSize)
	for _, users := range userIDsInBatches {

		placeHolder := U.GetValuePlaceHolder(len(users))
		value := U.GetInterfaceList(users)

		contentGroupNamesToDummyNamesMap := model.GetContentGroupNamesToDummyNamesMap(contentGroupNamesList)
		caseSelectStmt := "CASE WHEN JSON_EXTRACT_STRING(sessions.properties, ?) IS NULL THEN ? " +
			" WHEN JSON_EXTRACT_STRING(sessions.properties, ?) = '' THEN ? ELSE JSON_EXTRACT_STRING(sessions.properties, ?) END"

		queryUserSessionTimeRange := "SELECT sessions.user_id, " +
			caseSelectStmt + " AS campaignID, " +
			caseSelectStmt + " AS campaignName, " +
			caseSelectStmt + " AS adgroupID, " +
			caseSelectStmt + " AS adgroupName, " +
			caseSelectStmt + " AS keywordName, " +
			caseSelectStmt + " AS keywordMatchType, " +
			caseSelectStmt + " AS source, " +
			caseSelectStmt + " AS channel, " +
			caseSelectStmt + " AS attribution_id, " +
			caseSelectStmt + " AS gcl_id, " +
			caseSelectStmt + " AS landingPageUrl, " +
			caseSelectStmt + " AS allPageViewUrl, "

		var qParams []interface{}

		qParams = append(qParams,
			U.EP_CAMPAIGN_ID, model.PropertyValueNone, U.EP_CAMPAIGN_ID, model.PropertyValueNone, U.EP_CAMPAIGN_ID,
			U.EP_CAMPAIGN, model.PropertyValueNone, U.EP_CAMPAIGN, model.PropertyValueNone, U.EP_CAMPAIGN,
			U.EP_ADGROUP_ID, model.PropertyValueNone, U.EP_ADGROUP_ID, model.PropertyValueNone, U.EP_ADGROUP_ID,
			U.EP_ADGROUP, model.PropertyValueNone, U.EP_ADGROUP, model.PropertyValueNone, U.EP_ADGROUP,
			U.EP_KEYWORD, model.PropertyValueNone, U.EP_KEYWORD, model.PropertyValueNone, U.EP_KEYWORD,
			U.EP_KEYWORD_MATCH_TYPE, model.PropertyValueNone, U.EP_KEYWORD_MATCH_TYPE, model.PropertyValueNone, U.EP_KEYWORD_MATCH_TYPE,
			U.EP_SOURCE, model.PropertyValueNone, U.EP_SOURCE, model.PropertyValueNone, U.EP_SOURCE,
			U.EP_CHANNEL, model.PropertyValueNone, U.EP_CHANNEL, model.PropertyValueNone, U.EP_CHANNEL,
			attributionEventKey, model.PropertyValueNone, attributionEventKey, model.PropertyValueNone, attributionEventKey,
			U.EP_GCLID, model.PropertyValueNone, U.EP_GCLID, model.PropertyValueNone, U.EP_GCLID,
			U.UP_INITIAL_PAGE_URL, model.PropertyValueNone, U.UP_INITIAL_PAGE_URL, model.PropertyValueNone, U.UP_INITIAL_PAGE_URL,
			U.EP_PAGE_URL, model.PropertyValueNone, U.EP_PAGE_URL, model.PropertyValueNone, U.EP_PAGE_URL)

		wStmt, wParams, err := getSelectSQLStmtForContentGroup(contentGroupNamesToDummyNamesMap)
		if err != nil {
			return nil, nil, err
		}
		queryUserSessionTimeRange = queryUserSessionTimeRange + wStmt
		qParams = append(qParams, wParams...)

		queryUserSessionTimeRange = queryUserSessionTimeRange +
			" sessions.timestamp FROM events AS sessions " +
			" WHERE sessions.project_id=? " +

			// Filter page view event.
			"AND sessions.user_id IN (" + placeHolder + " ) AND sessions.timestamp BETWEEN ? AND ?" +
			" AND (JSON_EXTRACT_STRING(sessions.properties, ?)=? )"

		wParams = []interface{}{projectId}
		wParams = append(wParams, value...)
		wParams = append(wParams, effectiveFrom, effectiveTo, U.EP_IS_PAGE_VIEW, "true")
		qParams = append(qParams, wParams...)

		rows, tx, err, reqID := store.ExecQueryWithContext(queryUserSessionTimeRange, qParams)
		if err != nil {
			logCtx.WithError(err).Error("SQL Query failed")
			return nil, nil, err
		}
		if C.GetAttributionDebug() == 1 {
			logCtx.Info("Attribution before ProcessEventRows")
		}
		processErr := model.ProcessEventRows(rows, query, reports, contentGroupNamesList, &attributedSessionsByUserId, &userIdsWithSession, logCtx, reqID)
		U.CloseReadQuery(rows, tx)
		if processErr != nil {
			return attributedSessionsByUserId, userIdsWithSession, processErr
		}
	}

	return attributedSessionsByUserId, userIdsWithSession, nil
}
