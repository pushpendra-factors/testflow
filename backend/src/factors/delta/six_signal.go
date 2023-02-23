package delta

import (
	"bytes"
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	T "factors/task"
	U "factors/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
)

func SixSignalAnalysis(projectIdArray []int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)

	for _, projectId := range projectIdArray {

		logCtx := log.WithFields(log.Fields{
			"projectId": projectId,
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

		logCtx.WithFields(log.Fields{"result": resultGroup}).Info("Printing the resultGroup")

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
		}

		pwmBytes, _ := json.Marshal(resultGroup)
		pString := string(pwmBytes)
		pString = pString + "\n"
		pBytes := []byte(pString)
		_, err = resultFile.Write(pBytes)
		logCtx.Info("ResultFile in SIxSignalAnalysis: ", resultFile)

		WriteSixSignalResultsToCloud(diskManager, folderName, projectId, logCtx)
	}
	return nil, true

}

func WriteSixSignalResultsToCloud(diskManager *serviceDisk.DiskDriver, queryId string, projectId int64, logCtx *log.Entry) error {

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

	cloudManager := C.GetCloudManager(projectId, true)
	path, _ = cloudManager.GetSixSignalAnalysisTempFilePathAndName(queryId, projectId)
	resultJson, err := json.Marshal(result)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("failed to unmarshal result Info.")
		return err
	}

	err = cloudManager.Create(path, "result.txt", bytes.NewReader(resultJson))
	if err != nil {
		log.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return err
	}
	return err
}

func GetSixSignalAnalysisData(projectId int64, id string) map[int]model.SixSignalResultGroup {

	cloudManager := C.GetCloudManager(projectId, true)
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
	return result
}
