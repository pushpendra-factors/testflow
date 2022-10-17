package memsql

import (
	"database/sql"
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// ExecuteAttributionQueryV1 Todo Pre-compute's online version - add details once available to run
// @Deprecated
func (store *MemSQL) ExecuteAttributionQueryV1(projectID int64, queryOriginal *model.AttributionQuery,
	debugQueryKey string, enableOptimisedFilterOnProfileQuery,
	enableOptimisedFilterOnEventUserQuery bool) (*model.QueryResult, error) {

	logFields := log.Fields{
		"project_id":        projectID,
		"debug_query_key":   debugQueryKey,
		"attribution_query": true,
	}

	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	queryStartTime := time.Now().UTC().Unix()

	var query *model.AttributionQuery
	U.DeepCopy(queryOriginal, &query)
	// supporting existing old/saved queries
	model.AddDefaultAnalyzeType(query)
	model.AddDefaultKeyDimensionsToAttributionQuery(query)
	model.AddDefaultMarketingEventTypeTacticOffer(query)

	if query.AttributionKey == model.AttributionKeyLandingPage && query.TacticOfferType != model.MarketingEventTypeOffer {
		return nil, errors.New("can not get landing page level report for Tactic/TacticOffer")
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
	logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Fetch marketing report took time")
	queryStartTime = time.Now().UTC().Unix()

	if err != nil {
		return nil, err
	}

	err = store.PullCustomDimensionData(projectID, query.AttributionKey, marketingReports, *logCtx)
	logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Pull Custom dimension data took time")
	queryStartTime = time.Now().UTC().Unix()

	if err != nil {
		return nil, err
	}

	logCtx.Info("Done PullCustomDimensionData")

	sessionEventNameID, eventNameToIDList, err := store.getEventInformation(projectID, query, *logCtx)
	if err != nil {
		return nil, err
	}

	var contentGroupNamesList []string
	if query.AttributionKey == model.AttributionKeyLandingPage {
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
		conversionTo = model.LookbackAdjustedTo(query.To, query.LookbackDays, U.TimeZoneString(query.Timezone))
	}

	coalUserIdConversionTimestamp, userInfo, kpiData, usersIDsToAttribute, err3 := store.PullConvertedUsers(projectID, query, conversionFrom, conversionTo, eventNameToIDList,
		kpiData, debugQueryKey, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery, logCtx)

	if err3 != nil {
		return nil, err3
	}

	sessions, err4 := store.PullSessionsOfConvertedUsers(projectID, query, sessionEventNameID, usersIDsToAttribute, marketingReports, contentGroupNamesList, logCtx)
	if err4 != nil {
		return nil, err4
	}

	// Pull Offline touch points for all the cases: "Tactic",  "Offer", "TacticOffer"
	store.AppendOTPSessions(projectID, query, &sessions, *logCtx)
	logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Pull Offline touch points user data took time")
	queryStartTime = time.Now().UTC().Unix()

	attributionData, isCompare, err2 := store.GetAttributionData(projectID, query, sessions, userInfo, coalUserIdConversionTimestamp, marketingReports, kpiData, logCtx)
	if err2 != nil {
		return nil, err2
	}

	// Filter out the key values from query (apply filter after performance enrichment)
	model.ApplyFilter(attributionData, query)
	logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Metrics, Performance report, filter took time")

	queryStartTime = time.Now().UTC().Unix()

	result := ProcessAttributionDataToResult(projectID, query, attributionData, isCompare, queryStartTime, marketingReports, kpiData, logCtx)
	result.Meta.Currency = ""
	if projectSetting.IntAdwordsCustomerAccountId != nil && *projectSetting.IntAdwordsCustomerAccountId != "" {
		currency, _ := store.GetAdwordsCurrency(projectID, *projectSetting.IntAdwordsCustomerAccountId, query.From, query.To, *logCtx)
		result.Meta.Currency = currency
	}

	logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Total query took time")

	model.SanitizeResult(result)
	return result, nil
}

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
	logCtx.Info("Hitting ExecuteAttributionQueryV1")
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	queryStartTime := time.Now().UTC().Unix()

	var query *model.AttributionQuery
	U.DeepCopy(queryOriginal, &query)
	// supporting existing old/saved queries
	model.AddDefaultAnalyzeType(query)
	model.AddDefaultKeyDimensionsToAttributionQuery(query)
	model.AddDefaultMarketingEventTypeTacticOffer(query)

	if query.AttributionKey == model.AttributionKeyLandingPage && query.TacticOfferType != model.MarketingEventTypeOffer {
		return nil, errors.New("can not get landing page level report for Tactic/TacticOffer")
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
	if C.GetAttributionDebug() == 1 {
		logCtx.Info("Done PullCustomDimensionData")
	}
	sessionEventNameID, eventNameToIDList, err := store.getEventInformation(projectID, query, *logCtx)
	if err != nil {
		return nil, err
	}

	var contentGroupNamesList []string
	if query.AttributionKey == model.AttributionKeyLandingPage {
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
		conversionTo = model.LookbackAdjustedTo(query.To, query.LookbackDays, U.TimeZoneString(query.Timezone))
	}

	coalUserIdConversionTimestamp, userInfo, kpiData, usersIDsToAttribute, err3 := store.PullConvertedUsers(projectID, query, conversionFrom, conversionTo, eventNameToIDList,
		kpiData, debugQueryKey, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery, logCtx)

	if err3 != nil {
		return nil, err3
	}

	sessions, err4 := store.PullSessionsOfConvertedUsers(projectID, query, sessionEventNameID, usersIDsToAttribute, marketingReports, contentGroupNamesList, logCtx)
	if err4 != nil {
		return nil, err4
	}

	// Pull Offline touch points for all the cases: "Tactic",  "Offer", "TacticOffer"
	store.AppendOTPSessions(projectID, query, &sessions, *logCtx)
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Pull Offline touch points user data took time")
	}
	queryStartTime = time.Now().UTC().Unix()

	attributionData, isCompare, err2 := store.GetAttributionData(projectID, query, sessions, userInfo, coalUserIdConversionTimestamp, marketingReports, kpiData, logCtx)
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

func (store *MemSQL) PullSessionsOfConvertedUsers(projectID int64, query *model.AttributionQuery, sessionEventNameID string, usersToBeAttributed []string,
	marketingReports *model.MarketingReports, contentGroupNamesList []string, logCtx *log.Entry) (map[string]map[string]model.UserSessionData, error) {

	sessions := make(map[string]map[string]model.UserSessionData)

	queryStartTime := time.Now().UTC().Unix()
	// Pull Sessions for the cases: "Tactic" and "TacticOffer".
	// If Landing Page level report, pull for offer as well.
	if query.TacticOfferType != model.MarketingEventTypeOffer || query.AttributionKey == model.AttributionKeyLandingPage {
		// Get all the sessions (userId, attributionId, UserSessionData) for given period by attribution key
		_sessions, sessionUsers, err := store.getAllTheSessionsV1(projectID, sessionEventNameID, query, usersToBeAttributed, marketingReports, contentGroupNamesList, *logCtx)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Pull Sessions data data took time")
		}
		queryStartTime = time.Now().UTC().Unix()

		if err != nil {
			return nil, err
		}

		usersInfo, _, err := store.GetCoalesceIDFromUserIDs(sessionUsers, projectID, *logCtx)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Get Coalesce user data took time")
		}
		queryStartTime = time.Now().UTC().Unix()
		if err != nil {
			return nil, err
		}

		model.UpdateSessionsMapWithCoalesceID(_sessions, usersInfo, &sessions)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Update Sessions Coalesce user data took time")
		}
		queryStartTime = time.Now().UTC().Unix()
	}
	return sessions, nil
}

func (store *MemSQL) PullConvertedUsers(projectID int64, query *model.AttributionQuery, conversionFrom int64, conversionTo int64,
	eventNameToIDList map[string][]interface{}, kpiData map[string]model.KPIInfo,
	debugQueryKey string, enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool,
	logCtx *log.Entry) (map[string]int64, []model.UserEventInfo, map[string]model.KPIInfo, []string, error) {

	var coalUserIdConversionTimestamp map[string]int64
	var usersToBeAttributed []model.UserEventInfo
	var usersIDsToAttribute []string
	var err error

	if query.AnalyzeType == model.AnalyzeTypeUsers {
		var _userIDToInfoConverted map[string]model.UserInfo
		_userIDToInfoConverted, usersToBeAttributed, coalUserIdConversionTimestamp, err = store.GetConvertedUsers(projectID,
			conversionFrom, conversionTo, query.ConversionEvent,
			query, eventNameToIDList, *logCtx)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		// Get user IDs for AnalyzeTypeUsers
		for id, _ := range _userIDToInfoConverted {
			usersIDsToAttribute = append(usersIDsToAttribute, id)
		}
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"UniqueUsers": len(usersIDsToAttribute)}).Info("Total users for the attribution query")
		}
	} else if query.AnalyzeType == model.AnalyzeTypeUserKPI {

		var err error
		queryStartTime := time.Now().UTC().Unix()
		kpiData, err = store.ExecuteUserKPIForAttribution(projectID, query, debugQueryKey,
			*logCtx, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("UserKPI query execution took time")
		}
		if err != nil {
			return nil, nil, nil, nil, err
		}

		if C.GetAttributionDebug() == 1 {
			log.WithFields(log.Fields{"UserKPIAttribution": "Debug", "kpiData": kpiData}).Info("UserKPI Attribution kpiData")
		}

		_uniqueUsers := make(map[string]int)
		// Get user IDs for Revenue Attribution
		for _, data := range kpiData {
			for _, userID := range data.KpiUserIds {
				_uniqueUsers[userID] = 1
			}
		}

		for id, _ := range _uniqueUsers {
			usersIDsToAttribute = append(usersIDsToAttribute, id)
		}
	} else {
		// This thread is for query.AnalyzeType == model.AnalyzeTypeHSDeals || query.AnalyzeType == model.AnalyzeTypeSFOpportunities.
		var err error
		queryStartTime := time.Now().UTC().Unix()
		kpiData, err = store.ExecuteKPIForAttribution(projectID, query, debugQueryKey,
			*logCtx, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("KPI query execution took time")
		}
		if err != nil {
			return nil, nil, nil, nil, err
		}

		if C.GetAttributionDebug() == 1 {
			log.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData}).Info("KPI Attribution kpiData")
		}

		_uniqueUsers := make(map[string]int)
		// Get user IDs for Revenue Attribution
		for _, data := range kpiData {
			for _, userID := range data.KpiUserIds {
				_uniqueUsers[userID] = 1
			}
		}

		for id, _ := range _uniqueUsers {
			usersIDsToAttribute = append(usersIDsToAttribute, id)
		}
	}
	return coalUserIdConversionTimestamp, usersToBeAttributed, kpiData, usersIDsToAttribute, nil
}

func (store *MemSQL) GetAttributionData(projectID int64, query *model.AttributionQuery, sessions map[string]map[string]model.UserSessionData,
	usersToBeAttributed []model.UserEventInfo, coalUserIdConversionTimestamp map[string]int64, marketingReports *model.MarketingReports,
	kpiData map[string]model.KPIInfo, logCtx *log.Entry) (*map[string]*model.AttributionData, bool, error) {

	queryStartTime := time.Now().UTC().Unix()

	isCompare := false
	var attributionData *map[string]*model.AttributionData

	if query.AnalyzeType == model.AnalyzeTypeUsers {

		// Build attribution weight
		sessionWT := make(map[string][]float64)
		for key := range sessions {
			// since we support only one event
			sessionWT[key] = []float64{float64(1)}
		}

		if C.GetAttributionDebug() == 1 {
			uniqUsers := len(sessions)
			logCtx.WithFields(log.Fields{"AttributionDebug": "sessions"}).Info(fmt.Sprintf("Total users with session: %d", uniqUsers))
		}
		var err error
		attributionData, isCompare, err = store.FireAttributionV1(query, &usersToBeAttributed, &coalUserIdConversionTimestamp, sessions, sessionWT, *logCtx)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("FireAttribution took time")
		}
		queryStartTime = time.Now().UTC().Unix()
		if C.GetAttributionDebug() == 1 {
			logCtx.Info("Done FireAttribution")
		}
		if err != nil {
			return nil, false, err
		}

		// single function for user type queries
		convAggFunctionType := []string{"unique"}
		for key, _ := range *attributionData {
			(*attributionData)[key].ConvAggFunctionType = convAggFunctionType
		}

		// Add the Added keys with no of conversion event = 1
		model.AddTheAddedKeysAndMetrics(attributionData, query, sessions, 1)

		// Add the performance information no of conversion event = 1
		model.AddPerformanceData(attributionData, query.AttributionKey, marketingReports, 1)
		for key, _ := range *attributionData {
			(*attributionData)[key].ConvAggFunctionType = convAggFunctionType
		}
		if C.GetAttributionDebug() == 1 {
			logCtx.Info("Done AddTheAddedKeysAndMetrics, AddPerformanceData")
		}
	} else {

		// creating group sessions by transforming sessions
		groupSessions := make(map[string]map[string]model.UserSessionData)

		for kpiID, kpiInfo := range kpiData {

			if _, exists := groupSessions[kpiID]; !exists {
				groupSessions[kpiID] = make(map[string]model.UserSessionData)
			}
			if kpiInfo.KpiCoalUserIds == nil || len(kpiInfo.KpiCoalUserIds) == 0 {
				if C.GetAttributionDebug() == 1 {
					logCtx.WithFields(log.Fields{"KpiInfo": kpiInfo, "KPI_ID": kpiID}).Info("no user found for the KPI group, ignoring")
				}
				//groupSessions[kpiID][noneKey] = model.UserSessionData{}
				continue
			}
			for _, user := range kpiInfo.KpiCoalUserIds {
				// check if user has session/otp
				if _, exists := sessions[user]; !exists {
					if C.GetAttributionDebug() == 1 {
						logCtx.WithFields(log.Fields{"User": user, "KPI_ID": kpiID}).Info("user without session/otp")
					}
					continue
				}

				userSession := sessions[user]

				for attributionKey, newUserSession := range userSession {

					if existingUserSession, exists := groupSessions[kpiID][attributionKey]; exists {
						// Update the existing attribution first and last touch.
						existingUserSession.MinTimestamp = U.Min(existingUserSession.MinTimestamp, newUserSession.MinTimestamp)
						existingUserSession.MaxTimestamp = U.Max(existingUserSession.MaxTimestamp, newUserSession.MaxTimestamp)
						// Merging timestamp of same customer having 2 userIds.
						existingUserSession.TimeStamps = append(existingUserSession.TimeStamps, newUserSession.TimeStamps...)
						existingUserSession.WithinQueryPeriod = existingUserSession.WithinQueryPeriod || newUserSession.WithinQueryPeriod
						groupSessions[kpiID][attributionKey] = existingUserSession
					} else {
						groupSessions[kpiID][attributionKey] = newUserSession
					}
				}
			}

			// for new users who may have customer id not set for global users
			for _, user := range kpiInfo.KpiUserIds {
				// check if user has session/otp
				if _, exists := sessions[user]; !exists {
					if C.GetAttributionDebug() == 1 {
						logCtx.WithFields(log.Fields{"User": user, "KPI_ID": kpiID}).Info("user without session/otp")
					}
					continue
				}

				userSession := sessions[user] // map[string]model.UserSessionData

				for attributionKey, newUserSession := range userSession {

					if existingUserSession, exists := groupSessions[kpiID][attributionKey]; exists {
						// Update the existing attribution first and last touch.
						existingUserSession.MinTimestamp = U.Min(existingUserSession.MinTimestamp, newUserSession.MinTimestamp)
						existingUserSession.MaxTimestamp = U.Max(existingUserSession.MaxTimestamp, newUserSession.MaxTimestamp)
						// Merging timestamp of same customer having 2 userIds.
						existingUserSession.TimeStamps = append(existingUserSession.TimeStamps, newUserSession.TimeStamps...)
						existingUserSession.WithinQueryPeriod = existingUserSession.WithinQueryPeriod || newUserSession.WithinQueryPeriod
						groupSessions[kpiID][attributionKey] = existingUserSession
					} else {
						groupSessions[kpiID][attributionKey] = newUserSession
					}
				}
			}
		}
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"KPIGroupSession": groupSessions}).Info("KPI-Attribution Group session 2")
		}
		found := false
		for _, data := range groupSessions {
			for _, journey := range data {
				if len(journey.TimeStamps) > 0 {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			if C.GetAttributionDebug() == 1 {
				logCtx.Info("no user journey found (neither sessions nor offline touch points)")
			}
			return nil, false, errors.New("no user journey found (neither sessions nor offline touch points)")
		}

		// Build attribution weight
		noOfConversionEvents := 1
		sessionWT := make(map[string][]float64)
		for key := range groupSessions {
			sessionWT[key] = kpiData[key].KpiValues
			if kpiData[key].KpiValues != nil || len(kpiData[key].KpiValues) > 1 {
				noOfConversionEvents = U.MaxInt(noOfConversionEvents, len(kpiData[key].KpiValues))
			}
		}

		if C.GetAttributionDebug() == 1 {
			uniqUsers := len(groupSessions)
			logCtx.WithFields(log.Fields{"AttributionDebug": "sessions"}).Info(fmt.Sprintf("Total users with session: %d", uniqUsers))
		}
		var err error
		attributionData, isCompare, err = store.FireAttributionForKPI(projectID, query, groupSessions, kpiData, sessionWT, *logCtx)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("FireAttribution KPI took time")
		}
		queryStartTime = time.Now().UTC().Unix()
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData")
		}

		if err != nil {
			return nil, false, err
		}

		if C.GetAttributionDebug() == 1 {
			uniqueKeys := len(*attributionData)
			logCtx.WithFields(log.Fields{"AttributionDebug": "attributionData"}).Info(fmt.Sprintf("Total users with session: %d", uniqueKeys))
		}

		// for KPI queries, use the kpiData.KpiAggFunctionTypes as ConvAggFunctionType
		var convAggFunctionType []string
		for _, val := range kpiData {
			if len(val.KpiAggFunctionTypes) > 0 {
				convAggFunctionType = val.KpiAggFunctionTypes
				break
			}
		}
		for key := range *attributionData {
			(*attributionData)[key].ConvAggFunctionType = convAggFunctionType
		}

		// Add the Added keys
		model.AddTheAddedKeysAndMetrics(attributionData, query, groupSessions, noOfConversionEvents)

		// Add the performance information
		model.AddPerformanceData(attributionData, query.AttributionKey, marketingReports, noOfConversionEvents)

		for key := range *attributionData {
			(*attributionData)[key].ConvAggFunctionType = convAggFunctionType
		}
	}
	return attributionData, isCompare, nil
}

func ProcessAttributionDataToResult(projectID int64, query *model.AttributionQuery,
	attributionData *map[string]*model.AttributionData, isCompare bool, queryStartTime int64,
	marketingReports *model.MarketingReports, kpiData map[string]model.KPIInfo, logCtx *log.Entry) *model.QueryResult {

	result := &model.QueryResult{}

	if query.AttributionKey == model.AttributionKeyLandingPage {

		result = model.ProcessQueryLandingPageUrl(query, attributionData, *logCtx, isCompare)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query Landing PageUrl took time")
		}
		queryStartTime = time.Now().UTC().Unix()

	} else if query.AnalyzeType == model.AnalyzeTypeHSDeals || query.AnalyzeType == model.AnalyzeTypeSFOpportunities {
		// execution similar to the normal run - still keeping it separate for better understanding
		result = model.ProcessQueryKPI(query, attributionData, marketingReports, isCompare, kpiData)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"result": result}).Info(fmt.Sprintf("KPI-Attribution result"))
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query KPI took time")
		}
		queryStartTime = time.Now().UTC().Unix()

	} else {
		result = model.ProcessQuery(query, attributionData, marketingReports, isCompare, projectID, *logCtx)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query Normal took time")
		}
		queryStartTime = time.Now().UTC().Unix()
	}

	return result
}

// pullConvertedUsers pulls converted users for the given Goal Event
func (store *MemSQL) GetConvertedUsers(projectID,
	conversionFrom, conversionTo int64, goalEvent model.QueryEventWithProperties,
	query *model.AttributionQuery, eventNameToIDList map[string][]interface{},
	logCtx log.Entry) (map[string]model.UserInfo, []model.UserEventInfo, map[string]int64, error) {

	goalEventName := goalEvent.Name
	goalEventProperties := goalEvent.Properties

	// 3. Fetch users who hit conversion event
	var userIDToInfoConverted map[string]model.UserInfo
	//var coalescedIDToInfoConverted map[string][]model.UserIDPropID
	var coalUserIdConversionTimestamp map[string]int64
	var usersToBeAttributed []model.UserEventInfo

	var err error
	// Fetch users who hit conversion event.
	userIDToInfoConverted, _, coalUserIdConversionTimestamp, err = store.GetConvertedUsersWithFilterV1(projectID,
		goalEventName, goalEventProperties, conversionFrom, conversionTo, eventNameToIDList, logCtx)
	if err != nil {
		return userIDToInfoConverted, usersToBeAttributed, coalUserIdConversionTimestamp, err
	}
	if projectID == 568 {
		_convertedUserCoalID := []string{}
		_convertedUserTimestamp := []int64{}
		for k, v := range coalUserIdConversionTimestamp {
			_convertedUserCoalID = append(_convertedUserCoalID, k)
			_convertedUserTimestamp = append(_convertedUserTimestamp, v)
		}
		logCtx.WithFields(log.Fields{"CleverTapConvertedUsersCoalIDInBatches": _convertedUserCoalID}).Warn("Printing Converted Users Coal ID")
		logCtx.WithFields(log.Fields{"CleverTapConvertedUsersTimeStampInBatches": _convertedUserTimestamp}).Warn("Printing Converted Users TimeStamp")

	}

	// Add users who hit conversion event
	for key, val := range coalUserIdConversionTimestamp {
		usersToBeAttributed = append(usersToBeAttributed, model.UserEventInfo{CoalUserID: key,
			EventName: goalEventName, Timestamp: val, EventType: 0})
	}

	err, linkedFunnelEventUsers := store.GetLinkedFunnelEventUsersFilter(projectID, conversionFrom, conversionTo, query.LinkedEvents, eventNameToIDList, userIDToInfoConverted, logCtx)
	if err != nil {
		return userIDToInfoConverted, usersToBeAttributed, coalUserIdConversionTimestamp, err
	}
	if projectID == 568 {
		_lfeUsers := []string{}
		_lfeTimeStamp := []int64{}

		for _, v := range linkedFunnelEventUsers {
			_lfeUsers = append(_lfeUsers, v.CoalUserID)
			_lfeTimeStamp = append(_lfeTimeStamp, v.Timestamp)
		}
		logCtx.WithFields(log.Fields{"CleverTapLinkedFunnelCoalId": _lfeUsers}).Warn("Printing Linked Funnel Event Users CoalId")
		logCtx.WithFields(log.Fields{"CleverTapLinkedFunnelTimeStamp": _lfeTimeStamp}).Warn("Printing Linked Funnel Event Users TimeStamp")

	}

	model.MergeUsersToBeAttributed(&usersToBeAttributed, linkedFunnelEventUsers)

	if projectID == 568 {
		_finalUsers := []string{}
		_finalTimeStamp := []int64{}
		_eventType := []int64{}
		for _, v := range usersToBeAttributed {
			_finalUsers = append(_finalUsers, v.CoalUserID)
			_finalTimeStamp = append(_finalTimeStamp, v.Timestamp)
			_eventType = append(_eventType, int64(v.EventType))
		}
		logCtx.WithFields(log.Fields{"CleverTapFinalUsersCoalID": _finalUsers}).Warn("Printing Final Users Coal ID")
		logCtx.WithFields(log.Fields{"CleverTapFinalUsersTimeStamp": _finalTimeStamp}).Warn("Printing Final Users TimeStamp")
		logCtx.WithFields(log.Fields{"CleverTapEventType": _eventType}).Warn("Printing Event Type")

	}

	return userIDToInfoConverted, usersToBeAttributed, coalUserIdConversionTimestamp, nil
}

func (store *MemSQL) FireAttributionV1(query *model.AttributionQuery,
	usersToBeAttributed *[]model.UserEventInfo, coalUserIdConversionTimestamp *map[string]int64, sessions map[string]map[string]model.UserSessionData,
	sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, bool, error) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	isCompare := false
	var err error

	var attributionData *map[string]*model.AttributionData
	if query.AttributionMethodologyCompare != "" {
		// Two AttributionMethodologies comparison
		isCompare = true
		attributionData, err = store.RunAttributionForMethodologyComparisonV1(query, usersToBeAttributed, coalUserIdConversionTimestamp, sessions, sessionWT, logCtx)

	} else if query.ConversionEventCompare.Name != "" {
		// Two events comparison
		isCompare = true
		attributionData, err = store.runAttributionV1(query.ConversionEvent, query,
			usersToBeAttributed, coalUserIdConversionTimestamp, sessions, sessionWT, logCtx)

		if err != nil {
			return nil, isCompare, err
		}
		// Running for ConversionEventCompare.
		attributionCompareData, err := store.runAttributionV1(query.ConversionEventCompare, query,
			usersToBeAttributed, coalUserIdConversionTimestamp, sessions, sessionWT, logCtx)

		if err != nil {
			return nil, isCompare, err
		}

		// Merge compare data into attributionData.
		for key := range *attributionData {
			if _, exists := (*attributionCompareData)[key]; exists {
				(*attributionData)[key].ConversionEventCompareCount = (*attributionCompareData)[key].ConversionEventCount
				(*attributionData)[key].ConversionEventCompareCountInfluence = (*attributionCompareData)[key].ConversionEventCountInfluence
			} else {
				(*attributionData)[key].ConversionEventCompareCount = []float64{float64(0)}
				(*attributionData)[key].ConversionEventCompareCountInfluence = []float64{float64(0)}
			}
		}
		// Filling any non-matched touch points.
		for missingKey := range *attributionCompareData {
			if _, exists := (*attributionData)[missingKey]; !exists {
				(*attributionData)[missingKey] = &model.AttributionData{}
				(*attributionData)[missingKey].ConversionEventCompareCount = (*attributionCompareData)[missingKey].ConversionEventCount
				(*attributionData)[missingKey].ConversionEventCompareCountInfluence = (*attributionCompareData)[missingKey].ConversionEventCountInfluence

			}
		}
	} else {

		// Single event attribution.
		attributionData, err = store.runAttributionV1(query.ConversionEvent, query,
			usersToBeAttributed, coalUserIdConversionTimestamp, sessions, sessionWT, logCtx)
	}
	return attributionData, isCompare, err
}

func (store *MemSQL) RunAttributionForMethodologyComparisonV1(query *model.AttributionQuery,
	usersToBeAttributed *[]model.UserEventInfo, coalUserIdConversionTimestamp *map[string]int64,
	sessions map[string]map[string]model.UserSessionData, sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, error) {

	// Attribution based on given attribution methodology.
	userConversionHit, _, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodology,
		query.ConversionEvent.Name, *usersToBeAttributed, sessions, *coalUserIdConversionTimestamp,
		query.LookbackDays, query.From, query.To, query.AttributionKey, logCtx)
	if err != nil {
		return nil, err
	}

	attributionData := model.AddUpConversionEventCount(userConversionHit, sessionWT)

	// Attribution based on given attributionMethodologyCompare methodology.
	userConversionCompareHit, _, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodologyCompare,
		query.ConversionEvent.Name, *usersToBeAttributed, sessions, *coalUserIdConversionTimestamp,
		query.LookbackDays, query.From, query.To, query.AttributionKey, logCtx)
	if err != nil {
		return nil, err
	}
	attributionDataCompare := model.AddUpConversionEventCount(userConversionCompareHit, sessionWT)

	// Merge compare data into attributionData.
	for key := range attributionData {
		if _, exists := attributionDataCompare[key]; exists {

			for len(attributionDataCompare[key].ConversionEventCount) < len(attributionData[key].ConversionEventCount) {
				attributionDataCompare[key].ConversionEventCount = append(attributionDataCompare[key].ConversionEventCount, float64(0))
				attributionDataCompare[key].ConversionEventCountInfluence = append(attributionDataCompare[key].ConversionEventCountInfluence, float64(0))
			}

			for idx := 0; idx < len(attributionDataCompare[key].ConversionEventCount); idx++ {
				attributionData[key].ConversionEventCompareCount = append(attributionData[key].ConversionEventCompareCount, attributionDataCompare[key].ConversionEventCount[idx])
				attributionData[key].ConversionEventCompareCountInfluence = append(attributionData[key].ConversionEventCompareCountInfluence, attributionDataCompare[key].ConversionEventCountInfluence[idx])
			}
		}
	}
	// filling any non-matched touch points
	for missingKey := range attributionDataCompare {
		if _, exists := attributionData[missingKey]; !exists {
			attributionData[missingKey] = &model.AttributionData{}
			attributionData[missingKey].ConversionEventCompareCount = attributionDataCompare[missingKey].ConversionEventCount
			attributionData[missingKey].ConversionEventCompareCountInfluence = attributionDataCompare[missingKey].ConversionEventCountInfluence
			for idx := 0; idx < len(attributionDataCompare[missingKey].ConversionEventCount); idx++ {
				attributionData[missingKey].ConversionEventCompareCount = append(attributionData[missingKey].ConversionEventCompareCount, float64(0))
				attributionData[missingKey].ConversionEventCompareCountInfluence = append(attributionData[missingKey].ConversionEventCompareCountInfluence, float64(0))
			}
		}
	}
	return &attributionData, nil
}

func (store *MemSQL) runAttributionV1(goalEvent model.QueryEventWithProperties,
	query *model.AttributionQuery, usersToBeAttributed *[]model.UserEventInfo, coalUserIdConversionTimestamp *map[string]int64,
	sessions map[string]map[string]model.UserSessionData,
	sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, error) {

	goalEventName := goalEvent.Name

	// 4. Apply attribution based on given attribution methodology
	userConversionHit, userLinkedFEHit, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodology,
		goalEventName, *usersToBeAttributed, sessions, *coalUserIdConversionTimestamp,
		query.LookbackDays, query.From, query.To, query.AttributionKey, logCtx)
	if err != nil {
		return nil, err
	}

	attributionData := make(map[string]*model.AttributionData)
	attributionData = model.AddUpConversionEventCount(userConversionHit, sessionWT)
	model.AddUpLinkedFunnelEventCount(query.LinkedEvents, attributionData, userLinkedFEHit)

	return &attributionData, nil
}

// Returns the all the sessions (userId,attributionId,minTimestamp,maxTimestamp) for given
// users from given period including lookback
func (store *MemSQL) getAllTheSessionsV1(projectId int64, sessionEventNameId string, query *model.AttributionQuery, usersToPullSessionFor []string,
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
		effectiveTo = model.LookbackAdjustedTo(query.To, query.LookbackDays, U.TimeZoneString(query.Timezone))
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
			caseSelectStmt + " AS landingPageUrl, "

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
			U.UP_INITIAL_PAGE_URL, model.PropertyValueNone, U.UP_INITIAL_PAGE_URL, model.PropertyValueNone, U.UP_INITIAL_PAGE_URL)

		wStmt, wParams, err := getSelectSQLStmtForContentGroup(contentGroupNamesToDummyNamesMap)
		if err != nil {
			return nil, nil, err
		}
		queryUserSessionTimeRange = queryUserSessionTimeRange + wStmt
		qParams = append(qParams, wParams...)

		queryUserSessionTimeRange = queryUserSessionTimeRange +
			" sessions.timestamp FROM events AS sessions " +
			" WHERE sessions.project_id=? AND sessions.event_name_id=? AND sessions.user_id IN (" + placeHolder + " ) AND sessions.timestamp BETWEEN ? AND ?"

		wParams = []interface{}{projectId, sessionEventNameId}
		wParams = append(wParams, value...)
		wParams = append(wParams, effectiveFrom, effectiveTo)
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

// FetchAllUsersAndCustomerUserDataInBatches returns usersIds for given list of customer_user_id (i.e. coal_id) in batches
func (store *MemSQL) FetchAllUsersAndCustomerUserDataInBatches(projectID int64, customerUserIdList []string, logCtx log.Entry) (map[string]string, map[string][]string, error) {

	if len(customerUserIdList) == 0 {
		return nil, nil, nil
	}

	userIdToCoalIds := make(map[string]string)
	custUserIdToUserIds := make(map[string][]string)

	coalUserIDsInBatches := U.GetStringListAsBatch(customerUserIdList, model.UserBatchSize)
	batch := 1
	for _, users := range coalUserIDsInBatches {

		placeHolder := U.GetValuePlaceHolder(len(users))
		value := U.GetInterfaceList(users)
		groupUserListQuery := "Select users.id, users.customer_user_id FROM users WHERE project_id=? " +
			" AND users.customer_user_id IN ( " + placeHolder + " ) "
		var gULParams []interface{}
		gULParams = append(gULParams, projectID)
		gULParams = append(gULParams, value...)
		gULRows, tx, err, reqID := store.ExecQueryWithContext(groupUserListQuery, gULParams)
		if err != nil {
			logCtx.WithError(err).Error("SQL Query failed")
			return nil, nil, errors.New("failed to get groupUserListQuery result for project")
		}
		batch++
		startReadTime := time.Now()
		for gULRows.Next() {
			var userIDNull sql.NullString
			var custUserIDNull sql.NullString
			if err = gULRows.Scan(&userIDNull, &custUserIDNull); err != nil {
				logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
				continue
			}

			userID := U.IfThenElse(userIDNull.Valid, userIDNull.String, model.PropertyValueNone).(string)
			custUserID := U.IfThenElse(custUserIDNull.Valid, custUserIDNull.String, model.PropertyValueNone).(string)
			if userID == model.PropertyValueNone || custUserID == model.PropertyValueNone {
				logCtx.WithError(err).Error("Values are not correct - userID & custUserID . Ignoring row. Continuing")
				continue
			}

			// Keeping userID to CoalID
			userIdToCoalIds[userID] = custUserID

			// Keeping CoalID to userID(s). Since many user_ids can be associated with one coal_id
			if _, exists := custUserIdToUserIds[custUserID]; exists {
				v := custUserIdToUserIds[custUserID]
				v = append(v, userID)
				custUserIdToUserIds[custUserID] = v
			} else {
				var users []string
				users = append(users, userID)
				custUserIdToUserIds[custUserID] = users
			}
		}

		err = gULRows.Err()
		if err != nil {
			// Error from DB is captured eg: timeout error
			logCtx.WithFields(log.Fields{"err": err, "batchNo": batch}).Error("Error in executing query in FetchAllUsersAndCustomerUserDataInBatches")
			return nil, nil, err
		}
		U.CloseReadQuery(gULRows, tx)
		U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &log.Fields{"project_id": projectID})
	}
	return userIdToCoalIds, custUserIdToUserIds, nil
}

// FetchAllUsersAndCustomerUserData returns usersIds for given list of customer_user_id (i.e. coal_id)
// @Deprecated
func (store *MemSQL) FetchAllUsersAndCustomerUserData(projectID int64, customerUserIdList []string, logCtx log.Entry) (map[string]string, map[string][]string, error) {

	if len(customerUserIdList) == 0 {
		return nil, nil, nil
	}
	userIdToCoalIds := make(map[string]string)
	custUserIdToUserIds := make(map[string][]string)

	custUserIDPlaceHolder := U.GetValuePlaceHolder(len(customerUserIdList))
	custUserIDs := U.GetInterfaceList(customerUserIdList)
	groupUserListQuery := "Select users.id, users.customer_user_id FROM users WHERE project_id=? " +
		" AND users.customer_user_id IN ( " + custUserIDPlaceHolder + " ) "
	var gULParams []interface{}
	gULParams = append(gULParams, projectID)
	gULParams = append(gULParams, custUserIDs...)
	gULRows, tx2, err, reqID := store.ExecQueryWithContext(groupUserListQuery, gULParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, errors.New("failed to get groupUserListQuery result for project")
	}

	startReadTime := time.Now()
	for gULRows.Next() {
		var userIDNull sql.NullString
		var custUserIDNull sql.NullString
		if err = gULRows.Scan(&userIDNull, &custUserIDNull); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
			continue
		}

		userID := U.IfThenElse(userIDNull.Valid, userIDNull.String, model.PropertyValueNone).(string)
		custUserID := U.IfThenElse(custUserIDNull.Valid, custUserIDNull.String, model.PropertyValueNone).(string)
		if userID == model.PropertyValueNone || custUserID == model.PropertyValueNone {
			logCtx.WithError(err).Error("Values are not correct - userID & custUserID . Ignoring row. Continuing")
			continue
		}

		// Keeping userID to CoalID
		userIdToCoalIds[userID] = custUserID

		// Keeping CoalID to userID(s). Since many user_ids can be associated with one coal_id
		if _, exists := custUserIdToUserIds[custUserID]; exists {
			v := custUserIdToUserIds[custUserID]
			v = append(v, userID)
			custUserIdToUserIds[custUserID] = v
		} else {
			var users []string
			users = append(users, userID)
			custUserIdToUserIds[custUserID] = users
		}
	}
	err = gULRows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in Fetch All Users")
		return nil, nil, err
	}

	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &log.Fields{})
	defer U.CloseReadQuery(gULRows, tx2)
	return userIdToCoalIds, custUserIdToUserIds, nil
}

// GetConvertedUsersWithFilterV1 Returns the list of eligible users who hit conversion
// event for userProperties from events table
func (store *MemSQL) GetConvertedUsersWithFilterV1(projectID int64, goalEventName string,
	goalEventProperties []model.QueryProperty, conversionFrom, conversionTo int64,
	eventNameToIdList map[string][]interface{}, logCtx log.Entry) (map[string]model.UserInfo,
	map[string][]model.UserIDPropID, map[string]int64, error) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	conversionEventNameIDs := eventNameToIdList[goalEventName]
	placeHolder := "?"
	for i := 0; i < len(conversionEventNameIDs)-1; i++ {
		placeHolder += ",?"
	}

	selectEventHits := "SELECT events.user_id, events.timestamp FROM events"
	whereEventHits := "WHERE events.project_id=? AND timestamp >= ? AND " +
		" timestamp <=? AND events.event_name_id IN (" + placeHolder + ") "
	qParams := []interface{}{projectID, conversionFrom, conversionTo}
	qParams = append(qParams, conversionEventNameIDs...)

	// add event filter
	wStmtEvent, wParamsEvent, eventJoinStmnt, err := getFilterSQLStmtForEventProperties(
		projectID, goalEventProperties, conversionFrom) // query.ConversionEvent.Properties)
	if err != nil {
		return nil, nil, nil, err
	}

	if wStmtEvent != "" {
		whereEventHits = whereEventHits + " AND " + fmt.Sprintf("( %s )", wStmtEvent)
		qParams = append(qParams, wParamsEvent...)
	}

	// add user filter
	wStmtUser, wParamsUser, eventUserJoinStmnt, err := getFilterSQLStmtForUserProperties(projectID,
		goalEventProperties, conversionFrom) // query.ConversionEvent.Properties)
	if err != nil {
		return nil, nil, nil, err
	}
	if wStmtUser != "" {
		whereEventHits = whereEventHits + " AND " + fmt.Sprintf("( %s )", wStmtUser)
		qParams = append(qParams, wParamsUser...)
	}

	// JOIN events_properties_json table, if there is
	// filter on event_properties or event_user_properties.
	if eventJoinStmnt == "" {
		eventJoinStmnt = eventUserJoinStmnt
	}

	queryEventHits := selectEventHits + " " + eventJoinStmnt + " " + whereEventHits
	if projectID == 568 {
		queryEventHits1, qParams1 := model.ExpandArrayWithIndividualValues(queryEventHits, qParams)
		logCtx.WithFields(log.Fields{"CleverTapQueryGetConvertedUsersWithFilterV1": U.DBDebugPreparedStatement(C.GetConfig().Env, queryEventHits1, qParams1)}).Info("Printing Query")
	}

	// fetch query results
	rows, tx, err, reqID := store.ExecQueryWithContext(queryEventHits, qParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed for queryEventHits")
		return nil, nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)
	var userIDList []string
	userIdHitGoalEventTimestamp := make(map[string]int64)
	startReadTime := time.Now()
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
		} else {
			// record the fist occurrence of the event by userID
			if timestamp < userIdHitGoalEventTimestamp[userID] {
				userIdHitGoalEventTimestamp[userID] = timestamp
			}
		}
	}
	err = rows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in GetConvertedUsersWithFilterV1")
		return nil, nil, nil, err
	}
	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &log.Fields{"project_id": projectID})

	// Get coalesced Id for converted user_ids (without filter)
	userIDToCoalIDInfo, coalIDs, err := store.GetCoalesceIDFromUserIDs(userIDList, projectID, logCtx)
	if err != nil {
		return nil, nil, nil, err
	}
	// Reverse lookup for all the converted userID's coalIDs to get the other users which are not marked 'converted'
	_userIDToCoalID, _custUserIdToUserIds, err := store.FetchAllUsersAndCustomerUserDataInBatches(projectID, coalIDs, logCtx)
	if err != nil {
		return nil, nil, nil, err
	}
	for userID, coalID := range _userIDToCoalID {

		if _, exists := userIDToCoalIDInfo[userID]; exists {
			continue
		}
		// userID was not considered for conversion, let's add it with other userIDs of same user
		sameUsers := _custUserIdToUserIds[coalID]
		for _, uID := range sameUsers {
			if _, exists := userIDToCoalIDInfo[uID]; exists {
				// adding userID with one of the data from same user
				userIDToCoalIDInfo[userID] = model.UserInfo{
					CoalUserID: userIDToCoalIDInfo[uID].CoalUserID,
					Timestamp:  userIDToCoalIDInfo[uID].Timestamp,
				}
				// add the user hit timing
				userIdHitGoalEventTimestamp[userID] = userIdHitGoalEventTimestamp[uID]
				break
			}
		}
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

		filteredCoalIDToUserIDInfo[coalUserID] = append(filteredCoalIDToUserIDInfo[coalUserID], model.UserIDPropID{UserID: userID, Timestamp: timestamp})
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

	if projectID == 1125899918000010 {
		logCtx.WithFields(log.Fields{"HevoDebug": "Hevo", "filteredUserIdToUserIDInfo": filteredUserIdToUserIDInfo}).Info("Debug GetConvertedUsersWithFilterV1")
		logCtx.WithFields(log.Fields{"HevoDebug": "Hevo", "filteredCoalIDToUserIDInfo": filteredCoalIDToUserIDInfo}).Info("Debug GetConvertedUsersWithFilterV1")
	}
	return filteredUserIdToUserIDInfo, filteredCoalIDToUserIDInfo, coalUserIdConversionTimestamp, nil
}
