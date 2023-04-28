package task

import (
	"bytes"
	"encoding/json"
	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	"fmt"
	"net/http"
	"os"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

func SixSignalAnalysis(projectIdArray []int64, configs map[string]interface{}) map[string][]int64 {
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

func WriteSixSignalResultsToCloud(cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, queryId string, projectId int64, logCtx *log.Entry) error {

	path, _ := diskManager.GetSixSignalAnalysisTempFilePathAndName(queryId, projectId)
	resultLocalPath := path + "results.txt"
	logCtx.Info("Path in WriteResultstoCloud: ", resultLocalPath)
	scanner, err := OpenEventFileAndGetScanner(resultLocalPath)
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
