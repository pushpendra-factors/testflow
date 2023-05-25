package memsql

import (
	"database/sql"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// RunSixSignalGroupQuery gets the query and the projectID, returns the resultGroup after fetching the results.
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

// runSingleSixSignalQuery calls the wrapper method that builds and executes the sixsignal query.
// After fetching the results, it calls the method to build the error result on basis of errMsg if any err is present.
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

// ExecuteSixSignalQuery validates the sixsignal query properties and returns the result from RunSixSignalInsightsQuery.
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

// RunSixSignalInsightsQuery calls the method to build the sixsignal SQL query and after successful creation of sql statement and params
// it calls the method to execute the query.
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

/*
- buildSixSignalQuery takes two parameters: a project ID and a SixSignalQuery object.
- The project ID is used to filter events related to a specific project, while the SixSignalQuery object contains the time range of the events to be considered.
- The query retrieves the following user properties and behaviors:
  - $6Signal_name, $6Signal_country, $6Signal_industry, $6Signal_employee_range, $6Signal_revenue_range, $session_spent_time,
    $page_count, $6Signal_domain, $initial_page_url, $campaign, $channel.

- The query also uses the ROW_NUMBER() function to assign a row number to each user based on their time spent partitioned by company,
and then selects only the row with the highest time spent for each company.

- The previous SQL query  have been using a JOIN operation to retrieve data from multiple tables, which could have led to slower query
execution times due to the amount of data being joined together. Additionally, the JOIN operation may have caused data redundancy issues
where the same data was being repeated in multiple rows, leading to inefficient use of storage space.

- The new SQL query appears to address these issues by using a sub-query to retrieve data from a single table and then applying a
window function to remove any duplicate data based on a specific ordering criteria. This approach should lead to faster query
execution times since it is only querying one table and should avoid data redundancy issues since duplicate data is being removed.
*/
func (store *MemSQL) buildSixSignalQuery(projectID int64, query model.SixSignalQuery) (string, []interface{}, error) {

	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	caseSelectStmntEventProperties := " CASE WHEN JSON_EXTRACT_STRING(events.properties, ?) IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(events.properties, ?) = '' THEN ? ELSE JSON_EXTRACT_STRING(events.properties, ?) END "

	caseSelectStmntUserProperties := " CASE WHEN JSON_EXTRACT_STRING(events.user_properties, ?) IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(events.user_properties, ?) = '' THEN ? ELSE JSON_EXTRACT_STRING(events.user_properties, ?) END "

	eventNameID := " SELECT id FROM event_names WHERE project_id=? AND name='$session' "

	subQuery := " SELECT JSON_EXTRACT_STRING(events.user_properties, ?) AS company, " +
		caseSelectStmntEventProperties + " AS time_spent, " +
		"JSON_EXTRACT_BIGINT(events.user_properties, ?) AS page_count, " +
		caseSelectStmntUserProperties + " AS country, " +
		caseSelectStmntUserProperties + " AS industry, " +
		caseSelectStmntUserProperties + " AS emp_range, " +
		caseSelectStmntUserProperties + " AS revenue_range, " +
		caseSelectStmntUserProperties + " AS domain, " +
		caseSelectStmntUserProperties + " AS page_seen, " +
		caseSelectStmntEventProperties + " AS campaign, " +
		caseSelectStmntEventProperties + " AS channel " +
		" FROM events " + " WHERE project_id=? AND timestamp >= ? AND timestamp <= ? AND events.event_name_id IN ( " + eventNameID + " ) "

	rowNumAssigned := " SELECT company, country, industry, emp_range, revenue_range, time_spent, page_count, domain, page_seen, campaign, channel, " +
		" ROW_NUMBER() OVER(PARTITION BY company ORDER BY time_spent DESC) as row_num FROM ( " + subQuery + " ) "

	qStmnt = " SELECT company, country, industry, emp_range, revenue_range, time_spent, page_count, domain, page_seen, campaign, channel " +
		" FROM (" + rowNumAssigned + " ) WHERE row_num=1 ORDER BY page_count DESC; "

	qParams = append(qParams, U.SIX_SIGNAL_NAME,
		U.SP_SPENT_TIME, model.PropertyValueZero, U.SP_SPENT_TIME, model.PropertyValueZero, U.SP_SPENT_TIME,
		U.UP_PAGE_COUNT,
		U.SIX_SIGNAL_COUNTRY, model.PropertyValueNone, U.SIX_SIGNAL_COUNTRY, model.PropertyValueNone, U.SIX_SIGNAL_COUNTRY,
		U.SIX_SIGNAL_INDUSTRY, model.PropertyValueNone, U.SIX_SIGNAL_INDUSTRY, model.PropertyValueNone, U.SIX_SIGNAL_INDUSTRY,
		U.SIX_SIGNAL_EMPLOYEE_RANGE, model.PropertyValueNone, U.SIX_SIGNAL_EMPLOYEE_RANGE, model.PropertyValueNone, U.SIX_SIGNAL_EMPLOYEE_RANGE,
		U.SIX_SIGNAL_REVENUE_RANGE, model.PropertyValueNone, U.SIX_SIGNAL_REVENUE_RANGE, model.PropertyValueNone, U.SIX_SIGNAL_REVENUE_RANGE,
		U.SIX_SIGNAL_DOMAIN, model.PropertyValueNone, U.SIX_SIGNAL_DOMAIN, model.PropertyValueNone, U.SIX_SIGNAL_DOMAIN,
		U.UP_INITIAL_PAGE_URL, model.PropertyValueNone, U.UP_INITIAL_PAGE_URL, model.PropertyValueNone, U.UP_INITIAL_PAGE_URL,
		U.EP_CAMPAIGN, model.PropertyValueNone, U.EP_CAMPAIGN, model.PropertyValueNone, U.EP_CAMPAIGN,
		U.EP_CHANNEL, model.PropertyValueNone, U.EP_CHANNEL, model.PropertyValueNone, U.EP_CHANNEL,
		projectID, query.From, query.To, projectID)

	log.WithFields(log.Fields{"SixSignalQuery": qStmnt, "Query parameters": qParams}).Info("Six Signal Query Statement and Parameters")

	return qStmnt, qParams, nil

}

// ExecSixSignalSQLQuery executes the SQL query and returns three values, a pointer to a SixSignalQueryResult struct, an error and a string representing a unique request ID.
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

// RunSixSignalPageViewQuery takes in two parameters projectId of type int64 and query of type model.SixSignalQuery.
// It returns a slice of strings, an integer and a string. Next, a SQL statement and its parameters are generated
// using the buildSixSignalPageViewQuery method. If the SQL statement and its parameters are generated successfully,
// the ExecSixSignalPageViewQuery method is called with the generated SQL statement and its parameters.
func (store *MemSQL) RunSixSignalPageViewQuery(projectId int64, query model.SixSignalQuery) ([]string, int, string) {
	logFields := log.Fields{
		"project_id": projectId,
		"query":      query,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	stmt, params := store.buildSixSignalPageViewQuery(projectId, query)

	logCtx := log.WithFields(logFields)
	if stmt == "" || len(params) == 0 {
		logCtx.Error("Failed generating SQL query.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	result, err, _ := store.ExecSixSignalPageViewQuery(stmt, params) //reqID marked as _
	if err != nil {
		logCtx.WithError(err).Error("Failed executing SQL query generated.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	return result, http.StatusOK, "Successfully executed query"

}

func (store *MemSQL) buildSixSignalPageViewQuery(projectId int64, query model.SixSignalQuery) (string, []interface{}) {
	logFields := log.Fields{
		"project_id": projectId,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	qParams := make([]interface{}, 0, 0)
	var qStmt string

	eventName, errCode := store.GetEventNamesByNames(projectId, query.PageView)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("failed to find event names")
		return qStmt, qParams
	}

	var eventNameId []string
	for i := range eventName {
		if eventName[i].Type == model.TYPE_AUTO_TRACKED_EVENT_NAME {
			eventNameId = append(eventNameId, eventName[i].ID)
		}
	}

	qStmt = "SELECT DISTINCT JSON_EXTRACT_STRING(events.user_properties,?) AS Company " + " FROM events " +
		" WHERE events.event_name_id IN ( ?) AND project_id=? AND timestamp>= ? AND timestamp<= ? ;"

	qParams = append(qParams, U.SIX_SIGNAL_NAME, eventNameId, projectId, query.From, query.To)

	return qStmt, qParams
}

func (store *MemSQL) ExecSixSignalPageViewQuery(stmt string, params []interface{}) ([]string, error, string) {
	logFields := log.Fields{
		"stmt":   stmt,
		"params": params,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	logCtx := log.WithFields(logFields)
	rows, tx, err, reqID := store.ExecQueryWithContext(stmt, params)
	if err != nil {
		return nil, err, reqID
	}

	defer U.CloseReadQuery(rows, tx)
	var companyList []string
	for rows.Next() {
		var companyTmp sql.NullString
		err = rows.Scan(&companyTmp)
		if err != nil {
			logCtx.WithError(err).Error("Error while scanning row")
			continue
		}
		if !companyTmp.Valid {
			continue
		}

		var company string
		company = companyTmp.String
		companyList = append(companyList, company)

	}

	return companyList, nil, reqID
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

// SendSixSignalReportViaEmail send email to the given list of emailIds along with info provided in the request Payload
func SendSixSignalReportViaEmail(requestPayload model.SixSignalEmailAndMessage) (string, []string) {

	fromDate := U.GetDateFromTimestampAndTimezone(requestPayload.From, requestPayload.Timezone)
	toDate := U.GetDateFromTimestampAndTimezone(requestPayload.To, requestPayload.Timezone)

	var success, fail int
	sendEmailFailIds := make([]string, 0)
	sub := "Latest accounts that visited " + requestPayload.Domain + " from " + fromDate + " to " + toDate
	html := U.GetSixSignalReportSharingTemplate(requestPayload.Url, requestPayload.Domain)
	text := ""
	for _, email := range requestPayload.EmailIDs {
		err := C.GetServices().Mailer.SendMail(email, C.GetFactorsSenderEmail(), sub, html, text)
		if err != nil {
			fail++
			sendEmailFailIds = append(sendEmailFailIds, email)
			log.WithError(err).Error("failed to send email via SendSixSignalReportViaEmail method")
			continue
		}
		success++
	}
	var msg string
	if success < len(requestPayload.EmailIDs) {
		msg = fmt.Sprintf("Email successfully sent to %d email id, failed to send email to %d", success, fail)
	} else {
		msg = "Email successfully sent to all the email ids"
	}

	return msg, sendEmailFailIds
}
