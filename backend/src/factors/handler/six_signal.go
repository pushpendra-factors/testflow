package handler

import (
	"encoding/json"
	C "factors/config"
	"factors/integration/six_signal"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	"factors/model/store/memsql"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	INVALID_PROJECT = "INVALID PROJECT"
	INVALID_INPUT   = "INVALID_INPUT"
)

// GetSixSignalReportHandler fetches the saved sixsignal report from cloud storage if the isSaved parameter in request payload is true
// if the isSaved parameter is false the handler computes the result on the go.
// The report fetched from the cloud are allowed to share and the result computed on the go is not allowed to share which is reflected
// in the response parameter isShareable.
func GetSixSignalReportHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	r := c.Request

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		log.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, "Query failed. Invalid project.", true
	}

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
	})

	var requestPayload model.SixSignalQueryGroup

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
	}

	if len(requestPayload.Queries) == 0 {
		logCtx.Error("Query failed. Empty query group.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Empty query group.", true
	}

	pageView := requestPayload.Queries[0].PageView

	result := make(map[int]model.SixSignalResultGroup)
	if requestPayload.Queries[0].IsSaved == true {

		folderName := getFolderName(requestPayload.Queries[0])
		logCtx.WithFields(log.Fields{"folder name": folderName}).Info("Folder name for reading the result")

		result = GetSixSignalAnalysisData(projectId, folderName)
		if result == nil {
			logCtx.Error("Report is not present for this date range")
			return result, http.StatusBadRequest, "", "Report is not present for this date range", true
		} else if len(result[1].Results[0].Rows) == 0 {
			logCtx.Warn("Data is not present for this date range")
			return result, http.StatusOK, "", "Data is not present for this date range", false
		}

		if len(pageView) > 0 && pageView != nil {
			res, err := FilterRowsForSixSignalPageView(projectId, requestPayload.Queries[0], result[1].Results[0])
			if err != http.StatusOK {
				return nil, err, "", "Failed fetching 6Signal company name for page view filter", false
			}
			result[1].Results[0] = res
		}
	} else {

		resultGroup, errCode := store.GetStore().RunSixSignalGroupQuery(requestPayload.Queries, projectId)
		if errCode != http.StatusOK {
			logCtx.WithField("err_code", errCode).Error("Six signal group query failed on report handler")
			return nil, http.StatusInternalServerError, "", "Failed to process Query", true
		}
		resultGroup.Query = requestPayload
		resultGroup.IsShareable = false

		//Adding cache meta to the result group
		meta := model.CacheMeta{
			Timezone:       string(requestPayload.Queries[0].Timezone),
			From:           requestPayload.Queries[0].From,
			To:             requestPayload.Queries[0].To,
			LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		}
		resultGroup.CacheMeta = meta

		result[1] = resultGroup
		if len(pageView) > 0 && pageView != nil {
			res, err := FilterRowsForSixSignalPageView(projectId, requestPayload.Queries[0], result[1].Results[0])
			if err != http.StatusOK {
				return nil, err, "", "Failed fetching 6Signal company name for page view filter", false
			}
			result[1].Results[0] = res
		}
	}
	return result, http.StatusOK, "", "", false
}

// GetSixSignalPublicReportHandler fetches the sixsignal report from cloud storage for public URLs
func GetSixSignalPublicReportHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	r := c.Request

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		log.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, "Query failed. Invalid project.", true
	}

	queryID := c.Query("query_id")

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
		"queryID":    queryID,
	})

	// reqPayload will contain the page_url filters list applied on the public url.
	var reqPayload model.SixSignalQuery
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&reqPayload); err != nil {
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
	}

	share, errCode := store.GetStore().GetShareableURLWithShareStringWithLargestScope(projectId, queryID, model.ShareableURLEntityTypeSixSignal)
	if errCode != http.StatusFound {
		logCtx.Error("Failed fetching Shareable URLs in GetSixSignalPublicReportHandler with errCode: ", errCode)
		return nil, http.StatusNotFound, "", "No Shareable URLs found", true
	}

	query, err := store.GetStore().GetSixSignalQueryWithQueryId(projectId, share.EntityID)
	if err != http.StatusFound {
		logCtx.Error("Failed fetching queries in GetSixSignalPublicReportHandler with errCode: ", errCode)
		return nil, http.StatusNotFound, "", "No Query found", true
	}

	var sixSignalQuery model.SixSignalQuery
	err1 := json.Unmarshal(query.Query.RawMessage, &sixSignalQuery)
	if err1 != nil {
		logCtx.WithError(err1).Error("Failed to unmarshal query in GetSixSignalPublicReportHandler with error: ", err1)
		return nil, http.StatusNotFound, "", "Failed to unmarshal query", true
	}

	pageView := reqPayload.PageView
	folderName := getFolderName(sixSignalQuery)
	logCtx.WithFields(log.Fields{"folder name": folderName}).Info("Folder name for reading the result")

	result := GetSixSignalAnalysisData(projectId, folderName)
	if result == nil {
		logCtx.Error("Report is not present for this date range")
		return result, http.StatusBadRequest, "", "Report is not present for this date range", true
	} else if len(result[1].Results[0].Rows) == 0 {
		logCtx.Warn("Data is not present for this date range")
		return result, http.StatusOK, "", "Data is not present for this date range", false
	}

	if len(pageView) > 0 && pageView != nil {
		res, err := FilterRowsForSixSignalPageView(projectId, sixSignalQuery, result[1].Results[0])
		if err != http.StatusOK {
			return nil, err, "", "Failed fetching 6Signal company name for page view filter", false
		}
		result[1].Results[0] = res
	}
	return result, http.StatusOK, "", "", false

}

// CreateSixSignalShareableURLHandler saves the query to the queries table and generate a queryID for shareable URL
func CreateSixSignalShareableURLHandler(c *gin.Context) (interface{}, int, string, bool) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Six Signal Shareable query creation failed. Invalid project."})
		return nil, http.StatusForbidden, "Create SixSignal Shareable URL Failed. Invalid project", true
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	logCtx := log.WithFields(log.Fields{
		"reqId":         U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"loggedInAgent": agentUUID,
		"project_id":    projectID,
	})

	logCtx.Info("Six Signal report access is being changed to public by agent: ", agentUUID)

	params := model.SixSignalShareableURLParams{}
	err := c.BindJSON(&params)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse sixsignal shareable url request body")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Six Signal Shareable query creation failed. Invalid params."})
		return nil, http.StatusBadRequest, "Six Signal Shareable query creation failed. Invalid params.", true
	}

	//Getting sixSignalQuery struct to create data for sha256 encryption
	var sixSignalQuery model.SixSignalQuery
	err1 := json.Unmarshal(params.Query.RawMessage, &sixSignalQuery)
	if err1 != nil {
		logCtx.Error("Cannot Unmarshal SixSignalQueryGroup json in CreateSixSignalShareableURLHandler with error: ", err1)
		return nil, http.StatusBadRequest, "Cannot Unmarshal SixSignalQueryGroup json in CreateSixSignalShareableURLHandler", true
	}

	//Checking if report is present for this date range
	folderName := getFolderName(sixSignalQuery)
	result := GetSixSignalAnalysisData(projectID, folderName)
	if result == nil {
		logCtx.Error("Report is not present for this date range")
		return nil, http.StatusBadRequest, "Report is not present for this date range", true
	}

	data := fmt.Sprintf("%d%s%d%d", projectID, sixSignalQuery.Timezone, sixSignalQuery.From, sixSignalQuery.To)

	queryRequest := &model.Queries{
		Query:     *params.Query,
		Title:     "Six Signal Report",
		CreatedBy: agentUUID,
		Settings:  postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		IdText:    U.HashKeyUsingSha256Checksum(data),
		Type:      model.QueryTypeSixSignalQuery,
	}

	queryId, errCode, errMsg := six_signal.CreateSixSignalShareableURL(queryRequest, projectID, agentUUID)
	if errCode != http.StatusCreated {
		logCtx.Error(errMsg)
		return nil, errCode, errMsg, true
	}

	response := model.SixSignalPublicURLResponse{
		QueryID:      queryId,
		RouteVersion: ROUTE_VERSION_V1_WITHOUT_SLASH,
	}

	return response, http.StatusCreated, "Shareable Query creation successful", false
}

// SendSixSignalReportViaEmailHandler SendSixSignalReportViaEmail sends mail to the emailIDs provided by clients
func SendSixSignalReportViaEmailHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	r := c.Request

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		log.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, "Invalid Project", true
	}

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
	})

	var requestPayload model.SixSignalEmailAndMessage
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Json decode failed in method SendSixSignalReportViaEmail.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Json Decode Failed", true
	}

	if len(requestPayload.EmailIDs) == 0 {
		logCtx.Error("No email id present to send mail for SixSignal Report")
		return nil, http.StatusBadRequest, INVALID_INPUT, "No email id provided", true
	}

	msg, _ := memsql.SendSixSignalReportViaEmail(requestPayload)

	return msg, http.StatusOK, "", "", false

}

// AddSixSignalEmailIDHandler adds emailIDs provided by clients to the DB
func AddSixSignalEmailIDHandler(c *gin.Context) (interface{}, int, string, bool) {
	r := c.Request
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		log.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, "Invalid Project", true
	}

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
	})

	var requestPayload model.SixSignalEmailAndMessage
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Json decode failed in method AddSixSignalEmailIDHandler.")
		return nil, http.StatusBadRequest, "Json Decode Failed", true
	}
	if len(requestPayload.EmailIDs) == 0 {
		logCtx.Error("No email id present to send mail for SixSignal Report")
		return nil, http.StatusBadRequest, "No email id provided", true
	}

	emailIdsToAdd := strings.Join(requestPayload.EmailIDs, ",")

	emailIds, errCode := store.GetStore().GetSixsignalEmailListFromProjectSetting(projectId)
	if errCode == http.StatusInternalServerError {
		logCtx.Warn("Could not find emailids from sixsignal_email_list table")
		return nil, http.StatusInternalServerError, "Could not find emailids from sixsignal_email_list table", true
	}

	var emailIdFinalList string
	if len(emailIds) > 0 {
		emailIdFinalList = emailIds + "," + emailIdsToAdd
	} else {
		emailIdFinalList = emailIdsToAdd
	}

	errCode1 := store.GetStore().AddSixsignalEmailList(projectId, emailIdFinalList)
	if errCode1 != http.StatusCreated {
		logCtx.Error("Failed to add data in sixsignal_email_list")
		return nil, errCode1, "Failed to add data in sixsignal_email_list", true
	}

	return "EmailID added successfully", http.StatusCreated, "", false
}

// FetchListofDatesForSixSignalReport fetches the list of dates for which the report is present in cloud storage
func FetchListofDatesForSixSignalReport(c *gin.Context) (interface{}, int, string, bool) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		log.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, "Invalid Project", true
	}

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
	})

	timezoneString, statusCode := store.GetStore().GetTimezoneForProject(projectId)
	if statusCode != http.StatusFound {
		logCtx.Error("Failed to get Timezone in FetchListofDatesForSixSignalReport", statusCode)
		return nil, http.StatusBadRequest, "Query failed. Failed to get Timezone.", true
	}

	path := fmt.Sprintf("projects/%d/sixSignal", projectId) //path= "projects/2/sixSignal"
	cloudManager := C.GetCloudManager()
	//filenames contains the complete path for the reports file
	//filenames=["projects/2/sixSignal/20230212-20230219/results.txt","projects/2/sixSignal/20230220-20230227/results.txt",...]
	filenames := cloudManager.ListFiles(path)

	//dateList will contain the from-to values for all the sixsignal reports presents for a particular project on cloud storage.
	dateList := make([]string, 0)

	//In this loop, the dates(from-to) present in YYYYMMDD format is extracted from filenames using string slicing and then these from and to values are converted
	//into epoch values based on the timezone. The epoch values of from and to are merged using a hyphen and then append to the dateList.
	for _, filename := range filenames {
		from := filename[len(path)+1 : len(path)+9]
		to := filename[len(path)+10 : len(path)+18]

		fromEpoch := U.GetBeginningoftheDayEpochForDateAndTimezone(from, string(timezoneString))
		toEpoch := U.GetEndoftheDayEpochForDateAndTimezone(to, string(timezoneString))

		dateRange := fmt.Sprintf("%d-%d", fromEpoch, toEpoch) //dateRange="1676140200-1676744999"
		dateList = append(dateList, dateRange)                //dateList=["1676140200-1676744999", "1676745000-1677349799",...]
	}

	return dateList, http.StatusFound, "", false
}

// GetPageViewForSixSignalReport fetches the page_view event from the DB for both app-login and public-url.
func GetPageViewForSixSignalReport(c *gin.Context) (interface{}, int, string, string, bool) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "Invalid project.", true
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	eventNames, err := store.GetStore().GetMostFrequentlyEventNamesByType(projectId, model.FilterValuesOrEventNamesLimit, C.GetLookbackWindowForEventUserCache(), "page_views")
	if err != nil {
		logCtx.WithError(err).Error("get event names ordered by occurence and recency")
		return nil, http.StatusInternalServerError, "", "Failed fetching page views", true
	}

	if len(eventNames) == 0 {
		logCtx.WithError(err).Error(fmt.Sprintf("No Events Returned - ProjectID - %d", projectId))
	}

	// Force add specific events.
	if fNames, pExists := FORCED_EVENT_NAMES[projectId]; pExists {
		eventNames = append(eventNames, fNames...)
	}

	return eventNames, http.StatusOK, "", "", false
}

// getFolderName generate folder name using from, to and timezone from sixsignal query
func getFolderName(query model.SixSignalQuery) string {
	commonQueryFrom := query.From
	commonQueryTo := query.To
	timezoneString := query.Timezone

	fromDate := U.GetDateOnlyFormatFromTimestampAndTimezone(commonQueryFrom, timezoneString)
	toDate := U.GetDateOnlyFormatFromTimestampAndTimezone(commonQueryTo, timezoneString)
	folderName := fmt.Sprintf("%v-%v", fromDate, toDate)
	return folderName
}

// GetSixSignalAnalysisData fetches the sixsignal report cloud storage path and reads the report file.
func GetSixSignalAnalysisData(projectId int64, id string) map[int]model.SixSignalResultGroup {

	cloudManager := C.GetCloudManager()
	path, _ := cloudManager.GetSixSignalAnalysisTempFilePathAndName(id, projectId)
	fmt.Println(path)
	reader, err := cloudManager.Get(path, "result.txt")
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file from Cloud Manager for projectid: ", projectId)
		return nil
	}
	result := make(map[int]model.SixSignalResultGroup)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file from ioutil for projectid: ", projectId)
		return nil
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.WithError(err).Error("Error in unmarshal JSON in GetSixSignalAnalysisData for projectId: ", projectId)
		return nil
	}

	return result
}

func FilterRowsForSixSignalPageView(projectId int64, reqPayload model.SixSignalQuery, result model.SixSignalQueryResult) (model.SixSignalQueryResult, int) {

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
		"reqPayload": reqPayload,
	})

	rows := result.Rows
	filteredRes := model.SixSignalQueryResult{Headers: result.Headers,
		Query: result.Query}

	companyList, err, _ := store.GetStore().RunSixSignalPageViewQuery(projectId, reqPayload)
	if err != http.StatusOK {
		logCtx.Error("Failed fetching 6Signal company name for page view filter")
		return filteredRes, err
	}

	companyMap := make(map[string]bool)
	for _, company := range companyList {
		companyMap[company] = true
	}

	var filteredRows [][]interface{}
	for _, row := range rows {
		if _, ok := companyMap[row[0].(string)]; ok {
			filteredRows = append(filteredRows, row)
		}
	}

	sortIndex := getIndexForPageCount(filteredRes.Headers)
	if sortIndex != -1 {
		sort.Slice(filteredRows, func(i, j int) bool {
			countI, _ := filteredRows[i][sortIndex].(int)
			countJ, _ := filteredRows[j][sortIndex].(int)
			return countI > countJ
		})
	}
	filteredRes.Rows = filteredRows
	return filteredRes, http.StatusOK
}

func getIndexForPageCount(headers []string) int {
	for i, v := range headers {
		if v == "page_count" {
			return i
		}
	}
	return -1
}
