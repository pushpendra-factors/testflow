package memsql

import (
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

func (store *MemSQL) PullSessionsOfConvertedUsers(projectID int64, query *model.AttributionQuery, sessionEventNameID string, usersToBeAttributed []string,
	marketingReports *model.MarketingReports, contentGroupNamesList []string, logCtx *log.Entry) (map[string]map[string]model.UserSessionData, error) {

	sessions := make(map[string]map[string]model.UserSessionData)

	queryStartTime := time.Now().UTC().Unix()
	// Pull Sessions for the cases: "Tactic" and "TacticOffer".
	// If Landing Page level report, pull for offer as well.
	if query.TacticOfferType != model.MarketingEventTypeOffer || query.AttributionKey == model.AttributionKeyLandingPage {
		// Get all the sessions (userId, attributionId, UserSessionData) for given period by attribution key
		_sessions, sessionUsers, err := store.getAllTheSessionsV1(projectID, sessionEventNameID, query, usersToBeAttributed, marketingReports, contentGroupNamesList, *logCtx)
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Pull Sessions data data took time")
		queryStartTime = time.Now().UTC().Unix()

		if err != nil {
			return nil, err
		}

		usersInfo, err := store.GetCoalesceIDFromUserIDs(sessionUsers, projectID, *logCtx)
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Get Coalesce user data took time")
		queryStartTime = time.Now().UTC().Unix()
		if err != nil {
			return nil, err
		}

		model.UpdateSessionsMapWithCoalesceID(_sessions, usersInfo, &sessions)
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Update Sessions Coalesce user data took time")
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
		_userIDToInfoConverted, usersToBeAttributed, coalUserIdConversionTimestamp, err = store.pullConvertedUsers(projectID,
			conversionFrom, conversionTo, query.ConversionEvent,
			query, eventNameToIDList, *logCtx)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		// Get user IDs for AnalyzeTypeUsers
		for id, _ := range _userIDToInfoConverted {
			usersIDsToAttribute = append(usersIDsToAttribute, id)
		}
		logCtx.WithFields(log.Fields{"UniqueUsers": len(usersIDsToAttribute)}).Info("Total users for the attribution query")
	} else if query.AnalyzeType == model.AnalyzeTypeUserKPI {

		var err error
		queryStartTime := time.Now().UTC().Unix()
		kpiData, err = store.ExecuteUserKPIForAttribution(projectID, query, debugQueryKey,
			*logCtx, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("UserKPI query execution took time")
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
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("KPI query execution took time")
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

		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("FireAttribution took time")
		queryStartTime = time.Now().UTC().Unix()

		logCtx.Info("Done FireAttribution")
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
		logCtx.Info("Done AddTheAddedKeysAndMetrics, AddPerformanceData")

	} else {

		// creating group sessions by transforming sessions
		groupSessions := make(map[string]map[string]model.UserSessionData)

		for kpiID, kpiInfo := range kpiData {

			if _, exists := groupSessions[kpiID]; !exists {
				groupSessions[kpiID] = make(map[string]model.UserSessionData)
			}
			if kpiInfo.KpiCoalUserIds == nil || len(kpiInfo.KpiCoalUserIds) == 0 {
				logCtx.WithFields(log.Fields{"KpiInfo": kpiInfo, "KPI_ID": kpiID}).Info("no user found for the KPI group, ignoring")
				//groupSessions[kpiID][noneKey] = model.UserSessionData{}
				continue
			}
			for _, user := range kpiInfo.KpiCoalUserIds {
				// check if user has session/otp
				if _, exists := sessions[user]; !exists {
					logCtx.WithFields(log.Fields{"User": user, "KPI_ID": kpiID}).Info("user without session/otp")
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
					logCtx.WithFields(log.Fields{"User": user, "KPI_ID": kpiID}).Info("user without session/otp")
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
		logCtx.WithFields(log.Fields{"KPIGroupSession": groupSessions}).Info("KPI-Attribution Group session 2")

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
			logCtx.Info("no user journey found (neither sessions nor offline touch points)")
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
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("FireAttribution KPI took time")
		queryStartTime = time.Now().UTC().Unix()
		logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData")

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
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query Landing PageUrl took time")
		queryStartTime = time.Now().UTC().Unix()

	} else if query.AnalyzeType == model.AnalyzeTypeHSDeals || query.AnalyzeType == model.AnalyzeTypeSFOpportunities {
		// execution similar to the normal run - still keeping it separate for better understanding
		result = model.ProcessQueryKPI(query, attributionData, marketingReports, isCompare, kpiData)
		logCtx.WithFields(log.Fields{"result": result}).Info(fmt.Sprintf("KPI-Attribution result"))
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query KPI took time")
		queryStartTime = time.Now().UTC().Unix()

	} else {
		result = model.ProcessQuery(query, attributionData, marketingReports, isCompare, projectID, *logCtx)
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query Normal took time")
		queryStartTime = time.Now().UTC().Unix()
	}

	return result
}

// pullConvertedUsers pulls converted users for the given Goal Event
func (store *MemSQL) pullConvertedUsers(projectID,
	conversionFrom, conversionTo int64, goalEvent model.QueryEventWithProperties,
	query *model.AttributionQuery, eventNameToIDList map[string][]interface{},
	logCtx log.Entry) (map[string]model.UserInfo, []model.UserEventInfo, map[string]int64, error) {

	goalEventName := goalEvent.Name
	goalEventProperties := goalEvent.Properties

	// 3. Fetch users who hit conversion event
	var userIDToInfoConverted map[string]model.UserInfo
	var coalescedIDToInfoConverted map[string][]model.UserIDPropID
	var coalUserIdConversionTimestamp map[string]int64
	var usersToBeAttributed []model.UserEventInfo

	var err error
	// Fetch users who hit conversion event.
	userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err = store.GetConvertedUsersWithFilter(projectID,
		goalEventName, goalEventProperties, conversionFrom, conversionTo, eventNameToIDList, logCtx)
	if err != nil {
		return userIDToInfoConverted, usersToBeAttributed, coalUserIdConversionTimestamp, err
	}
	if projectID == 568 {
		logCtx.WithFields(log.Fields{"CleverTapUserIDToInfoConverted": userIDToInfoConverted}).Info("Printing Conversion goal userIDToInfoConverted")
		logCtx.WithFields(log.Fields{"CleverTapCoalescedIDToInfoConverted": coalescedIDToInfoConverted}).Info("Printing Conversion goal coalescedIDToInfoConverted")
		logCtx.WithFields(log.Fields{"CleverTapCoalUserIdConversionTimestamp": coalUserIdConversionTimestamp}).Info("Printing Conversion goal coalUserIdConversionTimestamp")
	}

	// Add users who hit conversion event
	for key := range coalescedIDToInfoConverted {
		usersToBeAttributed = append(usersToBeAttributed, model.UserEventInfo{CoalUserID: key,
			EventName: goalEventName})
	}

	err, linkedFunnelEventUsers := store.GetLinkedFunnelEventUsersFilter(projectID, conversionFrom, conversionTo, query.LinkedEvents, eventNameToIDList, userIDToInfoConverted, logCtx)
	if err != nil {
		return userIDToInfoConverted, usersToBeAttributed, coalUserIdConversionTimestamp, err
	}
	if projectID == 568 {
		logCtx.WithFields(log.Fields{"CleverTapLinkedFunnelUsers": linkedFunnelEventUsers}).Info("Printing Linked Funnel Event Users")
	}

	model.MergeUsersToBeAttributed(&usersToBeAttributed, linkedFunnelEventUsers)

	if projectID == 568 {
		logCtx.WithFields(log.Fields{"CleverTapUserIDToInfoConvertedFinal": userIDToInfoConverted}).Info("Printing Final userIDToInfoConverted")
		logCtx.WithFields(log.Fields{"CleverTapUsersToBeAttributedFinal ": usersToBeAttributed}).Info("Printing Final usersToBeAttributed")
		logCtx.WithFields(log.Fields{"CleverTapCoalUserIdConversionTimestampFinal": coalUserIdConversionTimestamp}).Info("Printing Final coalUserIdConversionTimestamp")
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
		query.LookbackDays, query.From, query.To, query.AttributionKey)
	if err != nil {
		return nil, err
	}

	attributionData := model.AddUpConversionEventCount(userConversionHit, sessionWT)

	// Attribution based on given attributionMethodologyCompare methodology.
	userConversionCompareHit, _, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodologyCompare,
		query.ConversionEvent.Name, *usersToBeAttributed, sessions, *coalUserIdConversionTimestamp,
		query.LookbackDays, query.From, query.To, query.AttributionKey)
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
		query.LookbackDays, query.From, query.To, query.AttributionKey)
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
		logCtx.Info("Attribution before ProcessEventRows")
		processErr := model.ProcessEventRows(rows, query, reports, contentGroupNamesList, &attributedSessionsByUserId, &userIdsWithSession, logCtx, reqID)
		U.CloseReadQuery(rows, tx)
		if processErr != nil {
			return attributedSessionsByUserId, userIdsWithSession, processErr
		}
	}

	return attributedSessionsByUserId, userIdsWithSession, nil
}
