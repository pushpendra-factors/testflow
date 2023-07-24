package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// ExecuteUserKPIForAttribution Executes the KPI sub-query for Attribution
func (store *MemSQL) ExecuteUserKPIForAttribution(projectID int64, query *model.AttributionQuery, debugQueryKey string,
	logCtx log.Entry, enableOptimisedFilterOnProfileQuery bool,
	enableOptimisedFilterOnEventUserQuery bool) (map[string]model.KPIInfo, []string, []string, error) {

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	kpiData := make(map[string]model.KPIInfo)
	var kpiKeys []string
	var kpiHeaders []string
	var kpiAggFunctionType []string
	var err error
	err, kpiKeys, kpiHeaders, kpiAggFunctionType = store.RunUserKPIGroupQuery(projectID, query, &kpiData, enableOptimisedFilterOnProfileQuery,
		enableOptimisedFilterOnEventUserQuery, debugQueryKey, logCtx)
	if err != nil {
		return kpiData, kpiHeaders, kpiAggFunctionType, err
	}
	logCtx.WithFields(log.Fields{"UserKPIAttribution": "Debug", "kpiData": kpiData,
		"kpiKeys": kpiKeys}).Info("UserKPI-Attribution kpiData reports after RunUserKPIGroupQuery")

	err = store.AddCoalUserIDinKPIData(&kpiData)
	if err != nil {
		return kpiData, kpiHeaders, kpiAggFunctionType, err
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.Info("done pulling group user list ids for Deal or Opportunity")
		logCtx.WithFields(log.Fields{"UserKPIAttribution": "Debug", "kpiData": kpiData,
			"kpiKeys": kpiKeys}).Info("UserKPI-Attribution kpiData reports 1")
	}
	err = store.PullAllUsersByCustomerUserID(projectID, &kpiData, logCtx)
	if err != nil {
		return kpiData, kpiHeaders, kpiAggFunctionType, err
	}
	logCtx.WithFields(log.Fields{"UserKPIAttribution": "Debug", "kpiData": kpiData,
		"kpiKeys": kpiKeys}).Info("UserKPI-Attribution kpiData reports 2")

	return kpiData, kpiHeaders, kpiAggFunctionType, nil
}

// ExecuteUserKPIForAttributionV1 Executes the KPI sub-query for Attribution
func (store *MemSQL) ExecuteUserKPIForAttributionV1(projectID int64, query *model.AttributionKPIQueries, from int64, to int64, debugQueryKey string,
	logCtx log.Entry, enableOptimisedFilterOnProfileQuery bool,
	enableOptimisedFilterOnEventUserQuery bool) (map[string]model.KPIInfo, []string, []string, error) {

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	kpiData := make(map[string]model.KPIInfo)
	var kpiKeys []string
	var kpiHeaders []string
	var kpiAggFunctionType []string
	var err error
	err, kpiKeys, kpiHeaders, kpiAggFunctionType = store.RunUserKPIGroupQueryV1(projectID, query, from, to, &kpiData, enableOptimisedFilterOnProfileQuery,
		enableOptimisedFilterOnEventUserQuery, debugQueryKey, logCtx)
	if err != nil {
		return kpiData, kpiHeaders, kpiAggFunctionType, err
	}
	logCtx.WithFields(log.Fields{"UserKPIAttribution": "Debug", "kpiData": kpiData,
		"kpiKeys": kpiKeys}).Info("UserKPI-Attribution kpiData reports after RunUserKPIGroupQuery")

	err = store.AddCoalUserIDinKPIData(&kpiData)
	if err != nil {
		return kpiData, kpiHeaders, kpiAggFunctionType, err
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.Info("done pulling group user list ids for Deal or Opportunity")
		logCtx.WithFields(log.Fields{"UserKPIAttribution": "Debug", "kpiData": kpiData,
			"kpiKeys": kpiKeys}).Info("UserKPI-Attribution kpiData reports 1")
	}
	err = store.PullAllUsersByCustomerUserID(projectID, &kpiData, logCtx)
	if err != nil {
		return kpiData, kpiHeaders, kpiAggFunctionType, err
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"UserKPIAttribution": "Debug", "kpiData": kpiData,
			"kpiKeys": kpiKeys}).Info("UserKPI-Attribution kpiData reports 2")
	}

	err = store.AddUserIdInKPIDataWithoutCustomerUserId(&kpiData, logCtx)
	if err != nil {
		return kpiData, kpiHeaders, kpiAggFunctionType, err
	}
	logCtx.WithFields(log.Fields{"UserKPIAttribution": "Debug", "kpiData": kpiData,
		"kpiKeys": kpiKeys}).Info("UserKPI-Attribution kpiData reports 3")

	return kpiData, kpiHeaders, kpiAggFunctionType, nil
}

// RunUserKPIGroupQuery runs kpi group query and adds the result in kpiData
func (store *MemSQL) RunUserKPIGroupQuery(projectID int64, query *model.AttributionQuery, kpiData *map[string]model.KPIInfo,
	enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery bool, debugQueryKey string, logCtx log.Entry) (error, []string, []string, []string) {

	var kpiQueryResult model.QueryResult
	if query.AnalyzeType == model.AnalyzeTypeUserKPI {

		var duplicatedRequest model.KPIQueryGroup
		U.DeepCopy(&query.KPI, &duplicatedRequest)
		resultGroup, statusCode := store.ExecuteKPIQueryGroup(projectID, debugQueryKey,
			duplicatedRequest, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		log.WithFields(log.Fields{"ResultGroup": resultGroup, "Status": statusCode}).Info("UserKPI-Attribution result received")
		if statusCode != http.StatusOK {
			logCtx.WithField("err_code", statusCode).Error("failed to get userKPI result for attribution query")
			if statusCode == http.StatusPartialContent {
				return errors.New("failed to get userKPI result for attribution query - StatusPartialContent"), nil, nil, nil
			}
			return errors.New("failed to get userKPI result for attribution query"), nil, nil, nil
		}
		for _, res := range resultGroup {
			// Skip the datetime header and the other result is of format. ex. "headers": ["$hubspot_deal_hs_object_id", "Revenue", "Pipeline", ...],
			if res.Headers[0] == "datetime" {
				kpiQueryResult = res
				logCtx.WithFields(log.Fields{"KpiQueryResult": kpiQueryResult}).Info("UserKPI-Attribution result set")
				break
			}
		}
		if kpiQueryResult.Headers == nil || len(kpiQueryResult.Headers) == 0 {
			logCtx.Error("no-valid result for userKPI query")
			return errors.New("no-valid result for userKPI query"), nil, nil, nil
		}
		kpiKeys, kpiHeaders, kpiAggFunctionType := store.GetDataFromKPIResult(projectID, kpiQueryResult, kpiData, query, logCtx)
		return nil, kpiKeys, kpiHeaders, kpiAggFunctionType
	}
	return errors.New("not a valid type of query for userKPI Attribution"), nil, nil, nil
}

// RunUserKPIGroupQueryV1 runs kpi group query and adds the result in kpiData
func (store *MemSQL) RunUserKPIGroupQueryV1(projectID int64, query *model.AttributionKPIQueries, from int64, to int64, kpiData *map[string]model.KPIInfo,
	enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery bool, debugQueryKey string, logCtx log.Entry) (error, []string, []string, []string) {

	var kpiQueryResult model.QueryResult
	var kpiQueryResultWithTime model.QueryResult
	if query.AnalyzeType == model.AnalyzeTypeUserKPI {

		var duplicatedRequest model.KPIQueryGroup

		// Making sure the internal KPI group queries have same from to as parent attribution query
		for index := range query.KPI.Queries {
			query.KPI.Queries[index].From = from
			query.KPI.Queries[index].To = to
		}

		U.DeepCopy(&query.KPI, &duplicatedRequest)

		// Making sure the internal KPI group queries have same from to as parent attribution query
		for index := range duplicatedRequest.Queries {
			duplicatedRequest.Queries[index].From = from
			duplicatedRequest.Queries[index].To = to
		}
		resultGroup, statusCode := store.ExecuteKPIQueryGroup(projectID, debugQueryKey,
			duplicatedRequest, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		logCtx.WithFields(log.Fields{"ResultGroup": resultGroup, "Status": statusCode}).Info("UserKPI-Attribution result received")
		if statusCode != http.StatusOK {
			logCtx.WithField("err_code", statusCode).Error("failed to get userKPI result for attribution query")
			if statusCode == http.StatusPartialContent {
				return errors.New("failed to get userKPI result for attribution query - StatusPartialContent"), nil, nil, nil
			}
			return errors.New("failed to get userKPI result for attribution query"), nil, nil, nil
		}
		for _, res := range resultGroup {
			// Skip the datetime header and the other result is of format. ex. "headers": ["$hubspot_deal_hs_object_id", "Revenue", "Pipeline", ...],
			if res.Headers[0] == "datetime" {
				kpiQueryResultWithTime = res
				logCtx.WithFields(log.Fields{"kpiQueryResultWithTime": kpiQueryResultWithTime}).Info("UserKPI-Attribution result set")
			} else {
				kpiQueryResult = res
				logCtx.WithFields(log.Fields{"kpiQueryResult": kpiQueryResult}).Info("UserKPI-Attribution result set")
			}

		}
		if kpiQueryResult.Headers == nil || len(kpiQueryResult.Headers) == 0 {
			logCtx.Error("no-valid result for userKPI query")
			return errors.New("no-valid result for userKPI query"), nil, nil, nil
		}
		kpiKeys, kpiHeaders, kpiAggFunctionType := store.GetDataFromUserKPIResultV1(kpiQueryResult, kpiQueryResultWithTime, kpiData, from, to, logCtx)
		return nil, kpiKeys, kpiHeaders, kpiAggFunctionType
	}
	return errors.New("not a valid type of query for userKPI Attribution"), nil, nil, nil
}

// GetDataFromUserKPIResult adds values in kpiData from kpiQueryResult
func (store *MemSQL) GetDataFromUserKPIResult(kpiQueryResult model.QueryResult, kpiData *map[string]model.KPIInfo, query *model.AttributionQuery, logCtx log.Entry) ([]string, []string, []string) {

	datetimeIdx := 0
	keyIdx := 1
	valIdx := 2

	kpiValueHeaderLength := len(kpiQueryResult.Headers) - valIdx
	kpiAggFunctionType := make([]string, kpiValueHeaderLength)
	var kpiValueHeaders []string
	for idx := valIdx; idx < len(kpiQueryResult.Headers); idx++ {
		kpiValueHeaders = append(kpiValueHeaders, kpiQueryResult.Headers[idx])
	}

	if len(kpiValueHeaders) == 0 {
		return nil, nil, nil
	}

	for i := range kpiValueHeaders {
		kpiAggFunctionType[i] = "unique"
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"kpiValueHeaders": kpiValueHeaders}).Info("KPI-Attribution headers set")
	}
	kpiKeys := model.AddKPIKeyDataInMap(kpiQueryResult, logCtx, keyIdx, datetimeIdx, query.From, query.To, valIdx, kpiValueHeaders, kpiAggFunctionType, kpiData)
	return kpiKeys, kpiValueHeaders, kpiAggFunctionType
}

// GetDataFromUserKPIResultV1 adds values in kpiData from kpiQueryResult
func (store *MemSQL) GetDataFromUserKPIResultV1(kpiQueryResult model.QueryResult, kpiQueryResultWithTime model.QueryResult, kpiData *map[string]model.KPIInfo, from int64, to int64, logCtx log.Entry) ([]string, []string, []string) {

	keyIdx := 0
	valIdx := 1

	kpiValueHeaderLength := len(kpiQueryResult.Headers) - valIdx
	kpiAggFunctionType := make([]string, kpiValueHeaderLength)
	var kpiValueHeaders []string
	for idx := valIdx; idx < len(kpiQueryResult.Headers); idx++ {
		kpiValueHeaders = append(kpiValueHeaders, kpiQueryResult.Headers[idx])
	}

	if len(kpiValueHeaders) == 0 {
		return nil, nil, nil
	}

	for i := range kpiValueHeaders {
		kpiAggFunctionType[i] = "unique"
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"kpiValueHeaders": kpiValueHeaders}).Info("KPI-Attribution headers set")
	}
	kpiKeys := model.AddUserKPIKeyDataInMap(kpiQueryResult, kpiQueryResultWithTime, logCtx, keyIdx, from, to, valIdx, kpiValueHeaders, kpiAggFunctionType, kpiData)
	return kpiKeys, kpiValueHeaders, kpiAggFunctionType
}

func (store *MemSQL) AddCoalUserIDinKPIData(kpiData *map[string]model.KPIInfo) error {

	for kpiID, kpiInfo := range *kpiData {
		var coalUsers []string
		coalUsers = append(coalUsers, kpiID)
		kpiInfo.KpiCoalUserIds = coalUsers
		(*kpiData)[kpiID] = kpiInfo
	}
	return nil
}
