// Get all the processed delta
// insert task starting record
// update task done record
// GetAllToBeProcesssedDeltaGivenLookback
// precondition check
// preexecution phase
// post execution phase

package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// Given a startdate till now what are all date/hours have the job been completed for
func (pg *Postgres) GetAllProcessedIntervalsFromStartDate(taskID uint64, projectId uint64, startDate *time.Time) ([]uint64, int, string) {

	// get all the processed deltas with the given range
	logCtx := log.WithFields(log.Fields{"taskID": taskID})
	deltas := make([]uint64, 0)
	if taskID == 0 {
		logCtx.Error("missing taskID")
		return deltas, http.StatusBadRequest, "missing taskID"
	}

	baseTaskDetails, _, _ := pg.GetTaskDetailsById(taskID)
	isProjectEnabled := false
	if baseTaskDetails.IsProjectEnabled == true {
		isProjectEnabled = true
	}
	// find all the deltas for a given task id from startTime to infinity
	// if starttime not specified, > currenttime - infinity
	db := C.GetServices().Db

	var startDelta uint64
	var endDelta uint64
	if startDate == nil {
		endDelta = 9999999999
		startDelta = U.DateAsFormattedInt(U.TimeNow())
	} else {
		endDelta = 9999999999
		startDelta = U.DateAsFormattedInt((*startDate))
	}
	var taskExecutionDetails []model.TaskExecutionDetails
	if isProjectEnabled == false {
		if err := db.Where(
			"task_id = ? AND delta >= ? AND delta <= ? AND is_completed = true",
			taskID, startDelta, endDelta).Find(&taskExecutionDetails).Error; err != nil {
			logCtx.Error(err.Error())
			return deltas, http.StatusInternalServerError, err.Error()
		}
	} else {
		if err := db.Where(
			"task_id = ? AND delta >= ? AND delta <= ? AND is_completed = true AND project_id = ?",
			taskID, startDelta, endDelta, projectId).Find(&taskExecutionDetails).Error; err != nil {
			logCtx.Error(err.Error())
			return deltas, http.StatusInternalServerError, err.Error()
		}
	}
	for _, data := range taskExecutionDetails {
		deltas = append(deltas, data.Delta)
	}
	return deltas, http.StatusOK, ""
}

// Given a enddate and lookback what are all date/hours have the job been completed for
func (pg *Postgres) GetAllProcessedIntervals(taskID uint64, projectId uint64, lookbackInDays int, endDate *time.Time) ([]uint64, int, string) {

	// get all the processed deltas with the given range
	logCtx := log.WithFields(log.Fields{"taskID": taskID, "lookback": lookbackInDays})
	deltas := make([]uint64, 0)
	if taskID == 0 {
		logCtx.Error("missing taskID")
		return deltas, http.StatusBadRequest, "missing taskID"
	}

	baseTaskDetails, _, _ := pg.GetTaskDetailsById(taskID)
	isProjectEnabled := false
	if baseTaskDetails.IsProjectEnabled == true {
		isProjectEnabled = true
	}
	// find all the deltas for a given task id from endtime-lookback to endtime
	// if endtime not specified, > currenttime -lookback
	db := C.GetServices().Db

	var startDelta uint64
	var endDelta uint64
	if endDate == nil {
		endDelta = ^uint64(0)
		startDelta = U.DateAsFormattedInt(U.TimeNow().AddDate(0, 0, -lookbackInDays))
	} else {
		endDelta = U.DateAsFormattedInt(*endDate)
		startDelta = U.DateAsFormattedInt((*endDate).AddDate(0, 0, -lookbackInDays))
	}
	var taskExecutionDetails []model.TaskExecutionDetails
	if isProjectEnabled == false {
		if err := db.Where(
			"task_id = ? AND delta >= ? AND delta <= ? AND is_completed = true",
			taskID, startDelta, endDelta).Find(&taskExecutionDetails).Error; err != nil {
			logCtx.Error(err.Error())
			return deltas, http.StatusInternalServerError, err.Error()
		}
	} else {
		if err := db.Where(
			"task_id = ? AND delta >= ? AND delta <= ? AND is_completed = true AND project_id = ?",
			taskID, startDelta, endDelta, projectId).Find(&taskExecutionDetails).Error; err != nil {
			logCtx.Error(err.Error())
			return deltas, http.StatusInternalServerError, err.Error()
		}
	}
	for _, data := range taskExecutionDetails {
		deltas = append(deltas, data.Delta)
	}
	return deltas, http.StatusOK, ""
}

// Get all the date/hours which is in progress state
func (pg *Postgres) GetAllInProgressIntervals(taskID uint64, projectId uint64, lookbackInDays int, endDate *time.Time) ([]uint64, int, string) {

	// get all the processed deltas with the given range
	logCtx := log.WithFields(log.Fields{"taskID": taskID, "lookback": lookbackInDays})
	deltas := make([]uint64, 0)
	if taskID == 0 {
		logCtx.Error("missing taskID")
		return deltas, http.StatusBadRequest, "missing taskID"
	}

	baseTaskDetails, _, _ := pg.GetTaskDetailsById(taskID)
	isProjectEnabled := false
	if baseTaskDetails.IsProjectEnabled == true {
		isProjectEnabled = true
	}

	// find all the deltas for a given task id from endtime-lookback to endtime
	// if endtime not specified, > currenttime -lookback
	db := C.GetServices().Db

	var startDelta uint64
	var endDelta uint64
	if endDate == nil {
		endDelta = ^uint64(0)
		startDelta = U.DateAsFormattedInt(U.TimeNow().AddDate(0, 0, -lookbackInDays))
	} else {
		endDelta = U.DateAsFormattedInt(*endDate)
		startDelta = U.DateAsFormattedInt((*endDate).AddDate(0, 0, -lookbackInDays))
	}
	var taskExecutionDetails []model.TaskExecutionDetails
	if isProjectEnabled == false {
		if err := db.Where(
			"task_id = ? AND delta >= ? AND delta <= ? AND is_completed = false",
			taskID, startDelta, endDelta).Find(&taskExecutionDetails).Error; err != nil {
			logCtx.Error(err.Error())
			return deltas, http.StatusInternalServerError, err.Error()
		}
	} else {
		if err := db.Where(
			"task_id = ? AND delta >= ? AND delta <= ? AND is_completed = false AND project_id = ?",
			taskID, startDelta, endDelta, projectId).Find(&taskExecutionDetails).Error; err != nil {
			logCtx.Error(err.Error())
			return deltas, http.StatusInternalServerError, err.Error()
		}
	}
	for _, data := range taskExecutionDetails {
		deltas = append(deltas, data.Delta)
	}
	return deltas, http.StatusOK, ""
}

// Insert a record before starting execution
func (pg *Postgres) InsertTaskBeginRecord(taskId uint64, projectId uint64, delta uint64) (int, string) {
	// THROW CONFLICT ERROR IT ITS A DUPLICATE ENTRY
	// else insert
	logCtx := log.WithFields(log.Fields{"taskId": taskId, "delta": delta})

	if taskId == 0 || delta == 0 {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest, "Missing taskID/delta"
	}

	baseTaskDetails, _, _ := pg.GetTaskDetailsById(taskId)
	isProjectEnabled := false
	if baseTaskDetails.IsProjectEnabled == true {
		isProjectEnabled = true
	}

	db := C.GetServices().Db

	taskExecDetails := model.TaskExecutionDetails{
		TaskID:      taskId,
		Delta:       delta,
		IsCompleted: false,
		CreatedAt:   U.TimeNow(),
		UpdatedAt:   U.TimeNow(),
	}

	if isProjectEnabled == true {
		taskExecDetails.ProjectID = projectId
	} else {
		taskExecDetails.ProjectID = 0
	}

	if err := db.Create(&taskExecDetails).Error; err != nil {
		if U.IsPostgresUniqueIndexViolationError("task_id_delta_unique_idx", err) {
			logCtx.Error("Trying to insert duplicate record")
			return http.StatusConflict, "Trying to insert duplicate record"
		} else {
			logCtx.Error(err.Error())
			return http.StatusConflict, err.Error()
		}
	}
	return http.StatusCreated, ""
}

// Insert a record after execution
func (pg *Postgres) InsertTaskEndRecord(taskId uint64, projectId uint64, delta uint64) (int, string) {
	// THROW CONFLICT ERROR IT ITS A DUPLICATE ENTRY
	// else insert
	logCtx := log.WithFields(log.Fields{"taskId": taskId, "delta": delta})

	if taskId == 0 || delta == 0 {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest, "Missing taskID/delta"
	}

	baseTaskDetails, _, _ := pg.GetTaskDetailsById(taskId)
	isProjectEnabled := false
	if baseTaskDetails.IsProjectEnabled == true {
		isProjectEnabled = true
	}

	db := C.GetServices().Db

	updateFields := map[string]interface{}{
		"is_completed": true,
	}
	var queryFilterString string
	if isProjectEnabled == false {
		queryFilterString = fmt.Sprintf("task_id = %v AND delta = %v", taskId, delta)
	} else {
		queryFilterString = fmt.Sprintf("task_id = %v AND delta = %v AND project_id = %v", taskId, delta, projectId)
	}
	query := db.Model(&model.TaskExecutionDetails{}).Where(queryFilterString).Updates(updateFields)
	if err := query.Error; err != nil {
		logCtx.WithError(err).Error("Failed updating task completed status")
		return http.StatusInternalServerError, "Failed updating task completed status"
	}

	if query.RowsAffected == 0 {
		logCtx.Error("No such record found")
		return http.StatusInternalServerError, "No such record found"
	}
	return http.StatusCreated, ""
}

// Delete a record if failed execution
func (pg *Postgres) DeleteTaskEndRecord(taskId uint64, projectId uint64, delta uint64) (int, string) {
	// THROW CONFLICT ERROR IT ITS A DUPLICATE ENTRY
	// else insert
	logCtx := log.WithFields(log.Fields{"taskId": taskId, "delta": delta})

	if taskId == 0 || delta == 0 {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest, "Missing taskID/delta"
	}

	baseTaskDetails, _, _ := pg.GetTaskDetailsById(taskId)
	isProjectEnabled := false
	if baseTaskDetails.IsProjectEnabled == true {
		isProjectEnabled = true
	}

	db := C.GetServices().Db

	var queryFilterString string
	if isProjectEnabled == false {
		queryFilterString = fmt.Sprintf("task_id = %v AND delta = %v", taskId, delta)
	} else {
		queryFilterString = fmt.Sprintf("task_id = %v AND delta = %v AND project_id = %v", taskId, delta, projectId)
	}

	query := db.Where(queryFilterString).Delete(&model.TaskExecutionDetails{})
	if err := query.Error; err != nil {
		logCtx.WithError(err).Error("Failed deleting task completed status")
		return http.StatusInternalServerError, "Failed deleting task completed status"
	}

	if query.RowsAffected == 0 {
		logCtx.Error("No such record found")
		return http.StatusInternalServerError, "No such record found"
	}
	return http.StatusAccepted, ""
}

// Get All the execution date/hour in the given range
func (pg *Postgres) GetAllDeltasByConfiguration(taskID uint64, lookbackInDays int, endDate *time.Time) ([]uint64, int, string) {

	// get all the processed deltas with the given range
	logCtx := log.WithFields(log.Fields{"taskID": taskID, "lookback": lookbackInDays})
	deltas := make([]uint64, 0)
	if taskID == 0 || lookbackInDays == 0 {
		logCtx.Error("missing taskID/lookback")
		return deltas, http.StatusBadRequest, "missing taskID/lookback"
	}

	taskDetails, errCode, status := pg.GetTaskDetailsById(taskID)
	if errCode != 200 {
		logCtx.Error(status)
		return deltas, http.StatusInternalServerError, status
	}
	skipEndIndex := taskDetails.SkipEndIndex
	skipStartIndex := taskDetails.SkipStartIndex
	intervals := make([]int, 0)
	if taskDetails.Frequency == model.Hourly {
		if taskDetails.SkipEndIndex == -1 {
			skipEndIndex = 23
		}
	}
	if taskDetails.Frequency == model.Daily {
		if taskDetails.SkipEndIndex == -1 {
			skipEndIndex = 6
		}
	}
	if taskDetails.Frequency == model.Weekly {
		if taskDetails.SkipEndIndex == -1 {
			skipEndIndex = 3
		}
	}
	i := skipStartIndex
	for {
		if i > skipEndIndex {
			break
		}
		intervals = append(intervals, i)
		i = i + taskDetails.FrequencyInterval
		if taskDetails.Recurrence == false {
			break
		}
	}
	var startDateTime time.Time
	var endDateTime time.Time
	if endDate == nil {
		endDateTime = U.TimeNow()
	} else {
		endDateTime = *endDate
	}
	if lookbackInDays > 0 {
		startDateTime = endDateTime.AddDate(0, 0, -lookbackInDays)
	} else {
		startDateTime = endDateTime
		endDateTime = endDateTime.AddDate(0, 0, -lookbackInDays)
	}
	if taskDetails.Frequency == model.Hourly {
		i := startDateTime
		for {
			if i.After(endDateTime) {
				break
			}
			if arrayContains(intervals, i.Hour()) {
				deltas = append(deltas, U.DateAsFormattedInt(i))
			}
			i = i.Add(time.Hour * time.Duration(1))
		}
	}
	if taskDetails.Frequency == model.Daily {
		i, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, startDateTime.Format(U.DATETIME_FORMAT_YYYYMMDD))
		for {
			if i.After(endDateTime) {
				break
			}
			if arrayContains(intervals, int(i.Weekday())) {
				deltas = append(deltas, U.DateAsFormattedInt(i))
			}
			i = i.AddDate(0, 0, 1)
		}
	}
	if taskDetails.Frequency == model.Weekly {
		// Weekly doesnt support skipping a week
		weekday := startDateTime.Weekday()
		nearestSundayIndex := int(weekday)
		i, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, startDateTime.Format(U.DATETIME_FORMAT_YYYYMMDD))
		if nearestSundayIndex != 0 {
			i = startDateTime.AddDate(0, 0, -nearestSundayIndex)
		}
		for {
			if i.After(endDateTime) {
				break
			}
			deltas = append(deltas, U.DateAsFormattedInt(i))
			i = i.AddDate(0, 0, 7)
		}
	}
	if taskDetails.Frequency == model.Stateless {
		deltas = append(deltas, U.DateAsFormattedInt(U.TimeNow()))
	}
	return deltas, http.StatusOK, ""
}

// TODO: JANANI tasks with same frequency but different offset. how to handle that? - May be avoid adding such dependencies
// avoid adding offsets for stateless
// To check if all the dependent jobs for a give date/hour range is done
func (pg *Postgres) IsDependentTaskDone(taskId uint64, projectId uint64, delta uint64) bool {
	dependentTaskOffsetMap := make(map[uint64]int, 0)
	dependentTaskStateMap := make(map[uint64]bool, 0)
	dependentTasks, _, _ := pg.GetAllDependentTasks(taskId)
	for _, dependentTask := range dependentTasks {
		dependentTaskOffsetMap[dependentTask.DependentTaskID] = dependentTask.DependencyOffset
	}
	deltaDate := getDeltaAsTime(delta)
	baseTaskDetails, _, _ := pg.GetTaskDetailsById(taskId)
	if baseTaskDetails.Frequency == model.Hourly {
		//hour -> hour
		// hour -> stateless
		for depTaskId, offset := range dependentTaskOffsetMap {
			depTaskDetails, _, _ := pg.GetTaskDetailsById(depTaskId)
			deltaDateWithDepOffset := deltaDate.Add(time.Hour * time.Duration(offset))
			if depTaskDetails.Frequency == model.Hourly {
				processedDeltas, _, _ := pg.GetAllProcessedIntervals(depTaskId, projectId, 1, &deltaDateWithDepOffset)
				if arrayUint64Contains(processedDeltas, U.DateAsFormattedInt(deltaDateWithDepOffset)) {
					dependentTaskStateMap[depTaskId] = true
				}
			} else if depTaskDetails.Frequency == model.Stateless {
				processedDeltas, _, _ := pg.GetAllProcessedIntervalsFromStartDate(depTaskId, projectId, &deltaDateWithDepOffset)
				dependentTaskStateMap[depTaskId] = isAnyHigherDeltaPresent(processedDeltas, delta)
				// if anything greateer than or equal to delta is done . mark it true
			} else {
				dependentTaskStateMap[depTaskId] = false
			}
		}
	}
	if baseTaskDetails.Frequency == model.Daily {
		// daily - hour
		// daily - daily
		// daily - stateless
		for depTaskId, offset := range dependentTaskOffsetMap {
			depTaskDetails, _, _ := pg.GetTaskDetailsById(depTaskId)
			deltaDateWithDepOffset := deltaDate.AddDate(0, 0, offset)
			if depTaskDetails.Frequency == model.Hourly {
				deltaDateWithDepOffset := deltaDateWithDepOffset.AddDate(0, 0, 1)
				processedDeltas, _, _ := pg.GetAllProcessedIntervals(depTaskId, projectId, 7, &deltaDateWithDepOffset)
				configuredDeltas, _, _ := pg.GetAllDeltasByConfiguration(depTaskId, 1, &deltaDateWithDepOffset)
				configuredDeltas = configuredDeltas[0 : len(configuredDeltas)-1]
				dependentTaskStateMap[depTaskId] = isAllDeltaPresent(processedDeltas, configuredDeltas)
				// if all the hours are present
			} else if depTaskDetails.Frequency == model.Daily {
				processedDeltas, _, _ := pg.GetAllProcessedIntervals(depTaskId, projectId, 7, &deltaDateWithDepOffset)
				if arrayUint64Contains(processedDeltas, U.DateAsFormattedInt(deltaDateWithDepOffset)) {
					dependentTaskStateMap[depTaskId] = true
				}
			} else if depTaskDetails.Frequency == model.Stateless {
				processedDeltas, _, _ := pg.GetAllProcessedIntervalsFromStartDate(depTaskId, projectId, &deltaDateWithDepOffset)
				dependentTaskStateMap[depTaskId] = isAnyHigherDeltaPresent(processedDeltas, delta)
				// if anything greateer than or equal to delta is done . mark it true
			} else {
				dependentTaskStateMap[depTaskId] = false
			}
		}
	}
	if baseTaskDetails.Frequency == model.Weekly {
		// weekly - hour
		// weekly - daily
		// weekly - week
		// weekly - stateless
		for depTaskId, offset := range dependentTaskOffsetMap {
			deltaDateWithDepOffset := deltaDate.AddDate(0, 0, (offset * 7))
			depTaskDetails, _, _ := pg.GetTaskDetailsById(depTaskId)
			if depTaskDetails.Frequency == model.Hourly {
				deltaDateWithDepOffset := deltaDateWithDepOffset.AddDate(0, 0, 1)
				processedDeltas, _, _ := pg.GetAllProcessedIntervals(depTaskId, projectId, 28, &deltaDateWithDepOffset)
				configuredDeltas, _, _ := pg.GetAllDeltasByConfiguration(depTaskId, 7, &deltaDateWithDepOffset)
				configuredDeltas = configuredDeltas[0 : len(configuredDeltas)-1]
				dependentTaskStateMap[depTaskId] = isAllDeltaPresent(processedDeltas, configuredDeltas)
			} else if depTaskDetails.Frequency == model.Daily {
				deltaDateWithDepOffset := deltaDateWithDepOffset.AddDate(0, 0, 1)
				configuredDeltas, _, _ := pg.GetAllDeltasByConfiguration(depTaskId, 7, &deltaDateWithDepOffset)
				configuredDeltas = configuredDeltas[0 : len(configuredDeltas)-1]
				processedDeltas, _, _ := pg.GetAllProcessedIntervals(depTaskId, projectId, 28, &deltaDateWithDepOffset)
				dependentTaskStateMap[depTaskId] = isAllDeltaPresent(processedDeltas, configuredDeltas)
			} else if depTaskDetails.Frequency == model.Weekly {
				processedDeltas, _, _ := pg.GetAllProcessedIntervals(depTaskId, projectId, 28, &deltaDateWithDepOffset)
				if arrayUint64Contains(processedDeltas, U.DateAsFormattedInt(deltaDateWithDepOffset)) {
					dependentTaskStateMap[depTaskId] = true
				}
			} else if depTaskDetails.Frequency == model.Stateless {
				processedDeltas, _, _ := pg.GetAllProcessedIntervalsFromStartDate(depTaskId, projectId, &deltaDateWithDepOffset)
				dependentTaskStateMap[depTaskId] = isAnyHigherDeltaPresent(processedDeltas, delta)
				// if anything greateer than or equal to delta is done . mark it true
			} else {
				dependentTaskStateMap[depTaskId] = false
			}
		}
	}
	if baseTaskDetails.Frequency == model.Stateless {
		//stateless -> stateless
		for depTaskId, _ := range dependentTaskOffsetMap {
			depTaskDetails, _, _ := pg.GetTaskDetailsById(depTaskId)
			if depTaskDetails.Frequency == model.Stateless {
				processedDeltas, _, _ := pg.GetAllProcessedIntervalsFromStartDate(depTaskId, projectId, &deltaDate)
				dependentTaskStateMap[depTaskId] = isAnyHigherDeltaPresent(processedDeltas, delta)
				// if anything greateer than or equal to delta is done . mark it true
			} else {
				dependentTaskStateMap[depTaskId] = false
			}
		}
	}
	for taskId, _ := range dependentTaskOffsetMap {
		if dependentTaskStateMap[taskId] == false {
			return false
		}
	}
	return true
}

// Get All the date/time range that are yet to be executed
func (pg *Postgres) GetAllToBeExecutedDeltas(taskId uint64, projectId uint64, lookbackInDays int, endDate *time.Time) ([]uint64, int, string) {
	if endDate == nil {
		currentTime := U.TimeNow()
		endDate = &currentTime
	}
	taskDetails, _, _ := pg.GetTaskDetailsById(taskId)
	endDateWithOffset := *endDate
	endDateWithOffset = endDateWithOffset.Add(time.Minute * time.Duration(-taskDetails.OffsetStartMinutes))
	allDeltas, _, _ := pg.GetAllDeltasByConfiguration(taskId, lookbackInDays, &endDateWithOffset)
	processedDeltas, _, _ := pg.GetAllProcessedIntervals(taskId, projectId, lookbackInDays, &endDateWithOffset)
	inProgressDeltas, _, _ := pg.GetAllInProgressIntervals(taskId, projectId, lookbackInDays, &endDateWithOffset)
	processedDeltaMap := make(map[uint64]bool)
	unprocessedDeltas := make([]uint64, 0)
	for _, delta := range processedDeltas {
		processedDeltaMap[delta] = true
	}
	for _, delta := range inProgressDeltas {
		processedDeltaMap[delta] = true
	}
	for _, delta := range allDeltas {
		if !processedDeltaMap[delta] == true {
			unprocessedDeltas = append(unprocessedDeltas, delta)
		}
	}
	return unprocessedDeltas, http.StatusOK, ""
}

func isAnyHigherDeltaPresent(processedDeltas []uint64, baseDelta uint64) bool {
	for _, delta := range processedDeltas {
		if delta > baseDelta {
			return true
		}
	}
	return false
}

func isAllDeltaPresent(processedDeltas []uint64, configuredDeltas []uint64) bool {
	for _, delta := range configuredDeltas {
		isDone := arrayUint64Contains(processedDeltas, delta)
		if !isDone {
			return false
		}
	}
	return true
}

func getDeltaAsTime(delta uint64) time.Time {
	hours := delta % 100
	datePart := fmt.Sprintf("%v", delta/100)
	deltaDate, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, datePart)
	deltaDate = deltaDate.Add(time.Hour * time.Duration(hours))
	return deltaDate
}

func arrayContains(arraySlice []int, value int) bool {
	for _, element := range arraySlice {
		if element == value {
			return true
		}
	}
	return false
}

func arrayUint64Contains(arraySlice []uint64, value uint64) bool {
	for _, element := range arraySlice {
		if element == value {
			return true
		}
	}
	return false
}
