package memsql

import (
	"database/sql"
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// ExecuteKPIForAttribution Executes the KPI sub-query for Attribution
func (store *MemSQL) ExecuteKPIForAttribution(projectID int64, query *model.AttributionQuery, debugQueryKey string,
	logCtx log.Entry, enableOptimisedFilterOnProfileQuery bool,
	enableOptimisedFilterOnEventUserQuery bool) (map[string]model.KPIInfo, []string, []string, error) {

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	kpiData := make(map[string]model.KPIInfo)
	groupUserIDToKpiID := make(map[string]string)
	var kpiKeys []string
	var kpiHeaders []string
	var kpiAggFunctionType []string
	var err error
	err, kpiKeys, kpiHeaders, kpiAggFunctionType = store.RunKPIGroupQuery(projectID, query, &kpiData, enableOptimisedFilterOnProfileQuery,
		enableOptimisedFilterOnEventUserQuery, debugQueryKey, logCtx)
	if err != nil {
		return kpiData, kpiHeaders, kpiAggFunctionType, err
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData, "groupUserIDToKpiID": groupUserIDToKpiID,
			"kpiKeys": kpiKeys}).Info("KPI-Attribution kpiData reports after RunKPIGroupQuery")
	}
	err = store.FillKPIGroupUserData(projectID, query, &kpiData, &kpiKeys, &groupUserIDToKpiID, logCtx)
	if err != nil {
		return kpiData, kpiHeaders, kpiAggFunctionType, err
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.Info("done pulling group user list ids for Deal or Opportunity")
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData, "groupUserIDToKpiID": groupUserIDToKpiID,
			"kpiKeys": kpiKeys}).Info("KPI-Attribution kpiData reports 1")
	}
	err = store.PullAllUsersByCustomerUserID(projectID, &kpiData, logCtx)
	if err != nil {
		return kpiData, kpiHeaders, kpiAggFunctionType, err
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData, "groupUserIDToKpiID": groupUserIDToKpiID,
			"kpiKeys": kpiKeys}).Info("KPI-Attribution kpiData reports 2")
	}
	return kpiData, kpiHeaders, kpiAggFunctionType, nil
}

// ExecuteKPIForAttributionV1 Executes the KPI sub-query for Attribution
func (store *MemSQL) ExecuteKPIForAttributionV1(projectID int64, query *model.AttributionKPIQueries, from int64, to int64, debugQueryKey string,
	logCtx log.Entry, enableOptimisedFilterOnProfileQuery bool,
	enableOptimisedFilterOnEventUserQuery bool) (map[string]model.KPIInfo, []string, []string, error) {

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	kpiData := make(map[string]model.KPIInfo)
	var headers []string
	var kpiAggFunctionType []string
	groupUserIDToKpiID := make(map[string]string)
	var kpiKeys []string
	var err error
	err, kpiKeys, headers, kpiAggFunctionType = store.RunKPIGroupQueryV1(projectID, query, from, to, &kpiData, enableOptimisedFilterOnProfileQuery,
		enableOptimisedFilterOnEventUserQuery, debugQueryKey, logCtx)
	if err != nil {
		return kpiData, headers, kpiAggFunctionType, err
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData, "groupUserIDToKpiID": groupUserIDToKpiID,
			"kpiKeys": kpiKeys}).Info("KPI-Attribution kpiData reports after RunKPIGroupQuery")
	}
	err = store.FillKPIGroupUserDataV1(projectID, query, &kpiData, &kpiKeys, &groupUserIDToKpiID, logCtx)
	if err != nil {
		return kpiData, headers, kpiAggFunctionType, err
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.Info("done pulling group user list ids for Deal or Opportunity")
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData, "groupUserIDToKpiID": groupUserIDToKpiID,
			"kpiKeys": kpiKeys}).Info("KPI-Attribution kpiData reports 1")
	}
	err = store.PullAllUsersByCustomerUserID(projectID, &kpiData, logCtx)
	if err != nil {
		return kpiData, headers, kpiAggFunctionType, err
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData, "groupUserIDToKpiID": groupUserIDToKpiID,
			"kpiKeys": kpiKeys}).Info("KPI-Attribution kpiData reports 2")
	}
	return kpiData, headers, kpiAggFunctionType, nil
}

func getGroupKeys(queryRunAnalyzeType string, groups []model.Group) (string, string) {
	var _groupIDKey string
	var _groupIDUserKey string
	if queryRunAnalyzeType == model.RunTypeHSDeals {

		for _, group := range groups {
			if group.Name == model.GROUP_NAME_HUBSPOT_DEAL {
				//_groupIDNo = group.ID
				_groupIDKey = "group_" + strconv.Itoa(group.ID) + "_id"
				_groupIDUserKey = "group_" + strconv.Itoa(group.ID) + "_user_id"
			}
		}
	}

	if queryRunAnalyzeType == model.RunTypeSFOpportunities {
		for _, group := range groups {
			if group.Name == model.GROUP_NAME_SALESFORCE_OPPORTUNITY {
				//_groupIDNo = group.ID
				_groupIDKey = "group_" + strconv.Itoa(group.ID) + "_id"
				_groupIDUserKey = "group_" + strconv.Itoa(group.ID) + "_user_id"
			}
		}
	}

	if queryRunAnalyzeType == model.RunTypeSFAccounts {
		for _, group := range groups {
			if group.Name == model.GROUP_NAME_SALESFORCE_ACCOUNT {
				//_groupIDNo = group.ID
				_groupIDKey = "group_" + strconv.Itoa(group.ID) + "_id"
				_groupIDUserKey = "group_" + strconv.Itoa(group.ID) + "_user_id"
			}
		}
	}

	if queryRunAnalyzeType == model.RunTypeHSCompanies {
		for _, group := range groups {
			if group.Name == model.GROUP_NAME_HUBSPOT_COMPANY {
				//_groupIDNo = group.ID
				_groupIDKey = "group_" + strconv.Itoa(group.ID) + "_id"
				_groupIDUserKey = "group_" + strconv.Itoa(group.ID) + "_user_id"
			}
		}
	}

	return _groupIDKey, _groupIDUserKey
}

// GetDataFromKPIResult adds values in kpiData from kpiQueryResult
func (store *MemSQL) GetDataFromKPIResult(projectID int64, kpiQueryResult model.QueryResult, kpiData *map[string]model.KPIInfo, query *model.AttributionQuery, logCtx log.Entry) ([]string, []string, []string) {

	datetimeIdx := 0
	keyIdx := 1
	valIdx := 2
	var kpiAggFunctionType []string

	var kpiValueHeaders []string
	for idx := valIdx; idx < len(kpiQueryResult.Headers); idx++ {
		kpiValueHeaders = append(kpiValueHeaders, kpiQueryResult.Headers[idx])
	}

	if len(kpiValueHeaders) == 0 {
		return nil, nil, nil
	}

	customMetrics, errMsg, statusCode := store.GetCustomMetricsByProjectId(projectID)

	if statusCode != http.StatusFound {
		logCtx.WithFields(log.Fields{"messageFinder": "Failed to get custom metrics", "err_code": statusCode}).Error(errMsg)
		return nil, nil, nil
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"customMetrics": customMetrics}).Info("customMetrics for project in attribution query")
	}

	mapKpiAggFunctionType := make(map[string]string)
	for _, kpi := range customMetrics {
		var customMetricTransformation model.CustomMetricTransformation
		err := U.DecodePostgresJsonbToStructType(kpi.Transformations, &customMetricTransformation)
		if err != nil {
			continue
		}
		mapKpiAggFunctionType[kpi.Name] = customMetricTransformation.AggregateFunction
	}

	for _, kpiName := range kpiValueHeaders {
		if fType, exists := mapKpiAggFunctionType[kpiName]; exists {
			kpiAggFunctionType = append(kpiAggFunctionType, fType)
		}
	}

	if len(kpiValueHeaders) != len(kpiAggFunctionType) {
		logCtx.WithField("kpiAggFunctionType", kpiAggFunctionType).WithField("kpiValueHeaders",
			kpiValueHeaders).Warn("failed to get function types of all of given KPI")
		return nil, nil, nil
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"kpiValueHeaders": kpiValueHeaders}).Info("KPI-Attribution headers set")
	}
	kpiKeys := model.AddKPIKeyDataInMap(kpiQueryResult, logCtx, keyIdx, datetimeIdx, query.From, query.To, valIdx, kpiValueHeaders, kpiAggFunctionType, kpiData)

	return kpiKeys, kpiValueHeaders, kpiAggFunctionType
}

// GetDataFromKPIResultV1 adds values in kpiData from kpiQueryResult
func (store *MemSQL) GetDataFromKPIResultV1(projectID int64, kpiQueryResult model.QueryResult, kpiData *map[string]model.KPIInfo, from int64, to int64, logCtx log.Entry) ([]string, []string, []string) {

	datetimeIdx := 0
	keyIdx := 1
	valIdx := 2
	var kpiAggFunctionType []string

	var kpiValueHeaders []string
	for idx := valIdx; idx < len(kpiQueryResult.Headers); idx++ {
		kpiValueHeaders = append(kpiValueHeaders, kpiQueryResult.Headers[idx])
	}

	if len(kpiValueHeaders) == 0 {
		return nil, nil, nil
	}

	customMetrics, errMsg, statusCode := store.GetCustomMetricsByProjectId(projectID)

	if statusCode != http.StatusFound {
		logCtx.WithFields(log.Fields{"messageFinder": "Failed to get custom metrics", "err_code": statusCode}).Error(errMsg)
		return nil, nil, nil
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"customMetrics": customMetrics}).Info("customMetrics for project in attribution query")
	}

	mapKpiAggFunctionType := make(map[string]string)
	for _, kpi := range customMetrics {
		var customMetricTransformation model.CustomMetricTransformation
		err := U.DecodePostgresJsonbToStructType(kpi.Transformations, &customMetricTransformation)
		if err != nil {
			continue
		}
		mapKpiAggFunctionType[kpi.Name] = customMetricTransformation.AggregateFunction
	}

	for _, kpiName := range kpiValueHeaders {
		if fType, exists := mapKpiAggFunctionType[kpiName]; exists {
			kpiAggFunctionType = append(kpiAggFunctionType, fType)
		}
	}

	if len(kpiValueHeaders) != len(kpiAggFunctionType) {
		logCtx.WithField("kpiAggFunctionType", kpiAggFunctionType).WithField("kpiValueHeaders",
			kpiValueHeaders).Warn("failed to get function types of all of given KPI")
		return nil, nil, nil
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"kpiValueHeaders": kpiValueHeaders}).Info("KPI-Attribution headers set")
	}
	kpiKeys := model.AddKPIKeyDataInMap(kpiQueryResult, logCtx, keyIdx, datetimeIdx, from, to, valIdx, kpiValueHeaders, kpiAggFunctionType, kpiData)
	return kpiKeys, kpiValueHeaders, kpiAggFunctionType
}

func (store *MemSQL) RunKPIGroupQuery(projectID int64, query *model.AttributionQuery, kpiData *map[string]model.KPIInfo,
	enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery bool, debugQueryKey string, logCtx log.Entry) (error, []string, []string, []string) {

	var kpiQueryResult model.QueryResult
	if query.AnalyzeType == model.AnalyzeTypeHSDeals || query.AnalyzeType == model.AnalyzeTypeSFOpportunities {

		var duplicatedRequest model.KPIQueryGroup
		U.DeepCopy(&query.KPI, &duplicatedRequest)
		for index := range duplicatedRequest.Queries {
			duplicatedRequest.Queries[index].LimitNotApplicable = true
			// Making sure the internal KPI group queries have same from to as parent attribution query
			duplicatedRequest.Queries[index].From = query.From
			duplicatedRequest.Queries[index].To = query.To
		}
		resultGroup, statusCode := store.ExecuteKPIQueryGroup(projectID, debugQueryKey,
			duplicatedRequest, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		log.WithFields(log.Fields{"KPIQueryGroupDebug": duplicatedRequest, "ResultGroup": resultGroup, "Status": statusCode}).Info("KPI-Attribution result received")
		log.WithFields(log.Fields{"ResultGroup": resultGroup, "Status": statusCode}).Info("KPI-Attribution result received")
		if statusCode != http.StatusOK {

			logCtx.WithField("err_code", statusCode).Error("failed to get KPI result for attribution query")
			if statusCode == http.StatusPartialContent {
				return errors.New("failed to get KPI result for attribution query - StatusPartialContent"), nil, nil, nil
			}
			return errors.New("failed to get KPI result for attribution query"), nil, nil, nil
		}
		for _, res := range resultGroup {
			// Skip the datetime header and the other result is of format. ex. "headers": ["$hubspot_deal_hs_object_id", "Revenue", "Pipeline", ...],
			if res.Headers[0] == "datetime" {
				kpiQueryResult = res
				logCtx.WithFields(log.Fields{"KpiQueryResult": kpiQueryResult}).Info("KPI-Attribution result set")
				break
			}
		}
		if kpiQueryResult.Headers == nil || len(kpiQueryResult.Headers) == 0 {
			logCtx.Error("no-valid result for KPI query")
			return errors.New("no-valid result for KPI query"), nil, nil, nil
		}
		kpiKeys, kpiHeaders, kpiAggFunctionType := store.GetDataFromKPIResult(projectID, kpiQueryResult, kpiData, query, logCtx)
		return nil, kpiKeys, kpiHeaders, kpiAggFunctionType
	}
	return errors.New("not a valid type of query for KPI Attribution"), nil, nil, nil
}

// RunKPIGroupQueryV1 runs kpi group query and adds the result in kpiData
func (store *MemSQL) RunKPIGroupQueryV1(projectID int64, query *model.AttributionKPIQueries, from int64, to int64, kpiData *map[string]model.KPIInfo,
	enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery bool, debugQueryKey string, logCtx log.Entry) (error, []string, []string, []string) {

	var kpiQueryResult model.QueryResult
	if query.AnalyzeType == model.AnalyzeTypeHSDeals || query.AnalyzeType == model.AnalyzeTypeSFOpportunities {

		var duplicatedRequest model.KPIQueryGroup
		U.DeepCopy(&query.KPI, &duplicatedRequest)
		for index := range duplicatedRequest.Queries {
			duplicatedRequest.Queries[index].LimitNotApplicable = true
			// Making sure the internal KPI group queries have same from to as parent attribution query
			duplicatedRequest.Queries[index].From = from
			duplicatedRequest.Queries[index].To = to
		}
		resultGroup, statusCode := store.ExecuteKPIQueryGroup(projectID, debugQueryKey,
			duplicatedRequest, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		logCtx.WithFields(log.Fields{"KPIQueryGroupDebug": duplicatedRequest, "ResultGroup": resultGroup, "Status": statusCode}).Info("KPI-Attribution result received")
		if statusCode != http.StatusOK {
			logCtx.WithField("err_code", statusCode).Error("failed to get KPI result for attribution query")
			if statusCode == http.StatusPartialContent {
				return errors.New("failed to get KPI result for attribution query - StatusPartialContent"), nil, nil, nil
			}
			return errors.New("failed to get KPI result for attribution query"), nil, nil, nil
		}
		for _, res := range resultGroup {
			// Skip the datetime header and the other result is of format. ex. "headers": ["$hubspot_deal_hs_object_id", "Revenue", "Pipeline", ...],
			if res.Headers[0] == "datetime" {
				kpiQueryResult = res
				logCtx.WithFields(log.Fields{"KpiQueryResult": kpiQueryResult}).Info("KPI-Attribution result set")
				break
			}
		}
		if kpiQueryResult.Headers == nil || len(kpiQueryResult.Headers) == 0 {
			logCtx.Error("no-valid result for KPI query")
			return errors.New("no-valid result for KPI query"), nil, nil, nil
		}
		kpiKeys, kpiHeaders, kpiAggFunctionType := store.GetDataFromKPIResultV1(projectID, kpiQueryResult, kpiData, from, to, logCtx)
		return nil, kpiKeys, kpiHeaders, kpiAggFunctionType
	}
	return errors.New("not a valid type of query for KPI Attribution"), nil, nil, nil
}

func (store *MemSQL) FillKPIGroupUserData(projectID int64, query *model.AttributionQuery, kpiData *map[string]model.KPIInfo,
	kpiKeys *[]string, groupUserIDToKpiID *map[string]string, logCtx log.Entry) error {

	if len(*kpiKeys) == 0 {
		logCtx.WithFields(log.Fields{"Method": "FillKPIGroupUserData"}).Info("No kpiKeys found")
		return nil
	}
	// Pulling group ID (group user ID) for each KPI ID i.e. Deal ID or Opp ID
	logCtx.WithFields(log.Fields{"kpiKeys": kpiKeys}).Info("KPI-Attribution keys set")
	if len(*kpiKeys) == 0 {
		return errors.New("no valid KPIs found for this query to run")
	}

	_groups, errCode := store.GetGroups(projectID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("failed to get groups for project")
		return errors.New("failed to get groups for project")
	}

	// Get Group keys for Analyze Type because we need to pull deal/opp data for both deal/opp And company/acc level attribution
	_groupIDKey, _groupIDUserKey := getGroupKeys(query.AnalyzeType, _groups)
	kpiKeyGroupUserIDList, err := store.PullGroupUserIDs(projectID, kpiKeys, _groupIDKey, kpiData, groupUserIDToKpiID, logCtx)
	if err != nil {
		return errors.New("no valid KPIs found for this query to run")
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"kpiKeyGroupUserIDList": kpiKeyGroupUserIDList}).Info("KPI-Attribution group set FillKPIGroupUserData")
	}

	if query.RunType == model.RunTypeHSDeals || query.RunType == model.RunTypeSFOpportunities {

		_groupIDKey, _groupIDUserKey = getGroupKeys(query.AnalyzeType, _groups)
		err = store.FillNormalUsersForKPIDealOppUser(projectID, kpiKeyGroupUserIDList, _groupIDUserKey, kpiData, groupUserIDToKpiID, logCtx)

	} else if query.RunType == model.RunTypeSFAccounts || query.RunType == model.RunTypeHSCompanies {

		// Pull the common id field based on run type
		_groupIDKey, _groupIDUserKey = getGroupKeys(query.RunType, _groups)
		accCompIds, err := store.FillCompanyAccountIDs(projectID, kpiKeyGroupUserIDList, kpiData, groupUserIDToKpiID, logCtx)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"accCompIds": accCompIds}).Info("KPI-Attribution accCompIds set FillKPIGroupUserData")
		}
		if err != nil {
			return err
		}
		err = store.FillNormalUsersForKPIAccCompUser(projectID, accCompIds, _groupIDUserKey, kpiData, logCtx)
	}

	return err
}

// FillKPIGroupUserDataV1 adds kpi group user data in kpiData
func (store *MemSQL) FillKPIGroupUserDataV1(projectID int64, query *model.AttributionKPIQueries, kpiData *map[string]model.KPIInfo,
	kpiKeys *[]string, groupUserIDToKpiID *map[string]string, logCtx log.Entry) error {

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"query": query,
			"kpiData":            kpiData,
			"kpiKeys":            kpiKeys,
			"groupUserIDToKpiID": groupUserIDToKpiID}).Info("Values in FillKPIGroupUserDataV1")
	}

	if len(*kpiKeys) == 0 {
		logCtx.WithFields(log.Fields{"Method": "FillKPIGroupUserData"}).Info("No kpiKeys found")
		return nil
	}

	// Pulling group ID (group user ID) for each KPI ID i.e. Deal ID or Opp ID
	logCtx.WithFields(log.Fields{"kpiKeys": kpiKeys}).Info("KPI-Attribution keys set")
	if len(*kpiKeys) == 0 {
		return errors.New("no valid KPIs found for this query to run")
	}

	_groups, errCode := store.GetGroups(projectID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("failed to get groups for project")
		return errors.New("failed to get groups for project")
	}

	// Get Group keys for Analyze Type because we need to pull deal/opp data for both deal/opp And company/acc level attribution
	_groupIDKey, _groupIDUserKey := getGroupKeys(query.AnalyzeType, _groups)
	kpiKeyGroupUserIDList, err := store.PullGroupUserIDs(projectID, kpiKeys, _groupIDKey, kpiData, groupUserIDToKpiID, logCtx)
	if err != nil {
		return errors.New("no valid KPIs found for this query to run")
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"kpiKeyGroupUserIDList": kpiKeyGroupUserIDList}).Info("KPI-Attribution group set FillKPIGroupUserData")
	}

	if query.RunType == model.RunTypeHSDeals || query.RunType == model.RunTypeSFOpportunities {

		_groupIDKey, _groupIDUserKey = getGroupKeys(query.AnalyzeType, _groups)
		err = store.FillNormalUsersForKPIDealOppUser(projectID, kpiKeyGroupUserIDList, _groupIDUserKey, kpiData, groupUserIDToKpiID, logCtx)

	} else if query.RunType == model.RunTypeSFAccounts || query.RunType == model.RunTypeHSCompanies {

		// Pull the common id field based on run type
		_groupIDKey, _groupIDUserKey = getGroupKeys(query.RunType, _groups)
		accCompIds, err := store.FillCompanyAccountIDs(projectID, kpiKeyGroupUserIDList, kpiData, groupUserIDToKpiID, logCtx)
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"accCompIds": accCompIds}).Info("KPI-Attribution accCompIds set FillKPIGroupUserData")
		}
		if err != nil {
			return err
		}
		err = store.FillNormalUsersForKPIAccCompUser(projectID, accCompIds, _groupIDUserKey, kpiData, logCtx)
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"kpiData": kpiData}).Info("Values after FillKPIGroupUserDataV1")
	}
	return err
}

func (store *MemSQL) PullGroupUserIDs(projectID int64, kpiKeys *[]string, _groupIDKey string,
	kpiData *map[string]model.KPIInfo, groupUserIDToKpiID *map[string]string, logCtx log.Entry) ([]string, error) {
	logFields := log.Fields{
		"project_id":    projectID,
		"_group_id_key": _groupIDKey,
	}
	if kpiKeys == nil || len(*kpiKeys) == 0 {
		logCtx.WithFields(log.Fields{"Method": "PullGroupUserIDs"}).Info("No kpiKeys found")
		return nil, nil
	}

	var kpiKeyGroupUserIDList []string
	kpiKeysIdPlaceHolder := U.GetValuePlaceHolder(len(*kpiKeys))
	kpiKeysIdValue := U.GetInterfaceList(*kpiKeys)
	groupUserQuery := "Select id, " + _groupIDKey + " FROM users WHERE project_id=? AND " + _groupIDKey + " IN ( " + kpiKeysIdPlaceHolder + " ) "
	var gUParams []interface{}
	gUParams = append(gUParams, projectID)
	gUParams = append(gUParams, kpiKeysIdValue...)
	gURows, tx1, err, reqID := store.ExecQueryWithContext(groupUserQuery, gUParams)
	if err != nil {
		logCtx.WithField("query_id", reqID).WithError(err).Error("SQL Query failed")
		return kpiKeyGroupUserIDList, errors.New("failed to get groupUserQuery result for project")
	}

	startReadTime := time.Now()
	for gURows.Next() {
		var groupUserIDNull sql.NullString
		var kpiIDNull sql.NullString
		if err = gURows.Scan(&groupUserIDNull, &kpiIDNull); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
			continue
		}

		groupUserID := U.IfThenElse(groupUserIDNull.Valid, groupUserIDNull.String, model.PropertyValueNone).(string)
		kpiID := U.IfThenElse(kpiIDNull.Valid, kpiIDNull.String, model.PropertyValueNone).(string)

		if groupUserID == model.PropertyValueNone || kpiID == model.PropertyValueNone {
			continue
		}

		// enrich KPI group ID
		v := (*kpiData)[kpiID]
		v.KpiGroupID = groupUserID
		(*kpiData)[kpiID] = v

		(*groupUserIDToKpiID)[groupUserID] = kpiID
		kpiKeyGroupUserIDList = append(kpiKeyGroupUserIDList, groupUserID)

	}
	err = gURows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in PullGroupUserIDs")
		return nil, err
	}
	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &logFields)

	defer U.CloseReadQuery(gURows, tx1)
	return kpiKeyGroupUserIDList, nil
}

// FillCompanyAccountIDs pulls account/company IDs for deals/opp and returns the list of accounts/company IDs
func (store *MemSQL) FillCompanyAccountIDs(projectID int64, kpiKeyGroupUserIDList []string,
	kpiData *map[string]model.KPIInfo, groupUserIDToKpiID *map[string]string, logCtx log.Entry) ([]string, error) {

	logFields := log.Fields{
		"project_id": projectID,
	}

	var kpiKeyAccountCompanyGroupUserIDList []string
	kpiKeysIdPlaceHolder := U.GetValuePlaceHolder(len(kpiKeyGroupUserIDList))
	kpiKeysIdValue := U.GetInterfaceList(kpiKeyGroupUserIDList)
	groupUserQuery := "Select left_group_user_id, right_group_user_id from group_relationships where project_id=? AND ( left_group_user_id IN ( " + kpiKeysIdPlaceHolder + " )  OR right_group_user_id IN ( " + kpiKeysIdPlaceHolder + " )) "
	var gUParams []interface{}
	gUParams = append(gUParams, projectID)
	gUParams = append(gUParams, kpiKeysIdValue...) // for left_group_user_id
	gUParams = append(gUParams, kpiKeysIdValue...) // for right_group_user_id
	gURows, tx1, err, reqID := store.ExecQueryWithContext(groupUserQuery, gUParams)
	if err != nil {
		logCtx.WithField("query_id", reqID).WithError(err).Error("SQL Query failed")
		return kpiKeyAccountCompanyGroupUserIDList, errors.New("failed to get groupUserQuery result for project")
	}

	startReadTime := time.Now()
	for gURows.Next() {
		var leftGroupUserIDNull sql.NullString
		var rightGroupUserIDNull sql.NullString
		if err = gURows.Scan(&leftGroupUserIDNull, &rightGroupUserIDNull); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
			continue
		}

		leftGroupUserID := U.IfThenElse(leftGroupUserIDNull.Valid, leftGroupUserIDNull.String, model.PropertyValueNone).(string)
		rightGroupUserID := U.IfThenElse(rightGroupUserIDNull.Valid, rightGroupUserIDNull.String, model.PropertyValueNone).(string)

		if leftGroupUserID == model.PropertyValueNone || rightGroupUserID == model.PropertyValueNone {
			continue
		}

		// enrich KPI group ID
		if _, exists1 := (*groupUserIDToKpiID)[leftGroupUserID]; exists1 {

			kpiKeyID := (*groupUserIDToKpiID)[leftGroupUserID]
			if _, exists := (*kpiData)[kpiKeyID]; exists {
				v := (*kpiData)[kpiKeyID]
				v.KpiAccountCompanyGroupID = rightGroupUserID
				(*kpiData)[kpiKeyID] = v
				kpiKeyAccountCompanyGroupUserIDList = append(kpiKeyAccountCompanyGroupUserIDList, rightGroupUserID)
			}
		} else {
			if _, exists1 := (*groupUserIDToKpiID)[rightGroupUserID]; exists1 {
				kpiKeyID := (*groupUserIDToKpiID)[rightGroupUserID]
				// enrich KPI group ID
				if _, exists := (*kpiData)[kpiKeyID]; exists {
					v := (*kpiData)[kpiKeyID]
					v.KpiAccountCompanyGroupID = leftGroupUserID
					(*kpiData)[kpiKeyID] = v
					kpiKeyAccountCompanyGroupUserIDList = append(kpiKeyAccountCompanyGroupUserIDList, leftGroupUserID)
				}
			}
		}
	}
	err = gURows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in PullGroupUserIDs")
		return nil, err
	}
	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &logFields)

	defer U.CloseReadQuery(gURows, tx1)
	return kpiKeyAccountCompanyGroupUserIDList, nil
}

// FillNormalUsersForKPIAccCompUser fills all the users for given account in kpiData
func (store *MemSQL) FillNormalUsersForKPIAccCompUser(projectID int64, kpiKeyGroupUserIDList []string, accountCompanyKey string,
	kpiData *map[string]model.KPIInfo, logCtx log.Entry) error {
	// Pulling user ID for each KPI ID i.e. associated users with each KPI ID i.e. DealID or OppID - kpiIDToCoalUsers

	if kpiKeyGroupUserIDList == nil || len(kpiKeyGroupUserIDList) == 0 {
		logCtx.WithFields(log.Fields{"Method": "FillNormalUsersForKPIAccCompUser"}).Info("No kpiKey Group UserId found")
		return nil
	}

	_kpiAccCompData := make(map[string]model.AccountCompUsers)

	kpiKeysGroupUserIdPlaceHolder := U.GetValuePlaceHolder(len(kpiKeyGroupUserIDList))
	kpiKeysGroupUserIdValue := U.GetInterfaceList(kpiKeyGroupUserIDList)
	groupUserListQuery := "Select " + accountCompanyKey + ", users.id, COALESCE(users.customer_user_id,users.id) FROM users WHERE project_id=? " +
		" AND (is_group_user=false or is_group_user IS NULL) AND " + accountCompanyKey + " IN ( " + kpiKeysGroupUserIdPlaceHolder + " ) "
	var gULParams []interface{}
	gULParams = append(gULParams, projectID)
	gULParams = append(gULParams, kpiKeysGroupUserIdValue...)
	gULRows, tx2, err, _ := store.ExecQueryWithContext(groupUserListQuery, gULParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return errors.New("failed to get groupUserListQuery result for project")
	}

	for gULRows.Next() {
		var groupIDNull sql.NullString
		var userIDNull sql.NullString
		var coalUserIDNull sql.NullString
		if err = gULRows.Scan(&groupIDNull, &userIDNull, &coalUserIDNull); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
			continue
		}

		groupID := U.IfThenElse(groupIDNull.Valid, groupIDNull.String, model.PropertyValueNone).(string)
		userID := U.IfThenElse(userIDNull.Valid, userIDNull.String, model.PropertyValueNone).(string)
		coalUserID := U.IfThenElse(coalUserIDNull.Valid, coalUserIDNull.String, model.PropertyValueNone).(string)
		if coalUserID == model.PropertyValueNone || groupID == model.PropertyValueNone {
			logCtx.WithError(err).Error("Values are not correct - coalUserID & groupID. Ignoring row. Continuing")
			continue
		}

		if _, exists := (_kpiAccCompData)[groupID]; exists {
			v := (_kpiAccCompData)[groupID]
			v.KpiCoalUserIds = append(v.KpiCoalUserIds, coalUserID)
			v.KpiUserIds = append(v.KpiUserIds, userID)
			(_kpiAccCompData)[groupID] = v
		} else {
			var v model.AccountCompUsers
			v.KpiCoalUserIds = append(v.KpiCoalUserIds, coalUserID)
			v.KpiUserIds = append(v.KpiUserIds, userID)
			(_kpiAccCompData)[groupID] = v
		}
	}
	err = gULRows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in FillNormalUsersForKPIDealOppUser")
		return err
	}
	if C.GetAttributionDebug() == 1 {
		logFields := log.Fields{"_kpiAccCompData": _kpiAccCompData, "project_id": projectID}
		logCtx.WithFields(logFields).Info("KPI-Attribution group set _kpiAccCompData FillNormalUsersForKPIAccCompUser")
	}

	// Fill the pulled kpiAccCompData users into actual kpiData for Deals/Opportunities for acc/comp level attribution
	for k, v := range *kpiData {
		accCompID := v.KpiAccountCompanyGroupID
		if _, exists := (_kpiAccCompData)[accCompID]; exists {
			vAccComp := (_kpiAccCompData)[accCompID]

			vDealOpp := (*kpiData)[k]
			vDealOpp.KpiCoalUserIds = append(v.KpiCoalUserIds, vAccComp.KpiCoalUserIds...)
			vDealOpp.KpiUserIds = append(v.KpiUserIds, vAccComp.KpiUserIds...)
			(*kpiData)[k] = vDealOpp
		}
	}

	if C.GetAttributionDebug() == 1 {
		logFields := log.Fields{"kpiData": kpiData, "project_id": projectID}
		logCtx.WithFields(logFields).Info("KPI-Attribution group set kpiData FillNormalUsersForKPIAccCompUser")
	}
	defer U.CloseReadQuery(gULRows, tx2)

	return nil
}

func (store *MemSQL) FillNormalUsersForKPIDealOppUser(projectID int64, kpiKeyGroupUserIDList []string, dealOppKey string,
	kpiData *map[string]model.KPIInfo, groupUserIDToKpiID *map[string]string, logCtx log.Entry) error {
	// Pulling user ID for each KPI ID i.e. associated users with each KPI ID i.e. DealID or OppID - kpiIDToCoalUsers

	if kpiKeyGroupUserIDList == nil || len(kpiKeyGroupUserIDList) == 0 {
		logCtx.WithFields(log.Fields{"Method": "FillNormalUsersForKPIDealOppUser"}).Info("no group users found, exiting")
		return nil
	}
	kpiKeysGroupUserIdPlaceHolder := U.GetValuePlaceHolder(len(kpiKeyGroupUserIDList))
	kpiKeysGroupUserIdValue := U.GetInterfaceList(kpiKeyGroupUserIDList)
	groupUserListQuery := "Select " + dealOppKey + ", users.id, COALESCE(users.customer_user_id,users.id) FROM users WHERE project_id=? " +
		" AND (is_group_user=false or is_group_user IS NULL) AND " + dealOppKey + " IN ( " + kpiKeysGroupUserIdPlaceHolder + " ) "
	var gULParams []interface{}
	gULParams = append(gULParams, projectID)
	gULParams = append(gULParams, kpiKeysGroupUserIdValue...)
	gULRows, tx2, err, reqID := store.ExecQueryWithContext(groupUserListQuery, gULParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return errors.New("failed to get groupUserListQuery result for project")
	}

	startReadTime := time.Now()
	for gULRows.Next() {
		var groupIDNull sql.NullString
		var userIDNull sql.NullString
		var coalUserIDNull sql.NullString
		if err = gULRows.Scan(&groupIDNull, &userIDNull, &coalUserIDNull); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed. Ignoring row. Continuing")
			continue
		}

		groupID := U.IfThenElse(groupIDNull.Valid, groupIDNull.String, model.PropertyValueNone).(string)
		userID := U.IfThenElse(userIDNull.Valid, userIDNull.String, model.PropertyValueNone).(string)
		coalUserID := U.IfThenElse(coalUserIDNull.Valid, coalUserIDNull.String, model.PropertyValueNone).(string)
		if coalUserID == model.PropertyValueNone || groupID == model.PropertyValueNone {
			logCtx.WithError(err).Error("Values are not correct - coalUserID & groupID. Ignoring row. Continuing")
			continue
		}

		kpiID := (*groupUserIDToKpiID)[groupID]
		if _, exists := (*kpiData)[kpiID]; exists {
			v := (*kpiData)[kpiID]
			v.KpiCoalUserIds = append(v.KpiCoalUserIds, coalUserID)
			v.KpiUserIds = append(v.KpiUserIds, userID)
			(*kpiData)[kpiID] = v
		}
	}
	err = gULRows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing query in FillNormalUsersForKPIDealOppUser")
		return err
	}
	logFields := log.Fields{"kpiData": kpiData, "project_id": projectID}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(logFields).Info("KPI-Attribution group set")
	}
	defer U.CloseReadQuery(gULRows, tx2)

	U.LogReadTimeWithQueryRequestID(startReadTime, reqID, &logFields)

	return nil
}

func (store *MemSQL) PullAllUsersByCustomerUserID(projectID int64, kpiData *map[string]model.KPIInfo, logCtx log.Entry) error {
	// Pulling user ID for each KPI ID i.e. associated users with each KPI ID i.e. DealID or OppID - kpiIDToCoalUsers

	var customerUserIdList []string
	for _, v := range *kpiData {
		customerUserIdList = append(customerUserIdList, v.KpiCoalUserIds...)
	}

	_, custUserIdToUserIds, err := store.FetchAllUsersAndCustomerUserDataInBatches(projectID, customerUserIdList, logCtx)
	if err != nil {
		return err
	}

	for k, v := range *kpiData {
		userIdMap := make(map[string]bool)
		// Add new users
		for _, uid := range v.KpiUserIds {
			userIdMap[uid] = true
		}

		// Add new users
		for _, cid := range v.KpiCoalUserIds {
			if _, exists := custUserIdToUserIds[cid]; exists {
				for _, userID := range custUserIdToUserIds[cid] {
					userIdMap[userID] = true
				}
			}
		}
		var users []string
		for uID := range userIdMap {
			users = append(users, uID)
		}
		// Replace users & Update the KPIInfo
		v.KpiUserIds = users
		(*kpiData)[k] = v
	}
	return nil
}

// AddUserIdInKPIDataWithoutCustomerUserId adds KPI key as user_id for each KPI_id where customer_user_id is null.
// If customer_user_id was present, at least one user_id would have been added in KpiUserIds
func (store *MemSQL) AddUserIdInKPIDataWithoutCustomerUserId(kpiData *map[string]model.KPIInfo, logCtx log.Entry) error {
	var keysWithNoCustomerUserId []string
	for k, v := range *kpiData {
		if v.KpiUserIds == nil || len(v.KpiUserIds) == 0 {
			v.KpiUserIds = []string{k}
			(*kpiData)[k] = v
			keysWithNoCustomerUserId = append(keysWithNoCustomerUserId, k)
		}
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"keysWithNoCustomerUserId": keysWithNoCustomerUserId}).Info("AddUserIdInKPIDataWithoutCustomerUserId")
	}
	return nil
}

func (store *MemSQL) FireAttributionForKPI(query *model.AttributionQuery,
	sessions map[string]map[string]model.UserSessionData,
	kpiData map[string]model.KPIInfo,
	sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, bool, error) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	isCompare := false
	var err error

	var attributionData *map[string]*model.AttributionData
	if query.AttributionMethodologyCompare != "" {
		// Two AttributionMethodologies comparison
		isCompare = true
		attributionData, err = store.RunAttributionForMethodologyComparisonKpi(query, sessions, kpiData, sessionWT, logCtx)
	} else {
		// Single event attribution.
		attributionData, err = store.runAttributionKPI(query, sessions, kpiData, sessionWT, logCtx)
	}
	return attributionData, isCompare, err
}

// FireAttributionForKPIV1 returns attribution data using UserSessionData
func (store *MemSQL) FireAttributionForKPIV1(projectID int64, query *model.AttributionQueryV1,
	sessions map[string]map[string]model.UserSessionData,
	kpiData map[string]model.KPIInfo,
	sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, bool, error) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	isCompare := false
	var err error

	var attributionData *map[string]*model.AttributionData
	if query.AttributionMethodologyCompare != "" {
		// Two AttributionMethodologies comparison
		isCompare = true
		attributionData, err = store.RunAttributionForMethodologyComparisonKpiV1(query, sessions, kpiData, sessionWT, logCtx)
	} else {
		// Single event attribution.
		attributionData, err = store.runAttributionKPIV1(query, sessions, kpiData, sessionWT, logCtx)
	}
	return attributionData, isCompare, err
}

func (store *MemSQL) runAttributionKPI(
	query *model.AttributionQuery,
	sessions map[string]map[string]model.UserSessionData,
	kpiData map[string]model.KPIInfo,
	sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, error) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	logCtx.WithFields(log.Fields{"sessionWT": sessionWT}).Info("KPI-Attribution sessionWT")

	// Apply attribution based on given attribution methodology
	userConversionHit, err := model.ApplyAttributionKPI(
		query.QueryType,
		query.AttributionMethodology,
		sessions,
		kpiData,
		query.LookbackDays, query.From, query.To, query.AttributionKey)
	if err != nil {
		return nil, err
	}

	// update attribution weight
	updateSessionWT(sessionWT, kpiData)

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"userConversionHit": userConversionHit}).Info("KPI-Attribution userConversionHit")
	}
	attributionData := make(map[string]*model.AttributionData)
	attributionData = model.AddUpConversionEventCount(userConversionHit, sessionWT)
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData")
	}
	return &attributionData, nil
}

func (store *MemSQL) runAttributionKPIV1(
	query *model.AttributionQueryV1,
	sessions map[string]map[string]model.UserSessionData,
	kpiData map[string]model.KPIInfo,
	sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, error) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	logCtx.WithFields(log.Fields{"sessionWT": sessionWT}).Info("KPI-Attribution sessionWT")

	// Apply attribution based on given attribution methodology
	userConversionHit, err := model.ApplyAttributionKPI(
		query.QueryType,
		query.AttributionMethodology,
		sessions,
		kpiData,
		query.LookbackDays, query.From, query.To, query.AttributionKey)
	if err != nil {
		return nil, err
	}

	// update attribution weight
	updateSessionWT(sessionWT, kpiData)

	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"userConversionHit": userConversionHit}).Info("KPI-Attribution userConversionHit")
	}
	attributionData := make(map[string]*model.AttributionData)
	attributionData = model.AddUpConversionEventCount(userConversionHit, sessionWT)
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData")
	}
	return &attributionData, nil
}

func updateSessionWT(sessionWT map[string][]float64, kpiData map[string]model.KPIInfo) {
	for key := range sessionWT {
		for _, value := range kpiData[key].KpiValuesList {
			if !value.IsConverted {
				//not converted then subtract from final result
				for idx, val := range value.Values {
					sessionWT[key][idx] = sessionWT[key][idx] - val
				}
			}
		}
	}
}

func (store *MemSQL) RunAttributionForMethodologyComparisonKpi(query *model.AttributionQuery,
	sessions map[string]map[string]model.UserSessionData,
	kpiData map[string]model.KPIInfo,
	sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, error) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	// Attribution based on given attribution methodology.
	userConversionHit, err := model.ApplyAttributionKPI(
		query.QueryType,
		query.AttributionMethodology,
		sessions,
		kpiData,
		query.LookbackDays, query.From, query.To, query.AttributionKey)
	if err != nil {
		return nil, err
	}
	// update attribution weight
	updateSessionWT(sessionWT, kpiData)

	attributionData := model.AddUpConversionEventCount(userConversionHit, sessionWT)

	// Attribution based on given attributionMethodologyCompare methodology.
	userConversionCompareHit, err := model.ApplyAttributionKPI(
		query.QueryType,
		query.AttributionMethodologyCompare,
		sessions,
		kpiData,
		query.LookbackDays, query.From, query.To, query.AttributionKey)
	if err != nil {
		return nil, err
	}

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

// RunAttributionForMethodologyComparisonKpiV1 runs attribution for queries with compare model
func (store *MemSQL) RunAttributionForMethodologyComparisonKpiV1(query *model.AttributionQueryV1,
	sessions map[string]map[string]model.UserSessionData,
	kpiData map[string]model.KPIInfo,
	sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, error) {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)

	// Attribution based on given attribution methodology.
	userConversionHit, err := model.ApplyAttributionKPI(
		query.QueryType,
		query.AttributionMethodology,
		sessions,
		kpiData,
		query.LookbackDays, query.From, query.To, query.AttributionKey)
	if err != nil {
		return nil, err
	}
	// update attribution weight
	updateSessionWT(sessionWT, kpiData)

	attributionData := model.AddUpConversionEventCount(userConversionHit, sessionWT)

	// Attribution based on given attributionMethodologyCompare methodology.
	userConversionCompareHit, err := model.ApplyAttributionKPI(
		query.QueryType,
		query.AttributionMethodologyCompare,
		sessions,
		kpiData,
		query.LookbackDays, query.From, query.To, query.AttributionKey)
	if err != nil {
		return nil, err
	}

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
