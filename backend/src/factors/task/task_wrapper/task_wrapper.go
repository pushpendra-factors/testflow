package task_wrapper

import (
	"factors/model/model"
	"factors/model/store"
	"fmt"
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
			status, isSuccess := f(configs)
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

func TaskFuncWithProjectId(JobName string, lookback int, projectIds []uint64, f func(uint64, map[string]interface{}) (map[string]interface{}, bool), configs map[string]interface{}) map[string]interface{} {
	taskDetails, _, _ := store.GetStore().GetTaskDetailsByName(JobName)
	finalStatus := make(map[string]interface{})
	if taskDetails.IsProjectEnabled == false {
		finalStatus["status"] = "Call ProjectId Enabled Func"
		return finalStatus
	}
	for _, projectId := range projectIds {
		deltas, _, _ := store.GetStore().GetAllToBeExecutedDeltas(taskDetails.TaskID, projectId, lookback, nil)
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
				store.GetStore().InsertTaskBeginRecord(taskDetails.TaskID, projectId, delta)
				configs["startTimestamp"] = getDeltaAsTime(delta)
				configs["endTimestamp"] = getEndTimestamp(delta, taskDetails.Frequency, taskDetails.FrequencyInterval)
				status, isSuccess := f(projectId, configs)
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
