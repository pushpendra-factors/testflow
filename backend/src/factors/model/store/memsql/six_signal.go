package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

func (store *MemSQL) RunSixSignalGroupQuery(queriesOriginal []model.SixSignalQuery,
	projectId int64) (model.SixSignalResultGroup, int) {

	logFields := log.Fields{
		"queries_original": queriesOriginal,
		"project_id":       projectId,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	queries := make([]model.SixSignalQuery, 0, 0)
	U.DeepCopy(&queriesOriginal, &queries)

	var resultGroup model.SixSignalResultGroup
	resultGroup.Results = make([]model.SixSignalQueryResult, len(queries))
	var waitGroup sync.WaitGroup
	count := 0
	actualRoutineLimit := U.MinInt(len(queries), AllowedGoroutines)
	waitGroup.Add(actualRoutineLimit)
	for index, query := range queries {
		count++
		go store.runSingleSixSignalQuery(projectId, &resultGroup.Results[index], &waitGroup, query)
		if count%actualRoutineLimit == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(queries)-count, actualRoutineLimit))
		}
	}
	waitGroup.Wait()
	for _, result := range resultGroup.Results {
		if result.Headers == nil {
			log.Error("Header in nil")
			return resultGroup, http.StatusInternalServerError
		}
		if result.Headers[0] == model.AliasError {
			return resultGroup, http.StatusPartialContent
		}
	}

	return resultGroup, http.StatusOK

}

func (store *MemSQL) runSingleSixSignalQuery(projectId int64,
	resultHolder *model.SixSignalQueryResult, waitGroup *sync.WaitGroup, query model.SixSignalQuery) {

	logFields := log.Fields{

		"project_id": projectId,
		"wait_group": waitGroup,
	}
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	defer waitGroup.Done()
	result, errCode, errMsg := store.ExecuteSixSignalQuery(projectId, query)
	if errCode != http.StatusOK {
		errorResult := buildSixSignalErrorResult(errMsg)
		*resultHolder = *errorResult
	} else {
		*resultHolder = *result

	}
	return
}

func (store *MemSQL) ExecuteSixSignalQuery(projectId int64, query model.SixSignalQuery) (*model.SixSignalQueryResult, int, string) {

	logFields := log.Fields{
		"query":      query,
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	if errMsg, hasError := validateSixSignalQueryProps(&query); hasError {
		return nil, http.StatusBadRequest, errMsg
	}
	return store.RunSixSignalInsightsQuery(projectId, query)
}

func (store *MemSQL) RunSixSignalInsightsQuery(projectId int64, query model.SixSignalQuery) (*model.SixSignalQueryResult, int, string) {
	logFields := log.Fields{
		"query":      query,
		"project_id": projectId,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	stmnt, params, err := store.buildSixSignalQuery(projectId, query)
	if err != nil {
		log.WithError(err).Error(model.ErrMsgQueryProcessingFailure)
		return &model.SixSignalQueryResult{}, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	logCtx := log.WithFields(logFields)
	if stmnt == "" || len(params) == 0 {
		logCtx.Error("Failed generating SQL query from analytics query.")
		return &model.SixSignalQueryResult{}, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	result, err, _ := store.ExecSixSignalSQLQuery(stmnt, params) //reqID marked as _
	if err != nil {
		logCtx.WithError(err).Error("Failed executing SQL query generated.")
		return &model.SixSignalQueryResult{}, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	return result, http.StatusOK, "Successfully executed query"
}

func (store *MemSQL) buildSixSignalQuery(projectID int64, query model.SixSignalQuery) (string, []interface{}, error) {

	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	caseSelectStmntEventProperties := "CASE WHEN JSON_EXTRACT_STRING(events.properties, ?) IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(events.properties, ?) = '' THEN ? ELSE JSON_EXTRACT_STRING(events.properties, ?) END "

	caseSelectStmntUserProperties := "CASE WHEN JSON_EXTRACT_STRING(events.user_properties, ?) IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(events.user_properties, ?) = '' THEN ? ELSE JSON_EXTRACT_STRING(events.user_properties, ?) END "

	eventNameID := " SELECT id FROM event_names WHERE project_id=? AND name='$session' "

	maxSessionTimeQuery := "SELECT JSON_EXTRACT_STRING(events.user_properties,?) AS company, " +
		caseSelectStmntEventProperties + " AS time_spent FROM events "

	maxSessionTimeStmnt := maxSessionTimeQuery + "WHERE project_id=? AND timestamp >= ? AND timestamp <= ? " +
		" AND events.event_name_id IN ( " + eventNameID + " ) " + " AND company IS NOT NULL " +
		" GROUP BY company "

	qParams = append(qParams, U.SIX_SIGNAL_NAME,
		U.SP_SPENT_TIME, model.PropertyValueZero, U.SP_SPENT_TIME, model.PropertyValueZero, U.SP_SPENT_TIME,
		projectID, query.From, query.To, projectID)

	sixSignalPropertiesQuery := "SELECT JSON_EXTRACT_STRING(events.user_properties,?) AS company, " +
		"JSON_EXTRACT_BIGINT(events.user_properties,?) AS page_count, " +
		caseSelectStmntEventProperties + "AS time_spent, " +
		caseSelectStmntUserProperties + "AS country, " +
		caseSelectStmntUserProperties + "AS industry, " +
		caseSelectStmntUserProperties + "AS emp_range, " +
		caseSelectStmntUserProperties + "AS revenue_range, " +
		caseSelectStmntUserProperties + "AS domain, " +
		caseSelectStmntUserProperties + "AS page_seen, " +
		caseSelectStmntEventProperties + "AS campaign, " +
		caseSelectStmntEventProperties + "AS channel " +
		"FROM events "

	sixSignalPropertiesStmnt := sixSignalPropertiesQuery + " WHERE project_id=? AND timestamp >= ? AND timestamp <= ?" +
		" AND events.event_name_id IN ( " + eventNameID + " )"

	qParams = append(qParams, U.SIX_SIGNAL_NAME, U.UP_PAGE_COUNT,
		U.SP_SPENT_TIME, model.PropertyValueZero, U.SP_SPENT_TIME, model.PropertyValueZero, U.SP_SPENT_TIME,
		U.SIX_SIGNAL_COUNTRY, model.PropertyValueNone, U.SIX_SIGNAL_COUNTRY, model.PropertyValueNone, U.SIX_SIGNAL_COUNTRY,
		U.SIX_SIGNAL_INDUSTRY, model.PropertyValueNone, U.SIX_SIGNAL_INDUSTRY, model.PropertyValueNone, U.SIX_SIGNAL_INDUSTRY,
		U.SIX_SIGNAL_EMPLOYEE_RANGE, model.PropertyValueNone, U.SIX_SIGNAL_EMPLOYEE_RANGE, model.PropertyValueNone, U.SIX_SIGNAL_EMPLOYEE_RANGE,
		U.SIX_SIGNAL_REVENUE_RANGE, model.PropertyValueNone, U.SIX_SIGNAL_REVENUE_RANGE, model.PropertyValueNone, U.SIX_SIGNAL_REVENUE_RANGE,
		U.SIX_SIGNAL_DOMAIN, model.PropertyValueNone, U.SIX_SIGNAL_DOMAIN, model.PropertyValueNone, U.SIX_SIGNAL_DOMAIN,
		U.UP_INITIAL_PAGE_URL, model.PropertyValueNone, U.UP_INITIAL_PAGE_URL, model.PropertyValueNone, U.UP_INITIAL_PAGE_URL,
		U.EP_CAMPAIGN, model.PropertyValueNone, U.EP_CAMPAIGN, model.PropertyValueNone, U.EP_CAMPAIGN,
		U.EP_CHANNEL, model.PropertyValueNone, U.EP_CHANNEL, model.PropertyValueNone, U.EP_CHANNEL,
		projectID, query.From, query.To, projectID)

	selectStmnt := "SELECT t1.company, t2.country, t2.industry, t2.emp_range, t2.revenue_range, t1.time_spent, t2.page_count, t2.domain, t2.page_seen, t2.campaign, t2.channel " + "FROM "

	qStmnt = selectStmnt + "( " + maxSessionTimeStmnt + " ) AS t1 " + " JOIN " + "( " + sixSignalPropertiesStmnt + " ) AS t2 " +
		"ON t1.company=t2.company " +
		"AND t1.time_spent=t2.time_spent " +
		"ORDER BY t2.page_count DESC; "

	log.WithFields(log.Fields{"SixSignalQuery": qStmnt, "Query parameters": qParams}).Info("Six Signal Query Statement and Parameters")

	return qStmnt, qParams, nil

}

func (store *MemSQL) ExecSixSignalSQLQuery(stmnt string, params []interface{}) (*model.SixSignalQueryResult, error, string) {
	logFields := log.Fields{
		"stmnt":  stmnt,
		"params": params,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	rows, tx, err, reqID := store.ExecQueryWithContext(stmnt, params)
	if err != nil {
		return nil, err, reqID
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows, tx, reqID)
	if err != nil {
		return nil, err, reqID
	}

	result := &model.SixSignalQueryResult{Headers: resultHeaders, Rows: resultRows}
	return result, nil, reqID
}

// buildSixSignalErrorResult takes the failure msg and wraps it into a model.SixSignalQueryResult object
func buildSixSignalErrorResult(errMsg string) *model.SixSignalQueryResult {
	logFields := log.Fields{
		"err_msg": errMsg,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	errMsg = "Query failed:" + " - " + errMsg
	headers := []string{model.AliasError}
	rows := make([][]interface{}, 0, 0)
	row := make([]interface{}, 0, 0)
	row = append(row, errMsg)
	rows = append(rows, row)
	errorResult := &model.SixSignalQueryResult{Headers: headers, Rows: rows}
	return errorResult
}

func validateSixSignalQueryProps(query *model.SixSignalQuery) (string, bool) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if query.From == 0 || query.To == 0 {
		return "Invalid query time range", true
	}

	return "", false
}
