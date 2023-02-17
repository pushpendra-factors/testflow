package task_wrapper

import (
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"time"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

// Get the task details by name
// get all the deltas to be executed
// sort it
// check if dependent events are done
// then execute

func TaskFunc(JobName string, lookback int, f func(map[string]interface{}) (map[string]interface{}, bool), configs map[string]interface{}) map[string]interface{} {
	taskDetails, _, _ := store.GetStore().GetTaskDetailsByName(JobName)
	finalStatus := make(map[string]interface{})
	if taskDetails.IsProjectEnabled == true {
		finalStatus["status"] = "Call ProjectId Enabled Func"
		return finalStatus
	}
	deltas, _, _ := store.GetStore().GetAllToBeExecutedDeltas(taskDetails.TaskID, 0, lookback, nil)
	log.WithFields(log.Fields{"Deltas": deltas}).Info("Deltas to be processed")
	sort.SliceStable(deltas, func(i, j int) bool {
		return deltas[i] < deltas[j]
	})
	for _, delta := range deltas {
		log.Info("Checking dependency")
		isDone := store.GetStore().IsDependentTaskDone(taskDetails.TaskID, 0, delta)
		deltaString := fmt.Sprintf("%v", delta)
		if isDone == true {
			log.WithFields(log.Fields{"delta": delta}).Info("processing")
			store.GetStore().InsertTaskBeginRecord(taskDetails.TaskID, 0, delta)
			configs["startTimestamp"] = getDeltaAsTime(delta)
			configs["endTimestamp"] = getEndTimestamp(delta, taskDetails.Frequency, taskDetails.FrequencyInterval)
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 1024)
					runtime.Stack(buf, false)
					msg := fmt.Sprintf("Panic CausedBy: %v\nStackTrace: %v\n", r, string(buf))
					log.Errorf("Recovering from panic: %v", msg)
					log.WithFields(log.Fields{"delta": delta, "is_project_enabled": taskDetails.IsProjectEnabled}).Info("processing failed")
					// Report Error here
					store.GetStore().DeleteTaskEndRecord(taskDetails.TaskID, 0, delta)
				}
			}()
			status, isSuccess := f(configs)
			if(status != nil){
				finalStatus = status
			}
			finalStatus[deltaString] = status
			finalStatus["Status"+deltaString] = isSuccess
			if isSuccess == false {
				log.WithFields(log.Fields{"delta": delta}).Info("processing failed")
				// Report Error here
				store.GetStore().DeleteTaskEndRecord(taskDetails.TaskID, 0, delta)
				break
			}
			log.WithFields(log.Fields{"delta": delta}).Info("processing success")
			store.GetStore().InsertTaskEndRecord(taskDetails.TaskID, 0, delta)
		} else {
			finalStatus[deltaString] = "dependency not done yet"
			finalStatus["Status"+deltaString] = false
			log.Info(fmt.Sprintf("%v - dependency not done yet", deltaString))
			// have alerts here if over the threshold, report error
		}
	}
	return finalStatus
}

func TaskFuncWithProjectId(JobName string, lookback int, projectIds []int64, f func(int64, map[string]interface{}) (map[string]interface{}, bool), configs map[string]interface{}) map[string]interface{} {
	taskDetails, _, _ := store.GetStore().GetTaskDetailsByName(JobName)
	finalStatus := make(map[string]interface{})
	if taskDetails.IsProjectEnabled == false {
		finalStatus["status"] = "Call ProjectId Enabled Func"
		return finalStatus
	}
	for _, projectId := range projectIds {
		deltasPreFilter, _, _ := store.GetStore().GetAllToBeExecutedDeltas(taskDetails.TaskID, projectId, lookback, nil)
		projectDetails, _ := store.GetStore().GetProject(projectId)
		deltas := make([]uint64, 0)
		if projectDetails.TimeZone != "" && taskDetails.Frequency != 1{
			currentMaxDelta := U.DateAsFormattedInt(U.TimeNowIn(U.TimeZoneString(projectDetails.TimeZone)))
			for _, delta := range deltasPreFilter {
				if delta <= currentMaxDelta {
					deltas = append(deltas, delta)
				}
			}
		} else {
			deltas = deltasPreFilter
		}
		log.WithFields(log.Fields{"Deltas": deltas, "projectId": projectId}).Info("Deltas to be processed")
		sort.SliceStable(deltas, func(i, j int) bool {
			return deltas[i] < deltas[j]
		})
		for _, delta := range deltas {
			log.Info("Checking dependency")
			isDone := store.GetStore().IsDependentTaskDone(taskDetails.TaskID, projectId, delta)
			deltaString := fmt.Sprintf("%v-%v", delta, projectId)
			if isDone == true {
				log.WithFields(log.Fields{"delta": delta, "projectId": projectId}).Info("processing")
				errCode, _ := store.GetStore().InsertTaskBeginRecord(taskDetails.TaskID, projectId, delta)
				if errCode != http.StatusCreated {
					continue
				}
				configs["startTimestamp"] = getDeltaAsTime(delta)
				configs["endTimestamp"] = getEndTimestamp(delta, taskDetails.Frequency, taskDetails.FrequencyInterval)
				defer func() {
					if r := recover(); r != nil {
						buf := make([]byte, 1024)
						runtime.Stack(buf, false)
						msg := fmt.Sprintf("Panic CausedBy: %v\nStackTrace: %v\n", r, string(buf))
						log.Errorf("Recovering from panic: %v", msg)
						log.WithFields(log.Fields{"delta": delta, "projectId": projectId}).Info("processing failed")
						// Report Error here
						store.GetStore().DeleteTaskEndRecord(taskDetails.TaskID, projectId, delta)
					}
				}()
				status, isSuccess := f(projectId, configs)
				if(status != nil){
					finalStatus = status
				}
				finalStatus[deltaString] = status
				finalStatus["Status"+deltaString] = isSuccess
				if isSuccess == false {
					log.WithFields(log.Fields{"delta": delta, "projectId": projectId}).Info("processing failed")
					// Report Error here
					store.GetStore().DeleteTaskEndRecord(taskDetails.TaskID, projectId, delta)
					break
				}
				log.WithFields(log.Fields{"delta": delta, "projectId": projectId}).Info("processing success")
				store.GetStore().InsertTaskEndRecord(taskDetails.TaskID, projectId, delta)
			} else {
				finalStatus[deltaString] = "dependency not done yet"
				finalStatus["Status"+deltaString] = false
				log.Info(fmt.Sprintf("%v - dependency not done yet", deltaString))
				// have alerts here if over the threshold, report error
			}
		}
	}
	return finalStatus
}

func GetTaskDeltaAsTime(delta uint64) int64 {
	return getDeltaAsTime(delta)
}

func GetTaskEndTimestamp(delta uint64, frequency int, frequencyInterval int) int64 {
	return getEndTimestamp(delta, frequency, frequencyInterval)
}

func getDeltaAsTime(delta uint64) int64 {
	hours := delta % 100
	datePart := fmt.Sprintf("%v", delta/100)
	deltaDate, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, datePart)
	deltaDate = deltaDate.Add(time.Hour * time.Duration(hours))
	return deltaDate.Unix()
}

func getEndTimestamp(delta uint64, frequency int, frequencyInterval int) int64 {
	hours := delta % 100
	datePart := fmt.Sprintf("%v", delta/100)
	deltaDate, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, datePart)
	deltaDate = deltaDate.Add(time.Hour * time.Duration(hours))
	if frequency == model.Hourly {
		return deltaDate.Add(time.Hour*time.Duration(frequencyInterval)).Unix() - 1
	}
	if frequency == model.Daily {
		return deltaDate.AddDate(0, 0, frequencyInterval).Unix() - 1
	}
	if frequency == model.Weekly {
		return deltaDate.AddDate(0, 0, 7*frequencyInterval).Unix() - 1
	}
	if frequency == model.Monthly {
		return deltaDate.AddDate(0, frequencyInterval, 0).Unix() - 1
	}
	if frequency == model.Quarterly {
		return deltaDate.AddDate(0, 3*frequencyInterval, 0).Unix() - 1
	}
	return deltaDate.Unix()
}
