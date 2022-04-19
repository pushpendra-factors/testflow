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
func (store *MemSQL) ExecuteKPIForAttribution(projectID uint64, query *model.AttributionQuery, debugQueryKey string, logCtx log.Entry) (map[string]model.KPIInfo, map[string]string, []string, error) {

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	kpiData := make(map[string]model.KPIInfo)
	groupUserIDToKpiID := make(map[string]string)
	var kpiKeys []string

	var kpiQueryResult model.QueryResult
	if query.AnalyzeType == model.AnalyzeTypeHSDeals || query.AnalyzeType == model.AnalyzeTypeSFOpportunities {

		var duplicatedRequest model.KPIQueryGroup
		U.DeepCopy(&query.KPI, &duplicatedRequest)
		resultGroup, statusCode := store.ExecuteKPIQueryGroup(projectID, debugQueryKey, duplicatedRequest)
		log.WithFields(log.Fields{"ResultGroup": resultGroup, "Status": statusCode}).Info("KPI-Attribution result received")
		if statusCode != http.StatusOK {
			logCtx.Error("failed to get KPI result for attribution query")
			if statusCode == http.StatusPartialContent {
				return kpiData, groupUserIDToKpiID, kpiKeys, errors.New("failed to get KPI result for attribution query - StatusPartialContent")
			}
			return kpiData, groupUserIDToKpiID, kpiKeys, errors.New("failed to get KPI result for attribution query")
		}
		for _, res := range resultGroup {
			// Skip the datetime header and the other result is of format. ex. "headers": ["$hubspot_deal_hs_object_id", "Revenue", "Pipeline", ...],
			if res.Headers[0] == "datetime" {
				kpiQueryResult = res
				log.WithFields(log.Fields{"KpiQueryResult": kpiQueryResult}).Info("KPI-Attribution result set")
				break
			}
		}
		if kpiQueryResult.Headers == nil || len(kpiQueryResult.Headers) == 0 {
			logCtx.Error("no-valid result for KPI query")
			return kpiData, groupUserIDToKpiID, kpiKeys, errors.New("no-valid result for KPI query")
		}

		kpiKeys = store.GetDataFromKPIResult(projectID, kpiQueryResult, &kpiData, logCtx)
	}

	// Pulling group ID (group user ID) for each KPI ID i.e. Deal ID or Opp ID
	log.WithFields(log.Fields{"kpiKeys": kpiKeys}).Info("KPI-Attribution keys set")
	if len(kpiKeys) == 0 {
		return kpiData, groupUserIDToKpiID, kpiKeys, errors.New("no valid KPIs found for this query to run")
	}

	groups, errCode := store.GetGroups(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("failed to get groups for project")
		return kpiData, groupUserIDToKpiID, kpiKeys, errors.New("failed to get groups for project")
	}

	_groupIDKey, _groupIDUserKey := getGroupKeys(query, groups)

	kpiKeyGroupUserIDList, err := store.PullGroupUserIDs(projectID, kpiKeys, _groupIDKey, &kpiData, &groupUserIDToKpiID, logCtx)
	if err != nil {
		return kpiData, groupUserIDToKpiID, kpiKeys, errors.New("no valid KPIs found for this query to run")
	}
	log.WithFields(log.Fields{"kpiKeyGroupUserIDList": kpiKeyGroupUserIDList}).Info("KPI-Attribution group set")

	err = store.PullKPIKeyUserGroupInfo(projectID, kpiKeyGroupUserIDList, _groupIDUserKey, &kpiData, &groupUserIDToKpiID, logCtx)
	if err != nil {
		return kpiData, groupUserIDToKpiID, kpiKeys, err
	}

	logCtx.Info("done pulling group user list ids for Deal or Opportunity")
	log.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData, "groupUserIDToKpiID": groupUserIDToKpiID,
		"kpiKeys": kpiKeys}).Info("KPI-Attribution kpiData reports 1")
	err = store.PullAllUsersByCustomerUserID(projectID, &kpiData, logCtx)
	if err != nil {
		return kpiData, groupUserIDToKpiID, kpiKeys, err
	}
	log.WithFields(log.Fields{"KPIAttribution": "Debug", "kpiData": kpiData, "groupUserIDToKpiID": groupUserIDToKpiID,
		"kpiKeys": kpiKeys}).Info("KPI-Attribution kpiData reports 2")

	return kpiData, groupUserIDToKpiID, kpiKeys, nil
}

func getGroupKeys(query *model.AttributionQuery, groups []model.Group) (string, string) {
	var _groupIDKey string
	var _groupIDUserKey string
	if query.AnalyzeType == model.AnalyzeTypeHSDeals {

		for _, group := range groups {
			if group.Name == model.GROUP_NAME_HUBSPOT_DEAL {
				//_groupIDNo = group.ID
				_groupIDKey = "group_" + strconv.Itoa(group.ID) + "_id"
				_groupIDUserKey = "group_" + strconv.Itoa(group.ID) + "_user_id"
			}
		}
	}

	if query.AnalyzeType == model.AnalyzeTypeSFOpportunities {
		for _, group := range groups {
			if group.Name == model.GROUP_NAME_SALESFORCE_OPPORTUNITY {
				//_groupIDNo = group.ID
				_groupIDKey = "group_" + strconv.Itoa(group.ID) + "_id"
				_groupIDUserKey = "group_" + strconv.Itoa(group.ID) + "_user_id"
			}
		}
	}
	return _groupIDKey, _groupIDUserKey
}

func (store *MemSQL) GetDataFromKPIResult(projectID uint64, kpiQueryResult model.QueryResult, kpiData *map[string]model.KPIInfo, logCtx log.Entry) []string {

	datetimeIdx := 0
	keyIdx := 1
	valIdx := 2
	var kpiKeys []string
	var kpiAggFunctionType []string

	var kpiValueHeaders []string
	for idx := valIdx; idx < len(kpiQueryResult.Headers); idx++ {
		kpiValueHeaders = append(kpiValueHeaders, kpiQueryResult.Headers[idx])
	}

	if len(kpiValueHeaders) == 0 {
		return kpiKeys
	}

	customMetrics, errMsg, statusCode := store.GetCustomMetricsByProjectId(projectID)
	if statusCode != http.StatusFound {
		logCtx.WithField("messageFinder", "Failed to get custom metrics").Error(errMsg)
		return kpiKeys
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
		logCtx.WithField("kpiAggFunctionType", kpiAggFunctionType).WithField("kpiValueHeaders", kpiValueHeaders).Warn("failed to get function types of all of given KPI")
		return kpiKeys
	}

	log.WithFields(log.Fields{"kpiValueHeaders": kpiValueHeaders}).Info("KPI-Attribution headers set")

	for _, row := range kpiQueryResult.Rows {

		log.WithFields(log.Fields{"Row": row}).Info("KPI-Attribution KPI Row")

		var kpiDetail model.KPIInfo

		// get ID
		kpiID := row[keyIdx].(string)
		kpiKeys = append(kpiKeys, kpiID)

		// get time
		eventTime, err := time.Parse(time.RFC3339, row[datetimeIdx].(string))
		if err != nil {
			logCtx.WithError(err).WithFields(log.Fields{"timestamp": row[datetimeIdx]}).Error("couldn't parse the timestamp for KPI query, continuing")
			continue
		}
		kpiDetail.Timestamp = eventTime.Unix()
		kpiDetail.TimeString = row[datetimeIdx].(string)

		// add kpi values
		var kpiVals []float64
		for vi := valIdx; vi < len(row); vi++ {
			val := float64(0)
			vInt, okInt := row[vi].(int)
			if !okInt {
				vFloat, okFloat := row[vi].(float64)
				if !okFloat {
					logCtx.WithError(err).WithFields(log.Fields{"value": row[vi]}).Error("couldn't parse the value for KPI query, continuing")
					val = 0.0
				} else {
					val = vFloat
				}
			} else {
				val = float64(vInt)
			}
			kpiVals = append(kpiVals, val)
		}
		kpiDetail.KpiValues = kpiVals

		// add headers
		kpiDetail.KpiHeaderNames = kpiValueHeaders
		// add aggregate function type
		kpiDetail.KpiAggFunctionTypes = kpiAggFunctionType

		// map to kpi data to key - final data
		(*kpiData)[kpiID] = kpiDetail
	}
	return kpiKeys
}

func (store *MemSQL) PullGroupUserIDs(projectID uint64, kpiKeys []string, _groupIDKey string, kpiData *map[string]model.KPIInfo, groupUserIDToKpiID *map[string]string, logCtx log.Entry) ([]string, error) {

	var kpiKeyGroupUserIDList []string
	kpiKeysIdPlaceHolder := U.GetValuePlaceHolder(len(kpiKeys))
	kpiKeysIdValue := U.GetInterfaceList(kpiKeys)
	groupUserQuery := "Select id, " + _groupIDKey + " FROM users WHERE project_id=? AND " + _groupIDKey + " IN ( " + kpiKeysIdPlaceHolder + " ) "
	var gUParams []interface{}
	gUParams = append(gUParams, projectID)
	gUParams = append(gUParams, kpiKeysIdValue...)
	gURows, tx1, err := store.ExecQueryWithContext(groupUserQuery, gUParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return kpiKeyGroupUserIDList, errors.New("failed to get groupUserQuery result for project")
	}
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
	defer U.CloseReadQuery(gURows, tx1)
	return kpiKeyGroupUserIDList, nil
}

func (store *MemSQL) PullKPIKeyUserGroupInfo(projectID uint64, kpiKeyGroupUserIDList []string, _groupIDUserKey string, kpiData *map[string]model.KPIInfo, groupUserIDToKpiID *map[string]string, logCtx log.Entry) error {
	// Pulling user ID for each KPI ID i.e. associated users with each KPI ID i.e. DealID or OppID - kpiIDToCoalUsers

	kpiKeysGroupUserIdPlaceHolder := U.GetValuePlaceHolder(len(kpiKeyGroupUserIDList))
	kpiKeysGroupUserIdValue := U.GetInterfaceList(kpiKeyGroupUserIDList)
	groupUserListQuery := "Select " + _groupIDUserKey + ", users.id, COALESCE(users.customer_user_id,users.id) FROM users WHERE project_id=? " +
		" AND (is_group_user=false or is_group_user IS NULL) AND " + _groupIDUserKey + " IN ( " + kpiKeysGroupUserIdPlaceHolder + " ) "
	var gULParams []interface{}
	gULParams = append(gULParams, projectID)
	gULParams = append(gULParams, kpiKeysGroupUserIdValue...)
	gULRows, tx2, err := store.ExecQueryWithContext(groupUserListQuery, gULParams)
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

		kpiID := (*groupUserIDToKpiID)[groupID]
		if _, exists := (*kpiData)[kpiID]; exists {
			v := (*kpiData)[kpiID]
			v.KpiCoalUserIds = append(v.KpiCoalUserIds, coalUserID)
			v.KpiUserIds = append(v.KpiUserIds, userID)
			(*kpiData)[kpiID] = v
		}
	}
	log.WithFields(log.Fields{"kpiData": kpiData}).Info("KPI-Attribution group set")
	defer U.CloseReadQuery(gULRows, tx2)
	return nil
}

func (store *MemSQL) PullAllUsersByCustomerUserID(projectID uint64, kpiData *map[string]model.KPIInfo, logCtx log.Entry) error {
	// Pulling user ID for each KPI ID i.e. associated users with each KPI ID i.e. DealID or OppID - kpiIDToCoalUsers

	var customerUserIdList []string
	for _, v := range *kpiData {
		customerUserIdList = append(customerUserIdList, v.KpiCoalUserIds...)
	}

	custUserIdToUserIds := make(map[string][]string)
	custUserIDPlaceHolder := U.GetValuePlaceHolder(len(customerUserIdList))
	custUserIDs := U.GetInterfaceList(customerUserIdList)
	groupUserListQuery := "Select users.id, users.customer_user_id FROM users WHERE project_id=? " +
		" AND users.customer_user_id IN ( " + custUserIDPlaceHolder + " ) "
	var gULParams []interface{}
	gULParams = append(gULParams, projectID)
	gULParams = append(gULParams, custUserIDs...)
	gULRows, tx2, err := store.ExecQueryWithContext(groupUserListQuery, gULParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return errors.New("failed to get groupUserListQuery result for project")
	}
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
	log.WithFields(log.Fields{"custUserIdToUserIds": custUserIdToUserIds}).Info("KPI-Attribution custUserIdToUserIds set")

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
	defer U.CloseReadQuery(gULRows, tx2)
	return nil
}

func (store *MemSQL) FireAttributionForKPI(projectID uint64, query *model.AttributionQuery,
	sessions map[string]map[string]model.UserSessionData,
	kpiData map[string]model.KPIInfo,
	sessionWT map[string][]float64, logCtx log.Entry) (*map[string]*model.AttributionData, bool, error) {

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
		attributionData, err = store.RunAttributionForMethodologyComparisonKpi(projectID,
			conversionFrom, conversionTo, query, sessions, kpiData, sessionWT, logCtx)
	} else {
		// Single event attribution.
		attributionData, err = store.runAttributionKPI(projectID,
			conversionFrom, conversionTo,
			query, sessions, kpiData, sessionWT, logCtx)
	}
	return attributionData, isCompare, err
}

func (store *MemSQL) runAttributionKPI(projectID uint64,
	conversionFrom, conversionTo int64,
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

	logCtx.WithFields(log.Fields{"userConversionHit": userConversionHit}).Info("KPI-Attribution userConversionHit")

	attributionData := make(map[string]*model.AttributionData)
	attributionData = model.AddUpConversionEventCount(userConversionHit, sessionWT)
	logCtx.WithFields(log.Fields{"attributionData": attributionData}).Info("KPI-Attribution attributionData")

	return &attributionData, nil
}

func (store *MemSQL) RunAttributionForMethodologyComparisonKpi(projectID uint64,
	conversionFrom, conversionTo int64, query *model.AttributionQuery,
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
			attributionData[key].ConversionEventCompareCount = attributionDataCompare[key].ConversionEventCount
		} else {
			attributionData[key].ConversionEventCompareCount = []float64{float64(0)}
		}
	}
	// filling any non-matched touch points
	for missingKey := range attributionDataCompare {
		if _, exists := attributionData[missingKey]; !exists {
			attributionData[missingKey] = &model.AttributionData{}
			attributionData[missingKey].ConversionEventCompareCount = attributionDataCompare[missingKey].ConversionEventCount
			attributionData[missingKey].ConversionEventCount = []float64{float64(0)}
		}
	}
	return &attributionData, nil
}
