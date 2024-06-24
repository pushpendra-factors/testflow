package task

import (
	"bufio"
	"bytes"
	"encoding/json"
	"factors/filestore"
	"factors/merge"
	M "factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	"factors/pull"
	U "factors/util"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

const positive_accounts_threshold = 150
const number_of_events_to_keep = 300

const firstEventDayCol = "firstEventDay"
const lastEventDayCol = "lastEventDay"
const othersCol = "ev#$others"

var dataLog = log.WithField("task", "predictive_scoring_data")

func CreatePredictiveAnalysisData(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	dataLog := dataLog.WithField("projectId", projectId)
	status := make(map[string]interface{})

	sortedCloudManager := configs["sortedCloudManager"].(*filestore.FileManager)
	archiveCloudManager := configs["archiveCloudManager"].(*filestore.FileManager)

	minDaysOfInput := (configs["minDaysOfInput"].(int))
	daysOfOutput := (configs["daysOfOutput"].(int))
	windowStartShift := configs["windowStartShift"].(int)
	windowEndShift := configs["windowEndShift"].(int)

	startTimestamp := configs["startTimestamp"].(int64)
	endTimestamp := configs["endTimestamp"].(int64)

	targetEvent := configs["targetEvent"].(string)
	targetEv := TargetEvent{EventName: targetEvent}

	targets := []TargetEvent{targetEv}

	for _, tg := range targets {
		writeFileDir, fileName := (*sortedCloudManager).GetPredictiveScoringTrainingDataFilePathAndName(projectId, startTimestamp, endTimestamp, tg.EventName, minDaysOfInput, daysOfOutput, windowStartShift, windowEndShift)
		cloudWriter, err := (*sortedCloudManager).GetWriter(writeFileDir, fileName)
		if err != nil {
			dataLog.WithFields(log.Fields{"fileDir": writeFileDir, "fileName": fileName}).Error("unable to get writer for file")
			status["err"] = err.Error()
			return status, false
		}

		dataLog.Info("createDailyTargetFilesIfNotExist")
		err = createDailyTargetFilesIfNotExist(projectId, startTimestamp, endTimestamp, archiveCloudManager, tg)
		if err != nil {
			dataLog.WithError(err).Error("error createDailyTargetFilesIfNotExist")
			status["err"] = err.Error()
			return status, false
		}

		var truncateEvents map[string]bool
		var keepEvents map[string]bool
		dataLog.Info("getTotalEventsCountMap")
		if totalEventsCountMap, err := getTotalEventsCountMap(projectId, startTimestamp, endTimestamp, archiveCloudManager); err != nil {
			dataLog.WithError(err).Error("error getTotalEventsCountMap")
			status["err"] = err.Error()
			return status, false
		} else {
			writeFileDir, _ := (*sortedCloudManager).GetPredictiveScoringEventsCountsFilePathAndName(projectId, startTimestamp, endTimestamp, tg.EventName, minDaysOfInput, daysOfOutput, windowStartShift, windowEndShift)
			fileName := "totalEventsCountMap.txt"
			err = createFileFromMap(writeFileDir, fileName, sortedCloudManager, totalEventsCountMap)
			if err != nil {
				dataLog.WithFields(log.Fields{"map": totalEventsCountMap, "err": err}).Error("error creating total event counts file")
				status["err"] = err.Error()
				return status, false
			}
			var finalEventsCounts map[string]int
			dataLog.Info("getEventsToTruncateAndEventsToKeep")
			truncateEvents, keepEvents, finalEventsCounts = getEventsToTruncateAndEventsToKeep(totalEventsCountMap)
			// dataLog.WithFields(log.Fields{"truncateEvents": truncateEvents, "keepEvents": keepEvents, "finalEventsCounts": finalEventsCounts, "err": err}).Info("events info")
			writeFileDir, fileName = (*sortedCloudManager).GetPredictiveScoringEventsCountsFilePathAndName(projectId, startTimestamp, endTimestamp, tg.EventName, minDaysOfInput, daysOfOutput, windowStartShift, windowEndShift)
			err = createFileFromMap(writeFileDir, fileName, sortedCloudManager, finalEventsCounts)
			if err != nil {
				dataLog.WithFields(log.Fields{"map": finalEventsCounts, "err": err}).Error("error creating total event counts file")
				status["err"] = err.Error()
				return status, false
			}
		}

		countDatapoints := 0

		inputStartDayTimestamp := U.GetBeginningOfDayTimestamp(startTimestamp)
		inputEndDayTimestamp := inputStartDayTimestamp + (int64(minDaysOfInput) * U.Per_day_epoch) - 1
		outputStartDayTimestamp := inputEndDayTimestamp + 1
		outputEndDayTimestamp := outputStartDayTimestamp + (int64(daysOfOutput) * U.Per_day_epoch) - 1

		dataLog.Info("getActiveAccounts")
		activeAccounts, err := getActiveAccounts(projectId, startTimestamp, endTimestamp, positive_accounts_threshold, archiveCloudManager, tg.EventName)
		if err != nil {
			dataLog.WithError(err).Error("unable to get active accounts")
			status["err"] = err.Error()
			return status, false
		}

		var inputTargetMap = make(map[string]bool)
		var accDatapointMap = make(map[string]map[string]interface{})
		var dayCount int64 = 0
		log.Infof("starting getting datapoints, %d-%d, minDaysOfInput:%d, daysOfOutput:%d, windowStartShift:%d, windowEndShift:%d", startTimestamp, endTimestamp, minDaysOfInput, daysOfOutput, windowStartShift, windowEndShift)
		for outputEndDayTimestamp <= endTimestamp {
			// numFiles := (inputEndDayTimestamp + 1 - inputStartDayTimestamp) / U.Per_day_epoch
			// if nextInputMode == "distinct" {
			// 	inputTargetMap = make(map[string]bool)
			// 	userPropCounts = make(map[string]map[string]float64)
			// 	accDatapointMap = make(map[string]map[string]interface{})
			// 	fileCount = 0
			// }
			accountsUsed := make(map[string]bool)
			timestamp := inputStartDayTimestamp
			for timestamp <= inputEndDayTimestamp {
				dayCount += 1

				err := addDailyTargetMap(projectId, inputTargetMap, timestamp, archiveCloudManager, tg.EventName)
				if err != nil {
					dataLog.WithFields(log.Fields{"err": err}).Error("Unable to add daily target map")
					status["err"] = err.Error()
					return status, false
				}

				err = addDailyInputMap(projectId, accDatapointMap, accountsUsed, timestamp, activeAccounts, inputTargetMap, dayCount, archiveCloudManager, truncateEvents, keepEvents)
				if err != nil {
					dataLog.WithFields(log.Fields{"err": err}).Error("Unable to add daily target map")
					status["err"] = err.Error()
					return status, false
				}
				timestamp += U.Per_day_epoch
			}

			dataLog.Info("getOutputTargetMap")
			outputTargetMap, err := getAggregateTargetMap(projectId, outputStartDayTimestamp, outputEndDayTimestamp, archiveCloudManager, tg.EventName)
			if err != nil {
				dataLog.WithError(err).Error("error getAggregateTargetMap")
				status["err"] = err.Error()
				return status, false
			}

			if cnt, err := writeDatapointsToFile(accDatapointMap, accountsUsed, int(dayCount), cloudWriter, outputTargetMap, true, false); err != nil {
				dataLog.WithFields(log.Fields{"err": err}).Error("Unable to write datapoints to file.")
				status["err"] = err.Error()
				return status, false
			} else {
				dataLog.Info("countIds: ", cnt, ", dayCount: ", dayCount)
				countDatapoints += cnt
			}

			outputStartDayTimestamp += int64(windowStartShift) * U.Per_day_epoch
			outputEndDayTimestamp += int64(windowEndShift) * U.Per_day_epoch

			inputStartDayTimestamp += int64(windowStartShift) * U.Per_day_epoch
			inputEndDayTimestamp += int64(windowEndShift) * U.Per_day_epoch
		}
		dataLog.Info("countDatapoints: ", countDatapoints)

		err = cloudWriter.Close()
		if err != nil {
			dataLog.WithFields(log.Fields{"fileDir": writeFileDir, "fileName": fileName}).Error("unable to close writer for file")
			status["err"] = err.Error()
			return status, false
		}

		writeFileDir, fileName = (*sortedCloudManager).GetPredictiveScoringPredictDataFilePathAndName(projectId, startTimestamp, endTimestamp, tg.EventName, minDaysOfInput, daysOfOutput, windowStartShift, windowEndShift)
		cloudWriter, err = (*sortedCloudManager).GetWriter(writeFileDir, fileName)
		if err != nil {
			dataLog.WithFields(log.Fields{"fileDir": writeFileDir, "fileName": fileName}).Error("unable to get writer for file")
			status["err"] = err.Error()
			return status, false
		}

		accountsUsed := make(map[string]bool)
		for inputEndDayTimestamp <= endTimestamp {
			timestamp := inputStartDayTimestamp
			for timestamp <= inputEndDayTimestamp {
				dayCount += 1

				err := addDailyTargetMap(projectId, inputTargetMap, timestamp, archiveCloudManager, tg.EventName)
				if err != nil {
					dataLog.WithFields(log.Fields{"err": err}).Error("Unable to add daily target map")
					status["err"] = err.Error()
					return status, false
				}

				err = addDailyInputMap(projectId, accDatapointMap, accountsUsed, timestamp, activeAccounts, inputTargetMap, dayCount, archiveCloudManager, truncateEvents, keepEvents)
				if err != nil {
					dataLog.WithFields(log.Fields{"err": err}).Error("Unable to add daily input map")
					status["err"] = err.Error()
					return status, false
				}
				timestamp += U.Per_day_epoch
			}
			inputStartDayTimestamp += int64(windowStartShift) * U.Per_day_epoch
			inputEndDayTimestamp += int64(windowEndShift) * U.Per_day_epoch
		}

		if cnt, err := writeDatapointsToFile(accDatapointMap, nil, int(dayCount), cloudWriter, nil, false, true); err != nil {
			dataLog.WithFields(log.Fields{"err": err}).Error("Unable to write datapoints to file.")
			status["err"] = err.Error()
			return status, false
		} else {
			dataLog.Info("countIds: ", cnt, ", dayCount: ", dayCount)
		}

		err = cloudWriter.Close()
		if err != nil {
			dataLog.WithFields(log.Fields{"fileDir": writeFileDir, "fileName": fileName}).Error("unable to close writer for file")
			status["err"] = err.Error()
			return status, false
		}
	}

	dataLog.Info("CreatePredictiveAnalysisData successful")
	return status, true
}

func getAggregateTargetMap(projectId, startTimestamp, endTimestamp int64, cloudManager *filestore.FileManager, targetEvent string) (map[string]bool, error) {
	var targetMap = make(map[string]bool)
	timestamp := startTimestamp
	for timestamp <= endTimestamp {
		addDailyTargetMap(projectId, targetMap, timestamp, cloudManager, targetEvent)
		timestamp += U.Per_day_epoch
	}
	return targetMap, nil
}

func getActiveAccounts(projectId, startTimestamp, endTimestamp int64, posAccountsThreshold int, cloudManager *filestore.FileManager, targetEvent string) (map[string]bool, error) {
	var targetMap = make(map[string]bool)
	timestamp := endTimestamp + 1 - U.Per_day_epoch
	num_days := 0
	num_pos_acc := 0
	for timestamp >= startTimestamp {
		num_days += 1
		addDailyTargetMap(projectId, targetMap, timestamp, cloudManager, targetEvent)
		num_pos_acc = getPositiveUsersCountFromTargetMap(targetMap)
		if num_pos_acc >= posAccountsThreshold {
			break
		}
		timestamp -= U.Per_day_epoch
	}
	log.Infof("Taking %d days of active users: %d (%d positive)", num_days, len(targetMap), num_pos_acc)
	return targetMap, nil
}

func getPositiveUsersCountFromTargetMap(targetMap map[string]bool) int {
	cnt := 0
	for _, val := range targetMap {
		if val {
			cnt += 1
		}
	}
	return cnt
}

func addDailyTargetMap(projectId int64, targetMap map[string]bool, timestamp int64, cloudManager *filestore.FileManager, targetEvent string) error {
	var target map[string]bool
	fileDir, fileName := (*cloudManager).GetEventsAggregateDailyTargetFilePathAndName(projectId, timestamp, targetEvent)
	target, err := readTargetMapFromFile(fileDir, fileName, cloudManager)
	if err != nil {
		dataLog.WithError(err).Error("error reading target map")
		return err
	}
	for x, y := range target {
		if val, ok := targetMap[x]; !ok || !val {
			targetMap[x] = y
		}
	}
	return nil
}

func addDailyInputMap(projectId int64, inputMap map[string]map[string]interface{}, inputAccountsUsed map[string]bool, timestamp int64, activeAccounts map[string]bool, inputTargetMap map[string]bool, fileCount int64, cloudManager *filestore.FileManager, truncateEventsPrefixes map[string]bool, keepEventsMap map[string]bool) error {

	fileDir, fileName := (*cloudManager).GetEventsAggregateDailyDataFilePathAndName(projectId, timestamp)
	log.Infof("Reading predictive scoring daily input data file :%s, %s", fileDir, fileName)
	file, err := (*cloudManager).Get(fileDir, fileName)
	if err != nil {
		dataLog.WithError(err).Error("error reading file")
		return err
	}
	buf := make([]byte, maxCapacity)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()

		var dataPoint = make(map[string]interface{})
		err := json.Unmarshal([]byte(line), &dataPoint)
		if err != nil {
			dataLog.WithError(err).Error("error unmarshalling datapoint")
			return err
		}
		AccID := dataPoint[IdColumnName].(string)

		if _, ok := activeAccounts[AccID]; ok {
			inputAccountsUsed[AccID] = true
		}
		if _, ok := inputMap[AccID]; !ok {
			var newMap = make(map[string]interface{})
			newMap[firstEventTimeCol] = float64(((fileCount - 1) * U.Per_day_epoch)) + dataPoint[firstEventTimeCol].(float64)
			inputMap[AccID] = newMap
		}

		if yes, ok := inputTargetMap[AccID]; ok && yes {
			// delete(activeAccounts, AccID)
			// delete(inputMap, AccID)
			delete(inputAccountsUsed, AccID)
		}
		finalDataPoint := inputMap[AccID]
		finalDataPoint[lastEventTimeCol] = float64(((fileCount - 1) * U.Per_day_epoch)) + dataPoint[lastEventTimeCol].(float64)

		for prop, val := range dataPoint {

			if strings.HasPrefix(prop, "ev#") {
				//handle events
				propArr := strings.Split(prop, "/")
				rest := strings.Join(propArr[:len(propArr)-1], "/")
				var key string
				if yes, ok := truncateEventsPrefixes[rest]; ok && yes {
					key = rest
				} else {
					key = prop
				}
				if _, ok := keepEventsMap[key]; !ok {
					if _, ok := finalDataPoint[othersCol]; !ok {
						finalDataPoint[othersCol] = 0.0
					}
					finalDataPoint[othersCol] = finalDataPoint[othersCol].(float64) + val.(float64)
				} else {
					if _, ok := finalDataPoint[key]; !ok {
						finalDataPoint[key] = 0.0
					}
					finalDataPoint[key] = finalDataPoint[key].(float64) + val.(float64)
				}
			} else if strings.HasPrefix(prop, "up#") {
				if val == "" || val == nil {
					continue
				}
				if _, ok := propsToUse[prop]; !ok {
					continue
				}
				finalDataPoint[prop] = val
				// } else if strings.HasPrefix(prop, "ep#") {
				// 	if val == "" || val == nil {
				// 		continue
				// 	}
				// 	op, ok := propsToUse[prop]
				// 	if !ok {
				// 		continue
				// 	}
				// 	if op == "pass" {
				// 		continue
				// 	}
				// 	if newId {
				// 		finalDataPoint[prop] = val
				// 		continue
				// 	}
				// 	if op == "first" {
				// 		if _, ok := finalDataPoint[prop]; !ok {
				// 			finalDataPoint[prop] = val
				// 		}
				// 	} else if op == "total" || op == "average" {
				// 		floatVal, err := U.GetPropertyValueAsFloat64(val)
				// 		if err != nil {
				// 			dataLog.WithError(err).Error("failed getting interface float value")
				// 			return err
				// 		}
				// 		if _, ok := finalDataPoint[prop]; !ok {
				// 			finalDataPoint[prop] = 0.0
				// 		}
				// 		var mult float64 = 1
				// 		var mult2 float64 = 1
				// 		var den float64 = 1
				// 		if op == "average" {
				// 			mult = userPropCounts[AccID][prop]
				// 			mult2 = dailyPropCounts[AccID][prop]
				// 			den = mult + mult2
				// 		}
				// 		finalDataPoint[prop] = ((finalDataPoint[prop].(float64) * mult) + (floatVal * mult2)) / den
				// 		userPropCounts[AccID][prop] += dailyPropCounts[AccID][prop]
				// 	} else if op == "last" {
				// 		finalDataPoint[prop] = val
				// 	}
			}
		}
	}
	return nil
}

func createDailyTargetFilesIfNotExist(projectId, startTimestamp, endTimestamp int64, cloudManager *filestore.FileManager, targetEvent TargetEvent) error {

	domain_group, httpStatus := store.GetStore().GetGroup(projectId, M.GROUP_NAME_DOMAINS)
	if httpStatus != http.StatusFound {
		err := fmt.Errorf("failed to get existing groups (%s) for project (%d)", M.GROUP_NAME_DOMAINS, projectId)
		return err
	}
	idNum := domain_group.ID

	timestamp := startTimestamp
	fileCount := 0
	for timestamp <= endTimestamp {
		var targetMap = make(map[string]bool)
		fileCount += 1
		fileDir, fileName := (*cloudManager).GetEventsAggregateDailyTargetFilePathAndName(projectId, timestamp, targetEvent.EventName)
		_, err := (*cloudManager).Get(fileDir, fileName)
		if err == nil {
			dataLog.Info("fileCount: ", fileCount, ", timestamp: ", timestamp, "- already present")
			timestamp += U.Per_day_epoch
			continue
		}
		partFilesDir, _ := pull.GetDailyArchiveFilePathAndName(cloudManager, U.DataTypeEvent, U.EVENTS_FILENAME_PREFIX, projectId, startTimestamp, 0, 0)
		listFiles := (*cloudManager).ListFiles(partFilesDir)
		for _, partFileFullName := range listFiles {
			partFNamelist := strings.Split(partFileFullName, "/")
			partFileName := partFNamelist[len(partFNamelist)-1]
			if !strings.HasPrefix(partFileName, U.EVENTS_FILENAME_PREFIX) {
				continue
			}
			file, err := (*cloudManager).Get(partFilesDir, partFileName)
			if err != nil {
				log.WithError(err).Error("error reading file")
				return err
			}

			scanner := bufio.NewScanner(file)
			buf := make([]byte, maxCapacity)
			scanner.Buffer(buf, maxCapacity)

			for scanner.Scan() {
				line := scanner.Text()

				var eventDetails *P.CounterEventFormat
				err = json.Unmarshal([]byte(line), &eventDetails)
				if err != nil {
					log.WithError(err).Error("error reading file")
					return err
				}
				AccID := merge.GetAptId(eventDetails, idNum)

				if _, ok := targetMap[AccID]; !ok {
					targetMap[AccID] = false
				}
				if eventMatchesTarget(eventDetails, targetEvent) {
					targetMap[AccID] = true
				}
			}
		}
		err = createFileFromMap(fileDir, fileName, cloudManager, targetMap)
		if err != nil {
			dataLog.WithFields(log.Fields{"map": targetMap, "err": err}).Error("error creating event counts file")
			return err
		}
		dataLog.Info("fileCount: ", fileCount, ", timestamp: ", timestamp, "- created target file")
		timestamp += U.Per_day_epoch
	}
	// timestamp := startTimestamp
	// fileCount := 0
	// for timestamp <= endTimestamp {
	// 	fileCount += 1
	// 	fileDir, fileName := (*cloudManager).GetEventsAggregateDailyTargetFilePathAndName(projectId, timestamp, targetEvent)
	// 	_, err := (*cloudManager).Get(fileDir, fileName)
	// 	if err == nil {
	// 		dataLog.Info("fileCount: ", fileCount, ", timestamp: ", timestamp, "- already present")
	// 		timestamp += U.Per_day_epoch
	// 		continue
	// 	}

	// 	targetMap := make(map[string]bool)
	// 	fileDir, fileName = (*cloudManager).GetEventsAggregateDailyDataFilePathAndName(projectId, timestamp)
	// 	file, err := (*cloudManager).Get(fileDir, fileName)
	// 	if err != nil {
	// 		dataLog.WithError(err).Error("error reading file")
	// 		return err
	// 	}
	// 	buf := make([]byte, maxCapacity)
	// 	scanner := bufio.NewScanner(file)
	// 	scanner.Buffer(buf, maxCapacity)

	// 	for scanner.Scan() {
	// 		line := scanner.Text()

	// 		var dataPoint = make(map[string]interface{})
	// 		err := json.Unmarshal([]byte(line), &dataPoint)
	// 		if err != nil {
	// 			dataLog.WithError(err).Error("error unmarshalling datapoint")
	// 			return err
	// 		}
	// 		AccID := dataPoint[IdColumnName].(string)

	// 		if freq, ok := dataPoint[getEventColumnName(targetEvent)]; ok && freq != 0 {
	// 			targetMap[AccID] = true
	// 		} else {
	// 			targetMap[AccID] = false
	// 		}
	// 	}
	// 	err = createFileFromMap(fileDir, fileName, cloudManager, targetMap)
	// 	if err != nil {
	// 		dataLog.WithFields(log.Fields{"map": targetMap, "err": err}).Error("error creating event counts file")
	// 		return err
	// 	}
	// 	dataLog.Info("fileCount: ", fileCount, ", timestamp: ", timestamp)
	// 	timestamp += U.Per_day_epoch
	// }
	return nil
}

func readTargetMapFromFile(fileDir, fileName string, cloudManager *filestore.FileManager) (map[string]bool, error) {
	var reqMap map[string]bool
	file, err := (*cloudManager).Get(fileDir, fileName)
	if err != nil {
		dataLog.WithError(err).Error("error reading file")
		return nil, err
	}
	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(file)
	if err != nil {
		dataLog.Error("Error reading properties map from File")
		return nil, err
	}
	json.Unmarshal(buffer.Bytes(), &reqMap)
	return reqMap, nil
}

func getTotalEventsCountMap(projectId, startTimestamp, endTimestamp int64, cloudManager *filestore.FileManager) (map[string]map[string]int, error) {
	var totalEventsCount = make(map[string]map[string]int)
	timestamp := startTimestamp
	for timestamp <= endTimestamp {
		filesDir, fileName := (*cloudManager).GetEventsAggregateDailyCountsFilePathAndName(projectId, timestamp)
		file, err := (*cloudManager).Get(filesDir, fileName)
		if err != nil {
			dataLog.WithError(err).Error("error reading file")
			return nil, err
		}
		var dailyEventsCount map[string]map[string]int
		var buf bytes.Buffer
		_, err = buf.ReadFrom(file)
		if err != nil {
			dataLog.Error("Error reading properties map from File")
			return nil, err
		}
		json.Unmarshal(buf.Bytes(), &dailyEventsCount)

		for x, y := range dailyEventsCount {
			if _, ok := totalEventsCount[x]; !ok {
				totalEventsCount[x] = make(map[string]int)
			}
			for z, val := range y {
				totalEventsCount[x][z] += val
			}
		}
		timestamp += U.Per_day_epoch
	}
	return totalEventsCount, nil
}

func getEventsToTruncateAndEventsToKeep(totalEventsCountMap map[string]map[string]int) (map[string]bool, map[string]bool, map[string]int) {
	var truncateEvents = make(map[string]bool)
	evCounts := make(map[string]int)
	evsToKeep := make(map[string]bool)
	finalEvsCounts := make(map[string]int)
	finalEvsCounts[othersCol] = 0
	numEvsToKeep := number_of_events_to_keep - 1
	for x, y := range totalEventsCountMap {
		if x == "" {
			continue
		}
		maxVal := 0
		numVals := 0
		total := 0
		for _, val := range y {
			numVals += 1
			total += val
			if val > maxVal {
				maxVal = val
			}
		}
		evCounts[x] = total
		if total/numVals == 1 && numVals > 10 {
			truncateEvents[x] = true
		}
	}
	for x, y := range totalEventsCountMap {
		if x == "" {
			continue
		}
		if _, ok := truncateEvents[x]; !ok {
			delete(evCounts, x)
			for z, val := range y {
				evCounts[strings.Join([]string{x, z}, "/")] = val
			}
		}
	}
	pairList := U.RankByWordCount(evCounts)
	for i, pr := range pairList {
		if i == numEvsToKeep {
			finalEvsCounts[othersCol] = pr.Value
		} else if i > numEvsToKeep {
			finalEvsCounts[othersCol] += pr.Value
		} else {
			evsToKeep[pr.Key] = true
			finalEvsCounts[pr.Key] = pr.Value
		}

	}
	return truncateEvents, evsToKeep, finalEvsCounts
}

func writeDatapointsToFile(accDatapointMap map[string]map[string]interface{}, accsToUse map[string]bool, dayCount int, cloudWriter io.WriteCloser, outputTargetMap map[string]bool, toUseTarget bool, toUseAccId bool) (int, error) {
	countPoints := 0
	var allAccounts bool = false
	if accsToUse == nil {
		allAccounts = true
	}
	for accId, dataPoint := range accDatapointMap {
		if !allAccounts {
			if _, ok := accsToUse[accId]; !ok {
				continue
			}
		}
		if toUseTarget {
			dataPoint["target"] = outputTargetMap[accId]
		}
		if toUseAccId {
			dataPoint[IdColumnName] = accId
		}
		dataPoint["num_days"] = dayCount
		dataPoint[firstEventDayCol] = 1 + (int64(dataPoint[firstEventTimeCol].(float64)) / U.Per_day_epoch)
		dataPoint[lastEventDayCol] = 1 + (int64(dataPoint[lastEventTimeCol].(float64)) / U.Per_day_epoch)
		if dataPointBytes, err := json.Marshal(dataPoint); err != nil {
			dataLog.WithFields(log.Fields{"dataPoint": dataPoint, "err": err}).Error("Marshal failed")
			return countPoints, err
		} else {
			lineWrite := string(dataPointBytes)
			if _, err := io.WriteString(cloudWriter, lineWrite+"\n"); err != nil {
				dataLog.WithFields(log.Fields{"line": lineWrite, "err": err}).Error("Unable to write to file.")
				return countPoints, err
			}
		}
		countPoints += 1
	}
	return countPoints, nil
}
