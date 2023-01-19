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
func (store *MemSQL) ExecuteAttributionQueryV1(projectID int64, queryOriginal *model.AttributionQueryV1,
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

	if C.GetAttributionDebug() == 1 {
		logCtx.Info("Hitting ExecuteAttributionQueryV1")
	}

	var query *model.AttributionQueryV1
	U.DeepCopy(queryOriginal, &query)

	// pulling project setting to build attribution query
	settings, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, errors.New("failed to get project settings during attribution call")
	}
	// enrich RunType for attribution query
	err := model.EnrichRequestUsingAttributionConfigV1(projectID, query, settings, logCtx)
	if err != nil {
		return nil, err
	}

	// supporting existing old/saved queries
	//model.AddDefaultAnalyzeType(query)
	model.AddDefaultKeyDimensionsToAttributionQueryV1(query)
	model.AddDefaultMarketingEventTypeTacticOfferV1(query)

	// LandingPage not allowed for tactic
	if query.AttributionKey == model.AttributionKeyLandingPage && query.TacticOfferType == model.MarketingEventTypeTactic {
		return nil, errors.New("can not get landing page level report for Tactic")
	}

	// AllPageView not allowed for tactic
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

	marketingReports, err := store.FetchMarketingReportsV1(projectID, *query, *projectSetting)
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

	sessionEventNameID, _, err := store.getEventInformationV1(projectID, query, *logCtx)
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

	coalUserIdConversionTimestamp, userInfo, kpiData, usersIDsToAttribute, err3 := store.PullConvertedUsersV1(projectID, query, debugQueryKey,
		enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery, logCtx)

	if err3 != nil {
		return nil, err3
	}

	var userData map[string]map[string]model.UserSessionData
	var err4 error
	if query.AttributionKey == model.AttributionKeyAllPageView {
		userData, err4 = store.PullPagesOfConvertedUsersV1(projectID, query, sessionEventNameID, usersIDsToAttribute, marketingReports, contentGroupNamesList, logCtx)
	} else {
		userData, err4 = store.PullSessionsOfConvertedUsersV1(projectID, query, sessionEventNameID, usersIDsToAttribute, marketingReports, contentGroupNamesList, logCtx)
	}
	if err4 != nil {
		return nil, err4
	}

	// Pull Offline touch points for all the cases: "Tactic",  "Offer", "TacticOffer"

	store.AppendOTPSessionsV1(projectID, query, &userData, *logCtx)
	logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Pull Offline touch points user data took time")
	queryStartTime = time.Now().UTC().Unix()

	attributionData, isCompare, err2 := store.GetAttributionDataV1(projectID, query, userData, userInfo, coalUserIdConversionTimestamp, marketingReports, kpiData, logCtx)

	if err2 != nil {
		return nil, err2
	}

	// Filter out the key values from query (apply filter after performance enrichment)
	model.ApplyFilterV1(attributionData, query)
	logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Metrics, Performance report, filter took time")

	queryStartTime = time.Now().UTC().Unix()

	result := ProcessAttributionDataToResultV1(projectID, query, attributionData, isCompare, queryStartTime, marketingReports, kpiData, logCtx)
	result.Meta.Currency = ""
	if projectSetting.IntAdwordsCustomerAccountId != nil && *projectSetting.IntAdwordsCustomerAccountId != "" {
		currency, _ := store.GetAdwordsCurrency(projectID, *projectSetting.IntAdwordsCustomerAccountId, query.From, query.To, *logCtx)
		result.Meta.Currency = currency
	}

	logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Total query took time")

	model.SanitizeResult(result)
	return result, nil
}

func (store *MemSQL) PullPagesOfConvertedUsers(projectID int64, query *model.AttributionQuery, sessionEventNameID string, usersToBeAttributed []string,
	marketingReports *model.MarketingReports, contentGroupNamesList []string, logCtx *log.Entry) (map[string]map[string]model.UserSessionData, error) {

	pages := make(map[string]map[string]model.UserSessionData)

	queryStartTime := time.Now().UTC().Unix()
	// Pull Sessions for the cases: "Tactic" and "TacticOffer".
	// If Landing Page level report, pull for offer as well.
	if query.TacticOfferType != model.MarketingEventTypeOffer || query.AttributionKey == model.AttributionKeyAllPageView {
		var _pages map[string]map[string]model.UserSessionData
		var users []string
		var err error

		// Get all the pages (userId, attributionId, UserSessionData) for given period by attribution key
		_pages, users, err = store.getAllThePages(projectID, sessionEventNameID, query, usersToBeAttributed, marketingReports, contentGroupNamesList, *logCtx)

		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Pull Sessions data data took time")
		}
		queryStartTime = time.Now().UTC().Unix()

		if err != nil {
			return nil, err
		}

		usersInfo, _, err := store.GetCoalesceIDFromUserIDs(users, projectID, *logCtx)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Get Coalesce user data took time")
		}
		queryStartTime = time.Now().UTC().Unix()
		if err != nil {
			return nil, err
		}

		model.UpdateSessionsMapWithCoalesceID(_pages, usersInfo, &pages)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Update Sessions Coalesce user data took time")
		}
		queryStartTime = time.Now().UTC().Unix()
	}
	return pages, nil
}

func (store *MemSQL) PullPagesOfConvertedUsersV1(projectID int64, query *model.AttributionQueryV1, sessionEventNameID string, usersToBeAttributed []string,
	marketingReports *model.MarketingReports, contentGroupNamesList []string, logCtx *log.Entry) (map[string]map[string]model.UserSessionData, error) {

	pages := make(map[string]map[string]model.UserSessionData)

	queryStartTime := time.Now().UTC().Unix()
	// Pull Sessions for the cases: "Tactic" and "TacticOffer".
	// If Landing Page level report, pull for offer as well.
	if query.TacticOfferType != model.MarketingEventTypeOffer || query.AttributionKey == model.AttributionKeyAllPageView {
		var _pages map[string]map[string]model.UserSessionData
		var users []string
		var err error

		// Get all the pages (userId, attributionId, UserSessionData) for given period by attribution key
		_pages, users, err = store.getAllThePagesV1(projectID, sessionEventNameID, query, usersToBeAttributed, marketingReports, contentGroupNamesList, *logCtx)

		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Pull Sessions data data took time")
		}
		queryStartTime = time.Now().UTC().Unix()

		if err != nil {
			return nil, err
		}

		usersInfo, _, err := store.GetCoalesceIDFromUserIDs(users, projectID, *logCtx)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Get Coalesce user data took time")
		}
		queryStartTime = time.Now().UTC().Unix()
		if err != nil {
			return nil, err
		}

		model.UpdateSessionsMapWithCoalesceID(_pages, usersInfo, &pages)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Update Sessions Coalesce user data took time")
		}
		queryStartTime = time.Now().UTC().Unix()
	}
	return pages, nil
}

func (store *MemSQL) PullSessionsOfConvertedUsers(projectID int64, query *model.AttributionQuery, sessionEventNameID string, usersToBeAttributed []string,
	marketingReports *model.MarketingReports, contentGroupNamesList []string, logCtx *log.Entry) (map[string]map[string]model.UserSessionData, error) {

	sessions := make(map[string]map[string]model.UserSessionData)

	queryStartTime := time.Now().UTC().Unix()
	// Pull Sessions for the cases: "Tactic" and "TacticOffer".
	// If Landing Page level report, pull for offer as well.
	if query.TacticOfferType != model.MarketingEventTypeOffer || query.AttributionKey == model.AttributionKeyLandingPage {
		// Get all the sessions (userId, attributionId, UserSessionData) for given period by attribution key
		_sessions, sessionUsers, err := store.getAllTheSessions(projectID, sessionEventNameID, query, usersToBeAttributed, marketingReports, contentGroupNamesList, *logCtx)
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

//PullSessionsOfConvertedUsersV1 returns session data of converted users for required time period
func (store *MemSQL) PullSessionsOfConvertedUsersV1(projectID int64, query *model.AttributionQueryV1, sessionEventNameID string, usersToBeAttributed []string,
	marketingReports *model.MarketingReports, contentGroupNamesList []string, logCtx *log.Entry) (map[string]map[string]model.UserSessionData, error) {

	sessions := make(map[string]map[string]model.UserSessionData)

	queryStartTime := time.Now().UTC().Unix()
	// Pull Sessions for the cases: "Tactic" and "TacticOffer".
	// If Landing Page level report, pull for offer as well.
	if query.TacticOfferType != model.MarketingEventTypeOffer || query.AttributionKey == model.AttributionKeyLandingPage || query.AttributionKey == model.AttributionKeyAllPageView {
		var _sessions map[string]map[string]model.UserSessionData
		var sessionUsers []string
		var err error

		// Get all the sessions (userId, attributionId, UserSessionData) for given period by attribution key
		if query.AttributionKey == model.AttributionKeyAllPageView {
			_sessions, sessionUsers, err = store.getAllThePagesV1(projectID, sessionEventNameID, query, usersToBeAttributed, marketingReports, contentGroupNamesList, *logCtx)
		} else {
			_sessions, sessionUsers, err = store.getAllTheSessionsV1(projectID, sessionEventNameID, query, usersToBeAttributed, marketingReports, contentGroupNamesList, *logCtx)
		}

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
	eventNameToIDList map[string][]interface{},
	debugQueryKey string, enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool,
	logCtx *log.Entry) (map[string]int64, []model.UserEventInfo, map[string]model.KPIInfo, []string, error) {

	var coalUserIdConversionTimestamp map[string]int64
	var usersToBeAttributed []model.UserEventInfo
	var usersIDsToAttribute []string
	kpiData := make(map[string]model.KPIInfo)
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

		log.WithFields(log.Fields{"UserKPIAttribution": "Debug", "kpiData": kpiData}).Info("UserKPI Attribution kpiData")

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
		// This thread is for AnalyzeTypeHSDeals, AnalyzeTypeSFOpportunities, AnalyzeTypeSFAccounts AnalyzeTypeHSCompanies
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
		if C.GetAttributionDebug() == 1 {
			log.WithFields(log.Fields{"KPIAttribution": "Debug", "usersIDsToAttribute": usersIDsToAttribute}).Info("KPI Attribution usersIDsToAttribute")
		}
	}
	return coalUserIdConversionTimestamp, usersToBeAttributed, kpiData, usersIDsToAttribute, nil
}

//PullConvertedUsersV1 runs kpi group query and returns converted users in the form of kpi data
func (store *MemSQL) PullConvertedUsersV1(projectID int64, query *model.AttributionQueryV1,
	debugQueryKey string, enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool,
	logCtx *log.Entry) (map[string]int64, []model.UserEventInfo, map[string]model.KPIInfo, []string, error) {

	var coalUserIdConversionTimestamp map[string]int64
	var usersToBeAttributed []model.UserEventInfo
	var usersIDsToAttribute []string
	var kpiDataForGroupQuery map[string]model.KPIInfo
	var kpiDataList []map[string]model.KPIInfo
	headerPositionMap := make(map[string]int64)
	var defaultHeader []string
	headerCount := int64(0)
	kpiData := make(map[string]model.KPIInfo)

	for index, individualKPIQuery := range query.KPIQueries {

		if query.KPIQueries[index].AnalyzeType == model.AnalyzeTypeUserKPI {

			var err error
			queryStartTime := time.Now().UTC().Unix()

			kpiDataForGroupQuery, err = store.ExecuteUserKPIForAttributionV1(projectID, &individualKPIQuery, query.From, query.To, debugQueryKey,
				*logCtx, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("UserKPI query execution took time")
			}
			if err != nil {
				return nil, nil, nil, nil, err
			}

			log.WithFields(log.Fields{"UserKPIAttribution": "Debug", "kpiData": kpiDataForGroupQuery}).Info("UserKPI Attribution kpiData")

			_uniqueUsers := make(map[string]int)
			// Get user IDs for Revenue Attribution
			for _, data := range kpiDataForGroupQuery {
				for _, userID := range data.KpiUserIds {
					_uniqueUsers[userID] = 1
				}
			}

			for id, _ := range _uniqueUsers {
				usersIDsToAttribute = append(usersIDsToAttribute, id)
			}
		} else {
			// This thread is for AnalyzeTypeHSDeals, AnalyzeTypeSFOpportunities, AnalyzeTypeSFAccounts AnalyzeTypeHSCompanies
			var err error
			queryStartTime := time.Now().UTC().Unix()
			kpiDataForGroupQuery, err = store.ExecuteKPIForAttributionV1(projectID, &individualKPIQuery, query.From, query.To, debugQueryKey,
				*logCtx, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("KPI query execution took time")
			}
			if err != nil {
				return nil, nil, nil, nil, err
			}

			if C.GetAttributionDebug() == 1 {
				log.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiDataForGroupQuery}).Info("KPI Attribution kpiData")
			}

			_uniqueUsers := make(map[string]int)
			// Get user IDs for Revenue Attribution
			for _, data := range kpiDataForGroupQuery {
				for _, userID := range data.KpiUserIds {
					_uniqueUsers[userID] = 1
				}
			}

			for id, _ := range _uniqueUsers {
				usersIDsToAttribute = append(usersIDsToAttribute, id)
			}
			if C.GetAttributionDebug() == 1 {
				log.WithFields(log.Fields{"KPIAttribution": "Debug", "usersIDsToAttribute": usersIDsToAttribute}).Info("KPI Attribution usersIDsToAttribute")
			}
		}

		// adding header position in headerPositionMap and creating default header
		kpiDataList = append(kpiDataList, kpiDataForGroupQuery)
		for _, val := range kpiDataForGroupQuery {
			for _, header := range val.KpiHeaderNames {
				if _, ok := headerPositionMap[header]; !ok {
					headerCount = headerCount + 1
					defaultHeader = append(defaultHeader, header)
					headerPositionMap[header] = headerCount
				}
			}
			break
		}

		// combining data with same kpiId
		for kpiId, val := range kpiDataForGroupQuery {
			if value, ok := kpiData[kpiId]; ok {
				value.KpiValuesList = append(value.KpiValuesList, val.KpiValuesList...)
				kpiData[kpiId] = value
			}
			kpiData[kpiId] = val
		}
	}

	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData}).Info("KPI Attribution kpiData with separate headers")
	}
	// updating kpiData with all the headers and respective values
	for kpiId, KpiInfo := range kpiData {
		var newKpiValuesList []model.KpiRowValue
		for _, kpiRowValue := range KpiInfo.KpiValuesList {
			var defaultValue = make([]float64, headerCount)
			for index, header := range KpiInfo.KpiHeaderNames {
				defaultValue[headerPositionMap[header]] = kpiRowValue.Values[index]
			}
			kpiRowValue.Values = defaultValue
			newKpiValuesList = append(newKpiValuesList, kpiRowValue)
		}
		KpiInfo.KpiValuesList = newKpiValuesList
		KpiInfo.KpiHeaderNames = defaultHeader
		kpiData[kpiId] = KpiInfo
	}
	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData}).Info("KPI Attribution kpiData with all the headers")
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

		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"attributionData": attributionData, "sessions": sessions}).Info("Done FireAttributionV1. Attribution debug AnalyzeTypeUsers.")
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

		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("Done AddTheAddedKeysAndMetrics. Attribution debug AnalyzeTypeUsers.")
		}
		// Add the performance information no of conversion event = 1
		model.AddPerformanceData(attributionData, query.AttributionKey, marketingReports, 1)
		for key, _ := range *attributionData {
			(*attributionData)[key].ConvAggFunctionType = convAggFunctionType
		}
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("Done AddTheAddedKeysAndMetrics, AddPerformanceData")
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
		if C.GetAttributionDebug() == 1 || query.AnalyzeType == model.AnalyzeTypeUserKPI {
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
			kpiValues := model.KPIValueListToValues(kpiData[key])
			sessionWT[key] = kpiValues
			if kpiValues != nil || len(kpiValues) > 1 {
				noOfConversionEvents = U.MaxInt(noOfConversionEvents, len(kpiValues))
			}
		}

		if C.GetAttributionDebug() == 1 || query.AnalyzeType == model.AnalyzeTypeUserKPI {
			uniqUsers := len(groupSessions)
			logCtx.WithFields(log.Fields{"AttributionDebug": groupSessions}).Info(fmt.Sprintf("Total users with session: %d", uniqUsers))
		}
		var err error
		attributionData, isCompare, err = store.FireAttributionForKPI(projectID, query, groupSessions, kpiData, sessionWT, *logCtx)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("FireAttribution KPI took time")
		}
		queryStartTime = time.Now().UTC().Unix()
		if C.GetAttributionDebug() == 1 || query.AnalyzeType == model.AnalyzeTypeUserKPI {
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
			if len(val.KpiAggFunctionTypes) > 0 && val.KpiAggFunctionTypes != nil {
				convAggFunctionType = val.KpiAggFunctionTypes
				break
			}
		}
		for key := range *attributionData {
			(*attributionData)[key].ConvAggFunctionType = convAggFunctionType
		}

		if C.GetAttributionDebug() == 1 || query.AnalyzeType == model.AnalyzeTypeUserKPI {
			logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData before merge")
		}

		// Add the Added keys
		model.AddTheAddedKeysAndMetrics(attributionData, query, groupSessions, noOfConversionEvents)

		if C.GetAttributionDebug() == 1 || query.AnalyzeType == model.AnalyzeTypeUserKPI {
			logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData AddTheAddedKeysAndMetrics")
		}

		// Add the performance information
		model.AddPerformanceData(attributionData, query.AttributionKey, marketingReports, noOfConversionEvents)

		if C.GetAttributionDebug() == 1 || query.AnalyzeType == model.AnalyzeTypeUserKPI {
			logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData AddPerformanceData")
		}

	}
	return attributionData, isCompare, nil
}

// GetAttributionDataV1 runs attribution on groupSessions
func (store *MemSQL) GetAttributionDataV1(projectID int64, query *model.AttributionQueryV1, sessions map[string]map[string]model.UserSessionData,
	usersToBeAttributed []model.UserEventInfo, coalUserIdConversionTimestamp map[string]int64, marketingReports *model.MarketingReports,
	kpiData map[string]model.KPIInfo, logCtx *log.Entry) (*map[string]*model.AttributionData, bool, error) {

	queryStartTime := time.Now().UTC().Unix()

	isCompare := false
	var attributionData *map[string]*model.AttributionData

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
		kpiValues := model.KPIValueListToValues(kpiData[key])
		sessionWT[key] = kpiValues
		if kpiValues != nil || len(kpiValues) > 1 {
			noOfConversionEvents = U.MaxInt(noOfConversionEvents, len(kpiValues))
		}
	}

	if C.GetAttributionDebug() == 1 {
		uniqUsers := len(groupSessions)
		logCtx.WithFields(log.Fields{"AttributionDebug": groupSessions}).Info(fmt.Sprintf("Total users with session: %d", uniqUsers))
	}
	var err error
	attributionData, isCompare, err = store.FireAttributionForKPIV1(projectID, query, groupSessions, kpiData, sessionWT, *logCtx)
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
		if len(val.KpiAggFunctionTypes) > 0 && val.KpiAggFunctionTypes != nil {
			convAggFunctionType = val.KpiAggFunctionTypes
			break
		}
	}
	for key := range *attributionData {
		(*attributionData)[key].ConvAggFunctionType = convAggFunctionType
	}

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData before merge")
	}

	// Add the Added keys
	model.AddTheAddedKeysAndMetricsV1(attributionData, query, groupSessions, noOfConversionEvents)

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData AddTheAddedKeysAndMetrics")
	}

	// Add the performance information
	model.AddPerformanceData(attributionData, query.AttributionKey, marketingReports, noOfConversionEvents)

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData AddPerformanceData")
	}

	return attributionData, isCompare, nil
}

func ProcessAttributionDataToResult(projectID int64, query *model.AttributionQuery,
	attributionData *map[string]*model.AttributionData, isCompare bool, queryStartTime int64,
	marketingReports *model.MarketingReports, kpiData map[string]model.KPIInfo, logCtx *log.Entry) *model.QueryResult {

	result := &model.QueryResult{}

	if query.AttributionKey == model.AttributionKeyAllPageView {

		if query.AnalyzeType == model.AnalyzeTypeHSDeals || query.AnalyzeType == model.AnalyzeTypeSFOpportunities || query.AnalyzeType == model.AnalyzeTypeUserKPI {

			result = model.ProcessQueryKPIPageUrl(query, attributionData, *logCtx, kpiData, isCompare)
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query Landing PageUrl took time")
			}
			queryStartTime = time.Now().UTC().Unix()

		} else {
			result = model.ProcessQueryPageUrl(query, attributionData, *logCtx, isCompare)
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query Landing PageUrl took time")
			}
			queryStartTime = time.Now().UTC().Unix()
		}

	} else if query.AttributionKey == model.AttributionKeyLandingPage {

		if query.AnalyzeType == model.AnalyzeTypeHSDeals || query.AnalyzeType == model.AnalyzeTypeSFOpportunities || query.AnalyzeType == model.AnalyzeTypeUserKPI {

			result = model.ProcessQueryKPILandingPageUrl(query, attributionData, *logCtx, kpiData, isCompare)
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query Landing PageUrl took time")
			}
			queryStartTime = time.Now().UTC().Unix()

		} else {
			result = model.ProcessQueryLandingPageUrl(query, attributionData, *logCtx, isCompare)
			if C.GetAttributionDebug() == 1 {
				logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query Landing PageUrl took time")
			}
			queryStartTime = time.Now().UTC().Unix()
		}

	} else if query.AnalyzeType == model.AnalyzeTypeHSDeals || query.AnalyzeType == model.AnalyzeTypeSFOpportunities {
		// execution similar to the normal run - still keeping it separate for better understanding
		result = model.ProcessQueryKPI(query, attributionData, marketingReports, isCompare, kpiData)
		if C.GetAttributionDebug() == 1 || query.AnalyzeType == model.AnalyzeTypeUserKPI {
			logCtx.WithFields(log.Fields{"result": result}).Info(fmt.Sprintf("KPI-Attribution result"))
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query KPI took time")
		}
		queryStartTime = time.Now().UTC().Unix()

	} else if query.AnalyzeType == model.AnalyzeTypeUserKPI {
		result = model.ProcessQueryUserKPI(query, attributionData, marketingReports, isCompare, kpiData)
		if C.GetAttributionDebug() == 1 || query.AnalyzeType == model.AnalyzeTypeUserKPI {
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

//ProcessAttributionDataToResultV1 converts attributionData to result for different types of attribution queries
func ProcessAttributionDataToResultV1(projectID int64, query *model.AttributionQueryV1,
	attributionData *map[string]*model.AttributionData, isCompare bool, queryStartTime int64,
	marketingReports *model.MarketingReports, kpiData map[string]model.KPIInfo, logCtx *log.Entry) *model.QueryResult {

	result := &model.QueryResult{}

	if query.AttributionKey == model.AttributionKeyAllPageView {

		result = model.ProcessQueryKPIPageUrlV1(query, attributionData, *logCtx, kpiData, isCompare)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query Landing PageUrl took time")
		}
		queryStartTime = time.Now().UTC().Unix()

	} else if query.AttributionKey == model.AttributionKeyLandingPage {

		result = model.ProcessQueryKPILandingPageUrlV1(query, attributionData, *logCtx, kpiData, isCompare)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query Landing PageUrl took time")
		}
		queryStartTime = time.Now().UTC().Unix()

	} else {
		// execution similar to the normal run - still keeping it separate for better understanding
		result = model.ProcessQueryKPIV1(query, attributionData, marketingReports, isCompare, kpiData)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"result": result}).Info(fmt.Sprintf("KPI-Attribution result"))
			logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("Process Query KPI took time")
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

	// Add users who hit conversion event
	for key, val := range coalUserIdConversionTimestamp {
		usersToBeAttributed = append(usersToBeAttributed, model.UserEventInfo{CoalUserID: key,
			EventName: goalEventName, Timestamp: val, EventType: 0})
	}

	err, linkedFunnelEventUsers := store.GetLinkedFunnelEventUsersFilter(projectID, conversionFrom, conversionTo, query.LinkedEvents, eventNameToIDList, userIDToInfoConverted, logCtx)
	if err != nil {
		return userIDToInfoConverted, usersToBeAttributed, coalUserIdConversionTimestamp, err
	}

	model.MergeUsersToBeAttributed(&usersToBeAttributed, linkedFunnelEventUsers)

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
func (store *MemSQL) getAllThePagesV1(projectId int64, sessionEventNameId string, query *model.AttributionQueryV1, usersToPullSessionFor []string,
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
		processErr := model.ProcessEventRowsV1(rows, query, reports, contentGroupNamesList, &attributedSessionsByUserId, &userIdsWithSession, logCtx, reqID)
		U.CloseReadQuery(rows, tx)
		if processErr != nil {
			return attributedSessionsByUserId, userIdsWithSession, processErr
		}
	}

	return attributedSessionsByUserId, userIdsWithSession, nil
}

// Returns the all the sessions (userId,attributionId,minTimestamp,maxTimestamp) for given
// users from given period including lookback
func (store *MemSQL) getAllTheSessions(projectId int64, sessionEventNameId string, query *model.AttributionQuery, usersToPullSessionFor []string,
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

// Returns the all the sessions (userId,attributionId,minTimestamp,maxTimestamp) for given
// users from given period including lookback
func (store *MemSQL) getAllTheSessionsV1(projectId int64, sessionEventNameId string, query *model.AttributionQueryV1, usersToPullSessionFor []string,
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
		processErr := model.ProcessEventRowsV1(rows, query, reports, contentGroupNamesList, &attributedSessionsByUserId, &userIdsWithSession, logCtx, reqID)
		U.CloseReadQuery(rows, tx)
		if processErr != nil {
			return attributedSessionsByUserId, userIdsWithSession, processErr
		}
	}

	return attributedSessionsByUserId, userIdsWithSession, nil
}

// Returns the all the sessions (userId,attributionId,minTimestamp,maxTimestamp) for given
// users from given period including lookback
func (store *MemSQL) getAllTheSessionsAttributionKPI(projectId int64, sessionEventNameId string, query *model.AttributionQueryV1, usersToPullSessionFor []string,
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
		processErr := model.ProcessEventRowsV1(rows, query, reports, contentGroupNamesList, &attributedSessionsByUserId, &userIdsWithSession, logCtx, reqID)
		U.CloseReadQuery(rows, tx)
		if processErr != nil {
			return attributedSessionsByUserId, userIdsWithSession, processErr
		}
	}

	return attributedSessionsByUserId, userIdsWithSession, nil
}

// FetchAllUsersAndCustomerUserDataInBatches returns usersIds for given list of customer_user_id (i.e. coal_id) in batches
func (store *MemSQL) FetchAllUsersAndCustomerUserDataInBatches(projectID int64, customerUserIdList []string, logCtx log.Entry) (map[string]string, map[string][]string, error) {

	if customerUserIdList == nil || len(customerUserIdList) == 0 {
		return nil, nil, errors.New("no customer user IDs found, exiting")
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

	if customerUserIdList == nil || len(customerUserIdList) == 0 {
		return nil, nil, errors.New("no customer user IDs found, exiting")
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

	return filteredUserIdToUserIDInfo, filteredCoalIDToUserIDInfo, coalUserIdConversionTimestamp, nil
}

// Return conversion event Id, list of all event_ids(Conversion and funnel events) and a Id to name mapping
func (store *MemSQL) getEventInformationV1(projectId int64,
	query *model.AttributionQueryV1, logCtx log.Entry) (string, map[string][]interface{}, error) {

	names := model.BuildEventNamesPlaceholderV1(query)
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
	//if len(eventNameToId[query.ConversionEvent.Name]) == 0 && query.AnalyzeType == model.AnalyzeTypeUsers {
	//	logCtx.Error("conversion event name : " + query.ConversionEvent.Name + " not found")
	//	return "", nil, errors.New("conversion event name : " + query.ConversionEvent.Name + " not found")
	//}
	for _, linkedEvent := range query.LinkedEvents {
		if len(eventNameToId[linkedEvent.Name]) == 0 {
			logCtx.Error("linked event name : " + linkedEvent.Name + " not found")
			return "", nil, errors.New("linked event name : " + linkedEvent.Name + " not found")
		}
	}
	sessionEventNameId := eventNameToId[U.EVENT_NAME_SESSION][0].(string)
	return sessionEventNameId, eventNameToId, nil
}

// AppendOTPSessionsV1 adds OTP sessions in user Session data as per attribution query
func (store *MemSQL) AppendOTPSessionsV1(projectID int64, query *model.AttributionQueryV1,
	sessions *map[string]map[string]model.UserSessionData, logCtx log.Entry) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	otpEvent, err := store.getOfflineEventData(projectID, logCtx)
	if err != nil {
		logCtx.Info("no OTP events/sessions found. Skipping computation")
		return
	}

	_sessionsOTP, sessionOTPUsers, err := store.fetchOTPSessionsV1(projectID, otpEvent.ID, query, logCtx)
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
