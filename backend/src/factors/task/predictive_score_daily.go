package task

import (
	"bufio"
	"bytes"
	"encoding/json"
	"factors/delta"
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

const firstEventTimeCol = "firstEventTime"
const lastEventTimeCol = "lastEventTime"

const maxCapacity = 30 * 1024 * 1024

var aggLog = log.WithField("task", "events_aggregate_daily")

type TargetEvent struct {
	EventName  string
	Properties []M.QueryProperty
}

func EventsAggregateDaily(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	aggLog = aggLog.WithField("project", projectId)
	status := make(map[string]interface{})

	// tmpCloudManager := configs["tmpCloudManager"].(*filestore.FileManager)
	archiveCloudManager := configs["archiveCloudManager"].(*filestore.FileManager)

	startTimestamp := configs["startTimestamp"].(int64)
	// endTimestamp := *(configs["endTimestamp"].(*int64))

	targetEvent := configs["targetEvent"].(string)

	targetEv := TargetEvent{EventName: targetEvent}

	propTimestamps := make(map[string]int64)
	targets := []TargetEvent{targetEv}
	// get id number for aaccount ids
	domain_group, httpStatus := store.GetStore().GetGroup(projectId, M.GROUP_NAME_DOMAINS)
	if httpStatus != http.StatusFound {
		err := fmt.Errorf("failed to get existing groups (%s) for project (%d)", M.GROUP_NAME_DOMAINS, projectId)
		aggLog.WithField("err_code", status).Error(err)
		status["err"] = err.Error()
		return status, false
	}
	idNum := domain_group.ID

	projectDetails, _ := store.GetStore().GetProject(projectId)
	startTimestampInProjectTimezone := startTimestamp
	if projectDetails.TimeZone != "" {
		// Input time is in UTC. We need the same time in the other timezone
		// if 2021-08-30 00:00:00 is UTC then we need the epoch equivalent in 2021-08-30 00:00:00 IST(project time zone)
		offset := U.FindOffsetInUTCForTimestamp(U.TimeZoneString(projectDetails.TimeZone), startTimestamp)
		startTimestampInProjectTimezone = startTimestamp - int64(offset)
	}

	var eventCounts = make(map[string]map[string]int)
	var accPropValCounts = make(map[string]map[string]map[string]int)
	var target = make(map[int]map[string]bool)
	for i, _ := range targets {
		target[i] = make(map[string]bool)
	}

	var accountInfos = make(map[string]map[string]interface{})

	fileDir, fileName := (*archiveCloudManager).GetEventsAggregateDailyDataFilePathAndName(projectId, startTimestamp)
	cloudWriter, err := (*archiveCloudManager).GetWriter(fileDir, fileName)
	if err != nil {
		aggLog.WithFields(log.Fields{"fileDir": fileDir, "fileName": fileName}).Error("unable to get writer for file")
		status["err"] = err.Error()
		return status, false
	}

	partFilesDir, _ := pull.GetDailyArchiveFilePathAndName(archiveCloudManager, U.DataTypeEvent, U.EVENTS_FILENAME_PREFIX, projectId, startTimestamp, 0, 0)
	listFiles := (*archiveCloudManager).ListFiles(partFilesDir)
	for _, partFileFullName := range listFiles {
		partFNamelist := strings.Split(partFileFullName, "/")
		partFileName := partFNamelist[len(partFNamelist)-1]
		if !strings.HasPrefix(partFileName, U.EVENTS_FILENAME_PREFIX) {
			continue
		}
		file, err := (*archiveCloudManager).Get(partFilesDir, partFileName)
		if err != nil {
			log.WithError(err).Error("error reading file")
			status["err"] = err.Error()
			return status, false
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
				status["err"] = err.Error()
				return status, false
			}
			AccID := merge.GetAptId(eventDetails, idNum)

			eventSeconds := eventDetails.EventTimestamp - startTimestampInProjectTimezone

			if _, ok := accountInfos[AccID]; !ok {
				var newMap = make(map[string]interface{})
				newMap[IdColumnName] = AccID
				newMap[firstEventTimeCol] = eventSeconds
				newMap[lastEventTimeCol] = eventSeconds
				accountInfos[AccID] = newMap
				accPropValCounts[AccID] = make(map[string]map[string]int)
			}
			dataPoint := accountInfos[AccID]

			eName := getEventColumnName(eventDetails.EventName)

			if eventSeconds < dataPoint[firstEventTimeCol].(int64) {
				dataPoint[firstEventTimeCol] = eventSeconds
			} else if eventSeconds > dataPoint[lastEventTimeCol].(int64) {
				dataPoint[lastEventTimeCol] = eventSeconds
			}

			updateTargetsMap(target, eventDetails, targets, AccID)

			eNameArr := strings.Split(eName, "/")
			last := eNameArr[len(eNameArr)-1]
			rest := strings.Join(eNameArr[:len(eNameArr)-1], "/")
			if _, ok := eventCounts[rest]; !ok {
				eventCounts[rest] = make(map[string]int)
			}
			eventCounts[rest][last] += 1

			if _, ok := dataPoint[eName]; !ok {
				dataPoint[eName] = 0
			}
			dataPoint[eName] = dataPoint[eName].(int) + 1

			propValCounts := accPropValCounts[AccID]

			for uKey, uVal := range eventDetails.UserProperties {
				uKey = getUserPropertyColumnName(uKey)
				if uVal == "" || uVal == nil {
					continue
				}
				if strVal, err := U.GetStringValueFromInterface(uVal); err != nil {
					if _, ok := propValCounts[uKey]; !ok {
						propValCounts[uKey] = make(map[string]int)
					}
					propValCounts[uKey][strVal] += 1
				}
				if time, ok := propTimestamps[uKey]; !ok || eventSeconds > time {
					propTimestamps[uKey] = eventSeconds
					dataPoint[uKey] = uVal
				}
			}

			// for eKey, eVal := range eventDetails.EventProperties {
			// 	eKey = getEventPropertyColumnName(eKey)
			// 	if eVal == "" || eVal == nil {
			// 		continue
			// 	}
			// 	if strVal, err := U.GetStringValueFromInterface(eVal); err != nil {
			// 		if _, ok := propValCounts[eKey]; !ok {
			// 			propValCounts[eKey] = make(map[string]int)
			// 		}
			// 		propValCounts[eKey][strVal] += 1
			// 	}
			// 	var op string
			// 	if givenOp, ok := propsToUse[eKey]; !ok {
			// 		op = "last"
			// 	} else {
			// 		op = givenOp
			// 	}
			// 	if op == "pass" {
			// 		continue
			// 	} else if op == "total" || op == "average" {
			// 		floatVal, err := U.GetPropertyValueAsFloat64(eVal)
			// 		if err != nil {
			// 			aggLog.WithError(err).Error("failed getting interface float value - prop: ", eKey)
			// 			continue
			// 		}
			// 		if _, ok := dataPoint[eKey]; !ok {
			// 			dataPoint[eKey] = 0.0
			// 		}
			// 		var mult float64 = 1
			// 		var den float64 = 1
			// 		if op == "average" {
			// 			mult = propCounts[eKey]
			// 			den = mult + 1
			// 		}
			// 		dataPoint[eKey] = ((dataPoint[eKey].(float64) * mult) + floatVal) / den
			// 		propCounts[eKey] += 1
			// 	} else if op == "first" || op == "last" {
			// 		if !(evType == op || newId) {
			// 			continue
			// 		}
			// 		dataPoint[eKey] = eVal
			// 	}
			// }
		}
	}

	for i, tgMap := range target {
		targ := targets[i]
		fileDir, fileName = (*archiveCloudManager).GetEventsAggregateDailyTargetFilePathAndName(projectId, startTimestamp, targ.EventName)
		err = createFileFromMap(fileDir, fileName, archiveCloudManager, tgMap)
		if err != nil {
			aggLog.WithFields(log.Fields{"map": target, "err": err}).Error("error creating target file")
			status["err"] = err.Error()
			return status, false
		}
	}

	fileDir, fileName = (*archiveCloudManager).GetEventsAggregateDailyPropsFilePathAndName(projectId, startTimestamp)
	err = createFileFromMap(fileDir, fileName, archiveCloudManager, accPropValCounts)
	if err != nil {
		aggLog.WithFields(log.Fields{"map": accPropValCounts, "err": err}).Error("error creating prop counts file")
		status["err"] = err.Error()
		return status, false
	}

	fileDir, fileName = (*archiveCloudManager).GetEventsAggregateDailyCountsFilePathAndName(projectId, startTimestamp)
	err = createFileFromMap(fileDir, fileName, archiveCloudManager, eventCounts)
	if err != nil {
		aggLog.WithFields(log.Fields{"map": eventCounts, "err": err}).Error("error creating event counts file")
		status["err"] = err.Error()
		return status, false
	}

	countDatapoints := 0
	for _, dataPoint := range accountInfos {
		if dataPointBytes, err := json.Marshal(dataPoint); err != nil {
			aggLog.WithFields(log.Fields{"dataPoint": dataPoint, "err": err}).Error("Marshal failed")
			status["err"] = err.Error()
			return status, false
		} else {
			lineWrite := string(dataPointBytes)
			if _, err := io.WriteString(cloudWriter, lineWrite+"\n"); err != nil {
				aggLog.WithFields(log.Fields{"line": lineWrite, "err": err}).Error("Unable to write to file.")
				status["err"] = err.Error()
				return status, false
			}
			countDatapoints += 1
		}
	}
	aggLog.Info("countDatapoints: ", countDatapoints)

	aggLog.Info("PredictiveScoringDaily successfully completed")
	// err := createTrainingData(projectId, archiveCloudManager, dailyFilesDir, daysOfInput, daysOfOutput, gapDaysForNextInput, startTimestamp, endTimestamp)
	// if err != nil {
	// 	aggLog.WithError(err).Error("error creating training data")
	// 	status["err"] = err.Error()
	// 	return status, false
	// }
	return status, true
}

func getEventColumnName(eventName string) string {
	return "ev#" + eventName
}

// func getEventPropertyColumnName(propName string) string {
// 	return "ep#" + propName
// }

func getUserPropertyColumnName(propName string) string {
	return "up#" + propName
}

func createFileFromMap(fileDir, fileName string, cloudManager *filestore.FileManager, Map interface{}) error {
	mapBytes, err := json.Marshal(Map)
	if err == nil {
		err = (*cloudManager).Create(fileDir, fileName, bytes.NewReader(mapBytes))
	}
	return err
}

func eventMatchesTarget(event *P.CounterEventFormat, target_event TargetEvent) bool {
	if target_event.EventName == "" {
		target_event.EventName = event.EventName
	}
	yes, _ := delta.ToCountEvent(*event, target_event.EventName, target_event.Properties)
	return yes
}

func updateTargetsMap(targetsMap map[int]map[string]bool, event *P.CounterEventFormat, targetEvents []TargetEvent, AccountId string) {
	for i, target := range targetEvents {
		specificTarget := targetsMap[i]
		if _, ok := specificTarget[AccountId]; !ok {
			specificTarget[AccountId] = false
		}
		if eventMatchesTarget(event, target) {
			specificTarget[AccountId] = true
		}
	}
}
