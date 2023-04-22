package delta

import (
	"bytes"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	"factors/model/store/memsql"
	serviceDisk "factors/services/disk"
	T "factors/task"
	U "factors/util"
	"fmt"
	"github.com/jinzhu/gorm/dialects/postgres"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func SixSignalAnalysis(projectIdArray []int64, configs map[string]interface{}) interface{} {

	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	modelCloudManager := configs["modelCloudManager"].(*filestore.FileManager)
	errMsgToProjectIDMap := make(map[string][]int64)

	for _, projectId := range projectIdArray {

		logCtx := log.WithFields(log.Fields{
			"project_id": projectId,
		})

		//fetching timezone, from and to timestamp
		var requestPayload model.SixSignalQueryGroup
		timezone, _ := store.GetStore().GetTimezoneForProject(projectId)
		from, to, _ := U.GetQueryRangePresetLastWeekIn(timezone)

		requestPayload.Queries = make([]model.SixSignalQuery, 1)
		requestPayload.Queries[0].From = from
		requestPayload.Queries[0].To = to
		requestPayload.Queries[0].Timezone = timezone

		resultGroup, errCode := store.GetStore().RunSixSignalGroupQuery(requestPayload.Queries, projectId)
		if errCode != http.StatusOK {
			logCtx.Error("Query failed. Failed to process query from DB with error: ", errCode)
			errMsg := fmt.Sprintf("%v", errCode)
			errMsgToProjectIDMap[errMsg] = append(errMsgToProjectIDMap[errMsg], projectId)
			continue
		}
		resultGroup.Query = requestPayload

		//Adding cache meta to the result group
		meta := model.CacheMeta{
			Timezone:       string(timezone),
			From:           from,
			To:             to,
			LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		}
		resultGroup.CacheMeta = meta
		resultGroup.IsShareable = true

		fromDate := U.GetDateOnlyFormatFromTimestampAndTimezone(from, timezone)
		toDate := U.GetDateOnlyFormatFromTimestampAndTimezone(to, timezone)
		folderName := fmt.Sprintf("%v-%v", fromDate, toDate)
		logCtx.WithFields(log.Fields{"folder name": folderName}).Info("Folder name for saving the result")

		//Creating temp path for the result file
		sixSignalAnalysisTempPath, sixSignalAnalysisTempName := diskManager.GetSixSignalAnalysisTempFilePathAndName(folderName, projectId)
		log.Info("creating path analysis temp file. Path: ", sixSignalAnalysisTempPath, " Name: ", sixSignalAnalysisTempName)
		if err := os.MkdirAll(sixSignalAnalysisTempPath, os.ModePerm); err != nil {
			log.Fatal(err)
		}

		resultFile, err := os.Create(sixSignalAnalysisTempPath + sixSignalAnalysisTempName)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Failed creating sixSignalAnalysis temp file")
			errMsg := fmt.Sprintf("%v", err)
			errMsgToProjectIDMap[errMsg] = append(errMsgToProjectIDMap[errMsg], projectId)
			continue
		}

		pwmBytes, _ := json.Marshal(resultGroup)
		pString := string(pwmBytes)
		pString = pString + "\n"
		pBytes := []byte(pString)
		_, err = resultFile.Write(pBytes)
		if err != nil {
			logCtx.Error("Failed to write results in result file with err: ", err)
			errMsg := fmt.Sprintf("%v", err)
			errMsgToProjectIDMap[errMsg] = append(errMsgToProjectIDMap[errMsg], projectId)
			continue
		}

		err1 := WriteSixSignalResultsToCloud(modelCloudManager, diskManager, folderName, projectId, logCtx)
		if err1 != nil {
			logCtx.Error("Writing sixsignal results to cloud failed with err: ", err1)
			errMsg := fmt.Sprintf("Writing sixsignal results to cloud failed with errCode %v ", err1)
			errMsgToProjectIDMap[errMsg] = append(errMsgToProjectIDMap[errMsg], projectId)
			continue
		}
	}
	return errMsgToProjectIDMap

}

//SendSixSignalEmailForSubscribe sends mail to all the email for which the sixsignal-report is subscribed on a weekly basis. The list of email id is fetched from the DB.
func SendSixSignalEmailForSubscribe(projectIdArray []int64) interface{} {

	projectIdToFailSendEmailIdsMap := make(map[int64][]string)
	for _, projectId := range projectIdArray {

		logCtx := log.WithFields(log.Fields{
			"project_id": projectId,
		})

		//Changing all the report generated share type to public and generating public url
		//Fetching date range and timezone for generating hash key and query
		timezone, _ := store.GetStore().GetTimezoneForProject(projectId)
		from, to, _ := U.GetQueryRangePresetLastWeekIn(timezone)

		data := fmt.Sprintf("%d%s%d%d", projectId, timezone, from, to)
		query := model.SixSignalQuery{From: from, To: to, Timezone: timezone}
		queryJson, _ := json.Marshal(query)

		queryRequest := &model.Queries{
			Query:     postgres.Jsonb{RawMessage: json.RawMessage(queryJson)},
			Title:     "Six Signal Report",
			CreatedBy: "",
			Settings:  postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
			IdText:    U.HashKeyUsingSha256Checksum(data),
			Type:      model.QueryTypeSixSignalQuery,
		}

		queryId, errCode, errMsg := CreateSixSignalShareableURL(queryRequest, projectId, "")
		if errCode != http.StatusCreated {
			logCtx.Error(errMsg)
			continue
		}

		publicUrlParams := fmt.Sprintf("/reports/visitor_report?queryId=%v&pId=%d&version=v1", queryId, projectId)
		publicURL := C.GetProtocol() + C.GetAPPDomain() + publicUrlParams

		//Fetching emailIds from database and converting the datatype to array
		emailIdsString, errCode1 := store.GetStore().GetSixsignalEmailListFromProjectSetting(projectId)
		if errCode1 != http.StatusFound {
			logCtx.Error("No email Ids for sixsignal report subscription is found.")
			continue
		}

		emailIds := strings.Split(emailIdsString, ",")
		if len(emailIds) == 0 {
			logCtx.Warn("No email id present for subscribe feature")
			continue
		}

		project, _ := store.GetStore().GetProject(projectId)
		reqPayload := model.SixSignalEmailAndMessage{
			EmailIDs: emailIds,
			Url:      publicURL,
			Domain:   project.Name,
			From:     from,
			To:       to,
			Timezone: timezone,
		}

		_, failToSendEmailIds := memsql.SendSixSignalReportViaEmail(reqPayload)
		projectIdToFailSendEmailIdsMap[projectId] = failToSendEmailIds

	}

	return projectIdToFailSendEmailIdsMap
}

func WriteSixSignalResultsToCloud(cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, queryId string, projectId int64, logCtx *log.Entry) error {

	path, _ := diskManager.GetSixSignalAnalysisTempFilePathAndName(queryId, projectId)
	resultLocalPath := path + "results.txt"
	logCtx.Info("Path in WriteResultstoCloud: ", resultLocalPath)
	scanner, err := T.OpenEventFileAndGetScanner(resultLocalPath)
	if err != nil {
		logCtx.Error("Error in scanner: ", err)
	}
	result := make(map[int]model.SixSignalResultGroup)
	i := 0
	for scanner.Scan() {
		txtline := scanner.Text()
		logCtx.Info("TextLine: ", txtline)
		i++
		var events model.SixSignalResultGroup
		if err := json.Unmarshal([]byte(txtline), &events); err != nil {
			log.WithFields(log.Fields{"Error": err}).Warn("Cannot unmarshal JSON")
		}
		result[i] = events
	}
	logCtx.Info("Result after Unmarshal in WriteResultsToCloud: ", result)

	path, _ = (*cloudManager).GetSixSignalAnalysisTempFilePathAndName(queryId, projectId)
	resultJson, err := json.Marshal(result)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("failed to unmarshal result Info.")
		return err
	}

	err = (*cloudManager).Create(path, "result.txt", bytes.NewReader(resultJson))
	if err != nil {
		log.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return err
	}
	return err
}

//CreateSixSignalShareableURL saves the query to the queries table and generate the queryID for public-URL for the given queryRequest and projectId
func CreateSixSignalShareableURL(queryRequest *model.Queries, projectId int64, agentUUID string) (string, int, string) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
		"query":      queryRequest,
	})
	queries, errCode, errMsg := store.GetStore().CreateQuery(projectId, queryRequest)
	if errCode != http.StatusCreated {
		return "", errCode, errMsg
	}

	isShared, _ := isReportShared(projectId, queries.IdText)
	if isShared {

		logCtx.Info("Query Id if shared already: ", queries.IdText)
		errCode1, errMsg1 := store.GetStore().DeleteQuery(projectId, queries.ID)
		if errCode1 != http.StatusAccepted {
			logCtx.Warn("Failed to Delete Query in CreateSixSignalShareableURLHandler: ", errMsg1)
		}
		return queries.IdText, http.StatusCreated, "Shareable Query already shared"
	}

	shareableUrlRequest := &model.ShareableURL{
		EntityType: model.ShareableURLEntityTypeSixSignal,
		EntityID:   queries.ID,
		ShareType:  model.ShareableURLShareTypePublic,
		ProjectID:  projectId,
		CreatedBy:  agentUUID,
		ExpiresAt:  time.Now().AddDate(0, 3, 0).Unix(),
	}

	valid, errMsg := store.GetStore().ValidateCreateShareableURLRequest(shareableUrlRequest, projectId, agentUUID)
	if !valid {
		logCtx.Error(errMsg)
		errCode2, errMsg2 := store.GetStore().DeleteQuery(projectId, queries.ID)
		if errCode2 != http.StatusAccepted {
			logCtx.Warn("Failed to Delete Query in CreateSixSignalShareableURLHandler: ", errMsg2)
			return "", http.StatusBadRequest, errMsg2
		}
		return "", http.StatusBadRequest, errMsg
	}

	logCtx.Info("Shareable urls after validation: ", shareableUrlRequest)
	shareableUrlRequest.QueryID = queries.IdText
	share, err := store.GetStore().CreateShareableURL(shareableUrlRequest)
	if err != http.StatusCreated {
		logCtx.Error("Failed to create shareable query")
		errCode3, errMsg3 := store.GetStore().DeleteQuery(projectId, queries.ID)
		if errCode3 != http.StatusAccepted {
			logCtx.Warn("Failed to Delete Query in CreateSixSignalShareableURLHandler: ", errMsg3)
		}
		return "", http.StatusInternalServerError, "Shareable query creation failed."
	}

	return share.QueryID, http.StatusCreated, "Shareable Query creation successful"
}

//isReportShared checks if the report has been already made public
func isReportShared(projectID int64, idText string) (bool, string) {

	share, err := store.GetStore().GetShareableURLWithShareStringWithLargestScope(projectID, idText, model.ShareableURLEntityTypeSixSignal)
	if err == http.StatusBadRequest || err == http.StatusInternalServerError {
		return false, "Shareable query fetch failed. DB error."
	} else if err == http.StatusFound {
		if share.ShareType == model.ShareableURLShareTypePublic {
			return true, "Shareable url already exists."
		}
	}
	return false, "Shareable url doesn't exist"

}
