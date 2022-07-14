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

// ExecuteAttributionQuery Executes the Attribution using following steps:
//	1. Get all the sessions data (userId, attributionId, timestamp) for given period by attribution key
// 	2. Add the website visitor info using session data from step 1
//	3. i) 	Find out users who hit conversion event applying filter
//	  ii)	Using users from 3.i) find out users who hit linked funnel event applying filter
//	4. Apply attribution methodology
//	5. Add performance data by attributionId
func (store *MemSQL) ExecuteAttributionQuery(projectID int64, queryOriginal *model.AttributionQuery,
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

	queryStartORG := time.Now().UTC().Unix()
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
	sessions := make(map[string]map[string]model.UserSessionData)

	// Pull Sessions for the cases: "Tactic" and "TacticOffer".
	// If Landing Page level report, pull for offer as well.
	if query.TacticOfferType != model.MarketingEventTypeOffer || query.AttributionKey == model.AttributionKeyLandingPage {
		// Get all the sessions (userId, attributionId, UserSessionData) for given period by attribution key
		_sessions, sessionUsers, err := store.getAllTheSessions(projectID, sessionEventNameID, query, marketingReports, contentGroupNamesList, *logCtx)
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

	// Pull Offline touch points for all the cases: "Tactic",  "Offer", "TacticOffer"
	store.AppendOTPSessions(projectID, query, &sessions, *logCtx)
	logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Pull Offline touch points user data took time")
	queryStartTime = time.Now().UTC().Unix()

	isCompare := false
	var attributionData *map[string]*model.AttributionData
	var kpiData map[string]model.KPIInfo
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

		attributionData, isCompare, err = store.FireAttribution(projectID, query, eventNameToIDList, sessions, sessionWT, *logCtx)

		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("FireAttribution took time")
		queryStartTime = time.Now().UTC().Unix()

		logCtx.Info("Done FireAttribution")
		if err != nil {
			return nil, err
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
		// This thread is for query.AnalyzeType == model.AnalyzeTypeHSDeals || query.AnalyzeType == model.AnalyzeTypeSFOpportunities.
		kpiData, _, _, err = store.ExecuteKPIForAttribution(projectID, query, debugQueryKey,
			*logCtx, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("KPI query execution took time")
		queryStartTime = time.Now().UTC().Unix()
		if err != nil {
			return nil, err
		}

		log.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData}).Info("KPI Attribution kpiData")
		/*emptyMarketingValue:= model.MarketingData{Channel: model.PropertyValueNone,
			CampaignID: model.PropertyValueNone, CampaignName: model.PropertyValueNone, AdgroupID: model.PropertyValueNone,
			AdgroupName: model.PropertyValueNone, KeywordName: model.PropertyValueNone, KeywordMatchType: model.PropertyValueNone,
			Source: model.PropertyValueNone, ChannelGroup: model.PropertyValueNone, LandingPageUrl: model.PropertyValueNone, ContentGroupValuesMap: nil}
		noneKey := model.GetMarketingDataKey(query.AttributionKey, emptyMarketingValue)*/

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
			return nil, errors.New("no user journey found (neither sessions nor offline touch points)")
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

		attributionData, isCompare, err = store.FireAttributionForKPI(projectID, query, groupSessions, kpiData, sessionWT, *logCtx)
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("FireAttribution KPI took time")
		queryStartTime = time.Now().UTC().Unix()
		logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData")

		if err != nil {
			return nil, err
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

	// Filter out the key values from query (apply filter after performance enrichment)
	model.ApplyFilter(attributionData, query)
	logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Metrics, Performance report, filter took time")
	queryStartTime = time.Now().UTC().Unix()

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
	result.Meta.Currency = ""
	if projectSetting.IntAdwordsCustomerAccountId != nil && *projectSetting.IntAdwordsCustomerAccountId != "" {
		currency, _ := store.GetAdwordsCurrency(projectID, *projectSetting.IntAdwordsCustomerAccountId, query.From, query.To, *logCtx)
		result.Meta.Currency = currency
	}
	logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartORG) / 60}).Info("Total query took time")
	queryStartTime = time.Now().UTC().Unix()

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

	usersInfoOTP, err := store.GetCoalesceIDFromUserIDs(sessionOTPUsers, projectID, logCtx)
	if err != nil {
		logCtx.Info("no users found for OTP events/sessions found. Skipping computation")
		return
	}

	model.UpdateSessionsMapWithCoalesceID(_sessionsOTP, usersInfoOTP, sessions)
}

func (store *MemSQL) FireAttribution(projectID int64, query *model.AttributionQuery, eventNameToIDList map[string][]interface{},
	sessions map[string]map[string]model.UserSessionData, sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, bool, error) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	isCompare := false
	var err error
	// Default conversion for AttributionQueryTypeConversionBased.
	conversionFrom := query.From
	conversionTo := query.To
	// Extend the campaign window for engagement based attribution.
	if query.QueryType == model.AttributionQueryTypeEngagementBased {
		conversionFrom = query.From
		conversionTo = model.LookbackAdjustedTo(query.To, query.LookbackDays)
	}
	var attributionData *map[string]*model.AttributionData
	if query.AttributionMethodologyCompare != "" {
		// Two AttributionMethodologies comparison
		isCompare = true
		attributionData, err = store.RunAttributionForMethodologyComparison(projectID,
			conversionFrom, conversionTo, query, eventNameToIDList, sessions, sessionWT, logCtx)

	} else if query.ConversionEventCompare.Name != "" {
		// Two events comparison
		isCompare = true
		attributionData, err = store.runAttribution(projectID,
			conversionFrom, conversionTo, query.ConversionEvent, query, eventNameToIDList, sessions, sessionWT, logCtx)

		if err != nil {
			return nil, isCompare, err
		}
		// Running for ConversionEventCompare.
		attributionCompareData, err := store.runAttribution(projectID,
			conversionFrom, conversionTo, query.ConversionEventCompare, query, eventNameToIDList, sessions, sessionWT, logCtx)

		if err != nil {
			return nil, isCompare, err
		}

		// Merge compare data into attributionData.
		for key := range *attributionData {
			if _, exists := (*attributionCompareData)[key]; exists {
				(*attributionData)[key].ConversionEventCompareCount = (*attributionCompareData)[key].ConversionEventCount
			} else {
				(*attributionData)[key].ConversionEventCompareCount = []float64{float64(0)}
			}
		}
		// Filling any non-matched touch points.
		for missingKey := range *attributionCompareData {
			if _, exists := (*attributionData)[missingKey]; !exists {
				(*attributionData)[missingKey] = &model.AttributionData{}
				(*attributionData)[missingKey].ConversionEventCompareCount = (*attributionCompareData)[missingKey].ConversionEventCount
			}
		}
	} else {
		// Single event attribution.
		attributionData, err = store.runAttribution(projectID,
			conversionFrom, conversionTo, query.ConversionEvent,
			query, eventNameToIDList, sessions, sessionWT, logCtx)
	}
	return attributionData, isCompare, err
}

func (store *MemSQL) RunAttributionForMethodologyComparison(projectID int64,
	conversionFrom, conversionTo int64, query *model.AttributionQuery, eventNameToIDList map[string][]interface{},
	sessions map[string]map[string]model.UserSessionData, sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, error) {

	// Empty linkedEvents as they are not analyzed in compare events.
	var linkedEvents []model.QueryEventWithProperties

	var userIDToInfoConverted map[string]model.UserInfo
	var coalescedIDToInfoConverted map[string][]model.UserIDPropID
	var coalUserIdConversionTimestamp map[string]int64
	var err error
	// Fetch users who hit conversion event.
	userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err = store.GetConvertedUsersWithFilter(projectID,
		query.ConversionEvent.Name, query.ConversionEvent.Properties, conversionFrom, conversionTo,
		eventNameToIDList, logCtx)
	if err != nil {
		return nil, err
	}

	// Add users who hit conversion event.
	var usersToBeAttributed []model.UserEventInfo
	for key := range coalescedIDToInfoConverted {
		usersToBeAttributed = append(usersToBeAttributed, model.UserEventInfo{CoalUserID: key,
			EventName: query.ConversionEvent.Name})
	}

	err, linkedFunnelEventUsers := store.GetLinkedFunnelEventUsersFilter(projectID, conversionFrom, conversionTo,
		linkedEvents, eventNameToIDList, userIDToInfoConverted, logCtx)
	if err != nil {
		return nil, err
	}
	model.MergeUsersToBeAttributed(&usersToBeAttributed, linkedFunnelEventUsers)

	// Attribution based on given attribution methodology.
	userConversionHit, _, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodology,
		query.ConversionEvent.Name, usersToBeAttributed, sessions, coalUserIdConversionTimestamp,
		query.LookbackDays, query.From, query.To, query.AttributionKey)
	if err != nil {
		return nil, err
	}

	attributionData := model.AddUpConversionEventCount(userConversionHit, sessionWT)

	// Attribution based on given attributionMethodologyCompare methodology.
	userConversionCompareHit, _, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodologyCompare,
		query.ConversionEvent.Name, usersToBeAttributed, sessions, coalUserIdConversionTimestamp,
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
			}

			for idx := 0; idx < len(attributionDataCompare[key].ConversionEventCount); idx++ {
				attributionData[key].ConversionEventCompareCount = append(attributionData[key].ConversionEventCompareCount, attributionDataCompare[key].ConversionEventCount[idx])
			}
		}
	}
	// filling any non-matched touch points
	for missingKey := range attributionDataCompare {
		if _, exists := attributionData[missingKey]; !exists {
			attributionData[missingKey] = &model.AttributionData{}
			attributionData[missingKey].ConversionEventCompareCount = attributionDataCompare[missingKey].ConversionEventCount
			for idx := 0; idx < len(attributionDataCompare[missingKey].ConversionEventCount); idx++ {
				attributionData[missingKey].ConversionEventCompareCount = append(attributionData[missingKey].ConversionEventCompareCount, float64(0))
			}
		}
	}
	return &attributionData, nil
}

func (store *MemSQL) runAttributionKpi(projectID int64,
	conversionFrom, conversionTo int64, goalEvent model.QueryEventWithProperties,
	query *model.AttributionQuery, eventNameToIDList map[string][]interface{},
	sessions map[string]map[string]model.UserSessionData, sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, error) {

	goalEventName := goalEvent.Name
	goalEventProperties := goalEvent.Properties

	// 3. Fetch users who hit conversion event
	var userIDToInfoConverted map[string]model.UserInfo
	var coalescedIDToInfoConverted map[string][]model.UserIDPropID
	var coalUserIdConversionTimestamp map[string]int64
	var err error
	// Fetch users who hit conversion event.
	userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err = store.GetConvertedUsersWithFilter(projectID,
		goalEventName, goalEventProperties, conversionFrom, conversionTo, eventNameToIDList, logCtx)
	if err != nil {
		return nil, err
	}

	// Add users who hit conversion event
	var usersToBeAttributed []model.UserEventInfo
	for key := range coalescedIDToInfoConverted {
		usersToBeAttributed = append(usersToBeAttributed, model.UserEventInfo{CoalUserID: key,
			EventName: goalEventName})
	}

	err, linkedFunnelEventUsers := store.GetLinkedFunnelEventUsersFilter(projectID, conversionFrom, conversionTo, query.LinkedEvents, eventNameToIDList, userIDToInfoConverted, logCtx)
	if err != nil {
		return nil, err
	}

	model.MergeUsersToBeAttributed(&usersToBeAttributed, linkedFunnelEventUsers)

	// 4. Apply attribution based on given attribution methodology
	userConversionHit, userLinkedFEHit, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodology,
		goalEventName, usersToBeAttributed, sessions, coalUserIdConversionTimestamp,
		query.LookbackDays, query.From, query.To, query.AttributionKey)
	if err != nil {
		return nil, err
	}

	attributionData := make(map[string]*model.AttributionData)
	attributionData = model.AddUpConversionEventCount(userConversionHit, sessionWT)
	model.AddUpLinkedFunnelEventCount(query.LinkedEvents, attributionData, userLinkedFEHit)
	return &attributionData, nil
}

func (store *MemSQL) runAttribution(projectID int64,
	conversionFrom, conversionTo int64, goalEvent model.QueryEventWithProperties,
	query *model.AttributionQuery, eventNameToIDList map[string][]interface{},
	sessions map[string]map[string]model.UserSessionData, sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, error) {

	goalEventName := goalEvent.Name
	goalEventProperties := goalEvent.Properties

	// 3. Fetch users who hit conversion event
	var userIDToInfoConverted map[string]model.UserInfo
	var coalescedIDToInfoConverted map[string][]model.UserIDPropID
	var coalUserIdConversionTimestamp map[string]int64
	var err error
	// Fetch users who hit conversion event.
	userIDToInfoConverted, coalescedIDToInfoConverted, coalUserIdConversionTimestamp, err = store.GetConvertedUsersWithFilter(projectID,
		goalEventName, goalEventProperties, conversionFrom, conversionTo, eventNameToIDList, logCtx)
	if err != nil {
		return nil, err
	}

	// Add users who hit conversion event
	var usersToBeAttributed []model.UserEventInfo
	for key := range coalescedIDToInfoConverted {
		usersToBeAttributed = append(usersToBeAttributed, model.UserEventInfo{CoalUserID: key,
			EventName: goalEventName})
	}

	err, linkedFunnelEventUsers := store.GetLinkedFunnelEventUsersFilter(projectID, conversionFrom, conversionTo, query.LinkedEvents, eventNameToIDList, userIDToInfoConverted, logCtx)
	if err != nil {
		return nil, err
	}

	model.MergeUsersToBeAttributed(&usersToBeAttributed, linkedFunnelEventUsers)

	// 4. Apply attribution based on given attribution methodology
	userConversionHit, userLinkedFEHit, err := model.ApplyAttribution(query.QueryType, query.AttributionMethodology,
		goalEventName, usersToBeAttributed, sessions, coalUserIdConversionTimestamp,
		query.LookbackDays, query.From, query.To, query.AttributionKey)
	if err != nil {
		return nil, err
	}

	attributionData := make(map[string]*model.AttributionData)
	attributionData = model.AddUpConversionEventCount(userConversionHit, sessionWT)
	model.AddUpLinkedFunnelEventCount(query.LinkedEvents, attributionData, userLinkedFEHit)
	return &attributionData, nil
}

// GetCoalesceIDFromUserIDs returns the map of coalesce userId for given list of users
func (store *MemSQL) GetCoalesceIDFromUserIDs(userIDs []string, projectID int64,
	logCtx log.Entry) (map[string]model.UserInfo, error) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	userIDsInBatches := U.GetStringListAsBatch(userIDs, model.UserBatchSize)
	userIDToCoalUserIDMap := make(map[string]model.UserInfo)
	batch := 1
	for _, users := range userIDsInBatches {
		placeHolder := U.GetValuePlaceHolder(len(users))
		value := U.GetInterfaceList(users)
		queryUserIDCoalID := "SELECT id, COALESCE(users.customer_user_id,users.id) AS coal_user_id" + " " +
			"FROM users WHERE id IN (" + placeHolder + ")"
		rows, tx, err, reqID := store.ExecQueryWithContext(queryUserIDCoalID, value)
		if err != nil {
			logCtx.WithError(err).Error("SQL Query failed for GetCoalesceIDFromUserIDs")
			return nil, err
		}
		logCtx.WithFields(log.Fields{"Batch": batch}).Info("Executing GetCoalesceIDFromUserIDs")
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
		U.CloseReadQuery(rows, tx)
		U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &log.Fields{"project_id": projectID})
	}
	logCtx.WithFields(log.Fields{"user_count": len(userIDToCoalUserIDMap)}).Info("GetCoalesceIDFromUserIDs 3")
	return userIDToCoalUserIDMap, nil
}

// Returns the all the sessions (userId,attributionId,minTimestamp,maxTimestamp) for given
// users from given period including lookback
func (store *MemSQL) getAllTheSessions(projectId int64, sessionEventNameId string, query *model.AttributionQuery, reports *model.MarketingReports, contentGroupNamesList []string, logCtx log.Entry) (map[string]map[string]model.UserSessionData, []string, error) {
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
	contentGroupNamesToDummyNamesMap := model.GetContentGroupNamesToDummyNamesMap(contentGroupNamesList)
	caseSelectStmt := "CASE WHEN JSON_EXTRACT_STRING(sessions.properties, ?) IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(sessions.properties, ?) = '' THEN ? ELSE JSON_EXTRACT_STRING(sessions.properties, ?) END"

	queryUserSessionTimeRange := "SELECT sessions.user_id, " +
		caseSelectStmt + " AS sessionTimeSpent, " +
		caseSelectStmt + " AS pageCount, " +
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
		U.SP_SPENT_TIME, 0, U.SP_SPENT_TIME, 0, U.SP_SPENT_TIME,
		U.SP_PAGE_COUNT, 0, U.SP_PAGE_COUNT, 0, U.SP_PAGE_COUNT,
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
		" WHERE sessions.project_id=? AND sessions.event_name_id=? AND sessions.timestamp BETWEEN ? AND ?"

	wParams = []interface{}{projectId, sessionEventNameId, effectiveFrom, effectiveTo}
	qParams = append(qParams, wParams...)
	rows, tx, err, reqID := store.ExecQueryWithContext(queryUserSessionTimeRange, qParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)
	logCtx.Info("Attribution before ProcessEventRows")
	return model.ProcessEventRows(rows, query, reports, contentGroupNamesList, logCtx, reqID)
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
		logCtx.Info("failed to find offline touch point event names, skipping OTP attribution computation")
		return model.EventName{}, errors.New("failed to find offline touch point event names, skipping OTP attribution computation")
	}
	return eventNames[0], nil
}

// Return conversion event Id, list of all event_ids(Conversion and funnel events) and a Id to name mapping
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
			wStmtEvent, wParamsEvent, eventJoinStmnt, err := getFilterSQLStmtForEventProperties(
				projectID, linkedEvent.Properties, queryFrom)
			if err != nil {
				return err, nil
			}

			if wStmtEvent != "" {
				whereEventHits = whereEventHits + " AND " + fmt.Sprintf("( %s )", wStmtEvent)
				qParams = append(qParams, wParamsEvent...)
			}

			// add user filter
			wStmtUser, wParamsUser, eventUserJoinStmnt, err := getFilterSQLStmtForUserProperties(projectID, linkedEvent.Properties, queryFrom)
			if err != nil {
				return err, nil
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
			U.CloseReadQuery(rows, tx)
			U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &log.Fields{"project_id": projectID})
		}
		// Get coalesced Id for Funnel Event user_ids
		userIDToCoalIDInfo, err := store.GetCoalesceIDFromUserIDs(userIDList, projectID, logCtx)
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

// GetConvertedUsersWithFilter Returns the list of eligible users who hit conversion
// event for userProperties from events table
func (store *MemSQL) GetConvertedUsersWithFilter(projectID int64, goalEventName string,
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
	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &log.Fields{"project_id": projectID})

	// Get coalesced Id for converted user_ids (without filter)
	userIDToCoalIDInfo, err := store.GetCoalesceIDFromUserIDs(userIDList, projectID, logCtx)
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

	return filteredUserIdToUserIDInfo, filteredCoalIDToUserIDInfo, coalUserIdConversionTimestamp, nil
}

// GetAdwordsCurrency Returns currency used for adwords customer_account_id
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
	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &logFields)

	return currency, nil
}
